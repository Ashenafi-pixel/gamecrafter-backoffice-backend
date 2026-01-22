package analytics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

func parseUint64Safely(s string) uint64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseDecimalSafely(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func (s *AnalyticsStorageImpl) buildGamingActivityQueryString(dateFrom, dateTo time.Time, userIDs []uuid.UUID) (string, []interface{}) {
	startOfDay := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)

	startDateStr := startOfDay.Format("2006-01-02 15:04:05")
	endDateStr := endOfDay.Format("2006-01-02 15:04:05")
	args := []interface{}{}
	userFilter := ""
	if len(userIDs) > 0 {
		placeholders := make([]string, len(userIDs))
		for i := range userIDs {
			placeholders[i] = "?"
		}
		userFilter = " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
		for _, userID := range userIDs {
			args = append(args, userID.String())
		}
	}

	dateFilter := fmt.Sprintf(" AND created_at >= '%s' AND created_at <= '%s'", startDateStr, endDateStr)
	query := `
		SELECT 
			user_id,
			amount,
			transaction_type,
			created_at,
			game_id,
			win_amount,
			session_id,
			bet_amount
		FROM tucanbit_analytics.transactions
		WHERE transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
			AND (
				(transaction_type IN ('groove_bet', 'groove_win') AND (status = 'completed' OR (status = 'pending' AND bet_amount IS NOT NULL AND win_amount IS NOT NULL)))
				OR (transaction_type NOT IN ('groove_bet', 'groove_win') AND (status = 'completed' OR status IS NULL OR status = ''))
			)` + dateFilter + userFilter + `
		
		UNION ALL
		
		SELECT 
			user_id,
			amount,
			'cashback_earning' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_analytics.cashback_analytics
		WHERE transaction_type = 'cashback_earning'` + dateFilter + userFilter + `
		
		UNION ALL
		
		SELECT 
			user_id,
			amount,
			'cashback_claim' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_analytics.cashback_analytics
		WHERE transaction_type = 'cashback_claim'` + dateFilter + userFilter

	if len(userIDs) > 0 {
		numSections := 3
		finalArgs := make([]interface{}, 0, len(args)*numSections)
		for i := 0; i < numSections; i++ {
			finalArgs = append(finalArgs, args...)
		}
		return query, finalArgs
	}

	return query, []interface{}{}
}

func (s *AnalyticsStorageImpl) buildUnionAllQueryString(dateFrom, dateTo time.Time, userIDs []uuid.UUID) (string, []interface{}) {
	startOfDay := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)

	startDateStr := startOfDay.Format("2006-01-02 15:04:05")
	endDateStr := endOfDay.Format("2006-01-02 15:04:05")
	args := []interface{}{}
	userFilter := ""
	if len(userIDs) > 0 {
		placeholders := make([]string, len(userIDs))
		for i := range userIDs {
			placeholders[i] = "?"
		}
		userFilter = " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
		for _, userID := range userIDs {
			args = append(args, userID.String())
		}
	}

	dateFilter := fmt.Sprintf(" AND created_at >= '%s' AND created_at <= '%s'", startDateStr, endDateStr)
	query := `
		SELECT 
			user_id,
			usd_amount as amount,
			'deposit' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_financial.deposits
		WHERE toString(status) IN ('completed') AND toString(status) != ''` + dateFilter + userFilter + `
		
		UNION ALL
		
		-- Withdrawals from tucanbit_financial
		SELECT 
			user_id,
			usd_amount as amount,
			'withdrawal' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_financial.withdrawals
		WHERE toString(status) IN ('completed') AND toString(status) != ''` + dateFilter + userFilter + `
		
		UNION ALL
		
		-- Betting transactions from tucanbit_analytics
		SELECT 
			user_id,
			amount,
			transaction_type,
			created_at,
			game_id,
			win_amount,
			session_id,
			bet_amount
		FROM tucanbit_analytics.transactions
		WHERE transaction_type NOT IN ('deposit', 'withdrawal')
			AND (
				(transaction_type IN ('groove_bet', 'groove_win') AND (status = 'completed' OR (status = 'pending' AND bet_amount IS NOT NULL AND win_amount IS NOT NULL)))
				OR (transaction_type NOT IN ('groove_bet', 'groove_win') AND (status = 'completed' OR status IS NULL OR status = ''))
			)` + dateFilter + userFilter + `
		
		UNION ALL
		
		-- Cashback earned
		SELECT 
			user_id,
			amount,
			'cashback_earning' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_analytics.cashback_analytics
		WHERE transaction_type = 'cashback_earning'` + dateFilter + userFilter + `
		
		UNION ALL
		
		-- Cashback claimed
		SELECT 
			user_id,
			amount,
			'cashback_claim' as transaction_type,
			created_at,
			NULL as game_id,
			NULL as win_amount,
			NULL as session_id,
			NULL as bet_amount
		FROM tucanbit_analytics.cashback_analytics
		WHERE transaction_type = 'cashback_claim'` + dateFilter + userFilter

	if len(userIDs) > 0 {
		numSections := 5
		finalArgs := make([]interface{}, 0, len(args)*numSections)
		for i := 0; i < numSections; i++ {
			finalArgs = append(finalArgs, args...)
		}
		return query, finalArgs
	}

	return query, []interface{}{}
}

func (s *AnalyticsStorageImpl) GetDashboardOverview(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID, includeDailyBreakdown bool) (*dto.DashboardOverviewResponse, error) {
	unionAllQuery, userArgs := s.buildUnionAllQueryString(dateFrom, dateTo, userIDs)

	summaryQuery := `
		SELECT 
			toString(ifNull(toUInt64(count()), 0)) as total_transactions,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8), toDecimal64(0, 8))) as total_deposits,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8), toDecimal64(0, 8))) as total_withdrawals,
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
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type IN ('groove_bet', 'groove_win') OR session_id IS NOT NULL OR game_id IS NOT NULL)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExact(game_id)), 0)) as active_games
		FROM (
			` + unionAllQuery + `
		) combined_data
	`

	var summaryArgs []interface{}
	if len(userArgs) > 0 {
		summaryArgs = make([]interface{}, 0, len(userArgs)*5)
		for i := 0; i < 5; i++ {
			summaryArgs = append(summaryArgs, userArgs...)
		}
	}

	var totalTransactionsStr, totalDepositsStr, totalWithdrawalsStr, totalBetsStr, totalWinsStr, cashbackClaimedStr sql.NullString
	var activeUsersStr, activeGamesStr sql.NullString

	var err error
	if len(summaryArgs) > 0 {
		err = s.clickhouse.QueryRow(ctx, summaryQuery, summaryArgs...).Scan(
			&totalTransactionsStr,
			&totalDepositsStr,
			&totalWithdrawalsStr,
			&totalBetsStr,
			&totalWinsStr,
			&cashbackClaimedStr,
			&activeUsersStr,
			&activeGamesStr,
		)
	} else {
		err = s.clickhouse.QueryRow(ctx, summaryQuery).Scan(
			&totalTransactionsStr,
			&totalDepositsStr,
			&totalWithdrawalsStr,
			&totalBetsStr,
			&totalWinsStr,
			&cashbackClaimedStr,
			&activeUsersStr,
			&activeGamesStr,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			s.logger.Warn("No rows returned for dashboard overview, returning zero values",
				zap.Time("date_from", dateFrom),
				zap.Time("date_to", dateTo))
		} else {
			s.logger.Error("Failed to get dashboard overview summary",
				zap.Error(err),
				zap.Time("date_from", dateFrom),
				zap.Time("date_to", dateTo))
			return nil, fmt.Errorf("failed to get dashboard overview: %w", err)
		}
	}

	summary := dto.DashboardOverviewSummary{
		TotalTransactions: parseUint64Safely(totalTransactionsStr.String),
		TotalDeposits:     parseDecimalSafely(totalDepositsStr.String),
		TotalWithdrawals:  parseDecimalSafely(totalWithdrawalsStr.String),
		TotalBets:         parseDecimalSafely(totalBetsStr.String),
		TotalWins:         parseDecimalSafely(totalWinsStr.String),
		CashbackClaimed:   parseDecimalSafely(cashbackClaimedStr.String),
		ActiveUsers:       parseUint64Safely(activeUsersStr.String),
		ActiveGames:       parseUint64Safely(activeGamesStr.String),
	}

	summary.GGR = summary.TotalBets.Sub(summary.TotalWins)
	summary.NGR = summary.GGR.Sub(summary.CashbackClaimed)

	response := &dto.DashboardOverviewResponse{
		DateRange: dto.DateRange{
			From: &dateFrom,
			To:   &dateTo,
		},
		Summary: summary,
	}

	if includeDailyBreakdown {
		dailyBreakdown, err := s.getDailyBreakdown(ctx, dateFrom, dateTo, userIDs)
		if err != nil {
			s.logger.Warn("Failed to get daily breakdown", zap.Error(err))
		} else {
			response.DailyBreakdown = dailyBreakdown
		}
	}

	return response, nil
}

func (s *AnalyticsStorageImpl) getDailyBreakdown(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID) ([]dto.DashboardDailyBreakdown, error) {
	unionAllQuery, userArgs := s.buildUnionAllQueryString(dateFrom, dateTo, userIDs)

	dailyQuery := `
		SELECT 
			toDate(created_at) as date,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8), toDecimal64(0, 8))) as deposits,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8), toDecimal64(0, 8))) as withdrawals,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as bets,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as wins,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type IN ('groove_bet', 'groove_win') OR session_id IS NOT NULL OR game_id IS NOT NULL)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExact(game_id)), 0)) as active_games
		FROM (
			` + unionAllQuery + `
		) combined_data
		GROUP BY date
		ORDER BY date ASC
	`

	dailyArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			dailyArgs = append(dailyArgs, userArgs...)
		}
	}

	rows, err := s.clickhouse.Query(ctx, dailyQuery, dailyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []dto.DashboardDailyBreakdown
	for rows.Next() {
		var day dto.DashboardDailyBreakdown
		var date time.Time
		var depositsStr, withdrawalsStr, betsStr, winsStr, cashbackClaimedStr sql.NullString
		var activeUsersStr, activeGamesStr sql.NullString

		err := rows.Scan(
			&date,
			&depositsStr,
			&withdrawalsStr,
			&betsStr,
			&winsStr,
			&cashbackClaimedStr,
			&activeUsersStr,
			&activeGamesStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily breakdown row: %w", err)
		}

		deposits := parseDecimalSafely(depositsStr.String)
		withdrawals := parseDecimalSafely(withdrawalsStr.String)
		bets := parseDecimalSafely(betsStr.String)
		wins := parseDecimalSafely(winsStr.String)
		cashbackClaimed := parseDecimalSafely(cashbackClaimedStr.String)
		activeUsers := parseUint64Safely(activeUsersStr.String)
		activeGames := parseUint64Safely(activeGamesStr.String)

		day.Date = date.Format("2006-01-02")
		day.Deposits = deposits
		day.Withdrawals = withdrawals
		day.Bets = bets
		day.Wins = wins
		day.CashbackClaimed = cashbackClaimed
		day.GGR = bets.Sub(wins)
		day.NGR = day.GGR.Sub(cashbackClaimed)
		day.ActiveUsers = activeUsers
		day.ActiveGames = activeGames

		breakdown = append(breakdown, day)
	}

	return breakdown, nil
}

func calculatePredefinedRange(rangeType string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	switch rangeType {
	case "today":
		return today, today.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	case "yesterday":
		return yesterday, yesterday.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	case "last_week":
		start := yesterday.AddDate(0, 0, -6) // 7 days ago to yesterday
		return start, yesterday.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	case "last_30_days":
		start := yesterday.AddDate(0, 0, -29) // 30 days ago to yesterday
		return start, yesterday.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	case "last_90_days":
		start := yesterday.AddDate(0, 0, -89) // 90 days ago to yesterday
		return start, yesterday.Add(24 * time.Hour).Add(-time.Nanosecond), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid range type: %s", rangeType)
	}
}

func (s *AnalyticsStorageImpl) GetPerformanceSummary(ctx context.Context, rangeType string, dateFrom, dateTo *time.Time, userIDs []uuid.UUID) (*dto.PerformanceSummaryResponse, error) {
	var startDate, endDate time.Time
	var err error

	if rangeType != "" && rangeType != "custom" {
		startDate, endDate, err = calculatePredefinedRange(rangeType)
		if err != nil {
			return nil, fmt.Errorf("invalid range type: %w", err)
		}
	} else if dateFrom != nil && dateTo != nil {
		startDate = *dateFrom
		endDate = *dateTo
		rangeType = "custom"
	} else {
		return nil, fmt.Errorf("either range or date_from/date_to must be provided")
	}

	unionAllQuery, userArgs := s.buildUnionAllQueryString(startDate, endDate, userIDs)

	financialQuery := `
		SELECT 
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			) - sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as ggr,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8), toDecimal64(0, 8))) as total_deposits,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8), toDecimal64(0, 8))) as total_withdrawals,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed
		FROM (
			` + unionAllQuery + `
		) combined_data
	`

	financialArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			financialArgs = append(financialArgs, userArgs...)
		}
	}

	var ggrStr, totalDepositsStr, totalWithdrawalsStr, cashbackClaimedStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, financialQuery, financialArgs...).Scan(
		&ggrStr,
		&totalDepositsStr,
		&totalWithdrawalsStr,
		&cashbackClaimedStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			s.logger.Warn("No rows returned for financial overview, returning zero values")
		} else {
			return nil, fmt.Errorf("failed to get financial overview: %w", err)
		}
	}

	var financial dto.PerformanceSummaryFinancialOverview
	financial.GGR = parseDecimalSafely(ggrStr.String)
	financial.TotalDeposits = parseDecimalSafely(totalDepositsStr.String)
	financial.TotalWithdrawals = parseDecimalSafely(totalWithdrawalsStr.String)
	financial.CashbackClaimed = parseDecimalSafely(cashbackClaimedStr.String)

	financial.NGR = financial.GGR.Sub(financial.CashbackClaimed)
	financial.NetDeposits = financial.TotalDeposits.Sub(financial.TotalWithdrawals)

	bettingQuery := `
		SELECT 
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
			toString(ifNull(toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))), 0)) as bet_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))), 0)) as win_count,
			toString(ifNull(toDecimal64(avgIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as avg_bet_amount
		FROM (
			` + unionAllQuery + `
		) combined_data
	`

	bettingArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			bettingArgs = append(bettingArgs, userArgs...)
		}
	}

	var totalBetsStr, totalWinsStr, betCountStr, winCountStr, avgBetAmountStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, bettingQuery, bettingArgs...).Scan(
		&totalBetsStr,
		&totalWinsStr,
		&betCountStr,
		&winCountStr,
		&avgBetAmountStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			s.logger.Warn("No rows returned for betting metrics, returning zero values")
		} else {
			return nil, fmt.Errorf("failed to get betting metrics: %w", err)
		}
	}

	var betting dto.PerformanceSummaryBettingMetrics
	betting.TotalBets = parseDecimalSafely(totalBetsStr.String)
	betting.TotalWins = parseDecimalSafely(totalWinsStr.String)
	betting.BetCount = parseUint64Safely(betCountStr.String)
	betting.WinCount = parseUint64Safely(winCountStr.String)
	betting.AvgBetAmount = parseDecimalSafely(avgBetAmountStr.String)

	// Calculate RTP
	if betting.TotalBets.GreaterThan(decimal.Zero) {
		betting.RTP = betting.TotalWins.Div(betting.TotalBets).Mul(decimal.NewFromInt(100))
	}

	userActivityQuery := `
		SELECT 
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type IN ('groove_bet', 'groove_win') OR session_id IS NOT NULL OR game_id IS NOT NULL)), 0)) as active_users,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type = 'registration')), 0)) as new_users,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type = 'deposit')), 0)) as unique_depositors,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type = 'withdrawal')), 0)) as unique_withdrawers,
			toString(ifNull(toDecimal64(
				CASE 
					WHEN uniqExactIf(user_id, transaction_type = 'deposit') > 0 
					THEN sumIf(amount, transaction_type = 'deposit') / uniqExactIf(user_id, transaction_type = 'deposit')
					ELSE 0
				END, 8
			), toDecimal64(0, 8))) as avg_deposit_per_user,
			toString(ifNull(toDecimal64(
				CASE 
					WHEN uniqExactIf(user_id, transaction_type = 'withdrawal') > 0 
					THEN sumIf(amount, transaction_type = 'withdrawal') / uniqExactIf(user_id, transaction_type = 'withdrawal')
					ELSE 0
				END, 8
			), toDecimal64(0, 8))) as avg_withdrawal_per_user
		FROM (
			` + unionAllQuery + `
		) combined_data
	`

	userActivityArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			userActivityArgs = append(userActivityArgs, userArgs...)
		}
	}

	var activeUsersStr, newUsersStr, uniqueDepositorsStr, uniqueWithdrawersStr, avgDepositPerUserStr, avgWithdrawalPerUserStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, userActivityQuery, userActivityArgs...).Scan(
		&activeUsersStr,
		&newUsersStr,
		&uniqueDepositorsStr,
		&uniqueWithdrawersStr,
		&avgDepositPerUserStr,
		&avgWithdrawalPerUserStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			s.logger.Warn("No rows returned for user activity, returning zero values")
		} else {
			return nil, fmt.Errorf("failed to get user activity: %w", err)
		}
	}

	var userActivity dto.PerformanceSummaryUserActivity
	userActivity.ActiveUsers = parseUint64Safely(activeUsersStr.String)
	userActivity.NewUsers = parseUint64Safely(newUsersStr.String)
	userActivity.UniqueDepositors = parseUint64Safely(uniqueDepositorsStr.String)
	userActivity.UniqueWithdrawers = parseUint64Safely(uniqueWithdrawersStr.String)
	userActivity.AvgDepositPerUser = parseDecimalSafely(avgDepositPerUserStr.String)
	userActivity.AvgWithdrawalPerUser = parseDecimalSafely(avgWithdrawalPerUserStr.String)

	transactionVolumeQuery := `
		SELECT 
			toString(ifNull(toUInt64(count()), 0)) as total_transactions,
			toString(ifNull(toUInt64(countIf(transaction_type = 'deposit')), 0)) as deposit_count,
			toString(ifNull(toUInt64(countIf(transaction_type = 'withdrawal')), 0)) as withdrawal_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))), 0)) as bet_count,
			toString(ifNull(toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))), 0)) as win_count,
			toString(ifNull(toUInt64(countIf(transaction_type = 'cashback_earning')), 0)) as cashback_earned_count,
			toString(ifNull(toUInt64(countIf(transaction_type = 'cashback_claim')), 0)) as cashback_claimed_count
		FROM (
			` + unionAllQuery + `
		) combined_data
	`

	transactionVolumeArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			transactionVolumeArgs = append(transactionVolumeArgs, userArgs...)
		}
	}

	var totalTransactionsStr, depositCountStr, withdrawalCountStr, transactionBetCountStr, transactionWinCountStr, cashbackEarnedCountStr, cashbackClaimedCountStr sql.NullString
	err = s.clickhouse.QueryRow(ctx, transactionVolumeQuery, transactionVolumeArgs...).Scan(
		&totalTransactionsStr,
		&depositCountStr,
		&withdrawalCountStr,
		&transactionBetCountStr,
		&transactionWinCountStr,
		&cashbackEarnedCountStr,
		&cashbackClaimedCountStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			s.logger.Warn("No rows returned for transaction volume, returning zero values")
		} else {
			return nil, fmt.Errorf("failed to get transaction volume: %w", err)
		}
	}

	var transactionVolume dto.PerformanceSummaryTransactionVolume
	transactionVolume.TotalTransactions = parseUint64Safely(totalTransactionsStr.String)
	transactionVolume.DepositCount = parseUint64Safely(depositCountStr.String)
	transactionVolume.WithdrawalCount = parseUint64Safely(withdrawalCountStr.String)
	transactionVolume.BetCount = parseUint64Safely(transactionBetCountStr.String)
	transactionVolume.WinCount = parseUint64Safely(transactionWinCountStr.String)
	transactionVolume.CashbackEarnedCount = parseUint64Safely(cashbackEarnedCountStr.String)
	transactionVolume.CashbackClaimedCount = parseUint64Safely(cashbackClaimedCountStr.String)

	dailyTrends, err := s.getPerformanceDailyTrends(ctx, startDate, endDate, userIDs)
	if err != nil {
		s.logger.Warn("Failed to get daily trends", zap.Error(err))
		dailyTrends = []dto.PerformanceSummaryDailyTrend{}
	}

	response := &dto.PerformanceSummaryResponse{
		RangeType:         rangeType,
		DateRange:         dto.DateRange{From: &startDate, To: &endDate},
		FinancialOverview: financial,
		BettingMetrics:    betting,
		UserActivity:      userActivity,
		TransactionVolume: transactionVolume,
		DailyTrends:       dailyTrends,
	}

	return response, nil
}

func (s *AnalyticsStorageImpl) getPerformanceDailyTrends(ctx context.Context, dateFrom, dateTo time.Time, userIDs []uuid.UUID) ([]dto.PerformanceSummaryDailyTrend, error) {
	unionAllQuery, userArgs := s.buildUnionAllQueryString(dateFrom, dateTo, userIDs)

	dailyQuery := `
		SELECT 
			toDate(created_at) as date,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			) - sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as ggr,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8), toDecimal64(0, 8))) as cashback_claimed,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8), toDecimal64(0, 8))) as deposits,
			toString(ifNull(toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8), toDecimal64(0, 8))) as withdrawals,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
					ELSE abs(amount)
				END,
				transaction_type IN ('bet', 'groove_bet')
			), 8), toDecimal64(0, 8))) as bets,
			toString(ifNull(toDecimal64(sumIf(
				CASE 
					WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
					ELSE COALESCE(win_amount, amount)
				END,
				transaction_type IN ('win', 'groove_win')
				OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
			), 8), toDecimal64(0, 8))) as wins,
			toString(ifNull(toUInt64(uniqExactIf(user_id, transaction_type IN ('groove_bet', 'groove_win') OR session_id IS NOT NULL OR game_id IS NOT NULL)), 0)) as active_users
		FROM (
			` + unionAllQuery + `
		) combined_data
		GROUP BY date
		ORDER BY date ASC
	`

	dailyArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			dailyArgs = append(dailyArgs, userArgs...)
		}
	}

	rows, err := s.clickhouse.Query(ctx, dailyQuery, dailyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily trends: %w", err)
	}
	defer rows.Close()

	var trends []dto.PerformanceSummaryDailyTrend
	for rows.Next() {
		var trend dto.PerformanceSummaryDailyTrend
		var date time.Time
		var ggrStr, cashbackClaimedStr, depositsStr, withdrawalsStr, betsStr, winsStr, activeUsersStr sql.NullString

		err := rows.Scan(
			&date,
			&ggrStr,
			&cashbackClaimedStr,
			&depositsStr,
			&withdrawalsStr,
			&betsStr,
			&winsStr,
			&activeUsersStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily trend row: %w", err)
		}

		ggr := parseDecimalSafely(ggrStr.String)
		cashbackClaimed := parseDecimalSafely(cashbackClaimedStr.String)
		deposits := parseDecimalSafely(depositsStr.String)
		withdrawals := parseDecimalSafely(withdrawalsStr.String)
		bets := parseDecimalSafely(betsStr.String)
		wins := parseDecimalSafely(winsStr.String)
		activeUsers := parseUint64Safely(activeUsersStr.String)

		trend.Date = date.Format("2006-01-02")
		trend.GGR = ggr
		trend.NGR = ggr.Sub(cashbackClaimed)
		trend.Deposits = deposits
		trend.Withdrawals = withdrawals
		trend.Bets = bets
		trend.Wins = wins
		trend.ActiveUsers = activeUsers

		trends = append(trends, trend)
	}

	return trends, nil
}

func (s *AnalyticsStorageImpl) GetTimeSeriesAnalytics(ctx context.Context, dateFrom, dateTo time.Time, granularity string, userIDs []uuid.UUID, metrics []string) (*dto.TimeSeriesResponse, error) {
	if granularity == "" {
		granularity = "day"
	}
	if granularity != "hour" && granularity != "day" {
		return nil, fmt.Errorf("invalid granularity: %s (must be 'hour' or 'day')", granularity)
	}

	unionAllQuery, userArgs := s.buildUnionAllQueryString(dateFrom, dateTo, userIDs)

	baseArgs := []interface{}{}
	if len(userArgs) > 0 {
		for i := 0; i < 5; i++ {
			baseArgs = append(baseArgs, userArgs...)
		}
	}

	includeAll := len(metrics) == 0
	includeRevenue := includeAll || contains(metrics, "revenue")
	includeUsers := includeAll || contains(metrics, "users")
	includeTransactions := includeAll || contains(metrics, "transactions")
	includeDepositsWithdrawals := includeAll || contains(metrics, "deposits") || contains(metrics, "withdrawals")

	response := &dto.TimeSeriesResponse{
		Granularity: granularity,
		DateRange:   dto.DateRange{From: &dateFrom, To: &dateTo},
	}

	timeGroupExpr := "toDate(created_at)"
	if granularity == "hour" {
		timeGroupExpr = "toStartOfHour(created_at)"
	}

	if includeRevenue {
		revenueQuery := `
			SELECT 
				` + timeGroupExpr + ` as timestamp,
				toDecimal64(sumIf(
					CASE 
						WHEN transaction_type = 'groove_bet' AND bet_amount IS NOT NULL THEN bet_amount
						ELSE abs(amount)
					END,
					transaction_type IN ('bet', 'groove_bet')
				) - sumIf(
					CASE 
						WHEN transaction_type = 'groove_bet' THEN COALESCE(win_amount, 0)
						ELSE COALESCE(win_amount, amount)
					END,
					transaction_type IN ('win', 'groove_win')
					OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0)
				), 8) as ggr,
				toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8) as cashback_claimed
			FROM (
				` + unionAllQuery + `
			) combined_data
			GROUP BY timestamp
			ORDER BY timestamp ASC
		`

		rows, err := s.clickhouse.Query(ctx, revenueQuery, baseArgs...)
		if err != nil {
			s.logger.Warn("Failed to get revenue trend", zap.Error(err))
		} else {
			defer rows.Close()
			var revenueTrend []dto.TimeSeriesRevenueTrend
			for rows.Next() {
				var trend dto.TimeSeriesRevenueTrend
				var ggr, cashbackClaimed decimal.Decimal

				err := rows.Scan(&trend.Timestamp, &ggr, &cashbackClaimed)
				if err != nil {
					s.logger.Warn("Failed to scan revenue trend row", zap.Error(err))
					continue
				}

				trend.GGR = ggr
				trend.NGR = ggr.Sub(cashbackClaimed)
				revenueTrend = append(revenueTrend, trend)
			}
			response.RevenueTrend = revenueTrend
		}
	}

	if includeUsers {
		userQuery := `
			SELECT 
				` + timeGroupExpr + ` as timestamp,
				toUInt64(uniqExactIf(user_id, transaction_type IN ('groove_bet', 'groove_win') OR session_id IS NOT NULL OR game_id IS NOT NULL)) as active_users,
				toUInt64(uniqExactIf(user_id, transaction_type = 'registration')) as new_users,
				toUInt64(uniqExactIf(user_id, transaction_type = 'deposit')) as unique_depositors,
				toUInt64(uniqExactIf(user_id, transaction_type = 'withdrawal')) as unique_withdrawers
			FROM (
				` + unionAllQuery + `
			) combined_data
			GROUP BY timestamp
			ORDER BY timestamp ASC
		`

		rows, err := s.clickhouse.Query(ctx, userQuery, baseArgs...)
		if err != nil {
			s.logger.Warn("Failed to get user activity", zap.Error(err))
		} else {
			defer rows.Close()
			var userActivity []dto.TimeSeriesUserActivity
			for rows.Next() {
				var activity dto.TimeSeriesUserActivity

				err := rows.Scan(
					&activity.Timestamp,
					&activity.ActiveUsers,
					&activity.NewUsers,
					&activity.UniqueDepositors,
					&activity.UniqueWithdrawers,
				)
				if err != nil {
					s.logger.Warn("Failed to scan user activity row", zap.Error(err))
					continue
				}

				userActivity = append(userActivity, activity)
			}
			response.UserActivity = userActivity
		}
	}

	// Transaction volume
	if includeTransactions {
		transactionQuery := `
			SELECT 
				` + timeGroupExpr + ` as timestamp,
				toUInt64(count()) as total_transactions,
				toUInt64(countIf(transaction_type = 'deposit')) as deposits,
				toUInt64(countIf(transaction_type = 'withdrawal')) as withdrawals,
				toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as bets,
				toUInt64(countIf(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))) as wins
			FROM (
				` + unionAllQuery + `
			) combined_data
			GROUP BY timestamp
			ORDER BY timestamp ASC
		`

		rows, err := s.clickhouse.Query(ctx, transactionQuery, baseArgs...)
		if err != nil {
			s.logger.Warn("Failed to get transaction volume", zap.Error(err))
		} else {
			defer rows.Close()
			var transactionVolume []dto.TimeSeriesTransactionVolume
			for rows.Next() {
				var volume dto.TimeSeriesTransactionVolume

				err := rows.Scan(
					&volume.Timestamp,
					&volume.TotalTransactions,
					&volume.Deposits,
					&volume.Withdrawals,
					&volume.Bets,
					&volume.Wins,
				)
				if err != nil {
					s.logger.Warn("Failed to scan transaction volume row", zap.Error(err))
					continue
				}

				transactionVolume = append(transactionVolume, volume)
			}
			response.TransactionVolume = transactionVolume
		}
	}

	// Deposits vs Withdrawals
	if includeDepositsWithdrawals {
		depositsQuery := `
			SELECT 
				` + timeGroupExpr + ` as timestamp,
				toDecimal64(sumIf(amount, transaction_type = 'deposit'), 8) as deposits,
				toDecimal64(sumIf(amount, transaction_type = 'withdrawal'), 8) as withdrawals
			FROM (
				` + unionAllQuery + `
			) combined_data
			GROUP BY timestamp
			ORDER BY timestamp ASC
		`

		rows, err := s.clickhouse.Query(ctx, depositsQuery, baseArgs...)
		if err != nil {
			s.logger.Warn("Failed to get deposits vs withdrawals", zap.Error(err))
		} else {
			defer rows.Close()
			var depositsVsWithdrawals []dto.TimeSeriesDepositsVsWithdrawals
			for rows.Next() {
				var dvw dto.TimeSeriesDepositsVsWithdrawals

				err := rows.Scan(
					&dvw.Timestamp,
					&dvw.Deposits,
					&dvw.Withdrawals,
				)
				if err != nil {
					s.logger.Warn("Failed to scan deposits vs withdrawals row", zap.Error(err))
					continue
				}

				depositsVsWithdrawals = append(depositsVsWithdrawals, dvw)
			}
			response.DepositsVsWithdrawals = depositsVsWithdrawals
		}
	}

	return response, nil
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
