package dto

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorSecret represents a generated 2FA secret
type TwoFactorSecret struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
	ManualKey string `json:"manual_key"`
}

// TwoFactorSetupRequest represents a request to setup 2FA
type TwoFactorSetupRequest struct {
	Token string `json:"token" binding:"required"`
}

// TwoFactorVerifyRequest represents a request to verify 2FA token
type TwoFactorVerifyRequest struct {
	Token      string `json:"token" binding:"required"`
	BackupCode string `json:"backup_code,omitempty"`
	Method     string `json:"method,omitempty"` // Method can be: totp, email_otp, sms_otp, backup_codes
}

// TwoFactorDisableRequest represents a request to disable 2FA
type TwoFactorDisableRequest struct {
	Token string `json:"token" binding:"required"`
}

// TwoFactorAuthSetupResponse represents the response for 2FA setup
type TwoFactorAuthSetupResponse struct {
	Secret     string `json:"secret"`
	QRCodeData string `json:"qr_code_data"` // Base64 encoded image or data URI
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// TwoFactorAuthEnableResponse represents the response for enabling 2FA
type TwoFactorAuthEnableResponse struct {
	Message string `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

// TwoFactorAuthStatusResponse represents the current 2FA status for a user
type TwoFactorAuthStatusResponse struct {
	IsEnabled bool      `json:"is_enabled"`
	SetupAt   *time.Time `json:"setup_at,omitempty"`
}

// TwoFactorBackupCodesResponse represents backup codes response
type TwoFactorBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// TwoFactorVerifyResponse represents the response for 2FA verification
type TwoFactorVerifyResponse struct {
	Message string `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

// TwoFactorSettings represents user's 2FA settings
type TwoFactorSettings struct {
	UserID    uuid.UUID `json:"user_id"`
	IsEnabled bool      `json:"is_enabled"`
	EnabledAt *time.Time `json:"enabled_at,omitempty"`
}

// TwoFactorAttempt represents a 2FA attempt log
type TwoFactorAttempt struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	AttemptType  string    `json:"attempt_type"`
	IsSuccessful bool      `json:"is_successful"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}

// TwoFactorResponse represents a generic 2FA API response
type TwoFactorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BackupCodesResponse represents backup codes response
type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
	Warning     string   `json:"warning"`
}

// TwoFactorStatus represents the current 2FA status
type TwoFactorStatus struct {
	IsEnabled       bool `json:"is_enabled"`
	IsSetupComplete bool `json:"is_setup_complete"`
	HasBackupCodes  bool `json:"has_backup_codes"`
}
