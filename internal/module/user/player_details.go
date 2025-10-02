package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// GetUserByID retrieves a user by their ID
func (u *User) GetUserByID(ctx context.Context, userID uuid.UUID) (dto.User, bool, error) {
	return u.userStorage.GetUserByID(ctx, userID)
}

// GetPlayerSuspensionHistory retrieves the suspension history for a player
func (u *User) GetPlayerSuspensionHistory(ctx context.Context, userID uuid.UUID) ([]dto.SuspensionHistory, error) {
	// Get suspension history from account_block table
	blockedAccounts, err := u.userStorage.GetBlockedAccountByUserIDWithPagination(ctx, userID, 100, 0)
	if err != nil {
		u.log.Error("Failed to get suspension history", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	return blockedAccounts, nil
}

// GetPlayerBalanceLogs retrieves the balance transaction logs for a player
func (u *User) GetPlayerBalanceLogs(ctx context.Context, userID uuid.UUID) ([]dto.BalanceLog, error) {
	// Get balance logs from balance_logs table
	balanceLogs, err := u.userStorage.GetBalanceLogsByUserID(ctx, userID, 100, 0)
	if err != nil {
		u.log.Error("Failed to get balance logs", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	var logs []dto.BalanceLog
	for _, log := range balanceLogs {
		logs = append(logs, dto.BalanceLog{
			ID:                  log.ID,
			UserID:              log.UserID,
			Component:           log.Component,
			Currency:            log.Currency,
			ChangeAmount:        log.ChangeAmount,
			OperationalGroupID:  log.OperationalGroupID,
			OperationalTypeID:   log.OperationalTypeID,
			Description:         log.Description,
			Timestamp:           log.Timestamp,
			BalanceAfterUpdate:  log.BalanceAfterUpdate,
			TransactionID:       log.TransactionID,
			Status:              log.Status,
			Type:                log.Type,
			OperationalTypeName: log.OperationalTypeName,
		})
	}

	return logs, nil
}

// GetPlayerBalances retrieves the current balances for a player
func (u *User) GetPlayerBalances(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error) {
	// Get current balances
	balances, err := u.balanceStorage.GetBalancesByUserID(ctx, userID)
	if err != nil {
		u.log.Error("Failed to get player balances", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	return balances, nil
}

// GetPlayerGameActivity retrieves the game activity for a player
func (u *User) GetPlayerGameActivity(ctx context.Context, userID uuid.UUID) ([]dto.GameActivity, error) {
	// For now, return empty array as game activity data might come from ClickHouse or other sources
	// This can be implemented later when game activity data is available
	u.log.Info("Game activity not yet implemented", zap.String("user_id", userID.String()))

	// Return mock data for now - this should be replaced with real data from ClickHouse
	gameActivity := []dto.GameActivity{
		{
			Game:         "Bitcoin Slots",
			Provider:     "Pragmatic Play",
			Sessions:     23,
			TotalWagered: decimal.NewFromInt(45000),
			NetResult:    decimal.NewFromInt(-2300),
			LastPlayed:   time.Now().Add(-24 * time.Hour),
			FavoriteGame: true,
		},
		{
			Game:         "Ethereum Poker",
			Provider:     "Evolution Gaming",
			Sessions:     18,
			TotalWagered: decimal.NewFromInt(38000),
			NetResult:    decimal.NewFromInt(1200),
			LastPlayed:   time.Now().Add(-48 * time.Hour),
			FavoriteGame: false,
		},
	}

	return gameActivity, nil
}
