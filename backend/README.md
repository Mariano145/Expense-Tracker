# Spendly Backend

Go HTTP API using the standard library (`net/http`). No frameworks.

## Prerequisites

- Go 1.22+
- PostgreSQL 16 (via Docker Compose at the repo root)

## Quick Start

```bash
# Copy environment variables
cp .env.example .env

# Run the server
go run ./cmd/server
```

The server starts on `http://localhost:8080` by default. Override with `PORT`.

## Project Layout

```
backend/
├── cmd/server/main.go     # Entrypoint — DI wiring + graceful shutdown
├── internal/
│   ├── auth/               # JWT (HS256), bcrypt, refresh token generation
│   ├── config/             # Env loading + validation
│   ├── database/           # pgxpool factory
│   ├── model/              # Domain models + sentinel errors
│   ├── repository/         # Postgres repositories (interfaces + impl)
│   ├── service/            # AuthService (business logic)
│   ├── handler/            # HTTP handlers + DTOs + input validation
│   ├── middleware/         # Bearer auth + CORS
│   ├── router/             # Route registration (http.ServeMux)
│   └── integration/        # End-to-end auth flow tests
├── migrations/             # SQL migrations (golang-migrate)
├── .env.example            # Environment variable template
└── go.mod
```

## Testing

Needs the test database once: `make db-test-setup` (from repo root).

```bash
go test -p 1 ./...   # -p 1: packages share the spendly_test DB, run sequentially
```

For details see [../docs/setup.md](../docs/setup.md).

## API Endpoints

| Method | Path                  | Auth            | Description                     |
|--------|-----------------------|-----------------|---------------------------------|
| GET    | `/health`             | —               | Health check (200)              |
| POST   | `/api/auth/register`  | —               | Create account                  |
| POST   | `/api/auth/login`     | —               | Start session                   |
| POST   | `/api/auth/refresh`   | refresh cookie  | Rotate refresh, new access      |
| POST   | `/api/auth/logout`    | refresh cookie  | Revoke family, clear cookie     |
| GET    | `/api/auth/me`        | Bearer token    | Current user profile            |

Full request/response reference: [../docs/api/auth.md](../docs/api/auth.md).
