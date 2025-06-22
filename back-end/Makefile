
# Makefile
.PHONY: help build run test clean migrate-up migrate-down seed docker-build docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make seed         - Seed the database"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

# Build the application
build:
	go build -o main cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f main
	rm -rf uploads/

# Run database migrations
migrate-up:
	migrate -path internal/database/migrations -database "${DATABASE_URL}" up

# Rollback database migrations
migrate-down:
	migrate -path internal/database/migrations -database "${DATABASE_URL}" down

# Seed the database
seed:
	go run scripts/seed.go

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Development with hot reload
dev:
	air

# Generate QR codes for all tables
generate-qr:
	go run scripts/generate_qr.go