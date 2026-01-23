package report

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"go.uber.org/zap"
)

func (r *report) sortPlayerMetrics(metrics []dto.PlayerMetric, sortBy *string, sortOrder *string) {
	if len(metrics) == 0 {
		return
	}

	order := "desc"
	if sortOrder != nil {
		order = strings.ToLower(*sortOrder)
	}

	sortField := "deposits"
	if sortBy != nil {
		sortField = strings.ToLower(*sortBy)
	}

	sort.Slice(metrics, func(i, j int) bool {
		var less bool
		switch sortField {
		case "deposits":
			less = metrics[i].TotalDeposits.LessThan(metrics[j].TotalDeposits)
		case "wagering":
			less = metrics[i].TotalWagered.LessThan(metrics[j].TotalWagered)
		case "net_loss":
			less = metrics[i].NetGamingResult.LessThan(metrics[j].NetGamingResult)
		case "activity":
			less = metrics[i].NumberOfBets < metrics[j].NumberOfBets
		case "registration":
			less = metrics[i].RegistrationDate.Before(metrics[j].RegistrationDate)
		default:
			less = metrics[i].TotalDeposits.LessThan(metrics[j].TotalDeposits)
		}
		if order == "asc" {
			return less
		}
		return !less
	})
}

func (r *report) GetPlayerMetrics(ctx context.Context, req dto.PlayerMetricsReportReq, userBrandIDs []uuid.UUID) (dto.PlayerMetricsReportRes, error) {
	var res dto.PlayerMetricsReportRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}
	if req.SortBy == nil {
		defaultSort := "deposits"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}

	// Parse date ranges
	var registrationFrom, registrationTo, lastActiveFrom, lastActiveTo *time.Time
	if req.RegistrationFrom != nil && *req.RegistrationFrom != "" {
		parsed, err := time.Parse("2006-01-02", *req.RegistrationFrom)
		if err == nil {
			registrationFrom = &parsed
		}
	}
	if req.RegistrationTo != nil && *req.RegistrationTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.RegistrationTo)
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			registrationTo = &endOfDay
		}
	}
	if req.LastActiveFrom != nil && *req.LastActiveFrom != "" {
		parsed, err := time.Parse("2006-01-02", *req.LastActiveFrom)
		if err == nil {
			lastActiveFrom = &parsed
		}
	}
	if req.LastActiveTo != nil && *req.LastActiveTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.LastActiveTo)
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			lastActiveTo = &endOfDay
		}
	}

	// Build WHERE conditions
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if len(userBrandIDs) > 0 {
		brandPlaceholders := []string{}
		for _, brandID := range userBrandIDs {
			brandPlaceholders = append(brandPlaceholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, brandID)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(u.brand_id IS NULL OR u.brand_id IN (%s))", strings.Join(brandPlaceholders, ",")))
	}

	// Additional brand filter from request
	if req.BrandID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.brand_id = $%d", argIndex))
		args = append(args, *req.BrandID)
		argIndex++
	}

	// Player search
	if req.PlayerSearch != nil && *req.PlayerSearch != "" {
		searchTerm := "%" + *req.PlayerSearch + "%"
		whereConditions = append(whereConditions, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d OR u.id::text = $%d)", argIndex, argIndex, argIndex))
		args = append(args, searchTerm)
		argIndex++
	}

	// Currency filter
	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
		argIndex++
	}

	// Country filter
	if req.Country != nil && *req.Country != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.country = $%d", argIndex))
		args = append(args, *req.Country)
		argIndex++
	}

	// Registration date range
	if registrationFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.created_at >= $%d", argIndex))
		args = append(args, *registrationFrom)
		argIndex++
	}
	if registrationTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.created_at <= $%d", argIndex))
		args = append(args, *registrationTo)
		argIndex++
	}

	// Last active date range - calculate last_activity from transactions
	if lastActiveFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE((SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id), u.created_at) >= $%d", argIndex))
		args = append(args, *lastActiveFrom)
		argIndex++
	}
	if lastActiveTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE((SELECT MAX(t3.created_at) FROM transactions t3 WHERE t3.user_id = u.id), u.created_at) <= $%d", argIndex))
		args = append(args, *lastActiveTo)
		argIndex++
	}

	if req.IsTestAccount != nil {
		if *req.IsTestAccount {
			whereConditions = append(whereConditions, "u.is_test_account = true")
		} else {
			whereConditions = append(whereConditions, "COALESCE(u.is_test_account, false) = false")
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build HAVING clause for deposit/withdrawal filters (PostgreSQL only)
	havingConditions := []string{}

	// Has deposited filter
	if req.HasDeposited != nil {
		if *req.HasDeposited {
			havingConditions = append(havingConditions, "COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) > 0")
		} else {
			havingConditions = append(havingConditions, "COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) = 0")
		}
	}

	// Has withdrawn filter
	if req.HasWithdrawn != nil {
		if *req.HasWithdrawn {
			havingConditions = append(havingConditions, "COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) > 0")
		} else {
			havingConditions = append(havingConditions, "COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) = 0")
		}
	}

	// Total deposits range (PostgreSQL only)
	if req.MinTotalDeposits != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinTotalDeposits))
		argIndex++
	}
	if req.MaxTotalDeposits != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxTotalDeposits))
		argIndex++
	}

	havingClause := ""
	if len(havingConditions) > 0 {
		havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "total_deposits DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "deposits":
			orderBy = "total_deposits"
		case "wagering":
			orderBy = "total_wagered"
		case "net_loss":
			orderBy = "net_gaming_result" // Negative = loss
		case "activity":
			orderBy = "number_of_bets"
		case "registration":
			orderBy = "registration_date"
		default:
			orderBy = "total_deposits"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	chConn, hasClickHouse := r.getClickHouseConn()
	gamingMetricsMap := make(map[uuid.UUID]struct {
		TotalWagered     decimal.Decimal
		TotalWon         decimal.Decimal
		NumberOfBets     int64
		LastActivity     *time.Time
		NumberOfSessions int64
	})
	cashbackMetricsMap := make(map[uuid.UUID]struct {
		RakebackEarned  decimal.Decimal
		RakebackClaimed decimal.Decimal
	})

	if hasClickHouse {
		chQuery := `
			SELECT 
				user_id,
				toDecimal64(sumIf(CASE 
					WHEN transaction_type IN ('groove_bet', 'groove_win') AND bet_amount IS NOT NULL THEN bet_amount
					WHEN transaction_type IN ('bet', 'groove_bet') THEN abs(amount)
					ELSE 0
				END, transaction_type IN ('bet', 'groove_bet')), 8) as total_wagered,
				toDecimal64(sumIf(CASE 
					WHEN transaction_type IN ('groove_bet', 'groove_win') AND win_amount IS NOT NULL THEN win_amount
					WHEN transaction_type IN ('win', 'groove_win') THEN abs(amount)
					ELSE 0
				END, transaction_type IN ('win', 'groove_win')), 8) as total_won,
				toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as number_of_bets,
				max(created_at) as last_activity,
				toUInt64(uniqExact(session_id)) as number_of_sessions
			FROM tucanbit_analytics.transactions
			WHERE transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
				AND (
					(transaction_type IN ('groove_bet', 'groove_win') AND (status = 'completed' OR (status = 'pending' AND bet_amount IS NOT NULL AND win_amount IS NOT NULL)))
					OR (transaction_type NOT IN ('groove_bet', 'groove_win') AND (status = 'completed' OR status IS NULL))
				)
			GROUP BY user_id
		`

		rows, err := chConn.Query(ctx, chQuery)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userIDStr string
				var totalWageredStr, totalWonStr string
				var numberOfBets uint64
				var lastActivity time.Time
				var numberOfSessions uint64

				if err := rows.Scan(&userIDStr, &totalWageredStr, &totalWonStr, &numberOfBets, &lastActivity, &numberOfSessions); err == nil {
					userID, err := uuid.Parse(userIDStr)
					if err == nil {
						totalWagered, _ := decimal.NewFromString(totalWageredStr)
						totalWon, _ := decimal.NewFromString(totalWonStr)
						gamingMetricsMap[userID] = struct {
							TotalWagered     decimal.Decimal
							TotalWon         decimal.Decimal
							NumberOfBets     int64
							LastActivity     *time.Time
							NumberOfSessions int64
						}{
							TotalWagered:     totalWagered,
							TotalWon:         totalWon,
							NumberOfBets:     int64(numberOfBets),
							LastActivity:     &lastActivity,
							NumberOfSessions: int64(numberOfSessions),
						}
					}
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for gaming metrics", zap.Error(err))
		}

		cashbackQuery := `
			SELECT 
				user_id,
				toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8) as rakeback_earned,
				toDecimal64(sumIf(amount, transaction_type = 'cashback_claim'), 8) as rakeback_claimed
			FROM tucanbit_analytics.cashback_analytics
			WHERE transaction_type IN ('cashback_earning', 'cashback_claim')
			GROUP BY user_id
		`

		cashbackRows, err := chConn.Query(ctx, cashbackQuery)
		if err == nil {
			defer cashbackRows.Close()
			for cashbackRows.Next() {
				var userIDStr string
				var rakebackEarnedStr, rakebackClaimedStr string

				if err := cashbackRows.Scan(&userIDStr, &rakebackEarnedStr, &rakebackClaimedStr); err == nil {
					userID, err := uuid.Parse(userIDStr)
					if err == nil {
						rakebackEarned, _ := decimal.NewFromString(rakebackEarnedStr)
						rakebackClaimed, _ := decimal.NewFromString(rakebackClaimedStr)
						cashbackMetricsMap[userID] = struct {
							RakebackEarned  decimal.Decimal
							RakebackClaimed decimal.Decimal
						}{
							RakebackEarned:  rakebackEarned,
							RakebackClaimed: rakebackClaimed,
						}
					}
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for cashback metrics", zap.Error(err))
		}
	}

	query := fmt.Sprintf(`
		WITH all_bet_transactions AS (
			-- All gaming betting data comes from ClickHouse, not PostgreSQL
			-- This CTE is kept for structure but returns no rows since all data is in ClickHouse
			SELECT 
				NULL::uuid as user_id,
				NULL::timestamp as transaction_date,
				NULL::text as transaction_type,
				NULL::decimal as amount
			WHERE false
		),
		player_transactions AS (
			SELECT 
				u.id as user_id,
				u.username,
				u.email,
				u.brand_id,
				b.name as brand_name,
				u.country,
				u.created_at as registration_date,
				u.status as account_status,
				u.kyc_status,
				COALESCE(u.default_currency, 'USD') as currency,
				COALESCE(bal.amount_units, 0) as main_balance,
				t.transaction_type,
				CASE 
					WHEN t.transaction_type = 'deposit' THEN COALESCE(t.usd_amount_cents, 0) / 100.0
					WHEN t.transaction_type = 'withdrawal' THEN COALESCE(t.usd_amount_cents, 0) / 100.0
					ELSE t.amount
				END as amount,
				t.created_at as transaction_date
			FROM users u
			LEFT JOIN brands b ON u.brand_id = b.id
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN transactions t ON u.id = t.user_id
				AND t.transaction_type IN ('deposit', 'withdrawal')
				AND t.status = 'verified'
			%s
		),
		game_sessions_activity AS (
			SELECT 
				gs.user_id,
				MAX(gs.last_activity) as last_activity
			FROM game_sessions gs
			GROUP BY gs.user_id
		),
		player_metrics AS (
			SELECT 
				u.id as user_id,
				u.username,
				u.email,
				u.brand_id,
				b.name as brand_name,
				u.country,
				u.created_at as registration_date,
				u.status as account_status,
				u.kyc_status,
				COALESCE(u.default_currency, 'USD') as currency,
				COALESCE(bal.amount_units, 0) as main_balance,
				COALESCE(
					gsa.last_activity,
					MAX(abt.transaction_date),
					MAX(pt.transaction_date),
					u.created_at
				) as last_activity,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) as total_withdrawals,
				COALESCE(SUM(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN abt.amount ELSE 0 END), 0) as non_gaming_wagered,
				COALESCE(SUM(CASE WHEN abt.transaction_type = 'win' THEN abt.amount ELSE 0 END), 0) as non_gaming_won,
				0::numeric as rakeback_earned,
				0::numeric as rakeback_claimed,
				MIN(CASE WHEN pt.transaction_type = 'deposit' THEN pt.transaction_date ELSE NULL END) as first_deposit_date,
				MAX(CASE WHEN pt.transaction_type = 'deposit' THEN pt.transaction_date ELSE NULL END) as last_deposit_date,
				COUNT(DISTINCT COALESCE(abt.transaction_date, pt.transaction_date)::date) as non_gaming_sessions,
				COUNT(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN 1 ELSE NULL END) as non_gaming_bets
			FROM users u
			LEFT JOIN brands b ON u.brand_id = b.id
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN game_sessions_activity gsa ON u.id = gsa.user_id
			LEFT JOIN player_transactions pt ON u.id = pt.user_id
			LEFT JOIN all_bet_transactions abt ON u.id = abt.user_id
			%s
			GROUP BY u.id, u.username, u.email, u.brand_id, b.name, u.country, u.created_at, u.status, u.kyc_status, u.default_currency, bal.amount_units, gsa.last_activity
			%s
		)
		SELECT 
			pm.user_id as player_id,
			pm.username,
			pm.email,
			pm.brand_id,
			pm.brand_name,
			pm.country,
			pm.registration_date,
			pm.last_activity,
			pm.main_balance,
			pm.currency,
			pm.total_deposits,
			pm.total_withdrawals,
			pm.total_deposits - pm.total_withdrawals as net_deposits,
			pm.non_gaming_wagered as total_wagered,
			pm.non_gaming_won as total_won,
			pm.rakeback_earned,
			pm.rakeback_claimed,
			pm.non_gaming_wagered - pm.non_gaming_won - pm.rakeback_earned as net_gaming_result,
			pm.non_gaming_sessions as number_of_sessions,
			pm.non_gaming_bets as number_of_bets,
			pm.account_status,
			NULL::text as device_type,
			pm.kyc_status,
			pm.first_deposit_date,
			pm.last_deposit_date
		FROM player_metrics pm
		`, whereClause, whereClause, havingClause)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get player metrics", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get player metrics")
	}
	defer rows.Close()

	var metrics []dto.PlayerMetric
	for rows.Next() {
		var metric dto.PlayerMetric
		var email sql.NullString
		var brandID uuid.NullUUID
		var brandName sql.NullString
		var country sql.NullString
		var lastActivity sql.NullTime
		var kycStatus sql.NullString
		var firstDepositDate sql.NullTime
		var lastDepositDate sql.NullTime
		var deviceType sql.NullString

		err := rows.Scan(
			&metric.PlayerID,
			&metric.Username,
			&email,
			&brandID,
			&brandName,
			&country,
			&metric.RegistrationDate,
			&lastActivity,
			&metric.MainBalance,
			&metric.Currency,
			&metric.TotalDeposits,
			&metric.TotalWithdrawals,
			&metric.NetDeposits,
			&metric.TotalWagered,
			&metric.TotalWon,
			&metric.RakebackEarned,
			&metric.RakebackClaimed,
			&metric.NetGamingResult,
			&metric.NumberOfSessions,
			&metric.NumberOfBets,
			&metric.AccountStatus,
			&deviceType,
			&kycStatus,
			&firstDepositDate,
			&lastDepositDate,
		)
		if err != nil {
			r.log.Error("failed to scan player metric row", zap.Error(err))
			continue
		}

		if email.Valid {
			metric.Email = &email.String
		}
		if brandID.Valid {
			metric.BrandID = &brandID.UUID
		}
		if brandName.Valid {
			metric.BrandName = &brandName.String
		}
		if country.Valid {
			metric.Country = &country.String
		}
		if lastActivity.Valid {
			metric.LastActivity = &lastActivity.Time
		}
		if kycStatus.Valid {
			metric.KYCStatus = &kycStatus.String
		}
		if firstDepositDate.Valid {
			metric.FirstDepositDate = &firstDepositDate.Time
		}
		if lastDepositDate.Valid {
			metric.LastDepositDate = &lastDepositDate.Time
		}
		if deviceType.Valid {
			metric.DeviceType = &deviceType.String
		}

		if gamingMetrics, exists := gamingMetricsMap[metric.PlayerID]; exists {
			metric.TotalWagered = metric.TotalWagered.Add(gamingMetrics.TotalWagered)
			metric.TotalWon = metric.TotalWon.Add(gamingMetrics.TotalWon)
			metric.NumberOfBets = metric.NumberOfBets + gamingMetrics.NumberOfBets
			metric.NumberOfSessions = metric.NumberOfSessions + gamingMetrics.NumberOfSessions
			if gamingMetrics.LastActivity != nil {
				if metric.LastActivity == nil || gamingMetrics.LastActivity.After(*metric.LastActivity) {
					metric.LastActivity = gamingMetrics.LastActivity
				}
			}
		}

		if cashbackMetrics, exists := cashbackMetricsMap[metric.PlayerID]; exists {
			metric.RakebackEarned = metric.RakebackEarned.Add(cashbackMetrics.RakebackEarned)
			metric.RakebackClaimed = metric.RakebackClaimed.Add(cashbackMetrics.RakebackClaimed)
		}

		metric.NetGamingResult = metric.TotalWagered.Sub(metric.TotalWon).Sub(metric.RakebackEarned)

		if req.MinTotalWagers != nil {
			minWagers := decimal.NewFromFloat(*req.MinTotalWagers)
			if metric.TotalWagered.LessThan(minWagers) {
				continue
			}
		}
		if req.MaxTotalWagers != nil {
			maxWagers := decimal.NewFromFloat(*req.MaxTotalWagers)
			if metric.TotalWagered.GreaterThan(maxWagers) {
				continue
			}
		}

		// Net result filter (deposits - withdrawals - net gaming result)
		netResult := metric.TotalDeposits.Sub(metric.TotalWithdrawals).Sub(metric.NetGamingResult)
		if req.MinNetResult != nil {
			minNetResult := decimal.NewFromFloat(*req.MinNetResult)
			if netResult.LessThan(minNetResult) {
				continue
			}
		}
		if req.MaxNetResult != nil {
			maxNetResult := decimal.NewFromFloat(*req.MaxNetResult)
			if netResult.GreaterThan(maxNetResult) {
				continue
			}
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating player metrics rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating player metrics rows")
	}

	// Store total before pagination
	total := int64(len(metrics))

	// Apply sorting (since we filtered after merge, we need to sort here)
	r.sortPlayerMetrics(metrics, req.SortBy, req.SortOrder)

	// Apply pagination
	startIdx := (req.Page - 1) * req.PerPage
	endIdx := startIdx + req.PerPage
	if startIdx > len(metrics) {
		metrics = []dto.PlayerMetric{}
	} else if endIdx > len(metrics) {
		metrics = metrics[startIdx:]
	} else {
		metrics = metrics[startIdx:endIdx]
	}

	// Calculate summary
	var summary dto.PlayerMetricsSummary
	summary.PlayerCount = total
	for _, metric := range metrics {
		summary.TotalDeposits = summary.TotalDeposits.Add(metric.TotalDeposits)
		summary.TotalWithdrawals = summary.TotalWithdrawals.Add(metric.TotalWithdrawals)
		summary.TotalWagers = summary.TotalWagers.Add(metric.TotalWagered)
		summary.TotalNGR = summary.TotalNGR.Add(metric.NetGamingResult)
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Player metrics retrieved successfully"
	res.Data = metrics
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage
	res.Summary = summary

	return res, nil
}

// GetPlayerTransactions retrieves detailed transactions for a specific player (drill-down)
func (r *report) GetPlayerTransactions(ctx context.Context, req dto.PlayerTransactionsReq) (dto.PlayerTransactionsRes, error) {
	var res dto.PlayerTransactionsRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 50
	}
	if req.SortBy == nil {
		defaultSort := "date"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}

	// Parse date range
	var dateFrom, dateTo *time.Time
	if req.DateFrom != nil && *req.DateFrom != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05", *req.DateFrom)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", *req.DateFrom)
		}
		if err == nil {
			dateFrom = &parsed
		}
	}
	if req.DateTo != nil && *req.DateTo != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05", *req.DateTo)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", *req.DateTo)
		}
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &endOfDay
		}
	}

	// Build WHERE conditions
	whereConditions := []string{}
	args := []interface{}{req.PlayerID}
	argIndex := 2

	// Date range
	if dateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("transaction_date >= $%d", argIndex))
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("transaction_date <= $%d", argIndex))
		args = append(args, *dateTo)
		argIndex++
	}

	// Transaction type filter
	if req.TransactionType != nil && *req.TransactionType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("transaction_type = $%d", argIndex))
		args = append(args, *req.TransactionType)
		argIndex++
	}

	// Game provider filter
	if req.GameProvider != nil && *req.GameProvider != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_provider = $%d", argIndex))
		args = append(args, *req.GameProvider)
		argIndex++
	}

	// Game ID filter
	if req.GameID != nil && *req.GameID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_id = $%d", argIndex))
		args = append(args, *req.GameID)
		argIndex++
	}

	// Amount range
	if req.MinAmount != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("amount >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinAmount))
		argIndex++
	}
	if req.MaxAmount != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("amount <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxAmount))
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "AND " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "transaction_date DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "date":
			orderBy = "transaction_date"
		case "amount":
			orderBy = "amount"
		case "net":
			orderBy = "net"
		case "game":
			orderBy = "game_name"
		default:
			orderBy = "transaction_date"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	var allTransactions []dto.PlayerTransactionDetail
	chConn, hasClickHouse := r.getClickHouseConn()

	if hasClickHouse {
		userIDStr := req.PlayerID.String()
		chWhereConditions := []string{"user_id = ?"}
		chArgs := []interface{}{userIDStr}

		if dateFrom != nil {
			chWhereConditions = append(chWhereConditions, "created_at >= ?")
			chArgs = append(chArgs, dateFrom.Format("2006-01-02 15:04:05"))
		}
		if dateTo != nil {
			chWhereConditions = append(chWhereConditions, "created_at <= ?")
			chArgs = append(chArgs, dateTo.Format("2006-01-02 15:04:05"))
		}

		if req.TransactionType != nil && *req.TransactionType != "" {
			if *req.TransactionType == "bet" {
				chWhereConditions = append(chWhereConditions, "transaction_type IN ('bet', 'groove_bet')")
			} else if *req.TransactionType == "win" {
				chWhereConditions = append(chWhereConditions, "(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))")
			}
		} else {
			chWhereConditions = append(chWhereConditions, "transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')")
		}

		if req.GameProvider != nil && *req.GameProvider != "" {
			chWhereConditions = append(chWhereConditions, "provider = ?")
			chArgs = append(chArgs, *req.GameProvider)
		}

		if req.GameID != nil && *req.GameID != "" {
			chWhereConditions = append(chWhereConditions, "game_id = ?")
			chArgs = append(chArgs, *req.GameID)
		}

		chWhereClause := strings.Join(chWhereConditions, " AND ")
		chQuery := fmt.Sprintf(`
			SELECT 
				id,
				toString(id) as transaction_id,
				transaction_type,
				created_at,
				CASE 
					WHEN transaction_type IN ('groove_bet', 'groove_win') AND bet_amount IS NOT NULL THEN bet_amount
					WHEN transaction_type IN ('bet', 'groove_bet') THEN abs(amount)
					WHEN transaction_type IN ('win', 'groove_win') THEN COALESCE(win_amount, abs(amount))
					ELSE abs(amount)
				END as amount,
				currency,
				status,
				provider as game_provider,
				game_id,
				game_name,
				toString(id) as bet_id,
				round_id,
				CASE 
					WHEN transaction_type IN ('groove_bet', 'groove_win') AND bet_amount IS NOT NULL THEN bet_amount
					WHEN transaction_type IN ('bet', 'groove_bet') THEN abs(amount)
					ELSE NULL
				END as bet_amount,
				CASE 
					WHEN transaction_type IN ('groove_bet', 'groove_win') THEN COALESCE(win_amount, 0)
					WHEN transaction_type IN ('win', 'groove_win') THEN COALESCE(win_amount, abs(amount))
					ELSE NULL
				END as win_amount,
				NULL as rakeback_earned,
				NULL as rakeback_claimed,
				NULL as rtp,
				CASE 
					WHEN (transaction_type IN ('groove_bet', 'groove_win') AND bet_amount IS NOT NULL AND bet_amount > 0) THEN COALESCE(win_amount, 0) / bet_amount
					WHEN (transaction_type IN ('bet', 'groove_bet') AND abs(amount) > 0) THEN COALESCE(win_amount, 0) / abs(amount)
					ELSE NULL
				END as multiplier,
				NULL as ggr,
				NULL as net,
				'cash' as bet_type,
				NULL as payment_method,
				NULL as tx_hash,
				NULL as network,
				NULL as chain_id,
				NULL as fees,
				NULL as device,
				NULL as ip_address,
				session_id
			FROM tucanbit_analytics.transactions
			WHERE %s
				AND (
					(transaction_type IN ('groove_bet', 'groove_win') AND (status = 'completed' OR (status = 'pending' AND bet_amount IS NOT NULL AND win_amount IS NOT NULL)))
					OR (transaction_type NOT IN ('groove_bet', 'groove_win') AND (status = 'completed' OR status IS NULL))
				)
			ORDER BY created_at DESC
		`, chWhereClause)

		chRows, err := chConn.Query(ctx, chQuery, chArgs...)
		if err == nil {
			defer chRows.Close()
			for chRows.Next() {
				var tx dto.PlayerTransactionDetail
				var idStr, transactionIDStr, transactionTypeStr, amountStr, currencyStr, statusStr string
				var gameProviderStr, gameIDStr, gameNameStr, betIDStr, roundIDStr sql.NullString
				var betAmountStr, winAmountStr, multiplierStr sql.NullString
				var sessionIDStr sql.NullString
				var createdAt time.Time

				err := chRows.Scan(
					&idStr, &transactionIDStr, &transactionTypeStr, &createdAt,
					&amountStr, &currencyStr, &statusStr,
					&gameProviderStr, &gameIDStr, &gameNameStr, &betIDStr, &roundIDStr,
					&betAmountStr, &winAmountStr,
					&multiplierStr,
					&sessionIDStr,
				)
				if err != nil {
					r.log.Warn("Failed to scan ClickHouse transaction row", zap.Error(err))
					continue
				}

				txID, err := uuid.Parse(idStr)
				if err != nil {
					continue
				}
				tx.ID = txID
				tx.TransactionID = transactionIDStr
				tx.Type = transactionTypeStr
				tx.DateTime = createdAt
				tx.Currency = currencyStr
				tx.Status = statusStr

				amount, err := decimal.NewFromString(amountStr)
				if err == nil {
					tx.Amount = amount
				}

				if gameProviderStr.Valid {
					tx.GameProvider = &gameProviderStr.String
				}
				if gameIDStr.Valid {
					tx.GameID = &gameIDStr.String
				}
				if gameNameStr.Valid {
					tx.GameName = &gameNameStr.String
				}
				if betIDStr.Valid {
					tx.BetID = &betIDStr.String
				}
				if roundIDStr.Valid {
					tx.RoundID = &roundIDStr.String
				}
				if betAmountStr.Valid {
					betAmt, err := decimal.NewFromString(betAmountStr.String)
					if err == nil {
						tx.BetAmount = &betAmt
					}
				}
				if winAmountStr.Valid {
					winAmt, err := decimal.NewFromString(winAmountStr.String)
					if err == nil {
						tx.WinAmount = &winAmt
					}
				}
				if multiplierStr.Valid && multiplierStr.String != "0" {
					mult, err := decimal.NewFromString(multiplierStr.String)
					if err == nil && !mult.IsZero() {
						tx.Multiplier = &mult
					}
				}
				if sessionIDStr.Valid {
					tx.SessionID = &sessionIDStr.String
				}

				betType := "cash"
				tx.BetType = &betType

				allTransactions = append(allTransactions, tx)
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for player transactions", zap.Error(err))
		}
	}

	query := fmt.Sprintf(`
		WITH all_transactions AS (
			SELECT 
				b.id,
				b.client_transaction_id as transaction_id,
				CASE WHEN COALESCE(b.payout, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				COALESCE(b.timestamp, NOW()) as transaction_date,
				COALESCE(b.payout, b.amount) as amount,
				b.currency,
				COALESCE(b.status, 'completed') as status,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				b.id::text as bet_id,
				b.round_id::text as round_id,
				b.amount as bet_amount,
				COALESCE(b.payout, 0) as win_amount,
				NULL::decimal as rakeback_earned,
				NULL::decimal as rakeback_claimed,
				NULL::decimal as rtp,
				CASE WHEN b.amount > 0 THEN COALESCE(b.payout, 0) / b.amount ELSE NULL END as multiplier,
				b.amount - COALESCE(b.payout, 0) as ggr,
				COALESCE(b.payout, 0) - b.amount as net,
				'cash' as bet_type,
				NULL::text as payment_method,
				NULL::text as tx_hash,
				NULL::text as network,
				NULL::text as chain_id,
				NULL::decimal as fees,
				NULL::text as device,
				NULL::text as ip_address,
				NULL::text as session_id
			FROM bets b
			WHERE b.user_id = $1
				AND (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				%s

			UNION ALL

			SELECT 
				w.id,
				w.withdrawal_id as transaction_id,
				'withdrawal' as transaction_type,
				w.created_at as transaction_date,
				(w.usd_amount_cents::decimal / 100) as amount,
				w.currency_code as currency,
				w.status::text,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				NULL::text as bet_id,
				NULL::text as round_id,
				NULL::decimal as bet_amount,
				NULL::decimal as win_amount,
				NULL::decimal as rakeback_earned,
				NULL::decimal as rakeback_claimed,
				NULL::decimal as rtp,
				NULL::decimal as multiplier,
				NULL::decimal as ggr,
				NULL::decimal as net,
				NULL::text as bet_type,
				w.protocol as payment_method,
				w.tx_hash,
				NULL::text as network,
				w.chain_id,
				(w.fee_cents::decimal / 100) as fees,
				NULL::text as device,
				NULL::text as ip_address,
				NULL::text as session_id
			FROM withdrawals w
			WHERE w.user_id = $1
				%s
		)
		SELECT 
			id,
			transaction_id,
			transaction_type,
			transaction_date,
			amount,
			currency,
			status,
			game_provider,
			game_id,
			game_name,
			bet_id,
			round_id,
			bet_amount,
			win_amount,
			rakeback_earned,
			rakeback_claimed,
			rtp,
			multiplier,
			ggr,
			net,
			bet_type,
			payment_method,
			tx_hash,
			network,
			chain_id,
			fees,
			device,
			ip_address,
			session_id
		FROM all_transactions
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, whereClause, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get player transactions", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get player transactions")
	}
	defer rows.Close()

	var transactions []dto.PlayerTransactionDetail
	for rows.Next() {
		var tx dto.PlayerTransactionDetail
		var gameProvider sql.NullString
		var gameID sql.NullString
		var gameName sql.NullString
		var betID sql.NullString
		var roundID sql.NullString
		var betAmount sql.NullString
		var winAmount sql.NullString
		var rakebackEarned sql.NullString
		var rakebackClaimed sql.NullString
		var rtp sql.NullString
		var multiplier sql.NullString
		var ggr sql.NullString
		var net sql.NullString
		var betType sql.NullString
		var paymentMethod sql.NullString
		var txHash sql.NullString
		var network sql.NullString
		var chainID sql.NullString
		var fees sql.NullString
		var device sql.NullString
		var ipAddress sql.NullString
		var sessionID sql.NullString

		err := rows.Scan(
			&tx.ID,
			&tx.TransactionID,
			&tx.Type,
			&tx.DateTime,
			&tx.Amount,
			&tx.Currency,
			&tx.Status,
			&gameProvider,
			&gameID,
			&gameName,
			&betID,
			&roundID,
			&betAmount,
			&winAmount,
			&rakebackEarned,
			&rakebackClaimed,
			&rtp,
			&multiplier,
			&ggr,
			&net,
			&betType,
			&paymentMethod,
			&txHash,
			&network,
			&chainID,
			&fees,
			&device,
			&ipAddress,
			&sessionID,
		)
		if err != nil {
			r.log.Warn("Failed to scan PostgreSQL transaction row", zap.Error(err), zap.Int("row_index", len(transactions)))
			continue
		}

		if gameProvider.Valid {
			tx.GameProvider = &gameProvider.String
		}
		if gameID.Valid {
			tx.GameID = &gameID.String
		}
		if gameName.Valid {
			tx.GameName = &gameName.String
		}
		if betID.Valid {
			tx.BetID = &betID.String
		}
		if roundID.Valid {
			tx.RoundID = &roundID.String
		}
		if betAmount.Valid {
			betAmt, err := decimal.NewFromString(betAmount.String)
			if err == nil {
				tx.BetAmount = &betAmt
			}
		}
		if winAmount.Valid {
			winAmt, err := decimal.NewFromString(winAmount.String)
			if err == nil {
				tx.WinAmount = &winAmt
			}
		}
		if rakebackEarned.Valid {
			rbEarned, err := decimal.NewFromString(rakebackEarned.String)
			if err == nil {
				tx.RakebackEarned = &rbEarned
			}
		}
		if rakebackClaimed.Valid {
			rbClaimed, err := decimal.NewFromString(rakebackClaimed.String)
			if err == nil {
				tx.RakebackClaimed = &rbClaimed
			}
		}
		if rtp.Valid {
			rtpVal, err := decimal.NewFromString(rtp.String)
			if err == nil {
				tx.RTP = &rtpVal
			}
		}
		if multiplier.Valid {
			mult, err := decimal.NewFromString(multiplier.String)
			if err == nil {
				tx.Multiplier = &mult
			}
		}
		if ggr.Valid {
			ggrVal, err := decimal.NewFromString(ggr.String)
			if err == nil {
				tx.GGR = &ggrVal
			}
		}
		if net.Valid {
			netVal, err := decimal.NewFromString(net.String)
			if err == nil {
				tx.Net = &netVal
			}
		}
		if betType.Valid {
			tx.BetType = &betType.String
		}
		if paymentMethod.Valid {
			tx.PaymentMethod = &paymentMethod.String
		}
		if txHash.Valid {
			tx.TXHash = &txHash.String
		}
		if network.Valid {
			tx.Network = &network.String
		}
		if chainID.Valid {
			tx.ChainID = &chainID.String
		}
		if fees.Valid {
			feesVal, err := decimal.NewFromString(fees.String)
			if err == nil {
				tx.Fees = &feesVal
			}
		}
		if device.Valid {
			tx.Device = &device.String
		}
		if ipAddress.Valid {
			tx.IPAddress = &ipAddress.String
		}
		if sessionID.Valid {
			tx.SessionID = &sessionID.String
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating player transactions rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating player transactions rows")
	}

	allTransactions = append(allTransactions, transactions...)

	var total int64
	if hasClickHouse {
		userIDStr := req.PlayerID.String()
		chCountWhereConditions := []string{"user_id = ?"}
		chCountArgs := []interface{}{userIDStr}

		if dateFrom != nil {
			chCountWhereConditions = append(chCountWhereConditions, "created_at >= ?")
			chCountArgs = append(chCountArgs, dateFrom.Format("2006-01-02 15:04:05"))
		}
		if dateTo != nil {
			chCountWhereConditions = append(chCountWhereConditions, "created_at <= ?")
			chCountArgs = append(chCountArgs, dateTo.Format("2006-01-02 15:04:05"))
		}

		if req.TransactionType != nil && *req.TransactionType != "" {
			if *req.TransactionType == "bet" {
				chCountWhereConditions = append(chCountWhereConditions, "transaction_type IN ('bet', 'groove_bet')")
			} else if *req.TransactionType == "win" {
				chCountWhereConditions = append(chCountWhereConditions, "(transaction_type IN ('win', 'groove_win') OR (transaction_type = 'groove_bet' AND win_amount IS NOT NULL AND win_amount > 0))")
			}
		} else {
			chCountWhereConditions = append(chCountWhereConditions, "transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')")
		}

		if req.GameProvider != nil && *req.GameProvider != "" {
			chCountWhereConditions = append(chCountWhereConditions, "provider = ?")
			chCountArgs = append(chCountArgs, *req.GameProvider)
		}

		if req.GameID != nil && *req.GameID != "" {
			chCountWhereConditions = append(chCountWhereConditions, "game_id = ?")
			chCountArgs = append(chCountArgs, *req.GameID)
		}

		chCountWhereClause := strings.Join(chCountWhereConditions, " AND ")
		chCountQuery := fmt.Sprintf(`
			SELECT count()
			FROM tucanbit_analytics.transactions
			WHERE %s
				AND (
					(transaction_type IN ('groove_bet', 'groove_win') AND (status = 'completed' OR (status = 'pending' AND bet_amount IS NOT NULL AND win_amount IS NOT NULL)))
					OR (transaction_type NOT IN ('groove_bet', 'groove_win') AND (status = 'completed' OR status IS NULL))
				)
		`, chCountWhereClause)

		var chCount uint64
		err := chConn.QueryRow(ctx, chCountQuery, chCountArgs...).Scan(&chCount)
		if err == nil {
			total += int64(chCount)
		}
	}

	countQuery := fmt.Sprintf(`
		WITH all_transactions AS (
			SELECT COALESCE(b.timestamp, NOW()) as transaction_date, 'bet' as type
			FROM bets b WHERE b.user_id = $1 %s
			UNION ALL
			SELECT w.created_at as transaction_date, 'withdrawal' as type
			FROM withdrawals w WHERE w.user_id = $1 %s
		)
		SELECT COUNT(*) as total
		FROM all_transactions
	`, whereClause, whereClause)

	var pgCount int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&pgCount)
	if err == nil {
		total += pgCount
	} else {
		r.log.Error("failed to get player transactions count", zap.Error(err))
		if total == 0 {
			total = int64(len(allTransactions))
		}
	}

	if req.MinAmount != nil || req.MaxAmount != nil {
		filtered := make([]dto.PlayerTransactionDetail, 0)
		for _, tx := range allTransactions {
			if req.MinAmount != nil && tx.Amount.LessThan(decimal.NewFromFloat(*req.MinAmount)) {
				continue
			}
			if req.MaxAmount != nil && tx.Amount.GreaterThan(decimal.NewFromFloat(*req.MaxAmount)) {
				continue
			}
			filtered = append(filtered, tx)
		}
		allTransactions = filtered
	}

	sort.Slice(allTransactions, func(i, j int) bool {
		if req.SortBy == nil {
			return allTransactions[i].DateTime.After(allTransactions[j].DateTime)
		}
		switch *req.SortBy {
		case "date":
			if req.SortOrder != nil && *req.SortOrder == "asc" {
				return allTransactions[i].DateTime.Before(allTransactions[j].DateTime)
			}
			return allTransactions[i].DateTime.After(allTransactions[j].DateTime)
		case "amount":
			if req.SortOrder != nil && *req.SortOrder == "asc" {
				return allTransactions[i].Amount.LessThan(allTransactions[j].Amount)
			}
			return allTransactions[i].Amount.GreaterThan(allTransactions[j].Amount)
		default:
			return allTransactions[i].DateTime.After(allTransactions[j].DateTime)
		}
	})

	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	startIdx := (req.Page - 1) * req.PerPage
	endIdx := startIdx + req.PerPage
	if startIdx > len(allTransactions) {
		startIdx = len(allTransactions)
	}
	if endIdx > len(allTransactions) {
		endIdx = len(allTransactions)
	}

	var paginatedTransactions []dto.PlayerTransactionDetail
	if startIdx < len(allTransactions) {
		paginatedTransactions = allTransactions[startIdx:endIdx]
	}

	res.Message = "Player transactions retrieved successfully"
	res.Data = paginatedTransactions
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage

	return res, nil
}
