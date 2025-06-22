.PHONY: help build run test clean seed docker-build docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make seed         - Seed the database"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

# Build the application
build:
	go build -o main .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f main

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

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run