package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PlayerTransactionHistoryRequest represents the request for player transaction history
type PlayerTransactionHistoryRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	AccountID string    `json:"account_id,omitempty"` // Optional: filter by specific GrooveTech account
	Type      string    `json:"type,omitempty"`       // Optional: filter by transaction type (wager, result, rollback, bet, sport_bet)
	Status    string    `json:"status,omitempty"`     // Optional: filter by status
	Category  string    `json:"category,omitempty"`   // Optional: filter by category (gaming, sports, general)
	Limit     int       `json:"limit,omitempty"`      // Optional: limit results (default 50, max 100)
	Offset    int       `json:"offset,omitempty"`     // Optional: offset for pagination
	StartDate string    `json:"start_date,omitempty"` // Optional: filter from date (YYYY-MM-DD)
	EndDate   string    `json:"end_date,omitempty"`   // Optional: filter to date (YYYY-MM-DD)
}

// PlayerTransactionHistoryResponse represents the response for player transaction history
type PlayerTransactionHistoryResponse struct {
	Transactions []PlayerTransaction `json:"transactions"`
	Total        int                 `json:"total"`
	Limit        int                 `json:"limit"`
	Offset       int                 `json:"offset"`
	HasMore      bool                `json:"has_more"`
}

// PlayerTransaction represents a single transaction in the history
type PlayerTransaction struct {
	ID            uuid.UUID       `json:"id"`
	TransactionID string          `json:"transaction_id"`
	AccountID     *string         `json:"account_id,omitempty"` // Only for GrooveTech transactions
	SessionID     *string         `json:"session_id,omitempty"`
	Type          string          `json:"type"`     // wager, result, rollback, jackpot, bet, sport_bet
	Category      string          `json:"category"` // gaming, sports, general
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	Metadata      *string         `json:"metadata,omitempty"` // JSON string of additional data

	// Game-specific information from metadata
	GameID   *string `json:"game_id,omitempty"`
	GameName *string `json:"game_name,omitempty"`
	RoundID  *string `json:"round_id,omitempty"`
	Provider *string `json:"provider,omitempty"`
	Device   *string `json:"device,omitempty"`

	// Sports betting specific fields
	BetReferenceNum *string          `json:"bet_reference_num,omitempty"`
	GameReference   *string          `json:"game_reference,omitempty"`
	BetMode         *string          `json:"bet_mode,omitempty"`
	Description     *string          `json:"description,omitempty"`
	PotentialWin    *decimal.Decimal `json:"potential_win,omitempty"`
	ActualWin       *decimal.Decimal `json:"actual_win,omitempty"`
	Odds            *decimal.Decimal `json:"odds,omitempty"`
	PlacedAt        *time.Time       `json:"placed_at,omitempty"`
	SettledAt       *time.Time       `json:"settled_at,omitempty"`

	// General betting fields
	ClientTransactionID *string          `json:"client_transaction_id,omitempty"`
	CashOutMultiplier   *decimal.Decimal `json:"cash_out_multiplier,omitempty"`
	Payout              *decimal.Decimal `json:"payout,omitempty"`
	HouseEdge           *decimal.Decimal `json:"house_edge,omitempty"`

	// Calculated fields
	IsWin           bool            `json:"is_win"`           // true if amount > 0 and type is result
	IsLoss          bool            `json:"is_loss"`          // true if amount < 0 and type is wager
	NetResult       decimal.Decimal `json:"net_result"`       // calculated net result for this transaction
	TransactionType string          `json:"transaction_type"` // human-readable transaction type
}

// PlayerTransactionSummary represents summary statistics for a player
type PlayerTransactionSummary struct {
	UserID           uuid.UUID       `json:"user_id"`
	TotalWagers      decimal.Decimal `json:"total_wagers"`
	TotalWins        decimal.Decimal `json:"total_wins"`
	TotalLosses      decimal.Decimal `json:"total_losses"`
	NetResult        decimal.Decimal `json:"net_result"`
	TransactionCount int             `json:"transaction_count"`
	WinCount         int             `json:"win_count"`
	LossCount        int             `json:"loss_count"`
	WinRate          float64         `json:"win_rate"` // percentage of winning transactions
	AverageBet       decimal.Decimal `json:"average_bet"`
	MaxWin           decimal.Decimal `json:"max_win"`
	MaxLoss          decimal.Decimal `json:"max_loss"`
	FirstTransaction *time.Time      `json:"first_transaction,omitempty"`
	LastTransaction  *time.Time      `json:"last_transaction,omitempty"`
}

// PlayerTransactionHistoryWithSummaryResponse includes both transactions and summary
type PlayerTransactionHistoryWithSummaryResponse struct {
	Transactions []PlayerTransaction      `json:"transactions"`
	Summary      PlayerTransactionSummary `json:"summary"`
	Total        int                      `json:"total"`
	Limit        int                      `json:"limit"`
	Offset       int                      `json:"offset"`
	HasMore      bool                     `json:"has_more"`
}
