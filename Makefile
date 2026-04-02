.PHONY: run docker-build docker-up docker-down migrate migrate-up migrate-down test generate db-reset

# Database connection string
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= greenpos
DB_PASSWORD ?= greenpos123
DB_NAME ?= greenpos
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

run:
	go run ./cmd/server

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

# Database migration commands
migrate:
	@echo "Running migrations..."
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/repository/migrations \
		-database "$(DATABASE_URL)" \
		-up

migrate-up:
	@echo "Running migrations..."
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/repository/migrations \
		-database "$(DATABASE_URL)" \
		-up

migrate-down:
	@echo "Rolling back last migration..."
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/repository/migrations \
		-database "$(DATABASE_URL)" \
		-down 1

migrate-force:
	@echo "Forcing migration version..."
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/repository/migrations \
		-database "$(DATABASE_URL)" \
		-force $(VERSION)

migrate-drop:
	@echo "Dropping all tables..."
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/repository/migrations \
		-database "$(DATABASE_URL)" \
		-down -1

db-reset: docker-down migrate-drop migrate docker-up
	@echo "Database reset complete!"

db-seed:
	@echo "Seeding database..."
	@go run ./cmd/server --seed

# SQL migration files (raw SQL for reference)
sql-migration-up:
	@cat internal/repository/migrations/000001_init_schema.up.sql | psql "$(DATABASE_URL)"

sql-migration-down:
	@cat internal/repository/migrations/000001_init_schema.down.sql | psql "$(DATABASE_URL)"

# Test commands
test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-watch:
	@which reflex > /dev/null || echo "Install reflex: go install github.com/cespare/reflex@latest"
	reflex -t 200ms -r '\.go$' -- go test ./...

# Generate GraphQL code
generate:
	go generate ./...

# Development helpers
dev:
	@echo "Starting development server..."
	go run ./cmd/server

dev-migrate:
	@echo "Running migrations for development..."
	migrate

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

# Dependency management
deps:
	go mod download
	go mod tidy

# Build commands
build:
	go build -o greenpos ./cmd/server

build-linux:
	GOOS=linux GOARCH=amd64 go build -o greenpos-linux ./cmd/server

# Docker build for production
docker-build-prod:
	docker build -t greenpos/api:latest -t greenpos/api:$(shell git rev-parse --short HEAD) .

# Help
help:
	@echo "GreenPOS Makefile Commands:"
	@echo ""
	@echo "  run              - Run the server locally"
	@echo "  docker-up        - Start Docker containers"
	@echo "  docker-down      - Stop Docker containers"
	@echo "  migrate          - Run database migrations"
	@echo "  migrate-down     - Rollback last migration"
	@echo "  migrate-drop     - Drop all tables"
	@echo "  db-reset         - Reset database (drop, migrate, start)"
	@echo "  db-seed          - Seed database with sample data"
	@echo "  test             - Run tests"
	@echo "  generate         - Generate GraphQL code"
	@echo "  build            - Build binary"
	@echo "  build-linux      - Build Linux binary"
	@echo "  docker-build-prod - Build production Docker image"
	@echo "  deps             - Download and tidy dependencies"
	@echo ""
	@echo "Environment Variables:"
	@echo "  DB_HOST          - Database host (default: localhost)"
	@echo "  DB_PORT          - Database port (default: 5432)"
	@echo "  DB_USER          - Database user (default: greenpos)"
	@echo "  DB_PASSWORD      - Database password (default: greenpos123)"
	@echo "  DB_NAME          - Database name (default: greenpos)"
	@echo "  DATABASE_URL     - Full database connection string"
