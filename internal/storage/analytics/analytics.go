package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform/clickhouse"
	"go.uber.org/zap"
)

type AnalyticsStorage interface {
	// Transaction methods
	InsertTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error
	InsertTransactions(ctx context.Context, transactions []*dto.AnalyticsTransaction) error
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)
	GetGameTransactions(ctx context.Context, gameID string, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)

	// Analytics methods
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, dateRange *dto.DateRange) (*dto.UserAnalytics, error)
	GetGameAnalytics(ctx context.Context, gameID string, dateRange *dto.DateRange) (*dto.GameAnalytics, error)
	GetSessionAnalytics(ctx context.Context, sessionID string) (*dto.SessionAnalytics, error)

	// Reporting methods
	GetDailyReport(ctx context.Context, date time.Time) (*dto.DailyReport, error)
	GetEnhancedDailyReport(ctx context.Context, date time.Time) (*dto.EnhancedDailyReport, error)
	GetMonthlyReport(ctx context.Context, year int, month int) (*dto.MonthlyReport, error)
	GetTopGames(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.GameStats, error)
	GetTopPlayers(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.PlayerStats, error)

	// Real-time methods
	GetRealTimeStats(ctx context.Context) (*dto.RealTimeStats, error)
	GetUserBalanceHistory(ctx context.Context, userID uuid.UUID, hours int) ([]*dto.BalanceSnapshot, error)
	InsertBalanceSnapshot(ctx context.Context, snapshot *dto.BalanceSnapshot) error

	// Summary methods
	GetTransactionSummary(ctx context.Context) (*dto.TransactionSummaryStats, error)
}

type AnalyticsStorageImpl struct {
	clickhouse *clickhouse.ClickHouseClient
	logger     *zap.Logger
}

func NewAnalyticsStorage(clickhouse *clickhouse.ClickHouseClient, logger *zap.Logger) AnalyticsStorage {
	return &AnalyticsStorageImpl{
		clickhouse: clickhouse,
		logger:     logger,
	}
}

func (s *AnalyticsStorageImpl) InsertTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error {
	query := `
		INSERT INTO tucanbit_analytics.transactions (
			id, user_id, transaction_type, amount, currency, status,
			game_id, game_name, provider, session_id, round_id,
			bet_amount, win_amount, net_result, balance_before, balance_after,
			payment_method, external_transaction_id, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	args := []interface{}{
		transaction.ID,
		transaction.UserID.String(),
		transaction.TransactionType,
		transaction.Amount,
		transaction.Currency,
		transaction.Status,
		transaction.GameID,
		transaction.GameName,
		transaction.Provider,
		transaction.SessionID,
		transaction.RoundID,
		safeDecimalPtr(transaction.BetAmount),
		safeDecimalPtr(transaction.WinAmount),
		safeDecimalPtr(transaction.NetResult),
		transaction.BalanceBefore,
		transaction.BalanceAfter,
		transaction.PaymentMethod,
		transaction.ExternalTransactionID,
		transaction.Metadata,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	}

	if err := s.clickhouse.Execute(ctx, query, args...); err != nil {
		s.logger.Error("Failed to insert transaction",
			zap.String("transaction_id", transaction.ID),
			zap.String("user_id", transaction.UserID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	s.logger.Debug("Transaction inserted successfully",
		zap.String("transaction_id", transaction.ID),
		zap.String("user_id", transaction.UserID.String()))

	return nil
}

// safeDecimalPtr safely converts a *decimal.Decimal to decimal.Decimal, returning decimal.Zero if nil
func safeDecimalPtr(ptr *decimal.Decimal) decimal.Decimal {
	if ptr == nil {
		return decimal.Zero
	}
	return *ptr
}

func (s *AnalyticsStorageImpl) InsertTransactions(ctx context.Context, transactions []*dto.AnalyticsTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	columns := []string{
		"id", "user_id", "transaction_type", "amount", "currency", "status",
		"game_id", "game_name", "provider", "session_id", "round_id",
		"bet_amount", "win_amount", "net_result", "balance_before", "balance_after",
		"payment_method", "external_transaction_id", "metadata", "created_at", "updated_at",
	}

	rows := make([][]interface{}, 0, len(transactions))
	for _, transaction := range transactions {
		row := []interface{}{
			transaction.ID,
			transaction.UserID.String(),
			transaction.TransactionType,
			transaction.Amount,
			transaction.Currency,
			transaction.Status,
			transaction.GameID,
			transaction.GameName,
			transaction.Provider,
			transaction.SessionID,
			transaction.RoundID,
			transaction.BetAmount,
			transaction.WinAmount,
			transaction.NetResult,
			transaction.BalanceBefore,
			transaction.BalanceAfter,
			transaction.PaymentMethod,
			transaction.ExternalTransactionID,
			transaction.Metadata,
			transaction.CreatedAt,
			transaction.UpdatedAt,
		}
		rows = append(rows, row)
	}

	if err := s.clickhouse.Insert(ctx, "transactions", columns, rows); err != nil {
		s.logger.Error("Failed to insert transactions batch",
			zap.Int("count", len(transactions)),
			zap.Error(err))
		return fmt.Errorf("failed to insert transactions batch: %w", err)
	}

	s.logger.Info("Transactions batch inserted successfully",
		zap.Int("count", len(transactions)))

	return nil
}

func (s *AnalyticsStorageImpl) GetUserTransactions(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error) {
	query := `
		SELECT 
			id, user_id, transaction_type, amount, currency, status,
			game_id, game_name, provider, session_id, round_id,
			bet_amount, win_amount, net_result, balance_before, balance_after,
			payment_method, external_transaction_id, metadata, created_at, updated_at
		FROM transactions
		WHERE user_id = ? AND amount > 0
	`

	args := []interface{}{userID.String()}

	// Add filters
	if filters != nil {
		if filters.DateFrom != nil {
			query += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
		if filters.TransactionType != nil {
			query += " AND transaction_type = ?"
			args = append(args, *filters.TransactionType)
		}
		if filters.GameID != nil {
			query += " AND game_id = ?"
			args = append(args, *filters.GameID)
		}
		if filters.Status != nil {
			query += " AND status = ?"
			args = append(args, *filters.Status)
		}
	}

	query += " ORDER BY created_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := s.clickhouse.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*dto.AnalyticsTransaction
	for rows.Next() {
		var transaction dto.AnalyticsTransaction
		var userIDStr string
		var gameID, gameName, provider, sessionID, roundID, paymentMethod, externalTransactionID, metadata *string

		err := rows.Scan(
			&transaction.ID,
			&userIDStr,
			&transaction.TransactionType,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Status,
			&gameID,
			&gameName,
			&provider,
			&sessionID,
			&roundID,
			&transaction.BetAmount,
			&transaction.WinAmount,
			&transaction.NetResult,
			&transaction.BalanceBefore,
			&transaction.BalanceAfter,
			&paymentMethod,
			&externalTransactionID,
			&metadata,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		// Handle NULL values by converting empty strings to nil pointers
		if gameID != nil && *gameID == "" {
			transaction.GameID = nil
		} else {
			transaction.GameID = gameID
		}

		if gameName != nil && *gameName == "" {
			transaction.GameName = nil
		} else {
			transaction.GameName = gameName
		}

		if provider != nil && *provider == "" {
			transaction.Provider = nil
		} else {
			transaction.Provider = provider
		}

		if sessionID != nil && *sessionID == "" {
			transaction.SessionID = nil
		} else {
			transaction.SessionID = sessionID
		}

		if roundID != nil && *roundID == "" {
			transaction.RoundID = nil
		} else {
			transaction.RoundID = roundID
		}

		if paymentMethod != nil && *paymentMethod == "" {
			transaction.PaymentMethod = nil
		} else {
			transaction.PaymentMethod = paymentMethod
		}

		if externalTransactionID != nil && *externalTransactionID == "" {
			transaction.ExternalTransactionID = nil
		} else {
			transaction.ExternalTransactionID = externalTransactionID
		}

		if metadata != nil && *metadata == "" {
			transaction.Metadata = nil
		} else {
			transaction.Metadata = metadata
		}

		transaction.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (s *AnalyticsStorageImpl) GetUserAnalytics(ctx context.Context, userID uuid.UUID, dateRange *dto.DateRange) (*dto.UserAnalytics, error) {
	query := `
		SELECT 
			user_id,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type = 'bonus'), 8) as total_bonuses,
			toDecimal64(sumIf(amount, transaction_type = 'cashback'), 8) as total_cashback,
			toUInt64(count()) as transaction_count,
			toUInt64(uniqExact(game_id)) as unique_games_played,
			toUInt64(uniqExact(session_id)) as session_count,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, avgIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as avg_bet_amount,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, maxIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as max_bet_amount,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, minIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as min_bet_amount,
			max(created_at) as last_activity
		FROM tucanbit_analytics.transactions
		WHERE user_id = ? AND amount > 0
		GROUP BY user_id
	`

	args := []interface{}{userID.String()}

	if dateRange != nil {
		if dateRange.From != nil {
			query += " AND created_at >= ?"
			args = append(args, *dateRange.From)
		}
		if dateRange.To != nil {
			query += " AND created_at <= ?"
			args = append(args, *dateRange.To)
		}
	}

	row := s.clickhouse.QueryRow(ctx, query, args...)

	var analytics dto.UserAnalytics
	var userIDStr string

	err := row.Scan(
		&userIDStr,
		&analytics.TotalDeposits,
		&analytics.TotalWithdrawals,
		&analytics.TotalBets,
		&analytics.TotalWins,
		&analytics.TotalBonuses,
		&analytics.TotalCashback,
		&analytics.TransactionCount,
		&analytics.UniqueGamesPlayed,
		&analytics.SessionCount,
		&analytics.AvgBetAmount,
		&analytics.MaxBetAmount,
		&analytics.MinBetAmount,
		&analytics.LastActivity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user analytics: %w", err)
	}

	analytics.UserID, err = uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	// Calculate net loss
	analytics.NetLoss = analytics.TotalBets.Sub(analytics.TotalWins)

	return &analytics, nil
}

func (s *AnalyticsStorageImpl) GetRealTimeStats(ctx context.Context) (*dto.RealTimeStats, error) {
	s.logger.Info("GetRealTimeStats called - checking ClickHouse connection")

	if s.clickhouse == nil {
		s.logger.Error("ClickHouse client is nil")
		return nil, fmt.Errorf("ClickHouse client is not initialized")
	}

	query := `
		SELECT 
			toUInt64(count()) as total_transactions,
			toUInt64(countIf(transaction_type = 'deposit')) as deposits_count,
			toUInt64(countIf(transaction_type = 'withdrawal')) as withdrawals_count,
			toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as bets_count,
			toUInt64(countIf(transaction_type IN ('win', 'groove_win'))) as wins_count,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toUInt64(uniqExact(user_id)) as active_users,
			toUInt64(uniqExact(game_id)) as active_games
		FROM tucanbit_analytics.transactions
	`

	s.logger.Info("Executing ClickHouse query", zap.String("query", query))
	row := s.clickhouse.QueryRow(ctx, query)

	var stats dto.RealTimeStats
	err := row.Scan(
		&stats.TotalTransactions,
		&stats.DepositsCount,
		&stats.WithdrawalsCount,
		&stats.BetsCount,
		&stats.WinsCount,
		&stats.TotalDeposits,
		&stats.TotalWithdrawals,
		&stats.TotalBets,
		&stats.TotalWins,
		&stats.ActiveUsers,
		&stats.ActiveGames,
	)
	if err != nil {
		s.logger.Error("Failed to scan real-time stats", zap.Error(err))
		return nil, fmt.Errorf("failed to scan real-time stats: %w", err)
	}

	s.logger.Info("Real-time stats scanned successfully",
		zap.Uint32("totalTransactions", stats.TotalTransactions),
		zap.Uint32("depositsCount", stats.DepositsCount),
		zap.String("totalDeposits", stats.TotalDeposits.String()),
		zap.Uint32("activeUsers", stats.ActiveUsers))

	stats.Timestamp = time.Now()
	stats.NetRevenue = stats.TotalBets.Sub(stats.TotalWins)

	return &stats, nil
}

func (s *AnalyticsStorageImpl) GetGameTransactions(ctx context.Context, gameID string, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error) {
	query := `
		SELECT 
			id, user_id, transaction_type, amount, currency, status,
			game_id, game_name, provider, session_id, round_id,
			bet_amount, win_amount, net_result, balance_before, balance_after,
			payment_method, external_transaction_id, metadata, created_at, updated_at
		FROM transactions
		WHERE game_id = ? AND amount > 0
	`

	args := []interface{}{gameID}

	// Add filters
	if filters != nil {
		if filters.DateFrom != nil {
			query += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
		if filters.TransactionType != nil {
			query += " AND transaction_type = ?"
			args = append(args, *filters.TransactionType)
		}
		if filters.Status != nil {
			query += " AND status = ?"
			args = append(args, *filters.Status)
		}
	}

	query += " ORDER BY created_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := s.clickhouse.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query game transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*dto.AnalyticsTransaction
	for rows.Next() {
		var transaction dto.AnalyticsTransaction
		var userIDStr string

		err := rows.Scan(
			&transaction.ID,
			&userIDStr,
			&transaction.TransactionType,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.Status,
			&transaction.GameID,
			&transaction.GameName,
			&transaction.Provider,
			&transaction.SessionID,
			&transaction.RoundID,
			&transaction.BetAmount,
			&transaction.WinAmount,
			&transaction.NetResult,
			&transaction.BalanceBefore,
			&transaction.BalanceAfter,
			&transaction.PaymentMethod,
			&transaction.ExternalTransactionID,
			&transaction.Metadata,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transaction.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (s *AnalyticsStorageImpl) GetGameAnalytics(ctx context.Context, gameID string, dateRange *dto.DateRange) (*dto.GameAnalytics, error) {
	query := `
		SELECT 
			game_id,
			game_name,
			provider,
			sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
			sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
			uniqExact(user_id) as total_players,
			uniqExact(session_id) as total_sessions,
			uniqExact(round_id) as total_rounds,
			avgIf(amount, transaction_type IN ('bet', 'groove_bet')) as avg_bet_amount,
			maxIf(amount, transaction_type IN ('bet', 'groove_bet')) as max_bet_amount,
			minIf(amount, transaction_type IN ('bet', 'groove_bet')) as min_bet_amount,
			CASE 
				WHEN sumIf(amount, transaction_type IN ('bet', 'groove_bet')) > 0 
				THEN (sumIf(amount, transaction_type IN ('win', 'groove_win')) / sumIf(amount, transaction_type IN ('bet', 'groove_bet'))) * 100
				ELSE 0 
			END as rtp
		FROM transactions
		WHERE game_id = ?
	`

	args := []interface{}{gameID}

	if dateRange != nil {
		if dateRange.From != nil {
			query += " AND created_at >= ?"
			args = append(args, *dateRange.From)
		}
		if dateRange.To != nil {
			query += " AND created_at <= ?"
			args = append(args, *dateRange.To)
		}
	}

	row := s.clickhouse.QueryRow(ctx, query, args...)

	var analytics dto.GameAnalytics
	var gameName, provider string

	err := row.Scan(
		&analytics.GameID,
		&gameName,
		&provider,
		&analytics.TotalBets,
		&analytics.TotalWins,
		&analytics.TotalPlayers,
		&analytics.TotalSessions,
		&analytics.TotalRounds,
		&analytics.AvgBetAmount,
		&analytics.MaxBetAmount,
		&analytics.MinBetAmount,
		&analytics.RTP,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan game analytics: %w", err)
	}

	analytics.GameName = gameName
	analytics.Provider = provider
	analytics.Volatility = "medium" // Default volatility

	return &analytics, nil
}

func (s *AnalyticsStorageImpl) GetSessionAnalytics(ctx context.Context, sessionID string) (*dto.SessionAnalytics, error) {
	query := `
		SELECT 
			session_id,
			user_id,
			game_id,
			game_name,
			provider,
			min(created_at) as start_time,
			max(created_at) as end_time,
			dateDiff('second', min(created_at), max(created_at)) as duration_seconds,
			sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
			sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
			sumIf(amount, transaction_type IN ('win', 'groove_win')) - sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as net_result,
			countIf(transaction_type IN ('bet', 'groove_bet')) as bet_count,
			countIf(transaction_type IN ('win', 'groove_win')) as win_count,
			max(balance_after) as max_balance,
			min(balance_before) as min_balance
		FROM transactions
		WHERE session_id = ?
		GROUP BY session_id, user_id, game_id, game_name, provider
	`

	row := s.clickhouse.QueryRow(ctx, query, sessionID)

	var analytics dto.SessionAnalytics
	var userIDStr string
	var gameID, gameName, provider *string
	var durationSeconds *uint64

	err := row.Scan(
		&analytics.SessionID,
		&userIDStr,
		&gameID,
		&gameName,
		&provider,
		&analytics.StartTime,
		&analytics.EndTime,
		&durationSeconds,
		&analytics.TotalBets,
		&analytics.TotalWins,
		&analytics.NetResult,
		&analytics.BetCount,
		&analytics.WinCount,
		&analytics.MaxBalance,
		&analytics.MinBalance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan session analytics: %w", err)
	}

	analytics.UserID, err = uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	analytics.GameID = gameID
	analytics.GameName = gameName
	analytics.Provider = provider
	analytics.DurationSeconds = durationSeconds
	analytics.SessionType = "regular"

	return &analytics, nil
}

func (s *AnalyticsStorageImpl) GetDailyReport(ctx context.Context, date time.Time) (*dto.DailyReport, error) {
	query := `
		SELECT 
			toUInt64(count()) as total_transactions,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as net_revenue,
			toUInt64(uniqExact(user_id)) as active_users,
			toUInt64(uniqExact(game_id)) as active_games,
			toUInt64(0) as new_users, -- TODO: Get from users table
			toUInt64(uniqExactIf(user_id, transaction_type = 'deposit')) as unique_depositors,
			toUInt64(uniqExactIf(user_id, transaction_type = 'withdrawal')) as unique_withdrawers,
			toUInt64(countIf(transaction_type = 'deposit')) as deposit_count,
			toUInt64(countIf(transaction_type = 'withdrawal')) as withdrawal_count,
			toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as bet_count,
			toUInt64(countIf(transaction_type IN ('win', 'groove_win'))) as win_count,
			toDecimal64(sumIf(amount, transaction_type = 'cashback'), 8) as cashback_earned,
			toDecimal64(sumIf(amount, transaction_type = 'cashback'), 8) as cashback_claimed,
			toDecimal64(0, 8) as admin_corrections -- TODO: Get from balance_logs or admin_fund_movements
		FROM tucanbit_analytics.transactions
		WHERE toDate(created_at + INTERVAL 3 HOUR) = ?
	`

	dateStr := date.Format("2006-01-02")
	row := s.clickhouse.QueryRow(ctx, query, dateStr)

	var report dto.DailyReport
	err := row.Scan(
		&report.TotalTransactions,
		&report.TotalDeposits,
		&report.TotalWithdrawals,
		&report.TotalBets,
		&report.TotalWins,
		&report.NetRevenue,
		&report.ActiveUsers,
		&report.ActiveGames,
		&report.NewUsers,
		&report.UniqueDepositors,
		&report.UniqueWithdrawers,
		&report.DepositCount,
		&report.WithdrawalCount,
		&report.BetCount,
		&report.WinCount,
		&report.CashbackEarned,
		&report.CashbackClaimed,
		&report.AdminCorrections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan daily report: %w", err)
	}

	report.Date = date

	// Create date range for the full day
	startOfDay := date
	endOfDay := date.Add(24 * time.Hour).Add(-time.Nanosecond) // End of day

	// Get top games for the day
	topGames, err := s.GetTopGames(ctx, 5, &dto.DateRange{
		From: &startOfDay,
		To:   &endOfDay,
	})
	if err != nil {
		s.logger.Warn("Failed to get top games for daily report", zap.Error(err))
	} else {
		// Convert []*dto.GameStats to []dto.GameStats
		report.TopGames = make([]dto.GameStats, len(topGames))
		for i, game := range topGames {
			report.TopGames[i] = *game
		}
	}

	// Get top players for the day
	topPlayers, err := s.GetTopPlayers(ctx, 5, &dto.DateRange{
		From: &startOfDay,
		To:   &endOfDay,
	})
	if err != nil {
		s.logger.Warn("Failed to get top players for daily report", zap.Error(err))
	} else {
		// Convert []*dto.PlayerStats to []dto.PlayerStats
		report.TopPlayers = make([]dto.PlayerStats, len(topPlayers))
		for i, player := range topPlayers {
			report.TopPlayers[i] = *player
		}
	}

	return &report, nil
}

// GetEnhancedDailyReport returns an enhanced daily report with comparison metrics
func (s *AnalyticsStorageImpl) GetEnhancedDailyReport(ctx context.Context, date time.Time) (*dto.EnhancedDailyReport, error) {
	// Get the base daily report
	baseReport, err := s.GetDailyReport(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get base daily report: %w", err)
	}

	// Convert base report to enhanced report
	enhancedReport := &dto.EnhancedDailyReport{
		Date:              baseReport.Date,
		TotalTransactions: baseReport.TotalTransactions,
		TotalDeposits:     baseReport.TotalDeposits,
		TotalWithdrawals:  baseReport.TotalWithdrawals,
		TotalBets:         baseReport.TotalBets,
		TotalWins:         baseReport.TotalWins,
		NetRevenue:        baseReport.NetRevenue,
		ActiveUsers:       baseReport.ActiveUsers,
		ActiveGames:       baseReport.ActiveGames,
		NewUsers:          baseReport.NewUsers,
		UniqueDepositors:  baseReport.UniqueDepositors,
		UniqueWithdrawers: baseReport.UniqueWithdrawers,
		DepositCount:      baseReport.DepositCount,
		WithdrawalCount:   baseReport.WithdrawalCount,
		BetCount:          baseReport.BetCount,
		WinCount:          baseReport.WinCount,
		CashbackEarned:    baseReport.CashbackEarned,
		CashbackClaimed:   baseReport.CashbackClaimed,
		AdminCorrections:  baseReport.AdminCorrections,
		TopGames:          baseReport.TopGames,
		TopPlayers:        baseReport.TopPlayers,
	}

	// Get previous day's report for comparison
	previousDay := date.AddDate(0, 0, -1)
	previousReport, err := s.GetDailyReport(ctx, previousDay)
	if err != nil {
		s.logger.Warn("Failed to get previous day report for comparison", zap.Error(err))
		// Set all changes to 0 if previous day data is not available
		enhancedReport.PreviousDayChange = dto.DailyReportComparison{}
	} else {
		// Calculate percentage changes
		enhancedReport.PreviousDayChange = s.calculatePercentageChange(baseReport, previousReport)
	}

	// Get MTD (Month To Date) data
	mtdData, err := s.getMTDData(ctx, date)
	if err != nil {
		s.logger.Warn("Failed to get MTD data", zap.Error(err))
		enhancedReport.MTD = dto.DailyReportMTD{}
	} else {
		enhancedReport.MTD = *mtdData
	}

	// Get SPLM (Same Period Last Month) data
	splmData, err := s.getSPLMData(ctx, date)
	if err != nil {
		s.logger.Warn("Failed to get SPLM data", zap.Error(err))
		enhancedReport.SPLM = dto.DailyReportSPLM{}
	} else {
		enhancedReport.SPLM = *splmData
	}

	// Calculate MTD vs SPLM percentage changes
	if mtdData != nil && splmData != nil {
		enhancedReport.MTDvsSPLMChange = s.calculateMTDvsSPLMChange(mtdData, splmData)
	}

	return enhancedReport, nil
}

// calculatePercentageChange calculates percentage change between two daily reports
func (s *AnalyticsStorageImpl) calculatePercentageChange(current, previous *dto.DailyReport) dto.DailyReportComparison {
	return dto.DailyReportComparison{
		TotalTransactionsChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.TotalTransactions)),
			decimal.NewFromInt(int64(previous.TotalTransactions)),
		),
		TotalDepositsChange:    s.calculatePercentageChangeDecimal(current.TotalDeposits, previous.TotalDeposits),
		TotalWithdrawalsChange: s.calculatePercentageChangeDecimal(current.TotalWithdrawals, previous.TotalWithdrawals),
		TotalBetsChange:        s.calculatePercentageChangeDecimal(current.TotalBets, previous.TotalBets),
		TotalWinsChange:        s.calculatePercentageChangeDecimal(current.TotalWins, previous.TotalWins),
		NetRevenueChange:       s.calculatePercentageChangeDecimal(current.NetRevenue, previous.NetRevenue),
		ActiveUsersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.ActiveUsers)),
			decimal.NewFromInt(int64(previous.ActiveUsers)),
		),
		ActiveGamesChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.ActiveGames)),
			decimal.NewFromInt(int64(previous.ActiveGames)),
		),
		NewUsersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.NewUsers)),
			decimal.NewFromInt(int64(previous.NewUsers)),
		),
		UniqueDepositorsChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.UniqueDepositors)),
			decimal.NewFromInt(int64(previous.UniqueDepositors)),
		),
		UniqueWithdrawersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.UniqueWithdrawers)),
			decimal.NewFromInt(int64(previous.UniqueWithdrawers)),
		),
		DepositCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.DepositCount)),
			decimal.NewFromInt(int64(previous.DepositCount)),
		),
		WithdrawalCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.WithdrawalCount)),
			decimal.NewFromInt(int64(previous.WithdrawalCount)),
		),
		BetCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.BetCount)),
			decimal.NewFromInt(int64(previous.BetCount)),
		),
		WinCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.WinCount)),
			decimal.NewFromInt(int64(previous.WinCount)),
		),
		CashbackEarnedChange:   s.calculatePercentageChangeDecimal(current.CashbackEarned, previous.CashbackEarned),
		CashbackClaimedChange:  s.calculatePercentageChangeDecimal(current.CashbackClaimed, previous.CashbackClaimed),
		AdminCorrectionsChange: s.calculatePercentageChangeDecimal(current.AdminCorrections, previous.AdminCorrections),
	}
}

// calculatePercentageChangeDecimal calculates percentage change between two decimal values and returns as string with % sign
func (s *AnalyticsStorageImpl) calculatePercentageChangeDecimal(current, previous decimal.Decimal) string {
	if previous.IsZero() {
		if current.IsZero() {
			return "0%"
		}
		// 100% increase from 0
		return "100%"
	}

	change := current.Sub(previous).Div(previous).Mul(decimal.NewFromInt(100))
	return change.StringFixed(2) + "%"
}

// getMTDData gets Month To Date data for the given date
func (s *AnalyticsStorageImpl) getMTDData(ctx context.Context, date time.Time) (*dto.DailyReportMTD, error) {
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

	query := `
		WITH mtd_data AS (
			-- Deposits from deposits table
			SELECT 
				user_id,
				amount,
				'deposit' as transaction_type,
				created_at,
				'deposits' as source
			FROM tucanbit_analytics.deposits
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Withdrawals from withdrawals table
			SELECT 
				user_id,
				amount,
				'withdrawal' as transaction_type,
				created_at,
				'withdrawals' as source
			FROM tucanbit_analytics.withdrawals
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- GrooveTech wagers and results
			SELECT 
				account_id as user_id,
				amount,
				CASE 
					WHEN type = 'wager' THEN 'groove_bet'
					WHEN type = 'result' THEN 'groove_win'
					ELSE type
				END as transaction_type,
				created_at,
				'groove' as source
			FROM tucanbit_analytics.groove_transactions
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Regular bets
			SELECT 
				user_id,
				amount,
				'bet' as transaction_type,
				timestamp as created_at,
				'bets' as source
			FROM tucanbit_analytics.bets
			WHERE toDate(timestamp) >= ? AND toDate(timestamp) <= ?
			
			UNION ALL
			
			-- Cashback earnings
			SELECT 
				user_id,
				earned_amount as amount,
				'cashback_earned' as transaction_type,
				created_at,
				'cashback_earnings' as source
			FROM tucanbit_analytics.cashback_earnings
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Cashback claims
			SELECT 
				user_id,
				claim_amount as amount,
				'cashback_claimed' as transaction_type,
				created_at,
				'cashback_claims' as source
			FROM tucanbit_analytics.cashback_claims
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
		)
		SELECT 
			toUInt64(count()) as total_transactions,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as net_revenue,
			toUInt64(uniqExact(user_id)) as active_users,
			toUInt64(0) as active_games, -- TODO: Get from games table
			toUInt64(0) as new_users, -- TODO: Get from users table
			toUInt64(uniqExactIf(user_id, transaction_type = 'deposit')) as unique_depositors,
			toUInt64(uniqExactIf(user_id, transaction_type = 'withdrawal')) as unique_withdrawers,
			toUInt64(countIf(transaction_type = 'deposit')) as deposit_count,
			toUInt64(countIf(transaction_type = 'withdrawal')) as withdrawal_count,
			toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as bet_count,
			toUInt64(countIf(transaction_type IN ('win', 'groove_win'))) as win_count,
			toDecimal64(sumIf(amount, transaction_type = 'cashback_earned'), 8) as cashback_earned,
			toDecimal64(sumIf(amount, transaction_type = 'cashback_claimed'), 8) as cashback_claimed,
			toDecimal64(0, 8) as admin_corrections -- TODO: Get from balance_logs or admin_fund_movements
		FROM mtd_data
	`

	startOfMonthStr := startOfMonth.Format("2006-01-02")
	dateStr := date.Format("2006-01-02")
	row := s.clickhouse.QueryRow(ctx, query, startOfMonthStr, dateStr, startOfMonthStr, dateStr, startOfMonthStr, dateStr, startOfMonthStr, dateStr, startOfMonthStr, dateStr, startOfMonthStr, dateStr)

	var mtd dto.DailyReportMTD
	err := row.Scan(
		&mtd.TotalTransactions,
		&mtd.TotalDeposits,
		&mtd.TotalWithdrawals,
		&mtd.TotalBets,
		&mtd.TotalWins,
		&mtd.NetRevenue,
		&mtd.ActiveUsers,
		&mtd.ActiveGames,
		&mtd.NewUsers,
		&mtd.UniqueDepositors,
		&mtd.UniqueWithdrawers,
		&mtd.DepositCount,
		&mtd.WithdrawalCount,
		&mtd.BetCount,
		&mtd.WinCount,
		&mtd.CashbackEarned,
		&mtd.CashbackClaimed,
		&mtd.AdminCorrections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan MTD data: %w", err)
	}

	return &mtd, nil
}

// getSPLMData gets Same Period Last Month data for the given date
func (s *AnalyticsStorageImpl) getSPLMData(ctx context.Context, date time.Time) (*dto.DailyReportSPLM, error) {
	// Calculate same period last month
	lastMonth := date.AddDate(0, -1, 0)
	startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
	endOfLastMonth := startOfLastMonth.AddDate(0, 1, -1)

	// Use the same day of the month as the current date, but cap at the end of last month
	splmEndDate := time.Date(lastMonth.Year(), lastMonth.Month(), date.Day(), 0, 0, 0, 0, lastMonth.Location())
	if splmEndDate.After(endOfLastMonth) {
		splmEndDate = endOfLastMonth
	}

	query := `
		WITH splm_data AS (
			-- Deposits from deposits table
			SELECT 
				user_id,
				amount,
				'deposit' as transaction_type,
				created_at,
				'deposits' as source
			FROM tucanbit_analytics.deposits
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Withdrawals from withdrawals table
			SELECT 
				user_id,
				amount,
				'withdrawal' as transaction_type,
				created_at,
				'withdrawals' as source
			FROM tucanbit_analytics.withdrawals
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- GrooveTech wagers and results
			SELECT 
				account_id as user_id,
				amount,
				CASE 
					WHEN type = 'wager' THEN 'groove_bet'
					WHEN type = 'result' THEN 'groove_win'
					ELSE type
				END as transaction_type,
				created_at,
				'groove' as source
			FROM tucanbit_analytics.groove_transactions
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Regular bets
			SELECT 
				user_id,
				amount,
				'bet' as transaction_type,
				timestamp as created_at,
				'bets' as source
			FROM tucanbit_analytics.bets
			WHERE toDate(timestamp) >= ? AND toDate(timestamp) <= ?
			
			UNION ALL
			
			-- Cashback earnings
			SELECT 
				user_id,
				earned_amount as amount,
				'cashback_earned' as transaction_type,
				created_at,
				'cashback_earnings' as source
			FROM tucanbit_analytics.cashback_earnings
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
			
			UNION ALL
			
			-- Cashback claims
			SELECT 
				user_id,
				claim_amount as amount,
				'cashback_claimed' as transaction_type,
				created_at,
				'cashback_claims' as source
			FROM tucanbit_analytics.cashback_claims
			WHERE toDate(created_at) >= ? AND toDate(created_at) <= ?
		)
		SELECT 
			toUInt64(count()) as total_transactions,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as net_revenue,
			toUInt64(uniqExact(user_id)) as active_users,
			toUInt64(0) as active_games, -- TODO: Get from games table
			toUInt64(0) as new_users, -- TODO: Get from users table
			toUInt64(uniqExactIf(user_id, transaction_type = 'deposit')) as unique_depositors,
			toUInt64(uniqExactIf(user_id, transaction_type = 'withdrawal')) as unique_withdrawers,
			toUInt64(countIf(transaction_type = 'deposit')) as deposit_count,
			toUInt64(countIf(transaction_type = 'withdrawal')) as withdrawal_count,
			toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as bet_count,
			toUInt64(countIf(transaction_type IN ('win', 'groove_win'))) as win_count,
			toDecimal64(sumIf(amount, transaction_type = 'cashback_earned'), 8) as cashback_earned,
			toDecimal64(sumIf(amount, transaction_type = 'cashback_claimed'), 8) as cashback_claimed,
			toDecimal64(0, 8) as admin_corrections -- TODO: Get from balance_logs or admin_fund_movements
		FROM splm_data
	`

	startOfLastMonthStr := startOfLastMonth.Format("2006-01-02")
	splmEndDateStr := splmEndDate.Format("2006-01-02")
	row := s.clickhouse.QueryRow(ctx, query, startOfLastMonthStr, splmEndDateStr, startOfLastMonthStr, splmEndDateStr, startOfLastMonthStr, splmEndDateStr, startOfLastMonthStr, splmEndDateStr, startOfLastMonthStr, splmEndDateStr, startOfLastMonthStr, splmEndDateStr)

	var splm dto.DailyReportSPLM
	err := row.Scan(
		&splm.TotalTransactions,
		&splm.TotalDeposits,
		&splm.TotalWithdrawals,
		&splm.TotalBets,
		&splm.TotalWins,
		&splm.NetRevenue,
		&splm.ActiveUsers,
		&splm.ActiveGames,
		&splm.NewUsers,
		&splm.UniqueDepositors,
		&splm.UniqueWithdrawers,
		&splm.DepositCount,
		&splm.WithdrawalCount,
		&splm.BetCount,
		&splm.WinCount,
		&splm.CashbackEarned,
		&splm.CashbackClaimed,
		&splm.AdminCorrections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan SPLM data: %w", err)
	}

	return &splm, nil
}

// calculateMTDvsSPLMChange calculates percentage change between MTD and SPLM
func (s *AnalyticsStorageImpl) calculateMTDvsSPLMChange(mtd *dto.DailyReportMTD, splm *dto.DailyReportSPLM) dto.DailyReportComparison {
	return dto.DailyReportComparison{
		TotalTransactionsChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.TotalTransactions)),
			decimal.NewFromInt(int64(splm.TotalTransactions)),
		),
		TotalDepositsChange:    s.calculatePercentageChangeDecimal(mtd.TotalDeposits, splm.TotalDeposits),
		TotalWithdrawalsChange: s.calculatePercentageChangeDecimal(mtd.TotalWithdrawals, splm.TotalWithdrawals),
		TotalBetsChange:        s.calculatePercentageChangeDecimal(mtd.TotalBets, splm.TotalBets),
		TotalWinsChange:        s.calculatePercentageChangeDecimal(mtd.TotalWins, splm.TotalWins),
		NetRevenueChange:       s.calculatePercentageChangeDecimal(mtd.NetRevenue, splm.NetRevenue),
		ActiveUsersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.ActiveUsers)),
			decimal.NewFromInt(int64(splm.ActiveUsers)),
		),
		ActiveGamesChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.ActiveGames)),
			decimal.NewFromInt(int64(splm.ActiveGames)),
		),
		NewUsersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.NewUsers)),
			decimal.NewFromInt(int64(splm.NewUsers)),
		),
		UniqueDepositorsChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.UniqueDepositors)),
			decimal.NewFromInt(int64(splm.UniqueDepositors)),
		),
		UniqueWithdrawersChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.UniqueWithdrawers)),
			decimal.NewFromInt(int64(splm.UniqueWithdrawers)),
		),
		DepositCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.DepositCount)),
			decimal.NewFromInt(int64(splm.DepositCount)),
		),
		WithdrawalCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.WithdrawalCount)),
			decimal.NewFromInt(int64(splm.WithdrawalCount)),
		),
		BetCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.BetCount)),
			decimal.NewFromInt(int64(splm.BetCount)),
		),
		WinCountChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(mtd.WinCount)),
			decimal.NewFromInt(int64(splm.WinCount)),
		),
		CashbackEarnedChange:   s.calculatePercentageChangeDecimal(mtd.CashbackEarned, splm.CashbackEarned),
		CashbackClaimedChange:  s.calculatePercentageChangeDecimal(mtd.CashbackClaimed, splm.CashbackClaimed),
		AdminCorrectionsChange: s.calculatePercentageChangeDecimal(mtd.AdminCorrections, splm.AdminCorrections),
	}
}

func (s *AnalyticsStorageImpl) GetMonthlyReport(ctx context.Context, year int, month int) (*dto.MonthlyReport, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	query := `
		SELECT 
			count() as total_transactions,
			sumIf(amount, transaction_type = 'deposit') as total_deposits,
			sumIf(amount, transaction_type = 'withdrawal') as total_withdrawals,
			sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
			sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
			sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')) as net_revenue,
			uniqExact(user_id) as active_users,
			uniqExact(game_id) as active_games,
			countIf(transaction_type = 'registration') as new_users
		FROM transactions
		WHERE toYear(created_at) = ? AND toMonth(created_at) = ?
	`

	row := s.clickhouse.QueryRow(ctx, query, year, month)

	var report dto.MonthlyReport
	err := row.Scan(
		&report.TotalTransactions,
		&report.TotalDeposits,
		&report.TotalWithdrawals,
		&report.TotalBets,
		&report.TotalWins,
		&report.NetRevenue,
		&report.ActiveUsers,
		&report.ActiveGames,
		&report.NewUsers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan monthly report: %w", err)
	}

	report.Year = year
	report.Month = month

	// Calculate average daily revenue
	daysInMonth := endDate.Day()
	if daysInMonth > 0 {
		report.AvgDailyRevenue = report.NetRevenue.Div(decimal.NewFromInt(int64(daysInMonth)))
	}

	// Get top games for the month
	topGames, err := s.GetTopGames(ctx, 10, &dto.DateRange{
		From: &startDate,
		To:   &endDate,
	})
	if err != nil {
		s.logger.Warn("Failed to get top games for monthly report", zap.Error(err))
	} else {
		// Convert []*dto.GameStats to []dto.GameStats
		report.TopGames = make([]dto.GameStats, len(topGames))
		for i, game := range topGames {
			report.TopGames[i] = *game
		}
	}

	// Get top players for the month
	topPlayers, err := s.GetTopPlayers(ctx, 10, &dto.DateRange{
		From: &startDate,
		To:   &endDate,
	})
	if err != nil {
		s.logger.Warn("Failed to get top players for monthly report", zap.Error(err))
	} else {
		// Convert []*dto.PlayerStats to []dto.PlayerStats
		report.TopPlayers = make([]dto.PlayerStats, len(topPlayers))
		for i, player := range topPlayers {
			report.TopPlayers[i] = *player
		}
	}

	return &report, nil
}

func (s *AnalyticsStorageImpl) GetTopGames(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.GameStats, error) {
	query := `
		SELECT 
			game_id,
			game_name,
			provider,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as net_revenue,
			toUInt64(uniqExact(user_id)) as player_count,
			toUInt64(uniqExact(session_id)) as session_count,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, avgIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as avg_bet_amount,
			CASE 
				WHEN sumIf(amount, transaction_type IN ('bet', 'groove_bet')) > 0 
				THEN (sumIf(amount, transaction_type IN ('win', 'groove_win')) / sumIf(amount, transaction_type IN ('bet', 'groove_bet'))) * 100
				ELSE 0 
			END as rtp
		FROM tucanbit_analytics.transactions
		WHERE game_id IS NOT NULL AND game_id != '' AND amount > 0
	`

	args := []interface{}{}

	if dateRange != nil {
		if dateRange.From != nil {
			query += " AND created_at >= ?"
			args = append(args, *dateRange.From)
		}
		if dateRange.To != nil {
			query += " AND created_at <= ?"
			args = append(args, *dateRange.To)
		}
	}

	query += `
		GROUP BY game_id, game_name, provider
		ORDER BY net_revenue DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := s.clickhouse.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top games: %w", err)
	}
	defer rows.Close()

	var games []*dto.GameStats
	rank := 1
	for rows.Next() {
		var game dto.GameStats
		var gameName, provider string

		err := rows.Scan(
			&game.GameID,
			&gameName,
			&provider,
			&game.TotalBets,
			&game.TotalWins,
			&game.NetRevenue,
			&game.PlayerCount,
			&game.SessionCount,
			&game.AvgBetAmount,
			&game.RTP,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game stats: %w", err)
		}

		game.GameName = gameName
		game.Provider = provider
		game.Rank = rank
		rank++

		games = append(games, &game)
	}

	return games, nil
}

func (s *AnalyticsStorageImpl) GetTopPlayers(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.PlayerStats, error) {
	query := `
		SELECT 
			user_id,
			toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as total_deposits,
			toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as total_withdrawals,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_bets,
			toDecimal64(sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(amount, transaction_type IN ('win', 'groove_win')), 8) as net_loss,
			toUInt64(count()) as transaction_count,
			toUInt64(uniqExact(game_id)) as unique_games_played,
			toUInt64(uniqExact(session_id)) as session_count,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, avgIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as avg_bet_amount,
			max(created_at) as last_activity
		FROM tucanbit_analytics.transactions
		WHERE user_id IS NOT NULL AND user_id != '' AND amount > 0
	`

	args := []interface{}{}

	if dateRange != nil {
		if dateRange.From != nil {
			query += " AND created_at >= ?"
			args = append(args, *dateRange.From)
		}
		if dateRange.To != nil {
			query += " AND created_at <= ?"
			args = append(args, *dateRange.To)
		}
	}

	query += `
		GROUP BY user_id
		ORDER BY total_bets DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := s.clickhouse.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top players: %w", err)
	}
	defer rows.Close()

	var players []*dto.PlayerStats
	rank := 1
	for rows.Next() {
		var player dto.PlayerStats
		var userIDStr string

		err := rows.Scan(
			&userIDStr,
			&player.TotalDeposits,
			&player.TotalWithdrawals,
			&player.TotalBets,
			&player.TotalWins,
			&player.NetLoss,
			&player.TransactionCount,
			&player.UniqueGamesPlayed,
			&player.SessionCount,
			&player.AvgBetAmount,
			&player.LastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player stats: %w", err)
		}

		// Handle both UUID and non-UUID user IDs
		if len(userIDStr) == 36 { // UUID format
			player.UserID, err = uuid.Parse(userIDStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse user ID: %w", err)
			}
		} else {
			// For non-UUID IDs (like GrooveTech account_id), create a placeholder UUID
			player.UserID = uuid.New()
		}

		player.Username = "Player_" + userIDStr[:8] // Placeholder username
		player.Rank = rank
		rank++

		players = append(players, &player)
	}

	return players, nil
}

func (s *AnalyticsStorageImpl) GetUserBalanceHistory(ctx context.Context, userID uuid.UUID, hours int) ([]*dto.BalanceSnapshot, error) {
	query := `
		SELECT 
			user_id,
			balance,
			currency,
			snapshot_time,
			transaction_id,
			transaction_type
		FROM balance_snapshots
		WHERE user_id = ? AND snapshot_time >= now() - INTERVAL ? HOUR
		ORDER BY snapshot_time DESC
	`

	rows, err := s.clickhouse.Query(ctx, query, userID.String(), hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance history: %w", err)
	}
	defer rows.Close()

	var snapshots []*dto.BalanceSnapshot
	for rows.Next() {
		var snapshot dto.BalanceSnapshot
		var userIDStr string
		var transactionID, transactionType *string

		err := rows.Scan(
			&userIDStr,
			&snapshot.Balance,
			&snapshot.Currency,
			&snapshot.SnapshotTime,
			&transactionID,
			&transactionType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance snapshot: %w", err)
		}

		snapshot.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID: %w", err)
		}

		snapshot.TransactionID = transactionID
		snapshot.TransactionType = transactionType

		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, nil
}

func (s *AnalyticsStorageImpl) InsertBalanceSnapshot(ctx context.Context, snapshot *dto.BalanceSnapshot) error {
	query := `
		INSERT INTO balance_snapshots (
			user_id, balance, currency, snapshot_time, 
			transaction_id, transaction_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	args := []interface{}{
		snapshot.UserID.String(),
		snapshot.Balance,
		snapshot.Currency,
		snapshot.SnapshotTime,
		snapshot.TransactionID,
		snapshot.TransactionType,
		time.Now(),
	}

	if err := s.clickhouse.Execute(ctx, query, args...); err != nil {
		s.logger.Error("Failed to insert balance snapshot",
			zap.String("user_id", snapshot.UserID.String()),
			zap.String("balance", snapshot.Balance.String()),
			zap.Error(err))
		return fmt.Errorf("failed to insert balance snapshot: %w", err)
	}

	s.logger.Debug("Balance snapshot inserted successfully",
		zap.String("user_id", snapshot.UserID.String()),
		zap.String("balance", snapshot.Balance.String()))

	return nil
}

func (s *AnalyticsStorageImpl) GetTransactionSummary(ctx context.Context) (*dto.TransactionSummaryStats, error) {
	s.logger.Info("GetTransactionSummary called - checking ClickHouse connection")

	if s.clickhouse == nil {
		s.logger.Error("ClickHouse client is nil in GetTransactionSummary")
		return nil, fmt.Errorf("ClickHouse client is not initialized")
	}

	query := `
		SELECT 
			count() as total_transactions,
			sum(amount) as total_volume,
			countIf(status = 'completed') as successful_transactions,
			countIf(status = 'failed') as failed_transactions,
			countIf(transaction_type = 'deposit') as deposit_count,
			countIf(transaction_type = 'withdrawal') as withdrawal_count,
			countIf(transaction_type IN ('bet', 'groove_bet')) as bet_count,
			countIf(transaction_type IN ('win', 'groove_win')) as win_count
		FROM tucanbit_analytics.transactions
	`

	s.logger.Info("Executing ClickHouse query for transaction summary", zap.String("query", query))
	row := s.clickhouse.QueryRow(ctx, query)

	var stats dto.TransactionSummaryStats
	err := row.Scan(
		&stats.TotalTransactions,
		&stats.TotalVolume,
		&stats.SuccessfulTransactions,
		&stats.FailedTransactions,
		&stats.DepositCount,
		&stats.WithdrawalCount,
		&stats.BetCount,
		&stats.WinCount,
	)
	if err != nil {
		s.logger.Error("Failed to scan transaction summary stats", zap.Error(err))
		return nil, fmt.Errorf("failed to scan transaction summary stats: %w", err)
	}

	s.logger.Info("Transaction summary stats retrieved successfully",
		zap.Int("totalTransactions", stats.TotalTransactions),
		zap.String("totalVolume", stats.TotalVolume.String()),
		zap.Int("successfulTransactions", stats.SuccessfulTransactions),
		zap.Int("failedTransactions", stats.FailedTransactions),
		zap.Int("depositCount", stats.DepositCount),
		zap.Int("withdrawalCount", stats.WithdrawalCount))

	return &stats, nil
}
