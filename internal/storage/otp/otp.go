package otp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
)

// OTP defines the interface for OTP storage operations
type OTP interface {
	CreateOTP(ctx context.Context, email, otpCode, otpType string, expiresAt time.Time) (*dto.OTPInfo, error)
	GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error)
	GetRecentOTPByEmail(ctx context.Context, email, otpType string) (*dto.OTPInfo, error)
	UpdateOTPStatus(ctx context.Context, otpID uuid.UUID, status string) error
	DeleteExpiredOTPs(ctx context.Context) error
	GetCurrentDBTime(ctx context.Context) (time.Time, error)
}

// otp implements the OTP interface
type otp struct {
	db Database
}

// Database defines the interface for database operations
type Database interface {
	CreateOTP(ctx context.Context, email, otpCode, otpType string, expiresAt time.Time) (*dto.OTPInfo, error)
	GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error)
	GetRecentOTPByEmail(ctx context.Context, email, otpType string) (*dto.OTPInfo, error)
	UpdateOTPStatus(ctx context.Context, otpID uuid.UUID, status string) error
	DeleteExpiredOTPs(ctx context.Context) error
	GetCurrentDBTime(ctx context.Context) (time.Time, error)
}

// NewOTP creates a new instance of OTP storage
func NewOTP(db Database) OTP {
	return &otp{
		db: db,
	}
}

// CreateOTP creates a new OTP record
func (o *otp) CreateOTP(ctx context.Context, email, otpCode, otpType string, expiresAt time.Time) (*dto.OTPInfo, error) {
	return o.db.CreateOTP(ctx, email, otpCode, otpType, expiresAt)
}

// GetOTPByID retrieves an OTP by ID
func (o *otp) GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error) {
	return o.db.GetOTPByID(ctx, otpID)
}

// GetRecentOTPByEmail retrieves the most recent OTP for an email
func (o *otp) GetRecentOTPByEmail(ctx context.Context, email, otpType string) (*dto.OTPInfo, error) {
	return o.db.GetRecentOTPByEmail(ctx, email, otpType)
}

// UpdateOTPStatus updates the status of an OTP
func (o *otp) UpdateOTPStatus(ctx context.Context, otpID uuid.UUID, status string) error {
	return o.db.UpdateOTPStatus(ctx, otpID, status)
}

// DeleteExpiredOTPs removes expired OTPs from the database
func (o *otp) DeleteExpiredOTPs(ctx context.Context) error {
	return o.db.DeleteExpiredOTPs(ctx)
}

// GetCurrentDBTime gets the current database time
func (o *otp) GetCurrentDBTime(ctx context.Context) (time.Time, error) {
	return o.db.GetCurrentDBTime(ctx)
}
