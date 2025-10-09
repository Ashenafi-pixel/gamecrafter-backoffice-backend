package twofactor

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	mathrand "math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// TwoFactorMethod represents different 2FA methods
type TwoFactorMethod string

const (
	MethodTOTP        TwoFactorMethod = "totp"         // Authenticator apps
	MethodEmailOTP    TwoFactorMethod = "email_otp"    // Email OTP
	MethodSMSOTP      TwoFactorMethod = "sms_otp"      // SMS OTP
	MethodBiometric   TwoFactorMethod = "biometric"    // Biometric (WebAuthn)
	MethodBackupCodes TwoFactorMethod = "backup_codes" // Backup codes
)

// TwoFactorStorage interface for 2FA storage operations
type TwoFactorStorage interface {
	// Secret management
	SaveSecret(ctx context.Context, userID uuid.UUID, secret string) error
	GetSecret(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteSecret(ctx context.Context, userID uuid.UUID) error

	// Backup codes management
	SaveBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error
	GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	DeleteBackupCodes(ctx context.Context, userID uuid.UUID) error
	UseBackupCode(ctx context.Context, userID uuid.UUID, code string) error

	// Settings management
	SaveSettings(ctx context.Context, settings *dto.TwoFactorSettings) error
	GetSettings(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error)
	UpdateSettings(ctx context.Context, settings *dto.TwoFactorSettings) error
	DeleteSettings(ctx context.Context, userID uuid.UUID) error

	// Attempts logging
	LogAttempt(ctx context.Context, attempt *dto.TwoFactorAttempt) error
	GetAttempts(ctx context.Context, userID uuid.UUID, limit int) ([]dto.TwoFactorAttempt, error)
	IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error)

	// OTP management for multiple methods
	SaveOTP(ctx context.Context, userID uuid.UUID, method, otp string, expiry time.Duration) error
	VerifyOTP(ctx context.Context, userID uuid.UUID, method, otp string) (bool, error)
	DeleteOTP(ctx context.Context, userID uuid.UUID, method string) error

	// Method management
	EnableMethod(ctx context.Context, userID uuid.UUID, method string) error
	DisableMethod(ctx context.Context, userID uuid.UUID, method string) error
	GetEnabledMethods(ctx context.Context, userID uuid.UUID) ([]string, error)
}

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

	// Multiple method support
	EnabledMethods   []TwoFactorMethod
	EmailOTPLength   int
	SMSOTPLength     int
	OTPExpiryMinutes int
}

type TwoFactorService interface {
	// Secret generation and QR code
	GenerateSecret(ctx context.Context, userID uuid.UUID, email string) (*dto.TwoFactorAuthSetupResponse, error)
	GenerateQRCode(secret, email string) (string, error)

	// Token verification
	VerifyToken(secret, token string) bool
	VerifyAndEnable2FA(ctx context.Context, userID uuid.UUID, secret, token string) error

	// 2FA management
	Disable2FA(ctx context.Context, userID uuid.UUID, token, ip, userAgent string) error
	Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error)

	// Backup codes
	GenerateBackupCodes() []string
	ValidateBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, token, ip, userAgent string) (*dto.TwoFactorBackupCodesResponse, error)

	// Login verification
	VerifyLoginToken(ctx context.Context, userID uuid.UUID, token, backupCode, ip, userAgent string) (bool, error)

	// Rate limiting
	IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error)
	LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error

	// Multiple 2FA methods
	GenerateEmailOTP(ctx context.Context, userID uuid.UUID, email string) error
	GenerateSMSOTP(ctx context.Context, userID uuid.UUID, phoneNumber string) error
	VerifyEmailOTP(ctx context.Context, userID uuid.UUID, otp string) (bool, error)
	VerifySMSOTP(ctx context.Context, userID uuid.UUID, otp string) (bool, error)
	GetAvailableMethods(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetEnabledMethods(ctx context.Context, userID uuid.UUID) ([]string, error)
	EnableMethod(ctx context.Context, userID uuid.UUID, method string, data map[string]interface{}) error
	DisableMethod(ctx context.Context, userID uuid.UUID, method, verificationData string) error
	VerifyLoginWithMethod(ctx context.Context, userID uuid.UUID, method, token, ip, userAgent string) (bool, error)
}

func NewTwoFactorService(storage TwoFactorStorage, log *zap.Logger, config TwoFactorConfig) TwoFactorService {
	return &twoFactorService{
		storage: storage,
		log:     log,
		config:  config,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (t *twoFactorService) GenerateSecret(ctx context.Context, userID uuid.UUID, email string) (*dto.TwoFactorAuthSetupResponse, error) {
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

	// Generate backup codes
	backupCodes := t.GenerateBackupCodes()

	return &dto.TwoFactorAuthSetupResponse{
		Secret:        secret.Secret(),
		QRCodeData:    qrCodeURL,
		RecoveryCodes: backupCodes,
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

	// Convert algorithm enum to proper string representation
	var algorithmStr string
	switch t.config.Algorithm {
	case otp.AlgorithmSHA1:
		algorithmStr = "SHA1"
	case otp.AlgorithmSHA256:
		algorithmStr = "SHA256"
	case otp.AlgorithmSHA512:
		algorithmStr = "SHA512"
	default:
		algorithmStr = "SHA1" // Default to SHA1
	}
	q.Set("algorithm", algorithmStr)

	// Convert digits enum to proper string representation
	var digitsStr string
	switch t.config.Digits {
	case otp.DigitsSix:
		digitsStr = "6"
	case otp.DigitsEight:
		digitsStr = "8"
	default:
		digitsStr = "6" // Default to 6 digits
	}
	q.Set("digits", digitsStr)
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
	now := time.Now()
	settings := &dto.TwoFactorSettings{
		UserID:    userID,
		IsEnabled: true,
		EnabledAt: &now,
	}
	err = t.storage.SaveSettings(ctx, settings)
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
func (t *twoFactorService) Disable2FA(ctx context.Context, userID uuid.UUID, token, ip, userAgent string) error {
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
	settings := &dto.TwoFactorSettings{
		UserID:    userID,
		IsEnabled: false,
		EnabledAt: nil,
	}
	err = t.storage.UpdateSettings(ctx, settings)
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
	err = t.LogAttempt(ctx, userID, "disable", true, ip, userAgent)
	if err != nil {
		t.log.Error("Failed to log successful disable attempt", zap.Error(err), zap.String("user_id", userID.String()))
	}

	t.log.Info("2FA disabled successfully", zap.String("user_id", userID.String()))
	return nil
}

// Get2FAStatus retrieves the 2FA status for a user
func (t *twoFactorService) Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error) {
	status, err := t.storage.GetSettings(ctx, userID)
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
	err := t.storage.UseBackupCode(ctx, userID, code)
	if err != nil {
		t.log.Error("Failed to validate backup code", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to validate backup code: %w", err)
	}

	t.log.Info("Backup code validated successfully", zap.String("user_id", userID.String()))
	return true, nil
}

// RegenerateBackupCodes generates new backup codes for a user
func (t *twoFactorService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, token, ip, userAgent string) (*dto.TwoFactorBackupCodesResponse, error) {
	// Verify the token first
	secret, err := t.storage.GetSecret(ctx, userID)
	if err != nil {
		t.log.Error("Failed to get secret for backup code regeneration", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get 2FA secret: %w", err)
	}

	valid := totp.Validate(token, secret)
	if !valid {
		t.LogAttempt(ctx, userID, "regenerate_backup_codes", false, ip, userAgent)
		return nil, fmt.Errorf("invalid 2FA token")
	}

	// Generate new backup codes
	codes := t.GenerateBackupCodes()
	err = t.storage.SaveBackupCodes(ctx, userID, codes)
	if err != nil {
		t.log.Error("Failed to save new backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to save backup codes: %w", err)
	}

	t.LogAttempt(ctx, userID, "regenerate_backup_codes", true, ip, userAgent)
	return &dto.TwoFactorBackupCodesResponse{
		BackupCodes: codes,
	}, nil
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
	return t.storage.IsRateLimited(ctx, userID)
}

// LogAttempt logs a 2FA attempt
func (t *twoFactorService) LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error {
	attempt := &dto.TwoFactorAttempt{
		UserID:       userID,
		AttemptType:  attemptType,
		IsSuccessful: success,
		IPAddress:    ip,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}
	return t.storage.LogAttempt(ctx, attempt)
}

// GenerateEmailOTP generates and sends an OTP via email
func (t *twoFactorService) GenerateEmailOTP(ctx context.Context, userID uuid.UUID, email string) error {
	// Generate random OTP
	otp := t.generateRandomOTP(t.config.EmailOTPLength)

	// Store OTP with expiry
	err := t.storage.SaveOTP(ctx, userID, "email_otp", otp, time.Duration(t.config.OTPExpiryMinutes)*time.Minute)
	if err != nil {
		t.log.Error("Failed to save email OTP", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save email OTP: %w", err)
	}

	// TODO: Send email via email service
	t.log.Info("Email OTP generated", zap.String("user_id", userID.String()), zap.String("email", email))

	return nil
}

// GenerateSMSOTP generates and sends an OTP via SMS
func (t *twoFactorService) GenerateSMSOTP(ctx context.Context, userID uuid.UUID, phoneNumber string) error {
	// Generate random OTP
	otp := t.generateRandomOTP(t.config.SMSOTPLength)

	// Store OTP with expiry
	err := t.storage.SaveOTP(ctx, userID, "sms_otp", otp, time.Duration(t.config.OTPExpiryMinutes)*time.Minute)
	if err != nil {
		t.log.Error("Failed to save SMS OTP", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save SMS OTP: %w", err)
	}

	// TODO: Send SMS via SMS service
	t.log.Info("SMS OTP generated", zap.String("user_id", userID.String()), zap.String("phone", phoneNumber))

	return nil
}

// VerifyEmailOTP verifies an email OTP
func (t *twoFactorService) VerifyEmailOTP(ctx context.Context, userID uuid.UUID, otp string) (bool, error) {
	valid, err := t.storage.VerifyOTP(ctx, userID, "email_otp", otp)
	if err != nil {
		t.log.Error("Failed to verify email OTP", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to verify email OTP: %w", err)
	}

	if valid {
		// Log successful attempt
		t.LogAttempt(ctx, userID, "email_otp_verify", true, "", "")
	} else {
		// Log failed attempt
		t.LogAttempt(ctx, userID, "email_otp_verify", false, "", "")
	}

	return valid, nil
}

// VerifySMSOTP verifies an SMS OTP
func (t *twoFactorService) VerifySMSOTP(ctx context.Context, userID uuid.UUID, otp string) (bool, error) {
	valid, err := t.storage.VerifyOTP(ctx, userID, "sms_otp", otp)
	if err != nil {
		t.log.Error("Failed to verify SMS OTP", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to verify SMS OTP: %w", err)
	}

	if valid {
		// Log successful attempt
		t.LogAttempt(ctx, userID, "sms_otp_verify", true, "", "")
	} else {
		// Log failed attempt
		t.LogAttempt(ctx, userID, "sms_otp_verify", false, "", "")
	}

	return valid, nil
}

// GetAvailableMethods returns enabled 2FA methods for a user
func (t *twoFactorService) GetAvailableMethods(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Get user's enabled methods instead of all configured methods
	enabledMethods, err := t.storage.GetEnabledMethods(ctx, userID)
	if err != nil {
		t.log.Error("Failed to get enabled methods", zap.Error(err), zap.String("user_id", userID.String()))
		return []string{}, fmt.Errorf("failed to get enabled methods: %w", err)
	}

	// Ensure we always return an array, never nil
	if enabledMethods == nil {
		return []string{}, nil
	}

	// Return the user's actual enabled methods
	return enabledMethods, nil
}

// GetEnabledMethods returns enabled 2FA methods for a user
func (t *twoFactorService) GetEnabledMethods(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Get user's enabled methods
	enabledMethods, err := t.storage.GetEnabledMethods(ctx, userID)
	if err != nil {
		t.log.Error("Failed to get enabled methods", zap.Error(err), zap.String("user_id", userID.String()))
		return []string{}, fmt.Errorf("failed to get enabled methods: %w", err)
	}

	// Ensure we always return an array, never nil
	if enabledMethods == nil {
		return []string{}, nil
	}

	// Return the user's actual enabled methods (empty if none enabled)
	return enabledMethods, nil
}

// EnableMethod enables a specific 2FA method for a user
func (t *twoFactorService) EnableMethod(ctx context.Context, userID uuid.UUID, method string, data map[string]interface{}) error {
	switch TwoFactorMethod(method) {
	case MethodTOTP:
		// Enable TOTP method
		if secret, ok := data["secret"].(string); ok {
			// Save the secret first
			err := t.storage.SaveSecret(ctx, userID, secret)
			if err != nil {
				return err
			}
			// Also enable the method in the multi-method system
			return t.storage.EnableMethod(ctx, userID, method)
		}
		return fmt.Errorf("secret required for TOTP method")

	case MethodEmailOTP:
		// Enable email OTP method
		return t.storage.EnableMethod(ctx, userID, method)

	case MethodSMSOTP:
		// Enable SMS OTP method
		return t.storage.EnableMethod(ctx, userID, method)

	case MethodBiometric:
		// Enable biometric method (WebAuthn)
		return t.storage.EnableMethod(ctx, userID, method)

	default:
		return fmt.Errorf("unsupported 2FA method: %s", method)
	}
}

// DisableMethod disables a specific 2FA method for a user
func (t *twoFactorService) DisableMethod(ctx context.Context, userID uuid.UUID, method, verificationData string) error {
	switch TwoFactorMethod(method) {
	case MethodTOTP:
		// Verify TOTP before disabling
		secret, err := t.storage.GetSecret(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		if !t.VerifyToken(secret, verificationData) {
			return fmt.Errorf("invalid verification token")
		}

		return t.storage.DeleteSecret(ctx, userID)

	case MethodEmailOTP, MethodSMSOTP, MethodBiometric:
		// For other methods, just disable them
		return t.storage.DisableMethod(ctx, userID, method)

	default:
		return fmt.Errorf("unsupported 2FA method: %s", method)
	}
}

// generateRandomOTP generates a random OTP of specified length
func (t *twoFactorService) generateRandomOTP(length int) string {
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = digits[mathrand.Intn(len(digits))]
	}
	return string(otp)
}

// VerifyLoginWithMethod verifies login using a specific 2FA method
func (t *twoFactorService) VerifyLoginWithMethod(ctx context.Context, userID uuid.UUID, method, token, ip, userAgent string) (bool, error) {
	// Check rate limiting first
	rateLimited, err := t.IsRateLimited(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	if rateLimited {
		return false, fmt.Errorf("too many failed attempts, please try again later")
	}

	var isValid bool
	var attemptType string

	switch TwoFactorMethod(method) {
	case MethodTOTP:
		// Verify TOTP token
		secret, err := t.storage.GetSecret(ctx, userID)
		if err != nil {
			t.log.Error("Failed to get TOTP secret", zap.Error(err), zap.String("user_id", userID.String()))
			return false, fmt.Errorf("failed to get TOTP secret: %w", err)
		}

		isValid = t.VerifyToken(secret, token)
		attemptType = "totp_login"

	case MethodEmailOTP:
		// Verify email OTP
		isValid, err = t.VerifyEmailOTP(ctx, userID, token)
		if err != nil {
			return false, err
		}
		attemptType = "email_otp_login"

	case MethodSMSOTP:
		// Verify SMS OTP
		isValid, err = t.VerifySMSOTP(ctx, userID, token)
		if err != nil {
			return false, err
		}
		attemptType = "sms_otp_login"

	case MethodBackupCodes:
		// Verify backup code
		isValid, err = t.ValidateBackupCode(ctx, userID, token)
		if err != nil {
			return false, err
		}
		attemptType = "backup_code_login"

	case MethodBiometric:
		// TODO: Implement biometric verification
		return false, fmt.Errorf("biometric authentication not yet implemented")

	default:
		return false, fmt.Errorf("unsupported 2FA method: %s", method)
	}

	// Log the attempt
	err = t.LogAttempt(ctx, userID, attemptType, isValid, ip, userAgent)
	if err != nil {
		t.log.Error("Failed to log attempt", zap.Error(err), zap.String("user_id", userID.String()))
	}

	if isValid {
		t.log.Info("2FA login verification successful",
			zap.String("user_id", userID.String()),
			zap.String("method", method),
			zap.String("ip", ip))
	} else {
		t.log.Warn("2FA login verification failed",
			zap.String("user_id", userID.String()),
			zap.String("method", method),
			zap.String("ip", ip))
	}

	return isValid, nil
}
