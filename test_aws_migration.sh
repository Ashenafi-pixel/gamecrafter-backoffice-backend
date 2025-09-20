#!/bin/bash

# TucanBIT GrooveTech AWS Migration Test Script
# This script tests the migrated GrooveTech APIs on AWS server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AWS_DOMAIN="${AWS_DOMAIN:-your-domain.com}"
TEST_ACCOUNT_ID="${TEST_ACCOUNT_ID:-a5e168fb-168e-4183-84c5-d49038ce00b5}"
TEST_SESSION_ID="${TEST_SESSION_ID:-Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7}"
TEST_GAME_ID="${TEST_GAME_ID:-82695}"

echo -e "${BLUE}🧪 TucanBIT GrooveTech AWS Migration Test${NC}"
echo -e "${BLUE}==========================================${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Test API endpoint
test_api() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"
    
    print_info "Testing $name..."
    
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$url")
    http_code="${response: -3}"
    
    if [ "$http_code" = "$expected_code" ]; then
        print_status "$name - HTTP $http_code"
        if [ -f /tmp/response.json ]; then
            echo "Response: $(cat /tmp/response.json | jq -r '.status // .message // "Success"' 2>/dev/null || cat /tmp/response.json)"
        fi
    else
        print_error "$name - HTTP $http_code (expected $expected_code)"
        if [ -f /tmp/response.json ]; then
            echo "Response: $(cat /tmp/response.json)"
        fi
    fi
    echo ""
}

# Test Game Launch API
test_game_launch() {
    print_info "Testing Game Launch API..."
    
    response=$(curl -s -X POST "https://$AWS_DOMAIN/api/groove/launch-game" \
        -H "Content-Type: application/json" \
        -d "{
            \"game_id\": \"$TEST_GAME_ID\",
            \"device_type\": \"desktop\",
            \"game_mode\": \"real\",
            \"country\": \"US\",
            \"currency\": \"USD\",
            \"language\": \"en_US\",
            \"is_test_account\": false,
            \"reality_check_elapsed\": 0,
            \"reality_check_interval\": 60
        }")
    
    if echo "$response" | jq -e '.groove_url' > /dev/null 2>&1; then
        print_status "Game Launch API - Success"
        echo "Groove URL: $(echo "$response" | jq -r '.groove_url')"
    else
        print_error "Game Launch API - Failed"
        echo "Response: $response"
    fi
    echo ""
}

# Test Get Account API
test_get_account() {
    test_api "Get Account API" \
        "https://$AWS_DOMAIN/groove-official/getaccount?request=getaccount&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&nogsgameid=$TEST_GAME_ID&apiversion=1.2"
}

# Test Get Balance API
test_get_balance() {
    test_api "Get Balance API" \
        "https://$AWS_DOMAIN/groove-official/balance?request=getbalance&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&nogsgameid=$TEST_GAME_ID&apiversion=1.2"
}

# Test Wager API
test_wager() {
    local tx_id="test_wager_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Wager API" \
        "https://$AWS_DOMAIN/groove-official/wager?request=wager&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&betamount=10.0&roundid=$round_id&transactionid=$tx_id"
}

# Test Result API
test_result() {
    local tx_id="test_result_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Result API" \
        "https://$AWS_DOMAIN/groove-official/result?request=result&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&result=15.0&roundid=$round_id&transactionid=$tx_id&gamestatus=completed"
}

# Test Rollback API
test_rollback() {
    local tx_id="test_rollback_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Rollback API" \
        "https://$AWS_DOMAIN/groove-official/rollback?request=rollback&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&rollbackamount=10.0&roundid=$round_id&transactionid=$tx_id"
}

# Test Jackpot API
test_jackpot() {
    local tx_id="test_jackpot_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Jackpot API" \
        "https://$AWS_DOMAIN/groove-official/jackpot?request=jackpot&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&gameid=$TEST_GAME_ID&apiversion=1.2&amount=50.0&roundid=$round_id&transactionid=$tx_id&gamestatus=completed"
}

# Test Rollback On Result API
test_rollback_on_result() {
    local tx_id="test_rollback_result_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Rollback On Result API" \
        "https://$AWS_DOMAIN/groove-official/reversewin?request=reversewin&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&amount=10.0&roundid=$round_id&transactionid=$tx_id"
}

# Test Rollback On Rollback API
test_rollback_on_rollback() {
    local tx_id="test_rollback_rollback_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Rollback On Rollback API" \
        "https://$AWS_DOMAIN/groove-official/rollbackrollback?request=rollbackrollback&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&rollbackAmount=10.0&roundid=$round_id&transactionid=$tx_id"
}

# Test Wager And Result API
test_wager_and_result() {
    local tx_id="test_wager_result_$(date +%s)"
    local round_id="test_round_$(date +%s)"
    
    test_api "Wager And Result API" \
        "https://$AWS_DOMAIN/groove-official/wager-and-result?request=wagerAndResult&accountid=$TEST_ACCOUNT_ID&gamesessionid=$TEST_SESSION_ID&device=desktop&gameid=$TEST_GAME_ID&apiversion=1.2&betamount=10.0&result=15.0&roundid=$round_id&transactionid=$tx_id&gamestatus=completed"
}

# Main test execution
main() {
    echo -e "${BLUE}Starting GrooveTech API tests on AWS server...${NC}"
    echo -e "${BLUE}Domain: $AWS_DOMAIN${NC}"
    echo -e "${BLUE}Account ID: $TEST_ACCOUNT_ID${NC}"
    echo -e "${BLUE}Session ID: $TEST_SESSION_ID${NC}"
    echo -e "${BLUE}Game ID: $TEST_GAME_ID${NC}"
    echo ""
    
    # Test all APIs
    test_game_launch
    test_get_account
    test_get_balance
    test_wager
    test_result
    test_rollback
    test_jackpot
    test_rollback_on_result
    test_rollback_on_rollback
    test_wager_and_result
    
    echo -e "${GREEN}🎉 GrooveTech API testing completed!${NC}"
    echo -e "${BLUE}====================================${NC}"
    echo -e "${YELLOW}Note: Some tests may fail if the account doesn't exist or has insufficient balance.${NC}"
    echo -e "${YELLOW}This is normal for testing purposes.${NC}"
    echo -e "${BLUE}====================================${NC}"
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    print_warning "jq is not installed. Installing..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y jq
    elif command -v yum &> /dev/null; then
        sudo yum install -y jq
    else
        print_error "Please install jq manually: https://stedolan.github.io/jq/"
        exit 1
    fi
fi

# Run main function
main "$@"