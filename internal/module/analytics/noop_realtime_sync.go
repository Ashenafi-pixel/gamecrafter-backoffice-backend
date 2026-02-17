package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
)

// noopRealtimeSync implements RealtimeSyncService with no-op methods (used when ClickHouse is disabled).
type noopRealtimeSync struct{}

// NewNoopRealtimeSyncService returns a RealtimeSyncService that does nothing.
func NewNoopRealtimeSyncService() RealtimeSyncService {
	return &noopRealtimeSync{}
}

func (n *noopRealtimeSync) StartRealtimeSync(ctx context.Context) error { return nil }
func (n *noopRealtimeSync) StopRealtimeSync() error                     { return nil }
func (n *noopRealtimeSync) SyncNewTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error {
	return nil
}
func (n *noopRealtimeSync) SyncNewBalance(ctx context.Context, userID uuid.UUID, balance decimal.Decimal, transactionID string, transactionType string) error {
	return nil
}
func (n *noopRealtimeSync) SyncNewGrooveTransaction(ctx context.Context, grooveTx *dto.GrooveTransaction, transactionType string) error {
	return nil
}
func (n *noopRealtimeSync) SyncNewCashbackEarning(ctx context.Context, earning *dto.CashbackEarning) error {
	return nil
}
