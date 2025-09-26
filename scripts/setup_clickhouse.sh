#!/bin/bash

# ClickHouse Setup Script for TucanBIT Casino Analytics
# This script sets up ClickHouse with Docker and initializes the schema

set -e

echo "🚀 Setting up ClickHouse for TucanBIT Casino Analytics..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

print_status "Docker is running ✓"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed. Please install docker-compose and try again."
    exit 1
fi

print_status "docker-compose is available ✓"

# Create network if it doesn't exist
print_status "Creating Docker network..."
docker network create tucanbit-network 2>/dev/null || print_warning "Network already exists"

# Start ClickHouse
print_status "Starting ClickHouse container..."
docker-compose -f docker-compose.clickhouse.yaml up -d

# Wait for ClickHouse to be ready
print_status "Waiting for ClickHouse to be ready..."
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if docker exec tucanbit-clickhouse wget --no-verbose --tries=1 --spider http://localhost:8123/ping 2>/dev/null; then
        print_success "ClickHouse is ready!"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        print_error "ClickHouse failed to start after $max_attempts attempts"
        exit 1
    fi
    
    print_status "Attempt $attempt/$max_attempts - ClickHouse not ready yet, waiting 2 seconds..."
    sleep 2
    attempt=$((attempt + 1))
done

# Initialize schema
print_status "Initializing ClickHouse schema..."
if docker exec -i tucanbit-clickhouse clickhouse-client --multiquery < clickhouse/schema.sql; then
    print_success "Schema initialized successfully!"
else
    print_error "Failed to initialize schema"
    exit 1
fi

# Test connection
print_status "Testing ClickHouse connection..."
if docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 'ClickHouse is working!' as status"; then
    print_success "ClickHouse connection test passed!"
else
    print_error "ClickHouse connection test failed"
    exit 1
fi

# Show connection information
print_success "ClickHouse setup completed successfully!"
echo ""
echo "📊 ClickHouse Connection Information:"
echo "   Host: localhost"
echo "   Port: 8123 (HTTP) / 9000 (Native)"
echo "   Database: tucanbit_analytics"
echo "   Username: tucanbit"
echo "   Password: tucanbit_clickhouse_password"
echo ""
echo "🔗 Web Interface: http://localhost:8123"
echo "📝 ClickHouse Client: docker exec -it tucanbit-clickhouse clickhouse-client"
echo ""
echo "📋 Available Tables:"
docker exec tucanbit-clickhouse clickhouse-client --query "SHOW TABLES FROM tucanbit_analytics"

echo ""
print_success "Setup complete! You can now start using ClickHouse for analytics."
echo ""
echo "Next steps:"
echo "1. Update your Go application configuration to include ClickHouse settings"
echo "2. Start your TucanBIT application"
echo "3. Begin syncing transaction data to ClickHouse"
echo "4. Use the analytics APIs to query data"