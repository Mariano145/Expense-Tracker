# Getting Started (Development)

Self-host / local development guide for Spendly. The end-user install paths
(Android app, hosted web) are not published yet — see the root [README](../README.md).

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://docs.docker.com/get-docker/) + Docker Compose

## First-time setup

```bash
git clone git@github.com:Mariano145/spendly.git && cd spendly

cd backend && go mod download && cd ..
cd frontend && npm install && cd ..

# Database
make db-up
make migrate-up        # applies SQL migrations via golang-migrate (Docker)

# Environment (first time only)
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
```

## Run

```bash
# Terminal 1 — API on :8080
cd backend && go run ./cmd/server
curl localhost:8080/health   # → {"status":"ok"}

# Terminal 2 — Web app on :3000
cd frontend && npm run dev
```

Open http://localhost:3000 → register → login → dashboard.

## Test

```bash
cd backend && go test -p 1 ./...   # ⚠️ -p 1 required (see gotcha below)
cd frontend && npm test            # vitest + RTL
```

DB-backed backend tests need the test database once:

```bash
make db-test-setup   # creates spendly_test + applies migrations
```

> **Gotcha:** `repository` and `integration` test packages share the
> `spendly_test` database and clean their tables at setup. Running packages
> in parallel (`go test ./...` without `-p 1`) makes them wipe each other's
> rows and fail intermittently with FK errors. Sequential run is the rule.

## Makefile

| Command              | Description                          |
|----------------------|--------------------------------------|
| `make db-up`         | Start PostgreSQL (Docker Compose)    |
| `make db-down`       | Stop PostgreSQL                      |
| `make migrate-up`    | Apply all pending migrations         |
| `make migrate-down`  | Rollback last migration              |
| `make db-test-setup` | Create + migrate `spendly_test` DB   |

## Environment variables

### Backend (`backend/.env`)

| Variable             | Default                                    | Description                        |
|----------------------|--------------------------------------------|------------------------------------|
| `PORT`               | `8080`                                     | Server port                        |
| `DATABASE_URL`       | —                                          | PostgreSQL connection string       |
| `JWT_SECRET`         | —                                          | HS256 signing key                  |
| `JWT_ACCESS_EXPIRY`  | `15m`                                      | Access token lifetime              |
| `JWT_REFRESH_EXPIRY` | `168h`                                     | Refresh token lifetime             |
| `BCRYPT_COST`        | `12`                                       | bcrypt cost factor                 |
| `CORS_ORIGIN`        | —                                          | Allowed CORS origin                |

### Frontend (`frontend/.env.local`)

| Variable              | Default | Description          |
|-----------------------|---------|----------------------|
| `NEXT_PUBLIC_API_URL` | —       | Backend API base URL |
