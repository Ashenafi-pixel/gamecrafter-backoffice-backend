package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/analytics"
	"go.uber.org/zap"
)

// provides hooks to integrate ClickHouse sync with existing services
type AnalyticsIntegration struct {
	realtimeSync analytics.RealtimeSyncService
	logger       *zap.Logger
}

func NewAnalyticsIntegration(realtimeSync analytics.RealtimeSyncService, logger *zap.Logger) *AnalyticsIntegration {
	return &AnalyticsIntegration{
		realtimeSync: realtimeSync,
		logger:       logger,
	}
}

// Hook for user registration
func (ai *AnalyticsIntegration) OnUserRegistration(ctx context.Context, userID uuid.UUID, registrationData map[string]interface{}) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "registration",
		Amount:          decimal.Zero,
		Currency:        "USD",
		Status:          "completed",
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		PaymentMethod:   stringPtr("registration"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Add registration metadata
	if metadata, err := json.Marshal(registrationData); err == nil {
		metadataStr := string(metadata)
		transaction.Metadata = &metadataStr
	}

	return ai.realtimeSync.SyncNewTransaction(ctx, transaction)
}

// Hook for balance changes
func (ai *AnalyticsIntegration) OnBalanceChange(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, transactionType string, transactionID string) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              transactionID,
		UserID:          userID,
		TransactionType: transactionType,
		Amount:          amount,
		Currency:        "USD",
		Status:          "completed",
		BalanceBefore:   decimal.Zero, // Would be fetched from current balance
		BalanceAfter:    decimal.Zero, // Would be calculated
		PaymentMethod:   stringPtr("balance_change"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Sync transaction
	if err := ai.realtimeSync.SyncNewTransaction(ctx, transaction); err != nil {
		return err
	}

	// Sync balance snapshot
	return ai.realtimeSync.SyncNewBalance(ctx, userID, amount, transactionID, transactionType)
}

// Hook for GrooveTech transactions
func (ai *AnalyticsIntegration) OnGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error {
	return ai.realtimeSync.SyncNewGrooveTransaction(ctx, grooveTx, transactionType)
}

// Hook for cashback earnings
func (ai *AnalyticsIntegration) OnCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error {
	return ai.realtimeSync.SyncNewCashbackEarning(ctx, earning)
}

// Hook for game sessions
func (ai *AnalyticsIntegration) OnGameSessionStart(ctx context.Context, userID uuid.UUID, gameID string, sessionID string) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "session_start",
		Amount:          decimal.Zero,
		Currency:        "USD",
		Status:          "completed",
		GameID:          &gameID,
		SessionID:       &sessionID,
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ai.realtimeSync.SyncNewTransaction(ctx, transaction)
}

// Hook for game sessions end
func (ai *AnalyticsIntegration) OnGameSessionEnd(ctx context.Context, userID uuid.UUID, gameID string, sessionID string, duration time.Duration) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "session_end",
		Amount:          decimal.Zero,
		Currency:        "USD",
		Status:          "completed",
		GameID:          &gameID,
		SessionID:       &sessionID,
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Add session metadata
	if metadata, err := json.Marshal(map[string]interface{}{
		"session_duration_seconds": duration.Seconds(),
		"session_end":              true,
	}); err == nil {
		metadataStr := string(metadata)
		transaction.Metadata = &metadataStr
	}

	return ai.realtimeSync.SyncNewTransaction(ctx, transaction)
}

// Hook for bet placement
func (ai *AnalyticsIntegration) OnBetPlaced(ctx context.Context, userID uuid.UUID, gameID string, sessionID string, betAmount decimal.Decimal, roundID string) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "bet",
		Amount:          betAmount,
		Currency:        "USD",
		Status:          "completed",
		GameID:          &gameID,
		SessionID:       &sessionID,
		RoundID:         &roundID,
		BetAmount:       &betAmount,
		NetResult:       func() *decimal.Decimal { neg := betAmount.Neg(); return &neg }(), // Negative for bets
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ai.realtimeSync.SyncNewTransaction(ctx, transaction)
}

// Hook for win
func (ai *AnalyticsIntegration) OnWin(ctx context.Context, userID uuid.UUID, gameID string, sessionID string, winAmount decimal.Decimal, roundID string) error {
	transaction := &dto.AnalyticsTransaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		TransactionType: "win",
		Amount:          winAmount,
		Currency:        "USD",
		Status:          "completed",
		GameID:          &gameID,
		SessionID:       &sessionID,
		RoundID:         &roundID,
		WinAmount:       &winAmount,
		NetResult:       &winAmount, // Positive for wins
		BalanceBefore:   decimal.Zero,
		BalanceAfter:    decimal.Zero,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return ai.realtimeSync.SyncNewTransaction(ctx, transaction)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create decimal pointer
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
