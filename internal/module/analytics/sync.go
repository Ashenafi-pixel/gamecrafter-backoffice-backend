package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type SyncService interface {
	SyncTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error
	SyncTransactionsBatch(ctx context.Context, transactions []*dto.AnalyticsTransaction) error
	SyncUserBalance(ctx context.Context, userID uuid.UUID, balance decimal.Decimal, transactionID string, transactionType string) error
	SyncGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error
	SyncCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error
}

type SyncServiceImpl struct {
	analyticsStorage storage.Analytics
	gameInfoProvider storage.GameInfoProvider
	logger           *zap.Logger
}

func NewSyncService(analyticsStorage storage.Analytics, gameInfoProvider storage.GameInfoProvider, logger *zap.Logger) SyncService {
	return &SyncServiceImpl{
		analyticsStorage: analyticsStorage,
		gameInfoProvider: gameInfoProvider,
		logger:           logger,
	}
}

func (s *SyncServiceImpl) SyncTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error {
	if err := s.analyticsStorage.InsertTransaction(ctx, transaction); err != nil {
		s.logger.Error("Failed to sync transaction to ClickHouse",
			zap.String("transaction_id", transaction.ID),
			zap.String("user_id", transaction.UserID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to sync transaction: %w", err)
	}

	s.logger.Debug("Transaction synced to ClickHouse successfully",
		zap.String("transaction_id", transaction.ID),
		zap.String("user_id", transaction.UserID.String()))

	return nil
}

func (s *SyncServiceImpl) SyncTransactionsBatch(ctx context.Context, transactions []*dto.AnalyticsTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	if err := s.analyticsStorage.InsertTransactions(ctx, transactions); err != nil {
		s.logger.Error("Failed to sync transactions batch to ClickHouse",
			zap.Int("count", len(transactions)),
			zap.Error(err))
		return fmt.Errorf("failed to sync transactions batch: %w", err)
	}

	s.logger.Info("Transactions batch synced to ClickHouse successfully",
		zap.Int("count", len(transactions)))

	return nil
}

func (s *SyncServiceImpl) SyncUserBalance(ctx context.Context, userID uuid.UUID, balance decimal.Decimal, transactionID string, transactionType string) error {
	// Create balance snapshot
	snapshot := &dto.BalanceSnapshot{
		UserID:          userID,
		Balance:         balance,
		Currency:        "USD", // Default currency
		SnapshotTime:    time.Now(),
		TransactionID:   &transactionID,
		TransactionType: &transactionType,
	}

	// Insert balance snapshot
	if err := s.analyticsStorage.InsertBalanceSnapshot(ctx, snapshot); err != nil {
		s.logger.Error("Failed to sync user balance to ClickHouse",
			zap.String("user_id", userID.String()),
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return fmt.Errorf("failed to sync user balance: %w", err)
	}

	s.logger.Debug("User balance synced to ClickHouse successfully",
		zap.String("user_id", userID.String()),
		zap.String("balance", balance.String()),
		zap.String("transaction_id", transactionID))

	return nil
}

func (s *SyncServiceImpl) SyncGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error {
	// Get game information - try to get game_id from session if not provided
	gameID := grooveTx.GameID
	if gameID == "" && grooveTx.GameSessionID != "" {
		// Try to get game_id from session_id
		if sessionGameID, err := s.getGameIDFromSession(ctx, grooveTx.GameSessionID); err == nil && sessionGameID != "" {
			gameID = sessionGameID
			s.logger.Debug("Retrieved game_id from session",
				zap.String("session_id", grooveTx.GameSessionID),
				zap.String("game_id", gameID))
		}
	}

	var gameName, provider *string
	if gameID != "" {
		if gameInfo, err := s.gameInfoProvider.GetGameInfo(ctx, gameID); err == nil {
			gameName = &gameInfo.GameName
			provider = &gameInfo.Provider
		} else {
			s.logger.Warn("Failed to get game info for analytics",
				zap.String("game_id", gameID),
				zap.Error(err))
		}
	}

	// Convert GrooveTech transaction to analytics transaction
	transaction := &dto.AnalyticsTransaction{
		ID:                    grooveTx.TransactionID,
		UserID:                uuid.MustParse(grooveTx.AccountID),
		TransactionType:       s.mapGrooveTransactionType(transactionType),
		Amount:                grooveTx.BetAmount,
		Currency:              "USD", // Default currency
		Status:                "completed",
		GameID:                &gameID,
		GameName:              gameName,
		Provider:              provider,
		SessionID:             &grooveTx.GameSessionID,
		RoundID:               &grooveTx.RoundID,
		BalanceBefore:         grooveTx.BalanceBefore,
		BalanceAfter:          grooveTx.BalanceAfter,
		ExternalTransactionID: &grooveTx.AccountTransactionID,
		CreatedAt:             grooveTx.CreatedAt,
		UpdatedAt:             grooveTx.CreatedAt,
	}

	// Add metadata
	if metadata, err := json.Marshal(map[string]interface{}{
		"groove_transaction": true,
		"game_status":        grooveTx.Status,
		"round_id":           grooveTx.RoundID,
	}); err == nil {
		metadataStr := string(metadata)
		transaction.Metadata = &metadataStr
	}

	// Set bet/win amounts based on transaction type
	switch transactionType {
	case "wager":
		betAmount := grooveTx.BetAmount
		transaction.BetAmount = &betAmount
		negAmount := grooveTx.BetAmount.Neg()
		transaction.NetResult = &negAmount // Negative for bets
	case "result":
		winAmount := grooveTx.BetAmount
		transaction.WinAmount = &winAmount
		netResult := grooveTx.BetAmount
		transaction.NetResult = &netResult // Positive for wins
	case "rollback":
		rollbackAmount := grooveTx.BetAmount
		transaction.NetResult = &rollbackAmount // Positive for rollbacks
	}

	if err := s.analyticsStorage.InsertTransaction(ctx, transaction); err != nil {
		s.logger.Error("Failed to sync GrooveTech transaction to ClickHouse",
			zap.String("transaction_id", grooveTx.TransactionID),
			zap.String("account_id", grooveTx.AccountID),
			zap.Error(err))
		return fmt.Errorf("failed to sync GrooveTech transaction: %w", err)
	}

	s.logger.Debug("GrooveTech transaction synced to ClickHouse successfully",
		zap.String("transaction_id", grooveTx.TransactionID),
		zap.String("account_id", grooveTx.AccountID),
		zap.String("type", transactionType))

	return nil
}

func (s *SyncServiceImpl) SyncCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error {
	// Create analytics transaction for cashback earning
	transaction := &dto.AnalyticsTransaction{
		ID:              earning.ID.String(),
		UserID:          earning.UserID,
		TransactionType: "cashback",
		Amount:          earning.EarnedAmount,
		Currency:        "USD", // Default currency
		Status:          "completed",
		BalanceBefore:   decimal.Zero, // We don't track balance before/after for cashback
		BalanceAfter:    decimal.Zero,
		GameID:          nil, // Cashback doesn't have a specific game
		RoundID:         nil, // Cashback doesn't have a specific round
		SessionID:       nil, // Cashback doesn't have a specific session
		CreatedAt:       earning.CreatedAt,
		UpdatedAt:       earning.UpdatedAt,
	}

	if err := s.analyticsStorage.InsertTransaction(ctx, transaction); err != nil {
		s.logger.Error("Failed to sync cashback earning to ClickHouse",
			zap.String("earning_id", earning.ID.String()),
			zap.String("user_id", earning.UserID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to sync cashback earning: %w", err)
	}

	s.logger.Debug("Cashback earning synced to ClickHouse successfully",
		zap.String("earning_id", earning.ID.String()),
		zap.String("user_id", earning.UserID.String()),
		zap.String("amount", earning.EarnedAmount.String()))

	return nil
}

func (s *SyncServiceImpl) mapGrooveTransactionType(grooveType string) string {
	switch grooveType {
	case "wager":
		return "groove_bet"
	case "result":
		return "groove_win"
	case "rollback":
		return "refund"
	default:
		return "groove_bet"
	}
}

// getGameIDFromSession retrieves game_id from session_id using the game_sessions table
func (s *SyncServiceImpl) getGameIDFromSession(ctx context.Context, sessionID string) (string, error) {
	// This would need to be implemented with database access
	// For now, we'll return empty string to avoid breaking the build
	// TODO: Implement database lookup for game_id from session_id
	s.logger.Debug("getGameIDFromSession called but not implemented",
		zap.String("session_id", sessionID))
	return "", fmt.Errorf("getGameIDFromSession not implemented")
}

// Helper function to convert PostgreSQL transaction to ClickHouse transaction
func ConvertPostgresTransactionToAnalytics(pgTx interface{}, userID uuid.UUID) *dto.AnalyticsTransaction {
	// This would be implemented based on your PostgreSQL transaction structure
	// For now, returning a placeholder
	return &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "deposit", // Placeholder
		Amount:          decimal.Zero,
		Currency:        "USD",
		Status:          "completed",
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
