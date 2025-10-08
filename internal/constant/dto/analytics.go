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
	DateFrom        *time.Time `json:"date_from,omitempty"`
	DateTo          *time.Time `json:"date_to,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	GameID          *string    `json:"game_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// DateRange for analytics queries
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// UserAnalytics represents user analytics data
type UserAnalytics struct {
	UserID            uuid.UUID       `json:"user_id"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	TotalBonuses      decimal.Decimal `json:"total_bonuses"`
	TotalCashback     decimal.Decimal `json:"total_cashback"`
	NetLoss           decimal.Decimal `json:"net_loss"`
	TransactionCount  uint32          `json:"transaction_count"`
	UniqueGamesPlayed uint32          `json:"unique_games_played"`
	SessionCount      uint32          `json:"session_count"`
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
	TotalPlayers  uint32          `json:"total_players"`
	TotalSessions uint32          `json:"total_sessions"`
	TotalRounds   uint32          `json:"total_rounds"`
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
	DurationSeconds *uint32         `json:"duration_seconds,omitempty"`
	TotalBets       decimal.Decimal `json:"total_bets"`
	TotalWins       decimal.Decimal `json:"total_wins"`
	NetResult       decimal.Decimal `json:"net_result"`
	BetCount        uint32          `json:"bet_count"`
	WinCount        uint32          `json:"win_count"`
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
	TotalTransactions uint32          `json:"total_transactions"`
	DepositsCount     uint32          `json:"deposits_count"`
	WithdrawalsCount  uint32          `json:"withdrawals_count"`
	BetsCount         uint32          `json:"bets_count"`
	WinsCount         uint32          `json:"wins_count"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
}

// DailyReport represents daily analytics report
type DailyReport struct {
	Date              time.Time       `json:"date"`
	TotalTransactions uint32          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
	NewUsers          uint32          `json:"new_users"`
	UniqueDepositors  uint32          `json:"unique_depositors"`
	UniqueWithdrawers uint32          `json:"unique_withdrawers"`

	// Additional metrics for comprehensive reporting
	DepositCount     uint32          `json:"deposit_count"`
	WithdrawalCount  uint32          `json:"withdrawal_count"`
	BetCount         uint32          `json:"bet_count"`
	WinCount         uint32          `json:"win_count"`
	CashbackEarned   decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections decimal.Decimal `json:"admin_corrections"`

	TopGames   []GameStats   `json:"top_games"`
	TopPlayers []PlayerStats `json:"top_players"`
}

// EnhancedDailyReport represents daily analytics report with comparison metrics
type EnhancedDailyReport struct {
	Date              time.Time       `json:"date"`
	TotalTransactions uint32          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
	NewUsers          uint32          `json:"new_users"`
	UniqueDepositors  uint32          `json:"unique_depositors"`
	UniqueWithdrawers uint32          `json:"unique_withdrawers"`

	// Additional metrics for comprehensive reporting
	DepositCount     uint32          `json:"deposit_count"`
	WithdrawalCount  uint32          `json:"withdrawal_count"`
	BetCount         uint32          `json:"bet_count"`
	WinCount         uint32          `json:"win_count"`
	CashbackEarned   decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections decimal.Decimal `json:"admin_corrections"`

	// Comparison metrics
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
	TotalTransactions uint32          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
	NewUsers          uint32          `json:"new_users"`
	UniqueDepositors  uint32          `json:"unique_depositors"`
	UniqueWithdrawers uint32          `json:"unique_withdrawers"`
	DepositCount      uint32          `json:"deposit_count"`
	WithdrawalCount   uint32          `json:"withdrawal_count"`
	BetCount          uint32          `json:"bet_count"`
	WinCount          uint32          `json:"win_count"`
	CashbackEarned    decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections  decimal.Decimal `json:"admin_corrections"`
}

// DailyReportSPLM represents Same Period Last Month metrics
type DailyReportSPLM struct {
	TotalTransactions uint32          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
	NewUsers          uint32          `json:"new_users"`
	UniqueDepositors  uint32          `json:"unique_depositors"`
	UniqueWithdrawers uint32          `json:"unique_withdrawers"`
	DepositCount      uint32          `json:"deposit_count"`
	WithdrawalCount   uint32          `json:"withdrawal_count"`
	BetCount          uint32          `json:"bet_count"`
	WinCount          uint32          `json:"win_count"`
	CashbackEarned    decimal.Decimal `json:"cashback_earned"`
	CashbackClaimed   decimal.Decimal `json:"cashback_claimed"`
	AdminCorrections  decimal.Decimal `json:"admin_corrections"`
}

// MonthlyReport represents monthly analytics report
type MonthlyReport struct {
	Year              int             `json:"year"`
	Month             int             `json:"month"`
	TotalTransactions uint32          `json:"total_transactions"`
	TotalDeposits     decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals  decimal.Decimal `json:"total_withdrawals"`
	TotalBets         decimal.Decimal `json:"total_bets"`
	TotalWins         decimal.Decimal `json:"total_wins"`
	NetRevenue        decimal.Decimal `json:"net_revenue"`
	ActiveUsers       uint32          `json:"active_users"`
	ActiveGames       uint32          `json:"active_games"`
	NewUsers          uint32          `json:"new_users"`
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
	PlayerCount  uint32          `json:"player_count"`
	SessionCount uint32          `json:"session_count"`
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
	TransactionCount  uint32          `json:"transaction_count"`
	UniqueGamesPlayed uint32          `json:"unique_games_played"`
	SessionCount      uint32          `json:"session_count"`
	AvgBetAmount      decimal.Decimal `json:"avg_bet_amount"`
	LastActivity      time.Time       `json:"last_activity"`
	Rank              int             `json:"rank"`
}

// BalanceSnapshot represents a balance snapshot at a point in time
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
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Pages    int `json:"pages"`
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
