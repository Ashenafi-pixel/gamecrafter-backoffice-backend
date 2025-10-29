package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type RealtimeSyncService interface {
	StartRealtimeSync(ctx context.Context) error
	StopRealtimeSync() error
	SyncNewTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error
	SyncNewBalance(ctx context.Context, userID uuid.UUID, balance decimal.Decimal, transactionID string, transactionType string) error
	SyncNewGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error
	SyncNewCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error
}

type RealtimeSyncServiceImpl struct {
	syncService      SyncService
	analyticsStorage storage.Analytics
	logger           *zap.Logger
	syncTicker       *time.Ticker
	stopChan         chan bool
	isRunning        bool
}

func NewRealtimeSyncService(syncService SyncService, analyticsStorage storage.Analytics, logger *zap.Logger) RealtimeSyncService {
	return &RealtimeSyncServiceImpl{
		syncService:      syncService,
		analyticsStorage: analyticsStorage,
		logger:           logger,
		stopChan:         make(chan bool),
		isRunning:        false,
	}
}

func (r *RealtimeSyncServiceImpl) StartRealtimeSync(ctx context.Context) error {
	if r.isRunning {
		return nil
	}

	r.logger.Info("Starting real-time ClickHouse sync service")

	// Start ticker for periodic sync (every 30 seconds)
	r.syncTicker = time.NewTicker(30 * time.Second)
	r.isRunning = true

	go func() {
		for {
			select {
			case <-r.syncTicker.C:
				r.performPeriodicSync(ctx)
			case <-r.stopChan:
				r.logger.Info("Real-time sync service stopped")
				return
			case <-ctx.Done():
				r.logger.Info("Real-time sync service stopped due to context cancellation")
				return
			}
		}
	}()

	r.logger.Info("Real-time ClickHouse sync service started successfully")
	return nil
}

func (r *RealtimeSyncServiceImpl) StopRealtimeSync() error {
	if !r.isRunning {
		return nil
	}

	r.logger.Info("Stopping real-time ClickHouse sync service")

	if r.syncTicker != nil {
		r.syncTicker.Stop()
	}

	r.stopChan <- true
	r.isRunning = false

	r.logger.Info("Real-time ClickHouse sync service stopped")
	return nil
}

func (r *RealtimeSyncServiceImpl) SyncNewTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error {
	if err := r.syncService.SyncTransaction(ctx, transaction); err != nil {
		r.logger.Error("Failed to sync transaction in real-time",
			zap.String("transaction_id", transaction.ID),
			zap.String("user_id", transaction.UserID.String()),
			zap.Error(err))
		return err
	}

	r.logger.Debug("Transaction synced to ClickHouse in real-time",
		zap.String("transaction_id", transaction.ID),
		zap.String("user_id", transaction.UserID.String()))

	return nil
}

func (r *RealtimeSyncServiceImpl) SyncNewBalance(ctx context.Context, userID uuid.UUID, balance decimal.Decimal, transactionID string, transactionType string) error {
	if err := r.syncService.SyncUserBalance(ctx, userID, balance, transactionID, transactionType); err != nil {
		r.logger.Error("Failed to sync balance in real-time",
			zap.String("user_id", userID.String()),
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return err
	}

	r.logger.Debug("Balance synced to ClickHouse in real-time",
		zap.String("user_id", userID.String()),
		zap.String("balance", balance.String()))

	return nil
}

func (r *RealtimeSyncServiceImpl) SyncNewGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error {
	if err := r.syncService.SyncGrooveTransaction(ctx, grooveTx, transactionType); err != nil {
		r.logger.Error("Failed to sync GrooveTech transaction in real-time",
			zap.String("transaction_id", grooveTx.TransactionID),
			zap.String("account_id", grooveTx.AccountID),
			zap.Error(err))
		return err
	}

	r.logger.Debug("GrooveTech transaction synced to ClickHouse in real-time",
		zap.String("transaction_id", grooveTx.TransactionID),
		zap.String("account_id", grooveTx.AccountID))

	return nil
}

func (r *RealtimeSyncServiceImpl) SyncNewCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error {
	if err := r.syncService.SyncCashbackEarning(ctx, earning); err != nil {
		r.logger.Error("Failed to sync cashback earning in real-time",
			zap.String("earning_id", earning.ID.String()),
			zap.String("user_id", earning.UserID.String()),
			zap.Error(err))
		return err
	}

	r.logger.Debug("Cashback earning synced to ClickHouse in real-time",
		zap.String("earning_id", earning.ID.String()),
		zap.String("user_id", earning.UserID.String()),
		zap.String("amount", earning.EarnedAmount.String()))

	return nil
}

func (r *RealtimeSyncServiceImpl) performPeriodicSync(ctx context.Context) {
	// This method can be used for periodic cleanup, optimization, or batch processing
	// For now, it's a placeholder for future enhancements

	r.logger.Debug("Performing periodic ClickHouse sync maintenance")

	// Example: Could perform table optimization, cleanup, or batch processing here
	// For now, just log that the periodic sync is running
}

// Helper function to convert balance log to analytics transaction
func ConvertBalanceLogToAnalyticsTransaction(balanceLog interface{}, userID uuid.UUID) *dto.AnalyticsTransaction {
	// This would be implemented based on your balance_logs table structure
	// For now, returning a placeholder
	return &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "deposit", // Would be determined from balance_log
		Amount:          decimal.Zero,
		Currency:        "USD",
		Status:          "completed",
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// Helper function to convert GrooveTech transaction to analytics transaction
func ConvertGrooveTransactionToAnalytics(grooveTx *dto.GrooveTransaction, transactionType string) *dto.AnalyticsTransaction {
	return &dto.AnalyticsTransaction{
		ID:                    grooveTx.TransactionID,
		UserID:                uuid.MustParse(grooveTx.AccountID),
		TransactionType:       mapGrooveTransactionType(transactionType),
		Amount:                grooveTx.BetAmount,
		Currency:              "USD",
		Status:                "completed",
		GameID:                &grooveTx.GameID,
		SessionID:             &grooveTx.GameSessionID,
		RoundID:               &grooveTx.RoundID,
		BalanceBefore:         decimal.Zero,
		BalanceAfter:          decimal.Zero,
		ExternalTransactionID: &grooveTx.AccountTransactionID,
		CreatedAt:             grooveTx.CreatedAt,
		UpdatedAt:             grooveTx.CreatedAt,
	}
}

func mapGrooveTransactionType(grooveType string) string {
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
