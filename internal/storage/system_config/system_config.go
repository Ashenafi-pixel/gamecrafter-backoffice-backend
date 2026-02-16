package system_config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

type WelcomeBonusSettings struct {
	// Legacy fields for backward compatibility (deprecated, use fixed_enabled and percentage_enabled instead)
	Type    string `json:"type,omitempty"`    // "fixed" or "percentage" (deprecated)
	Enabled bool   `json:"enabled,omitempty"` // true if this bonus type is enabled (deprecated)

	// New fields: separate toggles for fixed and percentage bonuses
	FixedEnabled      bool `json:"fixed_enabled"`      // true if fixed amount bonus is enabled
	PercentageEnabled bool `json:"percentage_enabled"` // true if percentage-based bonus is enabled

	// Anti-abuse / IP based protection
	// When ip_restriction_enabled = true and allow_multiple_bonuses_per_ip = false,
	// only the first account created from a given IP should receive a welcome bonus.
	// Enforcement is done by the signup backend; this struct only exposes configuration.
	IPRestrictionEnabled      bool `json:"ip_restriction_enabled"`        // enable/disable IP based restriction
	AllowMultipleBonusesPerIP bool `json:"allow_multiple_bonuses_per_ip"` // if false, only one welcome bonus per IP

	FixedAmount        float64 `json:"fixed_amount"`         // fixed bonus amount (for fixed type)
	Percentage         float64 `json:"percentage"`           // percentage value e.g., 50 for 50% (for percentage type)
	MaxDepositAmount   float64 `json:"max_deposit_amount"`   // maximum deposit amount for bonus calculation (for percentage type)
	MaxBonusPercentage float64 `json:"max_bonus_percentage"` // max bonus as % of deposit, default 90% to prevent 100% match
}

type WelcomeBonusChannelRule struct {
	ID                        string   `json:"id"`
	Channel                   string   `json:"channel"`
	ReferrerPatterns          []string `json:"referrer_patterns"`
	Enabled                   bool     `json:"enabled"`
	BonusType                 string   `json:"bonus_type"`
	FixedAmount               float64  `json:"fixed_amount"`
	Percentage                float64  `json:"percentage"`
	MaxBonusPercentage        float64  `json:"max_bonus_percentage"`
	MaxDepositAmount          float64  `json:"max_deposit_amount"`
	InheritIPPolicy           bool     `json:"inherit_ip_policy"`
	IPRestrictionEnabled      bool     `json:"ip_restriction_enabled"`
	AllowMultipleBonusesPerIP bool     `json:"allow_multiple_bonuses_per_ip"`
}

type WelcomeBonusChannelSettings struct {
	Channels []WelcomeBonusChannelRule `json:"channels"`
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

// GetWelcomeBonusSettings retrieves welcome bonus settings from system config
func (s *SystemConfig) GetWelcomeBonusSettings(ctx context.Context, brandID *uuid.UUID) (WelcomeBonusSettings, error) {
	s.log.Info("Getting welcome bonus settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting settings")
		return WelcomeBonusSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	// Get brand-specific settings (brand_id is required, no fallback)
	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "welcome_bonus_settings", brandID).Scan(&configValue)

	// Handle case where no configuration exists yet
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No welcome bonus settings found, returning defaults", zap.Any("brand_id", brandID))
			// Return default values when no configuration exists
			return WelcomeBonusSettings{
				FixedEnabled:              false,
				PercentageEnabled:         false,
				IPRestrictionEnabled:      true,  // by default, enable IP restriction
				AllowMultipleBonusesPerIP: false, // by default, only one bonus per IP
				FixedAmount:               0.0,
				Percentage:                0.0,
				MaxDepositAmount:          0.0,
				MaxBonusPercentage:        90.0,
			}, nil
		}
		s.log.Error("Failed to get welcome bonus settings from database", zap.Error(err))
		return WelcomeBonusSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to get welcome bonus settings")
	}

	var settings WelcomeBonusSettings
	err = json.Unmarshal(configValue, &settings)
	if err != nil {
		s.log.Error("Failed to unmarshal welcome bonus settings", zap.Error(err))
		return WelcomeBonusSettings{}, errors.ErrInternalServerError.Wrap(err, "failed to parse welcome bonus settings")
	}

	// Backward compatibility: migrate old type/enabled to new fixed_enabled/percentage_enabled
	if settings.Type != "" && !settings.FixedEnabled && !settings.PercentageEnabled {
		if settings.Type == "fixed" && settings.Enabled {
			settings.FixedEnabled = true
		} else if settings.Type == "percentage" && settings.Enabled {
			settings.PercentageEnabled = true
		}
	}

	// Backward compatibility: migrate old min_deposit_amount to max_deposit_amount
	// When JSON has min_deposit_amount, it won't populate MaxDepositAmount (different JSON tag)
	// So we need to check the raw JSON and copy the value
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(configValue, &rawSettings); err == nil {
		if minDepositVal, exists := rawSettings["min_deposit_amount"]; exists {
			if minDeposit, ok := minDepositVal.(float64); ok {
				// Always migrate if min_deposit_amount exists (even if max_deposit_amount was also set)
				settings.MaxDepositAmount = minDeposit
				s.log.Info("Migrated min_deposit_amount to max_deposit_amount", zap.Float64("value", minDeposit))
			}
		}
		// Also check if max_deposit_amount exists in raw JSON (in case it was set directly)
		if maxDepositVal, exists := rawSettings["max_deposit_amount"]; exists {
			if maxDeposit, ok := maxDepositVal.(float64); ok {
				settings.MaxDepositAmount = maxDeposit
			}
		}
	}

	return settings, nil
}

// UpdateWelcomeBonusSettings updates welcome bonus settings in system config for a specific brand
// brandID is required - checks if welcome_bonus_settings exists for that brand_id, updates if exists, inserts if not
func (s *SystemConfig) UpdateWelcomeBonusSettings(ctx context.Context, settings WelcomeBonusSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	if brandID == nil {
		s.log.Error("brand_id is required for welcome bonus settings")
		return errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required for welcome bonus settings")
	}

	s.log.Info("Updating welcome bonus settings", zap.Any("brand_id", brandID))

	// Backward compatibility: migrate old type/enabled to new fixed_enabled/percentage_enabled
	if settings.Type != "" {
		if settings.Type == "fixed" && settings.Enabled {
			settings.FixedEnabled = true
		} else if settings.Type == "percentage" && settings.Enabled {
			settings.PercentageEnabled = true
		}
	}

	// Validate that max_bonus_percentage < 100 for percentage type
	if settings.PercentageEnabled {
		if settings.MaxBonusPercentage >= 100 {
			s.log.Error("max_bonus_percentage must be less than 100 to prevent bonus from equaling deposit")
			return errors.ErrInvalidUserInput.Wrap(nil, "max_bonus_percentage must be less than 100")
		}
	}

	// Ensure we're using max_deposit_amount (not min_deposit_amount) in the saved JSON
	// Create a clean map to ensure old field names are not included
	cleanSettings := map[string]interface{}{
		"fixed_enabled":                 settings.FixedEnabled,
		"percentage_enabled":            settings.PercentageEnabled,
		"ip_restriction_enabled":        settings.IPRestrictionEnabled,
		"allow_multiple_bonuses_per_ip": settings.AllowMultipleBonusesPerIP,
		"fixed_amount":                  settings.FixedAmount,
		"percentage":                    settings.Percentage,
		"max_deposit_amount":            settings.MaxDepositAmount, // Always use max_deposit_amount
		"max_bonus_percentage":          settings.MaxBonusPercentage,
	}
	// Include legacy fields for backward compatibility if they exist
	if settings.Type != "" {
		cleanSettings["type"] = settings.Type
	}
	if settings.Enabled {
		cleanSettings["enabled"] = settings.Enabled
	}

	configValue, err := json.Marshal(cleanSettings)
	if err != nil {
		s.log.Error("Failed to marshal welcome bonus settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal welcome bonus settings")
	}

	// Check if welcome_bonus_settings exists for this brand_id
	var exists bool
	err = s.db.GetPool().QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM system_config 
			WHERE config_key = $1 AND brand_id = $2
		)
	`, "welcome_bonus_settings", brandID).Scan(&exists)

	if err != nil {
		s.log.Error("Failed to check if welcome bonus settings exist", zap.Error(err), zap.Any("brand_id", brandID))
		return errors.ErrInternalServerError.Wrap(err, "failed to check welcome bonus settings")
	}

	// Check if the admin user exists in the database
	var userExists bool
	err = s.db.GetPool().QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, adminID).Scan(&userExists)

	if err != nil {
		s.log.Warn("Failed to check if admin user exists, proceeding with NULL updated_by", zap.Error(err), zap.Any("admin_id", adminID))
		userExists = false
	}

	// Insert or update based on existence
	// ON CONFLICT will handle the update if it exists, or insert if it doesn't
	// Only set updated_by if user exists, otherwise keep existing value or set to NULL
	if userExists {
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "welcome_bonus_settings", configValue, "Welcome bonus settings", brandID, adminID)
	} else {
		s.log.Warn("Admin user not found in database, setting updated_by to NULL", zap.Any("admin_id", adminID))
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, NULL, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = system_config.updated_by,
				updated_at = NOW()
		`, "welcome_bonus_settings", configValue, "Welcome bonus settings", brandID)
	}

	if err != nil {
		s.log.Error("Failed to update welcome bonus settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
		return errors.ErrInternalServerError.Wrap(err, "failed to update welcome bonus settings")
	}

	if exists {
		s.log.Info("Welcome bonus settings updated for brand", zap.Any("brand_id", brandID))
	} else {
		s.log.Info("Welcome bonus settings created for brand", zap.Any("brand_id", brandID))
	}

	return nil
}

func (s *SystemConfig) GetWelcomeBonusChannelSettings(ctx context.Context, brandID *uuid.UUID) (WelcomeBonusChannelSettings, error) {
	s.log.Info("Getting welcome bonus channel settings from system config", zap.Any("brand_id", brandID))

	if brandID == nil {
		s.log.Error("brand_id is required for getting channel settings")
		return WelcomeBonusChannelSettings{}, errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required")
	}

	var configValue json.RawMessage
	err := s.db.GetPool().QueryRow(ctx, `
		SELECT config_value FROM system_config 
		WHERE config_key = $1 AND brand_id = $2
	`, "welcome_bonus_channel_settings", brandID).Scan(&configValue)

	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No welcome bonus channel settings found, returning empty channels", zap.Any("brand_id", brandID))
			return WelcomeBonusChannelSettings{
				Channels: []WelcomeBonusChannelRule{},
			}, nil
		}
		s.log.Error("Failed to get welcome bonus channel settings", zap.Error(err))
		return WelcomeBonusChannelSettings{}, errors.ErrUnableToGet.Wrap(err, "failed to get welcome bonus channel settings")
	}

	var settings WelcomeBonusChannelSettings
	if err := json.Unmarshal(configValue, &settings); err != nil {
		s.log.Error("Failed to unmarshal welcome bonus channel settings", zap.Error(err))
		return WelcomeBonusChannelSettings{}, errors.ErrUnableToGet.Wrap(err, "failed to unmarshal welcome bonus channel settings")
	}

	if settings.Channels == nil {
		settings.Channels = []WelcomeBonusChannelRule{}
	}

	return settings, nil
}

func (s *SystemConfig) UpdateWelcomeBonusChannelSettings(ctx context.Context, settings WelcomeBonusChannelSettings, adminID uuid.UUID, brandID *uuid.UUID) error {
	if brandID == nil {
		s.log.Error("brand_id is required for welcome bonus channel settings")
		return errors.ErrInvalidUserInput.Wrap(nil, "brand_id is required for welcome bonus channel settings")
	}

	s.log.Info("Updating welcome bonus channel settings", zap.Any("brand_id", brandID))

	if settings.Channels == nil {
		settings.Channels = []WelcomeBonusChannelRule{}
	}

	for i := range settings.Channels {
		if settings.Channels[i].BonusType == "percentage" && settings.Channels[i].MaxBonusPercentage >= 100 {
			s.log.Error("max_bonus_percentage must be less than 100", zap.String("channel_id", settings.Channels[i].ID))
			return errors.ErrInvalidUserInput.Wrap(nil, "max_bonus_percentage must be less than 100")
		}
	}

	configValue, err := json.Marshal(settings)
	if err != nil {
		s.log.Error("Failed to marshal welcome bonus channel settings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to marshal welcome bonus channel settings")
	}

	var userExists bool
	err = s.db.GetPool().QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, adminID).Scan(&userExists)

	if err != nil {
		s.log.Warn("Failed to check if admin user exists, proceeding with NULL updated_by", zap.Error(err), zap.Any("admin_id", adminID))
		userExists = false
	}

	if userExists {
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = EXCLUDED.updated_by,
				updated_at = NOW()
		`, "welcome_bonus_channel_settings", configValue, "Welcome bonus channel settings", brandID, adminID)
	} else {
		s.log.Warn("Admin user not found in database, setting updated_by to NULL", zap.Any("admin_id", adminID))
		_, err = s.db.GetPool().Exec(ctx, `
			INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, NULL, NOW())
			ON CONFLICT (brand_id, config_key) 
			DO UPDATE SET 
				config_value = EXCLUDED.config_value,
				updated_by = system_config.updated_by,
				updated_at = NOW()
		`, "welcome_bonus_channel_settings", configValue, "Welcome bonus channel settings", brandID)
	}

	if err != nil {
		s.log.Error("Failed to update welcome bonus channel settings for brand", zap.Error(err), zap.Any("brand_id", brandID))
		return errors.ErrInternalServerError.Wrap(err, "failed to update welcome bonus channel settings")
	}

	s.log.Info("Welcome bonus channel settings updated for brand", zap.Any("brand_id", brandID))
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

type GameImportConfig struct {
	ID                    uuid.UUID  `json:"id"`
	BrandID               uuid.UUID  `json:"brand_id"`
	ScheduleType          string     `json:"schedule_type"` // daily, weekly, monthly, custom
	ScheduleCron          *string    `json:"schedule_cron,omitempty"`
	Providers             []string   `json:"providers,omitempty"` // null or array of provider names
	DirectusURL           *string    `json:"directus_url,omitempty"`
	CheckFrequencyMinutes *int       `json:"check_frequency_minutes,omitempty"`
	IsActive              bool       `json:"is_active"`
	LastRunAt             *time.Time `json:"last_run_at,omitempty"`
	NextRunAt             *time.Time `json:"next_run_at,omitempty"`
	LastCheckAt           *time.Time `json:"last_check_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (s *SystemConfig) GetGameImportConfig(ctx context.Context, brandID uuid.UUID) (*GameImportConfig, error) {
	s.log.Info("Getting game import config", zap.String("brand_id", brandID.String()))

	var config GameImportConfig
	var providersJSON []byte
	var checkFreqMinutes *int

	err := s.db.GetPool().QueryRow(ctx, `
		SELECT id, brand_id, schedule_type, schedule_cron, providers, directus_url, 
		       check_frequency_minutes, is_active, last_run_at, next_run_at, 
		       last_check_at, created_at, updated_at
		FROM game_import_config
		WHERE brand_id = $1
	`, brandID).Scan(
		&config.ID, &config.BrandID, &config.ScheduleType, &config.ScheduleCron,
		&providersJSON, &config.DirectusURL, &checkFreqMinutes, &config.IsActive,
		&config.LastRunAt, &config.NextRunAt, &config.LastCheckAt,
		&config.CreatedAt, &config.UpdatedAt,
	)

	// Default check_frequency_minutes to 15 if NULL
	if checkFreqMinutes == nil {
		defaultFreq := 15
		config.CheckFrequencyMinutes = &defaultFreq
	} else {
		config.CheckFrequencyMinutes = checkFreqMinutes
	}

	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			s.log.Info("No game import config found, returning defaults", zap.String("brand_id", brandID.String()))
			// Return default configuration
			defaultFreq := 15
			return &GameImportConfig{
				ID:                    uuid.Nil,
				BrandID:               brandID,
				ScheduleType:          "daily",
				IsActive:              false,
				Providers:             nil, // nil means all providers
				CheckFrequencyMinutes: &defaultFreq,
			}, nil
		}
		s.log.Error("Failed to get game import config", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game import config")
	}

	// Parse providers JSON
	if len(providersJSON) > 0 && string(providersJSON) != "null" {
		if err := json.Unmarshal(providersJSON, &config.Providers); err != nil {
			s.log.Warn("Failed to parse providers JSON, using empty array", zap.Error(err))
			config.Providers = nil
		}
	}

	return &config, nil
}

// updates or creates game import configuration
func (s *SystemConfig) UpdateGameImportConfig(ctx context.Context, config GameImportConfig) error {
	s.log.Info("Updating game import config", zap.String("brand_id", config.BrandID.String()))

	// Validate schedule_type
	validScheduleTypes := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"custom":  true,
	}
	if !validScheduleTypes[config.ScheduleType] {
		return errors.ErrInvalidUserInput.Wrap(nil, "invalid schedule_type, must be: daily, weekly, monthly, or custom")
	}

	// Validate custom cron if schedule_type is custom
	if config.ScheduleType == "custom" && (config.ScheduleCron == nil || *config.ScheduleCron == "") {
		return errors.ErrInvalidUserInput.Wrap(nil, "schedule_cron is required when schedule_type is custom")
	}

	// Validate check_frequency_minutes
	if config.CheckFrequencyMinutes != nil {
		if *config.CheckFrequencyMinutes < 1 {
			return errors.ErrInvalidUserInput.Wrap(nil, "check_frequency_minutes must be at least 1")
		}
		if *config.CheckFrequencyMinutes > 10080 {
			return errors.ErrInvalidUserInput.Wrap(nil, "check_frequency_minutes cannot exceed 10080 (7 days)")
		}
	}

	// Marshal providers to JSON
	var providersJSON []byte
	var err error
	if len(config.Providers) == 0 {
		providersJSON = []byte("null")
	} else {
		providersJSON, err = json.Marshal(config.Providers)
		if err != nil {
			s.log.Error("Failed to marshal providers", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to marshal providers")
		}
	}

	// Default check_frequency_minutes to 15 if not provided
	checkFreq := config.CheckFrequencyMinutes
	if checkFreq == nil {
		defaultFreq := 15
		checkFreq = &defaultFreq
	}

	_, err = s.db.GetPool().Exec(ctx, `
		INSERT INTO game_import_config (brand_id, schedule_type, schedule_cron, providers, directus_url, check_frequency_minutes, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (brand_id)
		DO UPDATE SET
			schedule_type = EXCLUDED.schedule_type,
			schedule_cron = EXCLUDED.schedule_cron,
			providers = EXCLUDED.providers,
			directus_url = EXCLUDED.directus_url,
			check_frequency_minutes = EXCLUDED.check_frequency_minutes,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`, config.BrandID, config.ScheduleType, config.ScheduleCron, providersJSON, config.DirectusURL, checkFreq, config.IsActive)

	if err != nil {
		s.log.Error("Failed to update game import config", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update game import config")
	}

	s.log.Info("Game import config updated successfully", zap.String("brand_id", config.BrandID.String()))
	return nil
}

// updates the last_run_at and next_run_at timestamps
func (s *SystemConfig) UpdateGameImportLastRun(ctx context.Context, brandID uuid.UUID, lastRunAt time.Time, nextRunAt *time.Time) error {
	_, err := s.db.GetPool().Exec(ctx, `
		UPDATE game_import_config
		SET last_run_at = $1, next_run_at = $2, updated_at = NOW()
		WHERE brand_id = $3
	`, lastRunAt, nextRunAt, brandID)

	if err != nil {
		s.log.Error("Failed to update game import last run", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update game import last run")
	}

	return nil
}

// UpdateGameImportLastCheck updates the last_check_at timestamp for a brand
func (s *SystemConfig) UpdateGameImportLastCheck(ctx context.Context, brandID uuid.UUID, lastCheckAt time.Time) error {
	_, err := s.db.GetPool().Exec(ctx, `
		UPDATE game_import_config
		SET last_check_at = $1, updated_at = NOW()
		WHERE brand_id = $2
	`, lastCheckAt, brandID)

	if err != nil {
		s.log.Error("Failed to update game import last check", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update game import last check")
	}

	return nil
}

// GetAllActiveGameImportConfigs retrieves all active game import configurations
func (s *SystemConfig) GetAllActiveGameImportConfigs(ctx context.Context) ([]GameImportConfig, error) {
	s.log.Info("Getting all active game import configs")

	rows, err := s.db.GetPool().Query(ctx, `
		SELECT id, brand_id, schedule_type, schedule_cron, providers, directus_url, 
		       check_frequency_minutes, is_active, last_run_at, next_run_at, 
		       last_check_at, created_at, updated_at
		FROM game_import_config
		WHERE is_active = true
	`)
	if err != nil {
		s.log.Error("Failed to get active game import configs", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get active game import configs")
	}
	defer rows.Close()

	var configs []GameImportConfig
	for rows.Next() {
		var config GameImportConfig
		var providersJSON []byte
		var checkFreqMinutes *int

		err := rows.Scan(
			&config.ID, &config.BrandID, &config.ScheduleType, &config.ScheduleCron,
			&providersJSON, &config.DirectusURL, &checkFreqMinutes, &config.IsActive,
			&config.LastRunAt, &config.NextRunAt, &config.LastCheckAt,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			s.log.Error("Failed to scan game import config", zap.Error(err))
			continue
		}

		// Default check_frequency_minutes to 15 if NULL
		if checkFreqMinutes == nil {
			defaultFreq := 15
			config.CheckFrequencyMinutes = &defaultFreq
		} else {
			config.CheckFrequencyMinutes = checkFreqMinutes
		}

		// Parse providers JSON
		if len(providersJSON) > 0 && string(providersJSON) != "null" {
			if err := json.Unmarshal(providersJSON, &config.Providers); err != nil {
				s.log.Warn("Failed to parse providers JSON", zap.Error(err))
				config.Providers = nil
			}
		}

		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		s.log.Error("Error iterating game import configs", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "error iterating game import configs")
	}

	s.log.Info("Retrieved active game import configs", zap.Int("count", len(configs)))
	return configs, nil
}
