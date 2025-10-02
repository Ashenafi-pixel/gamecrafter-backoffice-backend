package groove

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/analytics"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

// BalanceSyncStatus represents the synchronization status of user balances
type BalanceSyncStatus struct {
	UserID             uuid.UUID       `json:"user_id"`
	MainBalance        decimal.Decimal `json:"main_balance"`
	GrooveBalance      decimal.Decimal `json:"groove_balance"`
	IsSynchronized     bool            `json:"is_synchronized"`
	Discrepancy        decimal.Decimal `json:"discrepancy"`
	LastSyncTime       *time.Time      `json:"last_sync_time"`
	LastValidationTime time.Time       `json:"last_validation_time"`
}

// BalanceDiscrepancy represents a balance discrepancy found in the system
type BalanceDiscrepancy struct {
	UserID              uuid.UUID       `json:"user_id"`
	Username            string          `json:"username"`
	Email               string          `json:"email"`
	MainBalance         decimal.Decimal `json:"main_balance"`
	GrooveBalance       decimal.Decimal `json:"groove_balance"`
	Discrepancy         decimal.Decimal `json:"discrepancy"`
	LastActivity        *time.Time      `json:"last_activity"`
	DiscrepancyDetected time.Time       `json:"discrepancy_detected"`
}

type GrooveStorage interface {
	// Account operations
	CreateAccount(ctx context.Context, account dto.GrooveAccount, userID uuid.UUID) (*dto.GrooveAccount, error)
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error)
	GetAccountByID(ctx context.Context, accountID string) (*dto.GrooveAccount, error)
	GetAccountByAccountID(ctx context.Context, accountID string) (*dto.GrooveAccount, error)
	UpdateAccount(ctx context.Context, account dto.GrooveAccount) (*dto.GrooveAccount, error)
	GetAccountBalance(ctx context.Context, accountID string) (decimal.Decimal, error)

	// Transaction operations
	ProcessTransaction(ctx context.Context, transaction dto.GrooveTransaction) (*dto.GrooveTransactionResponse, error)
	GetTransactionHistory(ctx context.Context, req dto.GrooveTransactionHistoryRequest) (*dto.GrooveTransactionHistory, error)

	// Game session operations
	CreateGameSession(ctx context.Context, session dto.GrooveGameSession) (*dto.GrooveGameSession, error)
	EndGameSession(ctx context.Context, sessionID string) error

	// Balance operations
	GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	UpdateUserBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, transactionType string) error
	DeductBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error)
	AddBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error)

	// Transaction storage for idempotency
	StoreTransaction(ctx context.Context, transaction *dto.GrooveTransaction, transactionType string) error
	GetTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error)
	GetResultTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error)
	GetWagerTransactionBySessionID(ctx context.Context, sessionID string) (*dto.GrooveTransaction, error)
	GetTransactionGameInfo(ctx context.Context, transactionID string) (gameID, gameType string, err error)

	// Balance synchronization and validation
	ValidateBalanceSync(ctx context.Context, userID uuid.UUID) (*BalanceSyncStatus, error)
	ReconcileBalances(ctx context.Context, userID uuid.UUID) error
	GetBalanceDiscrepancies(ctx context.Context) ([]BalanceDiscrepancy, error)

	// User profile operations
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.GrooveUserProfile, error)
}

type GrooveStorageImpl struct {
	db                   *persistencedb.PersistenceDB
	logger               *zap.Logger
	userWS               utils.UserWS
	analyticsIntegration *analytics.AnalyticsIntegration
}

func NewGrooveStorage(db *persistencedb.PersistenceDB, userWS utils.UserWS, analyticsIntegration *analytics.AnalyticsIntegration, logger *zap.Logger) GrooveStorage {
	return &GrooveStorageImpl{
		db:                   db,
		logger:               logger,
		userWS:               userWS,
		analyticsIntegration: analyticsIntegration,
	}
}

// CreateAccount creates a new GrooveTech account
func (s *GrooveStorageImpl) CreateAccount(ctx context.Context, account dto.GrooveAccount, userID uuid.UUID) (*dto.GrooveAccount, error) {
	s.logger.Info("Creating GrooveTech account",
		zap.String("account_id", account.AccountID),
		zap.String("user_id", userID.String()))

	// Check if account already exists for this user
	existingAccount, err := s.GetAccountByUserID(ctx, userID)
	if err == nil && existingAccount != nil {
		s.logger.Info("Account already exists for user, returning existing account",
			zap.String("account_id", existingAccount.AccountID),
			zap.String("user_id", userID.String()))
		return existingAccount, nil
	}

	query := `
		INSERT INTO groove_accounts (id, user_id, account_id, session_id, balance, currency, status, created_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, last_activity`

	var createdAt, lastActivity time.Time
	err = s.db.GetPool().QueryRow(ctx, query,
		uuid.New(), // Internal ID
		userID,
		account.AccountID,
		account.SessionID,
		account.Balance,
		account.Currency,
		account.Status,
		account.CreatedAt,
		account.LastActivity,
	).Scan(&createdAt, &lastActivity)

	if err != nil {
		s.logger.Error("Failed to create GrooveTech account", zap.Error(err))
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	account.CreatedAt = createdAt
	account.LastActivity = lastActivity

	s.logger.Info("GrooveTech account created successfully",
		zap.String("account_id", account.AccountID))

	return &account, nil
}

// GetAccountByUserID retrieves account by user ID
func (s *GrooveStorageImpl) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error) {
	s.logger.Info("Getting GrooveTech account by user ID", zap.String("user_id", userID.String()))

	query := `
		SELECT account_id, session_id, balance, currency, status, created_at, last_activity
		FROM groove_accounts
		WHERE user_id = $1`

	var account dto.GrooveAccount
	var sessionID *string
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(
		&account.AccountID,
		&sessionID,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.LastActivity,
	)

	// Handle NULL session_id
	if sessionID != nil {
		account.SessionID = *sessionID
	} else {
		account.SessionID = ""
	}

	if err != nil {
		s.logger.Info("GrooveTech account not found for user", zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("account not found: %w", err)
	}

	s.logger.Info("GrooveTech account retrieved successfully",
		zap.String("account_id", account.AccountID))

	return &account, nil
}

// GetAccountByID retrieves account by account ID
func (s *GrooveStorageImpl) GetAccountByID(ctx context.Context, accountID string) (*dto.GrooveAccount, error) {
	s.logger.Info("Getting GrooveTech account by ID", zap.String("account_id", accountID))

	query := `
		SELECT id, user_id, account_id, session_id, balance, currency, status, created_at, last_activity
		FROM groove_accounts
		WHERE account_id = $1`

	var account dto.GrooveAccount
	err := s.db.GetPool().QueryRow(ctx, query, accountID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountID,
		&account.SessionID,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.LastActivity,
	)

	if err != nil {
		s.logger.Error("Failed to get GrooveTech account", zap.Error(err))
		return nil, fmt.Errorf("account not found: %w", err)
	}

	s.logger.Info("GrooveTech account retrieved successfully",
		zap.String("account_id", account.AccountID))

	return &account, nil
}

// GetAccountByAccountID gets an account by account ID (alias for GetAccountByID)
func (s *GrooveStorageImpl) GetAccountByAccountID(ctx context.Context, accountID string) (*dto.GrooveAccount, error) {
	return s.GetAccountByID(ctx, accountID)
}

// UpdateAccount updates an existing account
func (s *GrooveStorageImpl) UpdateAccount(ctx context.Context, account dto.GrooveAccount) (*dto.GrooveAccount, error) {
	s.logger.Info("Updating GrooveTech account", zap.String("account_id", account.AccountID))

	query := `
		UPDATE groove_accounts 
		SET session_id = $2, balance = $3, status = $4, last_activity = $5
		WHERE account_id = $1
		RETURNING created_at`

	var createdAt time.Time
	err := s.db.GetPool().QueryRow(ctx, query,
		account.AccountID,
		account.SessionID,
		account.Balance,
		account.Status,
		account.LastActivity,
	).Scan(&createdAt)

	if err != nil {
		s.logger.Error("Failed to update GrooveTech account", zap.Error(err))
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	account.CreatedAt = createdAt

	s.logger.Info("GrooveTech account updated successfully",
		zap.String("account_id", account.AccountID))

	return &account, nil
}

// GetAccountBalance retrieves account balance
func (s *GrooveStorageImpl) GetAccountBalance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	s.logger.Info("Getting account balance", zap.String("account_id", accountID))

	query := `SELECT balance FROM groove_accounts WHERE account_id = $1`

	var balance decimal.Decimal
	err := s.db.GetPool().QueryRow(ctx, query, accountID).Scan(&balance)

	if err != nil {
		s.logger.Error("Failed to get account balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("account not found: %w", err)
	}

	s.logger.Info("Account balance retrieved successfully",
		zap.String("account_id", accountID),
		zap.String("balance", balance.String()))

	return balance, nil
}

// ProcessTransaction processes a GrooveTech transaction (legacy method - use ProcessWagerTransaction instead)
func (s *GrooveStorageImpl) ProcessTransaction(ctx context.Context, transaction dto.GrooveTransaction) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing GrooveTech transaction",
		zap.String("transaction_id", transaction.TransactionID))

	// This is a legacy method - redirect to the new wager transaction method
	wagerReq := dto.GrooveWagerRequest{
		AccountID:     transaction.AccountID,
		GameSessionID: transaction.GameSessionID,
		TransactionID: transaction.TransactionID,
		RoundID:       transaction.RoundID,
		GameID:        transaction.GameID,
		BetAmount:     transaction.BetAmount,
		Device:        transaction.Device,
		UserID:        transaction.UserID,
	}

	wagerResp, err := s.ProcessWagerTransaction(ctx, wagerReq)
	if err != nil {
		return nil, err
	}

	// Convert to legacy response format
	response := &dto.GrooveTransactionResponse{
		Success:       wagerResp.Code == 200,
		TransactionID: transaction.TransactionID,
		Balance:       wagerResp.RealBalance,
		Status:        wagerResp.Status,
	}

	return response, nil
}

// ProcessWagerTransaction processes a wager transaction according to GrooveTech spec
func (s *GrooveStorageImpl) ProcessWagerTransaction(ctx context.Context, req dto.GrooveWagerRequest) (*dto.GrooveWagerResponse, error) {
	s.logger.Info("Processing wager transaction",
		zap.String("transaction_id", req.TransactionID),
		zap.String("account_id", req.AccountID),
		zap.String("bet_amount", req.BetAmount.String()))

	// Generate account transaction ID
	accountTransactionID := fmt.Sprintf("TXN_%s_%d", req.TransactionID[:8], time.Now().Unix())

	// Deduct bet amount from user balance
	newBalance, err := s.DeductBalance(ctx, req.UserID, req.BetAmount)
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

	err = s.StoreTransaction(ctx, transaction, "wager")
	if err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
		// Don't fail the transaction if storage fails, but log it
	}

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

// GetTransactionHistory retrieves transaction history
func (s *GrooveStorageImpl) GetTransactionHistory(ctx context.Context, req dto.GrooveTransactionHistoryRequest) (*dto.GrooveTransactionHistory, error) {
	s.logger.Info("Getting transaction history",
		zap.String("account_id", req.AccountID),
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize))

	// Build query with filters
	query := `
		SELECT transaction_id, account_id, session_id, amount, metadata, created_at
		FROM groove_transactions
		WHERE account_id = $1`

	args := []interface{}{req.AccountID}
	argIndex := 2

	// Add date filters
	if !req.FromDate.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, req.FromDate)
		argIndex++
	}
	if !req.ToDate.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, req.ToDate)
		argIndex++
	}

	// Add transaction type filter
	if req.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, req.Type)
		argIndex++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, req.PageSize, (req.Page-1)*req.PageSize)

	rows, err := s.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to get transaction history", zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	defer rows.Close()

	var transactions []dto.GrooveTransaction
	for rows.Next() {
		var transaction dto.GrooveTransaction
		var metadata map[string]interface{}

		err := rows.Scan(
			&transaction.TransactionID,
			&transaction.AccountID,
			&transaction.GameSessionID,
			&transaction.BetAmount,
			&metadata,
			&transaction.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan transaction", zap.Error(err))
			continue
		}

		// Extract data from metadata
		if metadata != nil {
			if accountTxnID, ok := metadata["account_transaction_id"].(string); ok {
				transaction.AccountTransactionID = accountTxnID
			}
			if roundID, ok := metadata["round_id"].(string); ok {
				transaction.RoundID = roundID
			}
			if gameID, ok := metadata["game_id"].(string); ok {
				transaction.GameID = gameID
			}
			if device, ok := metadata["device"].(string); ok {
				transaction.Device = device
			}
			if frbid, ok := metadata["frbid"].(string); ok {
				transaction.FRBID = frbid
			}
			if userIDStr, ok := metadata["user_id"].(string); ok {
				userID, err := uuid.Parse(userIDStr)
				if err == nil {
					transaction.UserID = userID
				}
			}
		}

		transactions = append(transactions, transaction)
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM groove_transactions WHERE account_id = $1`
	countArgs := []interface{}{req.AccountID}

	if !req.FromDate.IsZero() {
		countQuery += " AND created_at >= $2"
		countArgs = append(countArgs, req.FromDate)
		if !req.ToDate.IsZero() {
			countQuery += " AND created_at <= $3"
			countArgs = append(countArgs, req.ToDate)
		}
	}

	var totalCount int
	err = s.db.GetPool().QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		s.logger.Error("Failed to get transaction count", zap.Error(err))
		totalCount = len(transactions) // Fallback
	}

	history := &dto.GrooveTransactionHistory{
		AccountID:    req.AccountID,
		SessionID:    req.SessionID,
		Transactions: transactions,
		TotalCount:   totalCount,
		Page:         req.Page,
		PageSize:     req.PageSize,
		HasMore:      (req.Page * req.PageSize) < totalCount,
	}

	s.logger.Info("Transaction history retrieved successfully",
		zap.Int("count", len(transactions)),
		zap.Int("total", totalCount))

	return history, nil
}

// CreateGameSession creates a new game session
func (s *GrooveStorageImpl) CreateGameSession(ctx context.Context, session dto.GrooveGameSession) (*dto.GrooveGameSession, error) {
	s.logger.Info("Creating game session",
		zap.String("session_id", session.SessionID),
		zap.String("game_id", session.GameID))

	query := `
		INSERT INTO groove_game_sessions (id, session_id, account_id, game_id, balance, currency, status, created_at, expires_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, expires_at, last_activity`

	var createdAt, expiresAt, lastActivity time.Time
	err := s.db.GetPool().QueryRow(ctx, query,
		uuid.New(), // Internal ID
		session.SessionID,
		session.AccountID,
		session.GameID,
		session.Balance,
		session.Currency,
		session.Status,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
	).Scan(&createdAt, &expiresAt, &lastActivity)

	if err != nil {
		s.logger.Error("Failed to create game session", zap.Error(err))
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	session.CreatedAt = createdAt
	session.ExpiresAt = expiresAt
	session.LastActivity = lastActivity

	s.logger.Info("Game session created successfully",
		zap.String("session_id", session.SessionID))

	return &session, nil
}

// EndGameSession ends a game session
func (s *GrooveStorageImpl) EndGameSession(ctx context.Context, sessionID string) error {
	s.logger.Info("Ending game session", zap.String("session_id", sessionID))

	query := `UPDATE groove_game_sessions SET status = 'ended', last_activity = $1 WHERE session_id = $2`

	result, err := s.db.GetPool().Exec(ctx, query, time.Now(), sessionID)
	if err != nil {
		s.logger.Error("Failed to end game session", zap.Error(err))
		return fmt.Errorf("failed to end game session: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("game session not found")
	}

	s.logger.Info("Game session ended successfully", zap.String("session_id", sessionID))
	return nil
}

// GetUserBalance retrieves user balance from the main balances table
func (s *GrooveStorageImpl) GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	s.logger.Info("Getting user balance", zap.String("user_id", userID.String()))

	// Get balance from main balances table (source of truth) - using real_money
	query := `SELECT real_money FROM balances WHERE user_id = $1 AND currency = 'USD' LIMIT 1`

	var balanceCents int64
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(&balanceCents)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Convert cents to dollars
	balance := decimal.NewFromInt(balanceCents).Div(decimal.NewFromInt(100))

	s.logger.Info("User balance retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("balance_cents", fmt.Sprintf("%d", balanceCents)),
		zap.String("balance_dollars", balance.String()))

	return balance, nil
}

// UpdateUserBalance updates user balance in the existing balance system
func (s *GrooveStorageImpl) UpdateUserBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	s.logger.Info("Updating user balance",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
		zap.String("type", transactionType))

	// This would integrate with your existing balance system
	// For now, we'll just log the update
	s.logger.Info("Balance update logged",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
		zap.String("type", transactionType))

	return nil
}

func (s *GrooveStorageImpl) GetUserProfile(ctx context.Context, userID uuid.UUID) (*dto.GrooveUserProfile, error) {
	s.logger.Info("Getting user profile", zap.String("user_id", userID.String()))

	// Get user information from the users table
	query := `
		SELECT city, country, currency_code 
		FROM users u
		LEFT JOIN balances b ON u.id = b.user_id AND b.currency = 'USD'
		WHERE u.id = $1`

	var city, country, currencyCode *string
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(&city, &country, &currencyCode)
	if err != nil {
		s.logger.Error("Failed to get user profile", zap.Error(err))
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Set defaults for null values
	profile := &dto.GrooveUserProfile{
		City:     "Unknown",
		Country:  "US",
		Currency: "USD",
	}

	if city != nil {
		profile.City = *city
	}
	if country != nil {
		profile.Country = *country
	}
	if currencyCode != nil {
		profile.Currency = *currencyCode
	}

	s.logger.Info("User profile retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("city", profile.City),
		zap.String("country", profile.Country),
		zap.String("currency", profile.Currency))

	return profile, nil
}

// DeductBalance deducts amount from user's balance and returns new balance
func (s *GrooveStorageImpl) DeductBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error) {
	s.logger.Info("Deducting balance",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()))

	// Start transaction
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current balance from main balances table - using real_money
	var currentBalanceCents int64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(real_money, 0) 
		FROM balances 
		WHERE user_id = $1 AND currency = 'USD'
	`, userID).Scan(&currentBalanceCents)
	if err != nil {
		s.logger.Error("Failed to get current balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Convert cents to dollars
	currentBalance := decimal.NewFromInt(currentBalanceCents).Div(decimal.NewFromInt(100))

	// Check if sufficient funds
	if currentBalance.LessThan(amount) {
		s.logger.Error("Insufficient funds",
			zap.String("current_balance", currentBalance.String()),
			zap.String("requested_amount", amount.String()))
		return decimal.Zero, fmt.Errorf("insufficient funds")
	}

	// Calculate new balance
	newBalance := currentBalance.Sub(amount)
	newBalanceCents := newBalance.Mul(decimal.NewFromInt(100)).IntPart()

	// Update main balances table - using real_money
	_, err = tx.Exec(ctx, `
		INSERT INTO balances (user_id, currency, real_money, bonus_money, updated_at)
		VALUES ($1, 'USD', $2, $3, NOW())
		ON CONFLICT (user_id, currency)
		DO UPDATE SET 
			real_money = $2,
			bonus_money = $3,
			updated_at = NOW()
	`, userID, newBalanceCents, newBalance)
	if err != nil {
		s.logger.Error("Failed to update balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to update balance: %w", err)
	}

	// Synchronize groove_accounts.balance with main wallet
	_, err = tx.Exec(ctx, `
		UPDATE groove_accounts 
		SET balance = $2, last_activity = NOW()
		WHERE user_id = $1
	`, userID, newBalance)
	if err != nil {
		s.logger.Error("Failed to update balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to update balance: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Balance deducted successfully",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
		zap.String("new_balance", newBalance.String()))

	// Trigger WebSocket balance update for real-time frontend updates
	if s.userWS != nil {
		s.userWS.TriggerBalanceWS(ctx, userID)
		s.logger.Debug("WebSocket balance update triggered for user",
			zap.String("user_id", userID.String()))
	}

	return newBalance, nil
}

// AddBalance adds funds to user balance
func (s *GrooveStorageImpl) AddBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error) {
	s.logger.Info("Adding balance",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()))

	// Start transaction
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current balance from main balances table - using real_money
	var currentBalanceCents int64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(real_money, 0) 
		FROM balances 
		WHERE user_id = $1 AND currency = 'USD'
	`, userID).Scan(&currentBalanceCents)
	if err != nil {
		s.logger.Error("Failed to get current balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Convert cents to dollars
	currentBalance := decimal.NewFromInt(currentBalanceCents).Div(decimal.NewFromInt(100))

	// Calculate new balance
	newBalance := currentBalance.Add(amount)
	newBalanceCents := newBalance.Mul(decimal.NewFromInt(100)).IntPart()

	// Update main balances table - using real_money
	_, err = tx.Exec(ctx, `
		INSERT INTO balances (user_id, currency, real_money, bonus_money, updated_at)
		VALUES ($1, 'USD', $2, $3, NOW())
		ON CONFLICT (user_id, currency)
		DO UPDATE SET 
			real_money = $2,
			bonus_money = $3,
			updated_at = NOW()
	`, userID, newBalanceCents, newBalance)
	if err != nil {
		s.logger.Error("Failed to update balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to update balance: %w", err)
	}

	// Synchronize groove_accounts.balance with main wallet
	_, err = tx.Exec(ctx, `
		UPDATE groove_accounts 
		SET balance = $2, last_activity = NOW()
		WHERE user_id = $1
	`, userID, newBalance)
	if err != nil {
		s.logger.Error("Failed to update balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to update balance: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Balance added successfully",
		zap.String("user_id", userID.String()),
		zap.String("amount", amount.String()),
		zap.String("new_balance", newBalance.String()))

	// Trigger WebSocket balance update for real-time frontend updates
	if s.userWS != nil {
		s.userWS.TriggerBalanceWS(ctx, userID)
		s.logger.Debug("WebSocket balance update triggered for user",
			zap.String("user_id", userID.String()))
	}

	return newBalance, nil
}

// StoreTransaction stores a transaction for idempotency checks
func (s *GrooveStorageImpl) StoreTransaction(ctx context.Context, transaction *dto.GrooveTransaction, transactionType string) error {
	s.logger.Info("Storing transaction",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("account_transaction_id", transaction.AccountTransactionID),
		zap.String("transaction_type", transactionType))

	// Store transaction metadata in JSONB format according to existing table structure
	metadata := map[string]interface{}{
		"account_transaction_id": transaction.AccountTransactionID,
		"game_session_id":        transaction.GameSessionID,
		"round_id":               transaction.RoundID,
		"game_id":                transaction.GameID,
		"device":                 transaction.Device,
		"frbid":                  transaction.FRBID,
		"user_id":                transaction.UserID.String(),
	}

	_, err := s.db.GetPool().Exec(ctx, `
		INSERT INTO groove_transactions (
			transaction_id, account_id, session_id, amount, currency, type, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (transaction_id) DO UPDATE SET
			type = EXCLUDED.type,
			amount = EXCLUDED.amount,
			metadata = EXCLUDED.metadata,
			created_at = EXCLUDED.created_at
	`, transaction.TransactionID, transaction.AccountID, transaction.GameSessionID,
		transaction.BetAmount, "USD", transactionType, "completed", metadata)
	if err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
		return fmt.Errorf("failed to store transaction: %w", err)
	}

	s.logger.Info("Transaction stored successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("transaction_type", transactionType))

	// Sync to ClickHouse analytics
	if s.analyticsIntegration != nil {
		if err := s.analyticsIntegration.OnGrooveTransaction(ctx, transaction, transactionType); err != nil {
			s.logger.Error("Failed to sync GrooveTech transaction to ClickHouse",
				zap.String("transaction_id", transaction.TransactionID),
				zap.String("transaction_type", transactionType),
				zap.Error(err))
			// Don't fail the transaction if analytics sync fails
		} else {
			s.logger.Debug("GrooveTech transaction synced to ClickHouse successfully",
				zap.String("transaction_id", transaction.TransactionID),
				zap.String("transaction_type", transactionType))
		}
	}

	return nil
}

// GetTransactionByID retrieves a transaction by its ID
func (s *GrooveStorageImpl) GetTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error) {
	s.logger.Info("Getting transaction by ID", zap.String("transaction_id", transactionID))

	var transaction dto.GrooveTransaction
	var metadata map[string]interface{}

	err := s.db.GetPool().QueryRow(ctx, `
		SELECT transaction_id, account_id, session_id, amount, metadata, created_at
		FROM groove_transactions
		WHERE transaction_id = $1
	`, transactionID).Scan(
		&transaction.TransactionID,
		&transaction.AccountID,
		&transaction.GameSessionID,
		&transaction.BetAmount,
		&metadata,
		&transaction.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Info("Transaction not found", zap.String("transaction_id", transactionID))
			return nil, nil
		}
		s.logger.Error("Failed to get transaction by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	// Extract data from metadata
	if metadata != nil {
		if accountTxnID, ok := metadata["account_transaction_id"].(string); ok {
			transaction.AccountTransactionID = accountTxnID
		}
		if roundID, ok := metadata["round_id"].(string); ok {
			transaction.RoundID = roundID
		}
		if gameID, ok := metadata["game_id"].(string); ok {
			transaction.GameID = gameID
		}
		if device, ok := metadata["device"].(string); ok {
			transaction.Device = device
		}
		if frbid, ok := metadata["frbid"].(string); ok {
			transaction.FRBID = frbid
		}
		if userIDStr, ok := metadata["user_id"].(string); ok {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				transaction.UserID = userID
			}
		}
	}

	s.logger.Info("Transaction retrieved successfully",
		zap.String("transaction_id", transactionID),
		zap.String("account_transaction_id", transaction.AccountTransactionID))

	return &transaction, nil
}

// GetResultTransactionByID retrieves a RESULT transaction by its ID (not wager transactions)
func (s *GrooveStorageImpl) GetResultTransactionByID(ctx context.Context, transactionID string) (*dto.GrooveTransaction, error) {
	s.logger.Info("Getting result transaction by ID", zap.String("transaction_id", transactionID))

	var transaction dto.GrooveTransaction
	var metadata map[string]interface{}

	err := s.db.GetPool().QueryRow(ctx, `
		SELECT transaction_id, account_id, session_id, amount, metadata, created_at
		FROM groove_transactions
		WHERE transaction_id = $1 AND type = 'result'
	`, transactionID).Scan(
		&transaction.TransactionID,
		&transaction.AccountID,
		&transaction.GameSessionID,
		&transaction.BetAmount,
		&metadata,
		&transaction.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Info("Result transaction not found", zap.String("transaction_id", transactionID))
			return nil, nil
		}
		s.logger.Error("Failed to get result transaction by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get result transaction by ID: %w", err)
	}

	// Extract data from metadata
	if metadata != nil {
		if accountTxnID, ok := metadata["account_transaction_id"].(string); ok {
			transaction.AccountTransactionID = accountTxnID
		}
		if roundID, ok := metadata["round_id"].(string); ok {
			transaction.RoundID = roundID
		}
		if gameID, ok := metadata["game_id"].(string); ok {
			transaction.GameID = gameID
		}
		if device, ok := metadata["device"].(string); ok {
			transaction.Device = device
		}
		if frbid, ok := metadata["frbid"].(string); ok {
			transaction.FRBID = frbid
		}
		if userIDStr, ok := metadata["user_id"].(string); ok {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				transaction.UserID = userID
			}
		}
	}

	s.logger.Info("Result transaction retrieved successfully",
		zap.String("transaction_id", transactionID),
		zap.String("account_transaction_id", transaction.AccountTransactionID))

	return &transaction, nil
}

// GetWagerTransactionBySessionID retrieves a WAGER transaction by session ID
func (s *GrooveStorageImpl) GetWagerTransactionBySessionID(ctx context.Context, sessionID string) (*dto.GrooveTransaction, error) {
	s.logger.Info("Getting wager transaction by session ID", zap.String("session_id", sessionID))

	var transaction dto.GrooveTransaction
	var metadata map[string]interface{}

	err := s.db.GetPool().QueryRow(ctx, `
		SELECT transaction_id, account_id, session_id, amount, metadata, created_at
		FROM groove_transactions
		WHERE session_id = $1 AND type = 'wager'
		ORDER BY created_at DESC
		LIMIT 1
	`, sessionID).Scan(
		&transaction.TransactionID,
		&transaction.AccountID,
		&transaction.GameSessionID,
		&transaction.BetAmount,
		&metadata,
		&transaction.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.logger.Info("Wager transaction not found for session", zap.String("session_id", sessionID))
			return nil, nil
		}
		s.logger.Error("Failed to get wager transaction by session ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get wager transaction by session ID: %w", err)
	}

	// Extract data from metadata
	if metadata != nil {
		if accountTxnID, ok := metadata["account_transaction_id"].(string); ok {
			transaction.AccountTransactionID = accountTxnID
		}
		if roundID, ok := metadata["round_id"].(string); ok {
			transaction.RoundID = roundID
		}
		if gameID, ok := metadata["game_id"].(string); ok {
			transaction.GameID = gameID
		}
		if device, ok := metadata["device"].(string); ok {
			transaction.Device = device
		}
		if frbid, ok := metadata["frbid"].(string); ok {
			transaction.FRBID = frbid
		}
		if userIDStr, ok := metadata["user_id"].(string); ok {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				transaction.UserID = userID
			}
		}
	}

	s.logger.Info("Wager transaction retrieved successfully",
		zap.String("session_id", sessionID),
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("account_transaction_id", transaction.AccountTransactionID))

	return &transaction, nil
}

// ValidateBalanceSync validates if user balances are synchronized between main and GrooveTech systems
func (s *GrooveStorageImpl) ValidateBalanceSync(ctx context.Context, userID uuid.UUID) (*BalanceSyncStatus, error) {
	s.logger.Info("Validating balance synchronization", zap.String("user_id", userID.String()))

	// Get main balance (source of truth)
	mainBalance, err := s.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get main balance", zap.Error(err))
		return nil, fmt.Errorf("failed to get main balance: %w", err)
	}

	// Get GrooveTech account balance
	grooveAccount, err := s.GetAccountByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get GrooveTech account", zap.Error(err))
		return nil, fmt.Errorf("failed to get GrooveTech account: %w", err)
	}

	grooveBalance := grooveAccount.Balance

	// Calculate discrepancy
	discrepancy := mainBalance.Sub(grooveBalance)
	isSynchronized := discrepancy.Abs().LessThan(decimal.NewFromFloat(0.01)) // Allow 1 cent tolerance

	// Get last activity time from GrooveTech account
	var lastSyncTime *time.Time
	if !grooveAccount.LastActivity.IsZero() {
		lastSyncTime = &grooveAccount.LastActivity
	}

	status := &BalanceSyncStatus{
		UserID:             userID,
		MainBalance:        mainBalance,
		GrooveBalance:      grooveBalance,
		IsSynchronized:     isSynchronized,
		Discrepancy:        discrepancy,
		LastSyncTime:       lastSyncTime,
		LastValidationTime: time.Now(),
	}

	s.logger.Info("Balance synchronization validation completed",
		zap.String("user_id", userID.String()),
		zap.String("main_balance", mainBalance.String()),
		zap.String("groove_balance", grooveBalance.String()),
		zap.String("discrepancy", discrepancy.String()),
		zap.Bool("is_synchronized", isSynchronized))

	return status, nil
}

// ReconcileBalances synchronizes GrooveTech account balance with main balance
func (s *GrooveStorageImpl) ReconcileBalances(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Reconciling balances", zap.String("user_id", userID.String()))

	// Start transaction
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get main balance (source of truth) - using real_money
	var mainBalanceCents int64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(real_money, 0) 
		FROM balances 
		WHERE user_id = $1 AND currency = 'USD'
	`, userID).Scan(&mainBalanceCents)
	if err != nil {
		s.logger.Error("Failed to get main balance", zap.Error(err))
		return fmt.Errorf("failed to get main balance: %w", err)
	}

	// Convert cents to dollars
	mainBalance := decimal.NewFromInt(mainBalanceCents).Div(decimal.NewFromInt(100))

	// Update GrooveTech account balance to match main balance
	result, err := tx.Exec(ctx, `
		UPDATE groove_accounts 
		SET balance = $2, last_activity = NOW()
		WHERE user_id = $1
	`, userID, mainBalance)
	if err != nil {
		s.logger.Error("Failed to update GrooveTech balance", zap.Error(err))
		return fmt.Errorf("failed to update GrooveTech balance: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Warn("No GrooveTech account found for user", zap.String("user_id", userID.String()))
		return fmt.Errorf("no GrooveTech account found for user")
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Balance reconciliation completed successfully",
		zap.String("user_id", userID.String()),
		zap.String("main_balance", mainBalance.String()))

	return nil
}

// GetBalanceDiscrepancies finds all users with balance discrepancies
func (s *GrooveStorageImpl) GetBalanceDiscrepancies(ctx context.Context) ([]BalanceDiscrepancy, error) {
	s.logger.Info("Getting balance discrepancies")

	query := `
		SELECT 
			u.id as user_id,
			u.username,
			u.email,
			COALESCE(b.bonus_money, 0) as main_balance,
			COALESCE(ga.balance, 0) as groove_balance,
			COALESCE(b.bonus_money, 0) - COALESCE(ga.balance, 0) as discrepancy,
			ga.last_activity
		FROM users u
		LEFT JOIN balances b ON u.id = b.user_id AND b.currency = 'USD'
		LEFT JOIN groove_accounts ga ON u.id = ga.user_id
		WHERE ga.user_id IS NOT NULL
		AND ABS(COALESCE(b.bonus_money, 0) - COALESCE(ga.balance, 0)) > 0.01
		ORDER BY ABS(COALESCE(b.bonus_money, 0) - COALESCE(ga.balance, 0)) DESC`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.logger.Error("Failed to get balance discrepancies", zap.Error(err))
		return nil, fmt.Errorf("failed to get balance discrepancies: %w", err)
	}
	defer rows.Close()

	var discrepancies []BalanceDiscrepancy
	for rows.Next() {
		var discrepancy BalanceDiscrepancy
		var lastActivity *time.Time

		err := rows.Scan(
			&discrepancy.UserID,
			&discrepancy.Username,
			&discrepancy.Email,
			&discrepancy.MainBalance,
			&discrepancy.GrooveBalance,
			&discrepancy.Discrepancy,
			&lastActivity,
		)
		if err != nil {
			s.logger.Error("Failed to scan discrepancy", zap.Error(err))
			continue
		}

		discrepancy.LastActivity = lastActivity
		discrepancy.DiscrepancyDetected = time.Now()

		discrepancies = append(discrepancies, discrepancy)
	}

	s.logger.Info("Balance discrepancies retrieved successfully",
		zap.Int("count", len(discrepancies)))

	return discrepancies, nil
}

// GetTransactionGameInfo retrieves game information from a transaction
func (s *GrooveStorageImpl) GetTransactionGameInfo(ctx context.Context, transactionID string) (gameID, gameType string, err error) {
	s.logger.Info("Getting transaction game info", zap.String("transaction_id", transactionID))

	var metadata map[string]interface{}

	query := `
		SELECT metadata 
		FROM groove_transactions 
		WHERE transaction_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`

	err = s.db.GetPool().QueryRow(ctx, query, transactionID).Scan(&metadata)
	if err != nil {
		s.logger.Warn("Failed to get transaction game info",
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return "", "", fmt.Errorf("failed to get transaction game info: %w", err)
	}

	// Extract game information from metadata
	if metadata != nil {
		if gameIDStr, ok := metadata["game_id"].(string); ok && gameIDStr != "" {
			gameID = gameIDStr
		}
		// Default game type for GrooveTech
		gameType = "groovetech"
	}

	s.logger.Info("Retrieved transaction game info",
		zap.String("transaction_id", transactionID),
		zap.String("game_id", gameID),
		zap.String("game_type", gameType))

	return gameID, gameType, nil
}
