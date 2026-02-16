package rakeback_override

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// GlobalRakebackOverride represents the global rakeback override configuration
type GlobalRakebackOverride struct {
	ID                uuid.UUID
	IsActive          bool
	RakebackPercentage decimal.Decimal
	StartTime         *time.Time
	EndTime           *time.Time
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedBy         *uuid.UUID
	UpdatedAt         time.Time
}

// GetID returns the override ID
func (g *GlobalRakebackOverride) GetID() uuid.UUID {
	return g.ID
}

// GetIsActive returns whether the override is active
func (g *GlobalRakebackOverride) GetIsActive() bool {
	return g.IsActive
}

// GetRakebackPercentage returns the rakeback percentage
func (g *GlobalRakebackOverride) GetRakebackPercentage() decimal.Decimal {
	return g.RakebackPercentage
}

// RakebackOverrideStorage defines the interface for rakeback override operations
type RakebackOverrideStorage interface {
	GetActiveOverride(ctx context.Context) (*GlobalRakebackOverride, error)
	GetOverride(ctx context.Context) (*GlobalRakebackOverride, error)
	CreateOverride(ctx context.Context, override GlobalRakebackOverride) (*GlobalRakebackOverride, error)
	UpdateOverride(ctx context.Context, override GlobalRakebackOverride) (*GlobalRakebackOverride, error)
	DisableOverride(ctx context.Context, overrideID uuid.UUID, updatedBy uuid.UUID) error
}

type rakebackOverrideStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func NewRakebackOverrideStorage(db *persistencedb.PersistenceDB, log *zap.Logger) RakebackOverrideStorage {
	return &rakebackOverrideStorage{
		db:  db,
		log: log,
	}
}

// GetActiveOverride retrieves the currently active global rakeback override
// It checks if is_active is true AND if current time is within the daily time window (start_time to end_time)
func (r *rakebackOverrideStorage) GetActiveOverride(ctx context.Context) (*GlobalRakebackOverride, error) {
	result, err := r.db.Queries.GetActiveGlobalRakebackOverride(ctx)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil // No active override
		}
		r.log.Error("Failed to get active rakeback override", zap.Error(err))
		return nil, err
	}

	override := r.convertToOverride(result)
	
	// Check if current time is within the daily time window
	if !r.isWithinTimeWindow(override) {
		// Override exists but is not within time window, return nil
		return nil, nil
	}
	
	return override, nil
}

// isWithinTimeWindow checks if the current UTC time is within the daily time window
// It compares only the time portion (HH:MM) for daily repetition
func (r *rakebackOverrideStorage) isWithinTimeWindow(override *GlobalRakebackOverride) bool {
	// If no time window is set, consider it always active (if is_active is true)
	if override.StartTime == nil && override.EndTime == nil {
		return true
	}
	
	// Get current UTC time
	now := time.Now().UTC()
	currentHours := now.Hour()
	currentMinutes := now.Minute()
	currentTimeMinutes := currentHours*60 + currentMinutes
	
	// Parse start time (extract time portion only)
	var startTimeMinutes *int
	if override.StartTime != nil {
		startHours := override.StartTime.Hour()
		startMins := override.StartTime.Minute()
		minutes := startHours*60 + startMins
		startTimeMinutes = &minutes
	}
	
	// Parse end time (extract time portion only)
	var endTimeMinutes *int
	if override.EndTime != nil {
		endHours := override.EndTime.Hour()
		endMins := override.EndTime.Minute()
		minutes := endHours*60 + endMins
		endTimeMinutes = &minutes
	}
	
	// Check if current time is within the daily time window
	if startTimeMinutes != nil && endTimeMinutes != nil {
		// Both start and end times are set
		if *endTimeMinutes <= *startTimeMinutes {
			// Time window spans midnight (e.g., 22:00 to 02:00)
			// Active if current time >= start OR current time <= end
			return currentTimeMinutes >= *startTimeMinutes || currentTimeMinutes <= *endTimeMinutes
		} else {
			// Normal time window (same day, e.g., 10:00 to 14:00)
			// Active if current time >= start AND current time <= end
			return currentTimeMinutes >= *startTimeMinutes && currentTimeMinutes <= *endTimeMinutes
		}
	} else if startTimeMinutes != nil {
		// Only start time is set - active if current time >= start
		return currentTimeMinutes >= *startTimeMinutes
	} else if endTimeMinutes != nil {
		// Only end time is set - active if current time <= end
		return currentTimeMinutes <= *endTimeMinutes
	}
	
	// No time window constraints
	return true
}

// GetOverride retrieves the most recent global rakeback override (active or not)
func (r *rakebackOverrideStorage) GetOverride(ctx context.Context) (*GlobalRakebackOverride, error) {
	result, err := r.db.Queries.GetGlobalRakebackOverride(ctx)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil // No override exists
		}
		r.log.Error("Failed to get rakeback override", zap.Error(err))
		return nil, err
	}

	override := r.convertToOverride(result)
	return override, nil
}

// CreateOverride creates a new global rakeback override
func (r *rakebackOverrideStorage) CreateOverride(ctx context.Context, override GlobalRakebackOverride) (*GlobalRakebackOverride, error) {
	var startTime, endTime interface{}
	if override.StartTime != nil {
		startTime = *override.StartTime
	} else {
		startTime = nil
	}
	if override.EndTime != nil {
		endTime = *override.EndTime
	} else {
		endTime = nil
	}

	result, err := r.db.Queries.CreateGlobalRakebackOverride(ctx, db.CreateGlobalRakebackOverrideParams{
		IsActive:          override.IsActive,
		RakebackPercentage: override.RakebackPercentage,
		StartTime:         startTime,
		EndTime:           endTime,
		CreatedBy:         override.CreatedBy,
		UpdatedBy:         override.UpdatedBy,
	})
	if err != nil {
		r.log.Error("Failed to create rakeback override", zap.Error(err))
		return nil, err
	}

	createdOverride := r.convertToOverride(result)
	return createdOverride, nil
}

// UpdateOverride updates an existing global rakeback override
func (r *rakebackOverrideStorage) UpdateOverride(ctx context.Context, override GlobalRakebackOverride) (*GlobalRakebackOverride, error) {
	var startTime, endTime interface{}
	if override.StartTime != nil {
		startTime = *override.StartTime
	} else {
		startTime = nil
	}
	if override.EndTime != nil {
		endTime = *override.EndTime
	} else {
		endTime = nil
	}

	result, err := r.db.Queries.UpdateGlobalRakebackOverride(ctx, db.UpdateGlobalRakebackOverrideParams{
		ID:                override.ID,
		IsActive:          override.IsActive,
		RakebackPercentage: override.RakebackPercentage,
		StartTime:         startTime,
		EndTime:           endTime,
		UpdatedBy:         override.UpdatedBy,
	})
	if err != nil {
		r.log.Error("Failed to update rakeback override", zap.Error(err))
		return nil, err
	}

	updatedOverride := r.convertToOverride(result)
	return updatedOverride, nil
}

// DisableOverride disables the global rakeback override
func (r *rakebackOverrideStorage) DisableOverride(ctx context.Context, overrideID uuid.UUID, updatedBy uuid.UUID) error {
	err := r.db.Queries.DisableGlobalRakebackOverride(ctx, db.DisableGlobalRakebackOverrideParams{
		ID:        overrideID,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		r.log.Error("Failed to disable rakeback override", zap.Error(err))
		return err
	}
	return nil
}

// convertToOverride converts SQLC result to GlobalRakebackOverride
func (r *rakebackOverrideStorage) convertToOverride(result interface{}) *GlobalRakebackOverride {
	// Handle different result types
	switch v := result.(type) {
	case db.GetActiveGlobalRakebackOverrideRow:
		return &GlobalRakebackOverride{
			ID:                v.ID,
			IsActive:          v.IsActive,
			RakebackPercentage: v.RakebackPercentage,
			StartTime:         r.convertNullTime(v.StartTime),
			EndTime:           r.convertNullTime(v.EndTime),
			CreatedBy:         r.convertNullUUID(v.CreatedBy),
			CreatedAt:         v.CreatedAt,
			UpdatedBy:         r.convertNullUUID(v.UpdatedBy),
			UpdatedAt:         v.UpdatedAt,
		}
	case db.GetGlobalRakebackOverrideRow:
		return &GlobalRakebackOverride{
			ID:                v.ID,
			IsActive:          v.IsActive,
			RakebackPercentage: v.RakebackPercentage,
			StartTime:         r.convertNullTime(v.StartTime),
			EndTime:           r.convertNullTime(v.EndTime),
			CreatedBy:         r.convertNullUUID(v.CreatedBy),
			CreatedAt:         v.CreatedAt,
			UpdatedBy:         r.convertNullUUID(v.UpdatedBy),
			UpdatedAt:         v.UpdatedAt,
		}
	case db.CreateGlobalRakebackOverrideRow:
		return &GlobalRakebackOverride{
			ID:                v.ID,
			IsActive:          v.IsActive,
			RakebackPercentage: v.RakebackPercentage,
			StartTime:         r.convertNullTime(v.StartTime),
			EndTime:           r.convertNullTime(v.EndTime),
			CreatedBy:         r.convertNullUUID(v.CreatedBy),
			CreatedAt:         v.CreatedAt,
			UpdatedBy:         r.convertNullUUID(v.UpdatedBy),
			UpdatedAt:         v.UpdatedAt,
		}
	case db.UpdateGlobalRakebackOverrideRow:
		return &GlobalRakebackOverride{
			ID:                v.ID,
			IsActive:          v.IsActive,
			RakebackPercentage: v.RakebackPercentage,
			StartTime:         r.convertNullTime(v.StartTime),
			EndTime:           r.convertNullTime(v.EndTime),
			CreatedBy:         r.convertNullUUID(v.CreatedBy),
			CreatedAt:         v.CreatedAt,
			UpdatedBy:         r.convertNullUUID(v.UpdatedBy),
			UpdatedAt:         v.UpdatedAt,
		}
	default:
		return nil
	}
}

func (r *rakebackOverrideStorage) convertNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func (r *rakebackOverrideStorage) convertNullUUID(nu pgtype.UUID) *uuid.UUID {
	if nu.Status == pgtype.Present {
		// Convert pgtype.UUID.Bytes (which is [16]byte) to uuid.UUID
		uid, err := uuid.FromBytes(nu.Bytes[:])
		if err != nil {
			return nil
		}
		return &uid
	}
	return nil
}

