.PHONY: help db-up db-down migrate-up migrate-down db-test-setup

# Show available commands
help:
	@echo "Available commands:"
	@echo "  make db-up          - Start PostgreSQL via Docker Compose"
	@echo "  make db-down        - Stop PostgreSQL"
	@echo "  make migrate-up     - Run all pending migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make db-test-setup  - Setup test database (create + migrate)"
	@echo ""

# Start PostgreSQL via Docker Compose
db-up:
	docker compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@docker compose exec -T postgres pg_isready -U billetera -d billetera --timeout 30

# Stop PostgreSQL
db-down:
	docker compose down

# Setup test database (create + migrate)
db-test-setup:
	@echo "Creating test database..."
	@docker compose exec -T postgres psql -U billetera -c "CREATE DATABASE billetera_test;" 2>/dev/null || echo "Database billetera_test already exists"
	@echo "Applying migrations to test database..."
	docker run --rm --network host -v $(PWD)/backend/migrations:/migrations:Z migrate/migrate \
		-path /migrations \
		-database "postgres://billetera:billetera@localhost:5432/billetera_test?sslmode=disable" \
		up

# Run all pending migrations (uses Docker to avoid local install)
migrate-up:
	@docker compose up -d postgres
	@docker compose exec -T postgres pg_isready -U billetera -d billetera --timeout 30 > /dev/null
	docker run --rm --network host -v $(PWD)/backend/migrations:/migrations:Z migrate/migrate \
		-path /migrations \
		-database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" \
		up

# Rollback last migration
migrate-down:
	docker run --rm --network host -v $(PWD)/backend/migrations:/migrations:Z migrate/migrate \
		-path /migrations \
		-database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" \
		down 1
