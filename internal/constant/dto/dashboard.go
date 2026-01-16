package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type DashboardOverviewRequest struct {
	DateFrom            time.Time `json:"date_from"`
	DateTo              time.Time `json:"date_to"`
	IsTestAccount       *bool     `json:"is_test_account,omitempty"`
	IncludeDailyBreakdown bool    `json:"include_daily_breakdown,omitempty"`
}

type DashboardOverviewSummary struct {
	TotalDeposits    decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals decimal.Decimal `json:"total_withdrawals"`
	TotalBets        decimal.Decimal `json:"total_bets"`
	TotalWins        decimal.Decimal `json:"total_wins"`
	GGR              decimal.Decimal `json:"ggr"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	NGR              decimal.Decimal `json:"ngr"`
	ActiveUsers      uint64          `json:"active_users"`
	ActiveGames      uint64          `json:"active_games"`
	TotalTransactions uint64         `json:"total_transactions"`
}

type DashboardDailyBreakdown struct {
	Date            string          `json:"date"`
	Deposits       decimal.Decimal `json:"deposits"`
	Withdrawals    decimal.Decimal `json:"withdrawals"`
	Bets           decimal.Decimal `json:"bets"`
	Wins           decimal.Decimal `json:"wins"`
	GGR            decimal.Decimal `json:"ggr"`
	CashbackClaimed decimal.Decimal `json:"cashback_claimed"`
	NGR            decimal.Decimal `json:"ngr"`
	ActiveUsers    uint64          `json:"active_users"`
	ActiveGames    uint64          `json:"active_games"`
}

type DashboardOverviewResponse struct {
	DateRange      DateRange                `json:"date_range"`
	Summary        DashboardOverviewSummary `json:"summary"`
	DailyBreakdown []DashboardDailyBreakdown `json:"daily_breakdown,omitempty"`
}

type PerformanceSummaryRequest struct {
	Range         string     `json:"range,omitempty"`
	DateFrom      *time.Time `json:"date_from,omitempty"`
	DateTo        *time.Time `json:"date_to,omitempty"`
	IsTestAccount *bool      `json:"is_test_account,omitempty"`
}

type PerformanceSummaryFinancialOverview struct {
	GGR              decimal.Decimal `json:"ggr"`
	NGR              decimal.Decimal `json:"ngr"`
	TotalDeposits    decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals decimal.Decimal `json:"total_withdrawals"`
	CashbackClaimed  decimal.Decimal `json:"cashback_claimed"`
	NetDeposits      decimal.Decimal `json:"net_deposits"`
}

type PerformanceSummaryBettingMetrics struct {
	TotalBets    decimal.Decimal `json:"total_bets"`
	TotalWins    decimal.Decimal `json:"total_wins"`
	BetCount     uint64          `json:"bet_count"`
	WinCount     uint64          `json:"win_count"`
	AvgBetAmount decimal.Decimal `json:"avg_bet_amount"`
	RTP          decimal.Decimal `json:"rtp"`
}

type PerformanceSummaryUserActivity struct {
	ActiveUsers          uint64          `json:"active_users"`
	NewUsers             uint64          `json:"new_users"`
	UniqueDepositors     uint64          `json:"unique_depositors"`
	UniqueWithdrawers    uint64          `json:"unique_withdrawers"`
	AvgDepositPerUser    decimal.Decimal `json:"avg_deposit_per_user"`
	AvgWithdrawalPerUser  decimal.Decimal `json:"avg_withdrawal_per_user"`
}

type PerformanceSummaryTransactionVolume struct {
	TotalTransactions      uint64 `json:"total_transactions"`
	DepositCount           uint64 `json:"deposit_count"`
	WithdrawalCount        uint64 `json:"withdrawal_count"`
	BetCount               uint64 `json:"bet_count"`
	WinCount               uint64 `json:"win_count"`
	CashbackEarnedCount    uint64 `json:"cashback_earned_count"`
	CashbackClaimedCount   uint64 `json:"cashback_claimed_count"`
}

type PerformanceSummaryDailyTrend struct {
	Date         string          `json:"date"`
	GGR          decimal.Decimal `json:"ggr"`
	NGR          decimal.Decimal `json:"ngr"`
	Deposits     decimal.Decimal `json:"deposits"`
	Withdrawals  decimal.Decimal `json:"withdrawals"`
	Bets         decimal.Decimal `json:"bets"`
	Wins         decimal.Decimal `json:"wins"`
	ActiveUsers  uint64          `json:"active_users"`
}

type PerformanceSummaryResponse struct {
	RangeType          string                           `json:"range_type"` 
	DateRange          DateRange                        `json:"date_range"`
	FinancialOverview  PerformanceSummaryFinancialOverview `json:"financial_overview"`
	BettingMetrics     PerformanceSummaryBettingMetrics    `json:"betting_metrics"`
	UserActivity       PerformanceSummaryUserActivity      `json:"user_activity"`
	TransactionVolume  PerformanceSummaryTransactionVolume  `json:"transaction_volume"`
	DailyTrends        []PerformanceSummaryDailyTrend       `json:"daily_trends"`
}

type TimeSeriesRequest struct {
	DateFrom      time.Time `json:"date_from"`
	DateTo        time.Time `json:"date_to"`
	Granularity   string    `json:"granularity,omitempty"`
	IsTestAccount *bool     `json:"is_test_account,omitempty"`
	Metrics       string    `json:"metrics,omitempty"`
}

type TimeSeriesRevenueTrend struct {
	Timestamp time.Time       `json:"timestamp"`
	GGR       decimal.Decimal `json:"ggr"`
	NGR       decimal.Decimal `json:"ngr"`
}

type TimeSeriesUserActivity struct {
	Timestamp         time.Time `json:"timestamp"`
	ActiveUsers       uint64    `json:"active_users"`
	NewUsers          uint64    `json:"new_users"`
	UniqueDepositors  uint64    `json:"unique_depositors"`
	UniqueWithdrawers uint64    `json:"unique_withdrawers"`
}

type TimeSeriesTransactionVolume struct {
	Timestamp        time.Time `json:"timestamp"`
	TotalTransactions uint64   `json:"total_transactions"`
	Deposits         uint64    `json:"deposits"`
	Withdrawals      uint64    `json:"withdrawals"`
	Bets             uint64    `json:"bets"`
	Wins             uint64    `json:"wins"`
}

type TimeSeriesDepositsVsWithdrawals struct {
	Timestamp   time.Time       `json:"timestamp"`
	Deposits    decimal.Decimal `json:"deposits"`
	Withdrawals decimal.Decimal `json:"withdrawals"`
}

type TimeSeriesResponse struct {
	Granularity          string                              `json:"granularity"`
	DateRange            DateRange                           `json:"date_range"`
	RevenueTrend         []TimeSeriesRevenueTrend             `json:"revenue_trend,omitempty"`
	UserActivity         []TimeSeriesUserActivity             `json:"user_activity,omitempty"`
	TransactionVolume    []TimeSeriesTransactionVolume        `json:"transaction_volume,omitempty"`
	DepositsVsWithdrawals []TimeSeriesDepositsVsWithdrawals   `json:"deposits_vs_withdrawals,omitempty"`
}

