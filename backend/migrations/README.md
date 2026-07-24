# Database Migrations

SQL migrations managed with [golang-migrate](https://github.com/golang-migrate/migrate).

## Usage

### Apply all pending migrations

```bash
# Using Docker (recommended — no local install needed)
make migrate-up

# Using migrate CLI directly
migrate -path backend/migrations -database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" up
```

### Rollback last migration

```bash
make migrate-down

# Or directly
migrate -path backend/migrations -database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" down 1
```

### Create a new migration

```bash
migrate create -ext sql -dir backend/migrations -seq <name>
```

This creates a pair of `.up.sql` and `.down.sql` files.

## Conventions

- File names: `NNNNNN_<name>.up.sql` / `NNNNNN_<name>.down.sql`
- Every `.up.sql` MUST have a corresponding `.down.sql` that reverses it
- Migrations are applied in order; never edit an already-applied migration
