#!/bin/bash

# Stop TucanBIT Background Script

echo "🛑 Stopping TucanBIT..."

# Check if PID file exists
if [ ! -f "tucanbit.pid" ]; then
    echo "❌ PID file not found. Checking if process is running..."
    if pgrep -f "tucanbit" > /dev/null; then
        echo "⚠️  Found running process, stopping it..."
        pkill -f "tucanbit"
        echo "✅ Process stopped!"
    else
        echo "ℹ️  No TucanBIT process found running."
    fi
    exit 0
fi

# Read PID from file
APP_PID=$(cat tucanbit.pid)

# Check if process is still running
if ! ps -p $APP_PID > /dev/null; then
    echo "ℹ️  Process $APP_PID is not running."
    rm -f tucanbit.pid
    exit 0
fi

# Stop the process
echo "🔄 Stopping process $APP_PID..."
kill $APP_PID

# Wait a bit for graceful shutdown
sleep 2

# Check if process is still running
if ps -p $APP_PID > /dev/null; then
    echo "⚠️  Process still running, force killing..."
    kill -9 $APP_PID
fi

# Clean up
rm -f tucanbit.pid
echo "✅ TucanBIT stopped successfully!"
echo "📁 Log file preserved: tucanbit.log" 