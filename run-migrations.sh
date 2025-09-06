#!/bin/bash

# Run database migrations for TucanBIT

echo "🔄 Running database migrations..."

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5433 -U tucanbit > /dev/null 2>&1; then
    echo " PostgreSQL is not running. Please start services first:"
    echo "   ./start-services.sh"
    exit 1
fi

# Set database URL
export DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable"

# Check if migrate tool is available
if ! command -v migrate > /dev/null 2>&1; then
    echo "📥 Installing migrate tool..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Run migrations
echo "🚀 Running migrations from ./migrations directory..."
migrate -database "$DB_URL" -path migrations -verbose up

echo "Migrations completed!"
echo "🎯 You can now start the application with: ./run-local.sh" 