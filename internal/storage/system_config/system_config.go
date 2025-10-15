package system_config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
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
