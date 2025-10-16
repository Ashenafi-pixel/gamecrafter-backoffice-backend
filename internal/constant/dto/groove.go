package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GrooveTech API DTOs for game integration
// Based on official documentation: https://groove-docs.pages.dev/transaction-api/

// GameInfo represents basic game information
type GameInfo struct {
	GameID             string `json:"game_id"`
	GameName           string `json:"game_name"`
	InternalName       string `json:"internal_name"`
	Provider           string `json:"provider"`
	IntegrationPartner string `json:"integration_partner"`
}

// WinnerNotificationData represents the data structure for winner notifications
type WinnerNotificationData struct {
	Username      string          `json:"username"`
	Email         string          `json:"email"`
	GameName      string          `json:"game_name"`
	GameID        string          `json:"game_id"`
	BetAmount     decimal.Decimal `json:"bet_amount"`
	WinAmount     decimal.Decimal `json:"win_amount"`
	NetWinnings   decimal.Decimal `json:"net_winnings"`
	Currency      string          `json:"currency"`
	Timestamp     time.Time       `json:"timestamp"`
	SessionID     string          `json:"session_id"`
	RoundID       string          `json:"round_id"`
	TransactionID string          `json:"transaction_id"`
}

// GameSession represents a game session for tracking
type GameSession struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	SessionID            string    `json:"session_id"`
	GameID               string    `json:"game_id"`
	DeviceType           string    `json:"device_type"`
	GameMode             string    `json:"game_mode"`
	GrooveURL            string    `json:"groove_url"`
	HomeURL              string    `json:"home_url"`
	ExitURL              string    `json:"exit_url"`
	HistoryURL           string    `json:"history_url"`
	LicenseType          string    `json:"license_type"`
	IsTestAccount        bool      `json:"is_test_account"`
	RealityCheckElapsed  int       `json:"reality_check_elapsed"`
	RealityCheckInterval int       `json:"reality_check_interval"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at"`
	IsActive             bool      `json:"is_active"`
	LastActivity         time.Time `json:"last_activity"`
}

// LaunchGameRequest represents the request to launch a game
type LaunchGameRequest struct {
	GameID     string `json:"game_id" validate:"required"`
	DeviceType string `json:"device_type" validate:"required,oneof=desktop mobile"`
	GameMode   string `json:"game_mode" validate:"required,oneof=demo real"`
	// CMA Compliance fields
	Country              string `json:"country,omitempty"`                // ISO 3166-1 alpha-2 code
	Currency             string `json:"currency,omitempty"`               // ISO 4217 currency code
	Language             string `json:"language,omitempty"`               // ISO format language
	IsTestAccount        *bool  `json:"is_test_account,omitempty"`        // Test account flag
	RealityCheckElapsed  int    `json:"reality_check_elapsed,omitempty"`  // Minutes elapsed since session start
	RealityCheckInterval int    `json:"reality_check_interval,omitempty"` // Minutes between reality checks
}

// LaunchGameResponse represents the response for game launch
type LaunchGameResponse struct {
	Success   bool   `json:"success"`
	GameURL   string `json:"game_url"`
	SessionID string `json:"session_id"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// GrooveAccount represents the account information for game launch
// GET /account endpoint response
type GrooveAccount struct {
	ID           string          `json:"id"`
	UserID       string          `json:"userId"`
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

// GrooveTransaction represents a stored transaction for idempotency
type GrooveTransaction struct {
	TransactionID        string          `json:"transaction_id"`
	AccountTransactionID string          `json:"account_transaction_id"`
	AccountID            string          `json:"account_id"`
	GameSessionID        string          `json:"game_session_id"`
	RoundID              string          `json:"round_id"`
	GameID               string          `json:"game_id"`
	BetAmount            decimal.Decimal `json:"bet_amount"`
	Device               string          `json:"device"`
	FRBID                string          `json:"frbid,omitempty"`
	UserID               uuid.UUID       `json:"user_id"`
	Status               string          `json:"status,omitempty"`
	BalanceBefore        decimal.Decimal `json:"balance_before"`
	BalanceAfter         decimal.Decimal `json:"balance_after"`
	CreatedAt            time.Time       `json:"created_at"`
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

// GrooveGetBalanceRequest represents the request for Get Balance (GrooveTech spec)
type GrooveGetBalanceRequest struct {
	AccountID  string `json:"accountid" validate:"required"`
	SessionID  string `json:"gamesessionid" validate:"required"`
	Device     string `json:"device" validate:"required,oneof=desktop mobile"`
	GameID     string `json:"nogsgameid" validate:"required"`
	APIVersion string `json:"apiversion" validate:"required"`
	Request    string `json:"request" validate:"required"`
}

// GrooveGetBalanceResponse represents the response for Get Balance (GrooveTech spec)
type GrooveGetBalanceResponse struct {
	Code         int             `json:"code"`
	Status       string          `json:"status"`
	Balance      decimal.Decimal `json:"balance"`
	BonusBalance decimal.Decimal `json:"bonus_balance"`
	RealBalance  decimal.Decimal `json:"real_balance"`
	GameMode     int             `json:"game_mode,omitempty"`
	Order        string          `json:"order,omitempty"`
	APIVersion   string          `json:"apiversion"`
	Message      string          `json:"message,omitempty"`
}

// GrooveWagerRequest represents a wager transaction request according to GrooveTech spec
type GrooveWagerRequest struct {
	AccountID     string          `json:"accountid"`
	GameSessionID string          `json:"gamesessionid"`
	TransactionID string          `json:"transactionid"`
	RoundID       string          `json:"roundid"`
	GameID        string          `json:"gameid"`
	BetAmount     decimal.Decimal `json:"betamount"`
	Device        string          `json:"device"`
	FRBID         string          `json:"frbid,omitempty"`   // Optional Free Round Bonus ID
	UserID        uuid.UUID       `json:"user_id,omitempty"` // Internal field
}

// GrooveWagerResponse represents a wager transaction response according to GrooveTech spec
type GrooveWagerResponse struct {
	Code                 int             `json:"code"`
	Status               string          `json:"status"`
	AccountTransactionID string          `json:"accounttransactionid"`
	Balance              decimal.Decimal `json:"balance"`
	BonusMoneyBet        decimal.Decimal `json:"bonusmoneybet"`
	RealMoneyBet         decimal.Decimal `json:"realmoneybet"`
	BonusBalance         decimal.Decimal `json:"bonus_balance"`
	RealBalance          decimal.Decimal `json:"real_balance"`
	GameMode             int             `json:"game_mode"`
	Order                string          `json:"order"`
	APIVersion           string          `json:"apiversion"`
	Message              string          `json:"message,omitempty"`
}

// GrooveResultRequest represents a result transaction request (GrooveTech Official API)
type GrooveResultRequest struct {
	Request       string          `json:"request"`         // Must be "result"
	AccountID     string          `json:"accountid"`       // Player's unique identifier
	APIVersion    string          `json:"apiversion"`      // API version (e.g., "1.2")
	Device        string          `json:"device"`          // Device type: "desktop" or "mobile"
	GameID        string          `json:"gameid"`          // Groove game ID
	GameSessionID string          `json:"gamesessionid"`   // Game session ID
	SessionID     string          `json:"sessionid"`       // Session ID (alias for GameSessionID)
	GameStatus    string          `json:"gamestatus"`      // Game status: "completed" or "pending"
	Result        decimal.Decimal `json:"result"`          // Win amount (0 if player lost)
	Amount        decimal.Decimal `json:"amount"`          // Amount (alias for Result)
	RoundID       string          `json:"roundid"`         // Round identifier
	TransactionID string          `json:"transactionid"`   // Unique transaction identifier
	FRBID         string          `json:"frbid,omitempty"` // Free Round Bonus ID (optional)
}

// GrooveResultResponse represents a result transaction response (GrooveTech Official API)
type GrooveResultResponse struct {
	Code          int             `json:"code"`                    // Response code (200 for success)
	Status        string          `json:"status"`                  // Response status ("Success")
	Success       bool            `json:"success"`                 // Success flag
	TransactionID string          `json:"transactionid"`           // Transaction ID
	AccountID     string          `json:"accountid"`               // Account ID
	SessionID     string          `json:"sessionid"`               // Session ID
	Amount        decimal.Decimal `json:"amount"`                  // Amount processed
	WalletTx      string          `json:"walletTx"`                // Casino's wallet transaction ID
	Balance       decimal.Decimal `json:"balance"`                 // Total player balance (real + bonus)
	BonusWin      decimal.Decimal `json:"bonusWin"`                // Portion of win allocated to bonus funds
	RealMoneyWin  decimal.Decimal `json:"realMoneyWin"`            // Portion of win allocated to real money
	BonusBalance  decimal.Decimal `json:"bonus_balance"`           // Player's bonus balance
	RealBalance   decimal.Decimal `json:"real_balance"`            // Player's real money balance
	GameMode      int             `json:"game_mode"`               // Game mode: 1 - Real money, 2 - Bonus mode
	Order         string          `json:"order"`                   // Order type: "cash_money" or "bonus_money"
	APIVersion    string          `json:"apiversion"`              // API version
	ErrorCode     string          `json:"error_code,omitempty"`    // Error code if failed
	ErrorMessage  string          `json:"error_message,omitempty"` // Error message if failed
}

// GrooveWagerAndResultRequest represents a combined wager and result request
type GrooveWagerAndResultRequest struct {
	TransactionID string          `json:"transactionId"`
	AccountID     string          `json:"accountId"`
	SessionID     string          `json:"sessionId"`
	Device        string          `json:"device"`
	GameID        string          `json:"gameId"`
	APIVersion    string          `json:"apiversion"`
	BetAmount     decimal.Decimal `json:"betamount"`
	WinAmount     decimal.Decimal `json:"result"`
	RoundID       string          `json:"roundid"`
	GameStatus    string          `json:"gamestatus"`
	FRBID         string          `json:"frbid,omitempty"`
}

// GrooveWagerAndResultResponse represents a combined wager and result response
type GrooveWagerAndResultResponse struct {
	Code          int             `json:"code"`                    // Response code (200 for success)
	Status        string          `json:"status"`                  // Response status ("Success")
	Success       bool            `json:"success"`                 // Success flag
	TransactionID string          `json:"transactionid"`           // Transaction ID
	AccountID     string          `json:"accountid"`               // Account ID
	SessionID     string          `json:"sessionid"`               // Session ID
	RoundID       string          `json:"roundid"`                 // Round ID
	GameStatus    string          `json:"gamestatus"`              // Game status
	WalletTx      string          `json:"walletTx"`                // Casino's wallet transaction ID
	Balance       decimal.Decimal `json:"balance"`                 // Total player balance (real + bonus)
	BonusWin      decimal.Decimal `json:"bonusWin"`                // Portion of win allocated to bonus funds
	RealMoneyWin  decimal.Decimal `json:"realMoneyWin"`            // Portion of win allocated to real money
	BonusMoneyBet decimal.Decimal `json:"bonusmoneybet"`           // Portion of bet from bonus funds
	RealMoneyBet  decimal.Decimal `json:"realmoneybet"`            // Portion of bet from real money
	BonusBalance  decimal.Decimal `json:"bonus_balance"`           // Player's bonus balance
	RealBalance   decimal.Decimal `json:"real_balance"`            // Player's real money balance
	GameMode      int             `json:"game_mode"`               // Game mode: 1 - Real money, 2 - Bonus mode
	Order         string          `json:"order"`                   // Order type: "cash_money" or "bonus_money"
	APIVersion    string          `json:"apiversion"`              // API version
	ErrorCode     string          `json:"error_code,omitempty"`    // Error code if failed
	ErrorMessage  string          `json:"error_message,omitempty"` // Error message if failed
}

// GrooveRollbackRequest represents a rollback transaction request (Legacy)
type GrooveRollbackRequest struct {
	TransactionID         string                 `json:"transactionId"`
	AccountID             string                 `json:"accountId"`
	SessionID             string                 `json:"sessionId"`
	Amount                decimal.Decimal        `json:"amount"`
	Currency              string                 `json:"currency"`
	OriginalTransactionID string                 `json:"originalTransactionId"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// GrooveRollbackResponse represents a rollback transaction response (Legacy)
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

// GrooveJackpotRequest represents a jackpot transaction request (Legacy)
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

// GrooveJackpotResponse represents a jackpot transaction response (Legacy)
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

// GrooveUserProfile represents user profile information for GrooveTech API
type GrooveUserProfile struct {
	City     string `json:"city"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
}

// GrooveRollbackRequestOfficial represents the request for rollback transaction (Official GrooveTech API)
type GrooveRollbackRequestOfficial struct {
	AccountID      string          `json:"accountid" validate:"required"`
	APIVersion     string          `json:"apiversion" validate:"required"`
	Device         string          `json:"device" validate:"required,oneof=desktop mobile"`
	GameID         string          `json:"gameid" validate:"required"`
	GameSessionID  string          `json:"gamesessionid" validate:"required"`
	Request        string          `json:"request" validate:"required,eq=rollback"`
	TransactionID  string          `json:"transactionid" validate:"required"`
	RollbackAmount decimal.Decimal `json:"rollbackamount,omitempty"`
	RoundID        string          `json:"roundid,omitempty"`
}

// GrooveRollbackResponseOfficial represents the response for rollback transaction (Official GrooveTech API)
type GrooveRollbackResponseOfficial struct {
	Code                 int             `json:"code"`
	Status               string          `json:"status"`
	AccountTransactionID string          `json:"accounttransactionid"`
	Balance              decimal.Decimal `json:"balance"`
	BonusBalance         decimal.Decimal `json:"bonus_balance"`
	RealBalance          decimal.Decimal `json:"real_balance"`
	GameMode             int             `json:"game_mode"`
	Order                string          `json:"order"`
	APIVersion           string          `json:"apiversion"`
}

// GrooveJackpotRequestOfficial represents the request for jackpot transaction (Official GrooveTech API)
type GrooveJackpotRequestOfficial struct {
	AccountID     string          `json:"accountid" validate:"required"`
	Amount        decimal.Decimal `json:"amount" validate:"required"`
	APIVersion    string          `json:"apiversion" validate:"required"`
	GameID        string          `json:"gameid" validate:"required"`
	GameSessionID string          `json:"gamesessionid" validate:"required"`
	GameStatus    string          `json:"gamestatus" validate:"required,oneof=completed pending"`
	Request       string          `json:"request" validate:"required,eq=jackpot"`
	RoundID       string          `json:"roundid" validate:"required"`
	TransactionID string          `json:"transactionid" validate:"required"`
}

// GrooveJackpotResponseOfficial represents the response for jackpot transaction (Official GrooveTech API)
type GrooveJackpotResponseOfficial struct {
	Code         int             `json:"code"`
	Status       string          `json:"status"`
	WalletTx     string          `json:"walletTx"`
	Balance      decimal.Decimal `json:"balance"`
	BonusWin     decimal.Decimal `json:"bonusWin"`
	RealMoneyWin decimal.Decimal `json:"realMoneyWin"`
	BonusBalance decimal.Decimal `json:"bonus_balance"`
	RealBalance  decimal.Decimal `json:"real_balance"`
	GameMode     int             `json:"game_mode"`
	Order        string          `json:"order"`
	APIVersion   string          `json:"apiversion"`
}

// GrooveRollbackOnResultRequest represents the request for rollback on result (reversewin)
type GrooveRollbackOnResultRequest struct {
	AccountID        string          `json:"accountid" validate:"required"`
	Amount           decimal.Decimal `json:"amount" validate:"required"`
	APIVersion       string          `json:"apiversion" validate:"required"`
	Device           string          `json:"device" validate:"required,oneof=desktop mobile"`
	GameID           string          `json:"gameid" validate:"required"`
	GameSessionID    string          `json:"gamesessionid" validate:"required"`
	Request          string          `json:"request" validate:"required,eq=reversewin"`
	RoundID          string          `json:"roundid" validate:"required"`
	TransactionID    string          `json:"transactionid" validate:"required"`
	WinTransactionID string          `json:"wintransactionid"` // Optional parameter for signature validation
}

// GrooveRollbackOnResultResponse represents the response for rollback on result
type GrooveRollbackOnResultResponse struct {
	Code                 int             `json:"code"`
	Status               string          `json:"status"`
	AccountTransactionID string          `json:"accounttransactionid"`
	Balance              decimal.Decimal `json:"balance"`
	BonusBalance         decimal.Decimal `json:"bonus_balance"`
	RealBalance          decimal.Decimal `json:"real_balance"`
	GameMode             int             `json:"game_mode"`
	Order                string          `json:"order"`
	APIVersion           string          `json:"apiversion"`
}

// GrooveRollbackOnRollbackRequest represents the request for rollback on rollback (rollbackrollback)
type GrooveRollbackOnRollbackRequest struct {
	AccountID      string          `json:"accountid" validate:"required"`
	RollbackAmount decimal.Decimal `json:"rollbackAmount" validate:"required"`
	APIVersion     string          `json:"apiversion" validate:"required"`
	Device         string          `json:"device" validate:"required,oneof=desktop mobile"`
	GameID         string          `json:"gameid" validate:"required"`
	GameSessionID  string          `json:"gamesessionid" validate:"required"`
	Request        string          `json:"request" validate:"required,eq=rollbackrollback"`
	RoundID        string          `json:"roundid" validate:"required"`
	TransactionID  string          `json:"transactionid" validate:"required"`
}

// GrooveRollbackOnRollbackResponse represents the response for rollback on rollback
type GrooveRollbackOnRollbackResponse struct {
	Code                 int             `json:"code"`
	Status               string          `json:"status"`
	AccountTransactionID string          `json:"accounttransactionid"`
	Balance              decimal.Decimal `json:"balance"`
	BonusBalance         decimal.Decimal `json:"bonus_balance"`
	RealBalance          decimal.Decimal `json:"real_balance"`
	GameMode             int             `json:"game_mode"`
	Order                string          `json:"order"`
	APIVersion           string          `json:"apiversion"`
}

// GrooveWagerByBatchRequest represents the request for wager by batch (sportsbook)
type GrooveWagerByBatchRequest struct {
	AccountID     string           `json:"account_id" validate:"required"`
	GameID        string           `json:"game_id" validate:"required"`
	GameSessionID string           `json:"game_session_id" validate:"required"`
	Device        string           `json:"device" validate:"required"`
	Bets          []GrooveBatchBet `json:"bets" validate:"required,min=1"`
}

// GrooveBatchBet represents a single bet within a batch
type GrooveBatchBet struct {
	FRBID         string          `json:"frb_id,omitempty"`
	Amount        decimal.Decimal `json:"amount" validate:"required"`
	RoundID       string          `json:"round_id" validate:"required"`
	TransactionID string          `json:"transaction_id" validate:"required"`
}

// GrooveWagerByBatchResponse represents the response for wager by batch
type GrooveWagerByBatchResponse struct {
	Status       string                 `json:"status"`
	Code         int                    `json:"code"`
	Message      string                 `json:"message"`
	Bets         []GrooveBatchBetResult `json:"bets"`
	Balance      decimal.Decimal        `json:"balance"`
	RealBalance  decimal.Decimal        `json:"real_balance"`
	BonusBalance decimal.Decimal        `json:"bonus_balance"`
}

// GrooveBatchBetResult represents the result of a single bet within a batch
type GrooveBatchBetResult struct {
	ProviderTransactionID string          `json:"provider_transaction_id"`
	TransactionID         string          `json:"transaction_id"`
	BonusMoneyBet         decimal.Decimal `json:"bonus_money_bet"`
	RealMoneyBet          decimal.Decimal `json:"real_money_bet"`
}
