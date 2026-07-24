.PHONY: db-up db-down migrate-up migrate-down

# Start PostgreSQL via Docker Compose
db-up:
	docker compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@docker compose exec -T postgres pg_isready -U billetera -d billetera --timeout 30

# Stop PostgreSQL
db-down:
	docker compose down

# Run all pending migrations (uses Docker to avoid local install)
migrate-up:
	@docker compose up -d postgres
	@docker compose exec -T postgres pg_isready -U billetera -d billetera --timeout 30 > /dev/null
	docker run --rm --network host -v $(PWD)/backend/migrations:/migrations migrate/migrate \
		-path /migrations \
		-database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" \
		up

# Rollback last migration
migrate-down:
	docker run --rm --network host -v $(PWD)/backend/migrations:/migrations migrate/migrate \
		-path /migrations \
		-database "postgres://billetera:billetera@localhost:5432/billetera?sslmode=disable" \
		down 1
