package cashback

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/cashback"
	"go.uber.org/zap"
)

type CashbackService struct {
	storage cashback.CashbackStorage
	logger  *zap.Logger
}

func NewCashbackService(storage cashback.CashbackStorage, logger *zap.Logger) *CashbackService {
	return &CashbackService{
		storage: storage,
		logger:  logger,
	}
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

	// Get user level
	userLevel, err := s.storage.GetUserLevel(ctx, bet.UserID)
	if err != nil {
		s.logger.Error("Failed to get user level", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get user level")
	}

	// Get current tier to get cashback rate
	currentTier, err := s.storage.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel)
	if err != nil {
		s.logger.Error("Failed to get current tier", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get current tier")
	}

	// Use default house edge for now (2%)
	houseEdge := decimal.NewFromFloat(0.02)

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

	_, err = s.storage.CreateCashbackEarning(ctx, cashbackEarning)
	if err != nil {
		s.logger.Error("Failed to create cashback earning", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to create cashback earning")
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
		s.logger.Error("Failed to update user level", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update user level")
	}

	s.logger.Info("Bet cashback processed successfully",
		zap.String("bet_id", bet.BetID.String()),
		zap.String("expected_ggr", expectedGGR.String()),
		zap.String("earned_cashback", earnedCashback.String()),
		zap.String("cashback_rate", cashbackRate.String()))

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

	// Create cashback claim record
	cashbackClaim := dto.CashbackClaim{
		ID:              claimID,
		UserID:          userID,
		ClaimAmount:     request.Amount,
		CurrencyCode:    "USD",
		Status:          "pending",
		ProcessingFee:   processingFee,
		NetAmount:       request.Amount.Sub(processingFee),
		ClaimedEarnings: claimedEarnings,
	}

	_, err = s.storage.CreateCashbackClaim(ctx, cashbackClaim)
	if err != nil {
		s.logger.Error("Failed to create cashback claim", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create cashback claim")
	}

	response := &dto.CashbackClaimResponse{
		ClaimID:       claimID,
		Amount:        request.Amount,
		NetAmount:     request.Amount.Sub(processingFee),
		ProcessingFee: processingFee,
		Status:        "pending",
		Message:       "Cashback claim submitted successfully",
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
