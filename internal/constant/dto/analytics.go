package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AnalyticsTransaction represents a transaction in ClickHouse
type AnalyticsTransaction struct {
	ID                    string           `json:"id"`
	UserID                uuid.UUID        `json:"user_id"`
	TransactionType       string           `json:"transaction_type"` // deposit, withdrawal, bet, win, bonus, cashback, etc.
	Amount                decimal.Decimal  `json:"amount"`
	Currency              string           `json:"currency"`
	Status                string           `json:"status"` // pending, completed, failed, cancelled
	GameID                *string          `json:"game_id,omitempty"`
	GameName              *string          `json:"game_name,omitempty"`
	Provider              *string          `json:"provider,omitempty"`
	SessionID             *string          `json:"session_id,omitempty"`
	RoundID               *string          `json:"round_id,omitempty"`
	BetAmount             *decimal.Decimal `json:"bet_amount,omitempty"`
	WinAmount             *decimal.Decimal `json:"win_amount,omitempty"`
	NetResult             *decimal.Decimal `json:"net_result,omitempty"`
	BalanceBefore         decimal.Decimal  `json:"balance_before"`
	BalanceAfter          decimal.Decimal  `json:"balance_after"`
	PaymentMethod         *string          `json:"payment_method,omitempty"`
	ExternalTransactionID *string          `json:"external_transaction_id,omitempty"`
	Metadata              *string          `json:"metadata,omitempty"` // JSON metadata
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}

// TransactionFilters for querying transactions
type TransactionFilters struct {
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	DateFrom        *time.Time `json:"date_from,omitempty"`
	DateTo          *time.Time `json:"date_to,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	GameID          *string    `json:"game_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	BrandID         *string    `json:"brand_id,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
	Search          *string    `json:"search,omitempty"`
}

// RakebackFilters for querying rakeback transactions
type RakebackFilters struct {
	TransactionType string     `json:"transaction_type"`    // "earned" or "claimed"
	Status          *string    `json:"status,omitempty"`    // available|completed|claimed
	DateFrom        *time.Time `json:"date_from,omitempty"` // optional, future-proof
	DateTo          *time.Time `json:"date_to,omitempty"`   // optional, future-proof
	BrandID         *string    `json:"brand_id,omitempty"`  // brand isolation
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// TipFilters for querying tip transactions
type TipFilters struct {
	Status   *string    `json:"status,omitempty"`
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
	BrandID  *string    `json:"brand_id,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	Offset   int        `json:"offset,omitempty"`
}

// for querying welcome bonus transactions
type WelcomeBonusFilters struct {
	UserID    *uuid.UUID       `json:"user_id,omitempty"` // Optional: filter by specific user
	Status    *string          `json:"status,omitempty"`
	DateFrom  *time.Time       `json:"date_from,omitempty"`
	DateTo    *time.Time       `json:"date_to,omitempty"`
	BrandID   *string          `json:"brand_id,omitempty"`
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"` // Filter by minimum amount
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"` // Filter by maximum amount
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// for analytics queries
type DateRange struct {
	From          *time.Time  `json:"from,omitempty"`
	To            *time.Time  `json:"to,omitempty"`
	IsTestAccount *bool       `json:"is_test_account,omitempty"` // nil = all, true = test accounts only, false = real accounts only
	UserIDs       []uuid.UUID `json:"user_ids,omitempty"`        // Filtered user IDs based on is_test_account
}

// represents user analytics data
type UserAnalytics struct {
	UserID            uuid.UUID       `json:"user_id"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	TotalBonuses      decimal.Decimal `json:"total_bonuses"`
	TotalCashback     decimal.Decimal `json:"total_cashback"`
	NetLoss           decimal.Decimal `json:"net_loss"`
	TransactionCount  uint64          `json:"transaction_count"`
	UniqueGamesPlayed uint64          `json:"unique_games_played"`
	SessionCount      uint64          `json:"session_count"`
	AvgBetAmount      decimal.Decimal `json:"avg_bet_amount"`
	MaxBetAmount      decimal.Decimal `json:"max_bet_amount"`
	MinBetAmount      decimal.Decimal `json:"min_bet_amount"`
	LastActivity      time.Time       `json:"last_activity"`
}

// GameAnalytics represents game analytics data
type GameAnalytics struct {
	GameID        string          `json:"game_id"`
	GameName      string          `json:"game_name"`
	Provider      string          `json:"provider"`
	TotalBets     decimal.Decimal `json:"total_bets"`
	TotalWins     decimal.Decimal `json:"total_wins"`
	TotalPlayers  uint64          `json:"total_players"`
	TotalSessions uint64          `json:"total_sessions"`
	TotalRounds   uint64          `json:"total_rounds"`
	AvgBetAmount  decimal.Decimal `json:"avg_bet_amount"`
	MaxBetAmount  decimal.Decimal `json:"max_bet_amount"`
	MinBetAmount  decimal.Decimal `json:"min_bet_amount"`
	RTP           decimal.Decimal `json:"rtp"`        // Return to Player percentage
	Volatility    string          `json:"volatility"` // low, medium, high
}

// SessionAnalytics represents session analytics data
type SessionAnalytics struct {
	SessionID       string          `json:"session_id"`
	UserID          uuid.UUID       `json:"user_id"`
	GameID          *string         `json:"game_id,omitempty"`
	GameName        *string         `json:"game_name,omitempty"`
	Provider        *string         `json:"provider,omitempty"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         *time.Time      `json:"end_time,omitempty"`
	DurationSeconds *uint64         `json:"duration_seconds,omitempty"`
	TotalBets       decimal.Decimal `json:"total_bets"`
	TotalWins       decimal.Decimal `json:"total_wins"`
	NetResult       decimal.Decimal `json:"net_result"`
	BetCount        uint64          `json:"bet_count"`
	WinCount        uint64          `json:"win_count"`
	MaxBalance      decimal.Decimal `json:"max_balance"`
	MinBalance      decimal.Decimal `json:"min_balance"`
	SessionType     string          `json:"session_type"` // regular, bonus, free_play
	DeviceType      *string         `json:"device_type,omitempty"`
	IPAddress       *string         `json:"ip_address,omitempty"`
	UserAgent       *string         `json:"user_agent,omitempty"`
}

// RealTimeStats represents real-time statistics
type RealTimeStats struct {
	Timestamp         time.Time       `json:"timestamp"`
	TotalTransactions uint64          `json:"total_transactions"`
	DepositsCount     uint64          `json:"deposits_count"`
	WithdrawalsCount  uint64          `json:"withdrawals_count"`
	BetsCount         uint64          `json:"bets_count"`
	WinsCount         uint64          `json:"wins_count"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
}

// DailyReport represents daily analytics report
type DailyReport struct {
	Date              time.Time       `json:"date"`
	TotalTransactions uint64          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
	NewUsers          uint64          `json:"new_users"`
	UniqueDepositors  uint64          `json:"unique_depositors"`
	UniqueWithdrawers uint64          `json:"unique_withdrawers"`

	// Additional metrics for comprehensive reporting
	DepositCount     uint64          `json:"deposit_count"`
	WithdrawalCount  uint64          `json:"withdrawal_count"`
	BetCount         uint64          `json:"bet_count"`
	WinCount         uint64          `json:"win_count"`
	CashbackEarned   decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections decimal.Decimal `json:"admin_corrections"`

	TopGames   []GameStats   `json:"top_games"`
	TopPlayers []PlayerStats `json:"top_players"`
}

// EnhancedDailyReport represents daily analytics report with comparison metrics
type EnhancedDailyReport struct {
	Date              time.Time       `json:"date"`
	TotalTransactions uint64          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	GGR               decimal.Decimal `json:"ggr"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
	NewUsers          uint64          `json:"new_users"`
	UniqueDepositors  uint64          `json:"unique_depositors"`
	UniqueWithdrawers uint64          `json:"unique_withdrawers"`

	DepositCount     uint64          `json:"deposit_count"`
	WithdrawalCount  uint64          `json:"withdrawal_count"`
	BetCount         uint64          `json:"bet_count"`
	WinCount         uint64          `json:"win_count"`
	CashbackEarned   decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections decimal.Decimal `json:"admin_corrections"`

	PreviousDayChange DailyReportComparison `json:"previous_day_change"`
	MTD               DailyReportMTD        `json:"mtd"`
	SPLM              DailyReportSPLM       `json:"splm"`
	MTDvsSPLMChange   DailyReportComparison `json:"mtd_vs_splm_change"`

	TopGames   []GameStats   `json:"top_games"`
	TopPlayers []PlayerStats `json:"top_players"`
}

// DailyReportComparison represents percentage change comparison
type DailyReportComparison struct {
	TotalTransactionsChange string `json:"total_transactions_change"`
	TotalDepositsChange     string `json:"total_deposits_change"`
	TotalWithdrawalsChange  string `json:"total_withdrawals_change"`
	TotalBetsChange         string `json:"total_bets_change"`
	TotalWinsChange         string `json:"total_wins_change"`
	GGRChange               string `json:"ggr_change"`
	NetRevenueChange        string `json:"net_revenue_change"`
	ActiveUsersChange       string `json:"active_users_change"`
	ActiveGamesChange       string `json:"active_games_change"`
	NewUsersChange          string `json:"new_users_change"`
	UniqueDepositorsChange  string `json:"unique_depositors_change"`
	UniqueWithdrawersChange string `json:"unique_withdrawers_change"`
	DepositCountChange      string `json:"deposit_count_change"`
	WithdrawalCountChange   string `json:"withdrawal_count_change"`
	BetCountChange          string `json:"bet_count_change"`
	WinCountChange          string `json:"win_count_change"`
	CashbackEarnedChange    string `json:"cashback_earned_change"`
	CashbackClaimedChange   string `json:"cashback_claimed_change"`
	AdminCorrectionsChange  string `json:"admin_corrections_change"`
}

// DailyReportMTD represents Month To Date metrics
type DailyReportMTD struct {
	TotalTransactions uint64          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	GGR               decimal.Decimal `json:"ggr"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
	NewUsers          uint64          `json:"new_users"`
	UniqueDepositors  uint64          `json:"unique_depositors"`
	UniqueWithdrawers uint64          `json:"unique_withdrawers"`
	DepositCount      uint64          `json:"deposit_count"`
	WithdrawalCount   uint64          `json:"withdrawal_count"`
	BetCount          uint64          `json:"bet_count"`
	WinCount          uint64          `json:"win_count"`
	CashbackEarned    decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections  decimal.Decimal `json:"admin_corrections"`
}

// DailyReportSPLM represents Same Period Last Month metrics
type DailyReportSPLM struct {
	TotalTransactions uint64          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	GGR               decimal.Decimal `json:"ggr"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
	NewUsers          uint64          `json:"new_users"`
	UniqueDepositors  uint64          `json:"unique_depositors"`
	UniqueWithdrawers uint64          `json:"unique_withdrawers"`
	DepositCount      uint64          `json:"deposit_count"`
	WithdrawalCount   uint64          `json:"withdrawal_count"`
	BetCount          uint64          `json:"bet_count"`
	WinCount          uint64          `json:"win_count"`
	CashbackEarned    decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections  decimal.Decimal `json:"admin_corrections"`
}

// DailyReportDataTableRow represents a single row in the daily report data table
type DailyReportDataTableRow struct {
	Date              string          `json:"date"`
	NewUsers          uint64          `json:"new_users"`
	UniqueDepositors  uint64          `json:"unique_depositors"`
	UniqueWithdrawers uint64          `json:"unique_withdrawers"`
	ActiveUsers       uint64          `json:"active_users"`
	BetCount          uint64          `json:"bet_count"`
	BetAmount         decimal.Decimal `json:"bet_amount"`
	WinAmount         decimal.Decimal `json:"win_amount"`
	GGR               decimal.Decimal `json:"ggr"`
	CashbackEarned    decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal `json:"cashback_claimed"`
	DepositCount      uint64          `json:"deposit_count"`
	DepositAmount     decimal.Decimal `json:"deposit_amount"`
	WithdrawalCount   uint64          `json:"withdrawal_count"`
	WithdrawalAmount  decimal.Decimal `json:"withdrawal_amount"`
	AdminCorrections  decimal.Decimal `json:"admin_corrections"`
}

// DailyReportDataTableResponse represents the data table response with rows and totals
type DailyReportDataTableResponse struct {
	Rows   []DailyReportDataTableRow `json:"rows"`
	Totals DailyReportDataTableRow   `json:"totals"`
}

// WeeklyReport represents weekly analytics report with comparison metrics
type WeeklyReport struct {
	WeekStart         time.Time                 `json:"week_start"`
	WeekEnd           time.Time                 `json:"week_end"`
	NewUsers          uint64                    `json:"new_users"`
	UniqueDepositors  uint64                    `json:"unique_depositors"`
	UniqueWithdrawers uint64                    `json:"unique_withdrawers"`
	ActiveUsers       uint64                    `json:"active_users"`
	BetCount          uint64                    `json:"bet_count"`
	BetAmount         decimal.Decimal           `json:"bet_amount"`
	WinAmount         decimal.Decimal           `json:"win_amount"`
	GGR               decimal.Decimal           `json:"ggr"`
	CashbackEarned    decimal.Decimal           `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal           `json:"cashback_claimed"`
	DepositCount      uint64                    `json:"deposit_count"`
	DepositAmount     decimal.Decimal           `json:"deposit_amount"`
	WithdrawalCount   uint64                    `json:"withdrawal_count"`
	WithdrawalAmount  decimal.Decimal           `json:"withdrawal_amount"`
	AdminCorrections  decimal.Decimal           `json:"admin_corrections"`
	DailyBreakdown    []DailyReportDataTableRow `json:"daily_breakdown"`
	MTD               DailyReportMTD            `json:"mtd"`
	SPLM              DailyReportSPLM           `json:"splm"`
	MTDvsSPLMChange   DailyReportComparison     `json:"mtd_vs_splm_change"`
}

// MonthlyReport represents monthly analytics report
type MonthlyReport struct {
	Year              int             `json:"year"`
	Month             int             `json:"month"`
	TotalTransactions uint64          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint64          `json:"active_users"`
	ActiveGames       uint64          `json:"active_games"`
	NewUsers          uint64          `json:"new_users"`
	AvgDailyRevenue   decimal.Decimal `json:"avg_daily_revenue"`
	TopGames          []GameStats     `json:"top_games"`
	TopPlayers        []PlayerStats   `json:"top_players"`
}

// GameStats represents game statistics
type GameStats struct {
	GameID       string          `json:"game_id"`
	GameName     string          `json:"game_name"`
	Provider     string          `json:"provider"`
	TotalBets    decimal.Decimal `json:"total_bets"`
	TotalWins    decimal.Decimal `json:"total_wins"`
	NetRevenue   decimal.Decimal `json:"net_revenue"`
	PlayerCount  uint64          `json:"player_count"`
	SessionCount uint64          `json:"session_count"`
	AvgBetAmount decimal.Decimal `json:"avg_bet_amount"`
	RTP          decimal.Decimal `json:"rtp"`
	Rank         int             `json:"rank"`
}

// PlayerStats represents player statistics
type PlayerStats struct {
	UserID            uuid.UUID       `json:"user_id"`
	Username          string          `json:"username"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetLoss           decimal.Decimal `json:"net_loss"`
	TransactionCount  uint64          `json:"transaction_count"`
	UniqueGamesPlayed uint64          `json:"unique_games_played"`
	SessionCount      uint64          `json:"session_count"`
	AvgBetAmount      decimal.Decimal `json:"avg_bet_amount"`
	LastActivity      time.Time       `json:"last_activity"`
	Rank              int             `json:"rank"`
}

// RakebackTransaction represents a single rakeback earning or claim row
type RakebackTransaction struct {
	ID              string           `json:"id"`
	UserID          uuid.UUID        `json:"user_id"`
	TransactionType string           `json:"transaction_type"` // "earned" or "claimed"
	RakebackAmount  decimal.Decimal  `json:"rakeback_amount"`
	Currency        string           `json:"currency"`
	Status          string           `json:"status"`
	GameID          *string          `json:"game_id,omitempty"`
	GameName        *string          `json:"game_name,omitempty"`
	Provider        *string          `json:"provider,omitempty"`
	ProcessingFee   *decimal.Decimal `json:"processing_fee,omitempty"`
	NetAmount       *decimal.Decimal `json:"net_amount,omitempty"`
	ClaimedAt       *time.Time       `json:"claimed_at,omitempty"`
	ClaimedEarnings *string          `json:"claimed_earnings,omitempty"`
	ClaimID         *string          `json:"claim_id,omitempty"`
	EarningID       *string          `json:"earning_id,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// TipTransaction represents a single tip sent/received row
type TipTransaction struct {
	ID                    string          `json:"id"`
	UserID                uuid.UUID       `json:"user_id"`
	TransactionType       string          `json:"transaction_type"` // "tip_sent" or "tip_received"
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Status                string          `json:"status"`
	BalanceBefore         decimal.Decimal `json:"balance_before"`
	BalanceAfter          decimal.Decimal `json:"balance_after"`
	ExternalTransactionID *string         `json:"external_transaction_id,omitempty"`
	Metadata              *string         `json:"metadata,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	SenderUsername        *string         `json:"sender_username,omitempty"`   // Populated from users table
	ReceiverUsername      *string         `json:"receiver_username,omitempty"` // Populated from users table
}

// WelcomeBonusTransaction represents a single welcome bonus transaction
type WelcomeBonusTransaction struct {
	ID                    string          `json:"id"`
	UserID                uuid.UUID       `json:"user_id"`
	Username              *string         `json:"username,omitempty"`
	Email                 *string         `json:"email,omitempty"`
	TransactionType       string          `json:"transaction_type"` // "welcome_bonus"
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Status                string          `json:"status"`
	BalanceBefore         decimal.Decimal `json:"balance_before"`
	BalanceAfter          decimal.Decimal `json:"balance_after"`
	ExternalTransactionID *string         `json:"external_transaction_id,omitempty"`
	Metadata              *string         `json:"metadata,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
}

// UserTransactionsTotals represents summary totals for game transactions
type UserTransactionsTotals struct {
	TotalCount     uint64          `json:"total_count"`
	TotalBetAmount decimal.Decimal `json:"total_bet_amount"`
	TotalWinAmount decimal.Decimal `json:"total_win_amount"`
	NetResult      decimal.Decimal `json:"net_result"`
}

// represents summary totals for rakeback
type UserRakebackTotals struct {
	TotalEarnedCount   uint64          `json:"total_earned_count"`
	TotalEarnedAmount  decimal.Decimal `json:"total_earned_amount"`
	TotalClaimedCount  uint64          `json:"total_claimed_count"`
	TotalClaimedAmount decimal.Decimal `json:"total_claimed_amount"`
	AvailableRakeback  decimal.Decimal `json:"available_rakeback"`
}

// represents summary totals for tip transactions
type UserTipsTotals struct {
	TotalTipsCount      uint64          `json:"total_tips_count"`
	TotalSentCount      uint64          `json:"total_sent_count"`
	TotalSentAmount     decimal.Decimal `json:"total_sent_amount"`
	TotalReceivedCount  uint64          `json:"total_received_count"`
	TotalReceivedAmount decimal.Decimal `json:"total_received_amount"`
	NetTips             decimal.Decimal `json:"net_tips"`
}

// represents summary totals for welcome bonus transactions
type UserWelcomeBonusTotals struct {
	TotalCount  uint64          `json:"total_count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}

// represents a balance snapshot at a point in time
type BalanceSnapshot struct {
	UserID          uuid.UUID       `json:"user_id"`
	Balance         decimal.Decimal `json:"balance"`
	Currency        string          `json:"currency"`
	SnapshotTime    time.Time       `json:"snapshot_time"`
	TransactionID   *string         `json:"transaction_id,omitempty"`
	TransactionType *string         `json:"transaction_type,omitempty"`
}

// AnalyticsResponse represents a generic analytics API response
type AnalyticsResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta represents pagination and metadata
type Meta struct {
	Total              int     `json:"total"`
	Page               int     `json:"page"`
	PageSize           int     `json:"page_size"`
	Pages              int     `json:"pages"`
	TotalClaimedAmount *string `json:"total_claimed_amount,omitempty"` // Only for rakeback endpoints
	TotalBetAmount     *string `json:"total_bet_amount,omitempty"`     // Only for transactions endpoints
	TotalWinAmount     *string `json:"total_win_amount,omitempty"`     // Only for transactions endpoints
}

// TransactionSummaryStats represents transaction summary statistics from ClickHouse
type TransactionSummaryStats struct {
	TotalTransactions      int             `json:"total_transactions"`
	TotalVolume            decimal.Decimal `json:"total_volume"`
	SuccessfulTransactions int             `json:"successful_transactions"`
	FailedTransactions     int             `json:"failed_transactions"`
	DepositCount           int             `json:"deposit_count"`
	WithdrawalCount        int             `json:"withdrawal_count"`
	BetCount               int             `json:"bet_count"`
	WinCount               int             `json:"win_count"`
}
