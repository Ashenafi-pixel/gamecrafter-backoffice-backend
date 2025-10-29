#!/bin/bash

# Test script for withdrawal pause system
# This script tests the API endpoints to ensure the system is working correctly

BASE_URL="http://localhost:8094"
ADMIN_TOKEN="your_admin_token_here"  # Replace with actual admin token

echo "Testing Withdrawal Pause System"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -n "Testing $description... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}PASS${NC}"
        if [ -f /tmp/response.json ]; then
            echo "  Response: $(cat /tmp/response.json | jq -r '.message // .data // .' 2>/dev/null || cat /tmp/response.json)"
        fi
    else
        echo -e "${RED}FAIL${NC} (Expected: $expected_status, Got: $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "  Response: $(cat /tmp/response.json)"
        fi
    fi
    echo
}

# Test 1: Get global status
test_endpoint "GET" "/api/v1/system-config/withdrawal/global-status" "" "200" "Get global withdrawal status"

# Test 2: Get thresholds
test_endpoint "GET" "/api/v1/system-config/withdrawal/thresholds" "" "200" "Get withdrawal thresholds"

# Test 3: Get manual review settings
test_endpoint "GET" "/api/v1/system-config/withdrawal/manual-review" "" "200" "Get manual review settings"

# Test 4: Get pause reasons
test_endpoint "GET" "/api/v1/system-config/withdrawal/pause-reasons" "" "200" "Get pause reasons"

# Test 5: Get paused withdrawals
test_endpoint "GET" "/api/v1/withdrawal-management/paused" "" "200" "Get paused withdrawals"

# Test 6: Get withdrawal stats
test_endpoint "GET" "/api/v1/withdrawal-management/stats" "" "200" "Get withdrawal statistics"

# Test 7: Update global status (pause withdrawals)
test_endpoint "PUT" "/api/v1/system-config/withdrawal/global-status" '{
    "enabled": false,
    "reason": "Test pause - system maintenance"
}' "200" "Pause withdrawals globally"

# Test 8: Update global status (enable withdrawals)
test_endpoint "PUT" "/api/v1/system-config/withdrawal/global-status" '{
    "enabled": true,
    "reason": "Test complete - system operational"
}' "200" "Enable withdrawals globally"

# Test 9: Update thresholds
test_endpoint "PUT" "/api/v1/system-config/withdrawal/thresholds" '{
    "hourly_volume": {
        "value": 75000,
        "currency": "USD",
        "active": true
    },
    "daily_volume": {
        "value": 1500000,
        "currency": "USD",
        "active": true
    },
    "single_transaction": {
        "value": 15000,
        "currency": "USD",
        "active": true
    },
    "user_daily": {
        "value": 7500,
        "currency": "USD",
        "active": true
    }
}' "200" "Update withdrawal thresholds"

# Test 10: Update manual review settings
test_endpoint "PUT" "/api/v1/system-config/withdrawal/manual-review" '{
    "enabled": true,
    "threshold_amount": 7500,
    "currency": "USD",
    "require_kyc": true
}' "200" "Update manual review settings"

echo "Test Summary"
echo "============"
echo "All tests completed. Check the results above."
echo
echo "Note: Some tests may fail if:"
echo "1. The server is not running on $BASE_URL"
echo "2. The admin token is invalid or missing"
echo "3. The database is not properly initialized"
echo
echo "To initialize the database, run:"
echo "psql -d tucanbit -f scripts/init_withdrawal_pause_system.sql"
