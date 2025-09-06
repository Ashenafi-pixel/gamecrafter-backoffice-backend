#!/bin/bash

# View TucanBIT Logs Script

echo "📋 Viewing TucanBIT logs..."

# Check if log file exists
if [ ! -f "tucanbit.log" ]; then
    echo " Log file not found. The application might not be running."
    echo "💡 Start the app first: ./start-app-background.sh"
    exit 1
fi

# Check if app is running
if [ ! -f "tucanbit.pid" ]; then
    echo "⚠️  PID file not found. Checking if process is running..."
    if pgrep -f "tucanbit" > /dev/null; then
        echo "Application is running. Viewing logs..."
    else
        echo " Application is not running."
        echo "💡 Start the app first: ./start-app-background.sh"
        exit 1
    fi
fi

echo "📊 Log file: tucanbit.log"
echo "🔄 Press Ctrl+C to stop viewing logs"
echo ""

# Show last 50 lines and then follow
echo "📋 Last 50 lines:"
echo "=================="
tail -n 50 tucanbit.log

echo ""
echo "🔄 Following logs in real-time..."
echo "=================="

# Follow logs in real-time
tail -f tucanbit.log 