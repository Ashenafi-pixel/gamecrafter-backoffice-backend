#!/bin/bash

# Local run script for TucanBIT (bypasses Docker network issues)

set -e

echo "🚀 Starting TucanBIT locally..."

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the project root."
    exit 1
fi

# Check if binary exists, if not build it
if [ ! -f "tucanbit" ]; then
    echo "🔨 Binary not found, building application..."
    export GOPROXY=direct
    export GOSUMDB=off
    go mod download
    go build -o tucanbit cmd/main.go
    echo "✅ Build completed!"
fi

# Check if required services are running
echo "🔍 Checking required services..."

# Check PostgreSQL
if ! pg_isready -h localhost -p 5433 -U tucanbit > /dev/null 2>&1; then
    echo "⚠️  PostgreSQL not running on port 5433"
    echo "💡 You can start it with: docker run -d --name tucanbit-db -p 5433:5432 -e POSTGRES_USER=tucanbit -e POSTGRES_PASSWORD=5kj0YmV5FKKpU9D50B7yH5A -e POSTGRES_DB=tucanbit postgres:13"
fi

# Check Redis
if ! redis-cli -p 63790 ping > /dev/null 2>&1; then
    echo "⚠️  Redis not running on port 63790"
    echo "💡 You can start it with: docker run -d --name redis -p 63790:6379 redis:6.2"
fi

# Set environment variables for local run
export CONFIG_FILE="./config.yaml"
export DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable"
export APP_HOST="0.0.0.0"
export APP_PORT="8080"
export JWT_SECRET="tokensecrethere"
export REDIS_ADDR="localhost:63790"
export KAFKA_BOOTSTRAP_SERVER="localhost:9093"
export KAFKA_TOPIC="events"

echo "🌐 Starting TucanBIT on http://localhost:8080"
echo "📊 Environment:"
echo "   - Database: $DB_URL"
echo "   - Redis: $REDIS_ADDR"
echo "   - Kafka: $KAFKA_BOOTSTRAP_SERVER"
echo "   - Port: $APP_PORT"
echo "   - Config: $CONFIG_FILE"

# Run the application
./tucanbit 