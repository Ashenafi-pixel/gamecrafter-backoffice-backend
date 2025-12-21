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

// GetPlayerMetrics retrieves player metrics report with filters
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
			// Set to end of day
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

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build HAVING clause for aggregated filters
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

	// Total deposits range
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

	// Total wagers range
	if req.MinTotalWagers != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("COALESCE(SUM(CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE 0 END), 0) >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinTotalWagers))
		argIndex++
	}
	if req.MaxTotalWagers != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("COALESCE(SUM(CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE 0 END), 0) <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxTotalWagers))
		argIndex++
	}

	// Net result range (deposits - withdrawals - net gaming result)
	if req.MinNetResult != nil {
		havingConditions = append(havingConditions, fmt.Sprintf(
			"(COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) - "+
				"COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) - "+
				"(COALESCE(SUM(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN abt.amount ELSE 0 END), 0) - "+
				"COALESCE(SUM(CASE WHEN abt.transaction_type = 'win' THEN abt.amount ELSE 0 END), 0))) >= $%d",
			argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinNetResult))
		argIndex++
	}
	if req.MaxNetResult != nil {
		havingConditions = append(havingConditions, fmt.Sprintf(
			"(COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) - "+
				"COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) - "+
				"(COALESCE(SUM(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN abt.amount ELSE 0 END), 0) - "+
				"COALESCE(SUM(CASE WHEN abt.transaction_type = 'win' THEN abt.amount ELSE 0 END), 0))) <= $%d",
			argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxNetResult))
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

	// Main query - aggregates player metrics from all transaction sources
	query := fmt.Sprintf(`
		WITH all_bet_transactions AS (
			-- GrooveTech transactions (wagers and results)
			SELECT 
				ga.user_id,
				gt.created_at as transaction_date,
				'wager' as transaction_type,
				ABS(gt.amount) as amount
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE gt.type = 'wager' AND gt.amount < 0

			UNION ALL

			SELECT 
				ga.user_id,
				gt.created_at as transaction_date,
				'win' as transaction_type,
				gt.amount as amount
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE gt.type = 'result' AND gt.amount > 0

			UNION ALL

			-- General bets (crash game)
			SELECT 
				b.user_id,
				COALESCE(b.timestamp, NOW()) as transaction_date,
				'bet' as transaction_type,
				b.amount as amount
			FROM bets b
			WHERE b.amount > 0

			UNION ALL

			SELECT 
				b.user_id,
				COALESCE(b.timestamp, NOW()) as transaction_date,
				'win' as transaction_type,
				COALESCE(b.payout, 0) as amount
			FROM bets b
			WHERE COALESCE(b.payout, 0) > 0

			UNION ALL

			-- Sport bets
			SELECT 
				sb.user_id,
				sb.created_at as transaction_date,
				'bet' as transaction_type,
				sb.bet_amount as amount
			FROM sport_bets sb
			WHERE sb.bet_amount > 0

			UNION ALL

			SELECT 
				sb.user_id,
				sb.created_at as transaction_date,
				'win' as transaction_type,
				COALESCE(sb.actual_win, 0) as amount
			FROM sport_bets sb
			WHERE COALESCE(sb.actual_win, 0) > 0

			UNION ALL

			-- Plinko bets
			SELECT 
				p.user_id,
				p.timestamp as transaction_date,
				'bet' as transaction_type,
				p.bet_amount as amount
			FROM plinko p
			WHERE p.bet_amount > 0

			UNION ALL

			SELECT 
				p.user_id,
				p.timestamp as transaction_date,
				'win' as transaction_type,
				p.win_amount as amount
			FROM plinko p
			WHERE p.win_amount > 0

			UNION ALL

			-- Transactions table (for any bets/wins stored there)
			SELECT 
				t.user_id,
				t.created_at as transaction_date,
				CASE 
					WHEN t.transaction_type IN ('bet', 'groove_bet') THEN 'bet'
					WHEN t.transaction_type IN ('win', 'groove_win') THEN 'win'
					ELSE t.transaction_type
				END as transaction_type,
				CASE 
					WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount)
					ELSE t.amount
				END as amount
			FROM transactions t
			WHERE t.transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
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
				t.amount,
				t.created_at as transaction_date
			FROM users u
			LEFT JOIN brands b ON u.brand_id = b.id
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN transactions t ON u.id = t.user_id
				AND t.transaction_type IN ('deposit', 'withdrawal', 'rakeback_earned', 'rakeback_claimed')
			%s
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
					MAX(abt.transaction_date),
					MAX(pt.transaction_date),
					u.created_at
				) as last_activity,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) as total_withdrawals,
				COALESCE(SUM(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN abt.amount ELSE 0 END), 0) as total_wagered,
				COALESCE(SUM(CASE WHEN abt.transaction_type = 'win' THEN abt.amount ELSE 0 END), 0) as total_won,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'rakeback_earned' THEN pt.amount ELSE 0 END), 0) as rakeback_earned,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'rakeback_claimed' THEN pt.amount ELSE 0 END), 0) as rakeback_claimed,
				MIN(CASE WHEN pt.transaction_type = 'deposit' THEN pt.transaction_date ELSE NULL END) as first_deposit_date,
				MAX(CASE WHEN pt.transaction_type = 'deposit' THEN pt.transaction_date ELSE NULL END) as last_deposit_date,
				COUNT(DISTINCT COALESCE(abt.transaction_date, pt.transaction_date)::date) as number_of_sessions,
				COUNT(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN 1 ELSE NULL END) as number_of_bets
			FROM users u
			LEFT JOIN brands b ON u.brand_id = b.id
			LEFT JOIN balances bal ON u.id = bal.user_id AND bal.currency_code = COALESCE(u.default_currency, 'USD')
			LEFT JOIN player_transactions pt ON u.id = pt.user_id
			LEFT JOIN all_bet_transactions abt ON u.id = abt.user_id
			%s
			GROUP BY u.id, u.username, u.email, u.brand_id, b.name, u.country, u.created_at, u.status, u.kyc_status, u.default_currency, bal.amount_units
			%s
		)
		SELECT 
			user_id as player_id,
			username,
			email,
			brand_id,
			brand_name,
			country,
			registration_date,
			last_activity,
			main_balance,
			currency,
			total_deposits,
			total_withdrawals,
			total_deposits - total_withdrawals as net_deposits,
			total_wagered,
			total_won,
			rakeback_earned,
			rakeback_claimed,
			total_wagered - total_won - rakeback_earned as net_gaming_result,
			number_of_sessions,
			number_of_bets,
			account_status,
			NULL::text as device_type,
			kyc_status,
			first_deposit_date,
			last_deposit_date
		FROM player_metrics
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, whereClause, havingClause, orderBy, argIndex, argIndex+1)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

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

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating player metrics rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating player metrics rows")
	}

	// Get total count (simplified - using same query without pagination)
	countQuery := fmt.Sprintf(`
		WITH all_bet_transactions AS (
			SELECT ga.user_id, 'wager' as transaction_type, ABS(gt.amount) as amount
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE gt.type = 'wager' AND gt.amount < 0
			UNION ALL
			SELECT ga.user_id, 'win' as transaction_type, gt.amount as amount
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE gt.type = 'result' AND gt.amount > 0
			UNION ALL
			SELECT b.user_id, 'bet' as transaction_type, b.amount as amount
			FROM bets b
			WHERE b.amount > 0
			UNION ALL
			SELECT b.user_id, 'win' as transaction_type, COALESCE(b.payout, 0) as amount
			FROM bets b
			WHERE COALESCE(b.payout, 0) > 0
			UNION ALL
			SELECT sb.user_id, 'bet' as transaction_type, sb.bet_amount as amount
			FROM sport_bets sb
			WHERE sb.bet_amount > 0
			UNION ALL
			SELECT sb.user_id, 'win' as transaction_type, COALESCE(sb.actual_win, 0) as amount
			FROM sport_bets sb
			WHERE COALESCE(sb.actual_win, 0) > 0
			UNION ALL
			SELECT p.user_id, 'bet' as transaction_type, p.bet_amount as amount
			FROM plinko p
			WHERE p.bet_amount > 0
			UNION ALL
			SELECT p.user_id, 'win' as transaction_type, p.win_amount as amount
			FROM plinko p
			WHERE p.win_amount > 0
			UNION ALL
			SELECT t.user_id,
				CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN 'bet'
					WHEN t.transaction_type IN ('win', 'groove_win') THEN 'win'
					ELSE t.transaction_type END as transaction_type,
				CASE WHEN t.transaction_type IN ('bet', 'groove_bet') THEN ABS(t.amount) ELSE t.amount END as amount
			FROM transactions t
			WHERE t.transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
		),
		player_transactions AS (
			SELECT u.id as user_id, t.transaction_type, t.amount
			FROM users u
			LEFT JOIN transactions t ON u.id = t.user_id
				AND t.transaction_type IN ('deposit', 'withdrawal', 'rakeback_earned', 'rakeback_claimed')
			%s
		),
		player_metrics AS (
			SELECT 
				u.id as user_id,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'deposit' THEN pt.amount ELSE 0 END), 0) as total_deposits,
				COALESCE(SUM(CASE WHEN pt.transaction_type = 'withdrawal' THEN pt.amount ELSE 0 END), 0) as total_withdrawals,
				COALESCE(SUM(CASE WHEN abt.transaction_type IN ('bet', 'wager') THEN abt.amount ELSE 0 END), 0) as total_wagered,
				COALESCE(SUM(CASE WHEN abt.transaction_type = 'win' THEN abt.amount ELSE 0 END), 0) as total_won
			FROM users u
			LEFT JOIN player_transactions pt ON u.id = pt.user_id
			LEFT JOIN all_bet_transactions abt ON u.id = abt.user_id
			%s
			GROUP BY u.id
			%s
		)
		SELECT COUNT(*) as total
		FROM player_metrics
	`, whereClause, whereClause, havingClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get player metrics count", zap.Error(err))
		total = int64(len(metrics)) // Fallback
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

	// Query combines transactions from multiple sources
	query := fmt.Sprintf(`
		WITH all_transactions AS (
			-- GrooveTech transactions
			SELECT 
				gt.id::uuid as id,
				gt.transaction_id,
				CASE 
					WHEN gt.type = 'result' AND gt.amount > 0 THEN 'win'
					WHEN gt.type = 'wager' THEN 'bet'
					ELSE gt.type
				END as transaction_type,
				gt.created_at as transaction_date,
				ABS(gt.amount) as amount,
				COALESCE(ga.currency, 'USD') as currency,
				gt.status,
				gt.game_id as game_provider,
				gt.game_id,
				NULL::text as game_name,
				gt.transaction_id as bet_id,
				gt.round_id,
				CASE WHEN gt.type = 'wager' THEN ABS(gt.amount) ELSE NULL END as bet_amount,
				CASE WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount ELSE NULL END as win_amount,
				NULL::decimal as rakeback_earned,
				NULL::decimal as rakeback_claimed,
				NULL::decimal as rtp,
				CASE 
					WHEN gt.type = 'wager' AND gt.type = 'result' AND gt.amount > 0 AND (SELECT amount FROM groove_transactions gt2 WHERE gt2.round_id = gt.round_id AND gt2.type = 'wager' LIMIT 1) > 0
					THEN gt.amount / (SELECT amount FROM groove_transactions gt2 WHERE gt2.round_id = gt.round_id AND gt2.type = 'wager' LIMIT 1)
					ELSE NULL
				END as multiplier,
				NULL::decimal as ggr,
				NULL::decimal as net,
				'cash' as bet_type,
				NULL::text as payment_method,
				NULL::text as tx_hash,
				NULL::text as network,
				NULL::text as chain_id,
				NULL::decimal as fees,
				NULL::text as device,
				NULL::text as ip_address,
				gt.game_session_id as session_id
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE ga.user_id = $1
				AND gt.type IN ('wager', 'result')
				%s

			UNION ALL

			-- General bets
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

			-- Sport bets
			SELECT 
				sb.id::uuid as id,
				sb.transaction_id,
				CASE WHEN COALESCE(sb.actual_win, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				sb.created_at as transaction_date,
				COALESCE(sb.actual_win, sb.bet_amount) as amount,
				sb.currency,
				COALESCE(sb.status, 'completed') as status,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				sb.transaction_id as bet_id,
				NULL::text as round_id,
				sb.bet_amount as bet_amount,
				COALESCE(sb.actual_win, 0) as win_amount,
				NULL::decimal as rakeback_earned,
				NULL::decimal as rakeback_claimed,
				NULL::decimal as rtp,
				CASE WHEN sb.bet_amount > 0 THEN COALESCE(sb.actual_win, 0) / sb.bet_amount ELSE NULL END as multiplier,
				sb.bet_amount - COALESCE(sb.actual_win, 0) as ggr,
				COALESCE(sb.actual_win, 0) - sb.bet_amount as net,
				'cash' as bet_type,
				NULL::text as payment_method,
				NULL::text as tx_hash,
				NULL::text as network,
				NULL::text as chain_id,
				NULL::decimal as fees,
				NULL::text as device,
				NULL::text as ip_address,
				NULL::text as session_id
			FROM sport_bets sb
			WHERE sb.user_id = $1
				AND (COALESCE(sb.actual_win, 0) > 0 OR sb.bet_amount > 0)
				%s

			UNION ALL

			-- Withdrawals
			SELECT 
				w.id,
				w.withdrawal_id as transaction_id,
				'withdrawal' as transaction_type,
				w.created_at as transaction_date,
				(w.usd_amount_cents::decimal / 100) as amount,
				w.currency_code as currency,
				w.status,
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
	`, whereClause, whereClause, whereClause, whereClause, whereClause, orderBy, argIndex, argIndex+1)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
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
			r.log.Error("failed to scan player transaction row", zap.Error(err))
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

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH all_transactions AS (
			SELECT gt.created_at as transaction_date, gt.type
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			WHERE ga.user_id = $1 AND gt.type IN ('wager', 'result')
			%s
			UNION ALL
			SELECT COALESCE(b.timestamp, NOW()) as transaction_date, 'bet' as type
			FROM bets b WHERE b.user_id = $1 %s
			UNION ALL
			SELECT sb.created_at as transaction_date, 'bet' as type
			FROM sport_bets sb WHERE sb.user_id = $1 %s
			UNION ALL
			SELECT w.created_at as transaction_date, 'withdrawal' as type
			FROM withdrawals w WHERE w.user_id = $1 %s
		)
		SELECT COUNT(*) as total
		FROM all_transactions
	`, whereClause, whereClause, whereClause, whereClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get player transactions count", zap.Error(err))
		total = int64(len(transactions))
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Player transactions retrieved successfully"
	res.Data = transactions
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage

	return res, nil
}
