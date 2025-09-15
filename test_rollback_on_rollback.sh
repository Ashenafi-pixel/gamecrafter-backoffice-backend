#!/bin/bash

# Test script for Rollback On Rollback API
# Based on the documentation example

echo "Testing Rollback On Rollback API..."

# Set the endpoint
ENDPOINT="http://localhost:8080/api/v1/groove/official/rollbackrollback"

# Test parameters
ACCOUNT_ID="a5e168fb-168e-4183-84c5-d49038ce00b5"
GAME_SESSION_ID="Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7"
DEVICE="desktop"
GAME_ID="82695"
ROLLBACK_AMOUNT="5.0"
ROUND_ID="rollback_rollback_test_final"
TRANSACTION_ID="rollback_rollback_test_final_tx"
API_VERSION="1.2"

# Build the request URL
REQUEST_URL="${ENDPOINT}?request=rollbackrollback&gamesessionid=${GAME_SESSION_ID}&accountid=${ACCOUNT_ID}&device=${DEVICE}&gameid=${GAME_ID}&rollbackAmount=${ROLLBACK_AMOUNT}&roundid=${ROUND_ID}&transactionid=${TRANSACTION_ID}&apiversion=${API_VERSION}"

echo "Request URL: ${REQUEST_URL}"
echo ""

# Make the request
echo "Making request..."
curl -X GET "${REQUEST_URL}" \
  -H "Content-Type: application/json" \
  -H "X-Groove-Signature: test_signature" \
  -v

echo ""
echo "Test completed."