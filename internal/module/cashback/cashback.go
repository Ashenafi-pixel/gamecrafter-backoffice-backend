package cashback

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/cashback"
	"github.com/tucanbit/internal/storage/groove"
	rakeback_override "github.com/tucanbit/internal/storage/rakeback_override"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

// DashboardStats represents comprehensive dashboard statistics
type DashboardStats struct {
	TotalUsers             int64                   `json:"total_users"`
	ActiveUsers            int64                   `json:"active_users"`
	TotalCashbackEarned    decimal.Decimal         `json:"total_cashback_earned"`
	TotalCashbackClaimed   decimal.Decimal         `json:"total_cashback_claimed"`
	PendingCashback        decimal.Decimal         `json:"pending_cashback"`
	AverageCashbackPerUser decimal.Decimal         `json:"average_cashback_per_user"`
	TopCashbackUsers       []UserCashbackSummary   `json:"top_cashback_users"`
	CashbackTiers          []dto.CashbackTier      `json:"cashback_tiers"`
	RecentClaims           []dto.CashbackClaim     `json:"recent_claims"`
	DailyCashbackTrend     []DailyCashbackData     `json:"daily_cashback_trend"`
	GameTypeStats          []GameTypeCashbackStats `json:"game_type_stats"`
}

type UserCashbackSummary struct {
	UserID          uuid.UUID       `json:"user_id"`
	Username        string          `json:"username"`
	Email           string          `json:"email"`
	CurrentLevel    int             `json:"current_level"`
	TotalCashback   decimal.Decimal `json:"total_cashback"`
	PendingCashback decimal.Decimal `json:"pending_cashback"`
	ClaimedCashback decimal.Decimal `json:"claimed_cashback"`
	LastActivity    *time.Time      `json:"last_activity"`
}

type DailyCashbackData struct {
	Date         string          `json:"date"`
	TotalEarned  decimal.Decimal `json:"total_earned"`
	TotalClaimed decimal.Decimal `json:"total_claimed"`
	ActiveUsers  int64           `json:"active_users"`
	NewUsers     int64           `json:"new_users"`
}

type GameTypeCashbackStats struct {
	GameType        string          `json:"game_type"`
	TotalBets       int64           `json:"total_bets"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	TotalCashback   decimal.Decimal `json:"total_cashback"`
	AverageCashback decimal.Decimal `json:"average_cashback"`
	HouseEdge       decimal.Decimal `json:"house_edge"`
}

type CashbackService struct {
	storage                 cashback.CashbackStorage
	grooveStorage           groove.GrooveStorage
	retryService            *RetryService
	levelProgressionService *LevelProgressionService
	userWS                  utils.UserWS
	rakebackOverrideStorage rakeback_override.RakebackOverrideStorage
	logger                  *zap.Logger
}

func NewCashbackService(storage cashback.CashbackStorage, grooveStorage groove.GrooveStorage, userWS utils.UserWS, logger *zap.Logger, rakebackOverrideStorage rakeback_override.RakebackOverrideStorage) *CashbackService {
	// Create the service first
	service := &CashbackService{
		storage:                 storage,
		grooveStorage:           grooveStorage,
		userWS:                  userWS,
		logger:                  logger,
		rakebackOverrideStorage: rakebackOverrideStorage,
	}

	// Create retry service with all dependencies
	retryService := NewRetryService(logger, storage, grooveStorage, service, DefaultRetryConfig())
	service.retryService = retryService

	// Create level progression service
	levelProgressionService := NewLevelProgressionService(logger, storage)
	service.levelProgressionService = levelProgressionService

	return service
}

// InitializeUserLevel creates a new user level entry for a user
func (s *CashbackService) InitializeUserLevel(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Initializing user level", zap.String("user_id", userID.String()))

	// Get the default tier (Bronze)
	defaultTier, err := s.storage.GetCashbackTierByLevel(ctx, 1)
	if err != nil {
		s.logger.Error("Failed to get default tier", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get default tier")
	}

	// Create user level
	userLevel := dto.UserLevel{
		UserID:           userID,
		CurrentLevel:     1,
		TotalExpectedGGR: decimal.Zero,
		TotalBets:        decimal.Zero,
		TotalWins:        decimal.Zero,
		LevelProgress:    decimal.Zero,
		CurrentTierID:    defaultTier.ID,
	}

	_, err = s.storage.CreateUserLevel(ctx, userLevel)
	if err != nil {
		s.logger.Error("Failed to create user level", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to create user level")
	}

	s.logger.Info("User level initialized successfully", zap.String("user_id", userID.String()))
	return nil
}

// ProcessBetCashback calculates and creates cashback earnings for a bet
func (s *CashbackService) ProcessBetCashback(ctx context.Context, bet dto.Bet) error {
	s.logger.Info("Processing bet cashback",
		zap.String("bet_id", bet.BetID.String()),
		zap.String("user_id", bet.UserID.String()))

	// Get user level, create if doesn't exist using dynamic initialization
	userLevel, err := s.storage.GetUserLevel(ctx, bet.UserID)
	if err != nil {
		s.logger.Info("User level not found, initializing new level dynamically",
			zap.String("user_id", bet.UserID.String()))

		// Use the proper InitializeUserLevel method for dynamic tier initialization
		err = s.InitializeUserLevel(ctx, bet.UserID)
		if err != nil {
			s.logger.Error("Failed to initialize user level", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to initialize user level")
		}

		// Get the newly created user level
		userLevel, err = s.storage.GetUserLevel(ctx, bet.UserID)
		if err != nil {
			s.logger.Error("Failed to retrieve newly created user level", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to retrieve newly created user level")
		}

		s.logger.Info("User level initialized successfully with dynamic tier",
			zap.String("user_id", bet.UserID.String()),
			zap.Int("level", userLevel.CurrentLevel),
			zap.String("tier_id", userLevel.CurrentTierID.String()))
	}

	// Get effective tier for cashback calculations
	effectiveTier, err := s.storage.GetCashbackTierByID(ctx, userLevel.EffectiveTierID)
	if err != nil {
		s.logger.Warn("Failed to get effective tier by ID, falling back to level",
			zap.Error(err),
			zap.String("user_id", bet.UserID.String()),
			zap.String("tier_id", userLevel.EffectiveTierID.String()))

		effectiveTier, err = s.storage.GetCashbackTierByLevel(ctx, userLevel.EffectiveLevel)
		if err != nil {
			s.logger.Error("Failed to get effective tier", zap.Error(err))
			return errors.ErrInternalServerError.Wrap(err, "failed to get effective tier")
		}
	}

	currentTier, err := s.storage.GetCashbackTierByID(ctx, userLevel.CurrentTierID)
	if err != nil {
		s.logger.Warn("Failed to get current tier", zap.Error(err))
		currentTier = effectiveTier
	}
	if currentTier != nil && effectiveTier != nil && currentTier.ID != effectiveTier.ID {
		s.logger.Info("Manual override tier applied for cashback",
			zap.String("user_id", bet.UserID.String()),
			zap.String("actual_tier", currentTier.TierName),
			zap.String("effective_tier", effectiveTier.TierName),
			zap.Int("actual_level", userLevel.CurrentLevel),
			zap.Int("effective_level", userLevel.EffectiveLevel))
	}

	// Get configurable house edge for the specific game
	var houseEdge decimal.Decimal
	gameType := "groovetech" // Default operator
	gameVariant := ""        // Will be extracted from bet data if available

	// Try to extract game information from bet data
	if bet.ClientTransactionID != "" {
		// For GrooveTech, we can extract game info from transaction context
		// This would typically come from the game session or transaction metadata
		gameVariant = s.extractGameVariantFromTransaction(ctx, bet.ClientTransactionID)
	}

	// Check for active global rakeback override (Happy Hour)
	if s.rakebackOverrideStorage != nil {
		override, err := s.rakebackOverrideStorage.GetActiveOverride(ctx)
		if err == nil && override != nil && override.GetIsActive() {
			// Convert percentage to decimal (e.g., 100% = 1.0)
			houseEdge = override.GetRakebackPercentage().Div(decimal.NewFromInt(100))
			s.logger.Info("Using global rakeback override",
				zap.String("override_id", override.GetID().String()),
				zap.String("rakeback_percentage", override.GetRakebackPercentage().String()),
				zap.String("house_edge", houseEdge.String()))
		} else {
			// Use normal game house edge
			gameHouseEdge, err := s.storage.GetGameHouseEdge(ctx, gameType, gameVariant)
			if err != nil {
				s.logger.Error("Failed to get game house edge, using default", zap.Error(err))
				houseEdge = decimal.NewFromFloat(0.0) // Default 0% house edge (no default cashback)
			} else {
				houseEdge = gameHouseEdge.HouseEdge
				s.logger.Info("Using configurable house edge",
					zap.String("game_type", gameType),
					zap.String("game_variant", gameVariant),
					zap.String("house_edge", houseEdge.String()))
			}
		}
	} else {
		// Use normal game house edge
		gameHouseEdge, err := s.storage.GetGameHouseEdge(ctx, gameType, gameVariant)
		if err != nil {
			s.logger.Error("Failed to get game house edge, using default", zap.Error(err))
			houseEdge = decimal.NewFromFloat(0.0) // Default 0% house edge (no default cashback)
		} else {
			houseEdge = gameHouseEdge.HouseEdge
			s.logger.Info("Using configurable house edge",
				zap.String("game_type", gameType),
				zap.String("game_variant", gameVariant),
				zap.String("house_edge", houseEdge.String()))
		}
	}

	// Calculate expected GGR (Expected Gross Gaming Revenue) - kept for logging purposes
	expectedGGR := bet.Amount.Mul(houseEdge)

	// Priority order for rakeback: Global Override > Scheduled Rakeback > VIP Tier
	var cashbackRate decimal.Decimal
	var isGlobalOverrideActive bool
	var isScheduledRakebackActive bool
	var scheduleName string

	// Check for global rakeback override first (Happy Hour Mode)
	globalOverride, err := s.storage.GetGlobalRakebackOverride(ctx)
	if err == nil && globalOverride != nil && globalOverride.IsEnabled {
		// Global override is active - use override percentage for all users
		cashbackRate = globalOverride.OverridePercentage
		isGlobalOverrideActive = true
		s.logger.Info("Applying global rakeback override (Happy Hour)",
			zap.String("override_percentage", cashbackRate.String()),
			zap.String("user_id", bet.UserID.String()),
			zap.String("original_house_edge", houseEdge.String()))
	} else {
		// Check for scheduled rakeback (if no global override)
		// Use gameVariant (which contains the game ID) for scheduled rakeback lookup
		activeSchedule, err := s.storage.GetActiveScheduleForBet(ctx, gameType, gameVariant)
		if err == nil && activeSchedule != nil {
			// Scheduled rakeback is active and applies to this bet
			cashbackRate = activeSchedule.Percentage
			isScheduledRakebackActive = true
			scheduleName = activeSchedule.Name
			s.logger.Info("Applying scheduled rakeback",
				zap.String("schedule_name", activeSchedule.Name),
				zap.String("scope_type", activeSchedule.ScopeType),
				zap.String("percentage", cashbackRate.String()),
				zap.String("user_id", bet.UserID.String()),
				zap.String("game_type", gameType))
		} else {
			// Use game house edge as cashback rate (per-wager cashback)
			cashbackRate = houseEdge.Mul(decimal.NewFromInt(100)) // Convert house edge to percentage
			isGlobalOverrideActive = false
			isScheduledRakebackActive = false
		}
	}

	// Check for active promotions (skip for now)
	// promotion, err := s.storage.GetCashbackPromotionForUser(ctx, userLevel.CurrentLevel, "default")
	// if err == nil && promotion != nil {
	// 	// Apply promotion boost
	// 	cashbackRate = cashbackRate.Mul(promotion.BoostMultiplier)
	// 	s.logger.Info("Applied promotion boost",
	// 		zap.String("promotion", promotion.PromotionName),
	// 		zap.String("boost", promotion.BoostMultiplier.String()))
	// }

	// Calculate earned cashback per wager
	var earnedCashback decimal.Decimal
	if isGlobalOverrideActive || isScheduledRakebackActive {
		// When override or scheduled rakeback is active, calculate based on percentage
		// Percentage is already in % format (e.g., 100.00 = 100%)
		earnedCashback = bet.Amount.Mul(cashbackRate).Div(decimal.NewFromInt(100))

		if isScheduledRakebackActive {
			s.logger.Info("Calculated scheduled rakeback cashback",
				zap.String("schedule_name", scheduleName),
				zap.String("bet_amount", bet.Amount.String()),
				zap.String("cashback_rate", cashbackRate.String()),
				zap.String("earned_cashback", earnedCashback.String()))
		}
	} else {
		// Normal calculation: bet amount * house edge
		earnedCashback = bet.Amount.Mul(houseEdge)
	}

	// Round to 2 decimal places
	earnedCashback = earnedCashback.Round(2)

	// Create cashback earning
	cashbackEarning := dto.CashbackEarning{
		UserID:            bet.UserID,
		TierID:            userLevel.EffectiveTierID,
		EarningType:       "bet",
		SourceBetID:       nil,         // Set to nil for GrooveTech transactions to avoid foreign key constraint
		ExpectedGGRAmount: expectedGGR, // Store actual expected GGR based on house edge
		CashbackRate:      cashbackRate,
		EarnedAmount:      earnedCashback,
		ClaimedAmount:     decimal.Zero,
		AvailableAmount:   earnedCashback,
		Status:            "available",
		ExpiresAt:         time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	// Use retry mechanism for critical operations
	err = s.retryService.RetryOperation(ctx, "process_bet_cashback", bet.UserID, map[string]interface{}{
		"bet_id":                bet.BetID.String(),
		"round_id":              bet.RoundID.String(),
		"client_transaction_id": bet.ClientTransactionID,
		"amount":                bet.Amount.String(),
		"currency":              bet.Currency,
		"payout":                bet.Payout.String(),
		"cash_out_multiplier":   bet.CashOutMultiplier.String(),
		"status":                bet.Status,
		"cashback_earning":      cashbackEarning,
		"updated_user_level": dto.UserLevel{
			UserID:           bet.UserID,
			CurrentLevel:     userLevel.CurrentLevel,
			TotalExpectedGGR: userLevel.TotalExpectedGGR.Add(expectedGGR), // Use actual expected GGR based on house edge
			TotalBets:        userLevel.TotalBets.Add(bet.Amount),
			TotalWins:        userLevel.TotalWins.Add(bet.Payout),
			LevelProgress:    userLevel.LevelProgress,
			CurrentTierID:    userLevel.CurrentTierID,
		},
	}, func() error {
		// Create cashback earning
		_, err := s.storage.CreateCashbackEarning(ctx, cashbackEarning)
		if err != nil {
			return fmt.Errorf("failed to create cashback earning: %w", err)
		}

		// Update user level statistics
		updatedUserLevel := dto.UserLevel{
			UserID:           bet.UserID,
			CurrentLevel:     userLevel.CurrentLevel,
			TotalExpectedGGR: userLevel.TotalExpectedGGR.Add(expectedGGR), // Use actual expected GGR based on house edge
			TotalBets:        userLevel.TotalBets.Add(bet.Amount),
			TotalWins:        userLevel.TotalWins.Add(bet.Payout), // We use payout as win amount
			LevelProgress:    userLevel.LevelProgress,
			CurrentTierID:    userLevel.CurrentTierID,
		}

		_, err = s.storage.UpdateUserLevel(ctx, updatedUserLevel)
		if err != nil {
			return fmt.Errorf("failed to update user level: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to process bet cashback with retry", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to process bet cashback")
	}

	s.logger.Info("Bet cashback processed successfully (per-wager calculation using game house edge)",
		zap.String("bet_id", bet.BetID.String()),
		zap.String("bet_amount", bet.Amount.String()),
		zap.String("house_edge", houseEdge.String()),
		zap.String("expected_ggr", expectedGGR.String()),
		zap.String("cashback_rate", cashbackRate.String()),
		zap.String("earned_cashback", earnedCashback.String()),
		zap.String("user_tier", effectiveTier.TierName))

	// Trigger WebSocket notification for new cashback availability with game details
	if s.userWS != nil {
		cashbackSummary, err := s.GetUserCashbackSummary(ctx, bet.UserID)
		if err == nil && cashbackSummary != nil {
			// Create enhanced cashback data with game-specific information
			enhancedCashbackData := s.createEnhancedCashbackData(ctx, *cashbackSummary, bet, houseEdge, cashbackRate)
			s.userWS.TriggerCashbackWS(ctx, bet.UserID, enhancedCashbackData)
			s.logger.Debug("Cashback WebSocket notification triggered with game details",
				zap.String("user_id", bet.UserID.String()),
				zap.String("available_cashback", cashbackSummary.AvailableCashback.String()),
				zap.String("game_type", gameType),
				zap.String("house_edge", houseEdge.String()),
				zap.String("cashback_rate", cashbackRate.String()))
		}
	}

	// Check and process automatic level progression after cashback processing
	s.logger.Info("Checking automatic level progression after cashback processing",
		zap.String("user_id", bet.UserID.String()))

	updatedUserLevel, err := s.CheckAndProcessLevelProgression(ctx, bet.UserID)
	if err != nil {
		s.logger.Error("Failed to check level progression", zap.Error(err))
		// Don't fail the cashback processing if level progression fails
	} else if updatedUserLevel != nil {
		s.logger.Info("Level progression completed",
			zap.String("user_id", bet.UserID.String()),
			zap.Int("new_level", updatedUserLevel.CurrentLevel),
			zap.String("level_progress", updatedUserLevel.LevelProgress.String()))
	}

	return nil
}

// GetUserCashbackSummary returns a comprehensive summary of user's cashback status
func (s *CashbackService) GetUserCashbackSummary(ctx context.Context, userID uuid.UUID) (*dto.UserCashbackSummary, error) {
	s.logger.Info("Getting user cashback summary", zap.String("user_id", userID.String()))

	summary, err := s.storage.GetUserCashbackSummary(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user cashback summary", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user cashback summary")
	}

	// Check for global rakeback override (Happy Hour Mode)
	globalOverride, err := s.storage.GetGlobalRakebackOverride(ctx)
	if err == nil && globalOverride != nil && globalOverride.IsEnabled {
		// Global override is active - update summary with override information
		summary.GlobalOverrideActive = true
		summary.EffectiveCashbackPercent = globalOverride.OverridePercentage
		summary.HappyHourMessage = fmt.Sprintf("🎉 Happy Hour Active! All users receive %s%% rakeback!", globalOverride.OverridePercentage.String())

		s.logger.Info("Global rakeback override active for user summary",
			zap.String("user_id", userID.String()),
			zap.String("override_percentage", globalOverride.OverridePercentage.String()))
	} else {
		// Normal VIP-based rakeback
		summary.GlobalOverrideActive = false
		summary.EffectiveCashbackPercent = summary.CurrentTier.CashbackPercentage
		summary.HappyHourMessage = ""
	}

	return summary, nil
}

// GetCashbackTiers returns all available cashback tiers
func (s *CashbackService) GetCashbackTiers(ctx context.Context) ([]dto.CashbackTier, error) {
	s.logger.Info("Getting cashback tiers")

	tiers, err := s.storage.GetCashbackTiers(ctx)
	if err != nil {
		s.logger.Error("Failed to get cashback tiers", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get cashback tiers")
	}

	s.logger.Info("Retrieved cashback tiers", zap.Int("count", len(tiers)))
	return tiers, nil
}

// ClaimCashback processes a cashback claim request
func (s *CashbackService) ClaimCashback(ctx context.Context, userID uuid.UUID, request dto.CashbackClaimRequest) (*dto.CashbackClaimResponse, error) {
	s.logger.Info("Processing cashback claim",
		zap.String("user_id", userID.String()),
		zap.String("amount", request.Amount.String()))

	// Get user level to check limits
	userLevel, err := s.storage.GetUserLevel(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user level", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user level")
	}

	// Get effective tier to check limits
	currentTier, err := s.storage.GetCashbackTierByID(ctx, userLevel.EffectiveTierID)
	if err != nil {
		s.logger.Warn("Failed to get effective tier by ID for claim limits",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("tier_id", userLevel.EffectiveTierID.String()))

		currentTier, err = s.storage.GetCashbackTierByLevel(ctx, userLevel.EffectiveLevel)
		if err != nil {
			s.logger.Error("Failed to get effective tier for claim limits", zap.Error(err))
			return nil, errors.ErrInternalServerError.Wrap(err, "failed to get effective tier")
		}
	}

	// Check daily limit
	if currentTier.DailyCashbackLimit != nil {
		dailyClaimed, err := s.storage.GetUserDailyCashbackLimit(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get daily cashback limit", zap.Error(err))
			return nil, errors.ErrInternalServerError.Wrap(err, "failed to get daily cashback limit")
		}

		if dailyClaimed.Add(request.Amount).GreaterThan(*currentTier.DailyCashbackLimit) {
			return nil, errors.ErrInvalidUserInput.New("daily cashback limit exceeded")
		}
	}

	// Get available cashback earnings
	availableEarnings, err := s.storage.GetAvailableCashbackEarnings(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get available cashback earnings", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get available cashback earnings")
	}

	// Calculate total available
	totalAvailable := decimal.Zero
	for _, earning := range availableEarnings {
		totalAvailable = totalAvailable.Add(earning.AvailableAmount)
	}

	if request.Amount.GreaterThan(totalAvailable) {
		return nil, errors.ErrInvalidUserInput.New("insufficient available cashback")
	}

	// Process the claim
	claimID := uuid.New()
	claimedEarnings := make(map[string]interface{})
	remainingAmount := request.Amount
	processingFee := decimal.Zero // No processing fee for now

	for _, earning := range availableEarnings {
		if remainingAmount.LessThanOrEqual(decimal.Zero) {
			break
		}

		claimAmount := decimal.Min(remainingAmount, earning.AvailableAmount)

		// Update earning status
		claimedAt := time.Now()
		updatedEarning := dto.CashbackEarning{
			ID:                earning.ID,
			UserID:            earning.UserID,
			TierID:            earning.TierID,
			EarningType:       earning.EarningType,
			SourceBetID:       earning.SourceBetID,
			ExpectedGGRAmount: earning.ExpectedGGRAmount,
			CashbackRate:      earning.CashbackRate,
			EarnedAmount:      earning.EarnedAmount,
			ClaimedAmount:     earning.ClaimedAmount.Add(claimAmount),
			AvailableAmount:   earning.AvailableAmount.Sub(claimAmount),
			Status:            "claimed",
			ExpiresAt:         earning.ExpiresAt,
			ClaimedAt:         &claimedAt,
		}

		_, err = s.storage.UpdateCashbackEarningStatus(ctx, updatedEarning)
		if err != nil {
			s.logger.Error("Failed to update cashback earning status", zap.Error(err))
			return nil, errors.ErrInternalServerError.Wrap(err, "failed to update cashback earning status")
		}

		claimedEarnings[earning.ID.String()] = claimAmount.String()
		remainingAmount = remainingAmount.Sub(claimAmount)
	}

	// Credit the user's balance immediately
	// Note: AddBalance automatically triggers balance WebSocket notification
	netAmount := request.Amount.Sub(processingFee)
	_, err = s.grooveStorage.AddBalance(ctx, userID, netAmount)
	if err != nil {
		s.logger.Error("Failed to credit user balance", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to credit user balance")
	}

	s.logger.Debug("Balance updated and WebSocket notification triggered after cashback claim",
		zap.String("user_id", userID.String()),
		zap.String("claimed_amount", request.Amount.String()),
		zap.String("net_amount", netAmount.String()))

	// Create cashback claim record
	cashbackClaim := dto.CashbackClaim{
		ID:              claimID,
		UserID:          userID,
		ClaimAmount:     request.Amount,
		CurrencyCode:    "USD",
		Status:          "completed",
		ProcessingFee:   processingFee,
		NetAmount:       netAmount,
		ClaimedEarnings: claimedEarnings,
		ProcessedAt:     &time.Time{},
	}

	_, err = s.storage.CreateCashbackClaim(ctx, cashbackClaim)
	if err != nil {
		s.logger.Error("Failed to create cashback claim", zap.Error(err))
		// If claim record creation fails, we should rollback the balance credit
		// For now, just log the error
		s.logger.Error("Balance credited but claim record creation failed - manual reconciliation required")
	}

	response := &dto.CashbackClaimResponse{
		ClaimID:       claimID,
		Amount:        request.Amount,
		NetAmount:     netAmount,
		ProcessingFee: processingFee,
		Status:        "completed",
		Message:       "Cashback claim processed successfully",
	}

	s.logger.Info("Cashback claim processed successfully",
		zap.String("claim_id", claimID.String()),
		zap.String("amount", request.Amount.String()))

	// Trigger WebSocket notification for cashback claim completion with enhanced data
	if s.userWS != nil {
		cashbackSummary, err := s.GetUserCashbackSummary(ctx, userID)
		if err == nil && cashbackSummary != nil {
			// Create enhanced cashback data for claim notification
			enhancedCashbackData := dto.EnhancedUserCashbackSummary{
				UserID:            cashbackSummary.UserID,
				CurrentTier:       cashbackSummary.CurrentTier,
				LevelProgress:     cashbackSummary.LevelProgress,
				TotalGGR:          cashbackSummary.TotalGGR,
				AvailableCashback: cashbackSummary.AvailableCashback,
				PendingCashback:   cashbackSummary.PendingCashback,
				TotalClaimed:      cashbackSummary.TotalClaimed,
				NextTierGGR:       cashbackSummary.NextTierGGR,
				DailyLimit:        cashbackSummary.DailyLimit,
				WeeklyLimit:       cashbackSummary.WeeklyLimit,
				MonthlyLimit:      cashbackSummary.MonthlyLimit,
				SpecialBenefits:   cashbackSummary.SpecialBenefits,
				// No game info for claims (not game-specific)
				LastGameInfo: nil,
			}
			s.userWS.TriggerCashbackWS(ctx, userID, enhancedCashbackData)
			s.logger.Debug("Cashback claim WebSocket notification triggered",
				zap.String("user_id", userID.String()),
				zap.String("claimed_amount", request.Amount.String()),
				zap.String("remaining_cashback", cashbackSummary.AvailableCashback.String()))
		}
	}

	return response, nil
}

// ProcessExpiredCashback marks expired cashback earnings as expired
func (s *CashbackService) ProcessExpiredCashback(ctx context.Context) error {
	s.logger.Info("Processing expired cashback earnings")

	err := s.storage.MarkCashbackEarningsAsExpired(ctx)
	if err != nil {
		s.logger.Error("Failed to mark expired cashback earnings", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to mark expired cashback earnings")
	}

	s.logger.Info("Expired cashback earnings processed successfully")
	return nil
}

// GetCashbackStats returns admin statistics for the cashback system
func (s *CashbackService) GetCashbackStats(ctx context.Context) (*dto.AdminCashbackStats, error) {
	s.logger.Info("Getting cashback statistics")

	stats, err := s.storage.GetCashbackStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get cashback stats", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get cashback stats")
	}

	tierDistribution, err := s.storage.GetTierDistribution(ctx)
	if err != nil {
		s.logger.Error("Failed to get tier distribution", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get tier distribution")
	}

	// Convert tier distribution to map
	tierMap := make(map[string]int)
	for _, tier := range tierDistribution {
		tierMap[tier.TierName] = tier.UserCount
	}

	stats.TierDistribution = tierMap

	return stats, nil
}

// GetComprehensiveStats returns comprehensive dashboard statistics
func (s *CashbackService) GetComprehensiveStats(ctx context.Context, startDate, endDate time.Time) (*DashboardStats, error) {
	s.logger.Info("Getting comprehensive dashboard statistics",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate))

	// This is a placeholder implementation
	// In a real implementation, you would query the database for comprehensive statistics
	stats := &DashboardStats{
		TotalUsers:             1000,
		ActiveUsers:            750,
		TotalCashbackEarned:    decimal.NewFromFloat(50000.00),
		TotalCashbackClaimed:   decimal.NewFromFloat(45000.00),
		PendingCashback:        decimal.NewFromFloat(5000.00),
		AverageCashbackPerUser: decimal.NewFromFloat(50.00),
		TopCashbackUsers:       []UserCashbackSummary{},
		CashbackTiers:          []dto.CashbackTier{},
		RecentClaims:           []dto.CashbackClaim{},
		DailyCashbackTrend:     []DailyCashbackData{},
		GameTypeStats:          []GameTypeCashbackStats{},
	}

	return stats, nil
}

// GetCashbackAnalytics returns detailed analytics for the cashback system
func (s *CashbackService) GetCashbackAnalytics(ctx context.Context, startDate, endDate time.Time, gameType string) (map[string]interface{}, error) {
	s.logger.Info("Getting cashback analytics",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.String("game_type", gameType))

	// This is a placeholder implementation
	analytics := map[string]interface{}{
		"total_cashback_earned":  decimal.NewFromFloat(50000.00),
		"total_cashback_claimed": decimal.NewFromFloat(45000.00),
		"average_daily_cashback": decimal.NewFromFloat(1666.67),
		"top_performing_games": []map[string]interface{}{
			{"game_type": "groovetech", "total_cashback": decimal.NewFromFloat(25000.00)},
			{"game_type": "crash", "total_cashback": decimal.NewFromFloat(15000.00)},
			{"game_type": "plinko", "total_cashback": decimal.NewFromFloat(10000.00)},
		},
		"user_retention_rate":      0.85,
		"cashback_conversion_rate": 0.90,
	}

	return analytics, nil
}

// CreateManualCashbackEarning creates a manual cashback earning for a user
func (s *CashbackService) CreateManualCashbackEarning(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, reason, gameType, gameID string) (*dto.CashbackEarning, error) {
	s.logger.Info("Creating manual cashback earning",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
		zap.String("reason", reason))

	// Create manual cashback earning
	earning := dto.CashbackEarning{
		ID:                uuid.New(),
		UserID:            userID,
		TierID:            uuid.New(), // Will be set by service based on user level
		EarningType:       "manual",
		SourceBetID:       nil,
		ExpectedGGRAmount: amount,
		CashbackRate:      decimal.NewFromFloat(0.05), // 5% default rate
		EarnedAmount:      amount,
		ClaimedAmount:     decimal.Zero,
		AvailableAmount:   amount,
		Status:            "available",
		ExpiresAt:         time.Now().AddDate(0, 0, 30), // 30 days expiry
		CreatedAt:         time.Now(),
	}

	createdEarning, err := s.storage.CreateCashbackEarning(ctx, earning)
	if err != nil {
		s.logger.Error("Failed to create manual cashback earning", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create manual cashback earning")
	}

	return &createdEarning, nil
}

// GetSystemHealth returns the health status of the cashback system
func (s *CashbackService) GetSystemHealth(ctx context.Context) (map[string]interface{}, error) {
	s.logger.Info("Checking cashback system health")

	// Check database connectivity
	_, err := s.storage.GetCashbackTiers(ctx)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		}, nil
	}

	// Check if tiers are properly configured
	tiers, err := s.storage.GetCashbackTiers(ctx)
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"tiers":  "error",
			"error":  err.Error(),
		}, nil
	}

	if len(tiers) == 0 {
		return map[string]interface{}{
			"status":   "warning",
			"database": "connected",
			"tiers":    "not_configured",
			"message":  "No cashback tiers configured",
		}, nil
	}

	return map[string]interface{}{
		"status":     "healthy",
		"database":   "connected",
		"tiers":      "configured",
		"tier_count": len(tiers),
		"timestamp":  time.Now(),
	}, nil
}

// GetUserLevel returns the user's current level
func (s *CashbackService) GetUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	s.logger.Info("Getting user level", zap.String("user_id", userID.String()))

	userLevel, err := s.storage.GetUserLevel(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user level", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user level")
	}

	return userLevel, nil
}

// GetUserCashbackEarnings returns the user's cashback earnings
func (s *CashbackService) GetUserCashbackEarnings(ctx context.Context, userID uuid.UUID) ([]dto.CashbackEarning, error) {
	s.logger.Info("Getting user cashback earnings", zap.String("user_id", userID.String()))

	earnings, err := s.storage.GetUserCashbackEarnings(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user cashback earnings", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user cashback earnings")
	}

	return earnings, nil
}

// GetUserCashbackClaims returns the user's cashback claims
func (s *CashbackService) GetUserCashbackClaims(ctx context.Context, userID uuid.UUID) ([]dto.CashbackClaim, error) {
	s.logger.Info("Getting user cashback claims", zap.String("user_id", userID.String()))

	claims, err := s.storage.GetUserCashbackClaims(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user cashback claims", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user cashback claims")
	}

	return claims, nil
}

// CreateCashbackTier creates a new cashback tier
func (s *CashbackService) CreateCashbackTier(ctx context.Context, tier dto.CashbackTier) (*dto.CashbackTier, error) {
	s.logger.Info("Creating cashback tier", zap.String("tier_name", tier.TierName))

	createdTier, err := s.storage.CreateCashbackTier(ctx, tier)
	if err != nil {
		s.logger.Error("Failed to create cashback tier", zap.Error(err))
		return nil, fmt.Errorf("failed to create cashback tier: %w", err)
	}

	s.logger.Info("Created cashback tier successfully",
		zap.String("tier_id", createdTier.ID.String()),
		zap.String("tier_name", createdTier.TierName))

	return createdTier, nil
}

// UpdateCashbackTier updates a cashback tier
func (s *CashbackService) UpdateCashbackTier(ctx context.Context, tierID uuid.UUID, tier dto.CashbackTier) (*dto.CashbackTier, error) {
	s.logger.Info("Updating cashback tier", zap.String("tier_id", tierID.String()))

	updatedTier, err := s.storage.UpdateCashbackTier(ctx, tierID, tier)
	if err != nil {
		s.logger.Error("Failed to update cashback tier", zap.Error(err))
		return nil, fmt.Errorf("failed to update cashback tier: %w", err)
	}

	s.logger.Info("Updated cashback tier successfully",
		zap.String("tier_id", tierID.String()),
		zap.String("tier_name", updatedTier.TierName))

	return updatedTier, nil
}

// DeleteCashbackTier deletes a cashback tier
func (s *CashbackService) DeleteCashbackTier(ctx context.Context, tierID uuid.UUID) error {
	s.logger.Info("Deleting cashback tier", zap.String("tier_id", tierID.String()))

	err := s.storage.DeleteCashbackTier(ctx, tierID)
	if err != nil {
		s.logger.Error("Failed to delete cashback tier", zap.Error(err))
		return fmt.Errorf("failed to delete cashback tier: %w", err)
	}

	s.logger.Info("Deleted cashback tier successfully", zap.String("tier_id", tierID.String()))
	return nil
}

// ReorderCashbackTiers reorders cashback tiers
func (s *CashbackService) ReorderCashbackTiers(ctx context.Context, tierOrder []uuid.UUID) error {
	s.logger.Info("Reordering cashback tiers", zap.Int("tier_count", len(tierOrder)))

	err := s.storage.ReorderCashbackTiers(ctx, tierOrder)
	if err != nil {
		s.logger.Error("Failed to reorder cashback tiers", zap.Error(err))
		return fmt.Errorf("failed to reorder cashback tiers: %w", err)
	}

	s.logger.Info("Reordered cashback tiers successfully")
	return nil
}

// CreateCashbackPromotion creates a new cashback promotion
func (s *CashbackService) CreateCashbackPromotion(ctx context.Context, promotion dto.CashbackPromotion) (*dto.CashbackPromotion, error) {
	s.logger.Info("Creating cashback promotion", zap.String("name", promotion.PromotionName))

	// This is a placeholder implementation
	// In a real implementation, you would create the promotion in the database
	promotion.ID = uuid.New()
	promotion.CreatedAt = time.Now()
	promotion.UpdatedAt = time.Now()

	return &promotion, nil
}

// ValidateBalanceSync validates balance synchronization for a user
func (s *CashbackService) ValidateBalanceSync(ctx context.Context, userID uuid.UUID) (*groove.BalanceSyncStatus, error) {
	s.logger.Info("Validating balance synchronization", zap.String("user_id", userID.String()))

	// Use GrooveStorage to validate balance sync
	status, err := s.grooveStorage.ValidateBalanceSync(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to validate balance sync", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to validate balance sync")
	}

	s.logger.Info("Balance sync validation completed",
		zap.String("user_id", userID.String()),
		zap.Bool("is_synchronized", status.IsSynchronized),
		zap.String("discrepancy", status.Discrepancy.String()))

	return status, nil
}

// ReconcileBalances reconciles user balances between main and GrooveTech systems
func (s *CashbackService) ReconcileBalances(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Reconciling balances", zap.String("user_id", userID.String()))

	// Use GrooveStorage to reconcile balances
	err := s.grooveStorage.ReconcileBalances(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to reconcile balances", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to reconcile balances")
	}

	s.logger.Info("Balance reconciliation completed successfully",
		zap.String("user_id", userID.String()))

	return nil
}

// GetGameHouseEdge returns the house edge configuration for a specific game type
func (s *CashbackService) GetGameHouseEdge(ctx context.Context, gameType, gameVariant string) (*dto.GameHouseEdge, error) {
	s.logger.Info("Getting game house edge", zap.String("game_type", gameType), zap.String("game_variant", gameVariant))

	houseEdge, err := s.storage.GetGameHouseEdge(ctx, gameType, gameVariant)
	if err != nil {
		s.logger.Error("Failed to get game house edge", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game house edge")
	}

	s.logger.Info("Retrieved game house edge",
		zap.String("game_type", gameType),
		zap.String("house_edge", houseEdge.HouseEdge.String()))

	return houseEdge, nil
}

// CreateGameHouseEdge creates a new game house edge configuration
func (s *CashbackService) CreateGameHouseEdge(ctx context.Context, houseEdge dto.GameHouseEdge) (*dto.GameHouseEdge, error) {
	s.logger.Info("Creating game house edge",
		zap.String("game_type", houseEdge.GameType),
		zap.String("house_edge", houseEdge.HouseEdge.String()))

	createdHouseEdge, err := s.storage.CreateGameHouseEdge(ctx, houseEdge)
	if err != nil {
		s.logger.Error("Failed to create game house edge", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create game house edge")
	}

	s.logger.Info("Game house edge created successfully",
		zap.String("game_type", houseEdge.GameType),
		zap.String("house_edge", houseEdge.HouseEdge.String()))

	return &createdHouseEdge, nil
}

// RetryFailedOperations retries all failed operations
func (s *CashbackService) RetryFailedOperations(ctx context.Context) error {
	s.logger.Info("Starting retry of failed operations")

	err := s.retryService.RetryFailedOperations(ctx)
	if err != nil {
		s.logger.Error("Failed to retry failed operations", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to retry failed operations")
	}

	s.logger.Info("Completed retry of failed operations")
	return nil
}

// GetRetryableOperations returns retryable operations for a user
func (s *CashbackService) GetRetryableOperations(ctx context.Context, userID uuid.UUID) ([]RetryableOperation, error) {
	s.logger.Info("Getting retryable operations", zap.String("user_id", userID.String()))

	operations, err := s.retryService.GetRetryableOperations(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get retryable operations", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get retryable operations")
	}

	s.logger.Info("Retrieved retryable operations", zap.Int("count", len(operations)))
	return operations, nil
}

// ManualRetryOperation manually retries a specific operation
func (s *CashbackService) ManualRetryOperation(ctx context.Context, operationID uuid.UUID) error {
	s.logger.Info("Manually retrying operation", zap.String("operation_id", operationID.String()))

	err := s.retryService.ManualRetryOperation(ctx, operationID)
	if err != nil {
		s.logger.Error("Failed to manually retry operation", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to manually retry operation")
	}

	s.logger.Info("Manual retry completed successfully", zap.String("operation_id", operationID.String()))
	return nil
}

// extractGameVariantFromTransaction extracts game variant information from transaction context
func (s *CashbackService) extractGameVariantFromTransaction(ctx context.Context, transactionID string) string {
	s.logger.Info("Extracting game variant from transaction",
		zap.String("transaction_id", transactionID))

	// Get game information from GrooveTech transaction
	gameID, gameType, err := s.grooveStorage.GetTransactionGameInfo(ctx, transactionID)
	if err != nil {
		s.logger.Warn("Failed to extract game variant from transaction",
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return ""
	}

	// Return the game ID as the variant for specific house edge lookup
	// This allows us to have game-specific RTP configurations
	// Use just the game ID as the variant since that's how it's stored in the database
	gameVariant := gameID

	s.logger.Info("Extracted game variant from transaction",
		zap.String("transaction_id", transactionID),
		zap.String("game_id", gameID),
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant))

	return gameVariant
}

// CheckAndProcessLevelProgression checks and processes automatic level progression for a user
func (s *CashbackService) CheckAndProcessLevelProgression(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	s.logger.Info("Checking level progression for user", zap.String("user_id", userID.String()))

	return s.levelProgressionService.CheckAndProcessLevelProgression(ctx, userID)
}

// GetLevelProgressionInfo returns detailed level progression information for a user
func (s *CashbackService) GetLevelProgressionInfo(ctx context.Context, userID uuid.UUID) (*dto.LevelProgressionInfo, error) {
	s.logger.Info("Getting level progression info", zap.String("user_id", userID.String()))

	return s.levelProgressionService.GetLevelProgressionInfo(ctx, userID)
}

// createEnhancedCashbackData creates enhanced cashback data with game-specific information
func (s *CashbackService) createEnhancedCashbackData(ctx context.Context, baseSummary dto.UserCashbackSummary, bet dto.Bet, houseEdge, cashbackRate decimal.Decimal) dto.EnhancedUserCashbackSummary {
	enhanced := dto.EnhancedUserCashbackSummary{
		UserID:            baseSummary.UserID,
		CurrentTier:       baseSummary.CurrentTier,
		LevelProgress:     baseSummary.LevelProgress,
		TotalGGR:          baseSummary.TotalGGR,
		AvailableCashback: baseSummary.AvailableCashback,
		PendingCashback:   baseSummary.PendingCashback,
		TotalClaimed:      baseSummary.TotalClaimed,
		NextTierGGR:       baseSummary.NextTierGGR,
		DailyLimit:        baseSummary.DailyLimit,
		WeeklyLimit:       baseSummary.WeeklyLimit,
		MonthlyLimit:      baseSummary.MonthlyLimit,
		SpecialBenefits:   baseSummary.SpecialBenefits,
	}

	// Extract game information from bet
	gameType := "groovetech"
	gameVariant := ""
	gameID := ""
	gameName := "Unknown Game"

	if bet.ClientTransactionID != "" {
		// Try to get game info from GrooveTech transaction
		if extractedGameID, extractedGameType, err := s.grooveStorage.GetTransactionGameInfo(ctx, bet.ClientTransactionID); err == nil {
			gameID = extractedGameID
			gameType = extractedGameType
			gameVariant = gameID
			gameName = s.getGameDisplayName(gameID, gameType)
		}
	}

	// Calculate expected GGR based on house edge - kept for display purposes
	expectedGGR := bet.Amount.Mul(houseEdge)

	// Use game house edge as cashback rate (per-wager cashback)
	cashbackRate = houseEdge.Mul(decimal.NewFromInt(100)) // Convert house edge to percentage

	// Calculate earned cashback per wager (bet amount * house edge)
	earnedCashback := bet.Amount.Mul(houseEdge)
	earnedCashback = earnedCashback.Round(2) // Round to 2 decimal places

	// Create game-specific information
	gameInfo := dto.GameCashbackInfo{
		GameID:           gameID,
		GameName:         gameName,
		GameType:         gameType,
		GameVariant:      gameVariant,
		HouseEdge:        houseEdge,
		HouseEdgePercent: fmt.Sprintf("%.2f%%", houseEdge.Mul(decimal.NewFromInt(100)).InexactFloat64()),
		CashbackRate:     cashbackRate,
		CashbackPercent:  fmt.Sprintf("%.1f%%", cashbackRate.InexactFloat64()),
		ExpectedGGR:      expectedGGR, // Store expected GGR for display purposes
		EarnedCashback:   earnedCashback,
		BetAmount:        bet.Amount,
		TransactionID:    bet.ClientTransactionID,
		Timestamp:        time.Now(),
	}

	enhanced.LastGameInfo = &gameInfo

	return enhanced
}

// getGameDisplayName returns a user-friendly display name for a game
func (s *CashbackService) getGameDisplayName(gameID, gameType string) string {
	// This could be enhanced to fetch from a games database or configuration
	// For now, return a formatted name based on game ID and type
	if gameID != "" {
		return fmt.Sprintf("%s Game %s", strings.Title(gameType), gameID)
	}
	return fmt.Sprintf("%s Game", strings.Title(gameType))
}

// GetGlobalRakebackOverride retrieves the current global rakeback override configuration
func (s *CashbackService) GetGlobalRakebackOverride(ctx context.Context) (*dto.GlobalCashbackOverride, error) {
	s.logger.Info("Getting global rakeback override configuration")

	override, err := s.storage.GetGlobalRakebackOverride(ctx)
	if err != nil {
		s.logger.Error("Failed to get global rakeback override", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get global rakeback override")
	}

	return override, nil
}

// UpdateGlobalRakebackOverride updates the global rakeback override configuration
func (s *CashbackService) UpdateGlobalRakebackOverride(ctx context.Context, adminUserID uuid.UUID, request dto.GlobalCashbackOverrideRequest) (*dto.GlobalCashbackOverride, error) {
	s.logger.Info("Updating global rakeback override",
		zap.String("admin_user_id", adminUserID.String()),
		zap.Bool("is_enabled", request.IsEnabled),
		zap.String("override_percentage", request.OverridePercentage.String()))

	// Get current override configuration
	currentOverride, err := s.storage.GetGlobalRakebackOverride(ctx)
	if err != nil {
		s.logger.Error("Failed to get current override configuration", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get current override configuration")
	}

	// Prepare the update
	now := time.Now()
	updatedOverride := dto.GlobalCashbackOverride{
		ID:                 currentOverride.ID,
		IsEnabled:          request.IsEnabled,
		OverridePercentage: request.OverridePercentage,
		CreatedAt:          currentOverride.CreatedAt,
		UpdatedAt:          now,
	}

	if request.IsEnabled {
		// Enabling override
		updatedOverride.EnabledBy = &adminUserID
		updatedOverride.EnabledAt = &now
		updatedOverride.DisabledBy = nil
		updatedOverride.DisabledAt = nil
	} else {
		// Disabling override - preserve enabled info, add disabled info
		updatedOverride.EnabledBy = currentOverride.EnabledBy
		updatedOverride.EnabledAt = currentOverride.EnabledAt
		updatedOverride.DisabledBy = &adminUserID
		updatedOverride.DisabledAt = &now
	}

	// Update in database
	updated, err := s.storage.UpdateGlobalRakebackOverride(ctx, updatedOverride)
	if err != nil {
		s.logger.Error("Failed to update global rakeback override", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update global rakeback override")
	}

	s.logger.Info("Global rakeback override updated successfully",
		zap.Bool("is_enabled", updated.IsEnabled),
		zap.String("override_percentage", updated.OverridePercentage.String()),
		zap.String("admin_user_id", adminUserID.String()))

	return updated, nil
}

// ProcessBulkLevelProgression processes level progression for multiple users
func (s *CashbackService) ProcessBulkLevelProgression(ctx context.Context, userIDs []uuid.UUID) ([]dto.LevelProgressionResult, error) {
	s.logger.Info("Processing bulk level progression", zap.Int("user_count", len(userIDs)))

	return s.levelProgressionService.ProcessBulkLevelProgression(ctx, userIDs)
}

// CreateLevelProgressionResult creates a detailed result for level progression
func (s *CashbackService) CreateLevelProgressionResult(ctx context.Context, userID uuid.UUID) dto.LevelProgressionResult {
	s.logger.Info("Creating level progression result", zap.String("user_id", userID.String()))

	return s.levelProgressionService.CreateLevelProgressionResult(ctx, userID)
}
