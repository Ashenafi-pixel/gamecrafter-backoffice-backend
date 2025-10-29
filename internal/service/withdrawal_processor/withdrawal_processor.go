package withdrawal_processor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/system_config"
	"github.com/tucanbit/internal/storage/withdrawal_management"
	"go.uber.org/zap"
)

type WithdrawalProcessor struct {
	db                   *persistencedb.PersistenceDB
	systemConfig         *system_config.SystemConfig
	withdrawalManagement *withdrawal_management.WithdrawalManagement
	log                  *zap.Logger
}

func NewWithdrawalProcessor(db *persistencedb.PersistenceDB, log *zap.Logger) *WithdrawalProcessor {
	return &WithdrawalProcessor{
		db:                   db,
		systemConfig:         system_config.NewSystemConfig(db, log),
		withdrawalManagement: withdrawal_management.NewWithdrawalManagement(db, log),
		log:                  log,
	}
}

// ProcessWithdrawalRequest processes a withdrawal request and checks if it should be paused
func (w *WithdrawalProcessor) ProcessWithdrawalRequest(ctx context.Context, withdrawalID string, userID uuid.UUID, amountCents int64, currency string) error {
	w.log.Info("Processing withdrawal request",
		zap.String("withdrawal_id", withdrawalID),
		zap.String("user_id", userID.String()),
		zap.Int64("amount_cents", amountCents),
		zap.String("currency", currency))

	// 1. Check if withdrawals are globally enabled
	allowed, reason, err := w.systemConfig.CheckWithdrawalAllowed(ctx)
	if err != nil {
		w.log.Error("Failed to check withdrawal global status", zap.Error(err))
		return err
	}

	if !allowed {
		w.log.Info("Withdrawal blocked - global pause", zap.String("reason", reason))
		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, nil, nil)
	}

	// 2. Check single transaction threshold
	amountUSD := float64(amountCents) / 100.0 // Convert cents to USD
	exceeds, reason, err := w.systemConfig.CheckWithdrawalThresholds(ctx, amountUSD, currency, "single_transaction")
	if err != nil {
		w.log.Error("Failed to check single transaction threshold", zap.Error(err))
		return err
	}

	if exceeds {
		w.log.Info("Withdrawal blocked - single transaction threshold exceeded", zap.String("reason", reason))
		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, stringPtr("single_transaction"), &amountUSD)
	}

	// 3. Check user daily threshold
	exceeds, reason, err = w.systemConfig.CheckWithdrawalThresholds(ctx, amountUSD, currency, "user_daily")
	if err != nil {
		w.log.Error("Failed to check user daily threshold", zap.Error(err))
		return err
	}

	if exceeds {
		w.log.Info("Withdrawal blocked - user daily threshold exceeded", zap.String("reason", reason))
		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, stringPtr("user_daily"), &amountUSD)
	}

	// 4. Check hourly volume threshold
	exceeds, reason, err = w.systemConfig.CheckWithdrawalThresholds(ctx, amountUSD, currency, "hourly_volume")
	if err != nil {
		w.log.Error("Failed to check hourly volume threshold", zap.Error(err))
		return err
	}

	if exceeds {
		w.log.Info("Withdrawal blocked - hourly volume threshold exceeded", zap.String("reason", reason))
		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, stringPtr("hourly_volume"), &amountUSD)
	}

	// 5. Check daily volume threshold
	exceeds, reason, err = w.systemConfig.CheckWithdrawalThresholds(ctx, amountUSD, currency, "daily_volume")
	if err != nil {
		w.log.Error("Failed to check daily volume threshold", zap.Error(err))
		return err
	}

	if exceeds {
		w.log.Info("Withdrawal blocked - daily volume threshold exceeded", zap.String("reason", reason))
		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, stringPtr("daily_volume"), &amountUSD)
	}

	// 6. Check manual review requirements
	manualReview, err := w.systemConfig.GetWithdrawalManualReview(ctx)
	if err != nil {
		w.log.Error("Failed to get manual review settings", zap.Error(err))
		return err
	}

	if manualReview.Enabled && amountUSD >= manualReview.ThresholdAmount {
		w.log.Info("Withdrawal requires manual review",
			zap.Float64("amount", amountUSD),
			zap.Float64("threshold", manualReview.ThresholdAmount))

		reason := fmt.Sprintf("Withdrawal amount %.2f %s requires manual review (threshold: %.2f %s)",
			amountUSD, currency, manualReview.ThresholdAmount, manualReview.Currency)

		return w.withdrawalManagement.PauseWithdrawal(ctx, withdrawalID, reason, nil, true, stringPtr("manual_review"), &amountUSD)
	}

	// If we get here, the withdrawal can proceed
	w.log.Info("Withdrawal approved for processing", zap.String("withdrawal_id", withdrawalID))
	return nil
}

// CheckVolumeThresholds checks if current volumes exceed thresholds and pauses withdrawals if needed
func (w *WithdrawalProcessor) CheckVolumeThresholds(ctx context.Context) error {
	w.log.Info("Checking volume thresholds")

	// Get current thresholds
	thresholds, err := w.systemConfig.GetWithdrawalThresholds(ctx)
	if err != nil {
		w.log.Error("Failed to get withdrawal thresholds", zap.Error(err))
		return err
	}

	// Check hourly volume
	if thresholds.HourlyVolume.Active {
		hourlyVolume, err := w.getHourlyWithdrawalVolume(ctx)
		if err != nil {
			w.log.Error("Failed to get hourly volume", zap.Error(err))
			return err
		}

		if hourlyVolume > thresholds.HourlyVolume.Value {
			w.log.Info("Hourly volume threshold exceeded, pausing new withdrawals",
				zap.Float64("current_volume", hourlyVolume),
				zap.Float64("threshold", thresholds.HourlyVolume.Value))

			// Pause global withdrawals
			status, err := w.systemConfig.GetWithdrawalGlobalStatus(ctx)
			if err != nil {
				return err
			}

			status.Enabled = false
			reason := fmt.Sprintf("Hourly volume threshold exceeded: %.2f %s (limit: %.2f %s)",
				hourlyVolume, thresholds.HourlyVolume.Currency,
				thresholds.HourlyVolume.Value, thresholds.HourlyVolume.Currency)
			status.Reason = &reason
			now := time.Now().Format(time.RFC3339)
			status.PausedAt = &now

			return w.systemConfig.UpdateWithdrawalGlobalStatus(ctx, status, uuid.Nil)
		}
	}

	// Check daily volume
	if thresholds.DailyVolume.Active {
		dailyVolume, err := w.getDailyWithdrawalVolume(ctx)
		if err != nil {
			w.log.Error("Failed to get daily volume", zap.Error(err))
			return err
		}

		if dailyVolume > thresholds.DailyVolume.Value {
			w.log.Info("Daily volume threshold exceeded, pausing new withdrawals",
				zap.Float64("current_volume", dailyVolume),
				zap.Float64("threshold", thresholds.DailyVolume.Value))

			// Pause global withdrawals
			status, err := w.systemConfig.GetWithdrawalGlobalStatus(ctx)
			if err != nil {
				return err
			}

			status.Enabled = false
			reason := fmt.Sprintf("Daily volume threshold exceeded: %.2f %s (limit: %.2f %s)",
				dailyVolume, thresholds.DailyVolume.Currency,
				thresholds.DailyVolume.Value, thresholds.DailyVolume.Currency)
			status.Reason = &reason
			now := time.Now().Format(time.RFC3339)
			status.PausedAt = &now

			return w.systemConfig.UpdateWithdrawalGlobalStatus(ctx, status, uuid.Nil)
		}
	}

	return nil
}

// Helper functions
func (w *WithdrawalProcessor) getHourlyWithdrawalVolume(ctx context.Context) (float64, error) {
	// This would query the database for hourly withdrawal volume
	// For now, return 0 as placeholder
	return 0, nil
}

func (w *WithdrawalProcessor) getDailyWithdrawalVolume(ctx context.Context) (float64, error) {
	// This would query the database for daily withdrawal volume
	// For now, return 0 as placeholder
	return 0, nil
}

func stringPtr(s string) *string {
	return &s
}
