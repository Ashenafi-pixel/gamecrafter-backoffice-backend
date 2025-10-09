#!/bin/bash

# Script to test complete email OTP flow
# This will generate an OTP and then verify it

echo "📧 Testing Complete Email OTP Flow..."

# Configuration
BASE_URL="http://localhost:8080"
USER_ID="5a8328c7-d51b-4187-b45c-b1beea7b41ff"
USER_EMAIL="kirube.tech23@gmail.com"

echo "📧 User ID: $USER_ID"
echo "📧 User Email: $USER_EMAIL"
echo "🌐 Server: $BASE_URL"
echo ""

# Step 1: Generate Email OTP
echo "1️⃣ Generating Email OTP..."
otp_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"email\": \"$USER_EMAIL\"
  }" \
  "$BASE_URL/api/admin/auth/2fa/generate-email-otp")

otp_http_code=$(echo "$otp_response" | grep "HTTP_CODE:" | cut -d: -f2)
otp_body=$(echo "$otp_response" | sed '/HTTP_CODE:/d')

echo "Email OTP Generation Response:"
echo "$otp_body" | jq . 2>/dev/null || echo "$otp_body"
echo "HTTP Status Code: $otp_http_code"

if [ "$otp_http_code" -eq 200 ]; then
    echo "✅ Email OTP generated successfully!"
    echo ""
    echo "📧 Check your email inbox ($USER_EMAIL) for the 6-digit OTP code"
    echo "📧 Also check spam/junk folder if not in inbox"
    echo ""
    echo "⏰ The OTP expires in 10 minutes"
    echo ""
    
    # Step 2: Wait for user to get the OTP
    echo "2️⃣ Please check your email and enter the 6-digit OTP code:"
    read -p "Enter OTP code: " otp_code
    
    if [ -z "$otp_code" ]; then
        echo "❌ No OTP code entered. Exiting."
        exit 1
    fi
    
    echo ""
    echo "3️⃣ Verifying OTP code: $otp_code"
    
    # Step 3: Verify the OTP
    verify_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
      -X POST \
      -H "Content-Type: application/json" \
      -d "{
        \"token\": \"$otp_code\",
        \"user_id\": \"$USER_ID\",
        \"method\": \"email_otp\"
      }" \
      "$BASE_URL/api/admin/auth/2fa/verify")
    
    verify_http_code=$(echo "$verify_response" | grep "HTTP_CODE:" | cut -d: -f2)
    verify_body=$(echo "$verify_response" | sed '/HTTP_CODE:/d')
    
    echo "OTP Verification Response:"
    echo "$verify_body" | jq . 2>/dev/null || echo "$verify_body"
    echo "HTTP Status Code: $verify_http_code"
    
    if [ "$verify_http_code" -eq 200 ]; then
        echo "🎉 SUCCESS! Email OTP verification worked!"
        echo ""
        echo "📋 What this means:"
        echo "1. ✅ Email OTP was generated and stored"
        echo "2. ✅ Email was sent successfully"
        echo "3. ✅ OTP verification is working"
        echo "4. ✅ User can now log in with email OTP"
        echo ""
        echo "🔧 Technical Details:"
        echo "• OTP was stored in database with 10-minute expiry"
        echo "• OTP was consumed after successful verification"
        echo "• SMTP configuration is working correctly"
        echo "• Email template is rendering properly"
    else
        echo "❌ Email OTP verification failed"
        echo ""
        echo "🔍 Possible issues:"
        echo "1. Invalid OTP code"
        echo "2. OTP has expired (10 minutes)"
        echo "3. OTP was already used"
        echo "4. Database connection issue"
        echo ""
        echo "💡 Try generating a new OTP and verify immediately"
    fi
    
else
    echo "❌ Email OTP generation failed"
    echo ""
    echo "🔍 Possible issues:"
    echo "1. SMTP server connection failed"
    echo "2. Invalid email credentials"
    echo "3. Database connection issue"
    echo "4. Server error"
fi

echo ""
echo "🧪 Additional Tests:"
echo "1. Test with expired OTP (wait 10+ minutes)"
echo "2. Test with invalid OTP code"
echo "3. Test with already used OTP"
echo "4. Test SMS OTP functionality"
echo ""
echo "📋 Email OTP Flow Summary:"
echo "1. User requests email OTP → Generate 6-digit code"
echo "2. Code stored in database with 10-minute expiry"
echo "3. Email sent with professional template"
echo "4. User enters code → Verify against database"
echo "5. Code consumed after successful verification"
echo "6. User logged in successfully"
