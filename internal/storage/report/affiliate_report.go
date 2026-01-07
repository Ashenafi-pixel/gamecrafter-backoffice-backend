package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"go.uber.org/zap"
)

// GetAffiliateReport retrieves affiliate report with daily metrics grouped by referral code
func (r *report) GetAffiliateReport(ctx context.Context, req dto.AffiliateReportReq, allowedReferralCodes []string) (dto.AffiliateReportRes, error) {
	var res dto.AffiliateReportRes

	// Parse date ranges
	var dateFrom, dateTo time.Time
	var err error

	if req.DateFrom == nil || *req.DateFrom == "" {
		// Default to last 30 days
		dateTo = time.Now()
		dateFrom = dateTo.AddDate(0, 0, -30)
	} else {
		dateFrom, err = time.Parse("2006-01-02", *req.DateFrom)
		if err != nil {
			r.log.Error("Invalid date_from format", zap.Error(err))
			return res, errors.ErrInvalidUserInput.Wrap(err, "invalid date_from format, expected YYYY-MM-DD")
		}
		// Set to start of day
		dateFrom = time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, dateFrom.Location())
	}

	if req.DateTo == nil || *req.DateTo == "" {
		dateTo = time.Now()
	} else {
		dateTo, err = time.Parse("2006-01-02", *req.DateTo)
		if err != nil {
			r.log.Error("Invalid date_to format", zap.Error(err))
			return res, errors.ErrInvalidUserInput.Wrap(err, "invalid date_to format, expected YYYY-MM-DD")
		}
		// Set to end of day
		dateTo = time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, dateTo.Location())
	}

	// Build referral code filter and test account filter
	referralCodeFilter := ""
	testAccountFilter := ""
	args := []interface{}{dateFrom, dateTo} // Start with date range for date_range CTE
	argIndex := 3

	if len(allowedReferralCodes) > 0 {
		placeholders := []string{}
		for _, code := range allowedReferralCodes {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, code)
			argIndex++
		}
		referralCodeFilter = fmt.Sprintf("AND COALESCE(u.refered_by_code, '') IN (%s)", strings.Join(placeholders, ","))
	}

	// Additional referral code filter from request
	if req.ReferralCode != nil && *req.ReferralCode != "" {
		referralCodeFilter += fmt.Sprintf(" AND COALESCE(u.refered_by_code, '') = $%d", argIndex)
		args = append(args, *req.ReferralCode)
		argIndex++
	}

	// Test account filter
	if req.IsTestAccount != nil {
		testAccountFilter = fmt.Sprintf(" AND u.is_test_account = $%d", argIndex)
		args = append(args, *req.IsTestAccount)
		argIndex++
	}

	// Main query - aggregates metrics by referral code (grouped, not by date)
	// Use dateFrom and dateTo multiple times, so we'll reuse $1 and $2
	query := fmt.Sprintf(`
		WITH affiliate_users AS (
			SELECT DISTINCT
				COALESCE(u.referal_code, '') as referral_code,
				COALESCE(u.username, '') as affiliate_username,
				COALESCE(u.email, '') as affiliate_email
			FROM users u
			WHERE COALESCE(u.referal_code, '') != ''
		),
		user_registrations AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COUNT(DISTINCT u.id) as registrations
			FROM users u
			WHERE DATE(u.created_at) >= $1
				AND DATE(u.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
				%s
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		user_deposits AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COUNT(DISTINCT CASE WHEN t.transaction_type = 'deposit' THEN u.id ELSE NULL END) as unique_depositors,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as deposits_usd
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'deposit'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
				%s
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		user_withdrawals AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'withdrawal' THEN t.amount ELSE 0 END), 0) as withdrawals_usd
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'withdrawal'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
				%s
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		user_activity AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COUNT(DISTINCT u.id) as active_customers
			FROM users u
			WHERE COALESCE(
				(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
				u.created_at
			) >= $1
				AND COALESCE(
					(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
					u.created_at
				) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
				%s
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		bet_metrics AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COUNT(CASE 
					WHEN gt.type = 'wager' AND gt.amount < 0 THEN 1
					WHEN b.amount > 0 THEN 1
					WHEN sb.bet_amount > 0 THEN 1
					WHEN p.bet_amount > 0 THEN 1
					ELSE NULL
				END) as total_bets,
				COALESCE(SUM(CASE 
					WHEN gt.type = 'wager' AND gt.amount < 0 THEN ABS(gt.amount)
					WHEN b.amount > 0 THEN b.amount
					WHEN sb.bet_amount > 0 THEN sb.bet_amount
					WHEN p.bet_amount > 0 THEN p.bet_amount
					ELSE 0
				END), 0) as total_stake,
				COALESCE(SUM(CASE 
					WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount
					WHEN b.payout > 0 THEN b.payout
					WHEN sb.actual_win > 0 THEN sb.actual_win
					WHEN p.win_amount > 0 THEN p.win_amount
					ELSE 0
				END), 0) as total_win,
				COALESCE(SUM(CASE 
					WHEN t.transaction_type = 'rakeback_earned' THEN t.amount
					ELSE 0
				END), 0) as rakeback_earned
			FROM users u
			LEFT JOIN groove_accounts ga ON u.id = ga.user_id
			LEFT JOIN groove_transactions gt ON ga.account_id = gt.account_id
				AND gt.type IN ('wager', 'result')
				AND DATE(gt.created_at) >= $1
				AND DATE(gt.created_at) <= $2
			LEFT JOIN bets b ON u.id = b.user_id
				AND DATE(COALESCE(b.timestamp, NOW())) >= $1
				AND DATE(COALESCE(b.timestamp, NOW())) <= $2
			LEFT JOIN sport_bets sb ON u.id = sb.user_id
				AND DATE(sb.created_at) >= $1
				AND DATE(sb.created_at) <= $2
			LEFT JOIN plinko p ON u.id = p.user_id
				AND DATE(p.timestamp) >= $1
				AND DATE(p.timestamp) <= $2
			LEFT JOIN transactions t ON u.id = t.user_id
				AND t.transaction_type = 'rakeback_earned'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
			WHERE COALESCE(u.refered_by_code, '') != ''
				%s
				%s
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		all_referral_codes AS (
			SELECT DISTINCT referral_code
			FROM (
				SELECT referral_code FROM user_registrations
				UNION
				SELECT referral_code FROM user_deposits
				UNION
				SELECT referral_code FROM user_withdrawals
				UNION
				SELECT referral_code FROM user_activity
				UNION
				SELECT referral_code FROM bet_metrics
			) combined
		)
		SELECT 
			arc.referral_code,
			COALESCE(au.affiliate_username, 'N/A') as affiliate_username,
			COALESCE(ur.registrations, 0) as registrations,
			COALESCE(ud.unique_depositors, 0) as unique_depositors,
			COALESCE(ua.active_customers, 0) as active_customers,
			COALESCE(bm.total_bets, 0) as total_bets,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) as ggr,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) - COALESCE(bm.rakeback_earned, 0) as ngr,
			COALESCE(ud.deposits_usd, 0) as deposits_usd,
			COALESCE(uw.withdrawals_usd, 0) as withdrawals_usd
		FROM all_referral_codes arc
		LEFT JOIN user_registrations ur ON arc.referral_code = ur.referral_code
		LEFT JOIN user_deposits ud ON arc.referral_code = ud.referral_code
		LEFT JOIN user_withdrawals uw ON arc.referral_code = uw.referral_code
		LEFT JOIN user_activity ua ON arc.referral_code = ua.referral_code
		LEFT JOIN bet_metrics bm ON arc.referral_code = bm.referral_code
		LEFT JOIN affiliate_users au ON arc.referral_code = au.referral_code
		WHERE COALESCE(ur.registrations, 0) > 0 
			OR COALESCE(ud.unique_depositors, 0) > 0
			OR COALESCE(ua.active_customers, 0) > 0
			OR COALESCE(bm.total_bets, 0) > 0
		ORDER BY referral_code
	`, referralCodeFilter, testAccountFilter, referralCodeFilter, testAccountFilter, referralCodeFilter, testAccountFilter, referralCodeFilter, testAccountFilter, referralCodeFilter, testAccountFilter)

	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to execute affiliate report query", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get affiliate report")
	}
	defer rows.Close()

	var data []dto.AffiliateReportRow
	var summary dto.AffiliateReportSummary

	for rows.Next() {
		var row dto.AffiliateReportRow
		// Set date to empty string since we're grouping by referral code only
		row.Date = ""
		err := rows.Scan(
			&row.ReferralCode,
			&row.AffiliateUsername,
			&row.Registrations,
			&row.UniqueDepositors,
			&row.ActiveCustomers,
			&row.TotalBets,
			&row.GGR,
			&row.NGR,
			&row.DepositsUSD,
			&row.WithdrawalsUSD,
		)
		if err != nil {
			r.log.Error("Failed to scan affiliate report row", zap.Error(err))
			continue
		}

		data = append(data, row)

		// Accumulate summary
		summary.TotalRegistrations += row.Registrations
		summary.TotalUniqueDepositors += row.UniqueDepositors
		summary.TotalActiveCustomers += row.ActiveCustomers
		summary.TotalBets += row.TotalBets
		summary.TotalGGR = summary.TotalGGR.Add(row.GGR)
		summary.TotalNGR = summary.TotalNGR.Add(row.NGR)
		summary.TotalDepositsUSD = summary.TotalDepositsUSD.Add(row.DepositsUSD)
		summary.TotalWithdrawalsUSD = summary.TotalWithdrawalsUSD.Add(row.WithdrawalsUSD)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error iterating affiliate report rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to process affiliate report")
	}

	res.Data = data
	res.Summary = summary
	res.Message = "Affiliate report retrieved successfully"

	return res, nil
}

// GetAffiliatePlayersReport retrieves player-level metrics for a specific referral code
func (r *report) GetAffiliatePlayersReport(ctx context.Context, req dto.AffiliatePlayersReportReq, allowedReferralCodes []string) (dto.AffiliatePlayersReportRes, error) {
	var res dto.AffiliatePlayersReportRes

	if req.ReferralCode == nil || *req.ReferralCode == "" {
		return res, errors.ErrInvalidUserInput.New("referral_code is required")
	}

	// Parse date ranges
	var dateFrom, dateTo time.Time
	var err error

	if req.DateFrom == nil || *req.DateFrom == "" {
		// Default to last 30 days
		dateTo = time.Now()
		dateFrom = dateTo.AddDate(0, 0, -30)
	} else {
		dateFrom, err = time.Parse("2006-01-02", *req.DateFrom)
		if err != nil {
			r.log.Error("Invalid date_from format", zap.Error(err))
			return res, errors.ErrInvalidUserInput.Wrap(err, "invalid date_from format, expected YYYY-MM-DD")
		}
		dateFrom = time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, dateFrom.Location())
	}

	if req.DateTo == nil || *req.DateTo == "" {
		dateTo = time.Now()
	} else {
		dateTo, err = time.Parse("2006-01-02", *req.DateTo)
		if err != nil {
			r.log.Error("Invalid date_to format", zap.Error(err))
			return res, errors.ErrInvalidUserInput.Wrap(err, "invalid date_to format, expected YYYY-MM-DD")
		}
		dateTo = time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, dateTo.Location())
	}

	// Build test account filter
	testAccountFilter := ""
	args := []interface{}{dateFrom, dateTo, *req.ReferralCode}
	argIndex := 4

	if req.IsTestAccount != nil {
		testAccountFilter = fmt.Sprintf(" AND u.is_test_account = $%d", argIndex)
		args = append(args, *req.IsTestAccount)
		argIndex++
	}

	// Query to get player-level metrics grouped by user
	query := fmt.Sprintf(`
		WITH user_deposits AS (
			SELECT 
				u.id as user_id,
				u.username,
				u.email,
				COUNT(DISTINCT CASE WHEN t.transaction_type = 'deposit' THEN t.id ELSE NULL END) as deposit_count,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'deposit' THEN t.amount ELSE 0 END), 0) as deposits_usd
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'deposit'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') = $3
				%s
			GROUP BY u.id, u.username, u.email
		),
		user_withdrawals AS (
			SELECT 
				u.id as user_id,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'withdrawal' THEN t.amount ELSE 0 END), 0) as withdrawals_usd
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'withdrawal'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') = $3
				%s
			GROUP BY u.id
		),
		user_activity AS (
			SELECT 
				u.id as user_id,
				1 as is_active
			FROM users u
			WHERE COALESCE(
				(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
				u.created_at
			) >= $1
				AND COALESCE(
					(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
					u.created_at
				) <= $2
				AND COALESCE(u.refered_by_code, '') = $3
				%s
		),
		bet_metrics AS (
			SELECT 
				u.id as user_id,
				COUNT(CASE 
					WHEN gt.type = 'wager' AND gt.amount < 0 THEN 1
					WHEN b.amount > 0 THEN 1
					WHEN sb.bet_amount > 0 THEN 1
					WHEN p.bet_amount > 0 THEN 1
					ELSE NULL
				END) as total_bets,
				COALESCE(SUM(CASE 
					WHEN gt.type = 'wager' AND gt.amount < 0 THEN ABS(gt.amount)
					WHEN b.amount > 0 THEN b.amount
					WHEN sb.bet_amount > 0 THEN sb.bet_amount
					WHEN p.bet_amount > 0 THEN p.bet_amount
					ELSE 0
				END), 0) as total_stake,
				COALESCE(SUM(CASE 
					WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount
					WHEN b.payout > 0 THEN b.payout
					WHEN sb.actual_win > 0 THEN sb.actual_win
					WHEN p.win_amount > 0 THEN p.win_amount
					ELSE 0
				END), 0) as total_win,
				COALESCE(SUM(CASE 
					WHEN t.transaction_type = 'rakeback_earned' THEN t.amount
					ELSE 0
				END), 0) as rakeback_earned
			FROM users u
			LEFT JOIN groove_accounts ga ON u.id = ga.user_id
			LEFT JOIN groove_transactions gt ON ga.account_id = gt.account_id
				AND gt.type IN ('wager', 'result')
				AND DATE(gt.created_at) >= $1
				AND DATE(gt.created_at) <= $2
			LEFT JOIN bets b ON u.id = b.user_id
				AND DATE(COALESCE(b.timestamp, NOW())) >= $1
				AND DATE(COALESCE(b.timestamp, NOW())) <= $2
			LEFT JOIN sport_bets sb ON u.id = sb.user_id
				AND DATE(sb.created_at) >= $1
				AND DATE(sb.created_at) <= $2
			LEFT JOIN plinko p ON u.id = p.user_id
				AND DATE(p.timestamp) >= $1
				AND DATE(p.timestamp) <= $2
			LEFT JOIN transactions t ON u.id = t.user_id
				AND t.transaction_type = 'rakeback_earned'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
			WHERE COALESCE(u.refered_by_code, '') = $3
				%s
			GROUP BY u.id
		),
		all_users AS (
			SELECT DISTINCT u.id, u.username, u.email, u.created_at
			FROM users u
			WHERE COALESCE(u.refered_by_code, '') = $3
				%s
		)
		SELECT 
			au.id::text as player_id,
			COALESCE(au.username, '') as username,
			COALESCE(au.email, '') as email,
			CASE WHEN DATE(au.created_at) >= $1 AND DATE(au.created_at) <= $2 THEN 1 ELSE 0 END as registrations,
			CASE WHEN ud.user_id IS NOT NULL THEN 1 ELSE 0 END as unique_depositors,
			CASE WHEN ua.user_id IS NOT NULL THEN 1 ELSE 0 END as active_customers,
			COALESCE(bm.total_bets, 0) as total_bets,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) as ggr,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) - COALESCE(bm.rakeback_earned, 0) as ngr,
			COALESCE(ud.deposits_usd, 0) as deposits_usd,
			COALESCE(uw.withdrawals_usd, 0) as withdrawals_usd
		FROM all_users au
		LEFT JOIN user_deposits ud ON au.id = ud.user_id
		LEFT JOIN user_withdrawals uw ON au.id = uw.user_id
		LEFT JOIN user_activity ua ON au.id = ua.user_id
		LEFT JOIN bet_metrics bm ON au.id = bm.user_id
		ORDER BY au.username
	`, testAccountFilter, testAccountFilter, testAccountFilter, testAccountFilter, testAccountFilter)

	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to execute affiliate players report query", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get affiliate players report")
	}
	defer rows.Close()

	var data []dto.AffiliatePlayerReportRow
	var summary dto.AffiliateReportSummary

	for rows.Next() {
		var row dto.AffiliatePlayerReportRow
		err := rows.Scan(
			&row.PlayerID,
			&row.Username,
			&row.Email,
			&row.Registrations,
			&row.UniqueDepositors,
			&row.ActiveCustomers,
			&row.TotalBets,
			&row.GGR,
			&row.NGR,
			&row.DepositsUSD,
			&row.WithdrawalsUSD,
		)
		if err != nil {
			r.log.Error("Failed to scan affiliate player report row", zap.Error(err))
			continue
		}

		data = append(data, row)

		// Accumulate summary
		summary.TotalRegistrations += row.Registrations
		summary.TotalUniqueDepositors += row.UniqueDepositors
		summary.TotalActiveCustomers += row.ActiveCustomers
		summary.TotalBets += row.TotalBets
		summary.TotalGGR = summary.TotalGGR.Add(row.GGR)
		summary.TotalNGR = summary.TotalNGR.Add(row.NGR)
		summary.TotalDepositsUSD = summary.TotalDepositsUSD.Add(row.DepositsUSD)
		summary.TotalWithdrawalsUSD = summary.TotalWithdrawalsUSD.Add(row.WithdrawalsUSD)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error iterating affiliate players report rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to process affiliate players report")
	}

	res.Data = data
	res.Summary = summary
	res.Message = "Affiliate players report retrieved successfully"

	return res, nil
}
