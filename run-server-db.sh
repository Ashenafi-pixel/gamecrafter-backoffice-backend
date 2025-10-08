#!/bin/bash

# Run TucanBIT Backend with Server Database
# This script runs the backend connected to the server database via SSH tunnel

echo "🚀 Starting TucanBIT Backend with Server Database..."
echo "=================================================="

# Check if SSH tunnel is running
if ! nc -z localhost 5433 2>/dev/null; then
    echo "❌ SSH tunnel not detected on port 5433"
    echo "💡 Please start the SSH tunnel first:"
    echo "   ssh -fN -L 5433:localhost:5433 ubuntu@13.51.168.77 -i ~/Developer/Upwork/Tucanbit/Tucanbit/TucanBIT.pem"
    exit 1
fi

echo "✅ SSH tunnel detected on port 5433"

# Set environment variables for server database
export SKIP_PERMISSION_INIT=true
export CONFIG_NAME=config
export REDIS_URL="redis://localhost:63790/0"

echo "🔧 Environment configured for server database"
echo "   - SKIP_PERMISSION_INIT=true (skips local permission creation)"
echo "   - Database: postgres://tucanbit:***@localhost:5433/tucanbit"
echo "   - Redis: redis://localhost:63790/0"

# Change to backend directory
cd "$(dirname "$0")"

echo "🏃 Starting backend..."
go run cmd/main.go
