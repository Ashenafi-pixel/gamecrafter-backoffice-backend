package cashback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/cashback"
)

// LevelProgressionService handles automatic user level progression
type LevelProgressionService struct {
	logger  *zap.Logger
	storage cashback.CashbackStorage
}

// NewLevelProgressionService creates a new level progression service
func NewLevelProgressionService(logger *zap.Logger, storage cashback.CashbackStorage) *LevelProgressionService {
	return &LevelProgressionService{
		logger:  logger,
		storage: storage,
	}
}

// CheckAndProcessLevelProgression checks if user qualifies for level progression and processes it
func (s *LevelProgressionService) CheckAndProcessLevelProgression(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	s.logger.Info("Checking level progression for user", zap.String("user_id", userID.String()))

	// Get current user level
	userLevel, err := s.storage.GetUserLevel(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user level", zap.Error(err))
		return nil, fmt.Errorf("failed to get user level: %w", err)
	}

	if userLevel == nil {
		s.logger.Warn("User level not found, initializing", zap.String("user_id", userID.String()))
		// Initialize user level if not found
		userLevel, err = s.storage.InitializeUserLevel(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to initialize user level", zap.Error(err))
			return nil, fmt.Errorf("failed to initialize user level: %w", err)
		}
	}

	// Get all tiers ordered by level
	tiers, err := s.storage.GetAllCashbackTiers(ctx)
	if err != nil {
		s.logger.Error("Failed to get cashback tiers", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback tiers: %w", err)
	}

	// Find the highest tier the user qualifies for
	var highestQualifyingTier *dto.CashbackTier
	for _, tier := range tiers {
		if userLevel.TotalExpectedGGR.GreaterThanOrEqual(tier.MinExpectedGGRRequired) {
			if highestQualifyingTier == nil || tier.TierLevel > highestQualifyingTier.TierLevel {
				highestQualifyingTier = &tier
			}
		}
	}

	if highestQualifyingTier == nil {
		s.logger.Warn("No qualifying tier found for user",
			zap.String("user_id", userID.String()),
			zap.String("current_expected_ggr", userLevel.TotalExpectedGGR.String()))
		return userLevel, nil
	}

	// Check if user needs to be upgraded
	if highestQualifyingTier.TierLevel > userLevel.CurrentLevel {
		s.logger.Info("User qualifies for level progression",
			zap.String("user_id", userID.String()),
			zap.Int("current_level", userLevel.CurrentLevel),
			zap.Int("new_level", highestQualifyingTier.TierLevel),
			zap.String("new_tier", highestQualifyingTier.TierName),
			zap.String("total_expected_ggr", userLevel.TotalExpectedGGR.String()),
			zap.String("required_expected_ggr", highestQualifyingTier.MinExpectedGGRRequired.String()))

		// Process the level progression
		updatedUserLevel, err := s.processLevelProgression(ctx, userLevel, highestQualifyingTier)
		if err != nil {
			s.logger.Error("Failed to process level progression", zap.Error(err))
			return nil, fmt.Errorf("failed to process level progression: %w", err)
		}

		// Create level progression notification
		err = s.createLevelProgressionNotification(ctx, userID, userLevel.CurrentLevel, highestQualifyingTier.TierLevel, highestQualifyingTier.TierName)
		if err != nil {
			s.logger.Error("Failed to create level progression notification", zap.Error(err))
			// Don't fail the progression if notification fails
		}

		return updatedUserLevel, nil
	}

	s.logger.Info("User already at appropriate level",
		zap.String("user_id", userID.String()),
		zap.Int("current_level", userLevel.CurrentLevel),
		zap.String("current_tier", highestQualifyingTier.TierName))

	return userLevel, nil
}

// processLevelProgression processes the actual level progression
func (s *LevelProgressionService) processLevelProgression(ctx context.Context, userLevel *dto.UserLevel, newTier *dto.CashbackTier) (*dto.UserLevel, error) {
	s.logger.Info("Processing level progression",
		zap.String("user_id", userLevel.UserID.String()),
		zap.Int("from_level", userLevel.CurrentLevel),
		zap.Int("to_level", newTier.TierLevel),
		zap.String("new_tier", newTier.TierName))

	// Calculate level progress percentage
	var levelProgress decimal.Decimal
	if newTier.TierLevel < 5 { // Not the highest tier
		// Find next tier for progress calculation
		nextTier, err := s.storage.GetCashbackTierByLevel(ctx, newTier.TierLevel+1)
		if err == nil && nextTier != nil {
			// Calculate progress: (current_expected_ggr - current_tier_min) / (next_tier_min - current_tier_min)
			currentTierMin := newTier.MinExpectedGGRRequired
			nextTierMin := nextTier.MinExpectedGGRRequired
			progressNumerator := userLevel.TotalExpectedGGR.Sub(currentTierMin)
			progressDenominator := nextTierMin.Sub(currentTierMin)

			if progressDenominator.GreaterThan(decimal.Zero) {
				levelProgress = progressNumerator.Div(progressDenominator)
				if levelProgress.GreaterThan(decimal.NewFromInt(1)) {
					levelProgress = decimal.NewFromInt(1) // Cap at 100%
				}
			}
		}
	} else {
		levelProgress = decimal.NewFromInt(1) // 100% for highest tier
	}

	// Update user level
	now := time.Now()
	updatedUserLevel := dto.UserLevel{
		ID:               userLevel.ID,
		UserID:           userLevel.UserID,
		CurrentLevel:     newTier.TierLevel,
		TotalExpectedGGR: userLevel.TotalExpectedGGR,
		TotalBets:        userLevel.TotalBets,
		TotalWins:        userLevel.TotalWins,
		LevelProgress:    levelProgress,
		CurrentTierID:    newTier.ID,
		LastLevelUp:      &now,
		CreatedAt:        userLevel.CreatedAt,
		UpdatedAt:        now,
	}

	// Save updated user level
	savedUserLevel, err := s.storage.UpdateUserLevel(ctx, updatedUserLevel)
	if err != nil {
		s.logger.Error("Failed to update user level", zap.Error(err))
		return nil, fmt.Errorf("failed to update user level: %w", err)
	}

	s.logger.Info("Level progression completed successfully",
		zap.String("user_id", userLevel.UserID.String()),
		zap.Int("new_level", newTier.TierLevel),
		zap.String("new_tier", newTier.TierName),
		zap.String("level_progress", levelProgress.String()),
		zap.String("cashback_percentage", newTier.CashbackPercentage.String()))

	return &savedUserLevel, nil
}

// createLevelProgressionNotification creates a notification for level progression
func (s *LevelProgressionService) createLevelProgressionNotification(ctx context.Context, userID uuid.UUID, fromLevel, toLevel int, tierName string) error {
	s.logger.Info("Creating level progression notification",
		zap.String("user_id", userID.String()),
		zap.Int("from_level", fromLevel),
		zap.Int("to_level", toLevel),
		zap.String("tier_name", tierName))

	// Create notification data
	notificationData := map[string]interface{}{
		"type":           "level_progression",
		"from_level":     fromLevel,
		"to_level":       toLevel,
		"tier_name":      tierName,
		"title":          "Level Up! 🎉",
		"message":        fmt.Sprintf("Congratulations! You've reached %s tier (Level %d)!", tierName, toLevel),
		"progression_at": time.Now(),
	}

	// Store notification (this would typically go to a notification service)
	s.logger.Info("Level progression notification created",
		zap.String("user_id", userID.String()),
		zap.String("notification_type", "level_progression"),
		zap.String("title", notificationData["title"].(string)),
		zap.String("message", notificationData["message"].(string)))

	return nil
}

// GetLevelProgressionInfo returns information about user's level progression
func (s *LevelProgressionService) GetLevelProgressionInfo(ctx context.Context, userID uuid.UUID) (*dto.LevelProgressionInfo, error) {
	s.logger.Info("Getting level progression info", zap.String("user_id", userID.String()))

	// Get current user level
	userLevel, err := s.storage.GetUserLevel(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user level", zap.Error(err))
		return nil, fmt.Errorf("failed to get user level: %w", err)
	}

	if userLevel == nil {
		return nil, fmt.Errorf("user level not found")
	}

	// Get current tier
	currentTier, err := s.storage.GetCashbackTierByID(ctx, userLevel.CurrentTierID)
	if err != nil {
		s.logger.Error("Failed to get current tier", zap.Error(err))
		return nil, fmt.Errorf("failed to get current tier: %w", err)
	}

	// Get next tier
	var nextTier *dto.CashbackTier
	if userLevel.CurrentLevel < 5 { // Not the highest tier
		nextTier, err = s.storage.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel+1)
		if err != nil {
			s.logger.Warn("Failed to get next tier", zap.Error(err))
			// Continue without next tier info
		}
	}

	// Calculate progress to next level
	var progressToNext decimal.Decimal
	var ggrToNext decimal.Decimal
	if nextTier != nil {
		ggrToNext = nextTier.MinExpectedGGRRequired.Sub(userLevel.TotalExpectedGGR)
		if ggrToNext.LessThanOrEqual(decimal.Zero) {
			progressToNext = decimal.NewFromInt(1) // Already qualified
		} else {
			currentTierMin := currentTier.MinExpectedGGRRequired
			nextTierMin := nextTier.MinExpectedGGRRequired
			progressNumerator := userLevel.TotalExpectedGGR.Sub(currentTierMin)
			progressDenominator := nextTierMin.Sub(currentTierMin)

			if progressDenominator.GreaterThan(decimal.Zero) {
				progressToNext = progressNumerator.Div(progressDenominator)
				if progressToNext.GreaterThan(decimal.NewFromInt(1)) {
					progressToNext = decimal.NewFromInt(1)
				}
			}
		}
	}

	progressionInfo := &dto.LevelProgressionInfo{
		UserID:                 userID,
		CurrentLevel:           userLevel.CurrentLevel,
		CurrentTier:            *currentTier,
		NextTier:               nextTier,
		TotalExpectedGGR:       userLevel.TotalExpectedGGR,
		ProgressToNext:         progressToNext,
		ExpectedGGRToNextLevel: ggrToNext,
		LastLevelUp:            userLevel.LastLevelUp,
		LevelProgress:          userLevel.LevelProgress,
	}

	s.logger.Info("Level progression info retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("current_level", userLevel.CurrentLevel),
		zap.String("current_tier", currentTier.TierName),
		zap.String("progress_to_next", progressToNext.String()))

	return progressionInfo, nil
}

// ProcessBulkLevelProgression processes level progression for multiple users
func (s *LevelProgressionService) ProcessBulkLevelProgression(ctx context.Context, userIDs []uuid.UUID) ([]dto.LevelProgressionResult, error) {
	s.logger.Info("Processing bulk level progression", zap.Int("user_count", len(userIDs)))

	var results []dto.LevelProgressionResult
	var errors []error

	for _, userID := range userIDs {
		result := dto.LevelProgressionResult{
			UserID: userID,
		}

		userLevel, err := s.CheckAndProcessLevelProgression(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to process level progression for user",
				zap.String("user_id", userID.String()),
				zap.Error(err))
			result.Error = err.Error()
			errors = append(errors, err)
		} else {
			result.Success = true
			result.NewLevel = userLevel.CurrentLevel
			result.UpdatedAt = userLevel.UpdatedAt
		}

		results = append(results, result)
	}

	s.logger.Info("Bulk level progression completed",
		zap.Int("total_users", len(userIDs)),
		zap.Int("successful", len(results)-len(errors)),
		zap.Int("failed", len(errors)))

	return results, nil
}
