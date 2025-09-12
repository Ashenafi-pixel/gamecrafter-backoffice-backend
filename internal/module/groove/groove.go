package groove

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/groove"
	"go.uber.org/zap"
)

type GrooveService interface {
	// Account operations
	GetAccount(ctx context.Context, sessionID string) (*dto.GrooveAccount, error)
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error)
	CreateAccount(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error)

	// Official GrooveTech Transaction API methods
	GetBalance(ctx context.Context, accountID string) (*dto.GrooveGetBalanceResponse, error)
	ProcessWager(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error)
	ProcessResult(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error)
	ProcessWagerAndResult(ctx context.Context, req dto.GrooveWagerAndResultRequest) (*dto.GrooveWagerAndResultResponse, error)
	ProcessRollback(ctx context.Context, req dto.GrooveRollbackRequest) (*dto.GrooveRollbackResponse, error)
	ProcessJackpot(ctx context.Context, req dto.GrooveJackpotRequest) (*dto.GrooveJackpotResponse, error)

	// Legacy transaction operations (for backward compatibility)
	ProcessDebit(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error)
	ProcessCredit(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error)
	ProcessBet(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error)
	ProcessWin(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error)

	// Transaction history
	GetTransactionHistory(ctx context.Context, req dto.GrooveTransactionHistoryRequest) (*dto.GrooveTransactionHistory, error)

	// Game session management
	CreateGameSession(ctx context.Context, accountID, gameID string) (*dto.GrooveGameSession, error)
	EndGameSession(ctx context.Context, sessionID string) error

	// Game launch functionality
	LaunchGame(ctx context.Context, userID uuid.UUID, req dto.LaunchGameRequest) (*dto.LaunchGameResponse, error)
	ValidateGameSession(ctx context.Context, sessionID string) (*dto.GameSession, error)

	// User profile operations
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.GrooveUserProfile, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
}

type GrooveServiceImpl struct {
	storage            groove.GrooveStorage
	gameSessionStorage groove.GameSessionStorage
	logger             *zap.Logger
}

func NewGrooveService(storage groove.GrooveStorage, gameSessionStorage groove.GameSessionStorage, logger *zap.Logger) GrooveService {
	return &GrooveServiceImpl{
		storage:            storage,
		gameSessionStorage: gameSessionStorage,
		logger:             logger,
	}
}

// GetAccount retrieves account information for game launch
func (s *GrooveServiceImpl) GetAccount(ctx context.Context, sessionID string) (*dto.GrooveAccount, error) {
	s.logger.Info("Getting GrooveTech account", zap.String("session_id", sessionID))

	// Extract user ID from session ID (access token)
	userID, err := s.extractUserIDFromSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to extract user ID from session", zap.Error(err))
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Get or create account
	account, err := s.storage.GetAccountByUserID(ctx, userID)
	if err != nil {
		// Create account if it doesn't exist
		account, err = s.CreateAccount(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to create account", zap.Error(err))
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
	}

	// Update session ID
	account.SessionID = sessionID
	account.LastActivity = time.Now()

	// Update account in storage
	updatedAccount, err := s.storage.UpdateAccount(ctx, *account)
	if err != nil {
		s.logger.Error("Failed to update account", zap.Error(err))
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	s.logger.Info("Account retrieved successfully",
		zap.String("account_id", updatedAccount.AccountID),
		zap.String("balance", updatedAccount.Balance.String()))

	return updatedAccount, nil
}

// GetAccountByUserID retrieves account by user ID
func (s *GrooveServiceImpl) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error) {
	s.logger.Info("Getting GrooveTech account by user ID", zap.String("user_id", userID.String()))

	account, err := s.storage.GetAccountByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get account by user ID", zap.Error(err))
		return nil, fmt.Errorf("account not found: %w", err)
	}

	s.logger.Info("Account retrieved successfully by user ID",
		zap.String("account_id", account.AccountID),
		zap.String("user_id", userID.String()))

	return account, nil
}

// CreateAccount creates a new GrooveTech account for a user
func (s *GrooveServiceImpl) CreateAccount(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error) {
	s.logger.Info("Creating GrooveTech account", zap.String("user_id", userID.String()))

	// Get user balance from existing system
	balance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	account := &dto.GrooveAccount{
		AccountID:    uuid.New().String(),
		SessionID:    "", // Will be set when user logs in
		Balance:      balance,
		Currency:     "USD", // Default currency
		Status:       "active",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	createdAccount, err := s.storage.CreateAccount(ctx, *account, userID)
	if err != nil {
		s.logger.Error("Failed to create account", zap.Error(err))
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Info("Account created successfully",
		zap.String("account_id", createdAccount.AccountID),
		zap.String("user_id", userID.String()))

	return createdAccount, nil
}

// ProcessDebit processes a debit transaction (betting)
func (s *GrooveServiceImpl) ProcessDebit(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing debit transaction",
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Generate transaction ID
	transactionID := uuid.New().String()

	// Validate transaction
	if err := s.validateTransaction(req); err != nil {
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "INVALID_TRANSACTION",
			ErrorMessage:  err.Error(),
		}, nil
	}

	// Check if account has sufficient balance
	balance, err := s.storage.GetAccountBalance(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account balance", zap.Error(err))
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "ACCOUNT_NOT_FOUND",
			ErrorMessage:  "Account not found",
		}, nil
	}

	if balance.LessThan(req.Amount) {
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "INSUFFICIENT_FUNDS",
			ErrorMessage:  "Insufficient funds",
		}, nil
	}

	// Process the debit
	transaction := dto.GrooveTransaction{
		TransactionID: transactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "debit",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process debit transaction", zap.Error(err))
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "TRANSACTION_FAILED",
			ErrorMessage:  "Transaction processing failed",
		}, nil
	}

	s.logger.Info("Debit transaction processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return response, nil
}

// ProcessCredit processes a credit transaction (winnings)
func (s *GrooveServiceImpl) ProcessCredit(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing credit transaction",
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Generate transaction ID
	transactionID := uuid.New().String()

	// Validate transaction
	if err := s.validateTransaction(req); err != nil {
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "INVALID_TRANSACTION",
			ErrorMessage:  err.Error(),
		}, nil
	}

	// Process the credit
	transaction := dto.GrooveTransaction{
		TransactionID: transactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "credit",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process credit transaction", zap.Error(err))
		return &dto.GrooveTransactionResponse{
			Success:       false,
			TransactionID: transactionID,
			ErrorCode:     "TRANSACTION_FAILED",
			ErrorMessage:  "Transaction processing failed",
		}, nil
	}

	s.logger.Info("Credit transaction processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return response, nil
}

// ProcessBet processes a bet transaction (combination of debit and potential credit)
func (s *GrooveServiceImpl) ProcessBet(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing bet transaction",
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Process as debit first
	debitReq := req
	debitReq.Type = "debit"

	response, err := s.ProcessDebit(ctx, debitReq)
	if err != nil || !response.Success {
		return response, err
	}

	// Note: Type field not available in GrooveTransactionResponse

	s.logger.Info("Bet transaction processed successfully",
		zap.String("transaction_id", response.TransactionID))

	return response, nil
}

// ProcessWin processes a win transaction (credit for winnings)
func (s *GrooveServiceImpl) ProcessWin(ctx context.Context, req dto.GrooveTransactionRequest) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing win transaction",
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Process as credit
	creditReq := req
	creditReq.Type = "credit"

	response, err := s.ProcessCredit(ctx, creditReq)
	if err != nil || !response.Success {
		return response, err
	}

	// Note: Type field not available in GrooveTransactionResponse

	s.logger.Info("Win transaction processed successfully",
		zap.String("transaction_id", response.TransactionID))

	return response, nil
}

// GetBalance retrieves account balance (Official GrooveTech API)
func (s *GrooveServiceImpl) GetBalance(ctx context.Context, accountID string) (*dto.GrooveGetBalanceResponse, error) {
	s.logger.Info("Getting account balance", zap.String("account_id", accountID))

	balance, err := s.storage.GetAccountBalance(ctx, accountID)
	if err != nil {
		s.logger.Error("Failed to get account balance", zap.Error(err))
		return &dto.GrooveGetBalanceResponse{
			Code:       1,
			Status:     "Technical error",
			Message:    "Account not found",
			APIVersion: "1.2",
		}, nil
	}

	response := &dto.GrooveGetBalanceResponse{
		Code:         200,
		Status:       "Success",
		Balance:      balance,
		BonusBalance: decimal.Zero, // No bonus balance for now
		RealBalance:  balance,
		GameMode:     1, // Real money mode
		Order:        "cash_money",
		APIVersion:   "1.2",
	}

	s.logger.Info("Balance retrieved successfully",
		zap.String("account_id", accountID),
		zap.String("balance", balance.String()))

	return response, nil
}

// ProcessWager processes a wager transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessWager(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error) {
	s.logger.Info("Processing wager",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Check if account has sufficient balance
	balance, err := s.storage.GetAccountBalance(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account balance", zap.Error(err))
		return &dto.GrooveWagerResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "ACCOUNT_NOT_FOUND",
			ErrorMessage:  "Account not found",
		}, nil
	}

	if balance.LessThan(req.Amount) {
		return &dto.GrooveWagerResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "INSUFFICIENT_FUNDS",
			ErrorMessage:  "Insufficient funds",
		}, nil
	}

	// Process the wager (debit)
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "wager",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process wager", zap.Error(err))
		return &dto.GrooveWagerResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "WAGER_FAILED",
			ErrorMessage:  "Failed to process wager",
		}, nil
	}

	s.logger.Info("Wager processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return &dto.GrooveWagerResponse{
		Success:       true,
		TransactionID: response.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		NewBalance:    response.Balance,
		Status:        "completed",
	}, nil
}

// ProcessResult processes a result transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessResult(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error) {
	s.logger.Info("Processing result",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Process the result (credit)
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "result",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process result", zap.Error(err))
		return &dto.GrooveResultResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "RESULT_FAILED",
			ErrorMessage:  "Failed to process result",
		}, nil
	}

	s.logger.Info("Result processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return &dto.GrooveResultResponse{
		Success:       true,
		TransactionID: response.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		NewBalance:    response.Balance,
		Status:        "completed",
	}, nil
}

// ProcessWagerAndResult processes a combined wager and result transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessWagerAndResult(ctx context.Context, req dto.GrooveWagerAndResultRequest) (*dto.GrooveWagerAndResultResponse, error) {
	s.logger.Info("Processing wager and result",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("wager_amount", req.WagerAmount.String()),
		zap.String("win_amount", req.WinAmount.String()))

	// Check if account has sufficient balance for wager
	balance, err := s.storage.GetAccountBalance(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account balance", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "ACCOUNT_NOT_FOUND",
			ErrorMessage:  "Account not found",
		}, nil
	}

	if balance.LessThan(req.WagerAmount) {
		return &dto.GrooveWagerAndResultResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "INSUFFICIENT_FUNDS",
			ErrorMessage:  "Insufficient funds for wager",
		}, nil
	}

	// Calculate net result (win amount - wager amount)
	netResult := req.WinAmount.Sub(req.WagerAmount)

	// Process the net result
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        netResult,
		Currency:      req.Currency,
		Type:          "wager_and_result",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process wager and result", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "WAGER_RESULT_FAILED",
			ErrorMessage:  "Failed to process wager and result",
		}, nil
	}

	s.logger.Info("Wager and result processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return &dto.GrooveWagerAndResultResponse{
		Success:       true,
		TransactionID: response.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		WagerAmount:   req.WagerAmount,
		WinAmount:     req.WinAmount,
		Currency:      req.Currency,
		NewBalance:    response.Balance,
		Status:        "completed",
	}, nil
}

// ProcessRollback processes a rollback transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessRollback(ctx context.Context, req dto.GrooveRollbackRequest) (*dto.GrooveRollbackResponse, error) {
	s.logger.Info("Processing rollback",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("original_transaction_id", req.OriginalTransactionID))

	// Process the rollback (credit back the amount)
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "rollback",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process rollback", zap.Error(err))
		return &dto.GrooveRollbackResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "ROLLBACK_FAILED",
			ErrorMessage:  "Failed to process rollback",
		}, nil
	}

	s.logger.Info("Rollback processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return &dto.GrooveRollbackResponse{
		Success:       true,
		TransactionID: response.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		NewBalance:    response.Balance,
		Status:        "completed",
	}, nil
}

// ProcessJackpot processes a jackpot transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessJackpot(ctx context.Context, req dto.GrooveJackpotRequest) (*dto.GrooveJackpotResponse, error) {
	s.logger.Info("Processing jackpot",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()),
		zap.String("jackpot_type", req.JackpotType))

	// Process the jackpot (credit)
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "jackpot",
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	response, err := s.storage.ProcessTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to process jackpot", zap.Error(err))
		return &dto.GrooveJackpotResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "JACKPOT_FAILED",
			ErrorMessage:  "Failed to process jackpot",
		}, nil
	}

	s.logger.Info("Jackpot processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.Balance.String()))

	return &dto.GrooveJackpotResponse{
		Success:       true,
		TransactionID: response.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		NewBalance:    response.Balance,
		Status:        "completed",
	}, nil
}

// GetTransactionHistory retrieves transaction history
func (s *GrooveServiceImpl) GetTransactionHistory(ctx context.Context, req dto.GrooveTransactionHistoryRequest) (*dto.GrooveTransactionHistory, error) {
	s.logger.Info("Getting transaction history",
		zap.String("account_id", req.AccountID),
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize))

	history, err := s.storage.GetTransactionHistory(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get transaction history", zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	s.logger.Info("Transaction history retrieved successfully",
		zap.Int("count", len(history.Transactions)))

	return history, nil
}

// CreateGameSession creates a new game session
func (s *GrooveServiceImpl) CreateGameSession(ctx context.Context, accountID, gameID string) (*dto.GrooveGameSession, error) {
	s.logger.Info("Creating game session",
		zap.String("account_id", accountID),
		zap.String("game_id", gameID))

	session := &dto.GrooveGameSession{
		SessionID:    uuid.New().String(),
		AccountID:    accountID,
		GameID:       gameID,
		Status:       "active",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour session
		LastActivity: time.Now(),
	}

	// Get current balance
	balance, err := s.storage.GetAccountBalance(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}
	session.Balance = balance
	session.Currency = "USD"

	createdSession, err := s.storage.CreateGameSession(ctx, *session)
	if err != nil {
		s.logger.Error("Failed to create game session", zap.Error(err))
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	s.logger.Info("Game session created successfully",
		zap.String("session_id", createdSession.SessionID),
		zap.String("game_id", gameID))

	return createdSession, nil
}

// EndGameSession ends a game session
func (s *GrooveServiceImpl) EndGameSession(ctx context.Context, sessionID string) error {
	s.logger.Info("Ending game session", zap.String("session_id", sessionID))

	err := s.storage.EndGameSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to end game session", zap.Error(err))
		return fmt.Errorf("failed to end game session: %w", err)
	}

	s.logger.Info("Game session ended successfully", zap.String("session_id", sessionID))
	return nil
}

// Helper methods

func (s *GrooveServiceImpl) extractUserIDFromSession(sessionID string) (uuid.UUID, error) {
	// Parse JWT token to extract user ID
	claims := &dto.Claim{}

	// Get JWT secret from config
	key := viper.GetString("app.jwt_secret")
	if key == "" {
		key = viper.GetString("auth.jwt_secret") // Fallback for backward compatibility
	}
	if key == "" {
		return uuid.Nil, fmt.Errorf("JWT secret not configured")
	}

	jwtKey := []byte(key)

	// Parse the token
	token, err := jwt.ParseWithClaims(sessionID, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	return claims.UserID, nil
}

func (s *GrooveServiceImpl) validateTransaction(req dto.GrooveTransactionRequest) error {
	if req.AccountID == "" {
		return fmt.Errorf("account ID is required")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than zero")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}

// LaunchGame creates a new game session and calls GrooveTech API to get game URL
func (s *GrooveServiceImpl) LaunchGame(ctx context.Context, userID uuid.UUID, req dto.LaunchGameRequest) (*dto.LaunchGameResponse, error) {
	s.logger.Info("Launching game",
		zap.String("user_id", userID.String()),
		zap.String("game_id", req.GameID),
		zap.String("device_type", req.DeviceType),
		zap.String("game_mode", req.GameMode))

	// Create game session in our database
	session, err := s.gameSessionStorage.CreateGameSession(ctx, userID, req.GameID, req.DeviceType, req.GameMode)
	if err != nil {
		s.logger.Error("Failed to create game session", zap.Error(err))
		return &dto.LaunchGameResponse{
			Success:   false,
			ErrorCode: "SESSION_CREATION_FAILED",
			Message:   "Failed to create game session",
		}, err
	}

	// Get or create user account for GrooveTech
	account, err := s.GetAccountByUserID(ctx, userID)
	if err != nil {
		s.logger.Info("GrooveTech account not found, creating new one", zap.String("user_id", userID.String()))

		// Create new GrooveTech account
		accountID := fmt.Sprintf("groove_%s", userID.String())
		newAccount := dto.GrooveAccount{
			AccountID:    accountID,
			SessionID:    session.SessionID, // Use the new session ID
			Balance:      decimal.Zero,      // Start with zero balance
			Currency:     req.Currency,
			Status:       "active",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		account, err = s.storage.CreateAccount(ctx, newAccount, userID)
		if err != nil {
			s.logger.Error("Failed to create GrooveTech account", zap.Error(err))
			return &dto.LaunchGameResponse{
				Success:   false,
				ErrorCode: "ACCOUNT_CREATION_FAILED",
				Message:   "Failed to create GrooveTech account",
			}, err
		}

		s.logger.Info("GrooveTech account created successfully",
			zap.String("account_id", account.AccountID),
			zap.String("session_id", session.SessionID))
	} else {
		// Update existing account with new session ID
		account.SessionID = session.SessionID
		account.LastActivity = time.Now()
		_, err = s.storage.UpdateAccount(ctx, *account)
		if err != nil {
			s.logger.Error("Failed to update GrooveTech account", zap.Error(err))
			return &dto.LaunchGameResponse{
				Success:   false,
				ErrorCode: "ACCOUNT_UPDATE_FAILED",
				Message:   "Failed to update GrooveTech account",
			}, err
		}

		s.logger.Info("GrooveTech account updated with new session",
			zap.String("account_id", account.AccountID),
			zap.String("session_id", session.SessionID))
	}

	// Use CMA compliance parameters from request
	country := req.Country
	currency := req.Currency
	language := req.Language
	isTestAccount := false
	if req.IsTestAccount != nil {
		isTestAccount = *req.IsTestAccount
	}
	realityCheckElapsed := req.RealityCheckElapsed
	realityCheckInterval := req.RealityCheckInterval

	// Build GrooveTech API URL with parameters
	grooveURL := s.buildGrooveGameURL(session.SessionID, account.AccountID, req.GameID, country, currency, language, req.DeviceType, req.GameMode, session.HomeURL, session.ExitURL, session.HistoryURL, session.LicenseType, isTestAccount, realityCheckElapsed, realityCheckInterval)

	// TODO: Make actual HTTP request to GrooveTech API when credentials are available
	// For now, we'll return the constructed URL for testing purposes
	s.logger.Info("GrooveTech API URL constructed",
		zap.String("groove_url", grooveURL),
		zap.String("note", "Actual GrooveTech API call not implemented yet - credentials needed"))

	// Update session with GrooveTech URL
	err = s.gameSessionStorage.UpdateGameSessionURL(ctx, session.SessionID, grooveURL)
	if err != nil {
		s.logger.Error("Failed to update game session URL", zap.Error(err))
		return &dto.LaunchGameResponse{
			Success:   false,
			ErrorCode: "URL_UPDATE_FAILED",
			Message:   "Failed to update game session URL",
		}, err
	}

	s.logger.Info("Game launched successfully",
		zap.String("session_id", session.SessionID),
		zap.String("game_url", grooveURL))

	return &dto.LaunchGameResponse{
		Success:   true,
		GameURL:   grooveURL,
		SessionID: session.SessionID,
	}, nil
}

// ValidateGameSession validates a game session and returns session details
func (s *GrooveServiceImpl) ValidateGameSession(ctx context.Context, sessionID string) (*dto.GameSession, error) {
	s.logger.Info("Validating game session", zap.String("session_id", sessionID))

	session, err := s.gameSessionStorage.GetGameSessionBySessionID(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to validate game session", zap.Error(err))
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.logger.Warn("Game session expired", zap.String("session_id", sessionID))
		return nil, fmt.Errorf("game session expired")
	}

	// Check if session is active
	if !session.IsActive {
		s.logger.Warn("Game session inactive", zap.String("session_id", sessionID))
		return nil, fmt.Errorf("game session inactive")
	}

	return session, nil
}

// buildGrooveGameURL constructs the GrooveTech game launch URL
func (s *GrooveServiceImpl) buildGrooveGameURL(sessionID, accountID, gameID, country, currency, language, deviceType, gameMode, homeURL, exitURL, historyURL, licenseType string, isTestAccount bool, realityCheckElapsed, realityCheckInterval int) string {
	operatorID := viper.GetString("groove.operator_id")
	if operatorID == "" {
		operatorID = "3818" // Real GrooveTech operator ID
	}

	grooveDomain := viper.GetString("groove.api_domain")
	if grooveDomain == "" {
		grooveDomain = "https://routerstg.groovegaming.com" // Real GrooveTech domain
	}

	// Build URL with all required parameters
	url := fmt.Sprintf("%s/game/?accountid=%s&country=%s&nogsgameid=%s&nogslang=%s&nogsmode=%s&nogsoperatorid=%s&nogscurrency=%s&sessionid=%s&homeurl=%s&license=%s&is_test_account=%t&device_type=%s&realityCheckElapsed=%d&realityCheckInterval=%d",
		grooveDomain,
		accountID,
		country,
		gameID,
		language,
		gameMode,
		operatorID,
		currency,
		sessionID,
		homeURL,
		licenseType,
		isTestAccount,
		deviceType,
		realityCheckElapsed,
		realityCheckInterval,
	)

	// Add optional parameters if provided
	if historyURL != "" {
		url += fmt.Sprintf("&historyUrl=%s", historyURL)
	}
	if exitURL != "" {
		url += fmt.Sprintf("&exitUrl=%s", exitURL)
	}

	return url
}

// GetUserProfile retrieves user profile information for GrooveTech API
func (s *GrooveServiceImpl) GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.GrooveUserProfile, error) {
	s.logger.Info("Getting user profile", zap.String("user_id", userID.String()))

	profile, err := s.storage.GetUserProfile(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user profile", zap.Error(err))
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	s.logger.Info("User profile retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("city", profile.City),
		zap.String("country", profile.Country),
		zap.String("currency", profile.Currency))

	return profile, nil
}

// GetUserBalance retrieves user balance from the existing balance system
func (s *GrooveServiceImpl) GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	s.logger.Info("Getting user balance", zap.String("user_id", userID.String()))

	// Use the storage layer method
	balance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get user balance: %w", err)
	}

	s.logger.Info("User balance retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("balance", balance.String()))

	return balance, nil
}
