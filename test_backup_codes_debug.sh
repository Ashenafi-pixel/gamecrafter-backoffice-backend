#!/bin/bash

# Script to test backup codes verification
# This will help debug the backup codes issue

echo "🔍 Testing Backup Codes Verification..."

# Configuration
BASE_URL="http://localhost:8080"
USER_ID="5a8328c7-d51b-4187-b45c-b1beea7b41ff"
BACKUP_CODE="4J2YM0UK"

echo "📧 User ID: $USER_ID"
echo "🔑 Backup Code: $BACKUP_CODE"
echo "🌐 Server: $BASE_URL"
echo ""

# Step 1: Check if user has backup codes
echo "1️⃣ Checking if user has backup codes..."
backup_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\"}" \
  "$BASE_URL/api/admin/auth/2fa/get-backup-codes")

backup_http_code=$(echo "$backup_response" | grep "HTTP_CODE:" | cut -d: -f2)
backup_body=$(echo "$backup_response" | sed '/HTTP_CODE:/d')

echo "Backup codes response:"
echo "$backup_body" | jq . 2>/dev/null || echo "$backup_body"
echo ""

if [ "$backup_http_code" -eq 200 ]; then
    echo "✅ User has backup codes available"
    
    # Extract backup codes from response
    backup_codes=$(echo "$backup_body" | jq -r '.data.backup_codes[]' 2>/dev/null)
    if [ -n "$backup_codes" ]; then
        echo "📋 Available backup codes:"
        echo "$backup_codes" | while read code; do
            echo "  - $code"
        done
        echo ""
        
        # Check if the provided code exists
        if echo "$backup_codes" | grep -q "$BACKUP_CODE"; then
            echo "✅ Backup code '$BACKUP_CODE' found in user's codes"
        else
            echo "❌ Backup code '$BACKUP_CODE' NOT found in user's codes"
            echo "💡 Try using one of the codes listed above"
        fi
    else
        echo "❌ No backup codes found in response"
    fi
else
    echo "❌ Failed to get backup codes"
    echo "💡 User might not have backup codes generated"
fi

echo ""
echo "2️⃣ Testing backup code verification..."

# Step 2: Test backup code verification
verify_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$BACKUP_CODE\",
    \"user_id\": \"$USER_ID\",
    \"method\": \"backup_codes\"
  }" \
  "$BASE_URL/api/admin/auth/2fa/verify")

verify_http_code=$(echo "$verify_response" | grep "HTTP_CODE:" | cut -d: -f2)
verify_body=$(echo "$verify_response" | sed '/HTTP_CODE:/d')

echo "Verification response:"
echo "$verify_body" | jq . 2>/dev/null || echo "$verify_body"
echo ""
echo "HTTP Status Code: $verify_http_code"

if [ "$verify_http_code" -eq 200 ]; then
    echo "🎉 SUCCESS! Backup code verification worked!"
    echo ""
    echo "📋 What this means:"
    echo "1. ✅ Backup code was valid"
    echo "2. ✅ Code was consumed (can't be used again)"
    echo "3. ✅ User should be logged in"
else
    echo "❌ Backup code verification failed"
    echo ""
    echo "🔍 Possible issues:"
    echo "1. Invalid backup code"
    echo "2. Code already used"
    echo "3. User doesn't have backup codes enabled"
    echo "4. Server error"
    echo ""
    echo "💡 Next steps:"
    echo "1. Check if user has backup codes (step 1 above)"
    echo "2. Use a valid backup code from the list"
    echo "3. Check server logs for detailed error messages"
fi

echo ""
echo "🔧 Manual Database Check (if needed):"
echo "SELECT backup_codes FROM user_2fa_settings WHERE user_id = '$USER_ID';"
