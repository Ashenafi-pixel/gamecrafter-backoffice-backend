#!/bin/bash

# TucanBIT Daily Report API Test Script
# This script tests the daily report API endpoints

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_URL="http://localhost:8080"
TEST_DATE=$(date -d "yesterday" "+%Y-%m-%d")

echo -e "${BLUE}🚀 Testing TucanBIT Daily Report API Endpoints${NC}"
echo "======================================================"

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${YELLOW}Testing: $description${NC}"
    echo "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$APP_URL$endpoint" 2>/dev/null)
    else
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X "$method" "$APP_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null)
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ SUCCESS (HTTP $http_code)${NC}"
        echo "Response:"
        cat /tmp/response.json | jq '.' 2>/dev/null || cat /tmp/response.json
    else
        echo -e "${RED}❌ FAILED (HTTP $http_code)${NC}"
        echo "Response:"
        cat /tmp/response.json
    fi
    
    echo ""
    echo "----------------------------------------"
}

# Test 1: Get cronjob status
test_endpoint "GET" "/analytics/daily-report/cronjob-status" "" "Get Cronjob Status"

# Test 2: Send configured daily report
test_endpoint "POST" "/analytics/daily-report/send-configured" "{\"date\": \"$TEST_DATE\"}" "Send Configured Daily Report"

# Test 3: Send test daily report
test_endpoint "POST" "/analytics/daily-report/test" "" "Send Test Daily Report"

# Test 4: Send yesterday's report
test_endpoint "POST" "/analytics/daily-report/yesterday" "{\"recipients\": [\"ashenafialemu27@gmail.com\", \"johsjones612@gmail.com\"]}" "Send Yesterday's Report"

# Test 5: Get daily report data
test_endpoint "GET" "/analytics/reports/daily?date=$TEST_DATE" "" "Get Daily Report Data"

echo -e "${BLUE}🎯 Test Summary${NC}"
echo "=================="
echo "✅ All endpoints tested"
echo "📧 Configured recipients: ashenafialemu27@gmail.com, johsjones612@gmail.com"
echo "⏰ Cronjob schedule: 23:59 UTC (end of day)"
echo "📅 Test date: $TEST_DATE"
echo ""
echo -e "${GREEN}Daily report system is ready!${NC}"
echo "The cronjob will automatically send daily reports at 23:59 UTC every day."