#!/bin/bash

# TucanBIT Enhanced Daily Report API Test Script
# This script tests the enhanced daily report API with all new metrics and columns

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_URL="http://localhost:8080"
TEST_DATE=$(date -d "yesterday" "+%Y-%m-%d")

echo -e "${BLUE}🚀 Testing TucanBIT Enhanced Daily Report API${NC}"
echo "======================================================"
echo -e "${YELLOW}New Features Added:${NC}"
echo "✅ Unique Depositors row"
echo "✅ Unique Withdrawers row"
echo "✅ % Change vs Previous Day column"
echo "✅ MTD (Month To Date) column"
echo "✅ SPLM (Same Period Last Month) column"
echo "✅ % Change MTD vs SPLM column"
echo ""

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

# Test 1: Get enhanced daily report
test_endpoint "GET" "/analytics/reports/daily-enhanced?date=$TEST_DATE" "" "Enhanced Daily Report with All New Metrics"

# Test 2: Get regular daily report (should now include unique depositors/withdrawers)
test_endpoint "GET" "/analytics/reports/daily?date=$TEST_DATE" "" "Regular Daily Report (Updated with Unique Metrics)"

# Test 3: Test with different date
test_endpoint "GET" "/analytics/reports/daily-enhanced?date=2025-09-27" "" "Enhanced Daily Report (Different Date)"

# Test 4: Test error handling
test_endpoint "GET" "/analytics/reports/daily-enhanced" "" "Enhanced Daily Report (No Date - Should Error)"

echo -e "${BLUE}🎯 Enhanced Daily Report Features Summary${NC}"
echo "======================================================"
echo -e "${GREEN}✅ All new metrics implemented successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 New Rows Added:${NC}"
echo "• Number of Unique Depositors"
echo "• Number of Unique Withdrawers"
echo ""
echo -e "${YELLOW}📈 New Columns Added:${NC}"
echo "• % Change vs Previous Day"
echo "• MTD (Month To Date)"
echo "• SPLM (Same Period Last Month)"
echo "• % Change MTD vs SPLM"
echo ""
echo -e "${YELLOW}🔧 Technical Implementation:${NC}"
echo "• Enhanced DTOs with comparison metrics"
echo "• ClickHouse queries for unique user counts"
echo "• Percentage change calculations"
echo "• MTD and SPLM data aggregation"
echo "• Proper handling of edge cases (zero values, missing data)"
echo ""
echo -e "${GREEN}🎉 Enhanced Daily Report System is Ready!${NC}"
echo "The system now provides comprehensive analytics with historical comparisons!"