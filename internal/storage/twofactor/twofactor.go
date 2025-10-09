package twofactor

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"sync"

	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type twoFactorStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger

	// In-memory storage for testing (TODO: replace with database)
	mu             sync.RWMutex
	enabledMethods map[uuid.UUID][]string
}

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

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) TwoFactorStorage {
	return &twoFactorStorage{
		db:             db,
		log:            log,
		enabledMethods: make(map[uuid.UUID][]string),
	}
}

// SaveSecret saves the 2FA secret for a user
func (t *twoFactorStorage) SaveSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `
		INSERT INTO user_2fa_settings (user_id, secret_key, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			secret_key = EXCLUDED.secret_key,
			updated_at = NOW()
	`

	_, err := t.db.GetPool().Exec(ctx, query, userID, secret)
	if err != nil {
		t.log.Error("Failed to save 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save 2FA secret: %w", err)
	}

	t.log.Info("2FA secret saved successfully", zap.String("user_id", userID.String()))
	return nil
}

// GetSecret retrieves the 2FA secret for a user
func (t *twoFactorStorage) GetSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	query := `SELECT secret_key FROM user_2fa_settings WHERE user_id = $1`

	var secret string
	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&secret)
	if err != nil {
		t.log.Error("Failed to get 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return "", fmt.Errorf("failed to get 2FA secret: %w", err)
	}

	return secret, nil
}

// DeleteSecret removes the 2FA secret for a user
func (t *twoFactorStorage) DeleteSecret(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_2fa_settings WHERE user_id = $1`

	_, err := t.db.GetPool().Exec(ctx, query, userID)
	if err != nil {
		t.log.Error("Failed to delete 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete 2FA secret: %w", err)
	}

	t.log.Info("2FA secret deleted successfully", zap.String("user_id", userID.String()))
	return nil
}

// SaveBackupCodes saves backup codes for a user
func (t *twoFactorStorage) SaveBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	query := `
		UPDATE user_2fa_settings 
		SET backup_codes = $2, updated_at = NOW()
		WHERE user_id = $1
	`

	result, err := t.db.GetPool().Exec(ctx, query, userID, codes)
	if err != nil {
		t.log.Error("Failed to save backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save backup codes: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		t.log.Error("No 2FA settings found for user", zap.String("user_id", userID.String()))
		return fmt.Errorf("no 2FA settings found for user")
	}

	t.log.Info("Backup codes saved successfully", zap.String("user_id", userID.String()))
	return nil
}

// GetBackupCodes retrieves backup codes for a user
func (t *twoFactorStorage) GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT backup_codes FROM user_2fa_settings WHERE user_id = $1`

	var codes []string
	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&codes)
	if err != nil {
		t.log.Error("Failed to get backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get backup codes: %w", err)
	}

	return codes, nil
}

// ValidateAndConsumeBackupCode validates and removes a backup code
func (t *twoFactorStorage) ValidateAndConsumeBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	query := `SELECT backup_codes FROM user_2fa_settings WHERE user_id = $1`

	var codes []string
	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&codes)
	if err != nil {
		t.log.Error("Failed to get backup codes for validation", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to get backup codes: %w", err)
	}

	// Find and remove the code
	for i, storedCode := range codes {
		if storedCode == code {
			// Remove the used code
			codes = append(codes[:i], codes[i+1:]...)

			// Update the database
			updateQuery := `UPDATE user_2fa_settings SET backup_codes = $1, updated_at = NOW() WHERE user_id = $2`
			_, err := t.db.GetPool().Exec(ctx, updateQuery, codes, userID)
			if err != nil {
				t.log.Error("Failed to update backup codes after consumption", zap.Error(err), zap.String("user_id", userID.String()))
				return false, fmt.Errorf("failed to update backup codes: %w", err)
			}

			t.log.Info("Backup code consumed successfully", zap.String("user_id", userID.String()))
			return true, nil
		}
	}

	t.log.Warn("Invalid backup code used", zap.String("user_id", userID.String()))
	return false, nil
}

// Enable2FA enables 2FA for a user
func (t *twoFactorStorage) Enable2FA(ctx context.Context, userID uuid.UUID) error {
	// Start transaction
	tx, err := t.db.GetPool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update user_2fa_settings
	query1 := `
		UPDATE user_2fa_settings 
		SET is_enabled = TRUE, enabled_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
	`
	_, err = tx.Exec(ctx, query1, userID)
	if err != nil {
		t.log.Error("Failed to enable 2FA in settings", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Update users table
	query2 := `
		UPDATE users 
		SET two_factor_enabled = TRUE, two_factor_setup_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, query2, userID)
	if err != nil {
		t.log.Error("Failed to enable 2FA in users table", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.log.Info("2FA enabled successfully", zap.String("user_id", userID.String()))
	return nil
}

// Disable2FA disables 2FA for a user
func (t *twoFactorStorage) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	// Start transaction
	tx, err := t.db.GetPool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update user_2fa_settings
	query1 := `
		UPDATE user_2fa_settings 
		SET is_enabled = FALSE, enabled_at = NULL, updated_at = NOW()
		WHERE user_id = $1
	`
	_, err = tx.Exec(ctx, query1, userID)
	if err != nil {
		t.log.Error("Failed to disable 2FA in settings", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Update users table
	query2 := `
		UPDATE users 
		SET two_factor_enabled = FALSE, two_factor_setup_at = NULL
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, query2, userID)
	if err != nil {
		t.log.Error("Failed to disable 2FA in users table", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.log.Info("2FA disabled successfully", zap.String("user_id", userID.String()))
	return nil
}

// Get2FAStatus retrieves the 2FA status for a user
func (t *twoFactorStorage) Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error) {
	query := `
		SELECT is_enabled, enabled_at, last_used_at, backup_codes
		FROM user_2fa_settings 
		WHERE user_id = $1
	`

	var settings dto.TwoFactorSettings
	var enabledAt, lastUsedAt *time.Time
	var backupCodes []string

	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&settings.IsEnabled, &enabledAt, &lastUsedAt, &backupCodes)
	if err != nil {
		// If no record exists, return default settings
		return &dto.TwoFactorSettings{
			UserID:    userID,
			IsEnabled: false,
			EnabledAt: nil,
		}, nil
	}

	if enabledAt != nil {
		settings.EnabledAt = enabledAt
	}

	return &settings, nil
}

// LogAttempt logs a 2FA attempt
func (t *twoFactorStorage) LogAttempt(ctx context.Context, attempt *dto.TwoFactorAttempt) error {
	query := `INSERT INTO user_2fa_attempts (user_id, attempt_type, is_successful, ip_address, user_agent, created_at) VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := t.db.GetPool().Exec(ctx, query, attempt.UserID, attempt.AttemptType, attempt.IsSuccessful, attempt.IPAddress, attempt.UserAgent)
	if err != nil {
		t.log.Error("Failed to log 2FA attempt", zap.Error(err), zap.String("user_id", attempt.UserID.String()))
		return fmt.Errorf("failed to log 2FA attempt: %w", err)
	}
	return nil
}

// GetRecentAttempts retrieves recent 2FA attempts for a user
func (t *twoFactorStorage) GetRecentAttempts(ctx context.Context, userID uuid.UUID, minutes int) ([]dto.TwoFactorAttempt, error) {
	query := `
		SELECT id, user_id, attempt_type, is_successful, ip_address, user_agent, created_at
		FROM user_2fa_attempts 
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '%d minutes'
		ORDER BY created_at DESC
	`

	rows, err := t.db.GetPool().Query(ctx, fmt.Sprintf(query, minutes), userID)
	if err != nil {
		t.log.Error("Failed to get recent 2FA attempts", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get recent attempts: %w", err)
	}
	defer rows.Close()

	var attempts []dto.TwoFactorAttempt
	for rows.Next() {
		var attempt dto.TwoFactorAttempt
		err := rows.Scan(&attempt.ID, &attempt.UserID, &attempt.AttemptType, &attempt.IsSuccessful, &attempt.IPAddress, &attempt.UserAgent, &attempt.CreatedAt)
		if err != nil {
			t.log.Error("Failed to scan 2FA attempt", zap.Error(err))
			continue
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// IsRateLimited checks if a user is rate limited
func (t *twoFactorStorage) IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM user_2fa_attempts WHERE user_id = $1 AND is_successful = false AND created_at > NOW() - INTERVAL '5 minutes'`
	var count int
	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		t.log.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	return count >= 5, nil
}

// DeleteBackupCodes deletes backup codes for a user
func (t *twoFactorStorage) DeleteBackupCodes(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE user_2fa_settings SET backup_codes = NULL WHERE user_id = $1`
	_, err := t.db.GetPool().Exec(ctx, query, userID)
	if err != nil {
		t.log.Error("Failed to delete backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}
	return nil
}

// UseBackupCode validates and consumes a backup code
func (t *twoFactorStorage) UseBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	// This would need to implement proper backup code validation and consumption
	// For now, we'll just return an error indicating the code was used
	return fmt.Errorf("backup code validation not implemented")
}

// SaveSettings saves 2FA settings
func (t *twoFactorStorage) SaveSettings(ctx context.Context, settings *dto.TwoFactorSettings) error {
	query := `
		UPDATE user_2fa_settings 
		SET is_enabled = $2, enabled_at = $3, updated_at = NOW()
		WHERE user_id = $1
	`

	result, err := t.db.GetPool().Exec(ctx, query, settings.UserID, settings.IsEnabled, settings.EnabledAt)
	if err != nil {
		t.log.Error("Failed to save 2FA settings", zap.Error(err), zap.String("user_id", settings.UserID.String()))
		return fmt.Errorf("failed to save 2FA settings: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		t.log.Error("No 2FA settings found for user", zap.String("user_id", settings.UserID.String()))
		return fmt.Errorf("no 2FA settings found for user")
	}

	return nil
}

// GetSettings gets 2FA settings for a user
func (t *twoFactorStorage) GetSettings(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error) {
	query := `SELECT is_enabled, enabled_at FROM user_2fa_settings WHERE user_id = $1`
	var settings dto.TwoFactorSettings
	var enabledAt *time.Time

	err := t.db.GetPool().QueryRow(ctx, query, userID).Scan(&settings.IsEnabled, &enabledAt)
	if err != nil {
		// If no record exists, return default settings
		return &dto.TwoFactorSettings{
			UserID:    userID,
			IsEnabled: false,
			EnabledAt: nil,
		}, nil
	}

	settings.UserID = userID
	if enabledAt != nil {
		settings.EnabledAt = enabledAt
	}

	return &settings, nil
}

// UpdateSettings updates 2FA settings
func (t *twoFactorStorage) UpdateSettings(ctx context.Context, settings *dto.TwoFactorSettings) error {
	query := `UPDATE user_2fa_settings SET is_enabled = $2, enabled_at = $3 WHERE user_id = $1`
	_, err := t.db.GetPool().Exec(ctx, query, settings.UserID, settings.IsEnabled, settings.EnabledAt)
	if err != nil {
		t.log.Error("Failed to update 2FA settings", zap.Error(err), zap.String("user_id", settings.UserID.String()))
		return fmt.Errorf("failed to update 2FA settings: %w", err)
	}
	return nil
}

// DeleteSettings deletes 2FA settings
func (t *twoFactorStorage) DeleteSettings(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_2fa_settings WHERE user_id = $1`
	_, err := t.db.GetPool().Exec(ctx, query, userID)
	if err != nil {
		t.log.Error("Failed to delete 2FA settings", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete 2FA settings: %w", err)
	}
	return nil
}

// GetAttempts gets recent 2FA attempts for a user
func (t *twoFactorStorage) GetAttempts(ctx context.Context, userID uuid.UUID, limit int) ([]dto.TwoFactorAttempt, error) {
	query := `SELECT id, user_id, attempt_type, is_successful, ip_address, user_agent, created_at FROM user_2fa_attempts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := t.db.GetPool().Query(ctx, query, userID, limit)
	if err != nil {
		t.log.Error("Failed to get 2FA attempts", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get 2FA attempts: %w", err)
	}
	defer rows.Close()

	var attempts []dto.TwoFactorAttempt
	for rows.Next() {
		var attempt dto.TwoFactorAttempt
		err := rows.Scan(&attempt.ID, &attempt.UserID, &attempt.AttemptType, &attempt.IsSuccessful, &attempt.IPAddress, &attempt.UserAgent, &attempt.CreatedAt)
		if err != nil {
			t.log.Error("Failed to scan 2FA attempt", zap.Error(err))
			return nil, fmt.Errorf("failed to scan 2FA attempt: %w", err)
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// generateBackupCodes generates secure backup codes
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate 8 random bytes
		bytes := make([]byte, 8)
		rand.Read(bytes)

		// Convert to base32 and take first 8 characters
		encoded := base32.StdEncoding.EncodeToString(bytes)
		codes[i] = strings.ToUpper(encoded[:8])
	}
	return codes
}

// SaveOTP saves an OTP for a specific method
func (t *twoFactorStorage) SaveOTP(ctx context.Context, userID uuid.UUID, method, otp string, expiry time.Duration) error {
	// TODO: Implement OTP storage in database
	// For now, just log the OTP
	t.log.Info("OTP saved", zap.String("user_id", userID.String()), zap.String("method", method), zap.String("otp", otp))
	return nil
}

// VerifyOTP verifies an OTP for a specific method
func (t *twoFactorStorage) VerifyOTP(ctx context.Context, userID uuid.UUID, method, otp string) (bool, error) {
	// TODO: Implement OTP verification from database
	// For now, return false
	t.log.Info("OTP verification attempted", zap.String("user_id", userID.String()), zap.String("method", method), zap.String("otp", otp))
	return false, nil
}

// DeleteOTP deletes an OTP for a specific method
func (t *twoFactorStorage) DeleteOTP(ctx context.Context, userID uuid.UUID, method string) error {
	// TODO: Implement OTP deletion from database
	t.log.Info("OTP deleted", zap.String("user_id", userID.String()), zap.String("method", method))
	return nil
}

// EnableMethod enables a specific 2FA method for a user
func (t *twoFactorStorage) EnableMethod(ctx context.Context, userID uuid.UUID, method string) error {
	// Save to database
	query := `
		INSERT INTO user_2fa_methods (user_id, method, enabled_at, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, method) 
		DO UPDATE SET enabled_at = NOW(), updated_at = NOW()
	`

	_, err := t.db.GetPool().Exec(ctx, query, userID, method)
	if err != nil {
		t.log.Error("Failed to enable method in database", zap.Error(err), zap.String("user_id", userID.String()), zap.String("method", method))
		return fmt.Errorf("failed to enable method: %w", err)
	}

	// Also update in-memory storage for immediate UI updates
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get current enabled methods
	methods := t.enabledMethods[userID]

	// Check if method is already enabled
	for _, m := range methods {
		if m == method {
			return nil // Already enabled
		}
	}

	// Add method to enabled list
	t.enabledMethods[userID] = append(methods, method)

	t.log.Info("Method enabled", zap.String("user_id", userID.String()), zap.String("method", method))
	return nil
}

// DisableMethod disables a specific 2FA method for a user
func (t *twoFactorStorage) DisableMethod(ctx context.Context, userID uuid.UUID, method string) error {
	// Save to database
	query := `
		UPDATE user_2fa_methods 
		SET enabled_at = NULL, disabled_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND method = $2
	`

	result, err := t.db.GetPool().Exec(ctx, query, userID, method)
	if err != nil {
		t.log.Error("Failed to disable method in database", zap.Error(err), zap.String("user_id", userID.String()), zap.String("method", method))
		return fmt.Errorf("failed to disable method: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		t.log.Warn("Method was not enabled for user", zap.String("user_id", userID.String()), zap.String("method", method))
		// Don't return error, just log warning
	}

	// Also update in-memory storage for immediate UI updates
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get current enabled methods
	methods := t.enabledMethods[userID]

	// Remove method from enabled list
	var newMethods []string
	for _, m := range methods {
		if m != method {
			newMethods = append(newMethods, m)
		}
	}

	t.enabledMethods[userID] = newMethods

	t.log.Info("Method disabled", zap.String("user_id", userID.String()), zap.String("method", method))
	return nil
}

// GetEnabledMethods returns enabled methods for a user
func (t *twoFactorStorage) GetEnabledMethods(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// First try to get from database (for refresh/restart scenarios)
	query := `
		SELECT method 
		FROM user_2fa_methods 
		WHERE user_id = $1 AND enabled_at IS NOT NULL
		ORDER BY enabled_at ASC
	`

	rows, err := t.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		t.log.Error("Failed to get enabled methods from database", zap.Error(err), zap.String("user_id", userID.String()))
		// Fall back to in-memory storage
		return t.getEnabledMethodsFromMemory(userID), nil
	}
	defer rows.Close()

	var methods []string
	for rows.Next() {
		var method string
		if err := rows.Scan(&method); err != nil {
			t.log.Error("Failed to scan method", zap.Error(err))
			continue
		}
		methods = append(methods, method)
	}

	if err := rows.Err(); err != nil {
		t.log.Error("Error iterating over methods", zap.Error(err))
		return t.getEnabledMethodsFromMemory(userID), nil
	}

	// Update in-memory storage with database data
	t.mu.Lock()
	t.enabledMethods[userID] = methods
	t.mu.Unlock()

	// Ensure we always return an array, never nil
	if methods == nil {
		return []string{}, nil
	}

	return methods, nil
}

// getEnabledMethodsFromMemory returns enabled methods from in-memory storage
func (t *twoFactorStorage) getEnabledMethodsFromMemory(userID uuid.UUID) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if methods, exists := t.enabledMethods[userID]; exists {
		// Ensure we always return an array, never nil
		if methods == nil {
			return []string{}
		}
		return methods
	}

	// Return empty slice if no methods enabled
	return []string{}
}
