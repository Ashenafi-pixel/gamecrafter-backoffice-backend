package twofactor

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

type twoFactorService struct {
	storage TwoFactorStorage
	log     *zap.Logger
	config  TwoFactorConfig
}

type TwoFactorConfig struct {
	Issuer           string
	Algorithm        otp.Algorithm
	Digits           otp.Digits
	Period           uint
	BackupCodesCount int
	MaxAttempts      int
	LockoutDuration  time.Duration
}

type TwoFactorService interface {
	// Secret generation and QR code
	GenerateSecret(userID uuid.UUID, email string) (*dto.TwoFactorSecret, error)
	GenerateQRCode(secret, email string) (string, error)
	
	// Token verification
	VerifyToken(secret, token string) bool
	VerifyAndEnable2FA(ctx context.Context, userID uuid.UUID, secret, token string) error
	
	// 2FA management
	Disable2FA(ctx context.Context, userID uuid.UUID, token string) error
	Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error)
	
	// Backup codes
	GenerateBackupCodes() []string
	ValidateBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	
	// Login verification
	VerifyLoginToken(ctx context.Context, userID uuid.UUID, token, backupCode, ip, userAgent string) (bool, error)
	
	// Rate limiting
	IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error)
	LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error
}

func NewTwoFactorService(storage TwoFactorStorage, log *zap.Logger, config TwoFactorConfig) TwoFactorService {
	return &twoFactorService{
		storage: storage,
		log:     log,
		config:  config,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (t *twoFactorService) GenerateSecret(userID uuid.UUID, email string) (*dto.TwoFactorSecret, error) {
	// Generate a random secret key
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.config.Issuer,
		AccountName: email,
		SecretSize:  20,
		Digits:      t.config.Digits,
		Algorithm:   t.config.Algorithm,
		Period:      t.config.Period,
	})
	if err != nil {
		t.log.Error("Failed to generate TOTP secret", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate QR code URL
	qrCodeURL, err := t.GenerateQRCode(secret.Secret(), email)
	if err != nil {
		t.log.Error("Failed to generate QR code", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	t.log.Info("TOTP secret generated successfully", zap.String("user_id", userID.String()))

	return &dto.TwoFactorSecret{
		Secret:    secret.Secret(),
		QRCodeURL: qrCodeURL,
		ManualKey: secret.Secret(),
	}, nil
}

// GenerateQRCode generates a QR code URL for the secret
func (t *twoFactorService) GenerateQRCode(secret, email string) (string, error) {
	// Create the TOTP URL
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("/%s:%s", t.config.Issuer, email),
	}
	
	q := u.Query()
	q.Set("secret", secret)
	q.Set("issuer", t.config.Issuer)
	q.Set("algorithm", string(t.config.Algorithm))
	q.Set("digits", string(t.config.Digits))
	q.Set("period", fmt.Sprintf("%d", t.config.Period))
	
	u.RawQuery = q.Encode()
	
	return u.String(), nil
}

// VerifyToken verifies a TOTP token against a secret
func (t *twoFactorService) VerifyToken(secret, token string) bool {
	valid := totp.Validate(token, secret)
	if !valid {
		t.log.Warn("Invalid TOTP token provided", zap.String("token", token))
		return false
	}
	
	t.log.Debug("TOTP token verified successfully")
	return true
}

// VerifyAndEnable2FA verifies a token and enables 2FA for a user
func (t *twoFactorService) VerifyAndEnable2FA(ctx context.Context, userID uuid.UUID, secret, token string) error {
	// Verify the token
	if !t.VerifyToken(secret, token) {
		t.log.Warn("Invalid token provided for 2FA enable", zap.String("user_id", userID.String()))
		return fmt.Errorf("invalid token")
	}

	// Check rate limiting
	rateLimited, err := t.IsRateLimited(ctx, userID)
	if err != nil {
		t.log.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if rateLimited {
		t.log.Warn("User is rate limited for 2FA enable", zap.String("user_id", userID.String()))
		return fmt.Errorf("too many attempts, please try again later")
	}

	// Save the secret
	err = t.storage.SaveSecret(ctx, userID, secret)
	if err != nil {
		t.log.Error("Failed to save 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save 2FA secret: %w", err)
	}

	// Generate backup codes
	backupCodes := t.GenerateBackupCodes()
	err = t.storage.SaveBackupCodes(ctx, userID, backupCodes)
	if err != nil {
		t.log.Error("Failed to save backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save backup codes: %w", err)
	}

	// Enable 2FA
	err = t.storage.Enable2FA(ctx, userID)
	if err != nil {
		t.log.Error("Failed to enable 2FA", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Log successful attempt
	err = t.LogAttempt(ctx, userID, "setup", true, "", "")
	if err != nil {
		t.log.Error("Failed to log successful setup attempt", zap.Error(err), zap.String("user_id", userID.String()))
	}

	t.log.Info("2FA enabled successfully", zap.String("user_id", userID.String()))
	return nil
}

// Disable2FA disables 2FA for a user
func (t *twoFactorService) Disable2FA(ctx context.Context, userID uuid.UUID, token string) error {
	// Get the user's secret
	secret, err := t.storage.GetSecret(ctx, userID)
	if err != nil {
		t.log.Error("Failed to get user secret for 2FA disable", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to get user secret: %w", err)
	}

	// Verify the token
	if !t.VerifyToken(secret, token) {
		t.log.Warn("Invalid token provided for 2FA disable", zap.String("user_id", userID.String()))
		return fmt.Errorf("invalid token")
	}

	// Check rate limiting
	rateLimited, err := t.IsRateLimited(ctx, userID)
	if err != nil {
		t.log.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if rateLimited {
		t.log.Warn("User is rate limited for 2FA disable", zap.String("user_id", userID.String()))
		return fmt.Errorf("too many attempts, please try again later")
	}

	// Disable 2FA
	err = t.storage.Disable2FA(ctx, userID)
	if err != nil {
		t.log.Error("Failed to disable 2FA", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Delete the secret
	err = t.storage.DeleteSecret(ctx, userID)
	if err != nil {
		t.log.Error("Failed to delete 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete 2FA secret: %w", err)
	}

	// Log successful attempt
	err = t.LogAttempt(ctx, userID, "disable", true, "", "")
	if err != nil {
		t.log.Error("Failed to log successful disable attempt", zap.Error(err), zap.String("user_id", userID.String()))
	}

	t.log.Info("2FA disabled successfully", zap.String("user_id", userID.String()))
	return nil
}

// Get2FAStatus retrieves the 2FA status for a user
func (t *twoFactorService) Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error) {
	status, err := t.storage.Get2FAStatus(ctx, userID)
	if err != nil {
		t.log.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get 2FA status: %w", err)
	}

	return status, nil
}

// GenerateBackupCodes generates secure backup codes
func (t *twoFactorService) GenerateBackupCodes() []string {
	codes := make([]string, t.config.BackupCodesCount)
	for i := 0; i < t.config.BackupCodesCount; i++ {
		// Generate 8 random bytes
		bytes := make([]byte, 8)
		rand.Read(bytes)
		
		// Convert to base32 and take first 8 characters
		encoded := base32.StdEncoding.EncodeToString(bytes)
		codes[i] = strings.ToUpper(encoded[:8])
	}
	return codes
}

// ValidateBackupCode validates and consumes a backup code
func (t *twoFactorService) ValidateBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	valid, err := t.storage.ValidateAndConsumeBackupCode(ctx, userID, code)
	if err != nil {
		t.log.Error("Failed to validate backup code", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to validate backup code: %w", err)
	}

	if valid {
		t.log.Info("Backup code validated successfully", zap.String("user_id", userID.String()))
	} else {
		t.log.Warn("Invalid backup code used", zap.String("user_id", userID.String()))
	}

	return valid, nil
}

// RegenerateBackupCodes generates new backup codes for a user
func (t *twoFactorService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Generate new backup codes
	newCodes := t.GenerateBackupCodes()
	
	// Save the new codes
	err := t.storage.SaveBackupCodes(ctx, userID, newCodes)
	if err != nil {
		t.log.Error("Failed to save new backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to save new backup codes: %w", err)
	}

	t.log.Info("Backup codes regenerated successfully", zap.String("user_id", userID.String()))
	return newCodes, nil
}

// VerifyLoginToken verifies a token during login (TOTP or backup code)
func (t *twoFactorService) VerifyLoginToken(ctx context.Context, userID uuid.UUID, token, backupCode, ip, userAgent string) (bool, error) {
	// Check rate limiting
	rateLimited, err := t.IsRateLimited(ctx, userID)
	if err != nil {
		t.log.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	if rateLimited {
		t.log.Warn("User is rate limited for login", zap.String("user_id", userID.String()))
		return false, fmt.Errorf("too many attempts, please try again later")
	}

	var success bool
	var attemptType string

	// Try backup code first if provided
	if backupCode != "" {
		valid, err := t.ValidateBackupCode(ctx, userID, backupCode)
		if err != nil {
			t.log.Error("Failed to validate backup code", zap.Error(err), zap.String("user_id", userID.String()))
			return false, fmt.Errorf("failed to validate backup code: %w", err)
		}
		if valid {
			success = true
			attemptType = "backup_code"
		}
	}

	// If backup code failed or not provided, try TOTP token
	if !success && token != "" {
		secret, err := t.storage.GetSecret(ctx, userID)
		if err != nil {
			t.log.Error("Failed to get user secret for login", zap.Error(err), zap.String("user_id", userID.String()))
			return false, fmt.Errorf("failed to get user secret: %w", err)
		}

		if t.VerifyToken(secret, token) {
			success = true
			attemptType = "totp"
		}
	}

	// Log the attempt
	err = t.LogAttempt(ctx, userID, attemptType, success, ip, userAgent)
	if err != nil {
		t.log.Error("Failed to log login attempt", zap.Error(err), zap.String("user_id", userID.String()))
	}

	if success {
		t.log.Info("2FA login verification successful", zap.String("user_id", userID.String()), zap.String("method", attemptType))
	} else {
		t.log.Warn("2FA login verification failed", zap.String("user_id", userID.String()))
	}

	return success, nil
}

// IsRateLimited checks if a user is rate limited
func (t *twoFactorService) IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error) {
	return t.storage.IsRateLimited(ctx, userID, t.config.MaxAttempts, int(t.config.LockoutDuration.Minutes()))
}

// LogAttempt logs a 2FA attempt
func (t *twoFactorService) LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error {
	return t.storage.LogAttempt(ctx, userID, attemptType, success, ip, userAgent)
}
