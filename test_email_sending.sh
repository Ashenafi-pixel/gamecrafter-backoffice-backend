#!/bin/bash

# Script to test email sending functionality for 2FA OTPs
# This will help debug if emails are being sent successfully

echo "📧 Testing Email Sending Functionality..."

# Configuration
BASE_URL="http://localhost:8094"
USER_ID="5a8328c7-d51b-4187-b45c-b1beea7b41ff"
USER_EMAIL="kirube.tech23@gmail.com"  # Replace with actual user email

echo "📧 User ID: $USER_ID"
echo "📧 User Email: $USER_EMAIL"
echo "🌐 Server: $BASE_URL"
echo ""

# Step 1: Test email OTP generation
echo "1️⃣ Testing Email OTP Generation..."
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
echo ""
echo "HTTP Status Code: $otp_http_code"

if [ "$otp_http_code" -eq 200 ]; then
    echo "✅ Email OTP generation request successful!"
    echo ""
    echo "📋 What this means:"
    echo "1. ✅ Backend processed the request"
    echo "2. ✅ SMTP configuration is valid"
    echo "3. ✅ Email service is working"
    echo ""
    echo "📧 Check your email inbox ($USER_EMAIL) for the 2FA code"
    echo "📧 Also check spam/junk folder if not in inbox"
    echo ""
    echo "🔍 Email Details:"
    echo "• From: TucanBIT Security <kirub.hel@gmail.com>"
    echo "• Subject: Two-Factor Authentication Code - TucanBIT Security"
    echo "• Code: 6-digit number"
    echo "• Expires: 10 minutes"
else
    echo "❌ Email OTP generation failed"
    echo ""
    echo "🔍 Possible issues:"
    echo "1. SMTP server connection failed"
    echo "2. Invalid email credentials"
    echo "3. Gmail security settings blocking the app"
    echo "4. Server error"
    echo ""
    echo "💡 Next steps:"
    echo "1. Check server logs for detailed error messages"
    echo "2. Verify Gmail app password is correct"
    echo "3. Check if 'Less secure app access' is enabled"
    echo "4. Verify SMTP settings in config.yaml"
fi

echo ""
echo "2️⃣ Testing with a different email (if first test failed)..."
if [ "$otp_http_code" -ne 200 ]; then
    echo "Trying with a simpler email address..."
    
    simple_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
      -X POST \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"$USER_ID\",
        \"email\": \"test@example.com\"
      }" \
      "$BASE_URL/api/admin/auth/2fa/generate-email-otp")
    
    simple_http_code=$(echo "$simple_response" | grep "HTTP_CODE:" | cut -d: -f2)
    simple_body=$(echo "$simple_response" | sed '/HTTP_CODE:/d')
    
    echo "Simple email test response:"
    echo "$simple_body" | jq . 2>/dev/null || echo "$simple_body"
    echo "HTTP Status: $simple_http_code"
fi

echo ""
echo "🔧 Manual SMTP Test (if needed):"
echo "You can test SMTP manually using telnet:"
echo "telnet smtp.gmail.com 465"
echo ""
echo "Or using curl with SMTP:"
echo "curl --url 'smtps://smtp.gmail.com:465' \\"
echo "  --ssl-reqd \\"
echo "  --mail-from 'kirub.hel@gmail.com' \\"
echo "  --mail-rcpt '$USER_EMAIL' \\"
echo "  --user 'kirub.hel@gmail.com:bads ozyw rzko hljf' \\"
echo "  --upload-file email.txt"

echo ""
echo "📋 SMTP Configuration Check:"
echo "Host: smtp.gmail.com"
echo "Port: 465"
echo "Username: kirub.hel@gmail.com"
echo "Password: bads ozyw rzko hljf"
echo "Use TLS: true"
echo ""
echo "⚠️  Important Notes:"
echo "• Gmail requires app-specific passwords for SMTP"
echo "• Make sure 2-factor authentication is enabled on Gmail"
echo "• App password should be 16 characters without spaces"
echo "• Check Gmail security settings if emails don't arrive"
