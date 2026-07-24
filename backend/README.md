# Billetera Backend

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
├── cmd/server/main.go        # Entrypoint — HTTP server + graceful shutdown
├── internal/
│   ├── handler/               # HTTP handlers (health, auth, expenses, …)
│   └── router/router.go       # Route registration (http.ServeMux)
├── migrations/                # SQL migrations (golang-migrate)
├── .env.example               # Environment variable template
└── go.mod
```

## Testing

```bash
go test ./...
```

## API Endpoints

| Method | Path      | Description         |
|--------|-----------|---------------------|
| GET    | `/health` | Health check (200)  |
