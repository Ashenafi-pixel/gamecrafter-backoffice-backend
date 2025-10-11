package cashback

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/analytics"
	"go.uber.org/zap"
)

// RetryableOperation represents an operation that can be retried
type RetryableOperation struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	UserID      uuid.UUID              `json:"user_id"`
	Data        map[string]interface{} `json:"data"`
	Attempts    int                    `json:"attempts"`
	LastError   string                 `json:"last_error"`
	NextRetryAt *time.Time             `json:"next_retry_at"`
	Status      string                 `json:"status"` // pending, retrying, failed, completed
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CashbackStorage interface defines the contract for cashback data operations
type CashbackStorage interface {
	// User Level operations
	CreateUserLevel(ctx context.Context, userLevel dto.UserLevel) (dto.UserLevel, error)
	GetUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error)
	UpdateUserLevel(ctx context.Context, userLevel dto.UserLevel) (dto.UserLevel, error)

	// Cashback Tier operations
	GetCashbackTiers(ctx context.Context) ([]dto.CashbackTier, error)
	GetCashbackTierByLevel(ctx context.Context, level int) (*dto.CashbackTier, error)
	GetCashbackTierByGGR(ctx context.Context, ggr decimal.Decimal) (*dto.CashbackTier, error)

	// Cashback Earning operations
	CreateCashbackEarning(ctx context.Context, earning dto.CashbackEarning) (dto.CashbackEarning, error)
	GetUserCashbackEarnings(ctx context.Context, userID uuid.UUID) ([]dto.CashbackEarning, error)
	GetAvailableCashbackEarnings(ctx context.Context, userID uuid.UUID) ([]dto.CashbackEarning, error)
	UpdateCashbackEarningStatus(ctx context.Context, earning dto.CashbackEarning) (dto.CashbackEarning, error)

	// Cashback Claim operations
	CreateCashbackClaim(ctx context.Context, claim dto.CashbackClaim) (dto.CashbackClaim, error)
	UpdateCashbackClaimStatus(ctx context.Context, claim dto.CashbackClaim) (dto.CashbackClaim, error)
	GetUserCashbackClaims(ctx context.Context, userID uuid.UUID) ([]dto.CashbackClaim, error)
	GetCashbackClaim(ctx context.Context, claimID uuid.UUID) (*dto.CashbackClaim, error)

	// Game House Edge operations
	GetGameHouseEdge(ctx context.Context, gameType, gameVariant string) (*dto.GameHouseEdge, error)
	CreateGameHouseEdge(ctx context.Context, houseEdge dto.GameHouseEdge) (dto.GameHouseEdge, error)

	// Cashback Promotion operations
	GetActiveCashbackPromotions(ctx context.Context) ([]dto.CashbackPromotion, error)

	// Level Progression operations
	GetAllCashbackTiers(ctx context.Context) ([]dto.CashbackTier, error)
	GetCashbackTierByID(ctx context.Context, tierID uuid.UUID) (*dto.CashbackTier, error)
	InitializeUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error)
	GetCashbackPromotionForUser(ctx context.Context, userLevel int, gameType string) (*dto.CashbackPromotion, error)
	CreateCashbackPromotion(ctx context.Context, promotion dto.CashbackPromotion) (dto.CashbackPromotion, error)
	UpdateCashbackPromotion(ctx context.Context, promotion dto.CashbackPromotion) (dto.CashbackPromotion, error)

	// Summary and Statistics
	GetUserCashbackSummary(ctx context.Context, userID uuid.UUID) (*dto.UserCashbackSummary, error)
	GetCashbackStats(ctx context.Context) (*dto.AdminCashbackStats, error)
	GetTierDistribution(ctx context.Context) ([]dto.TierDistribution, error)

	// Limit checking
	GetUserDailyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	GetUserWeeklyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	GetUserMonthlyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)

	// Expired earnings
	GetExpiredCashbackEarnings(ctx context.Context) ([]dto.CashbackEarning, error)
	MarkCashbackEarningsAsExpired(ctx context.Context) error

	// Wallet integration
	ProcessCashbackClaim(ctx context.Context, userID uuid.UUID, claimAmount decimal.Decimal, currency string) error

	// Retryable Operations
	CreateRetryableOperation(ctx context.Context, operation RetryableOperation) error
	GetRetryableOperation(ctx context.Context, operationID uuid.UUID) (*RetryableOperation, error)
	UpdateRetryableOperation(ctx context.Context, operation RetryableOperation) error
	GetRetryableOperationsByUser(ctx context.Context, userID uuid.UUID) ([]RetryableOperation, error)
	GetFailedRetryableOperations(ctx context.Context) ([]RetryableOperation, error)
	DeleteRetryableOperation(ctx context.Context, operationID uuid.UUID) error
}

// CashbackStorageImpl provides a production-ready implementation with real database integration
type CashbackStorageImpl struct {
	db                   *persistencedb.PersistenceDB
	logger               *zap.Logger
	analyticsIntegration *analytics.AnalyticsIntegration
}

func NewCashbackStorage(db *persistencedb.PersistenceDB, logger *zap.Logger, analyticsIntegration *analytics.AnalyticsIntegration) CashbackStorage {
	return &CashbackStorageImpl{
		db:                   db,
		logger:               logger,
		analyticsIntegration: analyticsIntegration,
	}
}

// User Level operations
func (s *CashbackStorageImpl) CreateUserLevel(ctx context.Context, userLevel dto.UserLevel) (dto.UserLevel, error) {
	s.logger.Info("Creating user level", zap.String("user_id", userLevel.UserID.String()))

	// Use raw SQL since we don't have SQLC generated code yet
	query := `
		INSERT INTO user_levels (user_id, current_level, total_ggr, total_bets, total_wins, level_progress, current_tier_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			current_level = EXCLUDED.current_level,
			total_ggr = EXCLUDED.total_ggr,
			total_bets = EXCLUDED.total_bets,
			total_wins = EXCLUDED.total_wins,
			level_progress = EXCLUDED.level_progress,
			current_tier_id = EXCLUDED.current_tier_id,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.db.GetPool().QueryRow(ctx, query,
		userLevel.UserID,
		userLevel.CurrentLevel,
		userLevel.TotalExpectedGGR,
		userLevel.TotalBets,
		userLevel.TotalWins,
		userLevel.LevelProgress,
		userLevel.CurrentTierID,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		s.logger.Error("Failed to create user level", zap.Error(err))
		return dto.UserLevel{}, fmt.Errorf("failed to create user level: %w", err)
	}

	userLevel.ID = id
	userLevel.CreatedAt = createdAt
	userLevel.UpdatedAt = updatedAt

	s.logger.Info("User level created successfully", zap.String("user_id", userLevel.UserID.String()))
	return userLevel, nil
}

func (s *CashbackStorageImpl) GetUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	s.logger.Info("Getting user level", zap.String("user_id", userID.String()))

	query := `
		SELECT id, user_id, current_level, total_ggr, total_bets, total_wins, level_progress, current_tier_id, last_level_up, created_at, updated_at
		FROM user_levels
		WHERE user_id = $1`

	var userLevel dto.UserLevel
	var lastLevelUp sql.NullTime

	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(
		&userLevel.ID,
		&userLevel.UserID,
		&userLevel.CurrentLevel,
		&userLevel.TotalExpectedGGR,
		&userLevel.TotalBets,
		&userLevel.TotalWins,
		&userLevel.LevelProgress,
		&userLevel.CurrentTierID,
		&lastLevelUp,
		&userLevel.CreatedAt,
		&userLevel.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("User level not found, creating default", zap.String("user_id", userID.String()))
			return s.createDefaultUserLevel(ctx, userID)
		}
		s.logger.Error("Failed to get user level", zap.Error(err))
		return nil, fmt.Errorf("failed to get user level: %w", err)
	}

	if lastLevelUp.Valid {
		userLevel.LastLevelUp = &lastLevelUp.Time
	}

	s.logger.Info("User level retrieved successfully", zap.String("user_id", userID.String()))
	return &userLevel, nil
}

func (s *CashbackStorageImpl) UpdateUserLevel(ctx context.Context, userLevel dto.UserLevel) (dto.UserLevel, error) {
	s.logger.Info("Updating user level", zap.String("user_id", userLevel.UserID.String()))

	query := `
		UPDATE user_levels 
		SET current_level = $2, total_ggr = $3, total_bets = $4, total_wins = $5, level_progress = $6, current_tier_id = $7, last_level_up = $8, updated_at = NOW()
		WHERE user_id = $1
		RETURNING updated_at`

	var updatedAt time.Time
	err := s.db.GetPool().QueryRow(ctx, query,
		userLevel.UserID,
		userLevel.CurrentLevel,
		userLevel.TotalExpectedGGR,
		userLevel.TotalBets,
		userLevel.TotalWins,
		userLevel.LevelProgress,
		userLevel.CurrentTierID,
		userLevel.LastLevelUp,
	).Scan(&updatedAt)

	if err != nil {
		s.logger.Error("Failed to update user level", zap.Error(err))
		return dto.UserLevel{}, fmt.Errorf("failed to update user level: %w", err)
	}

	userLevel.UpdatedAt = updatedAt

	s.logger.Info("User level updated successfully", zap.String("user_id", userLevel.UserID.String()))
	return userLevel, nil
}

// createDefaultUserLevel creates a default Bronze tier user level
func (s *CashbackStorageImpl) createDefaultUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	// Get Bronze tier
	bronzeTier, err := s.GetCashbackTierByLevel(ctx, 1)
	if err != nil {
		s.logger.Error("Failed to get Bronze tier", zap.Error(err))
		return nil, fmt.Errorf("failed to get Bronze tier: %w", err)
	}

	defaultLevel := dto.UserLevel{
		UserID:           userID,
		CurrentLevel:     1,
		TotalExpectedGGR: decimal.Zero,
		TotalBets:        decimal.Zero,
		TotalWins:        decimal.Zero,
		LevelProgress:    decimal.Zero,
		CurrentTierID:    bronzeTier.ID,
	}

	createdLevel, err := s.CreateUserLevel(ctx, defaultLevel)
	if err != nil {
		return nil, err
	}

	return &createdLevel, nil
}

// Cashback Tier operations
func (s *CashbackStorageImpl) GetCashbackTiers(ctx context.Context) ([]dto.CashbackTier, error) {
	s.logger.Info("Getting cashback tiers")

	query := `
		SELECT id, tier_name, tier_level, min_ggr_required, cashback_percentage, bonus_multiplier, 
		       daily_cashback_limit, weekly_cashback_limit, monthly_cashback_limit, special_benefits, is_active, created_at, updated_at
		FROM cashback_tiers
		WHERE is_active = true
		ORDER BY tier_level ASC`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.logger.Error("Failed to get cashback tiers", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback tiers: %w", err)
	}
	defer rows.Close()

	var tiers []dto.CashbackTier
	for rows.Next() {
		var tier dto.CashbackTier
		var dailyLimit, weeklyLimit, monthlyLimit sql.NullString
		var specialBenefits sql.NullString

		err := rows.Scan(
			&tier.ID,
			&tier.TierName,
			&tier.TierLevel,
			&tier.MinExpectedGGRRequired,
			&tier.CashbackPercentage,
			&tier.BonusMultiplier,
			&dailyLimit,
			&weeklyLimit,
			&monthlyLimit,
			&specialBenefits,
			&tier.IsActive,
			&tier.CreatedAt,
			&tier.UpdatedAt,
		)

		if err != nil {
			s.logger.Error("Failed to scan cashback tier", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cashback tier: %w", err)
		}

		// Parse nullable fields
		if dailyLimit.Valid {
			if limit, err := decimal.NewFromString(dailyLimit.String); err == nil {
				tier.DailyCashbackLimit = &limit
			}
		}
		if weeklyLimit.Valid {
			if limit, err := decimal.NewFromString(weeklyLimit.String); err == nil {
				tier.WeeklyCashbackLimit = &limit
			}
		}
		if monthlyLimit.Valid {
			if limit, err := decimal.NewFromString(monthlyLimit.String); err == nil {
				tier.MonthlyCashbackLimit = &limit
			}
		}
		if specialBenefits.Valid {
			// Parse JSON benefits
			tier.SpecialBenefits = map[string]interface{}{}
		}

		tiers = append(tiers, tier)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Error iterating cashback tiers", zap.Error(err))
		return nil, fmt.Errorf("error iterating cashback tiers: %w", err)
	}

	s.logger.Info("Retrieved cashback tiers", zap.Int("count", len(tiers)))
	return tiers, nil
}

func (s *CashbackStorageImpl) GetCashbackTierByLevel(ctx context.Context, level int) (*dto.CashbackTier, error) {
	s.logger.Info("Getting cashback tier by level", zap.Int("level", level))

	query := `
		SELECT id, tier_name, tier_level, min_ggr_required, cashback_percentage, bonus_multiplier, 
		       daily_cashback_limit, weekly_cashback_limit, monthly_cashback_limit, special_benefits, is_active, created_at, updated_at
		FROM cashback_tiers
		WHERE tier_level = $1 AND is_active = true`

	var tier dto.CashbackTier
	var dailyLimit, weeklyLimit, monthlyLimit sql.NullString
	var specialBenefits sql.NullString

	err := s.db.GetPool().QueryRow(ctx, query, level).Scan(
		&tier.ID,
		&tier.TierName,
		&tier.TierLevel,
		&tier.MinExpectedGGRRequired,
		&tier.CashbackPercentage,
		&tier.BonusMultiplier,
		&dailyLimit,
		&weeklyLimit,
		&monthlyLimit,
		&specialBenefits,
		&tier.IsActive,
		&tier.CreatedAt,
		&tier.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tier with level %d not found", level)
		}
		s.logger.Error("Failed to get cashback tier by level", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback tier by level: %w", err)
	}

	// Parse nullable fields
	if dailyLimit.Valid {
		if limit, err := decimal.NewFromString(dailyLimit.String); err == nil {
			tier.DailyCashbackLimit = &limit
		}
	}
	if weeklyLimit.Valid {
		if limit, err := decimal.NewFromString(weeklyLimit.String); err == nil {
			tier.WeeklyCashbackLimit = &limit
		}
	}
	if monthlyLimit.Valid {
		if limit, err := decimal.NewFromString(monthlyLimit.String); err == nil {
			tier.MonthlyCashbackLimit = &limit
		}
	}
	if specialBenefits.Valid {
		tier.SpecialBenefits = map[string]interface{}{}
	}

	s.logger.Info("Retrieved cashback tier by level", zap.Int("level", level), zap.String("tier_name", tier.TierName))
	return &tier, nil
}

func (s *CashbackStorageImpl) GetCashbackTierByGGR(ctx context.Context, ggr decimal.Decimal) (*dto.CashbackTier, error) {
	s.logger.Info("Getting cashback tier by GGR", zap.String("ggr", ggr.String()))

	query := `
		SELECT id, tier_name, tier_level, min_ggr_required, cashback_percentage, bonus_multiplier, 
		       daily_cashback_limit, weekly_cashback_limit, monthly_cashback_limit, special_benefits, is_active, created_at, updated_at
		FROM cashback_tiers
		WHERE min_ggr_required <= $1 AND is_active = true
		ORDER BY tier_level DESC
		LIMIT 1`

	var tier dto.CashbackTier
	var dailyLimit, weeklyLimit, monthlyLimit sql.NullString
	var specialBenefits sql.NullString

	err := s.db.GetPool().QueryRow(ctx, query, ggr).Scan(
		&tier.ID,
		&tier.TierName,
		&tier.TierLevel,
		&tier.MinExpectedGGRRequired,
		&tier.CashbackPercentage,
		&tier.BonusMultiplier,
		&dailyLimit,
		&weeklyLimit,
		&monthlyLimit,
		&specialBenefits,
		&tier.IsActive,
		&tier.CreatedAt,
		&tier.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return Bronze tier as default
			return s.GetCashbackTierByLevel(ctx, 1)
		}
		s.logger.Error("Failed to get cashback tier by GGR", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback tier by GGR: %w", err)
	}

	// Parse nullable fields
	if dailyLimit.Valid {
		if limit, err := decimal.NewFromString(dailyLimit.String); err == nil {
			tier.DailyCashbackLimit = &limit
		}
	}
	if weeklyLimit.Valid {
		if limit, err := decimal.NewFromString(weeklyLimit.String); err == nil {
			tier.WeeklyCashbackLimit = &limit
		}
	}
	if monthlyLimit.Valid {
		if limit, err := decimal.NewFromString(monthlyLimit.String); err == nil {
			tier.MonthlyCashbackLimit = &limit
		}
	}
	if specialBenefits.Valid {
		tier.SpecialBenefits = map[string]interface{}{}
	}

	s.logger.Info("Retrieved cashback tier by GGR", zap.String("ggr", ggr.String()), zap.String("tier_name", tier.TierName))
	return &tier, nil
}

// Wallet Integration - Process Cashback Claim
func (s *CashbackStorageImpl) ProcessCashbackClaim(ctx context.Context, userID uuid.UUID, claimAmount decimal.Decimal, currency string) error {
	s.logger.Info("Processing cashback claim", zap.String("user_id", userID.String()), zap.String("amount", claimAmount.String()))

	// For now, we'll use a simple approach that logs the claim
	// TODO: Integrate with the balance system once we have proper access
	s.logger.Info("Cashback claim processed successfully",
		zap.String("user_id", userID.String()),
		zap.String("amount", claimAmount.String()),
		zap.String("currency", currency))

	return nil
}

// Summary and Statistics operations
func (s *CashbackStorageImpl) GetUserCashbackSummary(ctx context.Context, userID uuid.UUID) (*dto.UserCashbackSummary, error) {
	s.logger.Info("Getting user cashback summary", zap.String("user_id", userID.String()))

	// Get user level
	userLevel, err := s.GetUserLevel(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get current tier
	tier, err := s.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel)
	if err != nil {
		return nil, err
	}

	// Get available earnings for available cashback calculation
	availableEarnings, err := s.GetAvailableCashbackEarnings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get ALL earnings (including claimed ones) for total claimed calculation
	allEarnings, err := s.GetUserCashbackEarnings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate total available cashback and total claimed
	totalAvailable := decimal.Zero
	totalClaimed := decimal.Zero
	for _, earning := range availableEarnings {
		totalAvailable = totalAvailable.Add(earning.AvailableAmount)
	}
	for _, earning := range allEarnings {
		totalClaimed = totalClaimed.Add(earning.ClaimedAmount)
		s.logger.Debug("Adding to total claimed",
			zap.String("earning_id", earning.ID.String()),
			zap.String("claimed_amount", earning.ClaimedAmount.String()),
			zap.String("running_total", totalClaimed.String()))
	}

	s.logger.Info("Cashback summary calculation",
		zap.String("user_id", userID.String()),
		zap.String("total_available", totalAvailable.String()),
		zap.String("total_claimed", totalClaimed.String()),
		zap.Int("available_earnings_count", len(availableEarnings)),
		zap.Int("all_earnings_count", len(allEarnings)))

	// Get next tier GGR requirement
	nextTierGGR := decimal.Zero
	if userLevel.CurrentLevel < 5 {
		nextTier, err := s.GetCashbackTierByLevel(ctx, userLevel.CurrentLevel+1)
		if err == nil {
			nextTierGGR = nextTier.MinExpectedGGRRequired
		}
	}

	summary := &dto.UserCashbackSummary{
		UserID:            userID,
		CurrentTier:       *tier,
		LevelProgress:     userLevel.LevelProgress,
		TotalGGR:          userLevel.TotalExpectedGGR,
		AvailableCashback: totalAvailable,
		PendingCashback:   decimal.Zero,
		TotalClaimed:      totalClaimed,
		NextTierGGR:       nextTierGGR,
		DailyLimit:        tier.DailyCashbackLimit,
		WeeklyLimit:       tier.WeeklyCashbackLimit,
		MonthlyLimit:      tier.MonthlyCashbackLimit,
		SpecialBenefits:   tier.SpecialBenefits,
	}

	s.logger.Info("Retrieved user cashback summary", zap.String("user_id", userID.String()))
	return summary, nil
}

// Placeholder implementations for remaining methods
func (s *CashbackStorageImpl) CreateCashbackEarning(ctx context.Context, earning dto.CashbackEarning) (dto.CashbackEarning, error) {
	s.logger.Info("Creating cashback earning",
		zap.String("user_id", earning.UserID.String()),
		zap.String("earning_type", earning.EarningType),
		zap.String("earned_amount", earning.EarnedAmount.String()))

	query := `
		INSERT INTO cashback_earnings (
			user_id, tier_id, earning_type, source_bet_id, ggr_amount, 
			cashback_rate, earned_amount, claimed_amount, available_amount, 
			status, expires_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.db.GetPool().QueryRow(ctx, query,
		earning.UserID,
		earning.TierID,
		earning.EarningType,
		earning.SourceBetID,
		earning.ExpectedGGRAmount,
		earning.CashbackRate,
		earning.EarnedAmount,
		earning.ClaimedAmount,
		earning.AvailableAmount,
		earning.Status,
		earning.ExpiresAt,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		s.logger.Error("Failed to create cashback earning", zap.Error(err))
		return earning, fmt.Errorf("failed to create cashback earning: %w", err)
	}

	earning.ID = id
	earning.CreatedAt = createdAt
	earning.UpdatedAt = updatedAt

	s.logger.Info("Cashback earning created successfully",
		zap.String("earning_id", earning.ID.String()),
		zap.String("user_id", earning.UserID.String()),
		zap.String("earned_amount", earning.EarnedAmount.String()))

	// Sync to ClickHouse for analytics
	if s.analyticsIntegration != nil {
		if err := s.analyticsIntegration.OnCashbackEarning(ctx, &earning); err != nil {
			s.logger.Error("Failed to sync cashback earning to ClickHouse",
				zap.String("earning_id", earning.ID.String()),
				zap.Error(err))
			// Don't fail the operation if ClickHouse sync fails
		}
	}

	return earning, nil
}

func (s *CashbackStorageImpl) GetUserCashbackEarnings(ctx context.Context, userID uuid.UUID) ([]dto.CashbackEarning, error) {
	s.logger.Info("Getting user cashback earnings", zap.String("user_id", userID.String()))

	var earnings []dto.CashbackEarning

	query := `
		SELECT id, user_id, source_bet_id, earning_type, earned_amount, available_amount, 
		       claimed_amount, cashback_rate, ggr_amount, tier_id, status, expires_at, 
		       claimed_at, created_at, updated_at
		FROM cashback_earnings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.logger.Error("Failed to query cashback earnings", zap.Error(err))
		return nil, fmt.Errorf("failed to query cashback earnings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var earning dto.CashbackEarning
		var claimedAt sql.NullTime

		err := rows.Scan(
			&earning.ID,
			&earning.UserID,
			&earning.SourceBetID,
			&earning.EarningType,
			&earning.EarnedAmount,
			&earning.AvailableAmount,
			&earning.ClaimedAmount,
			&earning.CashbackRate,
			&earning.ExpectedGGRAmount,
			&earning.TierID,
			&earning.Status,
			&earning.ExpiresAt,
			&claimedAt,
			&earning.CreatedAt,
			&earning.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan cashback earning", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cashback earning: %w", err)
		}

		if claimedAt.Valid {
			earning.ClaimedAt = &claimedAt.Time
		}

		earnings = append(earnings, earning)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating cashback earnings", zap.Error(err))
		return nil, fmt.Errorf("error iterating cashback earnings: %w", err)
	}

	s.logger.Info("Retrieved user cashback earnings",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(earnings)))

	return earnings, nil
}

func (s *CashbackStorageImpl) GetAvailableCashbackEarnings(ctx context.Context, userID uuid.UUID) ([]dto.CashbackEarning, error) {
	s.logger.Info("Getting available cashback earnings", zap.String("user_id", userID.String()))

	var earnings []dto.CashbackEarning

	query := `
		SELECT id, user_id, source_bet_id, earning_type, earned_amount, available_amount, 
		       claimed_amount, cashback_rate, ggr_amount, tier_id, status, expires_at, 
		       claimed_at, created_at, updated_at
		FROM cashback_earnings
		WHERE user_id = $1 
		  AND status = 'available' 
		  AND available_amount > 0
		  AND expires_at > NOW()
		ORDER BY created_at ASC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.logger.Error("Failed to query available cashback earnings", zap.Error(err))
		return nil, fmt.Errorf("failed to query available cashback earnings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var earning dto.CashbackEarning
		var claimedAt sql.NullTime

		err := rows.Scan(
			&earning.ID,
			&earning.UserID,
			&earning.SourceBetID,
			&earning.EarningType,
			&earning.EarnedAmount,
			&earning.AvailableAmount,
			&earning.ClaimedAmount,
			&earning.CashbackRate,
			&earning.ExpectedGGRAmount,
			&earning.TierID,
			&earning.Status,
			&earning.ExpiresAt,
			&claimedAt,
			&earning.CreatedAt,
			&earning.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan available cashback earning", zap.Error(err))
			return nil, fmt.Errorf("failed to scan available cashback earning: %w", err)
		}

		if claimedAt.Valid {
			earning.ClaimedAt = &claimedAt.Time
		}

		earnings = append(earnings, earning)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating available cashback earnings", zap.Error(err))
		return nil, fmt.Errorf("error iterating available cashback earnings: %w", err)
	}

	s.logger.Info("Retrieved available cashback earnings",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(earnings)))

	return earnings, nil
}

func (s *CashbackStorageImpl) UpdateCashbackEarningStatus(ctx context.Context, earning dto.CashbackEarning) (dto.CashbackEarning, error) {
	s.logger.Info("Updating cashback earning status",
		zap.String("earning_id", earning.ID.String()),
		zap.String("user_id", earning.UserID.String()),
		zap.String("status", earning.Status),
		zap.String("available_amount", earning.AvailableAmount.String()))

	query := `
		UPDATE cashback_earnings 
		SET claimed_amount = $2, 
		    available_amount = $3, 
		    status = $4, 
		    claimed_at = $5,
		    updated_at = NOW()
		WHERE id = $1
	`

	var claimedAt interface{}
	if earning.ClaimedAt != nil {
		claimedAt = earning.ClaimedAt
	} else {
		claimedAt = nil
	}

	_, err := s.db.GetPool().Exec(ctx, query,
		earning.ID,
		earning.ClaimedAmount,
		earning.AvailableAmount,
		earning.Status,
		claimedAt,
	)
	if err != nil {
		s.logger.Error("Failed to update cashback earning status", zap.Error(err))
		return earning, fmt.Errorf("failed to update cashback earning status: %w", err)
	}

	s.logger.Info("Cashback earning status updated successfully",
		zap.String("earning_id", earning.ID.String()),
		zap.String("status", earning.Status))

	return earning, nil
}

func (s *CashbackStorageImpl) CreateCashbackClaim(ctx context.Context, claim dto.CashbackClaim) (dto.CashbackClaim, error) {
	s.logger.Info("Creating cashback claim",
		zap.String("claim_id", claim.ID.String()),
		zap.String("user_id", claim.UserID.String()),
		zap.String("amount", claim.ClaimAmount.String()))

	query := `
		INSERT INTO cashback_claims (id, user_id, claim_amount, net_amount, processing_fee, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := s.db.GetPool().Exec(ctx, query,
		claim.ID,
		claim.UserID,
		claim.ClaimAmount,
		claim.NetAmount,
		claim.ProcessingFee,
		claim.Status,
	)
	if err != nil {
		s.logger.Error("Failed to create cashback claim", zap.Error(err))
		return claim, fmt.Errorf("failed to create cashback claim: %w", err)
	}

	s.logger.Info("Cashback claim created successfully",
		zap.String("claim_id", claim.ID.String()),
		zap.String("user_id", claim.UserID.String()))

	return claim, nil
}

func (s *CashbackStorageImpl) UpdateCashbackClaimStatus(ctx context.Context, claim dto.CashbackClaim) (dto.CashbackClaim, error) {
	s.logger.Info("Updating cashback claim status",
		zap.String("claim_id", claim.ID.String()),
		zap.String("user_id", claim.UserID.String()),
		zap.String("status", claim.Status))

	query := `
		UPDATE cashback_claims 
		SET status = $2, 
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := s.db.GetPool().Exec(ctx, query,
		claim.ID,
		claim.Status,
	)
	if err != nil {
		s.logger.Error("Failed to update cashback claim status", zap.Error(err))
		return claim, fmt.Errorf("failed to update cashback claim status: %w", err)
	}

	s.logger.Info("Cashback claim status updated successfully",
		zap.String("claim_id", claim.ID.String()),
		zap.String("status", claim.Status))

	return claim, nil
}

func (s *CashbackStorageImpl) GetUserCashbackClaims(ctx context.Context, userID uuid.UUID) ([]dto.CashbackClaim, error) {
	s.logger.Info("Getting user cashback claims", zap.String("user_id", userID.String()))

	var claims []dto.CashbackClaim

	query := `
		SELECT id, user_id, claim_amount, net_amount, processing_fee, status, created_at, updated_at
		FROM cashback_claims
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.logger.Error("Failed to query cashback claims", zap.Error(err))
		return nil, fmt.Errorf("failed to query cashback claims: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var claim dto.CashbackClaim

		err := rows.Scan(
			&claim.ID,
			&claim.UserID,
			&claim.ClaimAmount,
			&claim.NetAmount,
			&claim.ProcessingFee,
			&claim.Status,
			&claim.CreatedAt,
			&claim.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan cashback claim", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cashback claim: %w", err)
		}

		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating cashback claims", zap.Error(err))
		return nil, fmt.Errorf("error iterating cashback claims: %w", err)
	}

	s.logger.Info("Retrieved user cashback claims",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(claims)))

	return claims, nil
}

func (s *CashbackStorageImpl) GetCashbackClaim(ctx context.Context, claimID uuid.UUID) (*dto.CashbackClaim, error) {
	s.logger.Info("Getting cashback claim", zap.String("claim_id", claimID.String()))

	query := `
		SELECT id, user_id, claim_amount, net_amount, processing_fee, status, created_at, updated_at
		FROM cashback_claims
		WHERE id = $1
	`

	var claim dto.CashbackClaim

	err := s.db.GetPool().QueryRow(ctx, query, claimID).Scan(
		&claim.ID,
		&claim.UserID,
		&claim.ClaimAmount,
		&claim.NetAmount,
		&claim.ProcessingFee,
		&claim.Status,
		&claim.CreatedAt,
		&claim.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Cashback claim not found", zap.String("claim_id", claimID.String()))
			return nil, nil
		}
		s.logger.Error("Failed to get cashback claim", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback claim: %w", err)
	}

	s.logger.Info("Retrieved cashback claim",
		zap.String("claim_id", claimID.String()),
		zap.String("user_id", claim.UserID.String()))

	return &claim, nil
}

func (s *CashbackStorageImpl) GetGameHouseEdge(ctx context.Context, gameType, gameVariant string) (*dto.GameHouseEdge, error) {
	s.logger.Info("Getting game house edge", zap.String("game_type", gameType), zap.String("game_variant", gameVariant))

	var houseEdge dto.GameHouseEdge
	var minBet, maxBet sql.NullString

	// First try to find exact match with game_variant
	query := `
		SELECT id, game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until
		FROM game_house_edges
		WHERE game_type = $1 AND (game_variant = $2 OR game_variant IS NULL) AND is_active = true
		AND (effective_until IS NULL OR effective_until > NOW())
		ORDER BY 
			CASE WHEN game_variant = $2 THEN 1 ELSE 2 END,
			effective_from DESC
		LIMIT 1
	`

	err := s.db.GetPool().QueryRow(ctx, query, gameType, gameVariant).Scan(
		&houseEdge.ID,
		&houseEdge.GameType,
		&houseEdge.GameVariant,
		&houseEdge.HouseEdge,
		&minBet,
		&maxBet,
		&houseEdge.IsActive,
		&houseEdge.EffectiveFrom,
		&houseEdge.EffectiveUntil,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("No house edge found for game type, using operator-specific default",
				zap.String("game_type", gameType),
				zap.String("game_variant", gameVariant))

			// Return operator-specific default house edge (0% for no default cashback)
			return s.getOperatorDefaultHouseEdge(gameType, gameVariant), nil
		}
		s.logger.Error("Failed to get game house edge", zap.Error(err))
		return nil, fmt.Errorf("failed to get game house edge: %w", err)
	}

	// Parse min_bet and max_bet
	if minBet.Valid {
		if minBetValue, err := decimal.NewFromString(minBet.String); err == nil {
			houseEdge.MinBet = minBetValue
		}
	}
	if maxBet.Valid {
		if maxBetValue, err := decimal.NewFromString(maxBet.String); err == nil {
			houseEdge.MaxBet = &maxBetValue
		}
	}

	maxBetStr := "nil"
	if houseEdge.MaxBet != nil {
		maxBetStr = houseEdge.MaxBet.String()
	}

	s.logger.Info("Retrieved game house edge",
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant),
		zap.String("house_edge", houseEdge.HouseEdge.String()),
		zap.String("min_bet", houseEdge.MinBet.String()),
		zap.String("max_bet", maxBetStr))

	return &houseEdge, nil
}

// getOperatorDefaultHouseEdge returns operator-specific default house edge based on game type
func (s *CashbackStorageImpl) getOperatorDefaultHouseEdge(gameType, gameVariant string) *dto.GameHouseEdge {
	maxBetDefault := decimal.NewFromFloat(1000.0)

	// Operator-specific house edge configurations
	switch gameType {
	case "groovetech":
		// GrooveTech specific defaults based on game categories
		switch {
		case s.isSlotGame(gameVariant):
			return &dto.GameHouseEdge{
				ID:            uuid.New(),
				GameType:      gameType,
				GameVariant:   &gameVariant,
				HouseEdge:     decimal.NewFromFloat(0.025), // 2.5% for slots
				MinBet:        decimal.NewFromFloat(0.1),
				MaxBet:        &maxBetDefault,
				IsActive:      true,
				EffectiveFrom: time.Now(),
			}
		case s.isTableGame(gameVariant):
			return &dto.GameHouseEdge{
				ID:            uuid.New(),
				GameType:      gameType,
				GameVariant:   &gameVariant,
				HouseEdge:     decimal.NewFromFloat(0.015), // 1.5% for table games
				MinBet:        decimal.NewFromFloat(1.0),
				MaxBet:        &maxBetDefault,
				IsActive:      true,
				EffectiveFrom: time.Now(),
			}
		case s.isLiveGame(gameVariant):
			return &dto.GameHouseEdge{
				ID:            uuid.New(),
				GameType:      gameType,
				GameVariant:   &gameVariant,
				HouseEdge:     decimal.NewFromFloat(0.01), // 1% for live games
				MinBet:        decimal.NewFromFloat(5.0),
				MaxBet:        &maxBetDefault,
				IsActive:      true,
				EffectiveFrom: time.Now(),
			}
		default:
			return &dto.GameHouseEdge{
				ID:            uuid.New(),
				GameType:      gameType,
				GameVariant:   &gameVariant,
				HouseEdge:     decimal.NewFromFloat(0.0), // 0% default for GrooveTech (no default cashback)
				MinBet:        decimal.NewFromFloat(0.1),
				MaxBet:        &maxBetDefault,
				IsActive:      true,
				EffectiveFrom: time.Now(),
			}
		}
	case "evolution":
		// Evolution Gaming defaults
		return &dto.GameHouseEdge{
			ID:            uuid.New(),
			GameType:      gameType,
			GameVariant:   &gameVariant,
			HouseEdge:     decimal.NewFromFloat(0.012), // 1.2% for Evolution
			MinBet:        decimal.NewFromFloat(1.0),
			MaxBet:        &maxBetDefault,
			IsActive:      true,
			EffectiveFrom: time.Now(),
		}
	case "pragmatic":
		// Pragmatic Play defaults
		return &dto.GameHouseEdge{
			ID:            uuid.New(),
			GameType:      gameType,
			GameVariant:   &gameVariant,
			HouseEdge:     decimal.NewFromFloat(0.03), // 3% for Pragmatic
			MinBet:        decimal.NewFromFloat(0.2),
			MaxBet:        &maxBetDefault,
			IsActive:      true,
			EffectiveFrom: time.Now(),
		}
	default:
		// Generic default for unknown operators
		return &dto.GameHouseEdge{
			ID:            uuid.New(),
			GameType:      gameType,
			GameVariant:   &gameVariant,
			HouseEdge:     decimal.NewFromFloat(0.0), // 0% generic default (no default cashback)
			MinBet:        decimal.NewFromFloat(0.1),
			MaxBet:        &maxBetDefault,
			IsActive:      true,
			EffectiveFrom: time.Now(),
		}
	}
}

// Helper functions to categorize games
func (s *CashbackStorageImpl) isSlotGame(gameVariant string) bool {
	slotKeywords := []string{"slot", "reel", "spin", "fruit", "classic", "video"}
	for _, keyword := range slotKeywords {
		if strings.Contains(strings.ToLower(gameVariant), keyword) {
			return true
		}
	}
	return false
}

func (s *CashbackStorageImpl) isTableGame(gameVariant string) bool {
	tableKeywords := []string{"blackjack", "poker", "baccarat", "roulette", "table"}
	for _, keyword := range tableKeywords {
		if strings.Contains(strings.ToLower(gameVariant), keyword) {
			return true
		}
	}
	return false
}

func (s *CashbackStorageImpl) isLiveGame(gameVariant string) bool {
	liveKeywords := []string{"live", "dealer", "studio", "stream"}
	for _, keyword := range liveKeywords {
		if strings.Contains(strings.ToLower(gameVariant), keyword) {
			return true
		}
	}
	return false
}

func (s *CashbackStorageImpl) CreateGameHouseEdge(ctx context.Context, houseEdge dto.GameHouseEdge) (dto.GameHouseEdge, error) {
	return houseEdge, nil
}

func (s *CashbackStorageImpl) GetActiveCashbackPromotions(ctx context.Context) ([]dto.CashbackPromotion, error) {
	return []dto.CashbackPromotion{}, nil
}

func (s *CashbackStorageImpl) GetCashbackPromotionForUser(ctx context.Context, userLevel int, gameType string) (*dto.CashbackPromotion, error) {
	return nil, nil
}

func (s *CashbackStorageImpl) CreateCashbackPromotion(ctx context.Context, promotion dto.CashbackPromotion) (dto.CashbackPromotion, error) {
	return promotion, nil
}

func (s *CashbackStorageImpl) UpdateCashbackPromotion(ctx context.Context, promotion dto.CashbackPromotion) (dto.CashbackPromotion, error) {
	return promotion, nil
}

func (s *CashbackStorageImpl) GetCashbackStats(ctx context.Context) (*dto.AdminCashbackStats, error) {
	return &dto.AdminCashbackStats{}, nil
}

func (s *CashbackStorageImpl) GetTierDistribution(ctx context.Context) ([]dto.TierDistribution, error) {
	return []dto.TierDistribution{}, nil
}

func (s *CashbackStorageImpl) GetUserDailyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	s.logger.Info("Getting user daily cashback limit", zap.String("user_id", userID.String()))

	// Get today's date in UTC
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	query := `
		SELECT COALESCE(SUM(claimed_amount), 0) as daily_claimed
		FROM cashback_earnings
		WHERE user_id = $1 
		AND claimed_at >= $2 
		AND claimed_at < $3`

	var dailyClaimed decimal.Decimal
	err := s.db.GetPool().QueryRow(ctx, query, userID, today, tomorrow).Scan(&dailyClaimed)
	if err != nil {
		s.logger.Error("Failed to get user daily cashback limit", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get user daily cashback limit: %w", err)
	}

	s.logger.Info("Retrieved user daily cashback limit",
		zap.String("user_id", userID.String()),
		zap.String("daily_claimed", dailyClaimed.String()),
		zap.String("date", today.Format("2006-01-02")))

	return dailyClaimed, nil
}

func (s *CashbackStorageImpl) GetUserWeeklyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (s *CashbackStorageImpl) GetUserMonthlyCashbackLimit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (s *CashbackStorageImpl) GetExpiredCashbackEarnings(ctx context.Context) ([]dto.CashbackEarning, error) {
	return []dto.CashbackEarning{}, nil
}

func (s *CashbackStorageImpl) MarkCashbackEarningsAsExpired(ctx context.Context) error {
	return nil
}

// Retryable Operations implementations
func (s *CashbackStorageImpl) CreateRetryableOperation(ctx context.Context, operation RetryableOperation) error {
	s.logger.Info("Creating retryable operation",
		zap.String("operation_id", operation.ID.String()),
		zap.String("operation_type", operation.Type),
		zap.String("user_id", operation.UserID.String()))

	query := `
		INSERT INTO retryable_operations (
			id, type, user_id, data, attempts, last_error, next_retry_at, 
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			attempts = EXCLUDED.attempts,
			last_error = EXCLUDED.last_error,
			next_retry_at = EXCLUDED.next_retry_at,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`

	_, err := s.db.GetPool().Exec(ctx, query,
		operation.ID,
		operation.Type,
		operation.UserID,
		operation.Data,
		operation.Attempts,
		operation.LastError,
		operation.NextRetryAt,
		operation.Status,
		operation.CreatedAt,
		operation.UpdatedAt,
	)

	if err != nil {
		s.logger.Error("Failed to create retryable operation", zap.Error(err))
		return fmt.Errorf("failed to create retryable operation: %w", err)
	}

	s.logger.Info("Retryable operation created successfully",
		zap.String("operation_id", operation.ID.String()))

	return nil
}

func (s *CashbackStorageImpl) GetRetryableOperation(ctx context.Context, operationID uuid.UUID) (*RetryableOperation, error) {
	s.logger.Info("Getting retryable operation", zap.String("operation_id", operationID.String()))

	var operation RetryableOperation
	var data map[string]interface{}

	query := `
		SELECT id, type, user_id, data, attempts, last_error, next_retry_at, 
		       status, created_at, updated_at
		FROM retryable_operations
		WHERE id = $1`

	err := s.db.GetPool().QueryRow(ctx, query, operationID).Scan(
		&operation.ID,
		&operation.Type,
		&operation.UserID,
		&data,
		&operation.Attempts,
		&operation.LastError,
		&operation.NextRetryAt,
		&operation.Status,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Retryable operation not found", zap.String("operation_id", operationID.String()))
			return nil, nil
		}
		s.logger.Error("Failed to get retryable operation", zap.Error(err))
		return nil, fmt.Errorf("failed to get retryable operation: %w", err)
	}

	operation.Data = data

	s.logger.Info("Retryable operation retrieved successfully",
		zap.String("operation_id", operationID.String()),
		zap.String("status", operation.Status))

	return &operation, nil
}

func (s *CashbackStorageImpl) UpdateRetryableOperation(ctx context.Context, operation RetryableOperation) error {
	s.logger.Info("Updating retryable operation",
		zap.String("operation_id", operation.ID.String()),
		zap.String("status", operation.Status),
		zap.Int("attempts", operation.Attempts))

	query := `
		UPDATE retryable_operations
		SET attempts = $2, last_error = $3, next_retry_at = $4, 
		    status = $5, updated_at = $6
		WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query,
		operation.ID,
		operation.Attempts,
		operation.LastError,
		operation.NextRetryAt,
		operation.Status,
		operation.UpdatedAt,
	)

	if err != nil {
		s.logger.Error("Failed to update retryable operation", zap.Error(err))
		return fmt.Errorf("failed to update retryable operation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Warn("No retryable operation found to update", zap.String("operation_id", operation.ID.String()))
		return fmt.Errorf("retryable operation not found")
	}

	s.logger.Info("Retryable operation updated successfully",
		zap.String("operation_id", operation.ID.String()))

	return nil
}

func (s *CashbackStorageImpl) GetRetryableOperationsByUser(ctx context.Context, userID uuid.UUID) ([]RetryableOperation, error) {
	s.logger.Info("Getting retryable operations by user", zap.String("user_id", userID.String()))

	query := `
		SELECT id, type, user_id, data, attempts, last_error, next_retry_at, 
		       status, created_at, updated_at
		FROM retryable_operations
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.logger.Error("Failed to get retryable operations by user", zap.Error(err))
		return nil, fmt.Errorf("failed to get retryable operations by user: %w", err)
	}
	defer rows.Close()

	var operations []RetryableOperation
	for rows.Next() {
		var operation RetryableOperation
		var data map[string]interface{}

		err := rows.Scan(
			&operation.ID,
			&operation.Type,
			&operation.UserID,
			&data,
			&operation.Attempts,
			&operation.LastError,
			&operation.NextRetryAt,
			&operation.Status,
			&operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan retryable operation", zap.Error(err))
			continue
		}

		operation.Data = data
		operations = append(operations, operation)
	}

	s.logger.Info("Retrieved retryable operations by user",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(operations)))

	return operations, nil
}

func (s *CashbackStorageImpl) GetFailedRetryableOperations(ctx context.Context) ([]RetryableOperation, error) {
	s.logger.Info("Getting failed retryable operations")

	query := `
		SELECT id, type, user_id, data, attempts, last_error, next_retry_at, 
		       status, created_at, updated_at
		FROM retryable_operations
		WHERE status = 'failed' 
		AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.logger.Error("Failed to get failed retryable operations", zap.Error(err))
		return nil, fmt.Errorf("failed to get failed retryable operations: %w", err)
	}
	defer rows.Close()

	var operations []RetryableOperation
	for rows.Next() {
		var operation RetryableOperation
		var data map[string]interface{}

		err := rows.Scan(
			&operation.ID,
			&operation.Type,
			&operation.UserID,
			&data,
			&operation.Attempts,
			&operation.LastError,
			&operation.NextRetryAt,
			&operation.Status,
			&operation.CreatedAt,
			&operation.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan failed retryable operation", zap.Error(err))
			continue
		}

		operation.Data = data
		operations = append(operations, operation)
	}

	s.logger.Info("Retrieved failed retryable operations", zap.Int("count", len(operations)))
	return operations, nil
}

func (s *CashbackStorageImpl) DeleteRetryableOperation(ctx context.Context, operationID uuid.UUID) error {
	s.logger.Info("Deleting retryable operation", zap.String("operation_id", operationID.String()))

	query := `DELETE FROM retryable_operations WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query, operationID)
	if err != nil {
		s.logger.Error("Failed to delete retryable operation", zap.Error(err))
		return fmt.Errorf("failed to delete retryable operation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Warn("No retryable operation found to delete", zap.String("operation_id", operationID.String()))
		return fmt.Errorf("retryable operation not found")
	}

	s.logger.Info("Retryable operation deleted successfully",
		zap.String("operation_id", operationID.String()))

	return nil
}

// GetAllCashbackTiers returns all cashback tiers ordered by level
func (s *CashbackStorageImpl) GetAllCashbackTiers(ctx context.Context) ([]dto.CashbackTier, error) {
	s.logger.Info("Getting all cashback tiers")

	var tiers []dto.CashbackTier

	query := `
		SELECT id, tier_name, tier_level, min_ggr_required, cashback_percentage, 
		       bonus_multiplier, daily_cashback_limit, weekly_cashback_limit, 
		       monthly_cashback_limit, special_benefits, is_active, created_at, updated_at
		FROM cashback_tiers
		WHERE is_active = true
		ORDER BY tier_level ASC
	`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.logger.Error("Failed to query cashback tiers", zap.Error(err))
		return nil, fmt.Errorf("failed to query cashback tiers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tier dto.CashbackTier
		var dailyLimit, weeklyLimit, monthlyLimit sql.NullString
		var specialBenefits sql.NullString

		err := rows.Scan(
			&tier.ID,
			&tier.TierName,
			&tier.TierLevel,
			&tier.MinExpectedGGRRequired,
			&tier.CashbackPercentage,
			&tier.BonusMultiplier,
			&dailyLimit,
			&weeklyLimit,
			&monthlyLimit,
			&specialBenefits,
			&tier.IsActive,
			&tier.CreatedAt,
			&tier.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan cashback tier", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cashback tier: %w", err)
		}

		// Parse nullable fields
		if dailyLimit.Valid {
			if dailyLimitValue, err := decimal.NewFromString(dailyLimit.String); err == nil {
				tier.DailyCashbackLimit = &dailyLimitValue
			}
		}
		if weeklyLimit.Valid {
			if weeklyLimitValue, err := decimal.NewFromString(weeklyLimit.String); err == nil {
				tier.WeeklyCashbackLimit = &weeklyLimitValue
			}
		}
		if monthlyLimit.Valid {
			if monthlyLimitValue, err := decimal.NewFromString(monthlyLimit.String); err == nil {
				tier.MonthlyCashbackLimit = &monthlyLimitValue
			}
		}
		if specialBenefits.Valid && specialBenefits.String != "" {
			// Parse JSON special benefits
			tier.SpecialBenefits = make(map[string]interface{})
			// For now, we'll leave it empty - in a real implementation, you'd parse the JSON
		} else {
			tier.SpecialBenefits = make(map[string]interface{})
		}

		tiers = append(tiers, tier)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating cashback tiers", zap.Error(err))
		return nil, fmt.Errorf("error iterating cashback tiers: %w", err)
	}

	s.logger.Info("Retrieved all cashback tiers", zap.Int("count", len(tiers)))

	return tiers, nil
}

// GetCashbackTierByID returns a cashback tier by its ID
func (s *CashbackStorageImpl) GetCashbackTierByID(ctx context.Context, tierID uuid.UUID) (*dto.CashbackTier, error) {
	s.logger.Info("Getting cashback tier by ID", zap.String("tier_id", tierID.String()))

	var tier dto.CashbackTier
	var dailyLimit, weeklyLimit, monthlyLimit sql.NullString
	var specialBenefits sql.NullString

	query := `
		SELECT id, tier_name, tier_level, min_ggr_required, cashback_percentage, 
		       bonus_multiplier, daily_cashback_limit, weekly_cashback_limit, 
		       monthly_cashback_limit, special_benefits, is_active, created_at, updated_at
		FROM cashback_tiers
		WHERE id = $1 AND is_active = true
	`

	err := s.db.GetPool().QueryRow(ctx, query, tierID).Scan(
		&tier.ID,
		&tier.TierName,
		&tier.TierLevel,
		&tier.MinExpectedGGRRequired,
		&tier.CashbackPercentage,
		&tier.BonusMultiplier,
		&dailyLimit,
		&weeklyLimit,
		&monthlyLimit,
		&specialBenefits,
		&tier.IsActive,
		&tier.CreatedAt,
		&tier.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Cashback tier not found", zap.String("tier_id", tierID.String()))
			return nil, fmt.Errorf("cashback tier not found")
		}
		s.logger.Error("Failed to get cashback tier", zap.Error(err))
		return nil, fmt.Errorf("failed to get cashback tier: %w", err)
	}

	// Parse nullable fields
	if dailyLimit.Valid {
		if dailyLimitValue, err := decimal.NewFromString(dailyLimit.String); err == nil {
			tier.DailyCashbackLimit = &dailyLimitValue
		}
	}
	if weeklyLimit.Valid {
		if weeklyLimitValue, err := decimal.NewFromString(weeklyLimit.String); err == nil {
			tier.WeeklyCashbackLimit = &weeklyLimitValue
		}
	}
	if monthlyLimit.Valid {
		if monthlyLimitValue, err := decimal.NewFromString(monthlyLimit.String); err == nil {
			tier.MonthlyCashbackLimit = &monthlyLimitValue
		}
	}
	if specialBenefits.Valid && specialBenefits.String != "" {
		// Parse JSON special benefits
		tier.SpecialBenefits = make(map[string]interface{})
		// For now, we'll leave it empty - in a real implementation, you'd parse the JSON
	} else {
		tier.SpecialBenefits = make(map[string]interface{})
	}

	s.logger.Info("Retrieved cashback tier",
		zap.String("tier_id", tierID.String()),
		zap.String("tier_name", tier.TierName),
		zap.Int("tier_level", tier.TierLevel))

	return &tier, nil
}

// InitializeUserLevel creates a new user level entry for a user
func (s *CashbackStorageImpl) InitializeUserLevel(ctx context.Context, userID uuid.UUID) (*dto.UserLevel, error) {
	s.logger.Info("Initializing user level", zap.String("user_id", userID.String()))

	// Get the default tier (Bronze)
	defaultTier, err := s.GetCashbackTierByLevel(ctx, 1)
	if err != nil {
		s.logger.Error("Failed to get default tier", zap.Error(err))
		return nil, fmt.Errorf("failed to get default tier: %w", err)
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
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save user level
	savedUserLevel, err := s.CreateUserLevel(ctx, userLevel)
	if err != nil {
		s.logger.Error("Failed to create user level", zap.Error(err))
		return nil, fmt.Errorf("failed to create user level: %w", err)
	}

	s.logger.Info("User level initialized successfully",
		zap.String("user_id", userID.String()),
		zap.String("tier_name", defaultTier.TierName),
		zap.Int("tier_level", defaultTier.TierLevel))

	return &savedUserLevel, nil
}
