package dto

import (
	"time"

	"github.com/google/uuid"
)

// RegistrationPendingResponse represents the response when registration is pending email verification
type RegistrationPendingResponse struct {
	Message     string    `json:"message"`
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	OTPID       uuid.UUID `json:"otp_id"`
	ExpiresAt   string    `json:"expires_at"`
	ResendAfter string    `json:"resend_after"`
}

// CompleteRegistrationRequest represents the request to complete registration after email verification
type CompleteRegistrationRequest struct {
	UserID  uuid.UUID `json:"user_id" binding:"required"`
	OTPID   uuid.UUID `json:"otp_id" binding:"required"`
	OTPCode string    `json:"otp_code" binding:"required"`
}

// RegistrationCompleteResponse represents the response when registration is completed successfully
type RegistrationCompleteResponse struct {
	Message      string       `json:"message"`
	UserID       uuid.UUID    `json:"user_id"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	IsNewUser    bool         `json:"is_new_user"`
	UserProfile  *UserProfile `json:"user_profile,omitempty"`
}

// RegistrationData represents the temporary registration data stored during email verification
type RegistrationData struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	PhoneNumber     string    `json:"phone_number"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Password        string    `json:"password"`
	Type            string    `json:"type"`
	ReferalType     string    `json:"referal_type"`
	ReferedByCode   string    `json:"refered_by_code"`
	ReferralCode    string    `json:"referral_code"`
	Username        string    `json:"username"`
	City            string    `json:"city"`
	Country         string    `json:"country"`
	State           string    `json:"state"`
	StreetAddress   string    `json:"street_address"`
	PostalCode      string    `json:"postal_code"`
	DateOfBirth     string    `json:"date_of_birth"`
	DefaultCurrency string    `json:"default_currency"`
	KYCStatus       string    `json:"kyc_status"`
	ProfilePicture  string    `json:"profile_picture"`
	AgentRequestID  string    `json:"agent_request_id"`
	Accounts        []Account `json:"accounts"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// Account represents user account information
type Account struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Currency   string  `json:"currency"`
	RealMoney  float64 `json:"real_money"`
	BonusMoney float64 `json:"bonus_money"`
	UpdatedAt  string  `json:"updated_at"`
}
