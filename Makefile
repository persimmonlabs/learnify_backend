# Learnify API Makefile

.PHONY: help build run test clean docker-up docker-down migrate

# Default target
help:
	@echo "Learnify API - Available Commands"
	@echo ""
	@echo "  make build        - Build the API binary"
	@echo "  make run          - Run the API locally"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make docker-up    - Start Docker environment (DB + API)"
	@echo "  make docker-down  - Stop Docker environment"
	@echo "  make migrate      - Run database migrations"
	@echo "  make lint         - Run Go linter"
	@echo ""

# Build the API binary
build:
	@echo "Building API..."
	go build -o bin/api ./cmd/api

# Run the API locally
run:
	@echo "Running API..."
	go run ./cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

# Start Docker environment
docker-up:
	@echo "Starting Docker environment..."
	docker-compose up -d

# Stop Docker environment
docker-down:
	@echo "Stopping Docker environment..."
	docker-compose down

# Run database migrations
migrate:
	@echo "Running migrations..."
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/000_migration_tracker.sql
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/001_create_identity_tables.sql
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/002_create_curriculum_tables.sql
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/003_create_progress_tables.sql
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/004_create_social_tables.sql
	psql postgresql://postgres:postgres@localhost:5432/learnify -f migrations/005_create_discovery_tables.sql
	@echo "Migrations complete!"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Development workflow: format, lint, test
dev: fmt lint test
	@echo "Development checks complete!"
