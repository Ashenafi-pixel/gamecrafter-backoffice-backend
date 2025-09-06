# TucanBIT Makefile

.PHONY: help build clean test run docker-build docker-run docker-compose-up docker-compose-down services-up services-down migrate-up migrate-down start-bg stop-bg logs-bg status-bg

# Default target
help:
	@echo "🔧 TucanBIT Build Commands"
	@echo "=========================="
	@echo "make build          - Build the application locally"
	@echo "make clean          - Clean build artifacts"
	@echo "make test           - Run tests"
	@echo "make run            - Run the application locally"
	@echo "make docker-build   - Build Docker image"
	@echo "make docker-run     - Run Docker container"
	@echo "make up             - Start all services with Docker Compose"
	@echo "make down           - Stop all services"
	@echo "make logs           - View Docker Compose logs"
	@echo "make status         - Check service status"
	@echo ""
	@echo "Local Development (bypasses Docker network issues):"
	@echo "make services-up    - Start PostgreSQL and Redis locally"
	@echo "make services-down  - Stop local services"
	@echo "make migrate-up     - Run database migrations"
	@echo "make run-local      - Run app locally with local services"
	@echo ""
	@echo "🚀 Background App Management:"
	@echo "make start-bg       - Start app in background"
	@echo "make stop-bg        - Stop background app"
	@echo "make logs-bg        - View background app logs"
	@echo "make status-bg      - Check background app status"

# Build the application locally
build:
	@echo "🔨 Building TucanBIT..."
	export GOPROXY=direct && export GOSUMDB=off && \
	go mod download && \
	go build -o tucanbit cmd/main.go
	@echo "Build completed successfully!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f tucanbit
	@echo "Clean completed!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Run the application locally
run: build
	@echo "🚀 Starting TucanBIT..."
	./tucanbit

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t tucanbit:latest .
	@echo "Docker build completed!"

# Run Docker container
docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --name tucanbit-app tucanbit:latest

# Start all services with Docker Compose
up:
	@echo "🚀 Starting all services..."
	docker-compose up --build -d
	@echo "Services started! Check status with: make status"

# Stop all services
down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "Services stopped!"

# View logs
logs:
	@echo "📋 Viewing logs..."
	docker-compose logs -f

# Check service status
status:
	@echo "📊 Service status:"
	docker-compose ps

# Restart services
restart: down up

# Full rebuild and restart
rebuild: clean docker-build up

# Local development commands (bypass Docker network issues)
services-up:
	@echo "🚀 Starting local services..."
	./start-services.sh

services-down:
	@echo "🛑 Stopping local services..."
	docker stop tucanbit-db tucanbit-redis 2>/dev/null || true
	docker rm tucanbit-db tucanbit-redis 2>/dev/null || true
	@echo "Local services stopped!"

migrate-up:
	@echo "🔄 Running migrations..."
	./run-migrations.sh

run-local: build
	@echo "🚀 Starting TucanBIT locally..."
	./run-local.sh

# Background app management commands
start-bg:
	@echo "🚀 Starting TucanBIT in background..."
	./start-app-background.sh

stop-bg:
	@echo "🛑 Stopping background TucanBIT..."
	./stop-app.sh

logs-bg:
	@echo "📋 Viewing background app logs..."
	./view-logs.sh

status-bg:
	@echo "📊 Checking background app status..."
	./check-status.sh