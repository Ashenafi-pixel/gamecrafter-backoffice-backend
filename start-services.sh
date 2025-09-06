#!/bin/bash

# Start required services for TucanBIT

echo "🚀 Starting TucanBIT services..."

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker run -d --name tucanbit-db \
    -p 5433:5432 \
    -e POSTGRES_USER=tucanbit \
    -e POSTGRES_PASSWORD=5kj0YmV5FKKpU9D50B7yH5A \
    -e POSTGRES_DB=tucanbit \
    postgres:13

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Start Redis
echo "🔴 Starting Redis..."
docker run -d --name tucanbit-redis \
    -p 63790:6379 \
    redis:6.2

# Wait for Redis to be ready
echo "⏳ Waiting for Redis to be ready..."
sleep 5

echo "Services started successfully!"
echo "📊 Service status:"
echo "   - PostgreSQL: localhost:5433"
echo "   - Redis: localhost:63790"

echo ""
echo "🎯 Next steps:"
echo "   1. Run migrations: ./run-migrations.sh"
echo "   2. Start the app: ./run-local.sh"
echo "   3. Or use: make run" 