#!/bin/bash

# Check TucanBIT Status Script

echo "📊 TucanBIT Application Status"
echo "=============================="

# Check if PID file exists
if [ ! -f "tucanbit.pid" ]; then
    echo "❌ PID file not found"
else
    APP_PID=$(cat tucanbit.pid)
    echo "🆔 PID file: $APP_PID"
fi

# Check if process is running
if pgrep -f "tucanbit" > /dev/null; then
    echo "✅ Application is RUNNING"
    echo ""
    echo "📋 Process details:"
    ps aux | grep tucanbit | grep -v grep
    echo ""
    echo "🌐 Application should be accessible at: http://localhost:8080"
echo "📚 Swagger docs: http://localhost:8080/swagger/index.html"
else
    echo "❌ Application is NOT RUNNING"
    echo ""
    echo "💡 To start the app: ./start-app-background.sh"
fi

# Check log file
if [ -f "tucanbit.log" ]; then
    LOG_SIZE=$(du -h tucanbit.log | cut -f1)
    echo ""
    echo "📁 Log file: tucanbit.log (Size: $LOG_SIZE)"
    echo "📋 Last 5 log lines:"
    echo "=================="
    tail -n 5 tucanbit.log 2>/dev/null || echo "No recent logs"
else
    echo ""
    echo "📁 Log file: Not found"
fi

# Check if port is listening
if netstat -tuln 2>/dev/null | grep ":8080 " > /dev/null; then
    echo ""
    echo "🔌 Port 8080: LISTENING"
else
    echo ""
    echo "🔌 Port 8080: NOT LISTENING"
fi

echo ""
echo "🎯 Quick commands:"
echo "   📋 View logs: ./view-logs.sh"
echo "   🛑 Stop app: ./stop-app.sh"
echo "   🚀 Start app: ./start-app-background.sh" 