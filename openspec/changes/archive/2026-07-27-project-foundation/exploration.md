# Exploration: Project Foundation

## Current State

Greenfield project — no source code, no package manager, no test runner. Only SDD scaffolding exists (`openspec/config.yaml`, `.atl/skill-registry.md`). Stack confirmed: Go + net/http (backend), Next.js + React (frontend), PostgreSQL (database).

## Decision 1: Project Structure

### Options

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **Monorepo** (single repo, `backend/` + `frontend/` dirs) | Single git history, shared CI/CD, easy cross-stack changes, simpler for solo dev | Repo grows larger, Go modules need explicit root config | Low |
| **Polyrepo** (separate repos per service) | Independent deployment, team isolation | Overhead for solo dev, shared types harder, two CI/CD pipelines | Medium |

### Recommendation: **Monorepo**

Rationale:
- Solo developer (plus AI) — no team isolation needed
- Shared CI/CD pipeline simplifies DevOps
- Cross-stack changes (API contract + frontend consumer) in one PR = atomic reviews
- Go modules work fine in subdirectories with `go mod init` in `backend/`
- Next.js naturally lives in its own directory

**Proposed layout:**
```
billetera/
├── backend/                  # Go API
│   ├── cmd/
│   │   └── server/           # Main entrypoint
│   │       └── main.go
│   ├── internal/             # Private application code
│   │   ├── handler/          # HTTP handlers (controllers)
│   │   ├── model/            # Domain models
│   │   ├── repository/       # Data access layer
│   │   ├── service/          # Business logic
│   │   └── middleware/        # Auth, logging, CORS
│   ├── migrations/           # SQL migration files
│   ├── go.mod
│   └── go.sum
├── frontend/                 # Next.js app
│   ├── app/                  # App Router
│   │   ├── (auth)/           # Auth group (login, register)
│   │   ├── (dashboard)/      # Dashboard group
│   │   ├── layout.tsx
│   │   └── page.tsx
│   ├── components/
│   │   ├── ui/               # Shared UI primitives
│   │   └── features/         # Feature-specific components
│   ├── lib/                  # Utilities, API client
│   ├── stores/               # Zustand stores
│   ├── types/                # TypeScript types
│   ├── package.json
│   └── next.config.js
├── docker-compose.yml        # Local dev (PostgreSQL, backend, frontend)
├── .github/
│   └── workflows/            # CI/CD
├── openspec/                 # SDD artifacts (already exists)
└── README.md
```

---

## Decision 2: Authentication

### Options

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **JWT (custom)** | Stateless, simple, no external deps, full control | Must handle refresh tokens, token revocation is harder | Low |
| **JWT + Google OAuth** | Best UX for zero user, social login | More complex setup, Google Cloud Console config | Medium |
| **Session-based (cookie)** | Server-side revocation, familiar pattern | Requires session store (Redis), not RESTful by default | Medium |
| **NextAuth.js** | Built-in OAuth providers, session management | Couples auth to frontend, Go backend needs separate validation | Medium |

### Recommendation: **JWT with bcrypt, OAuth deferred**

Rationale:
- **MVP phase**: Custom JWT auth with bcrypt password hashing
  - Go's `golang.org/x/crypto/bcrypt` for password hashing
  - `github.com/golang-jwt/jwt/v5` for JWT creation/validation
  - Access token (15min) + Refresh token (7 days) pattern
  - Simple, no external dependencies, full control
- **Post-MVP**: Add Google OAuth when zero user needs frictionless signup
  - Keep JWT as the token format; OAuth just replaces the registration/login flow
- **Multi-user**: Each user has an account; expenses belong to a user via `user_id` foreign key
- **Cross-device sync**: JWT stored in httpOnly cookie (secure, works on mobile web)

**Auth flow:**
```
Register → POST /api/auth/register {email, password} → 201 + JWT
Login    → POST /api/auth/login    {email, password} → 200 + JWT
Refresh  → POST /api/auth/refresh  {refresh_token}   → 200 + new JWT
Logout   → POST /api/auth/logout                      → 204 (invalidate refresh)
```

---

## Decision 3: API Design

### Options

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **REST** | Simple, well-understood, Go net/http native, great tooling | Over/under-fetching possible | Low |
| **GraphQL** | Precise data fetching, single endpoint | Complexity for MVP, Go GraphQL libs are heavier, learning curve | High |
| **RPC (gRPC)** | Fast, typed contracts | Not browser-friendly without proxy, overkill | High |

### Recommendation: **REST**

Rationale:
- Go's `net/http` is built for REST — `http.ServeMux` handles routing naturally
- Zero external dependencies for MVP (Go 1.22+ has enhanced routing with method matching)
- Clear resource-based URLs are easy to understand and debug
- Frontend can fetch exactly what it needs with query params
- Well-supported by all tools (curl, Postman, Next.js fetch)

**Proposed endpoints:**
```
# Auth
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout

# Expenses
GET    /api/expenses                    # List (with filters: ?from=&to=&category=&min=&max=)
POST   /api/expenses                    # Create
GET    /api/expenses/{id}               # Get by ID
PUT    /api/expenses/{id}               # Update
DELETE /api/expenses/{id}               # Delete

# Categories
GET    /api/categories                  # List all (user's + defaults)
POST   /api/categories                  # Create custom
PUT    /api/categories/{id}             # Update
DELETE /api/categories/{id}             # Delete (only custom)

# Dashboard/Analytics
GET    /api/dashboard/summary           # Monthly summary, totals by category
GET    /api/dashboard/trends            # Spending trends over time

# User
GET    /api/user/profile                # Current user profile
PUT    /api/user/profile                # Update profile
```

**Response format (consistent envelope):**
```json
{
  "data": { ... },
  "meta": { "page": 1, "per_page": 20, "total": 150 }
}
```

**Error format:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Amount must be positive",
    "details": [{"field": "amount", "message": "must be > 0"}]
  }
}
```

---

## Decision 4: Testing Frameworks

### Go Backend

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **Standard `testing` + `testify`** | Idiomatic, testify adds assertions + mocks, widely used | testify is an external dep | Low |
| **Standard `testing` only** | Zero deps, idiomatic | Verbose assertions, manual mocks | Low |
| **GoMock** | Interface-based mocking, code generation | More ceremony, heavier setup | Medium |

### Recommendation: **Standard `testing` + `testify`**

Rationale:
- `testify/assert` and `testify/require` reduce boilerplate dramatically
- `testify/suite` for integration tests with setup/teardown
- `testify/mock` for interface mocking (simpler than GoMock for learning)
- Standard `testing` for benchmarks and simple unit tests
- Table-driven tests (idiomatic Go) + testify assertions = best of both worlds

```go
// Example pattern
func TestCreateExpense(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateExpenseRequest
        wantErr bool
    }{
        {"valid expense", CreateExpenseRequest{Amount: 25.50, Category: "food"}, false},
        {"negative amount", CreateExpenseRequest{Amount: -10, Category: "food"}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test logic with assert/require
        })
    }
}
```

### Next.js Frontend

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **Jest + React Testing Library** | Industry standard, great ecosystem | Slower than Vitest | Low |
| **Vitest + React Testing Library** | Fast, Vite-native, Jest-compatible API | Slightly less ecosystem coverage | Low |

### Recommendation: **Vitest + React Testing Library**

Rationale:
- Vitest is faster (uses Vite's transform pipeline)
- Jest-compatible API — minimal learning curve
- Next.js works well with Vitest (official examples exist)
- React Testing Library encourages testing user behavior, not implementation

---

## Decision 5: Frontend State Management

### Options

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **React Context API** | Built-in, zero deps | Re-renders on any change, boilerplate for multiple contexts | Low |
| **Zustand** | Tiny (~1KB), simple API, no boilerplate, TypeScript-first | External dep | Low |
| **Redux Toolkit** | Powerful, great DevTools | Heavy for this use case, steep learning curve | Medium |
| **Jotai** | Atomic state, fine-grained updates | Different paradigm, less intuitive for beginners | Medium |

### Recommendation: **Zustand**

Rationale:
- **~1KB bundle** — negligible impact
- **Simplest API** — `create()` with state + actions, that's it
- **TypeScript-first** — full type inference with `create<T>()`
- **No providers** — unlike Context/Redux, no wrapper components needed
- **Built-in persist middleware** — for offline-first / optimistic updates
- **Perfect for learning** — minimal concepts to understand

```typescript
// Example: expense store
interface ExpenseStore {
  expenses: Expense[]
  isLoading: boolean
  addExpense: (expense: CreateExpenseDTO) => Promise<void>
  fetchExpenses: (filters?: ExpenseFilters) => Promise<void>
}

const useExpenseStore = create<ExpenseStore>()((set, get) => ({
  expenses: [],
  isLoading: false,
  addExpense: async (expense) => {
    set({ isLoading: true })
    const newExpense = await api.createExpense(expense)
    set((state) => ({ expenses: [...state.expenses, newExpense], isLoading: false }))
  },
  fetchExpenses: async (filters) => {
    set({ isLoading: true })
    const expenses = await api.getExpenses(filters)
    set({ expenses, isLoading: false })
  },
}))
```

**Store structure:**
```
stores/
├── useAuthStore.ts       # Auth state (user, token, login/logout)
├── useExpenseStore.ts    # Expenses CRUD + filters
├── useCategoryStore.ts   # Categories list
└── useDashboardStore.ts  # Dashboard data
```

---

## Decision 6: Deployment Strategy

### Options

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **Vercel (FE) + Railway (BE)** | Zero-config deploys, free tiers, Git-based | Vendor lock-in, Railway costs at scale | Low |
| **Docker + VPS (Hetzner/DigitalOcean)** | Full control, cheaper at scale | Manual setup, maintenance burden | Medium |
| **Docker + AWS/GCP** | Enterprise-grade, scalable | Complex, expensive for MVP | High |
| **Vercel + Render** | Similar to Vercel+Railway, good free tier | Cold starts on free tier | Low |

### Recommendation: **Vercel (frontend) + Railway (backend) for MVP, Docker for local dev**

Rationale:
- **Vercel**: Zero-config Next.js deployment, edge functions, free tier sufficient for MVP
- **Railway**: Easy Go deployment, PostgreSQL add-on, $5/month starter
- **Docker Compose**: Local development environment (PostgreSQL + backend + frontend)
- **Migration path**: When scaling needs arise, Docker images can move to any VPS/cloud

**Deployment architecture:**
```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Vercel    │────▶│   Railway    │────▶│  PostgreSQL  │
│  (Next.js)  │ API │   (Go API)   │     │  (Railway)   │
└─────────────┘     └──────────────┘     └──────────────┘
```

---

## Decision 7: Development Workflow

### Local Setup

**Docker Compose** for local development:
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: billetera
      POSTGRES_USER: billetera
      POSTGRES_PASSWORD: devpass
    ports: ["5432:5432"]
    volumes: ["pgdata:/var/lib/postgresql/data"]

  backend:
    build: ./backend
    ports: ["8080:8080"]
    depends_on: [postgres]
    environment:
      DATABASE_URL: postgres://billetera:devpass@postgres:5432/billetera?sslmode=disable
      JWT_SECRET: dev-secret-change-in-prod
    volumes:
      - ./backend:/app
    command: go run cmd/server/main.go

  frontend:
    build: ./frontend
    ports: ["3000:3000"]
    depends_on: [backend]
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080/api
    volumes:
      - ./frontend:/app
      - /app/node_modules

volumes:
  pgdata:
```

### Database Migrations

**Recommendation: `golang-migrate/migrate`**

Rationale:
- CLI + Go library — can run migrations from code or CI/CD
- Simple SQL file-based migrations (no ORM magic)
- Supports PostgreSQL natively
- Easy to integrate with Railway deployment

```
backend/migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_categories.up.sql
├── 000002_create_categories.down.sql
├── 000003_create_expenses.up.sql
└── 000003_create_expenses.down.sql
```

### Environment Variables

```
# Backend (.env)
DATABASE_URL=postgres://billetera:devpass@localhost:5432/billetera?sslmode=disable
JWT_SECRET=dev-secret-change-in-prod
PORT=8080
CORS_ORIGIN=http://localhost:3000

# Frontend (.env.local)
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

### Git Workflow (confirmed)

```
main (production) ← develop ← feature/*
```

---

## Risks

- **Go learning curve**: Using standard library (no Gin/Echo) means more manual routing/handler code. Mitigates by keeping patterns simple and well-documented.
- **JWT refresh token rotation**: If not implemented correctly, can lead to security issues. Mitigate with one-time-use refresh tokens + token family tracking.
- **Next.js App Router complexity**: Server Components vs Client Components can confuse beginners. Mitigate by starting with Client Components for interactive parts.
- **Railway free tier limits**: May need to upgrade if usage grows. Low risk for MVP.

## Ready for Proposal

**Yes** — all foundational decisions are explored. The orchestrator should tell the user:

> Exploración completa. Las 7 decisiones fundamentales están analizadas con recomendaciones claras. Las decisiones clave son: monorepo, JWT custom, REST API, testify+Vitest, Zustand, Vercel+Railway, Docker Compose local. ¿Pasamos a proposal?
