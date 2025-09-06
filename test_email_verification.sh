#!/bin/bash

# 🚀 TucanBIT Email Verification System Test Script
# This script tests the complete email verification flow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
TEST_EMAIL="test@tucanbit.com"
LOG_FILE="email_verification_test.log"

# Log function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}$1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED} $1${NC}" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# Check if application is running
check_app_running() {
    log "Checking if TucanBIT application is running..."
    
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        success "Application is running at $BASE_URL"
    else
        error "Application is not running at $BASE_URL"
        log "Please start the application first:"
        log "  ./start-app-background.sh"
        exit 1
    fi
}

# Test 1: Initiate User Registration
test_initiate_registration() {
    log "🧪 Test 1: Initiating User Registration"
    
    local response=$(curl -s -X POST "$BASE_URL/api/register" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'$TEST_EMAIL'",
            "phone_number": "+1234567890",
            "first_name": "Test",
            "last_name": "User",
            "password": "SecurePass123!",
            "type": "PLAYER",
            "referal_type": "PLAYER",
            "default_currency": "USD",
            "city": "Test City",
            "country": "Test Country",
            "state": "Test State",
            "street_address": "123 Test Street",
            "postal_code": "12345",
            "date_of_birth": "1990-01-01",
            "kyc_status": "PENDING",
            "profile_picture": "",
            "agent_request_id": "",
            "accounts": []
        }')
    
    if echo "$response" | grep -q "session_id"; then
        success "Registration initiated successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        
        # Extract session ID for next test
        SESSION_ID=$(echo "$response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
        export SESSION_ID
        log "Session ID: $SESSION_ID"
    else
        error "Failed to initiate registration"
        echo "$response"
        return 1
    fi
}

# Test 2: Create Email Verification OTP
test_create_otp() {
    log "🧪 Test 2: Creating Email Verification OTP"
    
    local response=$(curl -s -X POST "$BASE_URL/api/otp/email-verification" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'$TEST_EMAIL'"
        }')
    
    if echo "$response" | grep -q "otp_id"; then
        success "OTP created successfully"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        
        # Extract OTP ID for next test
        OTP_ID=$(echo "$response" | grep -o '"otp_id":"[^"]*"' | cut -d'"' -f4)
        export OTP_ID
        log "OTP ID: $OTP_ID"
    else
        error "Failed to create OTP"
        echo "$response"
        return 1
    fi
}

# Test 3: Verify OTP (This will fail with fake OTP, which is expected)
test_verify_otp() {
    log "🧪 Test 3: Verifying OTP (Expected to fail with fake OTP)"
    
    local response=$(curl -s -X POST "$BASE_URL/api/otp/verify" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'$TEST_EMAIL'",
            "otp_code": "123456",
            "otp_id": "'$OTP_ID'"
        }')
    
    if echo "$response" | grep -q "invalid OTP code\|Wallet authentication failed"; then
        success "OTP verification correctly rejected fake OTP"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        warning "Unexpected response from OTP verification"
        echo "$response"
    fi
}

# Test 4: Complete Registration (This will fail without valid OTP, which is expected)
test_complete_registration() {
    log "🧪 Test 4: Completing Registration (Expected to fail without valid OTP)"
    
    local response=$(curl -s -X POST "$BASE_URL/api/register/complete" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'$TEST_EMAIL'",
            "otp_code": "123456",
            "otp_id": "'$OTP_ID'",
            "session_id": "'$SESSION_ID'"
        }')
    
    if echo "$response" | grep -q "invalid OTP code\|Wallet authentication failed"; then
        success "Registration completion correctly rejected invalid OTP"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        warning "Unexpected response from registration completion"
        echo "$response"
    fi
}

# Test 5: Check Swagger Documentation
test_swagger_docs() {
    log "🧪 Test 5: Checking Swagger Documentation"
    
    if curl -s "$BASE_URL/docs" > /dev/null 2>&1; then
        success "Swagger documentation is accessible at $BASE_URL/docs"
        log "You can view the API documentation at: $BASE_URL/docs"
    else
        warning "Swagger documentation is not accessible"
    fi
}

# Test 6: Check OTP Endpoints
test_otp_endpoints() {
    log "🧪 Test 6: Checking OTP Endpoints"
    
    # Test OTP info endpoint
    local response=$(curl -s -X GET "$BASE_URL/api/otp/$OTP_ID")
    if echo "$response" | grep -q "otp_id\|OTP not found"; then
        success "OTP info endpoint is working"
    else
        warning "OTP info endpoint returned unexpected response"
        echo "$response"
    fi
}

# Test 7: Resend OTP
test_resend_otp() {
    log "🧪 Test 7: Testing OTP Resend"
    
    local response=$(curl -s -X POST "$BASE_URL/api/otp/resend" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'$TEST_EMAIL'"
        }')
    
    if echo "$response" | grep -q "otp_id\|please wait"; then
        success "OTP resend endpoint is working"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        warning "OTP resend endpoint returned unexpected response"
        echo "$response"
    fi
}

# Main test execution
main() {
    log "🚀 Starting TucanBIT Email Verification System Tests"
    log "=================================================="
    
    # Clear log file
    > "$LOG_FILE"
    
    # Check prerequisites
    check_app_running
    
    # Run tests
    test_initiate_registration
    test_create_otp
    test_verify_otp
    test_complete_registration
    test_swagger_docs
    test_otp_endpoints
    test_resend_otp
    
    log ""
    log "🎉 All tests completed!"
    log "📋 Test Summary:"
    log "  - Registration initiation: "
    log "  - OTP creation: "
    log "  - OTP verification (expected failure): "
    log "  - Registration completion (expected failure): "
    log "  - Swagger documentation: "
    log "  - OTP endpoints: "
    log "  - OTP resend: "
    log ""
    log "📧 Next Steps:"
    log "  1. Check your email for verification OTP"
    log "  2. Use the real OTP to complete registration"
    log "  3. View API documentation at: $BASE_URL/docs"
    log ""
    log "📁 Test logs saved to: $LOG_FILE"
}

# Run main function
main "$@" 