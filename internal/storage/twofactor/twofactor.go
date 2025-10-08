package twofactor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
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
}

type TwoFactorStorage interface {
	// Secret management
	SaveSecret(ctx context.Context, userID uuid.UUID, secret string) error
	GetSecret(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteSecret(ctx context.Context, userID uuid.UUID) error

	// Backup codes management
	SaveBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error
	GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	ValidateAndConsumeBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// 2FA status management
	Enable2FA(ctx context.Context, userID uuid.UUID) error
	Disable2FA(ctx context.Context, userID uuid.UUID) error
	Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error)

	// Attempt logging
	LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error
	GetRecentAttempts(ctx context.Context, userID uuid.UUID, minutes int) ([]dto.TwoFactorAttempt, error)
	IsRateLimited(ctx context.Context, userID uuid.UUID, maxAttempts int, windowMinutes int) (bool, error)
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) TwoFactorStorage {
	return &twoFactorStorage{
		db:  db,
		log: log,
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
	
	_, err := t.db.DB.Exec(ctx, query, userID, secret)
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
	err := t.db.DB.QueryRow(ctx, query, userID).Scan(&secret)
	if err != nil {
		t.log.Error("Failed to get 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		return "", fmt.Errorf("failed to get 2FA secret: %w", err)
	}

	return secret, nil
}

// DeleteSecret removes the 2FA secret for a user
func (t *twoFactorStorage) DeleteSecret(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_2fa_settings WHERE user_id = $1`
	
	_, err := t.db.DB.Exec(ctx, query, userID)
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
		INSERT INTO user_2fa_settings (user_id, backup_codes, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			backup_codes = EXCLUDED.backup_codes,
			updated_at = NOW()
	`
	
	_, err := t.db.DB.Exec(ctx, query, userID, codes)
	if err != nil {
		t.log.Error("Failed to save backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to save backup codes: %w", err)
	}

	t.log.Info("Backup codes saved successfully", zap.String("user_id", userID.String()))
	return nil
}

// GetBackupCodes retrieves backup codes for a user
func (t *twoFactorStorage) GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT backup_codes FROM user_2fa_settings WHERE user_id = $1`
	
	var codes []string
	err := t.db.DB.QueryRow(ctx, query, userID).Scan(&codes)
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
	err := t.db.DB.QueryRow(ctx, query, userID).Scan(&codes)
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
			_, err := t.db.DB.Exec(ctx, updateQuery, codes, userID)
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
	tx, err := t.db.DB.Begin(ctx)
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
	tx, err := t.db.DB.Begin(ctx)
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
	
	err := t.db.DB.QueryRow(ctx, query, userID).Scan(&settings.IsEnabled, &enabledAt, &lastUsedAt, &backupCodes)
	if err != nil {
		// If no record exists, return default settings
		return &dto.TwoFactorSettings{
			IsEnabled:      false,
			BackupCodes:    []string{},
			HasBackupCodes: false,
		}, nil
	}

	if enabledAt != nil {
		settings.EnabledAt = *enabledAt
	}
	if lastUsedAt != nil {
		settings.LastUsedAt = *lastUsedAt
	}
	
	settings.BackupCodes = backupCodes
	settings.HasBackupCodes = len(backupCodes) > 0

	return &settings, nil
}

// LogAttempt logs a 2FA attempt
func (t *twoFactorStorage) LogAttempt(ctx context.Context, userID uuid.UUID, attemptType string, success bool, ip, userAgent string) error {
	query := `
		INSERT INTO user_2fa_attempts (user_id, attempt_type, is_successful, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	
	_, err := t.db.DB.Exec(ctx, query, userID, attemptType, success, ip, userAgent)
	if err != nil {
		t.log.Error("Failed to log 2FA attempt", zap.Error(err), zap.String("user_id", userID.String()))
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
	
	rows, err := t.db.DB.Query(ctx, fmt.Sprintf(query, minutes), userID)
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
func (t *twoFactorStorage) IsRateLimited(ctx context.Context, userID uuid.UUID, maxAttempts int, windowMinutes int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM user_2fa_attempts 
		WHERE user_id = $1 
		AND is_successful = FALSE 
		AND created_at >= NOW() - INTERVAL '%d minutes'
	`
	
	var count int
	err := t.db.DB.QueryRow(ctx, fmt.Sprintf(query, windowMinutes), userID).Scan(&count)
	if err != nil {
		t.log.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return count >= maxAttempts, nil
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
