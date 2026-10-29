# Billetera

Expense tracking web app — multi-user, cross-device sync, visual dashboards.

## Tech Stack

| Layer    | Technology                          |
|----------|-------------------------------------|
| Backend  | Go 1.22+ (`net/http`, no framework) |
| Frontend | Next.js 16 (App Router, TypeScript) |
| Database | PostgreSQL 16                       |
| Auth     | JWT + bcrypt                        |
| State    | Zustand                             |
| Testing  | testify (Go), Vitest + RTL (JS)     |
| CI/CD    | GitHub Actions                      |
| Deploy   | Hetzner VPS + Dokku (provisional)   |

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

## Quick Start

### 1. Clone and install

```bash
git clone <repo-url> && cd billetera

# Backend dependencies (none beyond stdlib + testify)
cd backend && go mod download && cd ..

# Frontend dependencies
cd frontend && npm install && cd ..
```

### 2. Start the database

```bash
# Start PostgreSQL via Docker Compose
make db-up

# Verify it's ready
docker compose exec postgres pg_isready -U billetera
```

### 3. Run migrations

```bash
make migrate-up
```

### 4. Start the backend

```bash
cd backend
cp .env.example .env   # first time only
go run ./cmd/server
```

Server runs on `http://localhost:8080`. Health check: `curl localhost:8080/health`

### 5. Start the frontend

```bash
cd frontend
cp .env.example .env.local   # first time only
npm run dev
```

App runs on `http://localhost:3000`.

## Testing

```bash
# Backend (unit + repository + integration tests)
cd backend
go test -p 1 ./...        # -p 1: packages share the billetera_test DB; run sequentially

# Frontend
cd frontend && npm test
```

Backend DB-backed tests need the test database up once:

```bash
make db-up && make db-test-setup   # creates billetera_test + applies migrations
```

## Makefile Commands

| Command             | Description                                |
|---------------------|--------------------------------------------|
| `make db-up`        | Start PostgreSQL                           |
| `make db-down`      | Stop PostgreSQL                            |
| `make migrate-up`   | Apply all pending migrations               |
| `make migrate-down` | Rollback last migration                    |
| `make db-test-setup`| Create + migrate `billetera_test` DB       |

## Project Layout

```
billetera/
├── backend/                 # Go API server
│   ├── cmd/server/          # Entrypoint (main.go, DI wiring, graceful shutdown)
│   ├── internal/
│   │   ├── auth/            # JWT sign/parse, bcrypt, refresh token generation
│   │   ├── config/          # Env loading + validation
│   │   ├── database/        # pgxpool factory
│   │   ├── handler/         # HTTP handlers (auth, health, DTOs, input validation)
│   │   ├── middleware/      # Bearer auth middleware, CORS
│   │   ├── model/           # Domain models + sentinel errors
│   │   ├── repository/      # Postgres repositories (users, refresh_tokens)
│   │   ├── router/          # Route registration
│   │   ├── service/         # AuthService (business logic)
│   │   └── integration/     # End-to-end auth flow tests
│   ├── migrations/          # SQL migrations (golang-migrate)
│   └── .env.example
├── frontend/                # Next.js 16 app (App Router)
│   ├── src/
│   │   ├── app/
│   │   │   ├── (auth)/      # Public pages: login, register
│   │   │   └── (dashboard)/ # Protected pages
│   │   ├── components/      # Reusable React components
│   │   ├── lib/api.ts       # Axios client + auto-refresh interceptors
│   │   ├── store/           # Zustand auth store (persist)
│   │   └── proxy.ts         # Route protection (Next.js 16 middleware)
│   ├── vitest.config.ts
│   └── .env.example
├── docs/api/                # API endpoint reference
├── .github/workflows/       # CI pipeline
├── docker-compose.yml       # PostgreSQL for local dev
├── Makefile                 # db-up, migrate-up, db-test-setup
└── openspec/                # SDD artifacts (specs, design, tasks)
```

## Authentication

The API uses short-lived JWT access tokens (HS256, 15 min) plus rotating
refresh tokens (7 days) delivered as an `httpOnly Secure SameSite=Strict`
cookie scoped to `/api/auth`. Refresh tokens use family-based rotation with
reuse detection: reusing a revoked token invalidates the whole family and
forces re-login.

### Endpoints

| Method | Path                 | Auth          | Description                          |
|--------|----------------------|---------------|--------------------------------------|
| POST   | `/api/auth/register` | —             | Create account (email + password)    |
| POST   | `/api/auth/login`    | —             | Obtain access token + refresh cookie |
| POST   | `/api/auth/refresh`  | refresh cookie| Rotate refresh token, new access     |
| POST   | `/api/auth/logout`   | refresh cookie| Revoke family, clear cookie          |
| GET    | `/api/auth/me`       | Bearer token  | Current user profile                 |
| GET    | `/health`            | —             | Liveness check                       |

Full request/response reference: [docs/api/auth.md](./docs/api/auth.md).

## Environment Variables

### Backend (`backend/.env`)

| Variable              | Default  | Description                          |
|-----------------------|----------|--------------------------------------|
| `PORT`                | `8080`   | Server port                          |
| `DATABASE_URL`        | —        | PostgreSQL connection string         |
| `JWT_SECRET`          | —        | Secret key for JWT signing (HS256)   |
| `JWT_ACCESS_EXPIRY`   | `15m`    | Access token lifetime                |
| `JWT_REFRESH_EXPIRY`  | `168h`   | Refresh token lifetime (7 days)      |
| `BCRYPT_COST`         | `12`     | bcrypt hashing cost                  |
| `CORS_ORIGIN`         | —        | Allowed CORS origin                  |

### Frontend (`frontend/.env.local`)

| Variable              | Default | Description            |
|-----------------------|---------|------------------------|
| `NEXT_PUBLIC_API_URL` | —       | Backend API base URL   |

## License

MIT — see [LICENSE](./LICENSE)
