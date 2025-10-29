package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateUTMReq struct {
	Name        string `json:"name"`
	UTMSource   string `json:"utm_source"`
	UTMCampaign string `json:"utm_campaign"`
	UTMMedium   string `json:"utm_medium"`
}

type UTM struct {
	ID          uuid.UUID `json:"id"`
	UTMSource   string    `json:"utm_source"`
	UTMCampaign string    `json:"utm_campaign"`
	UTMMedium   string    `json:"utm_medium"`
}
type CreateUTMRes struct {
	Message string `json:"message"`
	UTM     UTM    `json:"utm"`
}

type FinancialMatrix struct {
	Currency                 string          `json:"currency"`
	TotalDepositAmount       decimal.Decimal `json:"total_deposit_amount"`
	TotalWithdrawalAmount    decimal.Decimal `json:"total_withdrawal_amount"`
	NumberOfDeposites        int64           `json:"number_of_deposits"`
	NumberOfWithdrawals      int64           `json:"number_of_withdrawals"`
	AverageTransactionValues decimal.Decimal `json:"average_transaction_values"`
}
type GamePattern struct {
	TotalWins        int             `json:"total_wins"`
	TotalLosses      int             `json:"total_losses"`
	AvgBetAmount     decimal.Decimal `json:"avg_bet_amount"`
	AvgPayout        decimal.Decimal `json:"avg_payout"`
	WinPercentage    decimal.Decimal `json:"win_percentage"`
	LossPercentage   decimal.Decimal `json:"loss_percentage"`
	HighestBetAmount decimal.Decimal `json:"highest_bet_amount"`
	LowestBetAmount  decimal.Decimal `json:"lowest_bet_amount"`
	MaxMultiplier    decimal.Decimal `json:"max_multiplier"`
	MinMultiplier    decimal.Decimal `json:"min_multiplier"`
	TotalPayouts     decimal.Decimal `json:"total_payouts"`
}
type GameMatricsRes struct {
	TotalBets       int32           `json:"total_bets"`
	TotalBetAmount  decimal.Decimal `json:"total_bet_amount"`
	GGR             decimal.Decimal `json:"ggr"`
	BettingPatterns GamePattern     `json:"betting_patterns"`
}
