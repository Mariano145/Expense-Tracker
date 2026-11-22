# Architecture

Monorepo: `backend/` (Go) + `frontend/` (Next.js) + PostgreSQL 16.
API reference lives in [`docs/api/`](./api/auth.md).

## Backend — Go, no framework, 3-layer clean

```
HTTP → router → middleware (auth, CORS) → handler → service → repository → pgxpool → PostgreSQL
                                              (DTOs)     (rules)   (interfaces)
```

- **Handler** (`internal/handler/`): request decoding, input validation
  (email format, password length → `400`), DTO mapping, sentinel-error →
  HTTP status mapping. No business logic.
- **Service** (`internal/service/`): business rules — register, login,
  refresh with **family rotation + reuse detection**, logout, token
  validation. Depends on repository *interfaces*, never concrete types
  (mockable without a DB).
- **Repository** (`internal/repository/`): SQL via `pgx/v5`. Interfaces
  declared and implemented in the same package.
- **`internal/auth/`**: focused helpers — JWT HS256 sign/parse, bcrypt,
  random refresh tokens + SHA-256 hashing.
- **`internal/integration/`**: end-to-end auth flows against a real Postgres.
- **Wiring**: manual dependency injection in `cmd/server/main.go`
  (config → pool → repos → service → handler → router), graceful shutdown.
- **Errors**: sentinel errors in `internal/model/errors.go`
  (`ErrEmailExists` → 409, `ErrInvalidCredentials`/token errors → 401, …).
- **Migrations**: `backend/migrations/` via golang-migrate; UUIDs generated
  by PostgreSQL (`gen_random_uuid()`); `updated_at` triggers.

### Security model

| Concern            | Design                                                              |
|--------------------|---------------------------------------------------------------------|
| Passwords          | bcrypt, cost 12                                                     |
| Access token       | JWT HS256, 15 min, kept in memory/sessionStorage (never localStorage)|
| Refresh token      | random; stored as SHA-256 hash in DB; `httpOnly Secure SameSite=Strict` cookie scoped to `/api/auth`, 7 days |
| Rotation           | family-based: every refresh rotates the token; reuse of a revoked token revokes the whole family (theft detection → forced re-login) |
| Route middleware   | Bearer validation injects `userID`/`email` into request context     |
| CORS               | single configurable origin (`CORS_ORIGIN`)                          |

## Frontend — Next.js 16 (App Router), React 19, TypeScript, Tailwind, Zustand

```
src/
├── app/
│   ├── (auth)/       login + register (public route group)
│   └── (dashboard)/  protected area
├── lib/api.ts        Axios client: withCredentials, Bearer request interceptor,
│                     401 → refresh-once → retry; on failure → logout + redirect
├── store/            Zustand + persist (session: user, access token)
└── proxy.ts          route protection (⚠️ Next.js 16 renamed middleware.ts → proxy.ts,
                      Node.js runtime — not Edge)
```

## Data model (today)

```
users (id uuid pk, email unique, password_hash, created_at, updated_at)
refresh_tokens (id, user_id fk, token_hash, family_id, expires_at, revoked, created_at)
```

Domain tables (expenses, categories) are **not implemented yet** — the
product core is the next milestone.

## Conventions

- Commits: Conventional Commits (`feat|fix|docs|test|chore(scope): …`)
- Branches: `main` (stable) / `develop` (integration) / `feature/*`
- Tests colocated with code (`*_test.go`, `*.test.tsx`)
- Backend test run **must** use `go test -p 1 ./...` (shared test DB)
- Code/comments in English; user-facing strings in Spanish

## CI

GitHub Actions (`.github/workflows/ci.yml`): `go vet` + `go test` + `go build`
for backend; lint + vitest + production build for frontend.
