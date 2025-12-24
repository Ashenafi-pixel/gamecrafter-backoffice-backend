package report

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"go.uber.org/zap"
)

// GetCountryMetrics retrieves country-level aggregated metrics
func (r *report) GetCountryMetrics(ctx context.Context, req dto.CountryReportReq, userBrandIDs []uuid.UUID) (dto.CountryReportRes, error) {
	var res dto.CountryReportRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}
	if req.SortBy == nil {
		defaultSort := "ngr"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}
	if req.UserType == nil {
		defaultUserType := "all"
		req.UserType = &defaultUserType
	}

	// Parse date range (required)
	var dateFrom, dateTo *time.Time
	if req.DateFrom != nil && *req.DateFrom != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateFrom)
		if err == nil {
			dateFrom = &parsed
		}
	}
	if req.DateTo != nil && *req.DateTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateTo)
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &endOfDay
		}
	}

	// If no date range provided, default to last 30 days
	if dateFrom == nil && dateTo == nil {
		now := time.Now()
		dateTo = &now
		thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
		dateFrom = &thirtyDaysAgo
	}

	// Build WHERE conditions
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Filter out admin users - only include players
	whereConditions = append(whereConditions, "u.is_admin = false")

	// Brand filter - respect user's brand access
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

	// Currency filter
	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
		argIndex++
	}

	// Country filter (multi-select)
	if len(req.Countries) > 0 {
		countryPlaceholders := []string{}
		for _, country := range req.Countries {
			countryPlaceholders = append(countryPlaceholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, country)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("u.country IN (%s)", strings.Join(countryPlaceholders, ",")))
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build HAVING clause for user type filter
	havingConditions := []string{}
	if req.UserType != nil {
		switch *req.UserType {
		case "depositors":
			havingConditions = append(havingConditions, "total_depositors > 0")
		case "active":
			havingConditions = append(havingConditions, "active_players > 0")
		}
	}

	havingClause := ""
	if len(havingConditions) > 0 {
		havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "ngr DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "deposits":
			orderBy = "total_deposits"
		case "ngr":
			orderBy = "ngr"
		case "active_users":
			orderBy = "active_players"
		case "alphabetical":
			orderBy = "country"
		default:
			orderBy = "ngr"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	// Add date params to args - we'll reuse them for different calculations
	// Parameters will be: dateFrom (1), dateTo (2), dateFrom (3), dateTo (4), dateFrom (5), dateTo (6), dateFrom (7)
	dateFromParam1 := argIndex
	dateToParam1 := argIndex + 1
	dateFromParam2 := argIndex + 2
	dateToParam2 := argIndex + 3
	dateFromParam3 := argIndex + 4
	dateToParam3 := argIndex + 5
	dateFromParam4 := argIndex + 6

	if dateFrom != nil {
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		args = append(args, *dateTo)
		argIndex++
	}
	if dateFrom != nil {
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		args = append(args, *dateTo)
		argIndex++
	}
	if dateFrom != nil {
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		args = append(args, *dateTo)
		argIndex++
	}
	if dateFrom != nil {
		args = append(args, *dateFrom)
		argIndex++
	}

	// Build transaction date filter for JOIN
	transactionDateFilter := ""
	if dateFrom != nil && dateTo != nil {
		transactionDateFilter = fmt.Sprintf("AND t.created_at >= $%d AND t.created_at <= $%d", dateFromParam1, dateToParam1)
	} else if dateFrom != nil {
		transactionDateFilter = fmt.Sprintf("AND t.created_at >= $%d", dateFromParam1)
	} else if dateTo != nil {
		transactionDateFilter = fmt.Sprintf("AND t.created_at <= $%d", dateToParam1)
	}

	query := fmt.Sprintf(`
		WITH country_metrics AS (
			SELECT 
				COALESCE(u.country, 'Unknown') as country,
				COUNT(DISTINCT u.id) as total_registrations,
				COUNT(DISTINCT CASE 
					WHEN COALESCE((SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id), u.created_at) >= $%d 
					AND COALESCE((SELECT MAX(t3.created_at) FROM transactions t3 WHERE t3.user_id = u.id), u.created_at) <= $%d 
					THEN u.id 
					ELSE NULL 
				END) as active_players,
				COUNT(DISTINCT CASE 
					WHEN EXISTS (
						SELECT 1 FROM transactions t 
						WHERE t.user_id = u.id 
						AND t.transaction_type = 'deposit' 
						AND t.created_at >= $%d 
						AND t.created_at <= $%d
					) 
					AND NOT EXISTS (
						SELECT 1 FROM transactions t2 
						WHERE t2.user_id = u.id 
						AND t2.transaction_type = 'deposit' 
						AND t2.created_at < $%d
					)
					THEN u.id 
					ELSE NULL 
				END) as first_time_depositors,
				COUNT(DISTINCT CASE 
					WHEN EXISTS (
						SELECT 1 FROM transactions t 
						WHERE t.user_id = u.id 
						AND t.transaction_type = 'deposit'
					) 
					THEN u.id 
					ELSE NULL 
				END) as total_depositors,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'withdrawal' THEN t.amount ELSE 0 END), 0) as total_withdrawals,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE 0 END), 0) as total_wagered,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('win', 'groove_win') THEN t.amount ELSE 0 END), 0) as total_won,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'rakeback_earned' THEN t.amount ELSE 0 END), 0) as rakeback_earned,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'rakeback_claimed' THEN t.amount ELSE 0 END), 0) as rakeback_converted,
				COUNT(DISTINCT CASE 
					WHEN u.status = 'self_excluded' 
					THEN u.id 
					ELSE NULL 
				END) as self_exclusions
			FROM users u
			LEFT JOIN transactions t ON u.id = t.user_id %s
			%s
			GROUP BY COALESCE(u.country, 'Unknown')
		)
		SELECT 
			country,
			total_registrations,
			active_players,
			first_time_depositors,
			total_depositors,
			total_deposits,
			total_withdrawals,
			total_deposits - total_withdrawals as net_position,
			total_wagered,
			total_won,
			total_wagered - total_won as ggr,
			total_wagered - total_won - rakeback_earned as ngr,
			CASE 
				WHEN total_depositors > 0 THEN total_deposits / total_depositors
				ELSE 0
			END as average_deposit_per_player,
			CASE 
				WHEN total_registrations > 0 THEN total_wagered / total_registrations
				ELSE 0
			END as average_wager_per_player,
			rakeback_earned,
			rakeback_converted,
			self_exclusions
		FROM country_metrics
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, dateFromParam2, dateToParam2, dateFromParam3, dateToParam3, dateFromParam4, transactionDateFilter, whereClause, havingClause, orderBy, argIndex, argIndex+1)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get country metrics", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get country metrics")
	}
	defer rows.Close()

	var metrics []dto.CountryMetric
	for rows.Next() {
		var metric dto.CountryMetric

		err := rows.Scan(
			&metric.Country,
			&metric.TotalRegistrations,
			&metric.ActivePlayers,
			&metric.FirstTimeDepositors,
			&metric.TotalDepositors,
			&metric.TotalDeposits,
			&metric.TotalWithdrawals,
			&metric.NetPosition,
			&metric.TotalWagered,
			&metric.TotalWon,
			&metric.GGR,
			&metric.NGR,
			&metric.AverageDepositPerPlayer,
			&metric.AverageWagerPerPlayer,
			&metric.RakebackEarned,
			&metric.RakebackConverted,
			&metric.SelfExclusions,
		)
		if err != nil {
			r.log.Error("failed to scan country metric row", zap.Error(err))
			continue
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating country metrics rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating country metrics rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH country_metrics AS (
			SELECT 
				COALESCE(u.country, 'Unknown') as country
			FROM users u
			LEFT JOIN transactions t ON u.id = t.user_id
			%s
			GROUP BY COALESCE(u.country, 'Unknown')
			%s
		)
		SELECT COUNT(*) as total
		FROM country_metrics
	`, whereClause, havingClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get country metrics count", zap.Error(err))
		total = int64(len(metrics))
	}

	// Calculate summary from ALL data (not just paginated results)
	// We need to run a separate query to get the total summary
	summaryQuery := fmt.Sprintf(`
		WITH country_metrics AS (
			SELECT 
				COALESCE(u.country, 'Unknown') as country,
				COUNT(DISTINCT u.id) as total_registrations,
				COUNT(DISTINCT CASE 
					WHEN COALESCE((SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id), u.created_at) >= $%d 
					AND COALESCE((SELECT MAX(t3.created_at) FROM transactions t3 WHERE t3.user_id = u.id), u.created_at) <= $%d 
					THEN u.id 
					ELSE NULL 
				END) as active_players,
				COUNT(DISTINCT CASE 
					WHEN EXISTS (
						SELECT 1 FROM transactions t 
						WHERE t.user_id = u.id 
						AND t.transaction_type = 'deposit'
					) 
					THEN u.id 
					ELSE NULL 
				END) as total_depositors,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE 0 END), 0) as total_wagered,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('win', 'groove_win') THEN t.amount ELSE 0 END), 0) as total_won,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'rakeback_earned' THEN t.amount ELSE 0 END), 0) as rakeback_earned
			FROM users u
			LEFT JOIN transactions t ON u.id = t.user_id %s
			%s
			GROUP BY COALESCE(u.country, 'Unknown')
			%s
		)
		SELECT 
			SUM(total_registrations) as total_registrations,
			SUM(active_players) as total_active_users,
			SUM(total_depositors) as total_depositors,
			SUM(total_deposits) as total_deposits,
			SUM(total_wagered - total_won - rakeback_earned) as total_ngr
		FROM country_metrics
	`, dateFromParam2, dateToParam2, transactionDateFilter, whereClause, havingClause)

	var summary dto.CountryReportSummary
	var totalRegistrations, totalActiveUsers, totalDepositors sql.NullInt64
	var totalDeposits, totalNGR sql.NullString

	err = r.db.GetPool().QueryRow(ctx, summaryQuery, args[:len(args)-2]...).Scan(
		&totalRegistrations,
		&totalActiveUsers,
		&totalDepositors,
		&totalDeposits,
		&totalNGR,
	)
	if err != nil {
		r.log.Error("failed to get country metrics summary", zap.Error(err))
		// Fallback to calculating from paginated results
		for _, metric := range metrics {
			summary.TotalDeposits = summary.TotalDeposits.Add(metric.TotalDeposits)
			summary.TotalNGR = summary.TotalNGR.Add(metric.NGR)
			summary.TotalActiveUsers += metric.ActivePlayers
			summary.TotalDepositors += metric.TotalDepositors
			summary.TotalRegistrations += metric.TotalRegistrations
		}
	} else {
		// Use the summary from all data
		if totalRegistrations.Valid {
			summary.TotalRegistrations = int(totalRegistrations.Int64)
		}
		if totalActiveUsers.Valid {
			summary.TotalActiveUsers = int(totalActiveUsers.Int64)
		}
		if totalDepositors.Valid {
			summary.TotalDepositors = int(totalDepositors.Int64)
		}
		if totalDeposits.Valid {
			deposits, err := decimal.NewFromString(totalDeposits.String)
			if err == nil {
				summary.TotalDeposits = deposits
			}
		}
		if totalNGR.Valid {
			ngr, err := decimal.NewFromString(totalNGR.String)
			if err == nil {
				summary.TotalNGR = ngr
			}
		}
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Country metrics retrieved successfully"
	res.Data = metrics
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage
	res.Summary = summary

	return res, nil
}

// GetCountryPlayers retrieves players for a specific country (drill-down)
func (r *report) GetCountryPlayers(ctx context.Context, req dto.CountryPlayersReq, userBrandIDs []uuid.UUID) (dto.CountryPlayersRes, error) {
	var res dto.CountryPlayersRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 50
	}
	if req.SortBy == nil {
		defaultSort := "ngr"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}

	// Parse date ranges
	var dateFrom, dateTo, activityFrom, activityTo *time.Time
	if req.DateFrom != nil && *req.DateFrom != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateFrom)
		if err == nil {
			dateFrom = &parsed
		}
	}
	if req.DateTo != nil && *req.DateTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateTo)
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &endOfDay
		}
	}
	if req.ActivityFrom != nil && *req.ActivityFrom != "" {
		parsed, err := time.Parse("2006-01-02", *req.ActivityFrom)
		if err == nil {
			activityFrom = &parsed
		}
	}
	if req.ActivityTo != nil && *req.ActivityTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.ActivityTo)
		if err == nil {
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			activityTo = &endOfDay
		}
	}

	// Build WHERE conditions
	whereConditions := []string{fmt.Sprintf("u.country = $1")}
	args := []interface{}{req.Country}
	argIndex := 2

	// Filter out admin users - only include players
	whereConditions = append(whereConditions, "u.is_admin = false")

	// Brand filter
	if len(userBrandIDs) > 0 {
		brandPlaceholders := []string{}
		for _, brandID := range userBrandIDs {
			brandPlaceholders = append(brandPlaceholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, brandID)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(u.brand_id IS NULL OR u.brand_id IN (%s))", strings.Join(brandPlaceholders, ",")))
	}

	// Date range for transactions
	if dateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.created_at >= $%d", argIndex))
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("t.created_at <= $%d", argIndex))
		args = append(args, *dateTo)
		argIndex++
	}

	// Activity date range - calculate last_activity from transactions
	if activityFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE((SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id), u.created_at) >= $%d", argIndex))
		args = append(args, *activityFrom)
		argIndex++
	}
	if activityTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE((SELECT MAX(t3.created_at) FROM transactions t3 WHERE t3.user_id = u.id), u.created_at) <= $%d", argIndex))
		args = append(args, *activityTo)
		argIndex++
	}

	// KYC status filter
	if req.KYCStatus != nil && *req.KYCStatus != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.kyc_status = $%d", argIndex))
		args = append(args, *req.KYCStatus)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	// Build HAVING clause for deposit and balance filters
	havingConditions := []string{}
	if req.MinDeposits != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("total_deposits >= $%d::numeric", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinDeposits).String())
		argIndex++
	}
	if req.MaxDeposits != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("total_deposits <= $%d::numeric", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxDeposits).String())
		argIndex++
	}
	if req.MinBalance != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("balance >= $%d::numeric", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinBalance).String())
		argIndex++
	}
	if req.MaxBalance != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("balance <= $%d::numeric", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxBalance).String())
		argIndex++
	}

	havingClause := ""
	if len(havingConditions) > 0 {
		havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "ngr DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "deposits":
			orderBy = "total_deposits"
		case "ngr":
			orderBy = "ngr"
		case "activity":
			orderBy = "last_activity"
		case "registration":
			orderBy = "registration_date"
		default:
			orderBy = "ngr"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	// Main query
	query := fmt.Sprintf(`
		WITH player_metrics AS (
			SELECT 
				u.id as player_id,
				u.username,
				u.email,
				u.country,
				u.created_at as registration_date,
				COALESCE(bal.amount_units, 0) as balance,
				COALESCE(u.default_currency, 'USD') as currency,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'withdrawal' THEN t.amount ELSE 0 END), 0) as total_withdrawals,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE 0 END), 0) as total_wagered,
				COALESCE(SUM(CASE WHEN t.transaction_type IN ('win', 'groove_win') THEN t.amount ELSE 0 END), 0) as total_won,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'rakeback_earned' THEN t.amount ELSE 0 END), 0) as rakeback_earned,
				COALESCE(MAX(t.created_at), u.created_at) as last_activity
			FROM users u
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN transactions t ON u.id = t.user_id
			%s
			GROUP BY u.id, u.username, u.email, u.country, u.created_at, bal.amount_units, u.default_currency
		)
		SELECT 
			player_id,
			username,
			email,
			country,
			total_deposits,
			total_withdrawals,
			total_wagered,
			total_wagered - total_won - rakeback_earned as ngr,
			last_activity,
			registration_date,
			balance,
			currency
		FROM player_metrics
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, havingClause, orderBy, argIndex+1, argIndex+2)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get country players", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get country players")
	}
	defer rows.Close()

	var players []dto.CountryPlayer
	for rows.Next() {
		var player dto.CountryPlayer
		var email sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&player.PlayerID,
			&player.Username,
			&email,
			&player.Country,
			&player.TotalDeposits,
			&player.TotalWithdrawals,
			&player.TotalWagered,
			&player.NGR,
			&lastActivity,
			&player.RegistrationDate,
			&player.Balance,
			&player.Currency,
		)
		if err != nil {
			r.log.Error("failed to scan country player row", zap.Error(err))
			continue
		}

		if email.Valid {
			player.Email = &email.String
		}
		if lastActivity.Valid {
			player.LastActivity = &lastActivity.Time
		}

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating country players rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating country players rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH player_metrics AS (
			SELECT 
				u.id,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(bal.amount_units, 0) as balance
			FROM users u
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN transactions t ON u.id = t.user_id
			%s
			GROUP BY u.id, bal.amount_units
			%s
		)
		SELECT COUNT(*) as total
		FROM player_metrics
	`, whereClause, havingClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get country players count", zap.Error(err))
		total = int64(len(players))
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Country players retrieved successfully"
	res.Data = players
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage

	return res, nil
}

