package system_config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type SystemConfig struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

type WithdrawalGlobalStatus struct {
	Enabled  bool    `json:"enabled"`
	Reason   *string `json:"reason,omitempty"`
	PausedBy *string `json:"paused_by,omitempty"`
	PausedAt *string `json:"paused_at,omitempty"`
}

type WithdrawalThreshold struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
	Enabled  bool    `json:"enabled"`
}

type WithdrawalThresholds struct {
	HourlyVolume      WithdrawalThreshold `json:"hourly_volume"`
	DailyVolume       WithdrawalThreshold `json:"daily_volume"`
	SingleTransaction WithdrawalThreshold `json:"single_transaction"`
	UserDaily         WithdrawalThreshold `json:"user_daily"`
}

type WithdrawalManualReview struct {
	Enabled              bool  `json:"enabled"`
	ThresholdCents       int64 `json:"threshold_cents"`
	RequireAdminApproval bool  `json:"require_admin_approval"`
}

// Settings structures for the admin panel
type GeneralSettings struct {
	SiteName            string `json:"site_name"`
	SiteDescription     string `json:"site_description"`
	SupportEmail        string `json:"support_email"`
	Timezone            string `json:"timezone"`
	Language            string `json:"language"`
	MaintenanceMode     bool   `json:"maintenance_mode"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	DemoMode            bool   `json:"demo_mode"`
}

type PaymentSettings struct {
	MinDepositBTC    float64 `json:"min_deposit_btc"`
	MaxDepositBTC    float64 `json:"max_deposit_btc"`
	MinWithdrawalBTC float64 `json:"min_withdrawal_btc"`
	MaxWithdrawalBTC float64 `json:"max_withdrawal_btc"`
	KycRequired      bool    `json:"kyc_required"`
	KycThreshold     int     `json:"kyc_threshold"`
}

type SecuritySettings struct {
	SessionTimeout         int  `json:"session_timeout"`
	MaxLoginAttempts       int  `json:"max_login_attempts"`
	LockoutDuration        int  `json:"lockout_duration"`
	TwoFactorRequired      bool `json:"two_factor_required"`
	PasswordMinLength      int  `json:"password_min_length"`
	PasswordRequireSpecial bool `json:"password_require_special"`
	IpWhitelistEnabled     bool `json:"ip_whitelist_enabled"`
	RateLimitEnabled       bool `json:"rate_limit_enabled"`
	RateLimitRequests      int  `json:"rate_limit_requests"`
}

type TipSettings struct {
	TipTransactionFeeFromWho string  `json:"tip_transaction_fee_from_who"` // "sender" or "receiver"
	TransactionFee           float64 `json:"transaction_fee"`              // 1-100
}

type GeoBlockingSettings struct {
	EnableGeoBlocking bool     `json:"enable_geo_blocking"`
	DefaultAction     string   `json:"default_action"` // "allow" or "block"
	VpnDetection      bool     `json:"vpn_detection"`
	ProxyDetection    bool     `json:"proxy_detection"`
	TorBlocking       bool     `json:"tor_blocking"`
	LogAttempts       bool     `json:"log_attempts"`
	BlockedCountries  []string `json:"blocked_countries"`
	AllowedCountries  []string `json:"allowed_countries"`
	BypassCountries   []string `json:"bypass_countries"`
}

func NewSystemConfig(db *persistencedb.PersistenceDB, log *zap.Logger) *SystemConfig {
	return &SystemConfig{
		db:  db,
		log: log,
	}
}

// GetWithdrawalGlobalStatus retrieves the current global withdrawal status
func (s *SystemConfig) GetWithdrawalGlobalStatus(ctx context.Context) (WithdrawalGlobalStatus, error) {
	s.log.Info("Getting withdrawal global status from system config")

	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_global_status").Scan(&configValue)
	if err != nil {
		s.log.Error("Failed to get withdrawal global status", zap.Error(err))
		return WithdrawalGlobalStatus{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal global status")
	}

	var status WithdrawalGlobalStatus
	if err := json.Unmarshal(configValue, &status); err != nil {
		s.log.Error("Failed to unmarshal withdrawal global status", zap.Error(err))
		return WithdrawalGlobalStatus{}, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal withdrawal global status")
	}

	return status, nil
}

// UpdateWithdrawalGlobalStatus updates the global withdrawal status
func (s *SystemConfig) UpdateWithdrawalGlobalStatus(ctx context.Context, status WithdrawalGlobalStatus, adminID uuid.UUID) error {
	s.log.Info("Updating withdrawal global status",
		zap.Bool("enabled", status.Enabled),
		zap.String("reason", *status.Reason))

	configValue, err := json.Marshal(status)
	if err != nil {
		s.log.Error("Failed to marshal withdrawal global status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to marshal withdrawal global status")
	}

	_, err = s.db.GetPool().Exec(ctx, "UPDATE system_config SET config_value = $1, updated_by = $2, updated_at = NOW() WHERE config_key = $3", configValue, &adminID, "withdrawal_global_status")
	if err != nil {
		s.log.Error("Failed to update withdrawal global status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal global status")
	}

	s.log.Info("Successfully updated withdrawal global status")
	return nil
}

// GetWithdrawalThresholds retrieves the current withdrawal thresholds
func (s *SystemConfig) GetWithdrawalThresholds(ctx context.Context) (WithdrawalThresholds, error) {
	s.log.Info("Getting withdrawal thresholds from system config")

	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_thresholds").Scan(&configValue)
	if err != nil {
		s.log.Error("Failed to get withdrawal thresholds", zap.Error(err))
		return WithdrawalThresholds{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal thresholds")
	}

	var thresholds WithdrawalThresholds
	if err := json.Unmarshal(configValue, &thresholds); err != nil {
		s.log.Error("Failed to unmarshal withdrawal thresholds", zap.Error(err))
		return WithdrawalThresholds{}, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal withdrawal thresholds")
	}

	return thresholds, nil
}

// UpdateWithdrawalThresholds updates the withdrawal thresholds
func (s *SystemConfig) UpdateWithdrawalThresholds(ctx context.Context, thresholds WithdrawalThresholds, adminID uuid.UUID) error {
	s.log.Info("Updating withdrawal thresholds")

	configValue, err := json.Marshal(thresholds)
	if err != nil {
		s.log.Error("Failed to marshal withdrawal thresholds", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to marshal withdrawal thresholds")
	}

	_, err = s.db.GetPool().Exec(ctx, "UPDATE system_config SET config_value = $1, updated_by = $2, updated_at = NOW() WHERE config_key = $3", configValue, &adminID, "withdrawal_thresholds")
	if err != nil {
		s.log.Error("Failed to update withdrawal thresholds", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal thresholds")
	}

	s.log.Info("Successfully updated withdrawal thresholds")
	return nil
}

// GetWithdrawalManualReview retrieves the manual review settings
func (s *SystemConfig) GetWithdrawalManualReview(ctx context.Context) (WithdrawalManualReview, error) {
	s.log.Info("Getting withdrawal manual review settings from system config")

	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_manual_review").Scan(&configValue)
	if err != nil {
		s.log.Error("Failed to get withdrawal manual review settings", zap.Error(err))
		return WithdrawalManualReview{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal manual review settings")
	}

	var review WithdrawalManualReview
	if err := json.Unmarshal(configValue, &review); err != nil {
		s.log.Error("Failed to unmarshal withdrawal manual review settings", zap.Error(err))
		return WithdrawalManualReview{}, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal withdrawal manual review settings")
	}

	return review, nil
}

// UpdateWithdrawalManualReview updates the manual review settings
func (s *SystemConfig) UpdateWithdrawalManualReview(ctx context.Context, review WithdrawalManualReview, adminID uuid.UUID) error {
	s.log.Info("Updating withdrawal manual review settings")

	configValue, err := json.Marshal(review)
	if err != nil {
		s.log.Error("Failed to marshal withdrawal manual review settings", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to marshal withdrawal manual review settings")
	}

	_, err = s.db.GetPool().Exec(ctx, "UPDATE system_config SET config_value = $1, updated_by = $2, updated_at = NOW() WHERE config_key = $3", configValue, &adminID, "withdrawal_manual_review")
	if err != nil {
		s.log.Error("Failed to update withdrawal manual review settings", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal manual review settings")
	}

	s.log.Info("Successfully updated withdrawal manual review settings")
	return nil
}

// CheckWithdrawalAllowed checks if withdrawals are globally allowed
func (s *SystemConfig) CheckWithdrawalAllowed(ctx context.Context) (bool, string, error) {
	status, err := s.GetWithdrawalGlobalStatus(ctx)
	if err != nil {
		return false, "", err
	}

	if !status.Enabled {
		reason := "Withdrawals are currently disabled"
		if status.Reason != nil {
			reason = *status.Reason
		}
		return false, reason, nil
	}

	return true, "", nil
}

// CheckWithdrawalThresholds checks if a withdrawal amount exceeds any thresholds
func (s *SystemConfig) CheckWithdrawalThresholds(ctx context.Context, amount float64, currency string, thresholdType string) (bool, string, error) {
	thresholds, err := s.GetWithdrawalThresholds(ctx)
	if err != nil {
		return false, "", err
	}

	var threshold WithdrawalThreshold
	var thresholdName string

	switch thresholdType {
	case "hourly_volume":
		threshold = thresholds.HourlyVolume
		thresholdName = "hourly volume"
	case "daily_volume":
		threshold = thresholds.DailyVolume
		thresholdName = "daily volume"
	case "single_transaction":
		threshold = thresholds.SingleTransaction
		thresholdName = "single transaction"
	case "user_daily":
		threshold = thresholds.UserDaily
		thresholdName = "user daily"
	default:
		return false, "", fmt.Errorf("unknown threshold type: %s", thresholdType)
	}

	if !threshold.Enabled {
		return false, "", nil
	}

	if amount > threshold.Value {
		reason := fmt.Sprintf("Withdrawal amount %.2f %s exceeds %s threshold of %.2f %s",
			amount, currency, thresholdName, threshold.Value, threshold.Currency)
		return true, reason, nil
	}

	return false, "", nil
}

// GetWithdrawalPauseReasons retrieves the predefined pause reasons
func (s *SystemConfig) GetWithdrawalPauseReasons(ctx context.Context) ([]string, error) {
	s.log.Info("Getting withdrawal pause reasons from system config")

	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, "SELECT config_value FROM system_config WHERE config_key = $1", "withdrawal_pause_reasons").Scan(&configValue)
	if err != nil {
		s.log.Error("Failed to get withdrawal pause reasons", zap.Error(err))
		return []string{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal pause reasons")
	}

	var reasons []string
	if err := json.Unmarshal(configValue, &reasons); err != nil {
		s.log.Error("Failed to unmarshal withdrawal pause reasons", zap.Error(err))
		return []string{}, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal withdrawal pause reasons")
	}

	return reasons, nil
}

// GetGeneralSettings retrieves general settings from system config
func (s *SystemConfig) GetGeneralSettings(ctx context.Context, brandID *uuid.UUID) (GeneralSettings, error) {
	s.log.Info("Getting general settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return GeneralSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "general_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No general settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return GeneralSettings{
				SiteName:            "",
				SiteDescription:     "",
				SupportEmail:        "",
				Timezone:            "UTC",
				Language:            "en",
				MaintenanceMode:     false,
				RegistrationEnabled: true,
				DemoMode:            false,
			}, nil
		}
		s.log.Error("Failed to get general settings from database", zap.Error(err))
		return GeneralSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get general settings")
	}

	var settings GeneralSettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal general settings", zap.Error(err))
		return GeneralSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse general settings")
	}

	return settings, nil
}

// UpdateGeneralSettings updates general settings in system config
// If brandID is nil, updates ALL brands. If brandID is set, updates only that brand.
func (s *SystemConfig) UpdateGeneralSettings(ctx context.Context, settings GeneralSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	s.log.Info("Updating general settings", zap.Any("brand_id", brandID))

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal general settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal general settings")
	}

	if brandID == nil {
		// Update ALL brands when brandID is nil (global update)
		_, err = s.db.GetPool().Exec(ctx, `
			WITH brand_ids AS (
				SELECT id FROM brands
			)
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			SELECT 
				$1,
				$2,
				$3,
				bi.id,
				$4,
				NOW()
			FROM brand_ids bi
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "general_settings", configValue, "General application settings", adminID)

		if err != nil {
			s.log.Error("Failed to update general settings for all brands", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to update general settings for all brands")
		}
		s.log.Info("General settings updated for all brands")
	} else {
		// Update specific brand
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "general_settings", configValue, "General application settings", brandID, adminID)

		if err != nil {
			s.log.Error("Failed to update general settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
			return errors.ErrInternalServerError.Wrap(err, "failed to update general settings")
		}
		s.log.Info("General settings updated for specific brand", zap.Any("brand_id", brandID))
	}

	return nil
}

// GetPaymentSettings retrieves payment settings from system config
func (s *SystemConfig) GetPaymentSettings(ctx context.Context, brandID *uuid.UUID) (PaymentSettings, error) {
	s.log.Info("Getting payment settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return PaymentSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "payment_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No payment settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return PaymentSettings{
				MinDepositBTC:    0.001,
				MaxDepositBTC:    10.0,
				MinWithdrawalBTC: 0.001,
				MaxWithdrawalBTC: 10.0,
				KycRequired:      false,
				KycThreshold:     1000,
			}, nil
		}
		s.log.Error("Failed to get payment settings from database", zap.Error(err))
		return PaymentSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get payment settings")
	}

	var settings PaymentSettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal payment settings", zap.Error(err))
		return PaymentSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse payment settings")
	}

	return settings, nil
}

// UpdatePaymentSettings updates payment settings in system config
// If brandID is nil, updates ALL brands. If brandID is set, updates only that brand.
func (s *SystemConfig) UpdatePaymentSettings(ctx context.Context, settings PaymentSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	s.log.Info("Updating payment settings", zap.Any("brand_id", brandID))

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal payment settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal payment settings")
	}

	if brandID == nil {
		// Update ALL brands when brandID is nil (global update)
		_, err = s.db.GetPool().Exec(ctx, `
			WITH brand_ids AS (
				SELECT id FROM brands
			)
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			SELECT 
				$1,
				$2,
				$3,
				bi.id,
				$4,
				NOW()
			FROM brand_ids bi
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "payment_settings", configValue, "Payment processing settings", adminID)

		if err != nil {
			s.log.Error("Failed to update payment settings for all brands", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to update payment settings for all brands")
		}
		s.log.Info("Payment settings updated for all brands")
	} else {
		// Update specific brand
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "payment_settings", configValue, "Payment processing settings", brandID, adminID)

		if err != nil {
			s.log.Error("Failed to update payment settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
			return errors.ErrInternalServerError.Wrap(err, "failed to update payment settings")
		}
		s.log.Info("Payment settings updated for specific brand", zap.Any("brand_id", brandID))
	}

	return nil
}

// GetTipSettings retrieves tip settings from system config
func (s *SystemConfig) GetTipSettings(ctx context.Context, brandID *uuid.UUID) (TipSettings, error) {
	s.log.Info("Getting tip settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return TipSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "tip_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No tip settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return TipSettings{
				TipTransactionFeeFromWho: "sender",
				TransactionFee:           0.0,
			}, nil
		}
		s.log.Error("Failed to get tip settings from database", zap.Error(err))
		return TipSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get tip settings")
	}

	var settings TipSettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal tip settings", zap.Error(err))
		return TipSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse tip settings")
	}

	return settings, nil
}

// UpdateTipSettings updates tip settings in system config for a specific brand
// brandID is required - checks if tip_settings exists for that brand_id, updates if exists, inserts if not
func (s *SystemConfig) UpdateTipSettings(ctx context.Context, settings TipSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	if brandID == nil {
		s.log.Error("brand_id is required for tip settings")
		return errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required for tip settings")
	}

	s.log.Info("Updating tip settings", zap.Any("brand_id", brandID))

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal tip settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal tip settings")
	}

	// Check if tip_settings exists for this brand_id
	var exists bool
	err = s.db.GetPool().QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM system_config 
			WHERE config_key = $1 AND brand_id = $2
		)
	`, "tip_settings", brandID).Scan(&exists)

	if err != nil {
		s.log.Error("Failed to check if tip settings exist", zap.Error(err), zap.Any("brand_id", brandID))
		return errors.ErrInternalServerError.Wrap(err, "failed to check tip settings")
	}

	// Insert or update based on existence
	// ON CONFLICT will handle the update if it exists, or insert if it doesn't
	_, err = s.db.GetPool().Exec(ctx, `
		INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (brand_id, config_key) 
		DO UPDATE SET 
			config_value = EXCLUDED.config_value,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
	`, "tip_settings", configValue, "Tip transaction fee settings", brandID, adminID)

	if err != nil {
		s.log.Error("Failed to update tip settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
		return errors.ErrInternalServerError.Wrap(err, "failed to update tip settings")
	}

	if exists {
		s.log.Info("Tip settings updated for brand", zap.Any("brand_id", brandID))
	} else {
		s.log.Info("Tip settings created for brand", zap.Any("brand_id", brandID))
	}

	return nil
}

// GetSecuritySettings retrieves security settings from system config
func (s *SystemConfig) GetSecuritySettings(ctx context.Context, brandID *uuid.UUID) (SecuritySettings, error) {
	s.log.Info("Getting security settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return SecuritySettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "security_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No security settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return SecuritySettings{
				SessionTimeout:         30,
				MaxLoginAttempts:       5,
				LockoutDuration:        15,
				TwoFactorRequired:      false,
				PasswordMinLength:      8,
				PasswordRequireSpecial: true,
				IpWhitelistEnabled:     false,
				RateLimitEnabled:       true,
				RateLimitRequests:      100,
			}, nil
		}
		s.log.Error("Failed to get security settings from database", zap.Error(err))
		return SecuritySettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get security settings")
	}

	var settings SecuritySettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal security settings", zap.Error(err))
		return SecuritySettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse security settings")
	}

	return settings, nil
}

// UpdateSecuritySettings updates security settings in system config
// If brandID is nil, updates ALL brands. If brandID is set, updates only that brand.
func (s *SystemConfig) UpdateSecuritySettings(ctx context.Context, settings SecuritySettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	s.log.Info("Updating security settings", zap.Any("brand_id", brandID))

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal security settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal security settings")
	}

	if brandID == nil {
		// Update ALL brands when brandID is nil (global update)
		_, err = s.db.GetPool().Exec(ctx, `
			WITH brand_ids AS (
				SELECT id FROM brands
			)
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			SELECT 
				$1,
				$2,
				$3,
				bi.id,
				$4,
				NOW()
			FROM brand_ids bi
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "security_settings", configValue, "Security and authentication settings", adminID)

		if err != nil {
			s.log.Error("Failed to update security settings for all brands", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to update security settings for all brands")
		}
		s.log.Info("Security settings updated for all brands")
	} else {
		// Update specific brand
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "security_settings", configValue, "Security and authentication settings", brandID, adminID)

		if err != nil {
			s.log.Error("Failed to update security settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
			return errors.ErrInternalServerError.Wrap(err, "failed to update security settings")
		}
		s.log.Info("Security settings updated for specific brand", zap.Any("brand_id", brandID))
	}

	return nil
}

// GetGeoBlockingSettings retrieves geo blocking settings from system config
func (s *SystemConfig) GetGeoBlockingSettings(ctx context.Context, brandID *uuid.UUID) (GeoBlockingSettings, error) {
	s.log.Info("Getting geo blocking settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return GeoBlockingSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "geo_blocking_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No geo blocking settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return GeoBlockingSettings{
				EnableGeoBlocking: false,
				DefaultAction:     "allow",
				VpnDetection:      false,
				ProxyDetection:    false,
				TorBlocking:       false,
				LogAttempts:       true,
				BlockedCountries:  []string{},
				AllowedCountries:  []string{},
				BypassCountries:   []string{},
			}, nil
		}
		s.log.Error("Failed to get geo blocking settings from database", zap.Error(err))
		return GeoBlockingSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get geo blocking settings")
	}

	var settings GeoBlockingSettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal geo blocking settings", zap.Error(err))
		return GeoBlockingSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse geo blocking settings")
	}

	return settings, nil
}

// UpdateGeoBlockingSettings updates geo blocking settings in system config
// If brandID is nil, updates ALL brands. If brandID is set, updates only that brand.
func (s *SystemConfig) UpdateGeoBlockingSettings(ctx context.Context, settings GeoBlockingSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	s.log.Info("Updating geo blocking settings", zap.Any("brand_id", brandID))

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal geo blocking settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal geo blocking settings")
	}

	if brandID == nil {
		// Update ALL brands when brandID is nil (global update)
		_, err = s.db.GetPool().Exec(ctx, `
			WITH brand_ids AS (
				SELECT id FROM brands
			)
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			SELECT 
				$1,
				$2,
				$3,
				bi.id,
				$4,
				NOW()
			FROM brand_ids bi
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "geo_blocking_settings", configValue, "Geo blocking and location-based access control settings", adminID)

		if err != nil {
			s.log.Error("Failed to update geo blocking settings for all brands", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to update geo blocking settings for all brands")
		}
		s.log.Info("Geo blocking settings updated for all brands")
	} else {
		// Update specific brand
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "geo_blocking_settings", configValue, "Geo blocking and location-based access control settings", brandID, adminID)

		if err != nil {
			s.log.Error("Failed to update geo blocking settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
			return errors.ErrInternalServerError.Wrap(err, "failed to update geo blocking settings")
		}
		s.log.Info("Geo blocking settings updated for specific brand", zap.Any("brand_id", brandID))
	}

	return nil
}
