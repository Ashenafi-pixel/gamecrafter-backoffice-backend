package rakeback_override

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/rakeback_override"
	"go.uber.org/zap"
)

type RakebackOverride interface {
	GetActiveOverride(ctx context.Context) (*dto.GlobalRakebackOverride, error)
	GetOverride(ctx context.Context) (*dto.GlobalRakebackOverride, error)
	CreateOrUpdateOverride(ctx context.Context, req dto.CreateOrUpdateRakebackOverrideReq, adminID uuid.UUID) (*dto.GlobalRakebackOverride, error)
	ToggleOverride(ctx context.Context, isActive bool, adminID uuid.UUID) error
}

type rakebackOverride struct {
	storage rakeback_override.RakebackOverrideStorage
	log     *zap.Logger
}

func Init(storage rakeback_override.RakebackOverrideStorage, log *zap.Logger) RakebackOverride {
	return &rakebackOverride{
		storage: storage,
		log:     log,
	}
}

// GetActiveOverride retrieves the currently active global rakeback override
func (r *rakebackOverride) GetActiveOverride(ctx context.Context) (*dto.GlobalRakebackOverride, error) {
	override, err := r.storage.GetActiveOverride(ctx)
	if err != nil {
		r.log.Error("Failed to get active rakeback override", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get active rakeback override")
	}

	if override == nil {
		return nil, nil
	}

	return r.convertToDTO(override), nil
}

// GetOverride retrieves the most recent global rakeback override (active or not)
func (r *rakebackOverride) GetOverride(ctx context.Context) (*dto.GlobalRakebackOverride, error) {
	override, err := r.storage.GetOverride(ctx)
	if err != nil {
		r.log.Error("Failed to get rakeback override", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get rakeback override")
	}

	if override == nil {
		return nil, nil
	}

	return r.convertToDTO(override), nil
}

// CreateOrUpdateOverride creates a new override or updates an existing one
func (r *rakebackOverride) CreateOrUpdateOverride(ctx context.Context, req dto.CreateOrUpdateRakebackOverrideReq, adminID uuid.UUID) (*dto.GlobalRakebackOverride, error) {
	// Validate rakeback percentage (0-100, up to 2 decimal places)
	if req.RakebackPercentage.LessThan(decimal.Zero) || req.RakebackPercentage.GreaterThan(decimal.NewFromInt(100)) {
		err := errors.ErrInvalidUserInput.New("rakeback percentage must be between 0 and 100")
		return nil, err
	}

	// Get existing override
	existing, err := r.storage.GetOverride(ctx)
	if err != nil {
		r.log.Error("Failed to get existing override", zap.Error(err))
		return nil, errors.ErrUnableToGet.Wrap(err, "failed to get existing override")
	}

	var override *rakeback_override.GlobalRakebackOverride

	if existing != nil {
		// Update existing override
		override = &rakeback_override.GlobalRakebackOverride{
			ID:                existing.ID,
			IsActive:          req.IsActive,
			RakebackPercentage: req.RakebackPercentage,
			StartTime:         req.StartTime,
			EndTime:           req.EndTime,
			UpdatedBy:         &adminID,
		}

		override, err = r.storage.UpdateOverride(ctx, *override)
		if err != nil {
			r.log.Error("Failed to update rakeback override", zap.Error(err))
			return nil, errors.ErrUnableToUpdate.Wrap(err, "failed to update rakeback override")
		}
	} else {
		// Create new override
		override = &rakeback_override.GlobalRakebackOverride{
			IsActive:          req.IsActive,
			RakebackPercentage: req.RakebackPercentage,
			StartTime:         req.StartTime,
			EndTime:           req.EndTime,
			CreatedBy:         &adminID,
			UpdatedBy:         &adminID,
		}

		override, err = r.storage.CreateOverride(ctx, *override)
		if err != nil {
			r.log.Error("Failed to create rakeback override", zap.Error(err))
			return nil, errors.ErrUnableTocreate.Wrap(err, "failed to create rakeback override")
		}
	}

	r.log.Info("Rakeback override created/updated",
		zap.String("id", override.ID.String()),
		zap.Bool("is_active", override.IsActive),
		zap.String("percentage", override.RakebackPercentage.String()),
		zap.String("admin_id", adminID.String()))

	return r.convertToDTO(override), nil
}

// ToggleOverride enables or disables the global rakeback override
func (r *rakebackOverride) ToggleOverride(ctx context.Context, isActive bool, adminID uuid.UUID) error {
	override, err := r.storage.GetOverride(ctx)
	if err != nil {
		r.log.Error("Failed to get rakeback override", zap.Error(err))
		return errors.ErrUnableToGet.Wrap(err, "failed to get rakeback override")
	}

	if override == nil {
		err := errors.ErrResourceNotFound.New("no rakeback override found to toggle")
		return err
	}

	if !isActive {
		// Disable override
		err = r.storage.DisableOverride(ctx, override.ID, adminID)
		if err != nil {
			r.log.Error("Failed to disable rakeback override", zap.Error(err))
			return errors.ErrUnableToUpdate.Wrap(err, "failed to disable rakeback override")
		}
	} else {
		// Enable override
		override.IsActive = true
		override.UpdatedBy = &adminID
		_, err = r.storage.UpdateOverride(ctx, *override)
		if err != nil {
			r.log.Error("Failed to enable rakeback override", zap.Error(err))
			return errors.ErrUnableToUpdate.Wrap(err, "failed to enable rakeback override")
		}
	}

	r.log.Info("Rakeback override toggled",
		zap.String("id", override.ID.String()),
		zap.Bool("is_active", isActive),
		zap.String("admin_id", adminID.String()))

	return nil
}

// convertToDTO converts storage model to DTO
func (r *rakebackOverride) convertToDTO(override *rakeback_override.GlobalRakebackOverride) *dto.GlobalRakebackOverride {
	return &dto.GlobalRakebackOverride{
		ID:                override.ID,
		IsActive:          override.IsActive,
		RakebackPercentage: override.RakebackPercentage,
		StartTime:         override.StartTime,
		EndTime:           override.EndTime,
		CreatedBy:         override.CreatedBy,
		CreatedAt:         override.CreatedAt,
		UpdatedBy:         override.UpdatedBy,
		UpdatedAt:         override.UpdatedAt,
	}
}

