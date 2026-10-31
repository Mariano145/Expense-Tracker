# Design: User Authentication (JWT + bcrypt)

## Technical Approach

Implement a 3-layer Go backend (handler → service → repository) with `golang-jwt/jwt/v5` (HS256) for access tokens and `bcrypt` (cost 12) for password hashing. Refresh tokens are cryptographically random, stored as SHA-256 hashes in PostgreSQL with family-based rotation. Frontend uses Zustand for auth state, Axios interceptors for auto-refresh, and Next.js middleware for route protection. Maps to specs: `user-auth.md` (6 requirements) and `token-management.md` (6 requirements).

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| JWT library | `golang-jwt/jwt/v5` | lestrrat-go/jwx, kataras/jwt | De facto standard, best docs, simplest API for HS256 |
| Signing algorithm | HS256 | RS256, ES256 | Single secret env var, no key-pair management. Sufficient for monolith |
| Access token storage | In-memory (Zustand + sessionStorage hydration) | localStorage, httpOnly cookie | Minimizes XSS blast radius; sessionStorage clears on tab close |
| Refresh token storage | httpOnly Secure SameSite=Strict cookie | localStorage, both in cookies | Immune to XSS; SameSite=Strict eliminates CSRF for same-origin |
| Refresh token rotation | Family-based with reuse detection | Simple rotation (last wins) | Detects token theft — reuse revokes entire family, forcing re-auth |
| Password hashing | bcrypt cost 12 | argon2id, scrypt | OWASP recommendation, ~250ms/hash, well-understood, Go stdlib extension |
| DB driver | pgxpool (jackc/pgx/v5) | database/sql + lib/pq | Better performance, native PostgreSQL types, connection pool built-in |
| UUID generation | PostgreSQL `gen_random_uuid()` | Go-side UUID library | Zero dependency, DB-native, no application-side coordination |
| Frontend state | Zustand with persist middleware | Redux, Context API | Minimal boilerplate, built-in persist for hydration, no provider nesting |
| Route protection | Next.js middleware.ts | Server-side only, client-side guards | Runs at edge before page render, prevents flash of protected content |

## Data Flow

### Registration Flow

```
Client ──POST /api/auth/register──→ AuthHandler.Register
    │                                      │
    │                                      ├─→ validate email/password
    │                                      ├─→ AuthService.Register
    │                                      │       ├─→ bcrypt.Hash(password)
    │                                      │       └─→ UserRepository.Create(user)
    │                                      │
    │  ←── 201 {user: {id, email}} ────────┘
```

### Login Flow

```
Client ──POST /api/auth/login──→ AuthHandler.Login
    │                                  │
    │                                  ├─→ AuthService.Login(email, password)
    │                                  │       ├─→ UserRepository.FindByEmail
    │                                  │       ├─→ bcrypt.Compare
    │                                  │       ├─→ generate access token (JWT, 15m)
    │                                  │       ├─→ generate refresh token (random, 7d)
    │                                  │       ├─→ SHA-256 hash refresh token
    │                                  │       └─→ RefreshTokenRepository.Create
    │                                  │
    │  ←── 200 {access_token, user} ───┘
    │  ←── Set-Cookie: refresh_token (httpOnly)
```

### Token Refresh Flow (with rotation)

```
Client ──POST /api/auth/refresh──→ AuthHandler.Refresh
    │   (Cookie: refresh_token)          │
    │                                    ├─→ SHA-256(cookie value)
    │                                    ├─→ RefreshTokenRepository.FindByHash
    │                                    ├─→ IF revoked → revoke entire family (theft!)
    │                                    ├─→ IF expired → 401
    │                                    ├─→ mark old token revoked
    │                                    ├─→ generate new access token
    │                                    ├─→ generate new refresh token (same family)
    │                                    ├─→ RefreshTokenRepository.Create(new)
    │                                    │
    │  ←── 200 {access_token} ──────────┘
    │  ←── Set-Cookie: refresh_token (new, httpOnly)
```

### Protected Endpoint Flow

```
Client ──GET /api/auth/me──→ AuthMiddleware ──→ AuthHandler.Me
    │   (Authorization: Bearer <jwt>)    │
    │                                    ├─→ extract Bearer token
    │                                    ├─→ jwt.Parse + validate signature/expiry
    │                                    ├─→ inject userID/email into request context
    │                                    │
    │  ←── 200 {user: {id, email}} ────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/migrations/000002_create_users.up.sql` | Create | Users table with email uniqueness |
| `backend/migrations/000002_create_users.down.sql` | Create | Drop users table + trigger |
| `backend/migrations/000003_create_refresh_tokens.up.sql` | Create | Refresh tokens table with family support |
| `backend/migrations/000003_create_refresh_tokens.down.sql` | Create | Drop refresh_tokens table |
| `backend/internal/model/user.go` | Create | User domain type |
| `backend/internal/model/refresh_token.go` | Create | RefreshToken domain type |
| `backend/internal/model/errors.go` | Create | Domain errors (ErrDuplicateEmail, ErrInvalidCredentials, etc.) |
| `backend/internal/database/postgres.go` | Create | pgxpool connection setup + config |
| `backend/internal/repository/user_repository.go` | Create | UserRepository interface + pgx implementation |
| `backend/internal/repository/refresh_token_repository.go` | Create | RefreshTokenRepository interface + pgx implementation |
| `backend/internal/service/auth_service.go` | Create | AuthService: register, login, refresh, logout, token ops |
| `backend/internal/handler/auth.go` | Create | Auth HTTP handlers for 5 endpoints |
| `backend/internal/handler/dto.go` | Create | Request/response DTOs |
| `backend/internal/middleware/auth.go` | Create | JWT extraction + validation middleware |
| `backend/internal/middleware/cors.go` | Create | CORS middleware configuration |
| `backend/internal/config/config.go` | Create | Env var loading into typed Config struct |
| `backend/cmd/server/main.go` | Modify | Wire DB pool, repos, services, handlers, middleware |
| `backend/internal/router/router.go` | Modify | Register auth routes + middleware chain |
| `backend/.env.example` | Modify | Add JWT_ACCESS_EXPIRY, JWT_REFRESH_EXPIRY, BCRYPT_COST |
| `frontend/src/lib/api.ts` | Create | Axios instance with interceptors |
| `frontend/src/store/auth-store.ts` | Create | Zustand auth store with persist |
| `frontend/src/app/(auth)/login/page.tsx` | Create | Login page with form |
| `frontend/src/app/(auth)/register/page.tsx` | Create | Register page with form |
| `frontend/src/app/(auth)/layout.tsx` | Create | Auth layout (centered, no nav) |
| `frontend/src/app/(dashboard)/layout.tsx` | Create | Dashboard layout (with nav, protected) |
| `frontend/src/app/(dashboard)/page.tsx` | Create | Dashboard home (protected) |
| `frontend/src/middleware.ts` | Create | Route protection + redirect logic |
| `frontend/package.json` | Modify | Add zustand, axios |
| `frontend/.env.example` | Modify | No changes needed (NEXT_PUBLIC_API_URL already present) |

## Interfaces / Contracts

### Domain Models

```go
// model/user.go
type User struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// model/refresh_token.go
type RefreshToken struct {
    ID        string
    UserID    string
    TokenHash string
    FamilyID  string
    ExpiresAt time.Time
    Revoked   bool
    CreatedAt time.Time
}
```

### Domain Errors

```go
// model/errors.go — sentinel errors for service layer
var (
    ErrDuplicateEmail     = errors.New("email already registered")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrTokenExpired       = errors.New("token has expired")
    ErrTokenRevoked       = errors.New("token has been revoked")
    ErrTokenReuse         = errors.New("token reuse detected — family revoked")
    ErrInvalidToken       = errors.New("invalid token")
)
```

### Repository Interfaces

```go
// repository/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id string) (*model.User, error)
}

// repository/refresh_token_repository.go
type RefreshTokenRepository interface {
    Create(ctx context.Context, token *model.RefreshToken) error
    FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
    RevokeByHash(ctx context.Context, hash string) error
    RevokeByFamily(ctx context.Context, familyID string) error
}
```

### Service Interface

```go
// service/auth_service.go
type AuthService interface {
    Register(ctx context.Context, email, password string) (*model.User, error)
    Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
    Refresh(ctx context.Context, refreshToken string) (newAccess string, newRefresh string, err error)
    Logout(ctx context.Context, refreshToken string) error
    ValidateToken(ctx context.Context, tokenString string) (userID string, email string, err error)
}
```

### HTTP DTOs

```go
// handler/dto.go
type RegisterRequest  struct { Email string `json:"email"`; Password string `json:"password"` }
type LoginRequest     struct { Email string `json:"email"`; Password string `json:"password"` }
type AuthResponse     struct { AccessToken string `json:"access_token"`; User UserDTO `json:"user"` }
type RefreshResponse  struct { AccessToken string `json:"access_token"` }
type UserDTO          struct { ID string `json:"id"`; Email string `json:"email"`; CreatedAt string `json:"created_at,omitempty"` }
type ErrorResponse    struct { Error ErrorDetail `json:"error"` }
type ErrorDetail      struct { Code string `json:"code"`; Message string `json:"message"` }
```

### Frontend Types

```typescript
// store/auth-store.ts
interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<string | null>;
}
```

## Database Schema

```sql
-- 000002_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

```sql
-- 000003_create_refresh_tokens.up.sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit — JWT | Token generation, parsing, expiry, invalid signature | testify, no DB needed |
| Unit — bcrypt | Hash generation, password verification, wrong password | testify, no DB needed |
| Unit — Service | Register (success/duplicate/weak pw), Login (success/invalid), Refresh (rotation/reuse/expired), Logout | testify + mock repositories |
| Unit — Handler | HTTP status codes, request parsing, error response format | httptest + mock service |
| Integration — Auth flow | Register → Login → Refresh → Access protected → Logout | Real PostgreSQL (Docker), pgxpool |
| Integration — Token theft | Refresh → reuse old token → family revoked | Real PostgreSQL (Docker) |
| Frontend — Store | Login/logout state transitions, token refresh, persist hydration | Vitest, mock axios |
| Frontend — Pages | Form submission, validation errors, redirect on success | Vitest + RTL, mock store |
| Frontend — Middleware | Protected route redirect, public route pass-through | Vitest, mock request |

## Migration / Rollout

**Migration**: Sequential golang-migrate files (000002, 000003). Applied via `make migrate-up`.

**Rollback**: `make migrate-down` drops tables in reverse order. CASCADE on refresh_tokens → users ensures clean removal.

**No feature flags needed** — auth is foundational infrastructure, not a user-facing toggle.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `postgres://spendly:spendly@localhost:5432/spendly?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET` | (required, no default) | HMAC signing key, minimum 32 chars |
| `JWT_ACCESS_EXPIRY` | `15m` | Access token TTL |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token TTL (7 days) |
| `BCRYPT_COST` | `12` | bcrypt work factor |
| `CORS_ORIGIN` | `http://localhost:3000` | Allowed CORS origin |

## Open Questions

- [ ] None — all technical decisions are resolved. Ready for task breakdown.
