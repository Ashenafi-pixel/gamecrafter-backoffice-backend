#!/bin/bash

# Debug Analytics API 500 Error
echo "=== Debugging Analytics API 500 Error ==="
echo ""

# Test the specific endpoint that's failing
echo "1. Testing the failing endpoint:"
echo "URL: http://localhost:8080/analytics/reports/daily?date=2025-10-08"
echo ""

curl -v -H "Content-Type: application/json" \
     "http://localhost:8080/analytics/reports/daily?date=2025-10-08" 2>&1 | head -20

echo ""
echo ""

# Test if the server is running
echo "2. Testing if server is running:"
curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "Server is running"
else
    echo "Server is not responding"
fi

echo ""
echo ""

# Test basic analytics endpoint
echo "3. Testing basic analytics endpoint:"
curl -s -H "Content-Type: application/json" \
     "http://localhost:8080/analytics/realtime/stats" 2>/dev/null | head -5

echo ""
echo ""

# Check if ClickHouse is running (if applicable)
echo "4. Checking ClickHouse connection:"
echo "Note: ClickHouse might be required for analytics data"
echo ""

echo "=== Common Solutions ==="
echo "1. Make sure the backend server is running: ./tucanbit"
echo "2. Check if ClickHouse is running and accessible"
echo "3. Verify the analytics service is properly configured"
echo "4. Check server logs for detailed error messages"
echo "5. Ensure the database has data for the requested date"
