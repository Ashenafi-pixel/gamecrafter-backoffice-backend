#!/bin/bash

# Test script for password reset confirmation email functionality
# This script tests the complete password reset flow including the confirmation email

echo "🧪 Testing Password Reset Confirmation Email Functionality"
echo "=========================================================="

# Configuration
API_BASE_URL="http://localhost:8080"
TEST_EMAIL="test@example.com"
TEST_PASSWORD="NewSecurePassword123!"

echo "📧 Test Email: $TEST_EMAIL"
echo "🔐 Test Password: $TEST_PASSWORD"
echo ""

# Step 1: Request password reset
echo "Step 1: Requesting password reset..."
RESET_REQUEST=$(curl -s -X POST "$API_BASE_URL/api/user/password/forget" \
  -H "Content-Type: application/json" \
  -d "{\"username_or_phone_or_email\": \"$TEST_EMAIL\"}")

echo "Reset request response: $RESET_REQUEST"
echo ""

# Extract OTP ID from response (assuming JSON response format)
OTP_ID=$(echo "$RESET_REQUEST" | grep -o '"otp_id":"[^"]*"' | cut -d'"' -f4)
if [ -z "$OTP_ID" ]; then
    echo "❌ Failed to extract OTP ID from reset request response"
    exit 1
fi

echo "✅ OTP ID extracted: $OTP_ID"
echo ""

# Step 2: Verify OTP (you'll need to get the actual OTP from email/logs)
echo "Step 2: Verifying OTP..."
echo "⚠️  Note: You need to check the email or logs for the actual OTP code"
echo "Please enter the OTP code received:"
read -p "OTP Code: " OTP_CODE

if [ -z "$OTP_CODE" ]; then
    echo "❌ OTP code is required"
    exit 1
fi

VERIFY_RESPONSE=$(curl -s -X POST "$API_BASE_URL/api/user/password/forget/verify" \
  -H "Content-Type: application/json" \
  -d "{
    \"email_or_phone_or_username\": \"$TEST_EMAIL\",
    \"otp\": \"$OTP_CODE\",
    \"otp_id\": \"$OTP_ID\"
  }")

echo "Verify response: $VERIFY_RESPONSE"
echo ""

# Extract token from verify response
TOKEN=$(echo "$VERIFY_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "❌ Failed to extract token from verify response"
    exit 1
fi

echo "✅ Token extracted: ${TOKEN:0:20}..."
echo ""

# Step 3: Reset password with new password
echo "Step 3: Resetting password..."
RESET_RESPONSE=$(curl -s -X POST "$API_BASE_URL/api/user/password/reset" \
  -H "Content-Type: application/json" \
  -H "User-Agent: TestClient/1.0" \
  -d "{
    \"token\": \"$TOKEN\",
    \"new_password\": \"$TEST_PASSWORD\",
    \"confirm_password\": \"$TEST_PASSWORD\"
  }")

echo "Reset password response: $RESET_RESPONSE"
echo ""

# Check if reset was successful
if echo "$RESET_RESPONSE" | grep -q "success\|Success\|200"; then
    echo "✅ Password reset successful!"
    echo ""
    echo "📧 Password reset confirmation email should have been sent to: $TEST_EMAIL"
    echo ""
    echo "🔍 Check the following:"
    echo "   1. Email inbox for confirmation email"
    echo "   2. Application logs for email sending status"
    echo "   3. SMTP server logs if configured"
    echo ""
    echo "📋 Email should contain:"
    echo "   - Professional design with TucanBIT branding"
    echo "   - Confirmation of successful password reset"
    echo "   - Security information (device, location, IP)"
    echo "   - Security reminders and best practices"
    echo "   - Login link to access the account"
    echo ""
else
    echo "❌ Password reset failed"
    echo "Response: $RESET_RESPONSE"
    exit 1
fi

echo "🎉 Test completed successfully!"
echo ""
echo "📝 Next steps:"
echo "   1. Verify the confirmation email was received"
echo "   2. Check email content and design"
echo "   3. Test login with new password"
echo "   4. Review application logs for any errors"