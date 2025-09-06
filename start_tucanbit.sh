#!/bin/bash
# TucanBIT Complete Startup Script

echo "🚀 Starting TucanBIT Online Casino Application..."

# Start required Docker containers
echo "📦 Starting Docker containers..."
docker start tucanbit-db
docker start redis

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check if containers are running
echo "🔍 Checking container status..."
docker ps --filter "name=tucanbit-db" --filter "name=redis" --format "table {{.Names}}\t{{.Status}}"

# Set environment variables for local development
export CONFIG_NAME=config.yaml
export REDIS_ADDR=localhost:63790
export DB_URL=postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable

echo "Services started successfully!"
echo "Application will be available at: http://localhost:8080"
echo "📚 API Documentation: http://localhost:8080/swagger/index.html"
echo ""
echo "Starting TucanBIT application..."

# Start the application
./tucanbit
