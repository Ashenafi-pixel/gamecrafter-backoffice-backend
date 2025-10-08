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
}

// TwoFactorDisableRequest represents a request to disable 2FA
type TwoFactorDisableRequest struct {
	Token string `json:"token" binding:"required"`
}

// TwoFactorSettings represents user's 2FA settings
type TwoFactorSettings struct {
	IsEnabled     bool      `json:"is_enabled"`
	EnabledAt     time.Time `json:"enabled_at,omitempty"`
	LastUsedAt    time.Time `json:"last_used_at,omitempty"`
	BackupCodes   []string  `json:"backup_codes,omitempty"`
	HasBackupCodes bool     `json:"has_backup_codes"`
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
