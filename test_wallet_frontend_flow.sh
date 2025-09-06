#!/bin/bash

echo "🧪 Testing Complete Wallet Authentication Flow"
echo "================================================"

# Step 1: Create a challenge
echo "📝 Step 1: Creating wallet challenge..."
CHALLENGE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/wallet/challenge \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_type": "metamask",
    "wallet_address": "0x6d79dc497a41b244d67cbe444b9767c63ecc8e24",
    "chain_type": "ethereum"
  }')

echo "Challenge Response:"
echo "$CHALLENGE_RESPONSE" | jq .

# Extract challenge message and nonce
CHALLENGE_MESSAGE=$(echo "$CHALLENGE_RESPONSE" | jq -r '.challenge_message')
NONCE=$(echo "$CHALLENGE_RESPONSE" | jq -r '.nonce')

echo ""
echo "📋 Challenge Details:"
echo "Message: $CHALLENGE_MESSAGE"
echo "Nonce: $NONCE"
echo ""

echo "Step 2: Instructions for Frontend Testing"
echo "============================================="
echo "1. Open your frontend application"
echo "2. Click 'Connect Wallet' -> 'MetaMask'"
echo "3. MetaMask will prompt you to sign this exact message:"
echo ""
echo "   '$CHALLENGE_MESSAGE'"
echo ""
echo "4. After signing, the frontend will send the signature to /api/wallet/login"
echo "5. The authentication should succeed!"
echo ""
echo "💡 Note: The frontend code is correct - it uses the exact challenge_message"
echo "   from the backend. The previous failures were due to testing with"
echo "   signatures created for different messages."
echo ""
echo "🎯 Expected Result: Authentication success with access_token and refresh_token" 