package dto

import (
	"time"

	"github.com/google/uuid"
)

// OTPType represents the type of OTP
type OTPType string

const (
	OTPTypeEmailVerification OTPType = "email_verification"
	OTPTypePasswordReset     OTPType = "password_reset"
	OTPTypeLogin             OTPType = "login"
)

// OTPStatus represents the status of an OTP
type OTPStatus string

const (
	OTPStatusPending  OTPStatus = "pending"
	OTPStatusVerified OTPStatus = "verified"
	OTPStatusExpired  OTPStatus = "expired"
	OTPStatusUsed     OTPStatus = "used"
)

// EmailVerificationRequest represents email verification request
type EmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// EmailVerificationResponse represents email verification response
type EmailVerificationResponse struct {
	Message     string    `json:"message"`
	OTPID       uuid.UUID `json:"otp_id"`
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

// OTPVerificationRequest represents OTP verification request
type OTPVerificationRequest struct {
	OTPID   uuid.UUID `json:"otp_id" validate:"required"`
	OTPCode string    `json:"otp_code" validate:"required,min=6,max=6"`
	Email   string    `json:"email" validate:"required,email"`
}

// OTPVerificationResponse represents OTP verification response
type OTPVerificationResponse struct {
	Message      string    `json:"message"`
	IsVerified   bool      `json:"is_verified"`
	UserID       uuid.UUID `json:"user_id,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	VerifiedAt   time.Time `json:"verified_at,omitempty"`
}

// ResendOTPRequest represents resend OTP request
type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendOTPResponse represents resend OTP response
type ResendOTPResponse struct {
	Message     string    `json:"message"`
	OTPID       uuid.UUID `json:"otp_id"`
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

// OTPInfo represents OTP information
type OTPInfo struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	OTPCode    string     `json:"otp_code"`
	Type       OTPType    `json:"type"`
	Status     OTPStatus  `json:"status"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// OTPStatsResponse represents OTP statistics
type OTPStatsResponse struct {
	TotalOTPs           int     `json:"total_otps"`
	VerifiedOTPs        int     `json:"verified_otps"`
	ExpiredOTPs         int     `json:"expired_otps"`
	FailedAttempts      int     `json:"failed_attempts"`
	SuccessRate         float64 `json:"success_rate"`
	AverageResponseTime float64 `json:"average_response_time"`
	Period              int     `json:"period"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}
