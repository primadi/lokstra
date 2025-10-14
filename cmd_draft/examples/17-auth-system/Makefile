# Lokstra Auth System - Makefile
# Quick commands for common operations

.PHONY: help setup start stop clean test build docker-up docker-down db-setup db-reset

# Default target
help:
	@echo "Available commands:"
	@echo "  make setup       - Install dependencies and setup environment"
	@echo "  make start       - Start the application"
	@echo "  make stop        - Stop the application"
	@echo "  make test        - Run tests"
	@echo "  make build       - Build the application"
	@echo "  make clean       - Clean build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-up       - Start PostgreSQL and Redis in Docker"
	@echo "  make docker-down     - Stop and remove Docker containers"
	@echo "  make docker-tools    - Start with pgAdmin and Redis Commander"
	@echo ""
	@echo "Database commands:"
	@echo "  make db-setup    - Initialize database schema"
	@echo "  make db-reset    - Reset database (WARNING: deletes all data)"
	@echo "  make db-seed     - Seed database with test data"
	@echo ""

# Setup
setup:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Setup complete!"

# Start application
start:
	@echo "Starting Lokstra Auth System..."
	go run .

# Build application
build:
	@echo "Building application..."
	go build -o lokstra-auth-system .
	@echo "Build complete! Run with: ./lokstra-auth-system"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f lokstra-auth-system
	go clean
	@echo "Clean complete!"

# Docker operations
docker-up:
	@echo "Starting PostgreSQL and Redis..."
	docker-compose up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services are ready!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Containers stopped!"

docker-tools:
	@echo "Starting with management tools..."
	docker-compose --profile tools up -d
	@echo "Services started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "pgAdmin: http://localhost:5050 (admin@example.com / admin)"
	@echo "Redis Commander: http://localhost:8081"

docker-logs:
	docker-compose logs -f

# Database operations
db-setup:
	@echo "Setting up database..."
	@if command -v psql >/dev/null 2>&1; then \
		psql -U postgres -h localhost -d lokstra_auth_demo -f setup.sql; \
		echo "Database setup complete!"; \
	else \
		echo "psql not found. Using docker..."; \
		docker exec -i lokstra-postgres psql -U postgres -d lokstra_auth_demo < setup.sql; \
		echo "Database setup complete!"; \
	fi

db-reset:
	@echo "WARNING: This will delete all data in the database!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read confirm
	@echo "Resetting database..."
	@if command -v psql >/dev/null 2>&1; then \
		psql -U postgres -h localhost -d postgres -c "DROP DATABASE IF EXISTS lokstra_auth_demo;"; \
		psql -U postgres -h localhost -d postgres -c "CREATE DATABASE lokstra_auth_demo;"; \
		psql -U postgres -h localhost -d lokstra_auth_demo -f setup.sql; \
	else \
		docker exec -i lokstra-postgres psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS lokstra_auth_demo;"; \
		docker exec -i lokstra-postgres psql -U postgres -d postgres -c "CREATE DATABASE lokstra_auth_demo;"; \
		docker exec -i lokstra-postgres psql -U postgres -d lokstra_auth_demo < setup.sql; \
	fi
	@echo "Database reset complete!"

db-seed:
	@echo "Seeding database with test data..."
	@echo "This is already done by setup.sql. Check the file for seed data."

# Quick start (setup everything and run)
quickstart: docker-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) db-setup
	@echo ""
	@echo "=========================================="
	@echo "Setup complete! Starting application..."
	@echo "=========================================="
	@echo ""
	@$(MAKE) start

# Development mode (with auto-reload)
dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Generate password hash
hash-password:
	@echo "Enter password to hash:"
	@read -s password; \
	go run -tags tools ../../../common/utils/hash_password.go "$$password"
