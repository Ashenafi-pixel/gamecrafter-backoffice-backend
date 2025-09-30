#!/bin/bash

# WebSocket Test Script for TucanBIT Cashback System
# This script tests the WebSocket connection and cashback flow

echo "🔌 TucanBIT WebSocket Test Script"
echo "================================="

# Configuration
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1IiwiaXNfdmVyaWZpZWQiOmZhbHNlLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInBob25lX3ZlcmlmaWVkIjpmYWxzZSwiZXhwIjoxNzU5MjY4OTgyLCJpYXQiOjE3NTkxODI1ODIsImlzcyI6InR1Y2FuYml0Iiwic3ViIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1In0._VW0FM2BO4I1ukdkwcpaYpVs2QUXRYzhu4j7a72hodA"
BASE_URL="https://api.tucanbit.tv"
WS_URL="wss://api.tucanbit.tv/ws/balance/player"

echo "📋 Configuration:"
echo "  Base URL: $BASE_URL"
echo "  WebSocket URL: $WS_URL"
echo "  User ID: a5e168fb-168e-4183-84c5-d49038ce00b5"
echo ""

# Check if websocat is installed
if ! command -v websocat &> /dev/null; then
    echo "❌ websocat is not installed. Installing..."
    echo "Please install websocat first:"
    echo "  curl -LsSf https://astral.sh/uv/install.sh | sh"
    echo "  uv tool install websocat"
    echo ""
    echo "Or use the HTML test file: test_websocket_connection.html"
    exit 1
fi

echo "✅ websocat found. Starting WebSocket test..."
echo ""

# Test WebSocket connection
echo "🔌 Testing WebSocket connection..."
echo "Sending authentication message..."

# Create auth message
AUTH_MESSAGE='{"type":"auth","access_token":"'$ACCESS_TOKEN'"}'

echo "Auth message: $AUTH_MESSAGE"
echo ""

# Connect to WebSocket and send auth message
echo "📡 Connecting to WebSocket..."
echo "Press Ctrl+C to disconnect"
echo ""

# Use websocat to connect and send auth message
echo "$AUTH_MESSAGE" | websocat "$WS_URL" --text --ping-interval 30 --ping-timeout 10

echo ""
echo "🔌 WebSocket connection closed."