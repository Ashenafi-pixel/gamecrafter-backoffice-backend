package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/db"
	"go.uber.org/zap"
)

// GetBlockedAccountByUserIDWithPagination retrieves blocked accounts for a specific user with pagination
func (u *user) GetBlockedAccountByUserIDWithPagination(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.SuspensionHistory, error) {
	blockedAccounts, err := u.db.Queries.GetBlockedAccountByUserID(ctx, db.GetBlockedAccountByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		u.log.Error("Failed to get blocked accounts by user ID", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	var result []dto.SuspensionHistory
	for _, account := range blockedAccounts {
		var blockedFrom, blockedTo, unblockedAt *time.Time
		if account.BlockedFrom.Valid {
			blockedFrom = &account.BlockedFrom.Time
		}
		if account.BlockedTo.Valid {
			blockedTo = &account.BlockedTo.Time
		}
		if account.UnblockedAt.Valid {
			unblockedAt = &account.UnblockedAt.Time
		}

		result = append(result, dto.SuspensionHistory{
			ID:                account.ID,
			UserID:            account.UserID,
			BlockedBy:         account.BlockedBy,
			Duration:          account.Duration,
			Type:              account.Type,
			BlockedFrom:       blockedFrom,
			BlockedTo:         blockedTo,
			UnblockedAt:       unblockedAt,
			Reason:            account.Reason.String,
			Note:              account.Note.String,
			CreatedAt:         account.CreatedAt.Time,
			BlockedByUsername: account.BlockerAccountUserUsername.String,
			BlockedByEmail:    account.BlockerAccountUserEmail.String,
		})
	}

	return result, nil
}

// GetBalanceLogsByUserID retrieves balance logs for a specific user with pagination
func (u *user) GetBalanceLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.BalanceLog, error) {
	// This would need to be implemented in the balance logs storage
	// For now, return empty array as the balance logs storage doesn't have this method yet
	u.log.Info("GetBalanceLogsByUserID not yet implemented in balance logs storage", zap.String("user_id", userID.String()))

	// TODO: Implement this method in balance logs storage
	// balanceLogs, err := u.balanceLogsStorage.GetBalanceLogsByUserID(ctx, userID, limit, offset)

	return []dto.BalanceLog{}, nil
}
