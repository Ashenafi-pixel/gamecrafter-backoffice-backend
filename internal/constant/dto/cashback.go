package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GlobalRakebackOverride represents the global rakeback override configuration
type GlobalRakebackOverride struct {
	ID                uuid.UUID       `json:"id"`
	IsActive          bool            `json:"is_active"`
	RakebackPercentage decimal.Decimal `json:"rakeback_percentage"`
	StartTime         *time.Time      `json:"start_time,omitempty"`
	EndTime           *time.Time      `json:"end_time,omitempty"`
	CreatedBy         *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedBy         *uuid.UUID      `json:"updated_by,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// CreateOrUpdateRakebackOverrideReq represents the request to create or update a global rakeback override
type CreateOrUpdateRakebackOverrideReq struct {
	IsActive          bool            `json:"is_active"`
	RakebackPercentage decimal.Decimal `json:"rakeback_percentage" binding:"required"`
	StartTime         *time.Time      `json:"start_time,omitempty"`
	EndTime           *time.Time      `json:"end_time,omitempty"`
}

// UserLevel represents a user's current level and progress
type UserLevel struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserID           uuid.UUID       `json:"user_id" db:"user_id"`
	CurrentLevel     int             `json:"current_level" db:"current_level"`
	TotalExpectedGGR decimal.Decimal `json:"total_expected_ggr" db:"total_expected_ggr"`
	TotalBets        decimal.Decimal `json:"total_bets" db:"total_bets"`
	TotalWins        decimal.Decimal `json:"total_wins" db:"total_wins"`
	LevelProgress    decimal.Decimal `json:"level_progress" db:"level_progress"`
	CurrentTierID    uuid.UUID       `json:"current_tier_id" db:"current_tier_id"`
	LastLevelUp      *time.Time      `json:"last_level_up" db:"last_level_up"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// CashbackTier represents a cashback tier configuration
type CashbackTier struct {
	ID                     uuid.UUID              `json:"id" db:"id"`
	TierName               string                 `json:"tier_name" db:"tier_name"`
	TierLevel              int                    `json:"tier_level" db:"tier_level"`
	MinExpectedGGRRequired decimal.Decimal        `json:"min_expected_ggr_required" db:"min_expected_ggr_required"`
	CashbackPercentage     decimal.Decimal        `json:"cashback_percentage" db:"cashback_percentage"`
	BonusMultiplier        decimal.Decimal        `json:"bonus_multiplier" db:"bonus_multiplier"`
	DailyCashbackLimit     *decimal.Decimal       `json:"daily_cashback_limit" db:"daily_cashback_limit"`
	WeeklyCashbackLimit    *decimal.Decimal       `json:"weekly_cashback_limit" db:"weekly_cashback_limit"`
	MonthlyCashbackLimit   *decimal.Decimal       `json:"monthly_cashback_limit" db:"monthly_cashback_limit"`
	SpecialBenefits        map[string]interface{} `json:"special_benefits" db:"special_benefits"`
	IsActive               bool                   `json:"is_active" db:"is_active"`
	PlayerCount            int                    `json:"player_count" db:"player_count"`
	CreatedAt              time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
}

// CashbackEarning represents a cashback earning record
type CashbackEarning struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	UserID            uuid.UUID       `json:"user_id" db:"user_id"`
	TierID            uuid.UUID       `json:"tier_id" db:"tier_id"`
	EarningType       string          `json:"earning_type" db:"earning_type"`
	SourceBetID       *uuid.UUID      `json:"source_bet_id" db:"source_bet_id"`
	ExpectedGGRAmount decimal.Decimal `json:"expected_ggr_amount" db:"ggr_amount"`
	CashbackRate      decimal.Decimal `json:"cashback_rate" db:"cashback_rate"`
	EarnedAmount      decimal.Decimal `json:"earned_amount" db:"earned_amount"`
	ClaimedAmount     decimal.Decimal `json:"claimed_amount" db:"claimed_amount"`
	AvailableAmount   decimal.Decimal `json:"available_amount" db:"available_amount"`
	Status            string          `json:"status" db:"status"`
	ExpiresAt         time.Time       `json:"expires_at" db:"expires_at"`
	ClaimedAt         *time.Time      `json:"claimed_at" db:"claimed_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// CashbackClaim represents a cashback claim request
type CashbackClaim struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	UserID          uuid.UUID              `json:"user_id" db:"user_id"`
	ClaimAmount     decimal.Decimal        `json:"claim_amount" db:"claim_amount"`
	CurrencyCode    string                 `json:"currency_code" db:"currency_code"`
	Status          string                 `json:"status" db:"status"`
	TransactionID   *uuid.UUID             `json:"transaction_id" db:"transaction_id"`
	ProcessingFee   decimal.Decimal        `json:"processing_fee" db:"processing_fee"`
	NetAmount       decimal.Decimal        `json:"net_amount" db:"net_amount"`
	ClaimedEarnings map[string]interface{} `json:"claimed_earnings" db:"claimed_earnings"`
	AdminNotes      *string                `json:"admin_notes" db:"admin_notes"`
	ProcessedAt     *time.Time             `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// GameHouseEdge represents house edge configuration for games
type GameHouseEdge struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	GameType       string           `json:"game_type" db:"game_type"`
	GameVariant    *string          `json:"game_variant" db:"game_variant"`
	HouseEdge      decimal.Decimal  `json:"house_edge" db:"house_edge"`
	MinBet         decimal.Decimal  `json:"min_bet" db:"min_bet"`
	MaxBet         *decimal.Decimal `json:"max_bet" db:"max_bet"`
	IsActive       bool             `json:"is_active" db:"is_active"`
	EffectiveFrom  time.Time        `json:"effective_from" db:"effective_from"`
	EffectiveUntil *time.Time       `json:"effective_until" db:"effective_until"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

// CashbackPromotion represents special cashback promotions
type CashbackPromotion struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	PromotionName   string           `json:"promotion_name" db:"promotion_name"`
	Description     *string          `json:"description" db:"description"`
	PromotionType   string           `json:"promotion_type" db:"promotion_type"`
	BoostMultiplier decimal.Decimal  `json:"boost_multiplier" db:"boost_multiplier"`
	BonusAmount     decimal.Decimal  `json:"bonus_amount" db:"bonus_amount"`
	MinBetAmount    decimal.Decimal  `json:"min_bet_amount" db:"min_bet_amount"`
	MaxBonusAmount  *decimal.Decimal `json:"max_bonus_amount" db:"max_bonus_amount"`
	TargetTiers     []int            `json:"target_tiers" db:"target_tiers"`
	TargetGames     []string         `json:"target_games" db:"target_games"`
	IsActive        bool             `json:"is_active" db:"is_active"`
	StartsAt        time.Time        `json:"starts_at" db:"starts_at"`
	EndsAt          *time.Time       `json:"ends_at" db:"ends_at"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// UserCashbackSummary represents a user's cashback summary
type UserCashbackSummary struct {
	UserID                   uuid.UUID              `json:"user_id"`
	CurrentTier              CashbackTier           `json:"current_tier"`
	LevelProgress            decimal.Decimal        `json:"level_progress"`
	TotalGGR                 decimal.Decimal        `json:"total_ggr"`
	AvailableCashback        decimal.Decimal        `json:"available_cashback"`
	PendingCashback          decimal.Decimal        `json:"pending_cashback"`
	TotalClaimed             decimal.Decimal        `json:"total_claimed"`
	NextTierGGR              decimal.Decimal        `json:"next_tier_ggr"`
	DailyLimit               *decimal.Decimal       `json:"daily_limit"`
	WeeklyLimit              *decimal.Decimal       `json:"weekly_limit"`
	MonthlyLimit             *decimal.Decimal       `json:"monthly_limit"`
	SpecialBenefits          map[string]interface{} `json:"special_benefits"`
	GlobalOverrideActive     bool                   `json:"global_override_active"`
	EffectiveCashbackPercent decimal.Decimal        `json:"effective_cashback_percent"`
	HappyHourMessage         string                 `json:"happy_hour_message,omitempty"`
}

// EnhancedUserCashbackSummary represents a user's cashback summary with game-specific details
type EnhancedUserCashbackSummary struct {
	UserID            uuid.UUID              `json:"user_id"`
	CurrentTier       CashbackTier           `json:"current_tier"`
	LevelProgress     decimal.Decimal        `json:"level_progress"`
	TotalGGR          decimal.Decimal        `json:"total_ggr"`
	AvailableCashback decimal.Decimal        `json:"available_cashback"`
	PendingCashback   decimal.Decimal        `json:"pending_cashback"`
	TotalClaimed      decimal.Decimal        `json:"total_claimed"`
	NextTierGGR       decimal.Decimal        `json:"next_tier_ggr"`
	DailyLimit        *decimal.Decimal       `json:"daily_limit"`
	WeeklyLimit       *decimal.Decimal       `json:"weekly_limit"`
	MonthlyLimit      *decimal.Decimal       `json:"monthly_limit"`
	SpecialBenefits   map[string]interface{} `json:"special_benefits"`
	// Game-specific information
	LastGameInfo *GameCashbackInfo `json:"last_game_info,omitempty"`
}

// GameCashbackInfo represents game-specific cashback information
type GameCashbackInfo struct {
	GameID           string          `json:"game_id"`
	GameName         string          `json:"game_name"`
	GameType         string          `json:"game_type"`
	GameVariant      string          `json:"game_variant"`
	HouseEdge        decimal.Decimal `json:"house_edge"`
	HouseEdgePercent string          `json:"house_edge_percent"`
	CashbackRate     decimal.Decimal `json:"cashback_rate"`
	CashbackPercent  string          `json:"cashback_percent"`
	ExpectedGGR      decimal.Decimal `json:"expected_ggr"`
	EarnedCashback   decimal.Decimal `json:"earned_cashback"`
	BetAmount        decimal.Decimal `json:"bet_amount"`
	TransactionID    string          `json:"transaction_id"`
	Timestamp        time.Time       `json:"timestamp"`
}

// CashbackClaimRequest represents a request to claim cashback
type CashbackClaimRequest struct {
	Amount decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

// CashbackClaimResponse represents the response after claiming cashback
type CashbackClaimResponse struct {
	ClaimID       uuid.UUID       `json:"claim_id"`
	Amount        decimal.Decimal `json:"amount"`
	NetAmount     decimal.Decimal `json:"net_amount"`
	ProcessingFee decimal.Decimal `json:"processing_fee"`
	Status        string          `json:"status"`
	Message       string          `json:"message"`
}

// BetCashbackCalculation represents the calculation for a bet's cashback
type BetCashbackCalculation struct {
	BetID          uuid.UUID       `json:"bet_id"`
	UserID         uuid.UUID       `json:"user_id"`
	BetAmount      decimal.Decimal `json:"bet_amount"`
	GameType       string          `json:"game_type"`
	HouseEdge      decimal.Decimal `json:"house_edge"`
	ExpectedGGR    decimal.Decimal `json:"expected_ggr"`
	CashbackRate   decimal.Decimal `json:"cashback_rate"`
	EarnedCashback decimal.Decimal `json:"earned_cashback"`
	TierID         uuid.UUID       `json:"tier_id"`
	PromotionBoost decimal.Decimal `json:"promotion_boost"`
}

// AdminCashbackStats represents admin statistics for cashback system
type AdminCashbackStats struct {
	TotalUsersWithCashback int             `json:"total_users_with_cashback"`
	TotalCashbackEarned    decimal.Decimal `json:"total_cashback_earned"`
	TotalCashbackClaimed   decimal.Decimal `json:"total_cashback_claimed"`
	TotalCashbackPending   decimal.Decimal `json:"total_cashback_pending"`
	AverageCashbackRate    decimal.Decimal `json:"average_cashback_rate"`
	TierDistribution       map[string]int  `json:"tier_distribution"`
	DailyCashbackClaims    decimal.Decimal `json:"daily_cashback_claims"`
	WeeklyCashbackClaims   decimal.Decimal `json:"weekly_cashback_claims"`
	MonthlyCashbackClaims  decimal.Decimal `json:"monthly_cashback_claims"`
}

// CashbackTierUpdateRequest represents a request to update cashback tier
type CashbackTierUpdateRequest struct {
	TierName             string                 `json:"tier_name" validate:"required"`
	MinGGRRequired       decimal.Decimal        `json:"min_ggr_required" validate:"required,gte=0"`
	CashbackPercentage   decimal.Decimal        `json:"cashback_percentage" validate:"required,gte=0,lte=100"`
	BonusMultiplier      decimal.Decimal        `json:"bonus_multiplier" validate:"required,gte=1"`
	DailyCashbackLimit   *decimal.Decimal       `json:"daily_cashback_limit"`
	WeeklyCashbackLimit  *decimal.Decimal       `json:"weekly_cashback_limit"`
	MonthlyCashbackLimit *decimal.Decimal       `json:"monthly_cashback_limit"`
	SpecialBenefits      map[string]interface{} `json:"special_benefits"`
	IsActive             bool                   `json:"is_active"`
}

// CashbackPromotionRequest represents a request to create/update promotion
type CashbackPromotionRequest struct {
	PromotionName   string           `json:"promotion_name" validate:"required"`
	Description     string           `json:"description"`
	PromotionType   string           `json:"promotion_type" validate:"required,oneof=boost bonus special"`
	BoostMultiplier decimal.Decimal  `json:"boost_multiplier" validate:"required,gte=1"`
	BonusAmount     decimal.Decimal  `json:"bonus_amount" validate:"gte=0"`
	MinBetAmount    decimal.Decimal  `json:"min_bet_amount" validate:"gte=0"`
	MaxBonusAmount  *decimal.Decimal `json:"max_bonus_amount"`
	TargetTiers     []int            `json:"target_tiers"`
	TargetGames     []string         `json:"target_games"`
	StartsAt        time.Time        `json:"starts_at" validate:"required"`
	EndsAt          *time.Time       `json:"ends_at"`
	IsActive        bool             `json:"is_active"`
}

// TierDistribution represents tier distribution statistics
type TierDistribution struct {
	TierName  string `json:"tier_name"`
	UserCount int    `json:"user_count"`
}

// LevelProgressionInfo represents detailed level progression information
type LevelProgressionInfo struct {
	UserID                 uuid.UUID       `json:"user_id"`
	CurrentLevel           int             `json:"current_level"`
	CurrentTier            CashbackTier    `json:"current_tier"`
	NextTier               *CashbackTier   `json:"next_tier,omitempty"`
	TotalExpectedGGR       decimal.Decimal `json:"total_expected_ggr"`
	ProgressToNext         decimal.Decimal `json:"progress_to_next"`
	ExpectedGGRToNextLevel decimal.Decimal `json:"expected_ggr_to_next_level"`
	LastLevelUp            *time.Time      `json:"last_level_up"`
	LevelProgress          decimal.Decimal `json:"level_progress"`
}

// LevelProgressionResult represents the result of level progression processing
type LevelProgressionResult struct {
	UserID            uuid.UUID `json:"user_id"`
	Success           bool      `json:"success"`
	NewLevel          int       `json:"new_level,omitempty"`
	CurrentLevel      int       `json:"current_level,omitempty"`
	Error             string    `json:"error,omitempty"`
	Message           string    `json:"message,omitempty"`
	TotalExpectedGGR  string    `json:"total_expected_ggr,omitempty"`
	RequiredGGR       string    `json:"required_ggr,omitempty"`
	NextTierName      string    `json:"next_tier_name,omitempty"`
	CurrentTierName   string    `json:"current_tier_name,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// ReorderTiersRequest represents a request to reorder cashback tiers
type ReorderTiersRequest struct {
	TierOrder []uuid.UUID `json:"tier_order" validate:"required,min=1"`
}

// GlobalRakebackOverride represents the global rakeback override configuration (Happy Hour Mode)
type GlobalRakebackOverride struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	IsEnabled          bool            `json:"is_enabled" db:"is_enabled"`
	OverridePercentage decimal.Decimal `json:"override_percentage" db:"override_percentage"`
	EnabledBy          *uuid.UUID      `json:"enabled_by" db:"enabled_by"`
	EnabledAt          *time.Time      `json:"enabled_at" db:"enabled_at"`
	DisabledBy         *uuid.UUID      `json:"disabled_by" db:"disabled_by"`
	DisabledAt         *time.Time      `json:"disabled_at" db:"disabled_at"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// GlobalRakebackOverrideRequest represents a request to update global rakeback override
type GlobalRakebackOverrideRequest struct {
	IsEnabled          bool            `json:"is_enabled" validate:"required"`
	OverridePercentage decimal.Decimal `json:"override_percentage" validate:"required,gte=0,lte=100"`
}

// GlobalRakebackOverrideResponse represents the response for global rakeback override
type GlobalRakebackOverrideResponse struct {
	IsEnabled          bool            `json:"is_enabled"`
	OverridePercentage decimal.Decimal `json:"override_percentage"`
	EnabledBy          *uuid.UUID      `json:"enabled_by,omitempty"`
	EnabledByUsername  *string         `json:"enabled_by_username,omitempty"`
	EnabledAt          *time.Time      `json:"enabled_at,omitempty"`
	DisabledBy         *uuid.UUID      `json:"disabled_by,omitempty"`
	DisabledByUsername *string         `json:"disabled_by_username,omitempty"`
	DisabledAt         *time.Time      `json:"disabled_at,omitempty"`
	Message            string          `json:"message"`
}

// RakebackSchedule represents a scheduled rakeback event
type RakebackSchedule struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Description   *string         `json:"description,omitempty" db:"description"`
	StartTime     time.Time       `json:"start_time" db:"start_time"`
	EndTime       time.Time       `json:"end_time" db:"end_time"`
	Percentage    decimal.Decimal `json:"percentage" db:"percentage"`
	ScopeType     string          `json:"scope_type" db:"scope_type"`
	ScopeValue    *string         `json:"scope_value,omitempty" db:"scope_value"`
	Status        string          `json:"status" db:"status"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	ActivatedAt   *time.Time      `json:"activated_at,omitempty" db:"activated_at"`
	DeactivatedAt *time.Time      `json:"deactivated_at,omitempty" db:"deactivated_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateRakebackScheduleRequest represents a request to create a scheduled rakeback
type CreateRakebackScheduleRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	StartTime   time.Time       `json:"start_time" validate:"required"`
	EndTime     time.Time       `json:"end_time" validate:"required"`
	Percentage  decimal.Decimal `json:"percentage" validate:"required,gte=0,lte=100"`
	ScopeType   string          `json:"scope_type" validate:"required,oneof=all provider game"`
	ScopeValue  string          `json:"scope_value"`
}

// UpdateRakebackScheduleRequest represents a request to update a scheduled rakeback
type UpdateRakebackScheduleRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Percentage  decimal.Decimal `json:"percentage" validate:"gte=0,lte=100"`
	ScopeType   string          `json:"scope_type" validate:"omitempty,oneof=all provider game"`
	ScopeValue  string          `json:"scope_value"`
}

// RakebackScheduleResponse represents the response with schedule details
type RakebackScheduleResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Percentage      decimal.Decimal `json:"percentage"`
	ScopeType       string          `json:"scope_type"`
	ScopeValue      *string         `json:"scope_value,omitempty"`
	Status          string          `json:"status"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty"`
	CreatedByName   *string         `json:"created_by_name,omitempty"`
	ActivatedAt     *time.Time      `json:"activated_at,omitempty"`
	DeactivatedAt   *time.Time      `json:"deactivated_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	IsActive        bool            `json:"is_active"`
	TimeRemaining   *string         `json:"time_remaining,omitempty"`
	TimeUntilStart  *string         `json:"time_until_start,omitempty"`
}

// ListRakebackSchedulesResponse represents a paginated list of schedules
type ListRakebackSchedulesResponse struct {
	Schedules  []RakebackScheduleResponse `json:"schedules"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}
