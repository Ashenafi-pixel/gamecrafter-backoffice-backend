package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// GrooveTech API DTOs for game integration
// Based on official documentation: https://groove-docs.pages.dev/transaction-api/

// GrooveAccount represents the account information for game launch
// GET /account endpoint response
type GrooveAccount struct {
	AccountID    string          `json:"accountId"`
	SessionID    string          `json:"sessionId"`
	Balance      decimal.Decimal `json:"balance"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	LastActivity time.Time       `json:"lastActivity"`
}

// GrooveGetAccountRequest represents the request for Get Account
type GrooveGetAccountRequest struct {
	AccountID string `json:"accountId"`
	SessionID string `json:"sessionId"`
}

// GrooveGetAccountResponse represents the response for Get Account
type GrooveGetAccountResponse struct {
	Success      bool            `json:"success"`
	AccountID    string          `json:"accountId"`
	SessionID    string          `json:"sessionId"`
	Balance      decimal.Decimal `json:"balance"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	ErrorCode    string          `json:"errorCode,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// GrooveTransaction represents a transaction request/response
type GrooveTransaction struct {
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Type          string          `json:"type"` // "debit", "credit", "bet", "win"
	GameID        string          `json:"gameId,omitempty"`
	GameRoundID   string          `json:"gameRoundId,omitempty"`
	BetID         string          `json:"betId,omitempty"`
	Status        string          `json:"status"` // "pending", "completed", "failed", "cancelled"
	Balance       decimal.Decimal `json:"balance"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveTransactionRequest represents a transaction request
type GrooveTransactionRequest struct {
	AccountID   string          `json:"accountId" validate:"required"`
	SessionID   string          `json:"sessionId" validate:"required"`
	Amount      decimal.Decimal `json:"amount" validate:"required,gt=0"`
	Currency    string          `json:"currency" validate:"required"`
	Type        string          `json:"type" validate:"required,oneof=debit credit bet win"`
	GameID      string          `json:"gameId,omitempty"`
	GameRoundID string          `json:"gameRoundId,omitempty"`
	BetID       string          `json:"betId,omitempty"`
}

// GrooveTransactionResponse represents a transaction response
type GrooveTransactionResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	Balance       decimal.Decimal `json:"balance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveTech Official API DTOs based on documentation
// https://groove-docs.pages.dev/transaction-api/

// GrooveGetBalanceRequest represents the request for Get Balance
type GrooveGetBalanceRequest struct {
	AccountID string `json:"accountId"`
	SessionID string `json:"sessionId"`
}

// GrooveGetBalanceResponse represents the response for Get Balance
type GrooveGetBalanceResponse struct {
	Success      bool            `json:"success"`
	AccountID    string          `json:"accountId"`
	SessionID    string          `json:"sessionId"`
	Balance      decimal.Decimal `json:"balance"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	ErrorCode    string          `json:"errorCode,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// GrooveWagerRequest represents a wager transaction request
type GrooveWagerRequest struct {
	TransactionID string                 `json:"transactionId"`
	AccountID     string                 `json:"accountId"`
	SessionID     string                 `json:"sessionId"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	GameID        string                 `json:"gameId,omitempty"`
	RoundID       string                 `json:"roundId,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveWagerResponse represents a wager transaction response
type GrooveWagerResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	NewBalance    decimal.Decimal `json:"newBalance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveResultRequest represents a result transaction request
type GrooveResultRequest struct {
	TransactionID string                 `json:"transactionId"`
	AccountID     string                 `json:"accountId"`
	SessionID     string                 `json:"sessionId"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	GameID        string                 `json:"gameId,omitempty"`
	RoundID       string                 `json:"roundId,omitempty"`
	WinAmount     decimal.Decimal        `json:"winAmount,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveResultResponse represents a result transaction response
type GrooveResultResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	NewBalance    decimal.Decimal `json:"newBalance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveWagerAndResultRequest represents a combined wager and result request
type GrooveWagerAndResultRequest struct {
	TransactionID string                 `json:"transactionId"`
	AccountID     string                 `json:"accountId"`
	SessionID     string                 `json:"sessionId"`
	WagerAmount   decimal.Decimal        `json:"wagerAmount"`
	WinAmount     decimal.Decimal        `json:"winAmount"`
	Currency      string                 `json:"currency"`
	GameID        string                 `json:"gameId,omitempty"`
	RoundID       string                 `json:"roundId,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveWagerAndResultResponse represents a combined wager and result response
type GrooveWagerAndResultResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	WagerAmount   decimal.Decimal `json:"wagerAmount"`
	WinAmount     decimal.Decimal `json:"winAmount"`
	Currency      string          `json:"currency"`
	NewBalance    decimal.Decimal `json:"newBalance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveRollbackRequest represents a rollback transaction request
type GrooveRollbackRequest struct {
	TransactionID         string                 `json:"transactionId"`
	AccountID             string                 `json:"accountId"`
	SessionID             string                 `json:"sessionId"`
	Amount                decimal.Decimal        `json:"amount"`
	Currency              string                 `json:"currency"`
	OriginalTransactionID string                 `json:"originalTransactionId"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveRollbackResponse represents a rollback transaction response
type GrooveRollbackResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	NewBalance    decimal.Decimal `json:"newBalance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveJackpotRequest represents a jackpot transaction request
type GrooveJackpotRequest struct {
	TransactionID string                 `json:"transactionId"`
	AccountID     string                 `json:"accountId"`
	SessionID     string                 `json:"sessionId"`
	Amount        decimal.Decimal        `json:"amount"`
	Currency      string                 `json:"currency"`
	GameID        string                 `json:"gameId,omitempty"`
	RoundID       string                 `json:"roundId,omitempty"`
	JackpotType   string                 `json:"jackpotType,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveJackpotResponse represents a jackpot transaction response
type GrooveJackpotResponse struct {
	Success       bool            `json:"success"`
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	NewBalance    decimal.Decimal `json:"newBalance"`
	Status        string          `json:"status"`
	ErrorCode     string          `json:"errorCode,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
}

// GrooveBalanceRequest represents a balance check request
type GrooveBalanceRequest struct {
	AccountID string `json:"accountId" validate:"required"`
	SessionID string `json:"sessionId" validate:"required"`
}

// GrooveBalanceResponse represents a balance check response
type GrooveBalanceResponse struct {
	Success      bool            `json:"success"`
	AccountID    string          `json:"accountId"`
	Balance      decimal.Decimal `json:"balance"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	ErrorCode    string          `json:"errorCode,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// GrooveGameSession represents a game session
type GrooveGameSession struct {
	SessionID    string          `json:"sessionId"`
	AccountID    string          `json:"accountId"`
	GameID       string          `json:"gameId"`
	Balance      decimal.Decimal `json:"balance"`
	Currency     string          `json:"currency"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	ExpiresAt    time.Time       `json:"expiresAt"`
	LastActivity time.Time       `json:"lastActivity"`
}

// GrooveError represents an error response
type GrooveError struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Details      string `json:"details,omitempty"`
}

// GrooveTransactionHistory represents transaction history
type GrooveTransactionHistory struct {
	AccountID    string              `json:"accountId"`
	SessionID    string              `json:"sessionId"`
	Transactions []GrooveTransaction `json:"transactions"`
	TotalCount   int                 `json:"totalCount"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"pageSize"`
	HasMore      bool                `json:"hasMore"`
}

// GrooveTransactionHistoryRequest represents a transaction history request
type GrooveTransactionHistoryRequest struct {
	AccountID string    `json:"accountId" validate:"required"`
	SessionID string    `json:"sessionId" validate:"required"`
	FromDate  time.Time `json:"fromDate,omitempty"`
	ToDate    time.Time `json:"toDate,omitempty"`
	Page      int       `json:"page,omitempty"`
	PageSize  int       `json:"pageSize,omitempty"`
	Type      string    `json:"type,omitempty"`
}
