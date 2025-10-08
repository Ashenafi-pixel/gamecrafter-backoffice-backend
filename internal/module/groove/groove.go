package groove

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/groove"
	"github.com/tucanbit/platform/utils"
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
	ProcessWagerTransaction(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error)
	GetTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error)
	ProcessResult(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error)
	ProcessResultTransaction(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error)
	ProcessWagerAndResult(ctx context.Context, req dto.GrooveWagerAndResultRequest) (*dto.GrooveWagerAndResultResponse, error)
	ProcessRollback(ctx context.Context, req dto.GrooveRollbackRequestOfficial) (*dto.GrooveRollbackResponseOfficial, error)
	ProcessJackpot(ctx context.Context, req dto.GrooveJackpotRequestOfficial) (*dto.GrooveJackpotResponseOfficial, error)
	ProcessRollbackOnResult(ctx context.Context, req dto.GrooveRollbackOnResultRequest) (*dto.GrooveRollbackOnResultResponse, error)
	ProcessRollbackOnRollback(ctx context.Context, req dto.GrooveRollbackOnRollbackRequest) (*dto.GrooveRollbackOnRollbackResponse, error)
	ProcessWagerByBatch(ctx context.Context, req dto.GrooveWagerByBatchRequest) (*dto.GrooveWagerByBatchResponse, error)

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

// CashbackService interface for processing cashback
type CashbackService interface {
	ProcessBetCashback(ctx context.Context, bet dto.Bet) error
}

type GrooveServiceImpl struct {
	storage            groove.GrooveStorage
	gameSessionStorage groove.GameSessionStorage
	cashbackService    CashbackService
	userStorage        storage.User
	userWS             utils.UserWS
	logger             *zap.Logger
}

func NewGrooveService(storage groove.GrooveStorage, gameSessionStorage groove.GameSessionStorage, cashbackService CashbackService, userStorage storage.User, userWS utils.UserWS, logger *zap.Logger) GrooveService {
	return &GrooveServiceImpl{
		storage:            storage,
		gameSessionStorage: gameSessionStorage,
		cashbackService:    cashbackService,
		userStorage:        userStorage,
		userWS:             userWS,
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
		AccountID:    userID.String(), // Use user ID as account ID for consistency
		SessionID:    "",              // Will be set when user logs in
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
		GameSessionID: req.SessionID,
		BetAmount:     req.Amount,
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
		GameSessionID: req.SessionID,
		BetAmount:     req.Amount,
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

func (s *GrooveServiceImpl) ProcessWager(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error) {
	s.logger.Info("Processing wager",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("bet_amount", req.BetAmount.String()))

	return s.ProcessWagerTransaction(ctx, req)
}

func (s *GrooveServiceImpl) ProcessResult(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error) {
	s.logger.Info("Processing result",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("amount", req.Amount.String()))

	// Process the result (credit)
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		GameSessionID: req.SessionID,
		BetAmount:     req.Amount,
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
		Balance:       response.Balance,
		Status:        "Success",
		Code:          200,
	}, nil
}

// ProcessResultTransaction processes a result transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessResultTransaction(ctx context.Context, req dto.GrooveResultRequest) (*dto.GrooveResultResponse, error) {
	s.logger.Info("Processing result transaction",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("result", req.Result.String()),
		zap.String("game_status", req.GameStatus))

	// Validate game session (but don't fail if expired - Results must be processed even if session expired)
	_, err := s.ValidateGameSession(ctx, req.GameSessionID)
	if err != nil {
		s.logger.Warn("Game session validation failed, but processing result anyway",
			zap.String("session_id", req.GameSessionID), zap.Error(err))
		// Continue processing - Results must be accepted even if session expired
	}

	// Get account by account ID
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account", zap.Error(err))
		return &dto.GrooveResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: req.APIVersion,
		}, nil
	}

	// Check idempotency - if RESULT transaction already exists, return original response
	// Note: We don't check for wager transactions here since wager and result use the same transaction ID
	existingResultTransaction, err := s.storage.GetResultTransactionByID(ctx, req.TransactionID)
	if err == nil && existingResultTransaction != nil {
		s.logger.Info("Duplicate result transaction found", zap.String("transaction_id", req.TransactionID))

		// Get current balance
		userUUID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID", zap.Error(err))
			return &dto.GrooveResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: req.APIVersion,
			}, nil
		}
		currentBalance, err := s.storage.GetUserBalance(ctx, userUUID)
		if err != nil {
			s.logger.Error("Failed to get current balance", zap.Error(err))
			return &dto.GrooveResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: req.APIVersion,
			}, nil
		}

		// Return original response with current balance
		return &dto.GrooveResultResponse{
			Code:          200,
			Status:        "Success - duplicate request",
			Success:       true,
			TransactionID: req.TransactionID,
			AccountID:     req.AccountID,
			SessionID:     req.GameSessionID,
			Amount:        req.Result,
			WalletTx:      existingResultTransaction.AccountTransactionID,
			Balance:       currentBalance,
			BonusWin:      decimal.Zero, // We don't use bonuses
			RealMoneyWin:  req.Result,
			BonusBalance:  decimal.Zero,
			RealBalance:   currentBalance,
			GameMode:      1, // Real money mode
			Order:         "cash_money",
			APIVersion:    req.APIVersion,
		}, nil
	}

	// If result is 0, no balance change needed
	if req.Result.Equal(decimal.Zero) {
		s.logger.Info("Result is zero, no balance change needed", zap.String("transaction_id", req.TransactionID))

		// Get current balance
		userUUID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID", zap.Error(err))
			return &dto.GrooveResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: req.APIVersion,
			}, nil
		}
		currentBalance, err := s.storage.GetUserBalance(ctx, userUUID)
		if err != nil {
			s.logger.Error("Failed to get current balance", zap.Error(err))
			return &dto.GrooveResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: req.APIVersion,
			}, nil
		}

		// Store transaction for idempotency
		transaction := dto.GrooveTransaction{
			TransactionID: req.TransactionID,
			AccountID:     req.AccountID,
			GameSessionID: req.GameSessionID,
			GameID:        req.GameID,
			RoundID:       req.RoundID,
			BetAmount:     req.Result,
			Device:        req.Device,
			CreatedAt:     time.Now(),
		}

		err = s.storage.StoreTransaction(ctx, &transaction, "result")
		if err != nil {
			s.logger.Error("Failed to store transaction", zap.Error(err))
		}

		// Process cashback for zero result (player lost everything)
		s.processResultCashback(ctx, req, account.UserID)

		return &dto.GrooveResultResponse{
			Code:         200,
			Status:       "Success",
			WalletTx:     fmt.Sprintf("TXN_%s_%d", req.TransactionID[:8], time.Now().Unix()),
			Balance:      currentBalance,
			BonusWin:     decimal.Zero,
			RealMoneyWin: decimal.Zero,
			BonusBalance: decimal.Zero,
			RealBalance:  currentBalance,
			GameMode:     1,
			Order:        "cash_money",
			APIVersion:   req.APIVersion,
		}, nil
	}

	// Add winnings to balance
	userUUID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: req.APIVersion,
		}, nil
	}
	newBalance, err := s.storage.AddBalance(ctx, userUUID, req.Result)
	if err != nil {
		s.logger.Error("Failed to add balance", zap.Error(err))
		return &dto.GrooveResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: req.APIVersion,
		}, nil
	}

	// Generate wallet transaction ID
	walletTxID := fmt.Sprintf("TXN_%s_%d", req.TransactionID[:8], time.Now().Unix())

	// Store transaction for idempotency
	transaction := dto.GrooveTransaction{
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		GameSessionID: req.GameSessionID,
		GameID:        req.GameID,
		RoundID:       req.RoundID,
		BetAmount:     req.Result,
		Device:        req.Device,
		CreatedAt:     time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, &transaction, "result")
	if err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
	}

	s.logger.Info("Result transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("result", req.Result.String()),
		zap.String("new_balance", newBalance.String()))

	// Process cashback after result is known for accurate GGR calculation
	s.processResultCashback(ctx, req, account.UserID)

	return &dto.GrooveResultResponse{
		Code:          200,
		Status:        "Success",
		Success:       true,
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.GameSessionID,
		Amount:        req.Result,
		WalletTx:      walletTxID,
		Balance:       newBalance,
		BonusWin:      decimal.Zero, // We don't use bonuses
		RealMoneyWin:  req.Result,
		BonusBalance:  decimal.Zero,
		RealBalance:   newBalance,
		GameMode:      1, // Real money mode
		Order:         "cash_money",
		APIVersion:    req.APIVersion,
	}, nil
}

// ProcessWagerAndResult processes a combined wager and result transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessWagerAndResult(ctx context.Context, req dto.GrooveWagerAndResultRequest) (*dto.GrooveWagerAndResultResponse, error) {
	s.logger.Info("Processing wager and result",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("wager_amount", req.BetAmount.String()),
		zap.String("win_amount", req.WinAmount.String()))

	// Validate game session
	_, err := s.ValidateGameSession(ctx, req.SessionID)
	if err != nil {
		s.logger.Error("Invalid game session", zap.String("session_id", req.SessionID))
		return &dto.GrooveWagerAndResultResponse{
			Code:          1000,
			Status:        "Not logged on",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "INVALID_SESSION",
			ErrorMessage:  "Player session is invalid or expired",
		}, nil
	}

	// Get GrooveTech account
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get GrooveTech account", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Code:          110,
			Status:        "Operation not allowed",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "ACCOUNT_NOT_FOUND",
			ErrorMessage:  "Account not found",
		}, nil
	}

	// Check idempotency
	existingTransaction, err := s.storage.GetTransactionByID(ctx, req.TransactionID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		s.logger.Error("Failed to check transaction idempotency", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Code:          1,
			Status:        "Technical error",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "TECHNICAL_ERROR",
			ErrorMessage:  "Technical error",
		}, nil
	}

	if existingTransaction != nil {
		s.logger.Info("Duplicate wager and result transaction", zap.String("transaction_id", req.TransactionID))
		// Get current balance
		userID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID", zap.Error(err))
			return &dto.GrooveWagerAndResultResponse{
				Code:          1,
				Status:        "Technical error",
				Success:       false,
				TransactionID: req.TransactionID,
				ErrorCode:     "TECHNICAL_ERROR",
				ErrorMessage:  "Technical error",
			}, nil
		}
		currentBalance, err := s.storage.GetUserBalance(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get current balance", zap.Error(err))
			return &dto.GrooveWagerAndResultResponse{
				Code:          1,
				Status:        "Technical error",
				Success:       false,
				TransactionID: req.TransactionID,
				ErrorCode:     "TECHNICAL_ERROR",
				ErrorMessage:  "Technical error",
			}, nil
		}

		// Return original response for duplicate
		return &dto.GrooveWagerAndResultResponse{
			Code:          200,
			Status:        "Success - duplicate request",
			Success:       true,
			TransactionID: req.TransactionID,
			AccountID:     req.AccountID,
			SessionID:     req.SessionID,
			RoundID:       req.RoundID,
			GameStatus:    req.GameStatus,
			WalletTx:      existingTransaction.AccountTransactionID,
			Balance:       currentBalance,
			BonusWin:      decimal.Zero,
			RealMoneyWin:  req.WinAmount,
			BonusMoneyBet: decimal.Zero,
			RealMoneyBet:  req.BetAmount,
			BonusBalance:  decimal.Zero,
			RealBalance:   currentBalance,
			GameMode:      1,
			Order:         "cash_money",
			APIVersion:    req.APIVersion,
		}, nil
	}

	// Get current balance
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Code:          1,
			Status:        "Technical error",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "TECHNICAL_ERROR",
			ErrorMessage:  "Technical error",
		}, nil
	}
	currentBalance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return &dto.GrooveWagerAndResultResponse{
			Code:          1,
			Status:        "Technical error",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "TECHNICAL_ERROR",
			ErrorMessage:  "Technical error",
		}, nil
	}

	// Check if account has sufficient balance for wager
	if currentBalance.LessThan(req.BetAmount) {
		return &dto.GrooveWagerAndResultResponse{
			Code:          1006,
			Status:        "Out of money",
			Success:       false,
			TransactionID: req.TransactionID,
			ErrorCode:     "INSUFFICIENT_FUNDS",
			ErrorMessage:  "Insufficient funds to place the wager",
		}, nil
	}

	// Calculate net result (win amount - wager amount)
	netResult := req.WinAmount.Sub(req.BetAmount)

	// Process the net result (deduct wager, add win)
	var newBalance decimal.Decimal
	if netResult.GreaterThan(decimal.Zero) {
		// Player won - add net winnings
		newBalance, err = s.storage.AddBalance(ctx, userID, netResult)
		if err != nil {
			s.logger.Error("Failed to add balance", zap.Error(err))
			return &dto.GrooveWagerAndResultResponse{
				Code:          1,
				Status:        "Technical error",
				Success:       false,
				TransactionID: req.TransactionID,
				ErrorCode:     "TECHNICAL_ERROR",
				ErrorMessage:  "Technical error",
			}, nil
		}
	} else if netResult.LessThan(decimal.Zero) {
		// Player lost - deduct the loss amount
		lossAmount := netResult.Abs()
		newBalance, err = s.storage.DeductBalance(ctx, userID, lossAmount)
		if err != nil {
			s.logger.Error("Failed to deduct balance", zap.Error(err))
			return &dto.GrooveWagerAndResultResponse{
				Code:          1,
				Status:        "Technical error",
				Success:       false,
				TransactionID: req.TransactionID,
				ErrorCode:     "TECHNICAL_ERROR",
				ErrorMessage:  "Technical error",
			}, nil
		}
	} else {
		// Break even - no balance change
		newBalance = currentBalance
	}

	// Generate wallet transaction ID
	walletTxID := fmt.Sprintf("TXN_%s_%d", req.TransactionID, time.Now().Unix())

	// Store transaction
	transaction := dto.GrooveTransaction{
		TransactionID:        req.TransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.SessionID,
		BetAmount:            netResult,
		AccountTransactionID: walletTxID,
		CreatedAt:            time.Now(),
		Status:               "completed",
	}

	err = s.storage.StoreTransaction(ctx, &transaction, "wager")
	if err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
	}

	// Trigger winner notification if player won (netResult > 0)
	if netResult.GreaterThan(decimal.Zero) && s.userWS != nil {
		// Get user information for winner notification
		user, exists, err := s.userStorage.GetUserByID(ctx, userID)
		if err == nil && exists {
			// Get game information from session
			gameSession, err := s.gameSessionStorage.GetGameSessionBySessionID(ctx, req.SessionID)
			gameName := "GrooveTech Game"
			gameID := ""
			if err == nil && gameSession != nil {
				gameID = gameSession.GameID
				gameName = fmt.Sprintf("GrooveTech Game %s", gameSession.GameID)
			}

			// Create winner notification data
			winnerData := dto.WinnerNotificationData{
				Username:      user.Username,
				Email:         user.Email,
				GameName:      gameName,
				GameID:        gameID,
				BetAmount:     req.BetAmount,
				WinAmount:     req.WinAmount,
				NetWinnings:   netResult,
				Currency:      "USD",
				Timestamp:     time.Now(),
				SessionID:     req.SessionID,
				RoundID:       req.RoundID,
				TransactionID: req.TransactionID,
			}

			// Trigger winner notification WebSocket
			s.userWS.TriggerWinnerNotificationWS(ctx, userID, winnerData)
			s.logger.Info("Winner notification triggered",
				zap.String("user_id", userID.String()),
				zap.String("username", user.Username),
				zap.String("game_name", gameName),
				zap.String("win_amount", req.WinAmount.String()),
				zap.String("net_winnings", netResult.String()))
		}
	}

	s.logger.Info("Wager and result processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("net_result", netResult.String()),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveWagerAndResultResponse{
		Code:          200,
		Status:        "Success",
		Success:       true,
		TransactionID: req.TransactionID,
		AccountID:     req.AccountID,
		SessionID:     req.SessionID,
		RoundID:       req.RoundID,
		GameStatus:    req.GameStatus,
		WalletTx:      walletTxID,
		Balance:       newBalance,
		BonusWin:      decimal.Zero,
		RealMoneyWin:  req.WinAmount,
		BonusMoneyBet: decimal.Zero,
		RealMoneyBet:  req.BetAmount,
		BonusBalance:  decimal.Zero,
		RealBalance:   newBalance,
		GameMode:      1,
		Order:         "cash_money",
		APIVersion:    req.APIVersion,
	}, nil
}

// ProcessRollback processes a rollback transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessRollback(ctx context.Context, req dto.GrooveRollbackRequestOfficial) (*dto.GrooveRollbackResponseOfficial, error) {
	s.logger.Info("Processing rollback transaction", zap.String("transaction_id", req.TransactionID))

	// NOTE: Session validation is NOT performed for rollbacks
	// Rollbacks must be accepted even if the game session has expired
	// This is per GrooveTech specification

	// Check idempotency - look for existing rollback transaction
	rollbackTransactionID := req.TransactionID + "_rollback"
	existingRollback, err := s.storage.GetTransactionByID(ctx, rollbackTransactionID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		s.logger.Error("Failed to check rollback idempotency", zap.Error(err))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	if existingRollback != nil {
		s.logger.Info("Duplicate rollback transaction", zap.String("transaction_id", req.TransactionID))
		// Get current balance for duplicate response
		account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
		if err != nil {
			s.logger.Error("Failed to get account for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		userID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		currentBalance, err := s.storage.GetUserBalance(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get current balance for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		// Return original response for duplicate
		return &dto.GrooveRollbackResponseOfficial{
			Code:                 200,
			Status:               "Success - duplicate request",
			AccountTransactionID: existingRollback.AccountTransactionID,
			Balance:              currentBalance,
			BonusBalance:         decimal.Zero,
			RealBalance:          currentBalance,
			GameMode:             1,
			Order:                "cash_money",
			APIVersion:           "1.2",
		}, nil
	}

	// Get account
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account", zap.Error(err))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Parse user ID
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Transaction Validation: Find original wager transaction
	originalWager, err := s.storage.GetTransactionByID(ctx, req.TransactionID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		s.logger.Error("Failed to find original wager", zap.Error(err))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	if originalWager == nil {
		s.logger.Error("Original wager not found", zap.String("transaction_id", req.TransactionID))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       102,
			Status:     "Wager not found",
			APIVersion: "1.2",
		}, nil
	}

	// Verify rollback eligibility: Check if wager has already been rolled back
	if originalWager.Status == "rolled_back" {
		s.logger.Error("Wager already rolled back", zap.String("transaction_id", req.TransactionID))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       409,
			Status:     "Transaction ID exists",
			APIVersion: "1.2",
		}, nil
	}

	// Verify rollback eligibility: Check if wager has a result transaction
	// (This would require checking for result transactions, but for now we'll assume
	// that if the wager exists and isn't rolled back, it's eligible for rollback)

	// Restore player balance
	rollbackAmount := req.RollbackAmount
	if rollbackAmount.IsZero() {
		rollbackAmount = originalWager.BetAmount
	}

	newBalance, err := s.storage.AddBalance(ctx, userID, rollbackAmount)
	if err != nil {
		s.logger.Error("Failed to restore balance", zap.Error(err))
		return &dto.GrooveRollbackResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	// Generate account transaction ID
	accountTransactionID := uuid.New().String()

	// Store rollback transaction
	rollbackTransaction := &dto.GrooveTransaction{
		TransactionID:        req.TransactionID + "_rollback",
		AccountTransactionID: accountTransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.GameSessionID,
		RoundID:              req.RoundID,
		GameID:               req.GameID,
		BetAmount:            rollbackAmount,
		Device:               req.Device,
		FRBID:                "",
		UserID:               userID,
		CreatedAt:            time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, rollbackTransaction, "rollback")
	if err != nil {
		s.logger.Error("Failed to store rollback transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	// Update original wager status
	originalWager.Status = "rolled_back"
	err = s.storage.StoreTransaction(ctx, originalWager, "wager")
	if err != nil {
		s.logger.Error("Failed to update original wager status", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	s.logger.Info("Rollback transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_transaction_id", accountTransactionID),
		zap.String("rollback_amount", rollbackAmount.String()),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveRollbackResponseOfficial{
		Code:                 200,
		Status:               "Success",
		AccountTransactionID: accountTransactionID,
		Balance:              newBalance,
		BonusBalance:         decimal.Zero,
		RealBalance:          newBalance,
		GameMode:             1,
		Order:                "cash_money",
		APIVersion:           "1.2",
	}, nil
}

// ProcessJackpot processes a jackpot transaction (Official GrooveTech API)
func (s *GrooveServiceImpl) ProcessJackpot(ctx context.Context, req dto.GrooveJackpotRequestOfficial) (*dto.GrooveJackpotResponseOfficial, error) {
	s.logger.Info("Processing jackpot transaction", zap.String("transaction_id", req.TransactionID))

	// Check idempotency
	existingTransaction, err := s.storage.GetTransactionByID(ctx, req.TransactionID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		s.logger.Error("Failed to check transaction idempotency", zap.Error(err))
		return &dto.GrooveJackpotResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	if existingTransaction != nil {
		s.logger.Info("Duplicate jackpot transaction", zap.String("transaction_id", req.TransactionID))
		// Get current balance for duplicate response
		account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
		if err != nil {
			s.logger.Error("Failed to get account for duplicate response", zap.Error(err))
			return &dto.GrooveJackpotResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		userID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID for duplicate response", zap.Error(err))
			return &dto.GrooveJackpotResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		currentBalance, err := s.storage.GetUserBalance(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get current balance for duplicate response", zap.Error(err))
			return &dto.GrooveJackpotResponseOfficial{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		// Return original response for duplicate
		return &dto.GrooveJackpotResponseOfficial{
			Code:         200,
			Status:       "Success - duplicate request",
			WalletTx:     existingTransaction.AccountTransactionID,
			Balance:      currentBalance,
			BonusWin:     decimal.Zero,
			RealMoneyWin: existingTransaction.BetAmount,
			BonusBalance: decimal.Zero,
			RealBalance:  currentBalance,
			GameMode:     1,
			Order:        "cash_money",
			APIVersion:   "1.2",
		}, nil
	}

	// Get account - for jackpots, account might not exist yet
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Info("Account not found, creating new account for jackpot", zap.String("account_id", req.AccountID))
		// For jackpots, we need to create the account if it doesn't exist
		// This is because jackpots might not be related to rounds
		// We need to get the user ID from the account ID
		// For now, let's use a default user ID or handle this differently
		s.logger.Error("Account not found for jackpot", zap.Error(err))
		return &dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Parse user ID
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Add jackpot amount to balance
	newBalance, err := s.storage.AddBalance(ctx, userID, req.Amount)
	if err != nil {
		s.logger.Error("Failed to add jackpot amount", zap.Error(err))
		return &dto.GrooveJackpotResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	// Generate account transaction ID
	accountTransactionID := uuid.New().String()

	// Store jackpot transaction
	jackpotTransaction := &dto.GrooveTransaction{
		TransactionID:        req.TransactionID,
		AccountTransactionID: accountTransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.GameSessionID,
		RoundID:              req.RoundID,
		GameID:               req.GameID,
		BetAmount:            req.Amount,
		Device:               "desktop", // Default device for jackpot
		FRBID:                "",
		UserID:               userID,
		CreatedAt:            time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, jackpotTransaction, "jackpot")
	if err != nil {
		s.logger.Error("Failed to store jackpot transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	s.logger.Info("Jackpot transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_transaction_id", accountTransactionID),
		zap.String("jackpot_amount", req.Amount.String()),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveJackpotResponseOfficial{
		Code:         200,
		Status:       "Success",
		WalletTx:     accountTransactionID,
		Balance:      newBalance,
		BonusWin:     decimal.Zero,
		RealMoneyWin: req.Amount,
		BonusBalance: decimal.Zero,
		RealBalance:  newBalance,
		GameMode:     1,
		Order:        "cash_money",
		APIVersion:   "1.2",
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
		accountID := userID.String() // Use user ID as account ID for consistency
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
		grooveDomain = "https://gprouter.groovegaming.com" // Real GrooveTech domain
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

// ProcessWagerTransaction processes a wager transaction according to GrooveTech spec
func (s *GrooveServiceImpl) ProcessWagerTransaction(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error) {
	s.logger.Info("Processing wager transaction",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("bet_amount", req.BetAmount.String()))

	// Generate account transaction ID
	transactionPrefix := req.TransactionID
	if len(req.TransactionID) > 8 {
		transactionPrefix = req.TransactionID[:8]
	}
	accountTransactionID := fmt.Sprintf("TXN_%s_%d", transactionPrefix, time.Now().Unix())

	// Deduct bet amount from user balance
	newBalance, err := s.storage.DeductBalance(ctx, req.UserID, req.BetAmount)
	if err != nil {
		s.logger.Error("Failed to deduct balance", zap.Error(err))
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Store transaction for idempotency
	transaction := &dto.GrooveTransaction{
		TransactionID:        req.TransactionID,
		AccountTransactionID: accountTransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.GameSessionID,
		RoundID:              req.RoundID,
		GameID:               req.GameID,
		BetAmount:            req.BetAmount,
		Device:               req.Device,
		FRBID:                req.FRBID,
		UserID:               req.UserID,
		CreatedAt:            time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, transaction, "wager")
	if err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	// Cashback will be processed after result is known for accurate GGR calculation
	s.logger.Info("Wager processed - cashback will be calculated after result",
		zap.String("transaction_id", req.TransactionID),
		zap.String("user_id", req.UserID.String()))

	s.logger.Info("Wager transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_transaction_id", accountTransactionID),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveWagerResponse{
		Code:                 200,
		Status:               "Success",
		AccountTransactionID: accountTransactionID,
		Balance:              newBalance,
		BonusMoneyBet:        decimal.Zero, // No bonus money in our system
		RealMoneyBet:         req.BetAmount,
		BonusBalance:         decimal.Zero, // No bonus balance in our system
		RealBalance:          newBalance,
		GameMode:             1, // Real money mode
		Order:                "cash_money",
		APIVersion:           "1.2",
	}, nil
}

// GetTransactionByID retrieves a transaction by its ID for idempotency checks
func (s *GrooveServiceImpl) GetTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error) {
	s.logger.Info("Getting transaction by ID", zap.String("transaction_id", transactionID))

	transaction, err := s.storage.GetTransactionByID(ctx, transactionID)
	if err != nil {
		s.logger.Error("Failed to get transaction by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	if transaction == nil {
		s.logger.Info("Transaction not found", zap.String("transaction_id", transactionID))
		return nil, nil
	}

	s.logger.Info("Transaction retrieved successfully",
		zap.String("transaction_id", transactionID),
		zap.String("account_transaction_id", transaction.AccountTransactionID))

	return transaction, nil
}

// ProcessRollbackOnResult processes a rollback on result transaction
func (s *GrooveServiceImpl) ProcessRollbackOnResult(ctx context.Context, req dto.GrooveRollbackOnResultRequest) (*dto.GrooveRollbackOnResultResponse, error) {
	s.logger.Info("Processing rollback on result transaction", zap.String("transaction_id", req.TransactionID))

	// Check idempotency - look for existing rollback on result transaction
	rollbackTransactionID := req.TransactionID + "_rollback_result"
	existingTransaction, err := s.storage.GetTransactionByID(ctx, rollbackTransactionID)
	if err != nil {
		s.logger.Error("Failed to check transaction idempotency", zap.Error(err))
		return &dto.GrooveRollbackOnResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	if existingTransaction != nil {
		s.logger.Info("Duplicate rollback on result transaction", zap.String("transaction_id", req.TransactionID))
		// Get current balance for duplicate response
		account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
		if err != nil {
			s.logger.Error("Failed to get account for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		userID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		currentBalance, err := s.storage.GetUserBalance(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get current balance for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnResultResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		// Return original response for duplicate
		return &dto.GrooveRollbackOnResultResponse{
			Code:                 200,
			Status:               "Success - duplicate request",
			AccountTransactionID: existingTransaction.AccountTransactionID,
			Balance:              currentBalance,
			BonusBalance:         decimal.Zero,
			RealBalance:          currentBalance,
			GameMode:             1,
			Order:                "cash_money",
			APIVersion:           "1.2",
		}, nil
	}

	// Get account
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account", zap.Error(err))
		return &dto.GrooveRollbackOnResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Parse user ID
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveRollbackOnResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Deduct the rollback amount from balance
	newBalance, err := s.storage.DeductBalance(ctx, userID, req.Amount)
	if err != nil {
		s.logger.Error("Failed to deduct rollback amount", zap.Error(err))
		return &dto.GrooveRollbackOnResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	// Generate account transaction ID
	accountTransactionID := uuid.New().String()

	// Store rollback on result transaction
	rollbackTransaction := &dto.GrooveTransaction{
		TransactionID:        req.TransactionID,
		AccountTransactionID: accountTransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.GameSessionID,
		RoundID:              req.RoundID,
		GameID:               req.GameID,
		BetAmount:            req.Amount,
		Device:               req.Device,
		FRBID:                "",
		UserID:               userID,
		CreatedAt:            time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, rollbackTransaction, "rollback")
	if err != nil {
		s.logger.Error("Failed to store rollback on result transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	s.logger.Info("Rollback on result transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_transaction_id", accountTransactionID),
		zap.String("rollback_amount", req.Amount.String()),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveRollbackOnResultResponse{
		Code:                 200,
		Status:               "Success",
		AccountTransactionID: accountTransactionID,
		Balance:              newBalance,
		BonusBalance:         decimal.Zero,
		RealBalance:          newBalance,
		GameMode:             1,
		Order:                "cash_money",
		APIVersion:           "1.2",
	}, nil
}

// ProcessRollbackOnRollback processes a rollback on rollback transaction
func (s *GrooveServiceImpl) ProcessRollbackOnRollback(ctx context.Context, req dto.GrooveRollbackOnRollbackRequest) (*dto.GrooveRollbackOnRollbackResponse, error) {
	s.logger.Info("Processing rollback on rollback transaction", zap.String("transaction_id", req.TransactionID))

	// Check idempotency - look for existing rollback on rollback transaction
	rollbackTransactionID := req.TransactionID + "_rollback_rollback"
	existingTransaction, err := s.storage.GetTransactionByID(ctx, rollbackTransactionID)
	if err != nil {
		s.logger.Error("Failed to check transaction idempotency", zap.Error(err))
		return &dto.GrooveRollbackOnRollbackResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	if existingTransaction != nil {
		s.logger.Info("Duplicate rollback on rollback transaction", zap.String("transaction_id", req.TransactionID))
		// Get current balance for duplicate response
		account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
		if err != nil {
			s.logger.Error("Failed to get account for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnRollbackResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		userID, err := uuid.Parse(account.UserID)
		if err != nil {
			s.logger.Error("Failed to parse user ID for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnRollbackResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		currentBalance, err := s.storage.GetUserBalance(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get current balance for duplicate response", zap.Error(err))
			return &dto.GrooveRollbackOnRollbackResponse{
				Code:       1,
				Status:     "Technical error",
				APIVersion: "1.2",
			}, nil
		}

		// Return original response for duplicate
		return &dto.GrooveRollbackOnRollbackResponse{
			Code:                 200,
			Status:               "Success - duplicate request",
			AccountTransactionID: existingTransaction.AccountTransactionID,
			Balance:              currentBalance,
			BonusBalance:         decimal.Zero,
			RealBalance:          currentBalance,
			GameMode:             1,
			Order:                "cash_money",
			APIVersion:           "1.2",
		}, nil
	}

	// Get account
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account", zap.Error(err))
		return &dto.GrooveRollbackOnRollbackResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Parse user ID
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveRollbackOnRollbackResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		}, nil
	}

	// Add the rollback amount back to balance (reversing the previous rollback)
	newBalance, err := s.storage.AddBalance(ctx, userID, req.RollbackAmount)
	if err != nil {
		s.logger.Error("Failed to add rollback amount", zap.Error(err))
		return &dto.GrooveRollbackOnRollbackResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		}, nil
	}

	// Generate account transaction ID
	accountTransactionID := uuid.New().String()

	// Store rollback on rollback transaction
	rollbackTransaction := &dto.GrooveTransaction{
		TransactionID:        rollbackTransactionID,
		AccountTransactionID: accountTransactionID,
		AccountID:            req.AccountID,
		GameSessionID:        req.GameSessionID,
		RoundID:              req.RoundID,
		GameID:               req.GameID,
		BetAmount:            req.RollbackAmount,
		Device:               req.Device,
		FRBID:                "",
		UserID:               userID,
		Status:               "rollback_rollback",
		CreatedAt:            time.Now(),
	}

	err = s.storage.StoreTransaction(ctx, rollbackTransaction, "rollback")
	if err != nil {
		s.logger.Error("Failed to store rollback on rollback transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

	s.logger.Info("Rollback on rollback transaction processed successfully",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_transaction_id", accountTransactionID),
		zap.String("rollback_amount", req.RollbackAmount.String()),
		zap.String("new_balance", newBalance.String()))

	return &dto.GrooveRollbackOnRollbackResponse{
		Code:                 200,
		Status:               "Success",
		AccountTransactionID: accountTransactionID,
		Balance:              newBalance,
		BonusBalance:         decimal.Zero,
		RealBalance:          newBalance,
		GameMode:             1,
		Order:                "cash_money",
		APIVersion:           "1.2",
	}, nil
}

// ProcessWagerByBatch processes a wager by batch transaction (sportsbook)
func (s *GrooveServiceImpl) ProcessWagerByBatch(ctx context.Context, req dto.GrooveWagerByBatchRequest) (*dto.GrooveWagerByBatchResponse, error) {
	s.logger.Info("Processing wager by batch transaction", zap.Int("bet_count", len(req.Bets)))

	// Get account
	account, err := s.storage.GetAccountByAccountID(ctx, req.AccountID)
	if err != nil {
		s.logger.Error("Failed to get account", zap.Error(err))
		return &dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    110,
			Message: "Operation not allowed",
		}, nil
	}

	// Parse user ID
	userID, err := uuid.Parse(account.UserID)
	if err != nil {
		s.logger.Error("Failed to parse user ID", zap.Error(err))
		return &dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    110,
			Message: "Operation not allowed",
		}, nil
	}

	// Calculate total bet amount
	totalBetAmount := decimal.Zero
	for _, bet := range req.Bets {
		totalBetAmount = totalBetAmount.Add(bet.Amount)
	}

	// Check if user has sufficient balance
	currentBalance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return &dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    1,
			Message: "Technical error",
		}, nil
	}

	if currentBalance.LessThan(totalBetAmount) {
		s.logger.Error("Insufficient funds for batch wager",
			zap.String("current_balance", currentBalance.String()),
			zap.String("required_amount", totalBetAmount.String()))
		return &dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    1006,
			Message: "Out of money",
		}, nil
	}

	// Process all bets atomically
	var betResults []dto.GrooveBatchBetResult
	var finalBalance decimal.Decimal

	// Deduct total amount from balance
	finalBalance, err = s.storage.DeductBalance(ctx, userID, totalBetAmount)
	if err != nil {
		s.logger.Error("Failed to deduct batch amount", zap.Error(err))
		return &dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    1,
			Message: "Technical error",
		}, nil
	}

	// Process each bet
	for _, bet := range req.Bets {
		// Check idempotency for each bet
		existingTransaction, err := s.storage.GetTransactionByID(ctx, bet.TransactionID)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			s.logger.Error("Failed to check bet idempotency", zap.Error(err))
			return &dto.GrooveWagerByBatchResponse{
				Status:  "Failed",
				Code:    1,
				Message: "Technical error",
			}, nil
		}

		if existingTransaction != nil {
			// Return original response for duplicate bet
			betResults = append(betResults, dto.GrooveBatchBetResult{
				ProviderTransactionID: bet.TransactionID,
				TransactionID:         existingTransaction.AccountTransactionID,
				BonusMoneyBet:         decimal.Zero,
				RealMoneyBet:          bet.Amount,
			})
			continue
		}

		// Generate account transaction ID for this bet
		accountTransactionID := uuid.New().String()

		// Store bet transaction
		betTransaction := &dto.GrooveTransaction{
			TransactionID:        bet.TransactionID,
			AccountTransactionID: accountTransactionID,
			AccountID:            req.AccountID,
			GameSessionID:        req.GameSessionID,
			RoundID:              bet.RoundID,
			GameID:               req.GameID,
			BetAmount:            bet.Amount,
			Device:               req.Device,
			FRBID:                bet.FRBID,
			UserID:               userID,
			CreatedAt:            time.Now(),
		}

		err = s.storage.StoreTransaction(ctx, betTransaction, "wager")
		if err != nil {
			s.logger.Error("Failed to store bet transaction", zap.Error(err))
			return &dto.GrooveWagerByBatchResponse{
				Status:  "Failed",
				Code:    1,
				Message: "Technical error",
			}, nil
		}

		betResults = append(betResults, dto.GrooveBatchBetResult{
			ProviderTransactionID: bet.TransactionID,
			TransactionID:         accountTransactionID,
			BonusMoneyBet:         decimal.Zero,
			RealMoneyBet:          bet.Amount,
		})
	}

	s.logger.Info("Wager by batch transaction processed successfully",
		zap.Int("bet_count", len(betResults)),
		zap.String("total_amount", totalBetAmount.String()),
		zap.String("final_balance", finalBalance.String()))

	return &dto.GrooveWagerByBatchResponse{
		Status:       "Success",
		Code:         200,
		Message:      "OK",
		Bets:         betResults,
		Balance:      finalBalance,
		RealBalance:  finalBalance,
		BonusBalance: decimal.Zero,
	}, nil
}

// processBetCashback processes cashback for a GrooveTech bet
func (s *GrooveServiceImpl) processBetCashback(ctx context.Context, req dto.GrooveWagerRequest) {
	s.logger.Info("Processing cashback for GrooveTech bet",
		zap.String("transaction_id", req.TransactionID),
		zap.String("user_id", req.UserID.String()),
		zap.String("bet_amount", req.BetAmount.String()))

	// Process cashback for the bet
	if s.cashbackService != nil {
		// Create a bet DTO for cashback processing
		bet := dto.Bet{
			BetID:               uuid.New(),
			RoundID:             uuid.New(),
			UserID:              req.UserID,
			ClientTransactionID: req.TransactionID,
			Amount:              req.BetAmount,
			Currency:            "USD",
			Timestamp:           time.Now(),
			Payout:              decimal.Zero,
			CashOutMultiplier:   decimal.Zero,
			Status:              "completed",
		}

		// Process the bet for cashback
		err := s.cashbackService.ProcessBetCashback(ctx, bet)
		if err != nil {
			s.logger.Error("Failed to process cashback for GrooveTech bet",
				zap.String("transaction_id", req.TransactionID),
				zap.String("user_id", req.UserID.String()),
				zap.Error(err))
		} else {
			s.logger.Info("Cashback processed successfully for GrooveTech bet",
				zap.String("transaction_id", req.TransactionID),
				zap.String("user_id", req.UserID.String()),
				zap.String("bet_amount", req.BetAmount.String()))
		}
	} else {
		s.logger.Warn("Cashback service not available - skipping cashback processing",
			zap.String("transaction_id", req.TransactionID))
	}
}

// processResultCashback processes cashback after result is known for accurate GGR calculation
func (s *GrooveServiceImpl) processResultCashback(ctx context.Context, req dto.GrooveResultRequest, userIDStr string) {
	s.logger.Info("Processing result-based cashback",
		zap.String("transaction_id", req.TransactionID),
		zap.String("user_id", userIDStr),
		zap.String("result_amount", req.Result.String()))

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.logger.Error("Failed to parse user ID for cashback processing", zap.Error(err))
		return
	}

	// Get the original wager transaction to calculate net loss
	wagerTransaction, err := s.storage.GetWagerTransactionBySessionID(ctx, req.GameSessionID)
	if err != nil {
		s.logger.Error("Failed to get wager transaction for cashback calculation", zap.Error(err))
		return
	}

	if wagerTransaction == nil {
		s.logger.Warn("No wager transaction found for result - skipping cashback",
			zap.String("transaction_id", req.TransactionID))
		return
	}

	// Calculate net loss (bet amount - winnings)
	betAmount := wagerTransaction.BetAmount
	winAmount := req.Result
	netLoss := betAmount.Sub(winAmount)

	s.logger.Info("Calculating cashback based on net loss",
		zap.String("bet_amount", betAmount.String()),
		zap.String("win_amount", winAmount.String()),
		zap.String("net_loss", netLoss.String()))

	// Process cashback on every wager bet (per-spin cashback)
	s.logger.Info("Processing per-spin cashback for every wager",
		zap.String("transaction_id", req.TransactionID),
		zap.String("bet_amount", betAmount.String()))

	// Process cashback for the net loss
	if s.cashbackService != nil {
		// Create a bet DTO for cashback processing based on bet amount (not net loss)
		bet := dto.Bet{
			BetID:               uuid.New(),
			RoundID:             uuid.New(),
			UserID:              userID,
			ClientTransactionID: req.TransactionID,
			Amount:              betAmount, // Use bet amount for GGR calculation
			Currency:            "USD",
			Timestamp:           time.Now(),
			Payout:              winAmount, // Actual winnings
			CashOutMultiplier:   decimal.Zero,
			Status:              "completed",
		}

		// Process the bet for cashback
		err := s.cashbackService.ProcessBetCashback(ctx, bet)
		if err != nil {
			s.logger.Error("Failed to process result-based cashback",
				zap.String("transaction_id", req.TransactionID),
				zap.String("user_id", userID.String()),
				zap.String("net_loss", netLoss.String()),
				zap.Error(err))
		} else {
			s.logger.Info("Result-based cashback processed successfully",
				zap.String("transaction_id", req.TransactionID),
				zap.String("user_id", userID.String()),
				zap.String("net_loss", netLoss.String()),
				zap.String("cashback_based_on", "net_loss"))
		}
	} else {
		s.logger.Warn("Cashback service not available - skipping result-based cashback processing",
			zap.String("transaction_id", req.TransactionID))
	}
}
