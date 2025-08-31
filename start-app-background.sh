#!/bin/bash

# Start TucanBIT in Background Script

set -e

echo "🚀 Starting TucanBIT in the background..."

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

# Check if app is already running
if pgrep -f "tucanbit" > /dev/null; then
    echo "⚠️  TucanBIT is already running!"
    echo "📊 Process info:"
    ps aux | grep tucanbit | grep -v grep
    echo ""
    echo "🛑 To stop it: ./stop-app.sh"
    echo "📋 To view logs: ./view-logs.sh"
    exit 0
fi

# Set environment variables for background run
export CONFIG_FILE="./config/config.yaml"
export DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable"
export APP_HOST="0.0.0.0"
export APP_PORT="8080"
export JWT_SECRET="tokensecrethere"
export REDIS_ADDR="localhost:63790"
export KAFKA_BOOTSTRAP_SERVER="localhost:9093"
export KAFKA_TOPIC="events"

# Start the application in background
echo "🌐 Starting TucanBIT on http://localhost:8080"
echo "📊 Environment:"
echo "   - Database: $DB_URL"
echo "   - Redis: $REDIS_ADDR"
echo "   - Kafka: $KAFKA_BOOTSTRAP_SERVER"
echo "   - Port: $APP_PORT"
echo "   - Config: $CONFIG_FILE"

# Run in background and save PID
nohup ./tucanbit > tucanbit.log 2>&1 &
APP_PID=$!

# Save PID to file for management
echo $APP_PID > tucanbit.pid

echo "✅ TucanBIT started successfully in background!"
echo "🆔 Process ID: $APP_PID"
echo "📁 Log file: tucanbit.log"
echo "📁 PID file: tucanbit.pid"
echo ""
echo "🎯 Management commands:"
echo "   📋 View logs: ./view-logs.sh"
echo "   🛑 Stop app: ./stop-app.sh"
echo "   📊 Check status: ./check-status.sh"
echo "   🌐 Open Swagger: http://localhost:8080/swagger/index.html"
echo ""
echo "💡 The app will continue running even if you close this terminal!" 