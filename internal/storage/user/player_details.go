package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	// Query balance logs from database
	rows, err := u.db.GetPool().Query(ctx, `
		SELECT 
			id, user_id, component, change_cents, change_units, 
			operational_group_id, operational_type_id, description, 
			timestamp, balance_after_cents, balance_after_units, 
			transaction_id, status, currency_code
		FROM balance_logs 
		WHERE user_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		u.log.Error("Failed to get balance logs", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}
	defer rows.Close()

	var logs []dto.BalanceLog
	for rows.Next() {
		var log dto.BalanceLog
		var changeCents int64
		var changeUnits decimal.Decimal
		var balanceAfterCents int64
		var balanceAfterUnits decimal.Decimal
		var currencyCode string

		err := rows.Scan(
			&log.ID, &log.UserID, &log.Component, &changeCents, &changeUnits,
			&log.OperationalGroupID, &log.OperationalTypeID, &log.Description,
			&log.Timestamp, &balanceAfterCents, &balanceAfterUnits,
			&log.TransactionID, &log.Status, &currencyCode,
		)
		if err != nil {
			u.log.Error("Failed to scan balance log", zap.Error(err))
			continue
		}

		// Convert cents to decimal for display
		log.ChangeAmount = changeUnits
		log.BalanceAfterUpdate = balanceAfterUnits
		log.Currency = currencyCode

		logs = append(logs, log)
	}

	return logs, nil
}

// GetPlayerBettingStats retrieves betting statistics for a player
func (u *user) GetPlayerBettingStats(ctx context.Context, userID uuid.UUID) (dto.PlayerStatistics, error) {
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

	// Query betting statistics from database
	var totalWagered, totalPayout decimal.Decimal
	var totalBets int
	var lastActivity time.Time

	err := u.db.GetPool().QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(amount), 0) as total_wagered,
			COALESCE(SUM(payout), 0) as total_payout,
			COUNT(*) as total_bets,
			COALESCE(MAX(timestamp), '1970-01-01'::timestamp) as last_activity
		FROM bets 
		WHERE user_id = $1
	`, userID).Scan(&totalWagered, &totalPayout, &totalBets, &lastActivity)

	if err != nil {
		u.log.Error("Failed to get betting stats", zap.Error(err), zap.String("user_id", userID.String()))
		return stats, err
	}

	stats.TotalWagered = totalWagered
	stats.NetPL = totalPayout.Sub(totalWagered) // Net P&L = Payout - Wagered
	stats.TotalBets = totalBets
	stats.LastActivity = lastActivity

	// Calculate win rate and average bet size
	if totalBets > 0 {
		stats.AvgBetSize = totalWagered.Div(decimal.NewFromInt(int64(totalBets)))

		// Count wins (bets with payout > amount)
		var wins int
		err = u.db.GetPool().QueryRow(ctx, `
			SELECT COUNT(*) 
			FROM bets 
			WHERE user_id = $1 AND payout > amount
		`, userID).Scan(&wins)
		if err != nil {
			u.log.Error("Failed to count wins", zap.Error(err))
		} else {
			stats.TotalWins = wins
			stats.TotalLosses = totalBets - wins
			if totalBets > 0 {
				stats.WinRate = decimal.NewFromInt(int64(wins)).Div(decimal.NewFromInt(int64(totalBets))).Mul(decimal.NewFromInt(100))
			}
		}
	}

	return stats, nil
}

// GetPlayerSessionCount retrieves the number of sessions for a player
func (u *user) GetPlayerSessionCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var sessionCount int

	err := u.db.GetPool().QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM user_sessions 
		WHERE user_id = $1
	`, userID).Scan(&sessionCount)

	if err != nil {
		u.log.Error("Failed to get session count", zap.Error(err), zap.String("user_id", userID.String()))
		return 0, err
	}

	return sessionCount, nil
}
