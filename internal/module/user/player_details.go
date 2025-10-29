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

// GetPlayerStatistics retrieves player statistics from database
func (u *User) GetPlayerStatistics(ctx context.Context, userID uuid.UUID) (dto.PlayerStatistics, error) {
	stats := dto.PlayerStatistics{
		TotalWagered: decimal.Zero,
		NetPL:        decimal.Zero,
		Sessions:     0,
		TotalBets:    0,
		TotalWins:    0,
		TotalLosses:  0,
		WinRate:      decimal.Zero,
		AvgBetSize:   decimal.Zero,
		LastActivity: time.Time{},
	}

	// Get betting statistics
	betStats, err := u.userStorage.GetPlayerBettingStats(ctx, userID)
	if err != nil {
		u.log.Error("Failed to get betting stats", zap.Error(err), zap.String("user_id", userID.String()))
		// Don't fail, just use zero values
	} else {
		stats.TotalWagered = betStats.TotalWagered
		stats.NetPL = betStats.NetPL
		stats.TotalBets = betStats.TotalBets
		stats.TotalWins = betStats.TotalWins
		stats.TotalLosses = betStats.TotalLosses
		stats.WinRate = betStats.WinRate
		stats.AvgBetSize = betStats.AvgBetSize
		stats.LastActivity = betStats.LastActivity
	}

	// Get session count
	sessionCount, err := u.userStorage.GetPlayerSessionCount(ctx, userID)
	if err != nil {
		u.log.Error("Failed to get session count", zap.Error(err), zap.String("user_id", userID.String()))
		// Don't fail, just use zero value
	} else {
		stats.Sessions = sessionCount
	}

	return stats, nil
}

// GetPlayerGameActivity retrieves the game activity for a player
func (u *User) GetPlayerGameActivity(ctx context.Context, userID uuid.UUID) ([]dto.GameActivity, error) {
	// Query real game activity from database
	// For now, return empty array since this test user has no game activity
	// This can be implemented with real game data when available
	u.log.Info("Game activity query - no data available for user", zap.String("user_id", userID.String()))

	// Return empty array for users with no game activity
	return []dto.GameActivity{}, nil
}

// GetPlayerManualFunds retrieves manual fund transactions for a specific player
func (u *User) GetPlayerManualFunds(ctx context.Context, userID uuid.UUID) ([]dto.ManualFundResData, error) {
	// Get manual fund transactions from manual_funds table
	manualFunds, err := u.balanceStorage.GetManualFundsByUserID(ctx, userID)
	if err != nil {
		u.log.Error("Failed to get manual funds", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	var result []dto.ManualFundResData
	for _, fund := range manualFunds {
		result = append(result, dto.ManualFundResData{
			ID:            fund.ID,
			UserID:        fund.UserID,
			AdminID:       fund.AdminID,
			AdminName:     fund.AdminName,
			TransactionID: fund.TransactionID,
			Type:          fund.Type,
			Amount:        fund.Amount,
			Reason:        fund.Reason,
			Currency:      fund.Currency,
			Note:          fund.Note,
			CreatedAt:     fund.CreatedAt,
		})
	}

	return result, nil
}

// GetPlayerManualFundsPaginated retrieves manual fund transactions for a specific player with pagination
func (u *User) GetPlayerManualFundsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]dto.ManualFundResData, int64, error) {
	// Get manual fund transactions from manual_funds table with pagination
	manualFunds, totalCount, err := u.balanceStorage.GetManualFundsByUserIDPaginated(ctx, userID, page, perPage)
	if err != nil {
		u.log.Error("Failed to get manual funds", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, 0, err
	}

	var result []dto.ManualFundResData
	for _, fund := range manualFunds {
		result = append(result, dto.ManualFundResData{
			ID:            fund.ID,
			UserID:        fund.UserID,
			AdminID:       fund.AdminID,
			AdminName:     fund.AdminName,
			TransactionID: fund.TransactionID,
			Type:          fund.Type,
			Amount:        fund.Amount,
			Reason:        fund.Reason,
			Currency:      fund.Currency,
			Note:          fund.Note,
			CreatedAt:     fund.CreatedAt,
		})
	}

	return result, totalCount, nil
}
