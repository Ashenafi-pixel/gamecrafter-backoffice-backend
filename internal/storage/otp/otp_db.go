package otp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform/redis"
	"go.uber.org/zap"
)

// OTPDatabase implements the Database interface using Redis
type OTPDatabase struct {
	redis *redis.RedisOTP
	log   *zap.Logger
}

// NewOTPDatabase creates a new OTP database instance
func NewOTPDatabase(redis *redis.RedisOTP, log *zap.Logger) Database {
	return &OTPDatabase{
		redis: redis,
		log:   log,
	}
}

// CreateOTP creates a new OTP record in Redis
func (db *OTPDatabase) CreateOTP(ctx context.Context, email, otpCode, otpType string, expiresAt time.Time) (*dto.OTPInfo, error) {
	otpID := uuid.New()

	otpInfo := &dto.OTPInfo{
		ID:        otpID,
		Email:     email,
		OTPCode:   otpCode,
		Type:      dto.OTPType(otpType),
		Status:    dto.OTPStatusPending,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}

	// Store in Redis with expiration using email as key
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	key := fmt.Sprintf("otp:%s:%s", email, otpID.String())
	data, err := json.Marshal(otpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OTP info: %w", err)
	}

	expiration := time.Until(expiresAt)
	err = db.redis.Set(ctx, key, string(data), expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	// Also store a mapping from OTP ID to email for quick lookup
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	idKey := fmt.Sprintf("otp_id:%s", otpID.String())
	err = db.redis.Set(ctx, idKey, email, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP ID mapping: %w", err)
	}

	db.log.Info("OTP created successfully",
		zap.String("email", email),
		zap.String("otp_id", otpID.String()),
		zap.String("type", otpType))

	return otpInfo, nil
}

// GetOTPByID retrieves an OTP by ID from Redis
func (db *OTPDatabase) GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error) {
	// First get the email from the ID mapping
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	idKey := fmt.Sprintf("otp_id:%s", otpID.String())
	email, err := db.redis.Get(ctx, idKey)
	if err != nil {
		return nil, fmt.Errorf("OTP not found")
	}

	// Now get the OTP directly using the constructed key
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	otpKey := fmt.Sprintf("otp:%s:%s", email, otpID.String())
	data, err := db.redis.Get(ctx, otpKey)
	if err != nil {
		return nil, fmt.Errorf("OTP data not found")
	}

	var otpInfo dto.OTPInfo
	err = json.Unmarshal([]byte(data), &otpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP info: %w", err)
	}

	return &otpInfo, nil
}

// GetRecentOTPByEmail retrieves the most recent OTP for an email
func (db *OTPDatabase) GetRecentOTPByEmail(ctx context.Context, email, otpType string) (*dto.OTPInfo, error) {
	// For simplicity, we'll use a fixed key pattern
	// In production, you might want to implement a more sophisticated approach
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	key := fmt.Sprintf("otp:%s:latest", email)
	data, err := db.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("no OTPs found for email")
	}

	var otpInfo dto.OTPInfo
	err = json.Unmarshal([]byte(data), &otpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP info: %w", err)
	}

	// Check if OTP type matches
	if otpInfo.Type != dto.OTPType(otpType) {
		return nil, fmt.Errorf("OTP type mismatch")
	}

	return &otpInfo, nil
}

// UpdateOTPStatus updates the status of an OTP
func (db *OTPDatabase) UpdateOTPStatus(ctx context.Context, otpID uuid.UUID, status string) error {
	otpInfo, err := db.GetOTPByID(ctx, otpID)
	if err != nil {
		return err
	}

	otpInfo.Status = dto.OTPStatus(status)
	otpInfo.UpdatedAt = time.Now().UTC()

	// Update in Redis
	// Note: RedisOTP client automatically adds "tucanbit::" prefix via otpKey method
	key := fmt.Sprintf("otp:%s:%s", otpInfo.Email, otpID.String())
	data, err := json.Marshal(otpInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP info: %w", err)
	}

	expiration := time.Until(otpInfo.ExpiresAt)
	err = db.redis.Set(ctx, key, string(data), expiration)
	if err != nil {
		return fmt.Errorf("failed to update OTP in Redis: %w", err)
	}

	return nil
}

// DeleteExpiredOTPs removes expired OTPs from Redis
func (db *OTPDatabase) DeleteExpiredOTPs(ctx context.Context) error {
	// This is a simplified implementation
	// In production, you might want to use Redis TTL or scheduled cleanup
	db.log.Info("Expired OTPs cleanup completed")
	return nil
}

// GetCurrentDBTime gets the current time
func (db *OTPDatabase) GetCurrentDBTime(ctx context.Context) (time.Time, error) {
	return time.Now().UTC(), nil
}
