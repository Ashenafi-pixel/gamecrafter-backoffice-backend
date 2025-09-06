#!/bin/bash

echo "🧪 Testing TucanBIT Wallet APIs"
echo "=================================="

BASE_URL="http://localhost:8080"

echo ""
echo "1️⃣ Testing Wallet Challenge..."
echo "POST /api/wallet/challenge"
curl -s -X POST "$BASE_URL/api/wallet/challenge" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
  }' | jq .

echo ""
echo "2️⃣ Testing Wallet Login (with fake signature)..."
echo "POST /api/wallet/login"
curl -s -X POST "$BASE_URL/api/wallet/login" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Welcome to TucanBIT! Please sign this message to verify your wallet ownership. Nonce: 0x1234567890abcdef",
    "nonce": "0x1234567890abcdef",
    "signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "wallet_type": "metamask"
  }' | jq .

echo ""
echo "3️⃣ Testing Swagger Documentation..."
echo "GET /swagger/index.html"
curl -s -o /dev/null -w "Status: %{http_code}\n" "$BASE_URL/swagger/index.html"

echo ""
echo "Wallet API tests completed!"
echo ""
echo "📚 For full API documentation, visit: $BASE_URL/swagger/index.html"
echo "Note: Login with fake signature will fail (expected behavior)"
echo "💡 Use real wallet signatures for successful authentication" 