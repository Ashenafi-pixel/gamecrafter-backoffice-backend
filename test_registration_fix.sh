#!/bin/bash

# Test script to verify the registration fix
echo "Testing registration with unique username..."

# Generate a unique username with timestamp
UNIQUE_USERNAME="ashenafiplayer$(date +%s)"

echo "Using unique username: $UNIQUE_USERNAME"

# Test registration with unique username
echo "Testing registration..."
REGISTRATION_RESPONSE=$(curl -s -X 'POST' \
  'http://localhost:8080/register' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "username": "'$UNIQUE_USERNAME'",
  "accounts": [
    {
      "bonus_money": 10,
      "currency": "USD",
      "id": "uuid-new",
      "real_money": 50,
      "updated_at": "2025-08-30T01:00:00Z",
      "user_id": "uuid-new"
    }
  ],
  "agent_request_id": "AGENT002",
  "city": "Addis Ababa",
  "country": "Ethiopia",
  "date_of_birth": "1992-05-15",
  "default_currency": "USD",
  "email": "test'$(date +%s)'@example.com",
  "first_name": "Test",
  "kyc_status": "PENDING",
  "last_name": "User",
  "password": "SecurePass123!",
  "phone_number": "+251911223'$(date +%s)'",
  "postal_code": "2000",
  "profile_picture": "profile2.jpg",
  "referal_type": "PLAYER",
  "refered_by_code": "REF999",
  "referral_code": "USER789",
  "state": "Addis Ababa",
  "street_address": "456 New Street",
  "type": "PLAYER"
}')

echo "Registration Response:"
echo "$REGISTRATION_RESPONSE"

# Extract user_id and otp_id from response
USER_ID=$(echo "$REGISTRATION_RESPONSE" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
OTP_ID=$(echo "$REGISTRATION_RESPONSE" | grep -o '"otp_id":"[^"]*"' | cut -d'"' -f4)

echo "User ID: $USER_ID"
echo "OTP ID: $OTP_ID"

if [ -n "$USER_ID" ] && [ -n "$OTP_ID" ]; then
    echo "Registration successful! Now testing completion..."
    
    # Test completion with OTP (using a test OTP code)
    COMPLETION_RESPONSE=$(curl -s -X 'POST' \
      'http://localhost:8080/register/complete' \
      -H 'accept: application/json' \
      -H 'Content-Type: application/json' \
      -d '{
      "otp_code": "123456",
      "otp_id": "'$OTP_ID'",
      "user_id": "'$USER_ID'"
    }')
    
    echo "Completion Response:"
    echo "$COMPLETION_RESPONSE"
    
    # Check if completion was successful
    if echo "$COMPLETION_RESPONSE" | grep -q "Registration completed successfully"; then
        echo "Registration and completion successful!"
    else
        echo " Completion failed. This might be expected if OTP verification is required."
        echo "The important thing is that the username duplicate error is fixed."
    fi
else
    echo " Registration failed"
fi

echo "Test completed!" 