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
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

## Project Layout

```
billetera/
├── backend/                 # Go API server
│   ├── cmd/server/          # Entrypoint (main.go)
│   ├── internal/
│   │   ├── handler/         # HTTP handlers
│   │   └── router/          # Route registration
│   ├── migrations/          # SQL migrations (golang-migrate)
│   └── .env.example
├── frontend/                # Next.js app
│   ├── src/
│   │   ├── app/             # App Router pages & layouts
│   │   └── components/      # Reusable React components
│   ├── vitest.config.ts
│   └── .env.example
├── .github/workflows/       # CI pipeline
├── docker-compose.yml       # PostgreSQL for local dev
├── Makefile                 # db-up, db-down, migrate-up, migrate-down
└── openspec/                # SDD artifacts (specs, design, tasks)
```

## Environment Variables

### Backend (`backend/.env`)

| Variable       | Default | Description                  |
|----------------|---------|------------------------------|
| `PORT`         | `8080`  | Server port                  |
| `DATABASE_URL` | —       | PostgreSQL connection string |
| `JWT_SECRET`   | —       | Secret key for JWT signing   |
| `CORS_ORIGIN`  | —       | Allowed CORS origin          |

### Frontend (`frontend/.env.local`)

| Variable              | Default | Description            |
|-----------------------|---------|------------------------|
| `NEXT_PUBLIC_API_URL` | —       | Backend API base URL   |

## Makefile Commands

| Command        | Description                         |
|----------------|-------------------------------------|
| `make db-up`   | Start PostgreSQL                    |
| `make db-down` | Stop PostgreSQL                     |
| `make migrate-up`   | Apply all pending migrations  |
| `make migrate-down` | Rollback last migration        |

## License

MIT — see [LICENSE](./LICENSE)
