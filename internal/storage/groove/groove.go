package groove

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type GrooveStorage interface {
	// Account operations
	CreateAccount(ctx context.Context, account dto.GrooveAccount, userID uuid.UUID) (*dto.GrooveAccount, error)
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*dto.GrooveAccount, error)
	GetAccountByID(ctx context.Context, accountID string) (*dto.GrooveAccount, error)
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
}

type GrooveStorageImpl struct {
	db     *persistencedb.PersistenceDB
	logger *zap.Logger
}

func NewGrooveStorage(db *persistencedb.PersistenceDB, logger *zap.Logger) GrooveStorage {
	return &GrooveStorageImpl{
		db:     db,
		logger: logger,
	}
}

// CreateAccount creates a new GrooveTech account
func (s *GrooveStorageImpl) CreateAccount(ctx context.Context, account dto.GrooveAccount, userID uuid.UUID) (*dto.GrooveAccount, error) {
	s.logger.Info("Creating GrooveTech account",
		zap.String("account_id", account.AccountID),
		zap.String("user_id", userID.String()))

	query := `
		INSERT INTO groove_accounts (id, user_id, account_id, session_id, balance, currency, status, created_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, last_activity`

	var createdAt, lastActivity time.Time
	err := s.db.GetPool().QueryRow(ctx, query,
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
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(
		&account.AccountID,
		&account.SessionID,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.LastActivity,
	)

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
		SELECT account_id, session_id, balance, currency, status, created_at, last_activity
		FROM groove_accounts
		WHERE account_id = $1`

	var account dto.GrooveAccount
	err := s.db.GetPool().QueryRow(ctx, query, accountID).Scan(
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

// ProcessTransaction processes a GrooveTech transaction
func (s *GrooveStorageImpl) ProcessTransaction(ctx context.Context, transaction dto.GrooveTransaction) (*dto.GrooveTransactionResponse, error) {
	s.logger.Info("Processing GrooveTech transaction",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("type", transaction.Type))

	// Start transaction
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current account balance
	var currentBalance decimal.Decimal
	err = tx.QueryRow(ctx, `SELECT balance FROM groove_accounts WHERE account_id = $1`, transaction.AccountID).Scan(&currentBalance)
	if err != nil {
		s.logger.Error("Failed to get current balance", zap.Error(err))
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Calculate new balance
	var newBalance decimal.Decimal
	switch transaction.Type {
	case "debit":
		newBalance = currentBalance.Sub(transaction.Amount)
	case "credit":
		newBalance = currentBalance.Add(transaction.Amount)
	default:
		return nil, fmt.Errorf("invalid transaction type: %s", transaction.Type)
	}

	// Update account balance
	_, err = tx.Exec(ctx,
		`UPDATE groove_accounts SET balance = $1, last_activity = $2 WHERE account_id = $3`,
		newBalance, time.Now(), transaction.AccountID)
	if err != nil {
		s.logger.Error("Failed to update account balance", zap.Error(err))
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Insert transaction record
	_, err = tx.Exec(ctx, `
		INSERT INTO groove_transactions (id, transaction_id, account_id, session_id, amount, currency, type, status, created_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		uuid.New(), transaction.TransactionID, transaction.AccountID, transaction.SessionID,
		transaction.Amount, transaction.Currency, transaction.Type, transaction.Status,
		transaction.CreatedAt, nil)
	if err != nil {
		s.logger.Error("Failed to insert transaction record", zap.Error(err))
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	response := &dto.GrooveTransactionResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		Balance:       newBalance,
		Status:        "completed",
	}

	s.logger.Info("GrooveTech transaction processed successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("new_balance", newBalance.String()))

	return response, nil
}

// GetTransactionHistory retrieves transaction history
func (s *GrooveStorageImpl) GetTransactionHistory(ctx context.Context, req dto.GrooveTransactionHistoryRequest) (*dto.GrooveTransactionHistory, error) {
	s.logger.Info("Getting transaction history",
		zap.String("account_id", req.AccountID),
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize))

	// Build query with filters
	query := `
		SELECT transaction_id, account_id, session_id, amount, currency, type, status, created_at, metadata
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
		var metadata string

		err := rows.Scan(
			&transaction.TransactionID,
			&transaction.AccountID,
			&transaction.SessionID,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Type,
			&transaction.Status,
			&transaction.CreatedAt,
			&metadata,
		)
		if err != nil {
			s.logger.Error("Failed to scan transaction", zap.Error(err))
			continue
		}

		// Note: Metadata field not available in GrooveTransaction struct
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

// GetUserBalance retrieves user balance from the existing balance system
func (s *GrooveStorageImpl) GetUserBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	s.logger.Info("Getting user balance", zap.String("user_id", userID.String()))

	// Get balance from existing balance system
	query := `SELECT amount_units FROM balances WHERE user_id = $1 AND currency_code = 'USD' LIMIT 1`

	var balance decimal.Decimal
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(&balance)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return decimal.Zero, fmt.Errorf("failed to get user balance: %w", err)
	}

	s.logger.Info("User balance retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.String("balance", balance.String()))

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
