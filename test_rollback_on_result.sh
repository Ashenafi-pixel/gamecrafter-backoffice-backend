#!/bin/bash

# Test script for Rollback On Result API
# Based on the documentation example

echo "Testing Rollback On Result API..."

# Set the endpoint
ENDPOINT="http://localhost:8080/api/v1/groove/official/reversewin"

# Test parameters
ACCOUNT_ID="111"
GAME_SESSION_ID="123_jdhdujdk"
DEVICE="desktop"
GAME_ID="80102"
AMOUNT="10.0"
ROUND_ID="nc8n4nd87"
TRANSACTION_ID="trx_id"
WIN_TRANSACTION_ID="win_trx_id"
API_VERSION="1.2"

# Build the request URL
REQUEST_URL="${ENDPOINT}?request=reversewin&gamesessionid=${GAME_SESSION_ID}&accountid=${ACCOUNT_ID}&device=${DEVICE}&gameid=${GAME_ID}&amount=${AMOUNT}&roundid=${ROUND_ID}&transactionid=${TRANSACTION_ID}&wintransactionid=${WIN_TRANSACTION_ID}&apiversion=${API_VERSION}"

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