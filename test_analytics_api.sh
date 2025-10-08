#!/bin/bash

# Test Analytics API Endpoints
# This script tests the analytics endpoints to ensure they're working

echo "=== Testing Analytics API Endpoints ==="
echo ""

# Base URL (adjust if needed)
BASE_URL="http://localhost:8080"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1IiwiaXNfdmVyaWZpZWQiOmZhbHNlLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInBob25lX3ZlcmlmaWVkIjpmYWxzZSwiZXhwIjoxNzU5MjExNTAyLCJpYXQiOjE3NTkxMjUxMDIsImlzcyI6InR1Y2FuYml0Iiwic3ViIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1In0.gzyyL55i2OFVEtnaRBJ9632vomVt-iSh3Lel2ZsF3Lg"

# Test Real-time Stats
echo "1. Testing Real-time Stats..."
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "$BASE_URL/analytics/realtime/stats" | jq '.' 2>/dev/null || echo "Failed to parse JSON or endpoint not responding"
echo ""

# Test Daily Report (today)
TODAY=$(date +%Y-%m-%d)
echo "2. Testing Daily Report for $TODAY..."
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "$BASE_URL/analytics/reports/daily?date=$TODAY" | jq '.' 2>/dev/null || echo "Failed to parse JSON or endpoint not responding"
echo ""

# Test Daily Report (yesterday)
YESTERDAY=$(date -d "yesterday" +%Y-%m-%d)
echo "3. Testing Daily Report for $YESTERDAY..."
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "$BASE_URL/analytics/reports/daily?date=$YESTERDAY" | jq '.' 2>/dev/null || echo "Failed to parse JSON or endpoint not responding"
echo ""

# Test Enhanced Daily Report
echo "4. Testing Enhanced Daily Report..."
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     "$BASE_URL/analytics/reports/daily-enhanced?date=$TODAY" | jq '.' 2>/dev/null || echo "Failed to parse JSON or endpoint not responding"
echo ""

echo "=== Test Complete ==="
echo ""
echo "If you see 'Failed to parse JSON or endpoint not responding', check:"
echo "1. Backend server is running on $BASE_URL"
echo "2. Analytics endpoints are properly configured"
echo "3. ClickHouse database is running and accessible"
echo "4. Token is valid and has proper permissions"
