package cashback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/cashback"
	"github.com/tucanbit/internal/storage/groove"
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
	logger                  *zap.Logger
}

func NewCashbackService(storage cashback.CashbackStorage, grooveStorage groove.GrooveStorage, logger *zap.Logger) *CashbackService {
	// Create the service first
	service := &CashbackService{
		storage:       storage,
		grooveStorage: grooveStorage,
		logger:        logger,
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
		UserID:        userID,
		CurrentLevel:  1,
		TotalGGR:      decimal.Zero,
		TotalBets:     decimal.Zero,
		TotalWins:     decimal.Zero,
		LevelProgress: decimal.Zero,
		CurrentTierID: defaultTier.ID,
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

	// Get current tier to get cashback rate
	currentTier, err := s.storage.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel)
	if err != nil {
		s.logger.Error("Failed to get current tier", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get current tier")
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

	// Calculate expected GGR
	expectedGGR := bet.Amount.Mul(houseEdge)

	// Get current cashback rate
	cashbackRate := currentTier.CashbackPercentage

	// Check for active promotions (skip for now)
	// promotion, err := s.storage.GetCashbackPromotionForUser(ctx, userLevel.CurrentLevel, "default")
	// if err == nil && promotion != nil {
	// 	// Apply promotion boost
	// 	cashbackRate = cashbackRate.Mul(promotion.BoostMultiplier)
	// 	s.logger.Info("Applied promotion boost",
	// 		zap.String("promotion", promotion.PromotionName),
	// 		zap.String("boost", promotion.BoostMultiplier.String()))
	// }

	// Calculate earned cashback
	earnedCashback := expectedGGR.Mul(cashbackRate.Div(decimal.NewFromInt(100)))

	// Create cashback earning
	cashbackEarning := dto.CashbackEarning{
		UserID:          bet.UserID,
		TierID:          userLevel.CurrentTierID,
		EarningType:     "bet",
		SourceBetID:     &bet.BetID,
		GGRAmount:       expectedGGR,
		CashbackRate:    cashbackRate,
		EarnedAmount:    earnedCashback,
		ClaimedAmount:   decimal.Zero,
		AvailableAmount: earnedCashback,
		Status:          "available",
		ExpiresAt:       time.Now().Add(30 * 24 * time.Hour), // 30 days
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
			UserID:        bet.UserID,
			CurrentLevel:  userLevel.CurrentLevel,
			TotalGGR:      userLevel.TotalGGR.Add(expectedGGR),
			TotalBets:     userLevel.TotalBets.Add(bet.Amount),
			TotalWins:     userLevel.TotalWins.Add(bet.Payout),
			LevelProgress: userLevel.LevelProgress,
			CurrentTierID: userLevel.CurrentTierID,
		},
	}, func() error {
		// Create cashback earning
		_, err := s.storage.CreateCashbackEarning(ctx, cashbackEarning)
		if err != nil {
			return fmt.Errorf("failed to create cashback earning: %w", err)
		}

		// Update user level statistics
		updatedUserLevel := dto.UserLevel{
			UserID:        bet.UserID,
			CurrentLevel:  userLevel.CurrentLevel,
			TotalGGR:      userLevel.TotalGGR.Add(expectedGGR),
			TotalBets:     userLevel.TotalBets.Add(bet.Amount),
			TotalWins:     userLevel.TotalWins.Add(bet.Payout), // We use payout as win amount
			LevelProgress: userLevel.LevelProgress,
			CurrentTierID: userLevel.CurrentTierID,
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

	s.logger.Info("Bet cashback processed successfully",
		zap.String("bet_id", bet.BetID.String()),
		zap.String("expected_ggr", expectedGGR.String()),
		zap.String("earned_cashback", earnedCashback.String()),
		zap.String("cashback_rate", cashbackRate.String()))

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

	// Get current tier to check limits
	currentTier, err := s.storage.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel)
	if err != nil {
		s.logger.Error("Failed to get current tier", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get current tier")
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
		updatedEarning := dto.CashbackEarning{
			ID:              earning.ID,
			UserID:          earning.UserID,
			TierID:          earning.TierID,
			EarningType:     earning.EarningType,
			SourceBetID:     earning.SourceBetID,
			GGRAmount:       earning.GGRAmount,
			CashbackRate:    earning.CashbackRate,
			EarnedAmount:    earning.EarnedAmount,
			ClaimedAmount:   earning.ClaimedAmount.Add(claimAmount),
			AvailableAmount: earning.AvailableAmount.Sub(claimAmount),
			Status:          "claimed",
			ExpiresAt:       earning.ExpiresAt,
			ClaimedAt:       &time.Time{},
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
	netAmount := request.Amount.Sub(processingFee)
	_, err = s.grooveStorage.AddBalance(ctx, userID, netAmount)
	if err != nil {
		s.logger.Error("Failed to credit user balance", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to credit user balance")
	}

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
		ID:              uuid.New(),
		UserID:          userID,
		TierID:          uuid.New(), // Will be set by service based on user level
		EarningType:     "manual",
		SourceBetID:     nil,
		GGRAmount:       amount,
		CashbackRate:    decimal.NewFromFloat(0.05), // 5% default rate
		EarnedAmount:    amount,
		ClaimedAmount:   decimal.Zero,
		AvailableAmount: amount,
		Status:          "available",
		ExpiresAt:       time.Now().AddDate(0, 0, 30), // 30 days expiry
		CreatedAt:       time.Now(),
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

// UpdateCashbackTier updates a cashback tier
func (s *CashbackService) UpdateCashbackTier(ctx context.Context, tierID uuid.UUID, tier dto.CashbackTier) (*dto.CashbackTier, error) {
	s.logger.Info("Updating cashback tier", zap.String("tier_id", tierID.String()))

	// This is a placeholder implementation
	// In a real implementation, you would update the tier in the database
	updatedTier := &tier
	updatedTier.ID = tierID
	updatedTier.UpdatedAt = time.Now()

	return updatedTier, nil
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
	// This method would typically query the groove_transactions table
	// to get the game_id that was used in the original transaction
	// For now, we'll return a placeholder that can be enhanced later

	s.logger.Info("Extracting game variant from transaction",
		zap.String("transaction_id", transactionID))

	// TODO: Implement actual game variant extraction from groove_transactions table
	// This would involve:
	// 1. Querying groove_transactions table by transaction_id
	// 2. Getting the game_id from the transaction metadata
	// 3. Mapping game_id to game variant (slot, table, live, etc.)

	// For now, return empty string to use default categorization
	return ""
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

// ProcessBulkLevelProgression processes level progression for multiple users
func (s *CashbackService) ProcessBulkLevelProgression(ctx context.Context, userIDs []uuid.UUID) ([]dto.LevelProgressionResult, error) {
	s.logger.Info("Processing bulk level progression", zap.Int("user_count", len(userIDs)))

	return s.levelProgressionService.ProcessBulkLevelProgression(ctx, userIDs)
}
