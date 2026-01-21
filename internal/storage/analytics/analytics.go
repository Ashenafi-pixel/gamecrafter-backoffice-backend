package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform/clickhouse"
	"go.uber.org/zap"
)

type AnalyticsStorage interface {
	// Transaction methods
	InsertTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error
	InsertTransactions(ctx context.Context, transactions []*dto.AnalyticsTransaction) error
	// GetUserTransactions returns paginated transactions and the total count matching the filters
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, int, error)
	GetGameTransactions(ctx context.Context, gameID string, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)
	GetTransactionReport(ctx context.Context, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)

	// Player analytics extensions
	GetUserRakebackTransactions(ctx context.Context, userID uuid.UUID, filters *dto.RakebackFilters) ([]*dto.RakebackTransaction, int, dto.UserRakebackTotals, error)
	GetUserTips(ctx context.Context, userID uuid.UUID, filters *dto.TipFilters) ([]*dto.TipTransaction, int, error)
	GetUserWelcomeBonus(ctx context.Context, userID uuid.UUID, filters *dto.WelcomeBonusFilters) ([]*dto.WelcomeBonusTransaction, int, error)
	GetWelcomeBonusTransactions(ctx context.Context, filters *dto.WelcomeBonusFilters) ([]*dto.WelcomeBonusTransaction, int, error)

	// Totals endpoints
	GetUserTransactionsTotals(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) (*dto.UserTransactionsTotals, error)
	GetUserRakebackTotals(ctx context.Context, userID uuid.UUID, filters *dto.RakebackFilters) (*dto.UserRakebackTotals, error)
	GetUserTipsTotals(ctx context.Context, userID uuid.UUID, filters *dto.TipFilters) (*dto.UserTipsTotals, error)
	GetUserWelcomeBonusTotals(ctx context.Context, userID uuid.UUID, filters *dto.WelcomeBonusFilters) (*dto.UserWelcomeBonusTotals, error)

	// Analytics methods
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, dateRange *dto.DateRange) (*dto.UserAnalytics, error)
	GetGameAnalytics(ctx context.Context, gameID string, dateRange *dto.DateRange) (*dto.GameAnalytics, error)
	GetSessionAnalytics(ctx context.Context, sessionID string) (*dto.SessionAnalytics, error)

	// Reporting methods
	GetDailyReport(ctx context.Context, date time.Time) (*dto.DailyReport, error)
	GetEnhancedDailyReport(ctx context.Context, date time.Time) (*dto.EnhancedDailyReport, error)
	GetDailyReportDataTable(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID) (*dto.DailyReportDataTableResponse, error)
	GetWeeklyReport(ctx context.Context, weekStart time.Time, userIDs []uuid.UUID) (*dto.WeeklyReport, error)
	GetMonthlyReport(ctx context.Context, year int, month int) (*dto.MonthlyReport, error)
	GetTopGames(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.GameStats, error)
	GetTopPlayers(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.PlayerStats, error)

	// Real-time methods
	GetRealTimeStats(ctx context.Context) (*dto.RealTimeStats, error)
	GetUserBalanceHistory(ctx context.Context, userID uuid.UUID, hours int) ([]*dto.BalanceSnapshot, error)
	InsertBalanceSnapshot(ctx context.Context, snapshot *dto.BalanceSnapshot) error

	// Summary methods
	GetTransactionSummary(ctx context.Context) (*dto.TransactionSummaryStats, error)

	// Dashboard methods
	GetDashboardOverview(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID, includeDailyBreakdown bool) (*dto.DashboardOverviewResponse, error)
	GetPerformanceSummary(ctx context.Context, rangeType string, dateFrom, dateTo *time.Time, userIDs []uuid.UUID) (*dto.PerformanceSummaryResponse, error)
	GetTimeSeriesAnalytics(ctx context.Context, dateFrom, dateTo time.Time, granularity string, userIDs []uuid.UUID, metrics []string) (*dto.TimeSeriesResponse, error)
}

type AnalyticsStorageImpl struct {
	clickhouse *clickhouse.ClickHouseClient
	pgPool     *pgxpool.Pool
	logger     *zap.Logger
}

func NewAnalyticsStorage(clickhouse *clickhouse.ClickHouseClient, logger *zap.Logger) AnalyticsStorage {
	return &AnalyticsStorageImpl{
		clickhouse: clickhouse,
		logger:     logger,
	}
}

// SetPostgresPool sets the PostgreSQL pool for welcome bonus queries
func (s *AnalyticsStorageImpl) SetPostgresPool(pgPool *pgxpool.Pool) {
	s.pgPool = pgPool
}

func (s *AnalyticsStorageImpl) GetClickHouseClient() *clickhouse.ClickHouseClient {
	return s.clickhouse
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

func (s *AnalyticsStorageImpl) GetUserTransactions(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, int, error) {
	userIDStr := userID.String()

	isGroove := false
	if filters != nil && filters.TransactionType != nil {
		if *filters.TransactionType == "groove_bet" || *filters.TransactionType == "groove_win" {
			isGroove = true
		}
	}

	var where string
	var args []interface{}

	if isGroove {
		where = "WHERE user_id = ? AND transaction_type = ? AND bet_amount IS NOT NULL AND (status = 'completed' OR status = 'pending')"
		args = []interface{}{userIDStr, *filters.TransactionType}

		if filters.DateFrom != nil {
			where += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			where += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
		if filters.GameID != nil {
			where += " AND game_id = ?"
			args = append(args, *filters.GameID)
		}
	} else {
		where = "WHERE user_id = ?"
		args = []interface{}{userIDStr}

		if filters != nil {
			if filters.DateFrom != nil {
				where += " AND created_at >= ?"
				args = append(args, *filters.DateFrom)
			}
			if filters.DateTo != nil {
				where += " AND created_at <= ?"
				args = append(args, *filters.DateTo)
			}
			if filters.TransactionType != nil {
				where += " AND transaction_type = ?"
				args = append(args, *filters.TransactionType)
			}
			if filters.GameID != nil {
				where += " AND game_id = ?"
				args = append(args, *filters.GameID)
			}
			if filters.Status != nil {
				where += " AND status = ?"
				args = append(args, *filters.Status)
			}
		}
	}

	var countQuery string
	if isGroove {
		countQuery = `
		SELECT COUNT(*)
		FROM (
			SELECT 
				id,
				ROW_NUMBER() OVER (
					PARTITION BY id
					ORDER BY 
						(status = 'completed') DESC,
						updated_at DESC,
						created_at DESC
				) AS rn,
				status
			FROM tucanbit_analytics.transactions
			` + " " + where + `
		) deduped
		WHERE rn = 1 AND status = 'completed'
	`
	} else {
		countQuery = `
		SELECT COUNT(DISTINCT id)
		FROM (
			SELECT id,
				ROW_NUMBER() OVER (PARTITION BY id ORDER BY 
					CASE status 
						WHEN 'completed' THEN 1 
						ELSE 2 
					END,
					created_at DESC
				) as rn
			FROM tucanbit_analytics.transactions
			` + " " + where + `
		) deduped
		WHERE rn = 1
	`
	}

	var total64 uint64
	if err := s.clickhouse.QueryRow(ctx, countQuery, args...).Scan(&total64); err != nil {
		return nil, 0, fmt.Errorf("failed to count user transactions: %w", err)
	}
	total := int(total64)

	dataQuery := `
		SELECT 
			id, user_id, transaction_type, amount, currency, status,
			game_id, game_name, provider, session_id, round_id,
			bet_amount, win_amount, net_result, balance_before, balance_after,
			payment_method, external_transaction_id, metadata, created_at, updated_at
		FROM (
			SELECT 
				id, user_id, transaction_type, amount, currency, status,
				game_id, game_name, provider, session_id, round_id,
				bet_amount, win_amount, net_result, balance_before, balance_after,
				payment_method, external_transaction_id, metadata, created_at, updated_at,
				ROW_NUMBER() OVER (
					PARTITION BY id
					ORDER BY 
						(status = 'completed') DESC,
						updated_at DESC,
						created_at DESC
				) as rn
			FROM tucanbit_analytics.transactions
			` + " " + where + `
		) deduped
	`

	if isGroove {
		dataQuery += " WHERE rn = 1 AND status = 'completed'"
	} else {
		dataQuery += " WHERE rn = 1"
	}

	dataQuery += " ORDER BY created_at DESC"

	dataArgs := append([]interface{}{}, args...)
	if filters != nil && filters.Limit > 0 {
		dataQuery += " LIMIT ?"
		dataArgs = append(dataArgs, filters.Limit)
	}
	if filters != nil && filters.Offset > 0 {
		dataQuery += " OFFSET ?"
		dataArgs = append(dataArgs, filters.Offset)
	}

	rows, err := s.clickhouse.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user transactions: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
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

		// If net_result is less than or equal to zero, set it to zero
		if transaction.NetResult != nil && transaction.NetResult.LessThanOrEqual(decimal.Zero) {
			zero := decimal.Zero
			transaction.NetResult = &zero
		}

		transaction.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse user ID: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	return transactions, total, nil
}

// GetUserRakebackTransactions implements the /analytics/users/{user_id}/rakeback endpoint.
func (s *AnalyticsStorageImpl) GetUserRakebackTransactions(ctx context.Context, userID uuid.UUID, filters *dto.RakebackFilters) ([]*dto.RakebackTransaction, int, dto.UserRakebackTotals, error) {
	userIDStr := userID.String()

	// Map API transaction_type to underlying ClickHouse values
	var chType string
	switch filters.TransactionType {
	case "earned":
		chType = "cashback_earning"
	case "claimed":
		chType = "cashback_claim"
	default:
		chType = "cashback_earning"
	}

	// Base WHERE clause
	where := "WHERE user_id = ? AND transaction_type = ?"
	args := []interface{}{userIDStr, chType}

	if filters.Status != nil && *filters.Status != "" {
		where += " AND status = ?"
		args = append(args, *filters.Status)
	}
	if filters.BrandID != nil {
		// brand_id is not stored in cashback_analytics, so brand isolation must
		// be enforced at insert-time; we keep BrandID in filters for future use.
	}

	// Count and totals queries differ slightly between earned and claimed
	var totalCount uint64
	var totals dto.UserRakebackTotals

	// For "earned" transactions, fetch claim transactions to extract claimed earning IDs
	var claimedEarningIDs map[string]bool
	if filters.TransactionType == "earned" {
		claimedEarningIDs = make(map[string]bool)
		claimsQuery := `
			SELECT claimed_earnings
			FROM tucanbit_analytics.cashback_analytics
			WHERE user_id = ? 
				AND transaction_type = 'cashback_claim'
				AND claimed_earnings IS NOT NULL
				AND claimed_earnings != ''
		`
		claimsRows, err := s.clickhouse.Query(ctx, claimsQuery, userIDStr)
		if err == nil {
			defer claimsRows.Close()
			for claimsRows.Next() {
				var claimedEarningsJSON string
				if err := claimsRows.Scan(&claimedEarningsJSON); err == nil && claimedEarningsJSON != "" {
					// Parse JSON: claimed_earnings is a JSON object where keys are earning IDs
					// Example: {"earning_id_1": "0.5", "earning_id_2": "1.0"}
					// We extract the keys (earning IDs)
					var claimedMap map[string]interface{}
					if err := json.Unmarshal([]byte(claimedEarningsJSON), &claimedMap); err == nil {
						for earningID := range claimedMap {
							claimedEarningIDs[earningID] = true
						}
					}
				}
			}
		}
	}

	// Common count query
	countQuery := `
		SELECT COUNT(*)
		FROM tucanbit_analytics.cashback_analytics
	` + " " + where

	if err := s.clickhouse.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, dto.UserRakebackTotals{}, fmt.Errorf("failed to count rakeback rows: %w", err)
	}

	// Totals query (earned + claimed + available)
	totalsQuery := `
		SELECT
			COUNTIf(transaction_type = 'cashback_earning') AS total_earned_count,
			COALESCE(SUMIf(amount, transaction_type = 'cashback_earning'), 0) AS total_earned_amount,
			COUNTIf(transaction_type = 'cashback_claim') AS total_claimed_count,
			COALESCE(SUMIf(amount, transaction_type = 'cashback_claim'), 0) AS total_claimed_amount,
			COALESCE(
				SUMIf(available_amount, transaction_type = 'cashback_earning'),
				0
			) AS available_rakeback
		FROM tucanbit_analytics.cashback_analytics
		WHERE user_id = ?
	`

	var earnedCount, claimedCount uint64
	if err := s.clickhouse.QueryRow(ctx, totalsQuery, userIDStr).Scan(
		&earnedCount,
		&totals.TotalEarnedAmount,
		&claimedCount,
		&totals.TotalClaimedAmount,
		&totals.AvailableRakeback,
	); err != nil {
		return nil, 0, dto.UserRakebackTotals{}, fmt.Errorf("failed to get rakeback totals: %w", err)
	}
	totals.TotalEarnedCount = earnedCount
	totals.TotalClaimedCount = claimedCount

	// Data query
	dataQuery := `
		SELECT 
			id,
			user_id,
			transaction_type,
			amount,
			currency,
			status,
			game_id,
			game_name,
			provider,
			processing_fee,
			net_amount,
			claimed_at,
			claimed_earnings,
			claim_id,
			earning_id,
			created_at,
			updated_at
		FROM tucanbit_analytics.cashback_analytics
	` + " " + where + " ORDER BY created_at DESC"

	dataArgs := append([]interface{}{}, args...)
	if filters.Limit > 0 {
		dataQuery += " LIMIT ?"
		dataArgs = append(dataArgs, filters.Limit)
	}
	if filters.Offset > 0 {
		dataQuery += " OFFSET ?"
		dataArgs = append(dataArgs, filters.Offset)
	}

	rows, err := s.clickhouse.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, dto.UserRakebackTotals{}, fmt.Errorf("failed to query rakeback transactions: %w", err)
	}
	defer rows.Close()

	var result []*dto.RakebackTransaction
	for rows.Next() {
		var row dto.RakebackTransaction
		var userIDStrLocal, chTypeLocal string

		err := rows.Scan(
			&row.ID,
			&userIDStrLocal,
			&chTypeLocal,
			&row.RakebackAmount,
			&row.Currency,
			&row.Status,
			&row.GameID,
			&row.GameName,
			&row.Provider,
			&row.ProcessingFee,
			&row.NetAmount,
			&row.ClaimedAt,
			&row.ClaimedEarnings,
			&row.ClaimID,
			&row.EarningID,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, 0, dto.UserRakebackTotals{}, fmt.Errorf("failed to scan rakeback row: %w", err)
		}

		row.UserID, err = uuid.Parse(userIDStrLocal)
		if err != nil {
			return nil, 0, dto.UserRakebackTotals{}, fmt.Errorf("failed to parse user ID: %w", err)
		}

		// Map ClickHouse transaction_type back to API-level "earned"/"claimed"
		if chTypeLocal == "cashback_earning" {
			row.TransactionType = "earned"
			// For earned transactions, always override status based on claimed earnings logic
			// Status should be "available" or "claimed", not "completed"
			if filters.TransactionType == "earned" {
				if row.EarningID != nil && claimedEarningIDs[*row.EarningID] {
					// Mark as "claimed" if earning_id appears in any claim's claimed_earnings
					row.Status = "claimed"
				} else {
					// Otherwise set to "available" (earned but not yet claimed)
					row.Status = "available"
				}
			}
		} else if chTypeLocal == "cashback_claim" {
			row.TransactionType = "claimed"
		} else {
			row.TransactionType = chTypeLocal
		}

		result = append(result, &row)
	}

	return result, int(totalCount), totals, nil
}

// GetUserTips implements /analytics/users/{account_id}/tips backed by transactions.
func (s *AnalyticsStorageImpl) GetUserTips(ctx context.Context, userID uuid.UUID, filters *dto.TipFilters) ([]*dto.TipTransaction, int, error) {
	userIDStr := userID.String()

	where := "WHERE user_id = ? AND transaction_type IN ('tip_sent', 'tip_received')"
	args := []interface{}{userIDStr}

	if filters.DateFrom != nil {
		where += " AND created_at >= ?"
		args = append(args, *filters.DateFrom)
	}
	if filters.DateTo != nil {
		where += " AND created_at <= ?"
		args = append(args, *filters.DateTo)
	}
	if filters.Status != nil && *filters.Status != "" {
		where += " AND status = ?"
		args = append(args, *filters.Status)
	}
	if filters.BrandID != nil {
		where += " AND brand_id = ?"
		args = append(args, *filters.BrandID)
	}

	// Count
	countQuery := `
		SELECT COUNT(*)
		FROM tucanbit_analytics.transactions
	` + " " + where

	var total uint64
	if err := s.clickhouse.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tips: %w", err)
	}

	// Data
	dataQuery := `
		SELECT 
			id,
			user_id,
			transaction_type,
			amount,
			currency,
			status,
			balance_before,
			balance_after,
			external_transaction_id,
			metadata,
			created_at
		FROM tucanbit_analytics.transactions
	` + " " + where + " ORDER BY created_at DESC"

	dataArgs := append([]interface{}{}, args...)
	if filters.Limit > 0 {
		dataQuery += " LIMIT ?"
		dataArgs = append(dataArgs, filters.Limit)
	}
	if filters.Offset > 0 {
		dataQuery += " OFFSET ?"
		dataArgs = append(dataArgs, filters.Offset)
	}

	rows, err := s.clickhouse.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tips: %w", err)
	}
	defer rows.Close()

	var tips []*dto.TipTransaction
	for rows.Next() {
		var t dto.TipTransaction
		var userIDStrLocal string

		if err := rows.Scan(
			&t.ID,
			&userIDStrLocal,
			&t.TransactionType,
			&t.Amount,
			&t.Currency,
			&t.Status,
			&t.BalanceBefore,
			&t.BalanceAfter,
			&t.ExternalTransactionID,
			&t.Metadata,
			&t.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tip row: %w", err)
		}

		t.UserID, err = uuid.Parse(userIDStrLocal)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse user ID: %w", err)
		}

		tips = append(tips, &t)
	}

	return tips, int(total), nil
}

// implements /analytics/users/{user_id}/welcome_bonus backed by PostgreSQL groove_transactions.
func (s *AnalyticsStorageImpl) GetUserWelcomeBonus(ctx context.Context, userID uuid.UUID, filters *dto.WelcomeBonusFilters) ([]*dto.WelcomeBonusTransaction, int, error) {
	if s.pgPool == nil {
		return nil, 0, fmt.Errorf("PostgreSQL pool not available for welcome bonus queries")
	}

	// Build WHERE clause for PostgreSQL
	where := "WHERE ga.user_id = $1 AND gt.type = 'welcome_bonus'"
	args := []interface{}{userID}

	argIndex := 2

	if filters != nil {
		if filters.DateFrom != nil {
			where += fmt.Sprintf(" AND gt.created_at >= $%d", argIndex)
			args = append(args, *filters.DateFrom)
			argIndex++
		}
		if filters.DateTo != nil {
			where += fmt.Sprintf(" AND gt.created_at <= $%d", argIndex)
			args = append(args, *filters.DateTo)
			argIndex++
		}
		if filters.Status != nil && *filters.Status != "" {
			where += fmt.Sprintf(" AND gt.status = $%d", argIndex)
			args = append(args, *filters.Status)
			argIndex++
		}
		if filters.BrandID != nil {
			where += fmt.Sprintf(" AND gt.brand_id = $%d", argIndex)
			args = append(args, *filters.BrandID)
			argIndex++
		}
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM groove_transactions gt
		JOIN groove_accounts ga ON gt.account_id = ga.account_id
		JOIN users u ON ga.user_id = u.id
	` + " " + where + " AND gt.brand_id = u.brand_id"

	s.logger.Debug("Executing GetUserWelcomeBonus count query",
		zap.String("user_id", userID.String()),
		zap.String("query", countQuery),
		zap.Any("args", args))

	var total int
	if err := s.pgPool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		s.logger.Error("Failed to count user welcome bonus transactions",
			zap.String("user_id", userID.String()),
			zap.String("query", countQuery),
			zap.Any("args", args),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count welcome bonus transactions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT 
			gt.transaction_id,
			gt.account_id,
			gt.amount,
			gt.currency,
			gt.status,
			COALESCE(gt.balance_before, 0) as balance_before,
			COALESCE(gt.balance_after, 0) as balance_after,
			gt.account_transaction_id,
			gt.metadata::text,
			gt.created_at
		FROM groove_transactions gt
		JOIN groove_accounts ga ON gt.account_id = ga.account_id
		JOIN users u ON ga.user_id = u.id
	` + " " + where + " AND gt.brand_id = u.brand_id ORDER BY gt.created_at DESC"

	dataArgs := append([]interface{}{}, args...)
	if filters != nil {
		if filters.Limit > 0 {
			dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
			dataArgs = append(dataArgs, filters.Limit)
			argIndex++
		} else {
			dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
			dataArgs = append(dataArgs, 100)
			argIndex++
		}
		if filters.Offset > 0 {
			dataQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			dataArgs = append(dataArgs, filters.Offset)
		}
	} else {
		dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		dataArgs = append(dataArgs, 100)
	}

	s.logger.Debug("Executing GetUserWelcomeBonus data query",
		zap.String("user_id", userID.String()),
		zap.String("query", dataQuery),
		zap.Any("args", dataArgs))

	rows, err := s.pgPool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		s.logger.Error("Failed to query user welcome bonus transactions",
			zap.String("user_id", userID.String()),
			zap.String("query", dataQuery),
			zap.Any("args", dataArgs),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query welcome bonus transactions: %w", err)
	}
	defer rows.Close()

	var welcomeBonuses []*dto.WelcomeBonusTransaction
	for rows.Next() {
		var wb dto.WelcomeBonusTransaction
		var accountID string
		var metadataStr *string

		if err := rows.Scan(
			&wb.ID,
			&accountID,
			&wb.Amount,
			&wb.Currency,
			&wb.Status,
			&wb.BalanceBefore,
			&wb.BalanceAfter,
			&wb.ExternalTransactionID,
			&metadataStr,
			&wb.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan welcome bonus row: %w", err)
		}

		wb.UserID = userID
		wb.TransactionType = "welcome_bonus"
		if metadataStr != nil {
			wb.Metadata = metadataStr
		}

		welcomeBonuses = append(welcomeBonuses, &wb)
	}

	return welcomeBonuses, total, nil
}

func (s *AnalyticsStorageImpl) GetWelcomeBonusTransactions(ctx context.Context, filters *dto.WelcomeBonusFilters) ([]*dto.WelcomeBonusTransaction, int, error) {
	if s.pgPool == nil {
		return nil, 0, fmt.Errorf("PostgreSQL pool not available for welcome bonus queries")
	}

	// Build WHERE clause for PostgreSQL
	where := "WHERE gt.type = 'welcome_bonus'"
	args := []interface{}{}
	argIndex := 1

	if filters != nil {
		if filters.UserID != nil {
			where += fmt.Sprintf(" AND ga.user_id = $%d", argIndex)
			args = append(args, *filters.UserID)
			argIndex++
		}
		if filters.DateFrom != nil {
			where += fmt.Sprintf(" AND gt.created_at >= $%d", argIndex)
			args = append(args, *filters.DateFrom)
			argIndex++
		}
		if filters.DateTo != nil {
			where += fmt.Sprintf(" AND gt.created_at <= $%d", argIndex)
			args = append(args, *filters.DateTo)
			argIndex++
		}
		if filters.Status != nil && *filters.Status != "" {
			where += fmt.Sprintf(" AND gt.status = $%d", argIndex)
			args = append(args, *filters.Status)
			argIndex++
		}
		if filters.BrandID != nil {
			where += fmt.Sprintf(" AND gt.brand_id = $%d", argIndex)
			args = append(args, *filters.BrandID)
			argIndex++
		}
		if filters.MinAmount != nil {
			where += fmt.Sprintf(" AND gt.amount >= $%d", argIndex)
			args = append(args, *filters.MinAmount)
			argIndex++
		}
		if filters.MaxAmount != nil {
			where += fmt.Sprintf(" AND gt.amount <= $%d", argIndex)
			args = append(args, *filters.MaxAmount)
			argIndex++
		}
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM groove_transactions gt
		JOIN groove_accounts ga ON gt.account_id = ga.account_id
		JOIN users u ON ga.user_id = u.id
	` + " " + where
	if filters == nil || filters.BrandID == nil {
		countQuery += " AND gt.brand_id = u.brand_id"
	}

	s.logger.Debug("Executing welcome bonus count query",
		zap.String("query", countQuery),
		zap.Any("args", args))

	var total int
	if err := s.pgPool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		s.logger.Error("Failed to count welcome bonus transactions",
			zap.String("query", countQuery),
			zap.Any("args", args),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count welcome bonus transactions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT 
			gt.transaction_id,
			gt.account_id,
			ga.user_id,
			u.username,
			u.email,
			gt.amount,
			gt.currency,
			gt.status,
			COALESCE(gt.balance_before, 0) as balance_before,
			COALESCE(gt.balance_after, 0) as balance_after,
			gt.account_transaction_id,
			gt.metadata::text,
			gt.created_at
		FROM groove_transactions gt
		JOIN groove_accounts ga ON gt.account_id = ga.account_id
		JOIN users u ON ga.user_id = u.id
	` + " " + where
	if filters == nil || filters.BrandID == nil {
		dataQuery += " AND gt.brand_id = u.brand_id"
	}
	dataQuery += " ORDER BY gt.created_at DESC"

	dataArgs := append([]interface{}{}, args...)
	if filters != nil {
		if filters.Limit > 0 {
			dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
			dataArgs = append(dataArgs, filters.Limit)
			argIndex++
		} else {
			dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
			dataArgs = append(dataArgs, 100)
			argIndex++
		}
		if filters.Offset > 0 {
			dataQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			dataArgs = append(dataArgs, filters.Offset)
		}
	} else {
		dataQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		dataArgs = append(dataArgs, 100)
	}

	s.logger.Debug("Executing welcome bonus data query",
		zap.String("query", dataQuery),
		zap.Any("args", dataArgs))

	rows, err := s.pgPool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		s.logger.Error("Failed to query welcome bonus transactions",
			zap.String("query", dataQuery),
			zap.Any("args", dataArgs),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query welcome bonus transactions: %w", err)
	}
	defer rows.Close()

	var welcomeBonuses []*dto.WelcomeBonusTransaction
	for rows.Next() {
		var wb dto.WelcomeBonusTransaction
		var accountID string
		var userIDLocal uuid.UUID
		var username sql.NullString
		var email sql.NullString
		var metadataStr *string

		if err := rows.Scan(
			&wb.ID,
			&accountID,
			&userIDLocal,
			&username,
			&email,
			&wb.Amount,
			&wb.Currency,
			&wb.Status,
			&wb.BalanceBefore,
			&wb.BalanceAfter,
			&wb.ExternalTransactionID,
			&metadataStr,
			&wb.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan welcome bonus row: %w", err)
		}

		wb.UserID = userIDLocal
		if username.Valid {
			wb.Username = &username.String
		}
		if email.Valid {
			wb.Email = &email.String
		}
		wb.TransactionType = "welcome_bonus"
		if metadataStr != nil {
			wb.Metadata = metadataStr
		}

		welcomeBonuses = append(welcomeBonuses, &wb)
	}

	return welcomeBonuses, total, nil
}

// implements totals for game transactions.
func (s *AnalyticsStorageImpl) GetUserTransactionsTotals(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) (*dto.UserTransactionsTotals, error) {
	userIDStr := userID.String()

	where := "WHERE user_id = ?"
	args := []interface{}{userIDStr}

	isGroove := false
	if filters != nil && filters.TransactionType != nil {
		if *filters.TransactionType == "groove_bet" || *filters.TransactionType == "groove_win" {
			isGroove = true
		}
	}

	if isGroove {
		where += " AND transaction_type = ? AND bet_amount IS NOT NULL AND (status = 'completed' OR status = 'pending')"
		args = append(args, *filters.TransactionType)
	} else if filters != nil && filters.TransactionType != nil {
		where += " AND transaction_type = ?"
		args = append(args, *filters.TransactionType)
	}

	if filters != nil {
		if filters.DateFrom != nil {
			where += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			where += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
		if filters.Status != nil {
			where += " AND status = ?"
			args = append(args, *filters.Status)
		}
	}

	query := `
		SELECT 
			COUNT(*) AS total_count,
			COALESCE(SUMIf(bet_amount, transaction_type IN ('groove_bet','bet')), 0) AS total_bet_amount,
			COALESCE(SUMIf(win_amount, transaction_type IN ('groove_bet','bet','groove_win','win')), 0) AS total_win_amount
		FROM tucanbit_analytics.transactions
	` + " " + where

	var totals dto.UserTransactionsTotals
	if err := s.clickhouse.QueryRow(ctx, query, args...).Scan(
		&totals.TotalCount,
		&totals.TotalBetAmount,
		&totals.TotalWinAmount,
	); err != nil {
		return nil, fmt.Errorf("failed to get transaction totals: %w", err)
	}

	totals.NetResult = totals.TotalBetAmount.Sub(totals.TotalWinAmount)
	return &totals, nil
}

// wraps GetUserRakebackTransactions totals for convenience.
func (s *AnalyticsStorageImpl) GetUserRakebackTotals(ctx context.Context, userID uuid.UUID, filters *dto.RakebackFilters) (*dto.UserRakebackTotals, error) {
	_, _, totals, err := s.GetUserRakebackTransactions(ctx, userID, filters)
	if err != nil {
		return nil, err
	}
	return &totals, nil
}

// implements totals for tips.
func (s *AnalyticsStorageImpl) GetUserTipsTotals(ctx context.Context, userID uuid.UUID, filters *dto.TipFilters) (*dto.UserTipsTotals, error) {
	userIDStr := userID.String()

	where := "WHERE user_id = ? AND transaction_type IN ('tip_sent','tip_received')"
	args := []interface{}{userIDStr}

	if filters != nil {
		if filters.DateFrom != nil {
			where += " AND created_at >= ?"
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			where += " AND created_at <= ?"
			args = append(args, *filters.DateTo)
		}
		if filters.Status != nil {
			where += " AND status = ?"
			args = append(args, *filters.Status)
		}
		if filters.BrandID != nil {
			where += " AND brand_id = ?"
			args = append(args, *filters.BrandID)
		}
	}

	query := `
		SELECT 
			COUNT(*) AS total_tips_count,
			COUNTIf(transaction_type = 'tip_sent') AS total_sent_count,
			COALESCE(SUMIf(amount, transaction_type = 'tip_sent'), 0) AS total_sent_amount,
			COUNTIf(transaction_type = 'tip_received') AS total_received_count,
			COALESCE(SUMIf(amount, transaction_type = 'tip_received'), 0) AS total_received_amount
		FROM tucanbit_analytics.transactions
	` + " " + where

	var totals dto.UserTipsTotals
	if err := s.clickhouse.QueryRow(ctx, query, args...).Scan(
		&totals.TotalTipsCount,
		&totals.TotalSentCount,
		&totals.TotalSentAmount,
		&totals.TotalReceivedCount,
		&totals.TotalReceivedAmount,
	); err != nil {
		return nil, fmt.Errorf("failed to get tips totals: %w", err)
	}

	totals.NetTips = totals.TotalReceivedAmount.Sub(totals.TotalSentAmount)
	return &totals, nil
}

func (s *AnalyticsStorageImpl) GetUserWelcomeBonusTotals(ctx context.Context, userID uuid.UUID, filters *dto.WelcomeBonusFilters) (*dto.UserWelcomeBonusTotals, error) {
	if s.pgPool == nil {
		return nil, fmt.Errorf("PostgreSQL pool not available for welcome bonus queries")
	}

	// Build WHERE clause for PostgreSQL
	where := "WHERE ga.user_id = $1 AND gt.type = 'welcome_bonus'"
	args := []interface{}{userID}
	argIndex := 2

	if filters != nil {
		if filters.DateFrom != nil {
			where += fmt.Sprintf(" AND gt.created_at >= $%d", argIndex)
			args = append(args, *filters.DateFrom)
			argIndex++
		}
		if filters.DateTo != nil {
			where += fmt.Sprintf(" AND gt.created_at <= $%d", argIndex)
			args = append(args, *filters.DateTo)
			argIndex++
		}
		if filters.Status != nil && *filters.Status != "" {
			where += fmt.Sprintf(" AND gt.status = $%d", argIndex)
			args = append(args, *filters.Status)
			argIndex++
		}
		if filters.BrandID != nil {
			where += fmt.Sprintf(" AND gt.brand_id = $%d", argIndex)
			args = append(args, *filters.BrandID)
			argIndex++
		}
	}

	query := `
		SELECT 
			COUNT(*) AS total_count,
			COALESCE(SUM(gt.amount), 0) AS total_amount
		FROM groove_transactions gt
		JOIN groove_accounts ga ON gt.account_id = ga.account_id
		JOIN users u ON ga.user_id = u.id
	` + " " + where + " AND gt.brand_id = u.brand_id"

	var totals dto.UserWelcomeBonusTotals
	if err := s.pgPool.QueryRow(ctx, query, args...).Scan(
		&totals.TotalCount,
		&totals.TotalAmount,
	); err != nil {
		s.logger.Error("Failed to get welcome bonus totals",
			zap.String("user_id", userID.String()),
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get welcome bonus totals: %w", err)
	}

	return &totals, nil
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
		zap.Uint64("totalTransactions", stats.TotalTransactions),
		zap.Uint64("depositsCount", stats.DepositsCount),
		zap.String("totalDeposits", stats.TotalDeposits.String()),
		zap.Uint64("activeUsers", stats.ActiveUsers))

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

func (s *AnalyticsStorageImpl) GetTransactionReport(ctx context.Context, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error) {
	query := `
		SELECT 
			id, user_id, transaction_type, amount, currency, status,
			game_id, game_name, provider, session_id, round_id,
			bet_amount, win_amount, net_result, balance_before, balance_after,
			payment_method, external_transaction_id, metadata, created_at, updated_at
		FROM transactions
		WHERE amount > 0
	`

	args := []interface{}{}

	// Add filters
	if filters != nil {
		if filters.UserID != nil {
			query += " AND user_id = ?"
			args = append(args, filters.UserID.String())
		}
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

	if filters != nil && filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := s.clickhouse.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction report: %w", err)
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
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

	gamingQuery, gamingArgs := s.buildGamingActivityQueryString(startOfDay, endOfDay, nil)

	gamingMetricsQuery := `
		SELECT 
			toString(ifNull(toUInt64(count()), 0)) as total_transactions,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as total_bets,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as total_wins,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExact(game_id)), 0)) as active_games,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))), 0)) as bet_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))), 0)) as win_count,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8), toDecimal64(0, 8))) as cashback_earned,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed
		FROM (
			` + gamingQuery + `
		) gaming_data
	`

	var totalTransactionsStr, totalBetsStr, totalWinsStr, activeUsersStr, activeGamesStr, betCountStr, winCountStr, cashbackEarnedStr, cashbackClaimedStr sql.NullString

	var err error
	if len(gamingArgs) > 0 {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery, gamingArgs...).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	} else {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get gaming metrics: %w", err)
	}

	totalBets := parseDecimalSafely(totalBetsStr.String)
	totalWins := parseDecimalSafely(totalWinsStr.String)
	cashbackEarned := parseDecimalSafely(cashbackEarnedStr.String)
	cashbackClaimed := parseDecimalSafely(cashbackClaimedStr.String)

	depositsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_deposits,
			toString(ifNull(toUInt64(count()), 0)) as deposit_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_depositors
		FROM tucanbit_financial.deposits
		WHERE status = 'completed'
			AND toDate(created_at) = ?
	`

	dateStr := date.Format("2006-01-02")
	var totalDepositsStr, depositCountStr, uniqueDepositorsStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, depositsQuery, dateStr).Scan(
		&totalDepositsStr,
		&depositCountStr,
		&uniqueDepositorsStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get deposits from ClickHouse", zap.Error(err))
	}

	withdrawalsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_withdrawals,
			toString(ifNull(toUInt64(count()), 0)) as withdrawal_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_withdrawers
		FROM tucanbit_financial.withdrawals
		WHERE status = 'completed'
			AND toDate(created_at) = ?
	`

	var totalWithdrawalsStr, withdrawalCountStr, uniqueWithdrawersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, withdrawalsQuery, dateStr).Scan(
		&totalWithdrawalsStr,
		&withdrawalCountStr,
		&uniqueWithdrawersStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get withdrawals from ClickHouse", zap.Error(err))
	}

	var totalDeposits, totalWithdrawals decimal.Decimal
	var depositCount, withdrawalCount, uniqueDepositors, uniqueWithdrawers uint64

	totalDeposits = parseDecimalSafely(totalDepositsStr.String)
	depositCount = parseUint64Safely(depositCountStr.String)
	uniqueDepositors = parseUint64Safely(uniqueDepositorsStr.String)

	totalWithdrawals = parseDecimalSafely(totalWithdrawalsStr.String)
	withdrawalCount = parseUint64Safely(withdrawalCountStr.String)
	uniqueWithdrawers = parseUint64Safely(uniqueWithdrawersStr.String)

	if s.pgPool != nil {
		pgStartOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		pgEndOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

		chDepositHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.deposits
			WHERE status = 'completed'
				AND toDate(created_at) = ?
				AND tx_hash != ''
		`
		chDepositHashes := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chDepositHashesQuery, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chDepositHashes[txHash] = true
				}
			}
		}

		if len(chDepositHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chDepositHashes))
			hashArgs := make([]interface{}, 0, len(chDepositHashes)+2)
			hashArgs = append(hashArgs, pgStartOfDay, pgEndOfDay)
			argIndex := 3
			for hash := range chDepositHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgDepositsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, hashArgs...).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) = ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		} else {
			pgDepositsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, pgStartOfDay, pgEndOfDay).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) = ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, pgStartOfDay, pgEndOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		}

		chWithdrawalHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.withdrawals
			WHERE status = 'completed'
				AND toDate(created_at) = ?
				AND tx_hash != ''
		`
		chWithdrawalHashes := make(map[string]bool)
		rows, err = s.clickhouse.Query(ctx, chWithdrawalHashesQuery, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chWithdrawalHashes[txHash] = true
				}
			}
		}

		if len(chWithdrawalHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chWithdrawalHashes))
			hashArgs := make([]interface{}, 0, len(chWithdrawalHashes)+2)
			hashArgs = append(hashArgs, pgStartOfDay, pgEndOfDay)
			argIndex := 3
			for hash := range chWithdrawalHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgWithdrawalsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, hashArgs...).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) = ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		} else {
			pgWithdrawalsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, pgStartOfDay, pgEndOfDay).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) = ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, pgStartOfDay, pgEndOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		}
	}

	var newUsers uint64
	chNewUsersQuery := `
		SELECT toString(ifNull(toUInt64(count()), 0)) as new_users
		FROM tucanbit_analytics.transactions
		WHERE transaction_type = 'registration'
			AND toDate(created_at) = ?
	`
	var newUsersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, chNewUsersQuery, dateStr).Scan(&newUsersStr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get new users from ClickHouse", zap.Error(err))
	}
	newUsers = parseUint64Safely(newUsersStr.String)

	if s.pgPool != nil {
		chRegisteredUsersQuery := `
			SELECT DISTINCT user_id
			FROM tucanbit_analytics.transactions
			WHERE transaction_type = 'registration'
				AND toDate(created_at) = ?
		`
		chRegisteredUsers := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chRegisteredUsersQuery, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					chRegisteredUsers[userIDStr] = true
				}
			}
		}

		if len(chRegisteredUsers) > 0 {
			userPlaceholders := make([]string, 0, len(chRegisteredUsers))
			userArgs := make([]interface{}, 1)
			userArgs[0] = date
			argIndex := 2
			for userID := range chRegisteredUsers {
				userPlaceholders = append(userPlaceholders, fmt.Sprintf("$%d", argIndex))
				userArgs = append(userArgs, userID)
				argIndex++
			}
			userFilter := fmt.Sprintf("AND id::text NOT IN (%s)", strings.Join(userPlaceholders, ","))

			pgNewUsersQuery := fmt.Sprintf(`
				SELECT COUNT(*) as new_users
				FROM users
				WHERE DATE(created_at) = $1
					%s
			`, userFilter)

			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, userArgs...).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		} else {
			pgNewUsersQuery := `
				SELECT COUNT(*) as new_users
				FROM users
				WHERE DATE(created_at) = $1
			`
			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, date).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		}
	}

	ggr := totalBets.Sub(totalWins)
	ngr := ggr.Sub(cashbackClaimed)
	netRevenue := ngr

	report := &dto.DailyReport{
		Date:              date,
		TotalTransactions: parseUint64Safely(totalTransactionsStr.String),
		TotalDeposits:     totalDeposits,
		TotalWithdrawals:  totalWithdrawals,
		TotalBets:         totalBets,
		TotalWins:         totalWins,
		NetRevenue:        netRevenue,
		ActiveUsers:       parseUint64Safely(activeUsersStr.String),
		ActiveGames:       parseUint64Safely(activeGamesStr.String),
		NewUsers:          newUsers,
		UniqueDepositors:  uniqueDepositors,
		UniqueWithdrawers: uniqueWithdrawers,
		DepositCount:      depositCount,
		WithdrawalCount:   withdrawalCount,
		BetCount:          parseUint64Safely(betCountStr.String),
		WinCount:          parseUint64Safely(winCountStr.String),
		CashbackEarned:    cashbackEarned,
		CashbackClaimed:   cashbackClaimed,
		AdminCorrections:  decimal.Zero,
	}

	topGames, err := s.GetTopGames(ctx, 5, &dto.DateRange{
		From: &startOfDay,
		To:   &endOfDay,
	})
	if err != nil {
		s.logger.Warn("Failed to get top games for daily report", zap.Error(err))
	} else {
		report.TopGames = make([]dto.GameStats, len(topGames))
		for i, game := range topGames {
			report.TopGames[i] = *game
		}
	}

	topPlayers, err := s.GetTopPlayers(ctx, 5, &dto.DateRange{
		From: &startOfDay,
		To:   &endOfDay,
	})
	if err != nil {
		s.logger.Warn("Failed to get top players for daily report", zap.Error(err))
	} else {
		report.TopPlayers = make([]dto.PlayerStats, len(topPlayers))
		for i, player := range topPlayers {
			report.TopPlayers[i] = *player
		}
	}

	return report, nil
}

func (s *AnalyticsStorageImpl) GetEnhancedDailyReport(ctx context.Context, date time.Time) (*dto.EnhancedDailyReport, error) {
	baseReport, err := s.GetDailyReport(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get base daily report: %w", err)
	}

	ggr := baseReport.TotalBets.Sub(baseReport.TotalWins)

	enhancedReport := &dto.EnhancedDailyReport{
		Date:              baseReport.Date,
		TotalTransactions: baseReport.TotalTransactions,
		TotalDeposits:     baseReport.TotalDeposits,
		TotalWithdrawals:  baseReport.TotalWithdrawals,
		TotalBets:         baseReport.TotalBets,
		TotalWins:         baseReport.TotalWins,
		GGR:               ggr,
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

	previousDay := date.AddDate(0, 0, -1)
	previousReport, err := s.GetDailyReport(ctx, previousDay)
	if err != nil {
		s.logger.Warn("Failed to get previous day report for comparison", zap.Error(err))
		enhancedReport.PreviousDayChange = dto.DailyReportComparison{}
	} else {
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

func (s *AnalyticsStorageImpl) calculatePercentageChange(current, previous *dto.DailyReport) dto.DailyReportComparison {
	currentGGR := current.TotalBets.Sub(current.TotalWins)
	previousGGR := previous.TotalBets.Sub(previous.TotalWins)

	return dto.DailyReportComparison{
		TotalTransactionsChange: s.calculatePercentageChangeDecimal(
			decimal.NewFromInt(int64(current.TotalTransactions)),
			decimal.NewFromInt(int64(previous.TotalTransactions)),
		),
		TotalDepositsChange:    s.calculatePercentageChangeDecimal(current.TotalDeposits, previous.TotalDeposits),
		TotalWithdrawalsChange: s.calculatePercentageChangeDecimal(current.TotalWithdrawals, previous.TotalWithdrawals),
		TotalBetsChange:        s.calculatePercentageChangeDecimal(current.TotalBets, previous.TotalBets),
		TotalWinsChange:        s.calculatePercentageChangeDecimal(current.TotalWins, previous.TotalWins),
		GGRChange:              s.calculatePercentageChangeDecimal(currentGGR, previousGGR),
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

// gets Month To Date data for the given date
func (s *AnalyticsStorageImpl) getMTDData(ctx context.Context, date time.Time) (*dto.DailyReportMTD, error) {
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC)

	gamingQuery, gamingArgs := s.buildGamingActivityQueryString(startOfMonth, endOfDay, nil)

	gamingMetricsQuery := `
		SELECT 
			toString(ifNull(toUInt64(count()), 0)) as total_transactions,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as total_bets,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as total_wins,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExact(game_id)), 0)) as active_games,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))), 0)) as bet_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))), 0)) as win_count,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8), toDecimal64(0, 8))) as cashback_earned,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed
		FROM (
			` + gamingQuery + `
		) gaming_data
	`

	var totalTransactionsStr, totalBetsStr, totalWinsStr, activeUsersStr, activeGamesStr, betCountStr, winCountStr, cashbackEarnedStr, cashbackClaimedStr sql.NullString

	var err error
	if len(gamingArgs) > 0 {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery, gamingArgs...).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	} else {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get gaming metrics: %w", err)
	}

	totalBets := parseDecimalSafely(totalBetsStr.String)
	totalWins := parseDecimalSafely(totalWinsStr.String)
	cashbackEarned := parseDecimalSafely(cashbackEarnedStr.String)
	cashbackClaimed := parseDecimalSafely(cashbackClaimedStr.String)

	startOfMonthStr := startOfMonth.Format("2006-01-02")
	dateStr := date.Format("2006-01-02")

	depositsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_deposits,
			toString(ifNull(toUInt64(count()), 0)) as deposit_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_depositors
		FROM tucanbit_financial.deposits
		WHERE status = 'completed'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`

	var totalDepositsStr, depositCountStr, uniqueDepositorsStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, depositsQuery, startOfMonthStr, dateStr).Scan(
		&totalDepositsStr,
		&depositCountStr,
		&uniqueDepositorsStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get deposits from ClickHouse", zap.Error(err))
	}

	withdrawalsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_withdrawals,
			toString(ifNull(toUInt64(count()), 0)) as withdrawal_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_withdrawers
		FROM tucanbit_financial.withdrawals
		WHERE status = 'completed'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`

	var totalWithdrawalsStr, withdrawalCountStr, uniqueWithdrawersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, withdrawalsQuery, startOfMonthStr, dateStr).Scan(
		&totalWithdrawalsStr,
		&withdrawalCountStr,
		&uniqueWithdrawersStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get withdrawals from ClickHouse", zap.Error(err))
	}

	var totalDeposits, totalWithdrawals decimal.Decimal
	var depositCount, withdrawalCount, uniqueDepositors, uniqueWithdrawers uint64

	totalDeposits = parseDecimalSafely(totalDepositsStr.String)
	depositCount = parseUint64Safely(depositCountStr.String)
	uniqueDepositors = parseUint64Safely(uniqueDepositorsStr.String)

	totalWithdrawals = parseDecimalSafely(totalWithdrawalsStr.String)
	withdrawalCount = parseUint64Safely(withdrawalCountStr.String)
	uniqueWithdrawers = parseUint64Safely(uniqueWithdrawersStr.String)

	if s.pgPool != nil {
		chDepositHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.deposits
			WHERE status = 'completed'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
				AND tx_hash != ''
		`
		chDepositHashes := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chDepositHashesQuery, startOfMonthStr, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chDepositHashes[txHash] = true
				}
			}
		}

		if len(chDepositHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chDepositHashes))
			hashArgs := make([]interface{}, 0, len(chDepositHashes)+2)
			hashArgs = append(hashArgs, startOfMonth, endOfDay)
			argIndex := 3
			for hash := range chDepositHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgDepositsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, hashArgs...).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, startOfMonthStr, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		} else {
			pgDepositsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, startOfMonth, endOfDay).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, startOfMonthStr, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, startOfMonth, endOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		}

		chWithdrawalHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.withdrawals
			WHERE status = 'completed'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
				AND tx_hash != ''
		`
		chWithdrawalHashes := make(map[string]bool)
		rows, err = s.clickhouse.Query(ctx, chWithdrawalHashesQuery, startOfMonthStr, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chWithdrawalHashes[txHash] = true
				}
			}
		}

		if len(chWithdrawalHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chWithdrawalHashes))
			hashArgs := make([]interface{}, 0, len(chWithdrawalHashes)+2)
			hashArgs = append(hashArgs, startOfMonth, endOfDay)
			argIndex := 3
			for hash := range chWithdrawalHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgWithdrawalsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, hashArgs...).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, startOfMonthStr, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		} else {
			pgWithdrawalsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, startOfMonth, endOfDay).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, startOfMonthStr, dateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, startOfMonth, endOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		}
	}

	var newUsers uint64
	chNewUsersQuery := `
		SELECT toString(ifNull(toUInt64(count()), 0)) as new_users
		FROM tucanbit_analytics.transactions
		WHERE transaction_type = 'registration'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`
	var newUsersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, chNewUsersQuery, startOfMonthStr, dateStr).Scan(&newUsersStr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get new users from ClickHouse", zap.Error(err))
	}
	newUsers = parseUint64Safely(newUsersStr.String)

	if s.pgPool != nil {
		chRegisteredUsersQuery := `
			SELECT DISTINCT user_id
			FROM tucanbit_analytics.transactions
			WHERE transaction_type = 'registration'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
		`
		chRegisteredUsers := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chRegisteredUsersQuery, startOfMonthStr, dateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					chRegisteredUsers[userIDStr] = true
				}
			}
		}

		if len(chRegisteredUsers) > 0 {
			userPlaceholders := make([]string, 0, len(chRegisteredUsers))
			userArgs := make([]interface{}, 2)
			userArgs[0] = startOfMonth
			userArgs[1] = endOfDay
			argIndex := 3
			for userID := range chRegisteredUsers {
				userPlaceholders = append(userPlaceholders, fmt.Sprintf("$%d", argIndex))
				userArgs = append(userArgs, userID)
				argIndex++
			}
			userFilter := fmt.Sprintf("AND id::text NOT IN (%s)", strings.Join(userPlaceholders, ","))

			pgNewUsersQuery := fmt.Sprintf(`
				SELECT COUNT(*) as new_users
				FROM users
				WHERE created_at >= $1
					AND created_at <= $2
					%s
			`, userFilter)

			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, userArgs...).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		} else {
			pgNewUsersQuery := `
				SELECT COUNT(*) as new_users
				FROM users
				WHERE created_at >= $1
					AND created_at <= $2
			`
			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, startOfMonth, endOfDay).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		}
	}

	ggr := totalBets.Sub(totalWins)
	netRevenue := ggr

	mtd := &dto.DailyReportMTD{
		TotalTransactions: parseUint64Safely(totalTransactionsStr.String),
		TotalDeposits:     totalDeposits,
		TotalWithdrawals:  totalWithdrawals,
		TotalBets:         totalBets,
		TotalWins:         totalWins,
		GGR:               ggr,
		NetRevenue:        netRevenue,
		ActiveUsers:       parseUint64Safely(activeUsersStr.String),
		ActiveGames:       parseUint64Safely(activeGamesStr.String),
		NewUsers:          newUsers,
		UniqueDepositors:  uniqueDepositors,
		UniqueWithdrawers: uniqueWithdrawers,
		DepositCount:      depositCount,
		WithdrawalCount:   withdrawalCount,
		BetCount:          parseUint64Safely(betCountStr.String),
		WinCount:          parseUint64Safely(winCountStr.String),
		CashbackEarned:    cashbackEarned,
		CashbackClaimed:   cashbackClaimed,
		AdminCorrections:  decimal.Zero,
	}

	return mtd, nil
}

// getSPLMData gets Same Period Last Month data for the given date
func (s *AnalyticsStorageImpl) getSPLMData(ctx context.Context, date time.Time) (*dto.DailyReportSPLM, error) {
	lastMonth := date.AddDate(0, -1, 0)
	startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfLastMonth := startOfLastMonth.AddDate(0, 1, -1)

	splmEndDate := time.Date(lastMonth.Year(), lastMonth.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	if splmEndDate.After(endOfLastMonth) {
		splmEndDate = endOfLastMonth
	}
	splmEndOfDay := time.Date(splmEndDate.Year(), splmEndDate.Month(), splmEndDate.Day(), 23, 59, 59, 999999999, time.UTC)

	gamingQuery, gamingArgs := s.buildGamingActivityQueryString(startOfLastMonth, splmEndOfDay, nil)

	gamingMetricsQuery := `
		SELECT 
			toString(ifNull(toUInt64(count()), 0)) as total_transactions,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as total_bets,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as total_wins,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExact(game_id)), 0)) as active_games,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))), 0)) as bet_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))), 0)) as win_count,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8), toDecimal64(0, 8))) as cashback_earned,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed
		FROM (
			` + gamingQuery + `
		) gaming_data
	`

	var totalTransactionsStr, totalBetsStr, totalWinsStr, activeUsersStr, activeGamesStr, betCountStr, winCountStr, cashbackEarnedStr, cashbackClaimedStr sql.NullString

	var err error
	if len(gamingArgs) > 0 {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery, gamingArgs...).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	} else {
		err = s.clickhouse.QueryRow(ctx, gamingMetricsQuery).Scan(
			&totalTransactionsStr,
			&totalBetsStr,
			&totalWinsStr,
			&activeUsersStr,
			&activeGamesStr,
			&betCountStr,
			&winCountStr,
			&cashbackEarnedStr,
			&cashbackClaimedStr,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get gaming metrics: %w", err)
	}

	totalBets := parseDecimalSafely(totalBetsStr.String)
	totalWins := parseDecimalSafely(totalWinsStr.String)
	cashbackEarned := parseDecimalSafely(cashbackEarnedStr.String)
	cashbackClaimed := parseDecimalSafely(cashbackClaimedStr.String)

	startOfLastMonthStr := startOfLastMonth.Format("2006-01-02")
	splmEndDateStr := splmEndDate.Format("2006-01-02")

	depositsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_deposits,
			toString(ifNull(toUInt64(count()), 0)) as deposit_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_depositors
		FROM tucanbit_financial.deposits
		WHERE status = 'completed'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`

	var totalDepositsStr, depositCountStr, uniqueDepositorsStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, depositsQuery, startOfLastMonthStr, splmEndDateStr).Scan(
		&totalDepositsStr,
		&depositCountStr,
		&uniqueDepositorsStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get deposits from ClickHouse", zap.Error(err))
	}

	withdrawalsQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as total_withdrawals,
			toString(ifNull(toUInt64(count()), 0)) as withdrawal_count,
			toString(ifNull(toUInt64(uniqExact(user_id)), 0)) as unique_withdrawers
		FROM tucanbit_financial.withdrawals
		WHERE status = 'completed'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`

	var totalWithdrawalsStr, withdrawalCountStr, uniqueWithdrawersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, withdrawalsQuery, startOfLastMonthStr, splmEndDateStr).Scan(
		&totalWithdrawalsStr,
		&withdrawalCountStr,
		&uniqueWithdrawersStr,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get withdrawals from ClickHouse", zap.Error(err))
	}

	var totalDeposits, totalWithdrawals decimal.Decimal
	var depositCount, withdrawalCount, uniqueDepositors, uniqueWithdrawers uint64

	totalDeposits = parseDecimalSafely(totalDepositsStr.String)
	depositCount = parseUint64Safely(depositCountStr.String)
	uniqueDepositors = parseUint64Safely(uniqueDepositorsStr.String)

	totalWithdrawals = parseDecimalSafely(totalWithdrawalsStr.String)
	withdrawalCount = parseUint64Safely(withdrawalCountStr.String)
	uniqueWithdrawers = parseUint64Safely(uniqueWithdrawersStr.String)

	if s.pgPool != nil {
		chDepositHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.deposits
			WHERE status = 'completed'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
				AND tx_hash != ''
		`
		chDepositHashes := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chDepositHashesQuery, startOfLastMonthStr, splmEndDateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chDepositHashes[txHash] = true
				}
			}
		}

		if len(chDepositHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chDepositHashes))
			hashArgs := make([]interface{}, 0, len(chDepositHashes)+2)
			hashArgs = append(hashArgs, startOfLastMonth, splmEndOfDay)
			argIndex := 3
			for hash := range chDepositHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgDepositsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, hashArgs...).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, startOfLastMonthStr, splmEndDateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		} else {
			pgDepositsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_deposits,
					COUNT(*) as deposit_count,
					COUNT(DISTINCT user_id) as unique_depositors
				FROM transactions
				WHERE transaction_type = 'deposit'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalDeposits decimal.Decimal
			var pgDepositCount, pgUniqueDepositors int64
			err = s.pgPool.QueryRow(ctx, pgDepositsQuery, startOfLastMonth, splmEndOfDay).Scan(
				&pgTotalDeposits,
				&pgDepositCount,
				&pgUniqueDepositors,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced deposits from PostgreSQL", zap.Error(err))
			} else {
				totalDeposits = totalDeposits.Add(pgTotalDeposits)
				depositCount += uint64(pgDepositCount)

				chDepositorsMap := make(map[string]bool)
				chDepositorsQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.deposits
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chDepositorsQuery, startOfLastMonthStr, splmEndDateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chDepositorsMap[userIDStr] = true
						}
					}
				}

				pgDepositorsQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'deposit'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgDepositorsQuery, startOfLastMonth, splmEndOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chDepositorsMap[userIDStr] {
								uniqueDepositors++
							}
						}
					}
				}
			}
		}

		chWithdrawalHashesQuery := `
			SELECT DISTINCT tx_hash
			FROM tucanbit_financial.withdrawals
			WHERE status = 'completed'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
				AND tx_hash != ''
		`
		chWithdrawalHashes := make(map[string]bool)
		rows, err = s.clickhouse.Query(ctx, chWithdrawalHashesQuery, startOfLastMonthStr, splmEndDateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txHash string
				if err := rows.Scan(&txHash); err == nil && txHash != "" {
					chWithdrawalHashes[txHash] = true
				}
			}
		}

		if len(chWithdrawalHashes) > 0 {
			hashPlaceholders := make([]string, 0, len(chWithdrawalHashes))
			hashArgs := make([]interface{}, 0, len(chWithdrawalHashes)+2)
			hashArgs = append(hashArgs, startOfLastMonth, splmEndOfDay)
			argIndex := 3
			for hash := range chWithdrawalHashes {
				hashPlaceholders = append(hashPlaceholders, fmt.Sprintf("$%d", argIndex))
				hashArgs = append(hashArgs, hash)
				argIndex++
			}
			hashFilter := fmt.Sprintf("AND tx_hash NOT IN (%s)", strings.Join(hashPlaceholders, ","))

			pgWithdrawalsQuery := fmt.Sprintf(`
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
					%s
			`, hashFilter)

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, hashArgs...).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, startOfLastMonthStr, splmEndDateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := fmt.Sprintf(`
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
						%s
				`, hashFilter)
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, hashArgs...)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		} else {
			pgWithdrawalsQuery := `
				SELECT 
					COALESCE(SUM(CASE WHEN usd_amount_cents IS NOT NULL THEN usd_amount_cents::decimal / 100 ELSE amount END), 0) as total_withdrawals,
					COUNT(*) as withdrawal_count,
					COUNT(DISTINCT user_id) as unique_withdrawers
				FROM transactions
				WHERE transaction_type = 'withdrawal'
					AND status = 'verified'
					AND created_at >= $1
					AND created_at <= $2
			`

			var pgTotalWithdrawals decimal.Decimal
			var pgWithdrawalCount, pgUniqueWithdrawers int64
			err = s.pgPool.QueryRow(ctx, pgWithdrawalsQuery, startOfLastMonth, splmEndOfDay).Scan(
				&pgTotalWithdrawals,
				&pgWithdrawalCount,
				&pgUniqueWithdrawers,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced withdrawals from PostgreSQL", zap.Error(err))
			} else {
				totalWithdrawals = totalWithdrawals.Add(pgTotalWithdrawals)
				withdrawalCount += uint64(pgWithdrawalCount)

				chWithdrawersMap := make(map[string]bool)
				chWithdrawersQuery := `
					SELECT DISTINCT toString(user_id) as user_id
					FROM tucanbit_financial.withdrawals
					WHERE status = 'completed'
						AND toDate(created_at) >= ?
						AND toDate(created_at) <= ?
				`
				chRows, err := s.clickhouse.Query(ctx, chWithdrawersQuery, startOfLastMonthStr, splmEndDateStr)
				if err == nil {
					defer chRows.Close()
					for chRows.Next() {
						var userIDStr string
						if err := chRows.Scan(&userIDStr); err == nil {
							chWithdrawersMap[userIDStr] = true
						}
					}
				}

				pgWithdrawersQuery := `
					SELECT DISTINCT user_id::text
					FROM transactions
					WHERE transaction_type = 'withdrawal'
						AND status = 'verified'
						AND created_at >= $1
						AND created_at <= $2
				`
				pgRows, err := s.pgPool.Query(ctx, pgWithdrawersQuery, startOfLastMonth, splmEndOfDay)
				if err == nil {
					defer pgRows.Close()
					for pgRows.Next() {
						var userIDStr string
						if err := pgRows.Scan(&userIDStr); err == nil {
							if !chWithdrawersMap[userIDStr] {
								uniqueWithdrawers++
							}
						}
					}
				}
			}
		}
	}

	var newUsers uint64
	chNewUsersQuery := `
		SELECT toString(ifNull(toUInt64(count()), 0)) as new_users
		FROM tucanbit_analytics.transactions
		WHERE transaction_type = 'registration'
			AND toDate(created_at) >= ?
			AND toDate(created_at) <= ?
	`
	var newUsersStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, chNewUsersQuery, startOfLastMonthStr, splmEndDateStr).Scan(&newUsersStr)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "no rows") {
		s.logger.Warn("Failed to get new users from ClickHouse", zap.Error(err))
	}
	newUsers = parseUint64Safely(newUsersStr.String)

	if s.pgPool != nil {
		chRegisteredUsersQuery := `
			SELECT DISTINCT user_id
			FROM tucanbit_analytics.transactions
			WHERE transaction_type = 'registration'
				AND toDate(created_at) >= ?
				AND toDate(created_at) <= ?
		`
		chRegisteredUsers := make(map[string]bool)
		rows, err := s.clickhouse.Query(ctx, chRegisteredUsersQuery, startOfLastMonthStr, splmEndDateStr)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					chRegisteredUsers[userIDStr] = true
				}
			}
		}

		if len(chRegisteredUsers) > 0 {
			userPlaceholders := make([]string, 0, len(chRegisteredUsers))
			userArgs := make([]interface{}, 2)
			userArgs[0] = startOfLastMonth
			userArgs[1] = splmEndOfDay
			argIndex := 3
			for userID := range chRegisteredUsers {
				userPlaceholders = append(userPlaceholders, fmt.Sprintf("$%d", argIndex))
				userArgs = append(userArgs, userID)
				argIndex++
			}
			userFilter := fmt.Sprintf("AND id::text NOT IN (%s)", strings.Join(userPlaceholders, ","))

			pgNewUsersQuery := fmt.Sprintf(`
				SELECT COUNT(*) as new_users
				FROM users
				WHERE created_at >= $1
					AND created_at <= $2
					%s
			`, userFilter)

			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, userArgs...).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		} else {
			pgNewUsersQuery := `
				SELECT COUNT(*) as new_users
				FROM users
				WHERE created_at >= $1
					AND created_at <= $2
			`
			var pgNewUsers int64
			err = s.pgPool.QueryRow(ctx, pgNewUsersQuery, startOfLastMonth, splmEndOfDay).Scan(&pgNewUsers)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("Failed to get unsynced new users from PostgreSQL", zap.Error(err))
			} else {
				newUsers += uint64(pgNewUsers)
			}
		}
	}

	ggr := totalBets.Sub(totalWins)
	netRevenue := ggr

	splm := &dto.DailyReportSPLM{
		TotalTransactions: parseUint64Safely(totalTransactionsStr.String),
		TotalDeposits:     totalDeposits,
		TotalWithdrawals:  totalWithdrawals,
		TotalBets:         totalBets,
		TotalWins:         totalWins,
		GGR:               ggr,
		NetRevenue:        netRevenue,
		ActiveUsers:       parseUint64Safely(activeUsersStr.String),
		ActiveGames:       parseUint64Safely(activeGamesStr.String),
		NewUsers:          newUsers,
		UniqueDepositors:  uniqueDepositors,
		UniqueWithdrawers: uniqueWithdrawers,
		DepositCount:      depositCount,
		WithdrawalCount:   withdrawalCount,
		BetCount:          parseUint64Safely(betCountStr.String),
		WinCount:          parseUint64Safely(winCountStr.String),
		CashbackEarned:    cashbackEarned,
		CashbackClaimed:   cashbackClaimed,
		AdminCorrections:  decimal.Zero,
	}

	return splm, nil
}

// calculates percentage change between MTD and SPLM
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
		GGRChange:              s.calculatePercentageChangeDecimal(mtd.GGR, splm.GGR),
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
			toDecimal64(sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8) as total_wins,
			toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')) - sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8) as net_revenue,
			toUInt64(uniqExact(user_id)) as player_count,
			toUInt64(uniqExact(session_id)) as session_count,
			if(countIf(transaction_type IN ('bet', 'groove_bet')) > 0, avgIf(amount, transaction_type IN ('bet', 'groove_bet')), 0) as avg_bet_amount,
			CASE 
				WHEN sumIf(amount, transaction_type IN ('bet', 'groove_bet')) > 0 
				THEN (sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')) / sumIf(amount, transaction_type IN ('bet', 'groove_bet'))) * 100
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
		// Filter by user IDs if provided
		if len(dateRange.UserIDs) > 0 {
			// Build IN clause with placeholders
			placeholders := make([]string, len(dateRange.UserIDs))
			for i := range dateRange.UserIDs {
				placeholders[i] = "?"
				args = append(args, dateRange.UserIDs[i].String())
			}
			query += " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
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
		// Filter by user IDs if provided
		if len(dateRange.UserIDs) > 0 {
			// Build IN clause with placeholders
			placeholders := make([]string, len(dateRange.UserIDs))
			for i := range dateRange.UserIDs {
				placeholders[i] = "?"
				args = append(args, dateRange.UserIDs[i].String())
			}
			query += " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
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
		s.logger.Error("ClickHouse query failed for GetTopPlayers",
			zap.Error(err),
			zap.String("query", query),
			zap.Any("args", args))
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

func (s *AnalyticsStorageImpl) GetDailyReportDataTable(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID) (*dto.DailyReportDataTableResponse, error) {
	startOfDay := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)

	var rows []dto.DailyReportDataTableRow
	currentDate := startOfDay

	for !currentDate.After(endOfDay) {
		dayReport, err := s.GetDailyReport(ctx, currentDate)
		if err != nil {
			s.logger.Warn("Failed to get daily report for data table",
				zap.Time("date", currentDate),
				zap.Error(err))
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		ggr := dayReport.TotalBets.Sub(dayReport.TotalWins)

		row := dto.DailyReportDataTableRow{
			Date:              currentDate.Format("2006-01-02"),
			NewUsers:          dayReport.NewUsers,
			UniqueDepositors:  dayReport.UniqueDepositors,
			UniqueWithdrawers: dayReport.UniqueWithdrawers,
			ActiveUsers:       dayReport.ActiveUsers,
			BetCount:          dayReport.BetCount,
			BetAmount:         dayReport.TotalBets,
			WinAmount:         dayReport.TotalWins,
			GGR:               ggr,
			CashbackEarned:    dayReport.CashbackEarned,
			CashbackClaimed:   dayReport.CashbackClaimed,
			DepositCount:      dayReport.DepositCount,
			DepositAmount:     dayReport.TotalDeposits,
			WithdrawalCount:   dayReport.WithdrawalCount,
			WithdrawalAmount:  dayReport.TotalWithdrawals,
			AdminCorrections:  dayReport.AdminCorrections,
		}

		rows = append(rows, row)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	var totals dto.DailyReportDataTableRow
	totals.Date = "Totals"
	for _, row := range rows {
		totals.NewUsers += row.NewUsers
		totals.UniqueDepositors += row.UniqueDepositors
		totals.UniqueWithdrawers += row.UniqueWithdrawers
		totals.ActiveUsers += row.ActiveUsers
		totals.BetCount += row.BetCount
		totals.BetAmount = totals.BetAmount.Add(row.BetAmount)
		totals.WinAmount = totals.WinAmount.Add(row.WinAmount)
		totals.GGR = totals.GGR.Add(row.GGR)
		totals.CashbackEarned = totals.CashbackEarned.Add(row.CashbackEarned)
		totals.CashbackClaimed = totals.CashbackClaimed.Add(row.CashbackClaimed)
		totals.DepositCount += row.DepositCount
		totals.DepositAmount = totals.DepositAmount.Add(row.DepositAmount)
		totals.WithdrawalCount += row.WithdrawalCount
		totals.WithdrawalAmount = totals.WithdrawalAmount.Add(row.WithdrawalAmount)
		totals.AdminCorrections = totals.AdminCorrections.Add(row.AdminCorrections)
	}

	return &dto.DailyReportDataTableResponse{
		Rows:   rows,
		Totals: totals,
	}, nil
}

func (s *AnalyticsStorageImpl) GetWeeklyReport(ctx context.Context, weekStart time.Time, userIDs []uuid.UUID) (*dto.WeeklyReport, error) {
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 999999999, time.UTC)

	dataTable, err := s.GetDailyReportDataTable(ctx, weekStart, weekEnd, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily breakdown: %w", err)
	}

	var weeklySummary dto.DailyReportDataTableRow
	weeklySummary.Date = "Weekly Total"
	for _, row := range dataTable.Rows {
		weeklySummary.NewUsers += row.NewUsers
		weeklySummary.UniqueDepositors += row.UniqueDepositors
		weeklySummary.UniqueWithdrawers += row.UniqueWithdrawers
		weeklySummary.ActiveUsers += row.ActiveUsers
		weeklySummary.BetCount += row.BetCount
		weeklySummary.BetAmount = weeklySummary.BetAmount.Add(row.BetAmount)
		weeklySummary.WinAmount = weeklySummary.WinAmount.Add(row.WinAmount)
		weeklySummary.GGR = weeklySummary.GGR.Add(row.GGR)
		weeklySummary.CashbackEarned = weeklySummary.CashbackEarned.Add(row.CashbackEarned)
		weeklySummary.CashbackClaimed = weeklySummary.CashbackClaimed.Add(row.CashbackClaimed)
		weeklySummary.DepositCount += row.DepositCount
		weeklySummary.DepositAmount = weeklySummary.DepositAmount.Add(row.DepositAmount)
		weeklySummary.WithdrawalCount += row.WithdrawalCount
		weeklySummary.WithdrawalAmount = weeklySummary.WithdrawalAmount.Add(row.WithdrawalAmount)
		weeklySummary.AdminCorrections = weeklySummary.AdminCorrections.Add(row.AdminCorrections)
	}

	uniqueDepositorsMap := make(map[string]bool)
	uniqueWithdrawersMap := make(map[string]bool)
	activeUsersMap := make(map[string]bool)

	for _, row := range dataTable.Rows {
		date, err := time.Parse("2006-01-02", row.Date)
		if err != nil {
			continue
		}

		chDepositorsQuery := `
			SELECT DISTINCT toString(user_id) as user_id
			FROM tucanbit_financial.deposits
			WHERE status = 'completed'
				AND toDate(created_at) = ?
		`
		rows, err := s.clickhouse.Query(ctx, chDepositorsQuery, row.Date)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					uniqueDepositorsMap[userIDStr] = true
				}
			}
		}

		chWithdrawersQuery := `
			SELECT DISTINCT toString(user_id) as user_id
			FROM tucanbit_financial.withdrawals
			WHERE status = 'completed'
				AND toDate(created_at) = ?
		`
		rows, err = s.clickhouse.Query(ctx, chWithdrawersQuery, row.Date)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					uniqueWithdrawersMap[userIDStr] = true
				}
			}
		}

		gamingQuery, gamingArgs := s.buildGamingActivityQueryString(
			time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
			time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC),
			userIDs,
		)

		activeUsersQuery := `
			SELECT DISTINCT toString(user_id) as user_id
			FROM (
				` + gamingQuery + `
			) gaming_data
		`
		rows, err = s.clickhouse.Query(ctx, activeUsersQuery, gamingArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				if err := rows.Scan(&userIDStr); err == nil {
					activeUsersMap[userIDStr] = true
				}
			}
		}
	}

	weeklySummary.UniqueDepositors = uint64(len(uniqueDepositorsMap))
	weeklySummary.UniqueWithdrawers = uint64(len(uniqueWithdrawersMap))
	weeklySummary.ActiveUsers = uint64(len(activeUsersMap))

	mtdData, err := s.getMTDData(ctx, weekEnd)
	if err != nil {
		s.logger.Warn("Failed to get MTD data for weekly report", zap.Error(err))
		mtdData = &dto.DailyReportMTD{}
	}

	splmData, err := s.getSPLMData(ctx, weekEnd)
	if err != nil {
		s.logger.Warn("Failed to get SPLM data for weekly report", zap.Error(err))
		splmData = &dto.DailyReportSPLM{}
	}

	var mtdvsSPLMChange dto.DailyReportComparison
	if mtdData != nil && splmData != nil {
		mtdvsSPLMChange = s.calculateMTDvsSPLMChange(mtdData, splmData)
	}

	report := &dto.WeeklyReport{
		WeekStart:         weekStart,
		WeekEnd:           weekEnd,
		NewUsers:          weeklySummary.NewUsers,
		UniqueDepositors:  weeklySummary.UniqueDepositors,
		UniqueWithdrawers: weeklySummary.UniqueWithdrawers,
		ActiveUsers:       weeklySummary.ActiveUsers,
		BetCount:          weeklySummary.BetCount,
		BetAmount:         weeklySummary.BetAmount,
		WinAmount:         weeklySummary.WinAmount,
		GGR:               weeklySummary.GGR,
		CashbackEarned:    weeklySummary.CashbackEarned,
		CashbackClaimed:   weeklySummary.CashbackClaimed,
		DepositCount:      weeklySummary.DepositCount,
		DepositAmount:     weeklySummary.DepositAmount,
		WithdrawalCount:   weeklySummary.WithdrawalCount,
		WithdrawalAmount:  weeklySummary.WithdrawalAmount,
		AdminCorrections:  weeklySummary.AdminCorrections,
		DailyBreakdown:    dataTable.Rows,
		MTD:               *mtdData,
		SPLM:              *splmData,
		MTDvsSPLMChange:   mtdvsSPLMChange,
	}

	return report, nil
}
