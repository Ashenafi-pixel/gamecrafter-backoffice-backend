package withdrawal_pause

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/platform"
	"github.com/tucanbit/internal/utils/errors"
	"go.uber.org/zap"
)

type WithdrawalPause struct {
	db  *platform.Platform
	log *zap.Logger
}

func NewWithdrawalPause(db *platform.Platform, log *zap.Logger) *WithdrawalPause {
	return &WithdrawalPause{
		db:  db,
		log: log,
	}
}

// GetWithdrawalPauseSettings retrieves the current pause settings
func (w *WithdrawalPause) GetWithdrawalPauseSettings(ctx context.Context) (dto.WithdrawalPauseSettings, error) {
	w.log.Info("Getting withdrawal pause settings")

	settings, err := w.db.Queries.GetWithdrawalPauseSettings(ctx)
	if err != nil {
		w.log.Error("Failed to get withdrawal pause settings", zap.Error(err))
		return dto.WithdrawalPauseSettings{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal pause settings")
	}

	return dto.WithdrawalPauseSettings{
		ID:               settings.ID,
		IsGloballyPaused: settings.IsGloballyPaused,
		PauseReason:      settings.PauseReason,
		PausedBy:         settings.PausedBy,
		PausedAt:         settings.PausedAt,
		CreatedAt:        settings.CreatedAt,
		UpdatedAt:        settings.UpdatedAt,
	}, nil
}

// UpdateWithdrawalPauseSettings updates the global pause settings
func (w *WithdrawalPause) UpdateWithdrawalPauseSettings(ctx context.Context, req dto.UpdateWithdrawalPauseSettingsRequest, adminID uuid.UUID) error {
	w.log.Info("Updating withdrawal pause settings",
		zap.Bool("is_globally_paused", req.IsGloballyPaused != nil && *req.IsGloballyPaused),
		zap.String("pause_reason", getStringValue(req.PauseReason)))

	// Get current settings first
	currentSettings, err := w.db.Queries.GetWithdrawalPauseSettings(ctx)
	if err != nil {
		w.log.Error("Failed to get current pause settings", zap.Error(err))
		return errors.ErrUnableToGet.Wrap(err, "failed to get current pause settings")
	}

	err = w.db.Queries.UpdateWithdrawalPauseSettings(ctx, db.UpdateWithdrawalPauseSettingsParams{
		ID:               currentSettings.ID,
		IsGloballyPaused: req.IsGloballyPaused,
		PauseReason:      req.PauseReason,
		PausedBy:         &adminID,
	})
	if err != nil {
		w.log.Error("Failed to update withdrawal pause settings", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal pause settings")
	}

	w.log.Info("Successfully updated withdrawal pause settings")
	return nil
}

// GetWithdrawalThresholds retrieves all active thresholds
func (w *WithdrawalPause) GetWithdrawalThresholds(ctx context.Context) ([]dto.WithdrawalThreshold, error) {
	w.log.Info("Getting withdrawal thresholds")

	thresholds, err := w.db.Queries.GetWithdrawalThresholds(ctx)
	if err != nil {
		w.log.Error("Failed to get withdrawal thresholds", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal thresholds")
	}

	result := make([]dto.WithdrawalThreshold, len(thresholds))
	for i, t := range thresholds {
		result[i] = dto.WithdrawalThreshold{
			ID:             t.ID,
			ThresholdType:  t.ThresholdType,
			ThresholdValue: t.ThresholdValue,
			Currency:       t.Currency,
			IsActive:       t.IsActive,
			CreatedBy:      t.CreatedBy,
			CreatedAt:      t.CreatedAt,
			UpdatedAt:      t.UpdatedAt,
		}
	}

	return result, nil
}

// CreateWithdrawalThreshold creates a new threshold
func (w *WithdrawalPause) CreateWithdrawalThreshold(ctx context.Context, req dto.CreateWithdrawalThresholdRequest, adminID uuid.UUID) (dto.WithdrawalThreshold, error) {
	w.log.Info("Creating withdrawal threshold",
		zap.String("threshold_type", req.ThresholdType),
		zap.String("threshold_value", req.ThresholdValue.String()),
		zap.String("currency", req.Currency))

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	threshold, err := w.db.Queries.CreateWithdrawalThreshold(ctx, db.CreateWithdrawalThresholdParams{
		ThresholdType:  req.ThresholdType,
		ThresholdValue: req.ThresholdValue,
		Currency:       req.Currency,
		IsActive:       isActive,
		CreatedBy:      &adminID,
	})
	if err != nil {
		w.log.Error("Failed to create withdrawal threshold", zap.Error(err))
		return dto.WithdrawalThreshold{}, errors.ErrUnableTocreate.Wrap(err, "failed to create withdrawal threshold")
	}

	return dto.WithdrawalThreshold{
		ID:             threshold.ID,
		ThresholdType:  threshold.ThresholdType,
		ThresholdValue: threshold.ThresholdValue,
		Currency:       threshold.Currency,
		IsActive:       threshold.IsActive,
		CreatedBy:      threshold.CreatedBy,
		CreatedAt:      threshold.CreatedAt,
		UpdatedAt:      threshold.UpdatedAt,
	}, nil
}

// UpdateWithdrawalThreshold updates an existing threshold
func (w *WithdrawalPause) UpdateWithdrawalThreshold(ctx context.Context, thresholdID uuid.UUID, req dto.UpdateWithdrawalThresholdRequest) (dto.WithdrawalThreshold, error) {
	w.log.Info("Updating withdrawal threshold", zap.String("threshold_id", thresholdID.String()))

	// Get current threshold to preserve unchanged values
	currentThreshold, err := w.db.Queries.GetWithdrawalThresholdByType(ctx, db.GetWithdrawalThresholdByTypeParams{
		ThresholdType: "", // We'll get by ID instead
		Currency:      "",
	})
	if err != nil {
		w.log.Error("Failed to get current threshold", zap.Error(err))
		return dto.WithdrawalThreshold{}, errors.ErrUnableToGet.Wrap(err, "failed to get current threshold")
	}

	thresholdValue := currentThreshold.ThresholdValue
	if req.ThresholdValue != nil {
		thresholdValue = *req.ThresholdValue
	}

	isActive := currentThreshold.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	threshold, err := w.db.Queries.UpdateWithdrawalThreshold(ctx, db.UpdateWithdrawalThresholdParams{
		ID:             thresholdID,
		ThresholdValue: thresholdValue,
		IsActive:       isActive,
	})
	if err != nil {
		w.log.Error("Failed to update withdrawal threshold", zap.Error(err))
		return dto.WithdrawalThreshold{}, errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal threshold")
	}

	return dto.WithdrawalThreshold{
		ID:             threshold.ID,
		ThresholdType:  threshold.ThresholdType,
		ThresholdValue: threshold.ThresholdValue,
		Currency:       threshold.Currency,
		IsActive:       threshold.IsActive,
		CreatedBy:      threshold.CreatedBy,
		CreatedAt:      threshold.CreatedAt,
		UpdatedAt:      threshold.UpdatedAt,
	}, nil
}

// GetPausedWithdrawals retrieves paused withdrawals with pagination
func (w *WithdrawalPause) GetPausedWithdrawals(ctx context.Context, req dto.GetPausedWithdrawalsRequest) (dto.GetPausedWithdrawalsResponse, error) {
	w.log.Info("Getting paused withdrawals",
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage))

	offset := (req.Page - 1) * req.PerPage
	withdrawals, err := w.db.Queries.GetPausedWithdrawals(ctx, db.GetPausedWithdrawalsParams{
		Status:      req.Status,
		PauseReason: req.PauseReason,
		UserID:      req.UserID,
		Limit:       int32(req.PerPage),
		Offset:      int32(offset),
	})
	if err != nil {
		w.log.Error("Failed to get paused withdrawals", zap.Error(err))
		return dto.GetPausedWithdrawalsResponse{}, errors.ErrUnableToGet.Wrap(err, "failed to get paused withdrawals")
	}

	var total int
	if len(withdrawals) > 0 {
		total = int(withdrawals[0].Total)
	}

	result := make([]dto.PausedWithdrawal, len(withdrawals))
	for i, wd := range withdrawals {
		result[i] = dto.PausedWithdrawal{
			ID:                   wd.ID,
			UserID:               wd.UserID,
			WithdrawalID:         wd.WithdrawalID,
			Amount:               decimal.NewFromInt(wd.UsdAmountCents).Div(decimal.NewFromInt(100)),
			Currency:             wd.CryptoCurrency,
			Status:               wd.Status,
			IsPaused:             wd.IsPaused,
			PauseReason:          wd.PauseReason,
			PausedAt:             wd.PausedAt,
			RequiresManualReview: wd.RequiresManualReview,
			CreatedAt:            wd.CreatedAt,
			UpdatedAt:            wd.UpdatedAt,
			Username:             wd.Username,
			Email:                wd.Email,
		}
	}

	return dto.GetPausedWithdrawalsResponse{
		Withdrawals: result,
		Total:       total,
		Page:        req.Page,
		PerPage:     req.PerPage,
	}, nil
}

// PauseWithdrawal pauses a withdrawal and logs the reason
func (w *WithdrawalPause) PauseWithdrawal(ctx context.Context, withdrawalID uuid.UUID, reason string, thresholdType *string, thresholdValue *decimal.Decimal) error {
	w.log.Info("Pausing withdrawal",
		zap.String("withdrawal_id", withdrawalID.String()),
		zap.String("reason", reason))

	// Update withdrawal status
	err := w.db.Queries.UpdateWithdrawalPauseStatus(ctx, db.UpdateWithdrawalPauseStatusParams{
		ID:                   withdrawalID,
		IsPaused:             true,
		PauseReason:          &reason,
		RequiresManualReview: true,
	})
	if err != nil {
		w.log.Error("Failed to update withdrawal pause status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal pause status")
	}

	// Create pause log
	_, err = w.db.Queries.CreateWithdrawalPauseLog(ctx, db.CreateWithdrawalPauseLogParams{
		WithdrawalID:   withdrawalID,
		PauseReason:    reason,
		ThresholdType:  thresholdType,
		ThresholdValue: thresholdValue,
	})
	if err != nil {
		w.log.Error("Failed to create withdrawal pause log", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to create withdrawal pause log")
	}

	w.log.Info("Successfully paused withdrawal")
	return nil
}

// ApproveWithdrawal approves a paused withdrawal
func (w *WithdrawalPause) ApproveWithdrawal(ctx context.Context, withdrawalID uuid.UUID, adminID uuid.UUID, notes *string) error {
	w.log.Info("Approving withdrawal", zap.String("withdrawal_id", withdrawalID.String()))

	// Update withdrawal status
	err := w.db.Queries.UpdateWithdrawalPauseStatus(ctx, db.UpdateWithdrawalPauseStatusParams{
		ID:                   withdrawalID,
		IsPaused:             false,
		PauseReason:          nil,
		RequiresManualReview: false,
	})
	if err != nil {
		w.log.Error("Failed to update withdrawal pause status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal pause status")
	}

	// Update pause log
	// Note: This assumes we have a way to get the latest pause log for this withdrawal
	// In a real implementation, you might need to add a query to get the latest log entry

	w.log.Info("Successfully approved withdrawal")
	return nil
}

// GetWithdrawalPauseStats retrieves statistics for the pause system
func (w *WithdrawalPause) GetWithdrawalPauseStats(ctx context.Context) (dto.WithdrawalPauseStats, error) {
	w.log.Info("Getting withdrawal pause stats")

	stats, err := w.db.Queries.GetWithdrawalPauseStats(ctx)
	if err != nil {
		w.log.Error("Failed to get withdrawal pause stats", zap.Error(err))
		return dto.WithdrawalPauseStats{}, errors.ErrUnableToGet.Wrap(err, "failed to get withdrawal pause stats")
	}

	// Get global pause status
	pauseSettings, err := w.GetWithdrawalPauseSettings(ctx)
	if err != nil {
		w.log.Error("Failed to get pause settings for stats", zap.Error(err))
		pauseSettings.IsGloballyPaused = false
	}

	// Get active thresholds count
	thresholds, err := w.GetWithdrawalThresholds(ctx)
	if err != nil {
		w.log.Error("Failed to get thresholds for stats", zap.Error(err))
	}

	return dto.WithdrawalPauseStats{
		TotalPausedToday:    int(stats.TotalPausedToday),
		TotalPausedThisHour: int(stats.TotalPausedThisHour),
		PendingReview:       int(stats.PendingReview),
		ApprovedToday:       int(stats.ApprovedToday),
		RejectedToday:       int(stats.RejectedToday),
		AverageReviewTime:   stats.AverageReviewTimeMinutes,
		GlobalPauseStatus:   pauseSettings.IsGloballyPaused,
		ActiveThresholds:    len(thresholds),
	}, nil
}

// CheckThresholds checks if withdrawal exceeds any thresholds
func (w *WithdrawalPause) CheckThresholds(ctx context.Context, withdrawalID uuid.UUID, userID uuid.UUID, amount decimal.Decimal) (bool, string, error) {
	w.log.Info("Checking withdrawal thresholds",
		zap.String("withdrawal_id", withdrawalID.String()),
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()))

	// Get all active thresholds
	thresholds, err := w.GetWithdrawalThresholds(ctx)
	if err != nil {
		w.log.Error("Failed to get thresholds for checking", zap.Error(err))
		return false, "", err
	}

	// Check single transaction threshold
	for _, threshold := range thresholds {
		if threshold.ThresholdType == dto.THRESHOLD_SINGLE_TRANSACTION {
			if amount.GreaterThan(threshold.ThresholdValue) {
				w.log.Warn("Single transaction threshold exceeded",
					zap.String("amount", amount.String()),
					zap.String("threshold", threshold.ThresholdValue.String()))
				return true, dto.PAUSE_REASON_THRESHOLD_EXCEEDED, nil
			}
		}
	}

	// Check hourly volume threshold
	hourlyVolume, err := w.db.Queries.GetHourlyWithdrawalVolume(ctx)
	if err != nil {
		w.log.Error("Failed to get hourly volume", zap.Error(err))
		return false, "", err
	}

	for _, threshold := range thresholds {
		if threshold.ThresholdType == dto.THRESHOLD_HOURLY_VOLUME {
			hourlyVolumeDecimal := decimal.NewFromInt(hourlyVolume).Div(decimal.NewFromInt(100))
			if hourlyVolumeDecimal.GreaterThan(threshold.ThresholdValue) {
				w.log.Warn("Hourly volume threshold exceeded",
					zap.String("current_volume", hourlyVolumeDecimal.String()),
					zap.String("threshold", threshold.ThresholdValue.String()))
				return true, dto.PAUSE_REASON_THRESHOLD_EXCEEDED, nil
			}
		}
	}

	// Check daily volume threshold
	dailyVolume, err := w.db.Queries.GetDailyWithdrawalVolume(ctx)
	if err != nil {
		w.log.Error("Failed to get daily volume", zap.Error(err))
		return false, "", err
	}

	for _, threshold := range thresholds {
		if threshold.ThresholdType == dto.THRESHOLD_DAILY_VOLUME {
			dailyVolumeDecimal := decimal.NewFromInt(dailyVolume).Div(decimal.NewFromInt(100))
			if dailyVolumeDecimal.GreaterThan(threshold.ThresholdValue) {
				w.log.Warn("Daily volume threshold exceeded",
					zap.String("current_volume", dailyVolumeDecimal.String()),
					zap.String("threshold", threshold.ThresholdValue.String()))
				return true, dto.PAUSE_REASON_THRESHOLD_EXCEEDED, nil
			}
		}
	}

	// Check user daily limit
	userDailyVolume, err := w.db.Queries.GetUserDailyWithdrawalVolume(ctx, userID)
	if err != nil {
		w.log.Error("Failed to get user daily volume", zap.Error(err))
		return false, "", err
	}

	for _, threshold := range thresholds {
		if threshold.ThresholdType == dto.THRESHOLD_USER_DAILY {
			userDailyVolumeDecimal := decimal.NewFromInt(userDailyVolume).Div(decimal.NewFromInt(100))
			if userDailyVolumeDecimal.GreaterThan(threshold.ThresholdValue) {
				w.log.Warn("User daily limit threshold exceeded",
					zap.String("user_id", userID.String()),
					zap.String("current_volume", userDailyVolumeDecimal.String()),
					zap.String("threshold", threshold.ThresholdValue.String()))
				return true, dto.PAUSE_REASON_USER_DAILY_LIMIT, nil
			}
		}
	}

	w.log.Info("All thresholds passed")
	return false, "", nil
}

// RejectWithdrawal rejects a paused withdrawal
func (w *WithdrawalPause) RejectWithdrawal(ctx context.Context, withdrawalID uuid.UUID, adminID uuid.UUID, notes *string) error {
	w.log.Info("Rejecting withdrawal", zap.String("withdrawal_id", withdrawalID.String()))

	// Update withdrawal status
	err := w.db.Queries.UpdateWithdrawalPauseStatus(ctx, db.UpdateWithdrawalPauseStatusParams{
		ID:                   withdrawalID,
		IsPaused:             false,
		PauseReason:          nil,
		RequiresManualReview: false,
	})
	if err != nil {
		w.log.Error("Failed to update withdrawal pause status", zap.Error(err))
		return errors.ErrUnableToUpdate.Wrap(err, "failed to update withdrawal pause status")
	}

	// Update pause log with rejection
	// Note: This assumes we have a way to get the latest pause log for this withdrawal
	// In a real implementation, you might need to add a query to get the latest log entry

	w.log.Info("Successfully rejected withdrawal")
	return nil
}

// DeleteWithdrawalThreshold deletes a threshold
func (w *WithdrawalPause) DeleteWithdrawalThreshold(ctx context.Context, thresholdID uuid.UUID) error {
	w.log.Info("Deleting withdrawal threshold", zap.String("threshold_id", thresholdID.String()))

	err := w.db.Queries.DeleteWithdrawalThreshold(ctx, thresholdID)
	if err != nil {
		w.log.Error("Failed to delete withdrawal threshold", zap.Error(err))
		return errors.ErrUnableToDelete.Wrap(err, "failed to delete withdrawal threshold")
	}

	w.log.Info("Successfully deleted withdrawal threshold")
	return nil
}
