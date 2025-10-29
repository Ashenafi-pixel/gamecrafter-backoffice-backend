package otp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// RedisOTP implements OTP storage using Redis
type RedisOTP struct {
	client RedisClient
	logger *zap.Logger
}

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// NewRedisOTP creates a new instance of RedisOTP
func NewRedisOTP(client RedisClient, logger *zap.Logger) *RedisOTP {
	return &RedisOTP{
		client: client,
		logger: logger,
	}
}

// CreateOTP creates a new OTP record in Redis
func (r *RedisOTP) CreateOTP(ctx context.Context, email, otpCode, otpType string, expiresAt time.Time) (*dto.OTPInfo, error) {
	otpID := uuid.New()
	now := time.Now().UTC()

	otpInfo := &dto.OTPInfo{
		ID:        otpID,
		Email:     email,
		OTPCode:   otpCode,
		Type:      dto.OTPType(otpType),
		Status:    dto.OTPStatusPending,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Store OTP in Redis with expiration
	key := fmt.Sprintf("otp:%s", otpID.String())
	otpData, err := json.Marshal(otpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OTP data: %w", err)
	}

	// Calculate TTL based on expiration time
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil, fmt.Errorf("OTP expiration time is in the past")
	}

	err = r.client.Set(ctx, key, string(otpData), ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	// Also store by email for quick lookups
	emailKey := fmt.Sprintf("otp:email:%s:%s", email, otpType)
	err = r.client.Set(ctx, emailKey, otpID.String(), ttl)
	if err != nil {
		// Clean up main OTP record if email index fails
		_ = r.client.Delete(ctx, key)
		return nil, fmt.Errorf("failed to store OTP email index: %w", err)
	}

	r.logger.Info("OTP created successfully in Redis",
		zap.String("otp_id", otpID.String()),
		zap.String("email", email),
		zap.String("type", otpType),
		zap.Time("expires_at", expiresAt))

	return otpInfo, nil
}

// GetOTPByID retrieves an OTP by ID from Redis
func (r *RedisOTP) GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error) {
	key := fmt.Sprintf("otp:%s", otpID.String())

	otpData, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP from Redis: %w", err)
	}

	var otpInfo dto.OTPInfo
	err = json.Unmarshal([]byte(otpData), &otpInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
	}

	return &otpInfo, nil
}

// GetRecentOTPByEmail retrieves the most recent OTP for an email and type
func (r *RedisOTP) GetRecentOTPByEmail(ctx context.Context, email, otpType string) (*dto.OTPInfo, error) {
	emailKey := fmt.Sprintf("otp:email:%s:%s", email, otpType)

	otpIDStr, err := r.client.Get(ctx, emailKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP ID from email index: %w", err)
	}

	otpID, err := uuid.Parse(otpIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OTP ID: %w", err)
	}

	return r.GetOTPByID(ctx, otpID)
}

// UpdateOTPStatus updates the status of an OTP
func (r *RedisOTP) UpdateOTPStatus(ctx context.Context, otpID uuid.UUID, status string) error {
	otpInfo, err := r.GetOTPByID(ctx, otpID)
	if err != nil {
		return fmt.Errorf("failed to get OTP for status update: %w", err)
	}

	// Update status and timestamp
	otpInfo.Status = dto.OTPStatus(status)
	otpInfo.UpdatedAt = time.Now().UTC()

	if status == string(dto.OTPStatusVerified) {
		now := time.Now().UTC()
		otpInfo.VerifiedAt = &now
	}

	// Re-store updated OTP
	key := fmt.Sprintf("otp:%s", otpID.String())
	otpData, err := json.Marshal(otpInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated OTP data: %w", err)
	}

	// Calculate remaining TTL
	ttl := time.Until(otpInfo.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default TTL for verified/used OTPs
	}

	err = r.client.Set(ctx, key, string(otpData), ttl)
	if err != nil {
		return fmt.Errorf("failed to update OTP in Redis: %w", err)
	}

	r.logger.Info("OTP status updated successfully",
		zap.String("otp_id", otpID.String()),
		zap.String("email", otpInfo.Email),
		zap.String("status", status))

	return nil
}

// DeleteExpiredOTPs removes expired OTPs from Redis
func (r *RedisOTP) DeleteExpiredOTPs(ctx context.Context) error {
	// Get all OTP keys
	keys, err := r.client.Keys(ctx, "otp:*")
	if err != nil {
		return fmt.Errorf("failed to get OTP keys: %w", err)
	}

	deletedCount := 0
	for _, key := range keys {
		// Skip email index keys
		if len(key) > 4 && key[:4] == "otp:" && key[4:10] != "email" {
			otpData, err := r.client.Get(ctx, key)
			if err != nil {
				continue // Key might have expired
			}

			var otpInfo dto.OTPInfo
			if err := json.Unmarshal([]byte(otpData), &otpInfo); err != nil {
				continue // Skip malformed data
			}

			// Check if OTP is expired
			if time.Now().UTC().After(otpInfo.ExpiresAt) {
				// Delete main OTP record
				_ = r.client.Delete(ctx, key)

				// Delete email index
				emailKey := fmt.Sprintf("otp:email:%s:%s", otpInfo.Email, otpInfo.Type)
				_ = r.client.Delete(ctx, emailKey)

				deletedCount++
			}
		}
	}

	r.logger.Info("Expired OTPs cleanup completed",
		zap.Int("deleted_count", deletedCount))

	return nil
}

// GetCurrentDBTime gets the current time (Redis doesn't have a concept of DB time, so we use system time)
func (r *RedisOTP) GetCurrentDBTime(ctx context.Context) (time.Time, error) {
	// For Redis, we use system time since Redis doesn't have its own time concept
	// In production, you might want to use a centralized time service
	return time.Now().UTC(), nil
}
