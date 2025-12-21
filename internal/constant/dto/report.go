package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DailyReportReq struct {
	Date string `form:"date" validate:"required,datetime=2006-01-02"`
}

type DailyReportRes struct {
	TotalPlayers int64             `json:"total_players"`
	NewPlayers   int64             `json:"new_players"`
	BucksEarned  int64             `json:"bucks_earned"`
	BucksSpent   int64             `json:"bucks_spent"`
	NetBucksFlow int64             `json:"net_bucks_flow"`
	Revenue      RevenueStream     `json:"revenue_streams"`
	Store        StoreTransaction  `json:"store_transactions"`
	Airtime      AirtimeConversion `json:"airtime_conversions"`
}

type RevenueStream struct {
	AdRevenueViews      int     `json:"ad_revenue_views"`
	AdRevenueBucks      int     `json:"ad_revenue_bucks"`
	AdRevenueAvgPerView float64 `json:"ad_revenue_avg_per_view"`
}

type StoreTransaction struct {
	Purchases         int     `json:"purchases"`
	TotalSpentBucks   int     `json:"total_spent_bucks"`
	AvgPerTransaction float64 `json:"avg_per_transaction"`
}

type AirtimeConversion struct {
	Conversions      int     `json:"conversions"`
	TotalValueBucks  int     `json:"total_value_bucks"`
	AvgPerConversion float64 `json:"avg_per_conversion"`
}

func ValidateDailyReportReq(req DailyReportReq) error {
	validate := validator.New()
	return validate.Struct(req)
}

// DuplicateIPAccount represents an account created from a duplicate IP
type DuplicateIPAccount struct {
	UserID      string `json:"user_id" db:"user_id"`
	Username    string `json:"username" db:"username"`
	Email       string `json:"email" db:"email"`
	UserAgent   string `json:"user_agent" db:"user_agent"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	SessionDate string `json:"session_date" db:"session_date"`
}

// DuplicateIPAccountsReport represents the report of accounts created from duplicate IPs
type DuplicateIPAccountsReport struct {
	IPAddress string               `json:"ip_address"`
	Count     int                  `json:"count"`
	Accounts  []DuplicateIPAccount `json:"accounts"`
}

// SuspendAccountsByIPReq represents the request to suspend all accounts from an IP
type SuspendAccountsByIPReq struct {
	IPAddress string `json:"ip_address" binding:"required"`
	Reason    string `json:"reason"`
	Note      string `json:"note"`
}

// SuspendAccountsByIPRes represents the response for suspending accounts by IP
type SuspendAccountsByIPRes struct {
	Message        string   `json:"message"`
	IPAddress      string   `json:"ip_address"`
	AccountsSuspended int   `json:"accounts_suspended"`
	UserIDs        []string `json:"user_ids"`
}

// BigWinnersReportReq represents the request for Big Winners report
type BigWinnersReportReq struct {
	Page            int        `form:"page" json:"page"`
	PerPage         int        `form:"per_page" json:"per_page"`
	DateFrom        *string    `form:"date_from" json:"date_from"`
	DateTo          *string    `form:"date_to" json:"date_to"`
	BrandID         *uuid.UUID `form:"brand_id" json:"brand_id"`
	GameProvider    *string    `form:"game_provider" json:"game_provider"`
	GameID          *string    `form:"game_id" json:"game_id"`
	PlayerSearch    *string    `form:"player_search" json:"player_search"` // username, email, or user ID
	MinWinThreshold *float64   `form:"min_win_threshold" json:"min_win_threshold"`
	BetType         *string    `form:"bet_type" json:"bet_type"`     // "cash", "bonus", "both"
	SortBy          *string    `form:"sort_by" json:"sort_by"`       // "win_amount", "net_win", "multiplier", "date"
	SortOrder       *string    `form:"sort_order" json:"sort_order"` // "asc", "desc"
	IsTestAccount   *bool      `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// BigWinner represents a single big winner entry
type BigWinner struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	DateTime      time.Time        `json:"date_time" db:"date_time"`
	PlayerID      uuid.UUID        `json:"player_id" db:"player_id"`
	Username      string           `json:"username" db:"username"`
	Email         *string          `json:"email" db:"email"`
	BrandID       *uuid.UUID       `json:"brand_id" db:"brand_id"`
	BrandName     *string          `json:"brand_name" db:"brand_name"`
	GameProvider  *string          `json:"game_provider" db:"game_provider"`
	GameID        *string          `json:"game_id" db:"game_id"`
	GameName      *string          `json:"game_name" db:"game_name"`
	BetID         *string          `json:"bet_id" db:"bet_id"`
	RoundID       *string          `json:"round_id" db:"round_id"`
	StakeAmount   decimal.Decimal  `json:"stake_amount" db:"stake_amount"`
	WinAmount     decimal.Decimal  `json:"win_amount" db:"win_amount"`
	NetWin        decimal.Decimal  `json:"net_win" db:"net_win"`
	Currency      string           `json:"currency" db:"currency"`
	WinMultiplier *decimal.Decimal `json:"win_multiplier" db:"win_multiplier"`
	BetType       string           `json:"bet_type" db:"bet_type"` // "cash", "bonus", "mixed"
	IsJackpot     bool             `json:"is_jackpot" db:"is_jackpot"`
	JackpotName   *string          `json:"jackpot_name" db:"jackpot_name"`
	SessionID     *string          `json:"session_id" db:"session_id"`
	Country       *string          `json:"country" db:"country"`
	BetSource     string           `json:"bet_source" db:"bet_source"` // "bets", "groove", "sport_bets", "plinko"
}

// BigWinnersReportRes represents the response for Big Winners report
type BigWinnersReportRes struct {
	Message    string            `json:"message"`
	Data       []BigWinner       `json:"data"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"total_pages"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	Summary    BigWinnersSummary `json:"summary"`
}

// BigWinnersSummary represents summary statistics for the report
type BigWinnersSummary struct {
	TotalWins    decimal.Decimal `json:"total_wins"`
	TotalNetWins decimal.Decimal `json:"total_net_wins"`
	TotalStakes  decimal.Decimal `json:"total_stakes"`
	Count        int64           `json:"count"`
}

// PlayerMetricsReportReq represents the request for Player Metrics report
type PlayerMetricsReportReq struct {
	Page             int        `form:"page" json:"page"`
	PerPage          int        `form:"per_page" json:"per_page"`
	PlayerSearch     *string    `form:"player_search" json:"player_search"` // username, email, or user ID
	BrandID          *uuid.UUID `form:"brand_id" json:"brand_id"`
	Currency         *string    `form:"currency" json:"currency"`
	Country          *string    `form:"country" json:"country"`
	RegistrationFrom *string    `form:"registration_from" json:"registration_from"`
	RegistrationTo   *string    `form:"registration_to" json:"registration_to"`
	LastActiveFrom   *string    `form:"last_active_from" json:"last_active_from"`
	LastActiveTo     *string    `form:"last_active_to" json:"last_active_to"`
	HasDeposited     *bool      `form:"has_deposited" json:"has_deposited"`
	HasWithdrawn     *bool      `form:"has_withdrawn" json:"has_withdrawn"`
	MinTotalDeposits *float64   `form:"min_total_deposits" json:"min_total_deposits"`
	MaxTotalDeposits *float64   `form:"max_total_deposits" json:"max_total_deposits"`
	MinTotalWagers   *float64   `form:"min_total_wagers" json:"min_total_wagers"`
	MaxTotalWagers   *float64   `form:"max_total_wagers" json:"max_total_wagers"`
	MinNetResult     *float64   `form:"min_net_result" json:"min_net_result"`
	MaxNetResult     *float64   `form:"max_net_result" json:"max_net_result"`
	SortBy           *string    `form:"sort_by" json:"sort_by"`       // "deposits", "wagering", "net_loss", "activity", "registration"
	SortOrder        *string    `form:"sort_order" json:"sort_order"` // "asc", "desc"
	IsTestAccount    *bool      `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// PlayerMetric represents a single player's metrics
type PlayerMetric struct {
	PlayerID         uuid.UUID       `json:"player_id" db:"player_id"`
	Username         string          `json:"username" db:"username"`
	Email            *string         `json:"email" db:"email"`
	BrandID          *uuid.UUID      `json:"brand_id" db:"brand_id"`
	BrandName        *string         `json:"brand_name" db:"brand_name"`
	Country          *string         `json:"country" db:"country"`
	RegistrationDate time.Time       `json:"registration_date" db:"registration_date"`
	LastActivity     *time.Time      `json:"last_activity" db:"last_activity"`
	MainBalance      decimal.Decimal `json:"main_balance" db:"main_balance"`
	Currency         string          `json:"currency" db:"currency"`
	TotalDeposits    decimal.Decimal `json:"total_deposits" db:"total_deposits"`
	TotalWithdrawals decimal.Decimal `json:"total_withdrawals" db:"total_withdrawals"`
	NetDeposits      decimal.Decimal `json:"net_deposits" db:"net_deposits"`
	TotalWagered     decimal.Decimal `json:"total_wagered" db:"total_wagered"`
	TotalWon         decimal.Decimal `json:"total_won" db:"total_won"`
	RakebackEarned   decimal.Decimal `json:"rakeback_earned" db:"rakeback_earned"`
	RakebackClaimed  decimal.Decimal `json:"rakeback_claimed" db:"rakeback_claimed"`
	NetGamingResult  decimal.Decimal `json:"net_gaming_result" db:"net_gaming_result"`
	NumberOfSessions int64           `json:"number_of_sessions" db:"number_of_sessions"`
	NumberOfBets     int64           `json:"number_of_bets" db:"number_of_bets"`
	AccountStatus    string          `json:"account_status" db:"account_status"`
	DeviceType       *string         `json:"device_type" db:"device_type"`
	KYCStatus        *string         `json:"kyc_status" db:"kyc_status"`
	FirstDepositDate *time.Time      `json:"first_deposit_date" db:"first_deposit_date"`
	LastDepositDate  *time.Time      `json:"last_deposit_date" db:"last_deposit_date"`
}

// PlayerMetricsReportRes represents the response for Player Metrics report
type PlayerMetricsReportRes struct {
	Message    string               `json:"message"`
	Data       []PlayerMetric       `json:"data"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"total_pages"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	Summary    PlayerMetricsSummary `json:"summary"`
}

// PlayerMetricsSummary represents summary statistics for the report
type PlayerMetricsSummary struct {
	TotalDeposits    decimal.Decimal `json:"total_deposits"`
	TotalNGR         decimal.Decimal `json:"total_ngr"`
	TotalWagers      decimal.Decimal `json:"total_wagers"`
	TotalWithdrawals decimal.Decimal `json:"total_withdrawals"`
	PlayerCount      int64           `json:"player_count"`
}

// PlayerTransactionsReq represents the request for player transaction drill-down
type PlayerTransactionsReq struct {
	PlayerID        uuid.UUID `form:"player_id" json:"player_id" uri:"player_id"`
	Page            int       `form:"page" json:"page"`
	PerPage         int       `form:"per_page" json:"per_page"`
	DateFrom        *string   `form:"date_from" json:"date_from"`
	DateTo          *string   `form:"date_to" json:"date_to"`
	TransactionType *string   `form:"transaction_type" json:"transaction_type"` // "deposit", "withdrawal", "bet", "win", "bonus", "adjustment"
	GameProvider    *string   `form:"game_provider" json:"game_provider"`
	GameID          *string   `form:"game_id" json:"game_id"`
	MinAmount       *float64  `form:"min_amount" json:"min_amount"`
	MaxAmount       *float64  `form:"max_amount" json:"max_amount"`
	SortBy          *string   `form:"sort_by" json:"sort_by"`       // "date", "amount", "net", "game"
	SortOrder       *string   `form:"sort_order" json:"sort_order"` // "asc", "desc"
}

// PlayerTransactionDetail represents a single transaction in drill-down
type PlayerTransactionDetail struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	TransactionID   string           `json:"transaction_id" db:"transaction_id"`
	Type            string           `json:"type" db:"type"` // "deposit", "withdrawal", "bet", "win", "bonus", "adjustment"
	DateTime        time.Time        `json:"date_time" db:"date_time"`
	Amount          decimal.Decimal  `json:"amount" db:"amount"`
	Currency        string           `json:"currency" db:"currency"`
	Status          string           `json:"status" db:"status"`
	GameProvider    *string          `json:"game_provider" db:"game_provider"`
	GameID          *string          `json:"game_id" db:"game_id"`
	GameName        *string          `json:"game_name" db:"game_name"`
	BetID           *string          `json:"bet_id" db:"bet_id"`
	RoundID         *string          `json:"round_id" db:"round_id"`
	BetAmount       *decimal.Decimal `json:"bet_amount" db:"bet_amount"`
	WinAmount       *decimal.Decimal `json:"win_amount" db:"win_amount"`
	RakebackEarned  *decimal.Decimal `json:"rakeback_earned" db:"rakeback_earned"`
	RakebackClaimed *decimal.Decimal `json:"rakeback_claimed" db:"rakeback_claimed"`
	RTP             *decimal.Decimal `json:"rtp" db:"rtp"`
	Multiplier      *decimal.Decimal `json:"multiplier" db:"multiplier"`
	GGR             *decimal.Decimal `json:"ggr" db:"ggr"`
	Net             *decimal.Decimal `json:"net" db:"net"`
	BetType         *string          `json:"bet_type" db:"bet_type"` // "cash", "bonus"
	PaymentMethod   *string          `json:"payment_method" db:"payment_method"`
	TXHash          *string          `json:"tx_hash" db:"tx_hash"`
	Network         *string          `json:"network" db:"network"`
	ChainID         *string          `json:"chain_id" db:"chain_id"`
	Fees            *decimal.Decimal `json:"fees" db:"fees"`
	Device          *string          `json:"device" db:"device"`
	IPAddress       *string          `json:"ip_address" db:"ip_address"`
	SessionID       *string          `json:"session_id" db:"session_id"`
}

// PlayerTransactionsRes represents the response for player transactions drill-down
type PlayerTransactionsRes struct {
	Message    string                    `json:"message"`
	Data       []PlayerTransactionDetail `json:"data"`
	Total      int64                     `json:"total"`
	TotalPages int                       `json:"total_pages"`
	Page       int                       `json:"page"`
	PerPage    int                       `json:"per_page"`
}

// CountryReportReq represents the request for Country Report
type CountryReportReq struct {
	Page              int       `form:"page" json:"page"`
	PerPage           int       `form:"per_page" json:"per_page"`
	DateFrom          *string   `form:"date_from" json:"date_from"`
	DateTo            *string   `form:"date_to" json:"date_to"`
	BrandID           *uuid.UUID `form:"brand_id" json:"brand_id"`
	Currency          *string   `form:"currency" json:"currency"`
	Countries         []string  `form:"countries" json:"countries"` // Multi-select countries
	AcquisitionChannel *string  `form:"acquisition_channel" json:"acquisition_channel"`
	UserType          *string   `form:"user_type" json:"user_type"` // "depositors", "all", "active"
	SortBy            *string   `form:"sort_by" json:"sort_by"` // "deposits", "ngr", "active_users", "alphabetical"
	SortOrder         *string   `form:"sort_order" json:"sort_order"` // "asc", "desc"
	ConvertToBaseCurrency *bool `form:"convert_to_base_currency" json:"convert_to_base_currency"`
	IsTestAccount     *bool     `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// CountryMetric represents aggregated metrics for a single country
type CountryMetric struct {
	Country                string          `json:"country" db:"country"`
	TotalRegistrations     int64           `json:"total_registrations" db:"total_registrations"`
	ActivePlayers          int64           `json:"active_players" db:"active_players"`
	FirstTimeDepositors    int64           `json:"first_time_depositors" db:"first_time_depositors"`
	TotalDepositors        int64           `json:"total_depositors" db:"total_depositors"`
	TotalDeposits          decimal.Decimal `json:"total_deposits" db:"total_deposits"`
	TotalWithdrawals       decimal.Decimal `json:"total_withdrawals" db:"total_withdrawals"`
	NetPosition            decimal.Decimal  `json:"net_position" db:"net_position"`
	TotalWagered           decimal.Decimal `json:"total_wagered" db:"total_wagered"`
	TotalWon               decimal.Decimal `json:"total_won" db:"total_won"`
	GGR                    decimal.Decimal `json:"ggr" db:"ggr"`
	NGR                    decimal.Decimal `json:"ngr" db:"ngr"`
	AverageDepositPerPlayer decimal.Decimal `json:"average_deposit_per_player" db:"average_deposit_per_player"`
	AverageWagerPerPlayer  decimal.Decimal `json:"average_wager_per_player" db:"average_wager_per_player"`
	RakebackEarned         decimal.Decimal `json:"rakeback_earned" db:"rakeback_earned"`
	RakebackConverted      decimal.Decimal `json:"rakeback_converted" db:"rakeback_converted"`
	SelfExclusions         int64           `json:"self_exclusions" db:"self_exclusions"`
}

// CountryReportRes represents the response for Country Report
type CountryReportRes struct {
	Message   string          `json:"message"`
	Data      []CountryMetric `json:"data"`
	Total     int64           `json:"total"`
	TotalPages int            `json:"total_pages"`
	Page      int             `json:"page"`
	PerPage   int             `json:"per_page"`
	Summary   CountryReportSummary `json:"summary"`
}

// CountryReportSummary represents summary statistics for the report
type CountryReportSummary struct {
	TotalDeposits   decimal.Decimal `json:"total_deposits"`
	TotalNGR        decimal.Decimal `json:"total_ngr"`
	TotalActiveUsers int64          `json:"total_active_users"`
	TotalDepositors  int64          `json:"total_depositors"`
	TotalRegistrations int64        `json:"total_registrations"`
}

// CountryPlayersReq represents the request for country players drill-down
type CountryPlayersReq struct {
	Country      string   `form:"country" json:"country" uri:"country"`
	Page         int      `form:"page" json:"page"`
	PerPage      int      `form:"per_page" json:"per_page"`
	DateFrom     *string  `form:"date_from" json:"date_from"`
	DateTo       *string  `form:"date_to" json:"date_to"`
	MinDeposits  *float64 `form:"min_deposits" json:"min_deposits"`
	MaxDeposits  *float64 `form:"max_deposits" json:"max_deposits"`
	ActivityFrom *string  `form:"activity_from" json:"activity_from"`
	ActivityTo   *string  `form:"activity_to" json:"activity_to"`
	KYCStatus    *string  `form:"kyc_status" json:"kyc_status"`
	MinBalance   *float64 `form:"min_balance" json:"min_balance"`
	MaxBalance   *float64 `form:"max_balance" json:"max_balance"`
	SortBy       *string  `form:"sort_by" json:"sort_by"`
	SortOrder    *string  `form:"sort_order" json:"sort_order"`
}

// CountryPlayer represents a player in country drill-down
type CountryPlayer struct {
	PlayerID        uuid.UUID       `json:"player_id" db:"player_id"`
	Username        string          `json:"username" db:"username"`
	Email           *string         `json:"email" db:"email"`
	Country         string          `json:"country" db:"country"`
	TotalDeposits   decimal.Decimal `json:"total_deposits" db:"total_deposits"`
	TotalWithdrawals decimal.Decimal `json:"total_withdrawals" db:"total_withdrawals"`
	TotalWagered    decimal.Decimal `json:"total_wagered" db:"total_wagered"`
	NGR             decimal.Decimal `json:"ngr" db:"ngr"`
	LastActivity    *time.Time      `json:"last_activity" db:"last_activity"`
	RegistrationDate time.Time      `json:"registration_date" db:"registration_date"`
	Balance         decimal.Decimal `json:"balance" db:"balance"`
	Currency        string          `json:"currency" db:"currency"`
}

// CountryPlayersRes represents the response for country players drill-down
type CountryPlayersRes struct {
	Message   string          `json:"message"`
	Data      []CountryPlayer `json:"data"`
	Total     int64           `json:"total"`
	TotalPages int            `json:"total_pages"`
	Page      int             `json:"page"`
	PerPage   int             `json:"per_page"`
}

// GamePerformanceReportReq represents the request for Game Performance Report
type GamePerformanceReportReq struct {
	Page          int       `form:"page" json:"page"`
	PerPage       int       `form:"per_page" json:"per_page"`
	DateFrom      *string   `form:"date_from" json:"date_from"`
	DateTo        *string   `form:"date_to" json:"date_to"`
	BrandID       *uuid.UUID `form:"brand_id" json:"brand_id"`
	Currency      *string   `form:"currency" json:"currency"`
	GameProvider  *string   `form:"game_provider" json:"game_provider"`
	GameID        *string   `form:"game_id" json:"game_id"`
	Category      *string   `form:"category" json:"category"`
	SortBy        *string   `form:"sort_by" json:"sort_by"` // "ggr", "ngr", "most_played", "rtp", "bet_volume"
	SortOrder     *string   `form:"sort_order" json:"sort_order"` // "asc", "desc"
	IsTestAccount *bool     `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// GamePerformanceMetric represents aggregated metrics for a single game
type GamePerformanceMetric struct {
	GameID          string          `json:"game_id" db:"game_id"`
	GameName        string          `json:"game_name" db:"game_name"`
	Provider        string          `json:"provider" db:"provider"`
	Category        *string         `json:"category" db:"category"`
	TotalBets       int64           `json:"total_bets" db:"total_bets"`
	TotalRounds     int64           `json:"total_rounds" db:"total_rounds"`
	UniquePlayers   int64           `json:"unique_players" db:"unique_players"`
	TotalStake      decimal.Decimal `json:"total_stake" db:"total_stake"`
	TotalWin        decimal.Decimal `json:"total_win" db:"total_win"`
	GGR             decimal.Decimal `json:"ggr" db:"ggr"`
	NGR             decimal.Decimal `json:"ngr" db:"ngr"`
	EffectiveRTP    decimal.Decimal `json:"effective_rtp" db:"effective_rtp"`
	AvgBetAmount    decimal.Decimal `json:"avg_bet_amount" db:"avg_bet_amount"`
	RakebackEarned  decimal.Decimal `json:"rakeback_earned" db:"rakeback_earned"`
	BigWinsCount    int64           `json:"big_wins_count" db:"big_wins_count"`
	HighestWin      decimal.Decimal `json:"highest_win" db:"highest_win"`
	HighestMultiplier decimal.Decimal `json:"highest_multiplier" db:"highest_multiplier"`
}

// GamePerformanceReportRes represents the response for Game Performance Report
type GamePerformanceReportRes struct {
	Message   string                 `json:"message"`
	Data      []GamePerformanceMetric `json:"data"`
	Total     int64                  `json:"total"`
	TotalPages int                   `json:"total_pages"`
	Page      int                    `json:"page"`
	PerPage   int                    `json:"per_page"`
	Summary   GamePerformanceSummary `json:"summary"`
}

// GamePerformanceSummary represents summary statistics for the report
type GamePerformanceSummary struct {
	TotalBets      int64           `json:"total_bets"`
	TotalUniquePlayers int64       `json:"total_unique_players"`
	TotalWagered   decimal.Decimal `json:"total_wagered"`
	TotalGGR       decimal.Decimal `json:"total_ggr"`
	TotalRakeback  decimal.Decimal `json:"total_rakeback"`
	AverageRTP     decimal.Decimal `json:"average_rtp"`
}

// GamePlayersReq represents the request for game players drill-down
type GamePlayersReq struct {
	GameID      string   `form:"game_id" json:"game_id" uri:"game_id"`
	Page        int      `form:"page" json:"page"`
	PerPage     int      `form:"per_page" json:"per_page"`
	DateFrom    *string  `form:"date_from" json:"date_from"`
	DateTo      *string  `form:"date_to" json:"date_to"`
	Currency    *string  `form:"currency" json:"currency"`
	BetType     *string  `form:"bet_type" json:"bet_type"`
	MinStake    *float64 `form:"min_stake" json:"min_stake"`
	MaxStake    *float64 `form:"max_stake" json:"max_stake"`
	MinNetResult *float64 `form:"min_net_result" json:"min_net_result"`
	MaxNetResult *float64 `form:"max_net_result" json:"max_net_result"`
	SortBy      *string  `form:"sort_by" json:"sort_by"`
	SortOrder   *string  `form:"sort_order" json:"sort_order"`
}

// GamePlayer represents a player in game drill-down
type GamePlayer struct {
	PlayerID      uuid.UUID       `json:"player_id" db:"player_id"`
	Username      string          `json:"username" db:"username"`
	Email         *string         `json:"email" db:"email"`
	TotalStake    decimal.Decimal `json:"total_stake" db:"total_stake"`
	TotalWin      decimal.Decimal `json:"total_win" db:"total_win"`
	NGR           decimal.Decimal `json:"ngr" db:"ngr"`
	Rakeback      decimal.Decimal `json:"rakeback" db:"rakeback"`
	NumberOfRounds int64          `json:"number_of_rounds" db:"number_of_rounds"`
	LastPlayed    *time.Time      `json:"last_played" db:"last_played"`
	Currency      string          `json:"currency" db:"currency"`
}

// GamePlayersRes represents the response for game players drill-down
type GamePlayersRes struct {
	Message   string        `json:"message"`
	Data      []GamePlayer  `json:"data"`
	Total     int64         `json:"total"`
	TotalPages int          `json:"total_pages"`
	Page      int           `json:"page"`
	PerPage   int           `json:"per_page"`
}

// ProviderPerformanceReportReq represents the request for Provider Performance Report
type ProviderPerformanceReportReq struct {
	Page          int       `form:"page" json:"page"`
	PerPage       int       `form:"per_page" json:"per_page"`
	DateFrom      *string   `form:"date_from" json:"date_from"`
	DateTo        *string   `form:"date_to" json:"date_to"`
	BrandID       *uuid.UUID `form:"brand_id" json:"brand_id"`
	Currency      *string   `form:"currency" json:"currency"`
	Provider      *string   `form:"provider" json:"provider"`
	Category      *string   `form:"category" json:"category"`
	SortBy        *string   `form:"sort_by" json:"sort_by"` // "ggr", "ngr", "most_played", "rtp", "bet_volume"
	SortOrder     *string   `form:"sort_order" json:"sort_order"` // "asc", "desc"
	IsTestAccount *bool     `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// ProviderPerformanceMetric represents aggregated metrics for a single provider
type ProviderPerformanceMetric struct {
	Provider        string          `json:"provider" db:"provider"`
	TotalGames      int64           `json:"total_games" db:"total_games"`
	TotalBets       int64           `json:"total_bets" db:"total_bets"`
	TotalRounds     int64           `json:"total_rounds" db:"total_rounds"`
	UniquePlayers   int64           `json:"unique_players" db:"unique_players"`
	TotalStake      decimal.Decimal `json:"total_stake" db:"total_stake"`
	TotalWin        decimal.Decimal `json:"total_win" db:"total_win"`
	GGR             decimal.Decimal `json:"ggr" db:"ggr"`
	NGR             decimal.Decimal `json:"ngr" db:"ngr"`
	EffectiveRTP    decimal.Decimal `json:"effective_rtp" db:"effective_rtp"`
	AvgBetAmount    decimal.Decimal `json:"avg_bet_amount" db:"avg_bet_amount"`
	RakebackEarned  decimal.Decimal `json:"rakeback_earned" db:"rakeback_earned"`
	BigWinsCount    int64           `json:"big_wins_count" db:"big_wins_count"`
	HighestWin      decimal.Decimal `json:"highest_win" db:"highest_win"`
	HighestMultiplier decimal.Decimal `json:"highest_multiplier" db:"highest_multiplier"`
}

// ProviderPerformanceReportRes represents the response for Provider Performance Report
type ProviderPerformanceReportRes struct {
	Message   string                     `json:"message"`
	Data      []ProviderPerformanceMetric `json:"data"`
	Total     int64                     `json:"total"`
	TotalPages int                      `json:"total_pages"`
	Page      int                       `json:"page"`
	PerPage   int                       `json:"per_page"`
	Summary   GamePerformanceSummary    `json:"summary"`
}

// AffiliateReportReq represents the request for Affiliate Report
type AffiliateReportReq struct {
	DateFrom     *string `form:"date_from" json:"date_from"`         // YYYY-MM-DD format
	DateTo       *string `form:"date_to" json:"date_to"`             // YYYY-MM-DD format
	ReferralCode *string `form:"referral_code" json:"referral_code"` // Optional filter by referral code
	IsTestAccount *bool  `form:"is_test_account" json:"is_test_account"` // Filter by test account (false = real accounts only, true = test accounts only)
}

// AffiliateReportRow represents a single day's metrics for an affiliate
type AffiliateReportRow struct {
	Date              string          `json:"date" db:"date"`
	ReferralCode      string          `json:"referral_code" db:"referral_code"`
	Registrations     int64           `json:"registrations" db:"registrations"`
	UniqueDepositors  int64           `json:"unique_depositors" db:"unique_depositors"`
	ActiveCustomers   int64           `json:"active_customers" db:"active_customers"`
	TotalBets         int64           `json:"total_bets" db:"total_bets"`
	GGR               decimal.Decimal `json:"ggr" db:"ggr"`
	NGR               decimal.Decimal `json:"ngr" db:"ngr"`
	DepositsUSD       decimal.Decimal `json:"deposits_usd" db:"deposits_usd"`
	WithdrawalsUSD    decimal.Decimal `json:"withdrawals_usd" db:"withdrawals_usd"`
}

// AffiliateReportRes represents the response for Affiliate Report
type AffiliateReportRes struct {
	Message string              `json:"message"`
	Data    []AffiliateReportRow `json:"data"`
	Summary AffiliateReportSummary `json:"summary"`
}

// AffiliateReportSummary represents summary statistics for the report
type AffiliateReportSummary struct {
	TotalRegistrations    int64           `json:"total_registrations"`
	TotalUniqueDepositors int64           `json:"total_unique_depositors"`
	TotalActiveCustomers  int64           `json:"total_active_customers"`
	TotalBets             int64           `json:"total_bets"`
	TotalGGR              decimal.Decimal `json:"total_ggr"`
	TotalNGR              decimal.Decimal `json:"total_ngr"`
	TotalDepositsUSD      decimal.Decimal `json:"total_deposits_usd"`
	TotalWithdrawalsUSD   decimal.Decimal `json:"total_withdrawals_usd"`
}
