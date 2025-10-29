#!/bin/bash

# Script to test Gmail SMTP connection and authentication
# This will help identify the exact issue with email delivery

echo "🔍 Testing Gmail SMTP Connection..."

# Gmail SMTP Configuration
SMTP_HOST="smtp.gmail.com"
SMTP_PORT="465"
SMTP_USER="kirub.hel@gmail.com"
SMTP_PASS="dacc uhlb etak tpoo"
TEST_EMAIL="kirube.tech23@gmail.com"

echo "📧 SMTP Configuration:"
echo "Host: $SMTP_HOST"
echo "Port: $SMTP_PORT"
echo "Username: $SMTP_USER"
echo "Password: $SMTP_PASS (length: ${#SMTP_PASS})"
echo "Test Email: $TEST_EMAIL"
echo ""

# Test 1: Check if port 465 is accessible
echo "1️⃣ Testing SMTP Port Connectivity..."
if timeout 10 bash -c "</dev/tcp/$SMTP_HOST/$SMTP_PORT" 2>/dev/null; then
    echo "✅ Port $SMTP_PORT is accessible on $SMTP_HOST"
else
    echo "❌ Cannot connect to $SMTP_HOST:$SMTP_PORT"
    echo "💡 Check your internet connection and firewall settings"
    exit 1
fi

echo ""

# Test 2: Test SMTP authentication using curl
echo "2️⃣ Testing SMTP Authentication..."
echo "Creating test email content..."

cat > /tmp/test_email.txt << EOF
From: TucanBIT Security <$SMTP_USER>
To: $TEST_EMAIL
Subject: Test Email from TucanBIT
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

<html>
<body>
<h2>Test Email</h2>
<p>This is a test email to verify SMTP configuration.</p>
<p>If you receive this, the email system is working correctly.</p>
<p>Timestamp: $(date)</p>
</body>
</html>
EOF

echo "Sending test email via SMTP..."

# Use curl to test SMTP
smtp_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  --url "smtps://$SMTP_HOST:$SMTP_PORT" \
  --ssl-reqd \
  --mail-from "$SMTP_USER" \
  --mail-rcpt "$TEST_EMAIL" \
  --user "$SMTP_USER:$SMTP_PASS" \
  --upload-file /tmp/test_email.txt \
  --connect-timeout 30 \
  --max-time 60)

smtp_http_code=$(echo "$smtp_response" | grep "HTTP_CODE:" | cut -d: -f2)
smtp_body=$(echo "$smtp_response" | sed '/HTTP_CODE:/d')

echo "SMTP Response:"
echo "$smtp_body"
echo "HTTP Code: $smtp_http_code"

if [ "$smtp_http_code" -eq 0 ] && [ -z "$smtp_body" ]; then
    echo "✅ SMTP authentication successful!"
    echo "📧 Test email sent to $TEST_EMAIL"
    echo "📧 Check your inbox and spam folder"
else
    echo "❌ SMTP authentication failed"
    echo ""
    echo "🔍 Common Gmail SMTP Issues:"
    echo "1. App Password Required:"
    echo "   • Go to Google Account settings"
    echo "   • Security → 2-Step Verification → App passwords"
    echo "   • Generate a new app password for 'Mail'"
    echo "   • Use the 16-character password (no spaces)"
    echo ""
    echo "2. Enable 2-Factor Authentication:"
    echo "   • Gmail account must have 2FA enabled"
    echo "   • App passwords only work with 2FA enabled"
    echo ""
    echo "3. Check Gmail Security Settings:"
    echo "   • Go to myaccount.google.com"
    echo "   • Security → Less secure app access"
    echo "   • Make sure it's configured correctly"
    echo ""
    echo "4. Verify App Password Format:"
    echo "   • Should be 16 characters"
    echo "   • No spaces or special characters"
    echo "   • Generated specifically for 'Mail' app"
fi

echo ""

# Test 3: Alternative SMTP settings
echo "3️⃣ Testing Alternative SMTP Settings..."
echo "Trying port 587 with STARTTLS..."

smtp_587_response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  --url "smtp://$SMTP_HOST:587" \
  --mail-from "$SMTP_USER" \
  --mail-rcpt "$TEST_EMAIL" \
  --user "$SMTP_USER:$SMTP_PASS" \
  --upload-file /tmp/test_email.txt \
  --connect-timeout 30 \
  --max-time 60)

smtp_587_http_code=$(echo "$smtp_587_response" | grep "HTTP_CODE:" | cut -d: -f2)
smtp_587_body=$(echo "$smtp_587_response" | sed '/HTTP_CODE:/d')

echo "SMTP 587 Response:"
echo "$smtp_587_body"
echo "HTTP Code: $smtp_587_http_code"

if [ "$smtp_587_http_code" -eq 0 ] && [ -z "$smtp_587_body" ]; then
    echo "✅ SMTP 587 (STARTTLS) works!"
    echo "💡 Consider updating config.yaml to use port 587"
else
    echo "❌ SMTP 587 also failed"
fi

echo ""

# Cleanup
rm -f /tmp/test_email.txt

echo "🔧 Recommended Actions:"
echo "1. Generate a new Gmail App Password:"
echo "   • Go to https://myaccount.google.com/security"
echo "   • Enable 2-Step Verification if not already enabled"
echo "   • Go to App passwords → Generate password for 'Mail'"
echo "   • Copy the 16-character password"
echo ""
echo "2. Update config.yaml with new password:"
echo "   smtp:"
echo "     host: \"smtp.gmail.com\""
echo "     port: 465"
echo "     username: \"kirub.hel@gmail.com\""
echo "     password: \"YOUR_NEW_APP_PASSWORD\""
echo "     from: \"kirub.hel@gmail.com\""
echo "     from_name: \"TucanBIT Security\""
echo "     use_tls: true"
echo ""
echo "3. Restart the backend server after updating config"
echo ""
echo "4. Test again with the new app password"
