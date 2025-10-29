package dto

import (
	"time"

	"github.com/google/uuid"
)

// EnterpriseRegistrationRequest represents the initial registration request
type EnterpriseRegistrationRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FirstName    string `json:"first_name" validate:"required"`
	LastName     string `json:"last_name" validate:"required"`
	PhoneNumber  string `json:"phone_number,omitempty"`
	UserType     string `json:"user_type" validate:"required,oneof=PLAYER AGENT ADMIN"`
	CompanyName  string `json:"company_name,omitempty"`
	ReferralCode string `json:"referral_code,omitempty"`
}

// EnterpriseRegistration represents the database model for enterprise registration
type EnterpriseRegistration struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	UserID                  uuid.UUID              `json:"user_id" db:"user_id"`
	Email                   string                 `json:"email" db:"email"`
	FirstName               string                 `json:"first_name" db:"first_name"`
	LastName                string                 `json:"last_name" db:"last_name"`
	UserType                string                 `json:"user_type" db:"user_type"`
	PhoneNumber             *string                `json:"phone_number" db:"phone_number"`
	CompanyName             *string                `json:"company_name" db:"company_name"`
	RegistrationStatus      string                 `json:"registration_status" db:"registration_status"`
	VerificationOTP         *string                `json:"verification_otp" db:"verification_otp"`
	OTPExpiresAt            *time.Time             `json:"otp_expires_at" db:"otp_expires_at"`
	VerificationAttempts    int                    `json:"verification_attempts" db:"verification_attempts"`
	MaxVerificationAttempts int                    `json:"max_verification_attempts" db:"max_verification_attempts"`
	EmailVerifiedAt         *time.Time             `json:"email_verified_at" db:"email_verified_at"`
	PhoneVerifiedAt         *time.Time             `json:"phone_verified_at" db:"phone_verified_at"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at" db:"updated_at"`
	VerifiedAt              *time.Time             `json:"verified_at" db:"verified_at"`
	RejectedAt              *time.Time             `json:"rejected_at" db:"rejected_at"`
	RejectionReason         *string                `json:"rejection_reason" db:"rejection_reason"`
	Metadata                map[string]interface{} `json:"metadata" db:"metadata"`
}

// EnterpriseRegistrationStats represents registration statistics
type EnterpriseRegistrationStats struct {
	StatusCounts map[string]map[string]int `json:"status_counts"`
	TotalCount   int                       `json:"total_count"`
}

// EnterpriseRegistrationResponse represents the registration response
type EnterpriseRegistrationResponse struct {
	Message     string    `json:"message"`
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	OTPID       uuid.UUID `json:"otp_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	ResendAfter time.Time `json:"resend_after"`
}

// EnterpriseRegistrationCompleteRequest represents the registration completion request
type EnterpriseRegistrationCompleteRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	OTPID   uuid.UUID `json:"otp_id" validate:"required"`
	OTPCode string    `json:"otp_code" validate:"required,min=6,max=6"`
}

// EnterpriseRegistrationCompleteResponse represents the registration completion response
type EnterpriseRegistrationCompleteResponse struct {
	Message      string    `json:"message"`
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	IsVerified   bool      `json:"is_verified"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	VerifiedAt   time.Time `json:"verified_at"`
}

// EnterpriseRegistrationStatus represents the registration status
type EnterpriseRegistrationStatus struct {
	UserID       uuid.UUID  `json:"user_id"`
	Email        string     `json:"email"`
	Status       string     `json:"status"` // pending, verified, completed
	CreatedAt    time.Time  `json:"created_at"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	OTPExpiresAt time.Time  `json:"otp_expires_at"`
}

// ResendVerificationEmailRequest represents the resend verification email request
type ResendVerificationEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
