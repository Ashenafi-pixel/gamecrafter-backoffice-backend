package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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

	// Build referral code filter
	referralCodeFilter := ""
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

	// Main query - aggregates metrics by date and referral code
	// Use dateFrom and dateTo multiple times, so we'll reuse $1 and $2
	query := fmt.Sprintf(`
		WITH date_range AS (
			SELECT generate_series(
				$1::date,
				$2::date,
				'1 day'::interval
			)::date as date
		),
		user_registrations AS (
			SELECT 
				DATE(u.created_at) as date,
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COUNT(DISTINCT u.id) as registrations
			FROM users u
			WHERE DATE(u.created_at) >= $1
				AND DATE(u.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
			GROUP BY DATE(u.created_at), COALESCE(u.refered_by_code, 'N/A')
		),
		user_deposits AS (
			SELECT 
				DATE(t.created_at) as date,
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
			GROUP BY DATE(t.created_at), COALESCE(u.refered_by_code, 'N/A')
		),
		user_withdrawals AS (
			SELECT 
				DATE(t.created_at) as date,
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'withdrawal' THEN t.amount ELSE 0 END), 0) as withdrawals_usd
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'withdrawal'
				AND DATE(t.created_at) >= $1
				AND DATE(t.created_at) <= $2
				AND COALESCE(u.refered_by_code, '') != ''
				%s
			GROUP BY DATE(t.created_at), COALESCE(u.refered_by_code, 'N/A')
		),
		user_activity AS (
			SELECT 
				DATE(COALESCE(
					(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
					u.created_at
				)) as date,
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
			GROUP BY DATE(COALESCE(
				(SELECT MAX(t2.created_at) FROM transactions t2 WHERE t2.user_id = u.id),
				u.created_at
			)), COALESCE(u.refered_by_code, 'N/A')
		),
		bet_metrics AS (
			SELECT 
				DATE(COALESCE(gt.created_at, b.timestamp, sb.created_at, p.timestamp, NOW())) as date,
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
			GROUP BY DATE(COALESCE(gt.created_at, b.timestamp, sb.created_at, p.timestamp, NOW())), COALESCE(u.refered_by_code, 'N/A')
		)
		SELECT 
			dr.date::text as date,
			COALESCE(ur.referral_code, ud.referral_code, uw.referral_code, ua.referral_code, bm.referral_code, 'N/A') as referral_code,
			COALESCE(ur.registrations, 0) as registrations,
			COALESCE(ud.unique_depositors, 0) as unique_depositors,
			COALESCE(ua.active_customers, 0) as active_customers,
			COALESCE(bm.total_bets, 0) as total_bets,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) as ggr,
			COALESCE(bm.total_stake, 0) - COALESCE(bm.total_win, 0) - COALESCE(bm.rakeback_earned, 0) as ngr,
			COALESCE(ud.deposits_usd, 0) as deposits_usd,
			COALESCE(uw.withdrawals_usd, 0) as withdrawals_usd
		FROM date_range dr
		LEFT JOIN user_registrations ur ON dr.date = ur.date
		LEFT JOIN user_deposits ud ON dr.date = ud.date AND COALESCE(ur.referral_code, 'N/A') = COALESCE(ud.referral_code, 'N/A')
		LEFT JOIN user_withdrawals uw ON dr.date = uw.date AND COALESCE(ur.referral_code, 'N/A') = COALESCE(uw.referral_code, 'N/A')
		LEFT JOIN user_activity ua ON dr.date = ua.date AND COALESCE(ur.referral_code, 'N/A') = COALESCE(ua.referral_code, 'N/A')
		LEFT JOIN bet_metrics bm ON dr.date = bm.date AND COALESCE(ur.referral_code, 'N/A') = COALESCE(bm.referral_code, 'N/A')
		WHERE COALESCE(ur.registrations, 0) > 0 
			OR COALESCE(ud.unique_depositors, 0) > 0
			OR COALESCE(ua.active_customers, 0) > 0
			OR COALESCE(bm.total_bets, 0) > 0
		ORDER BY dr.date DESC, referral_code
	`, referralCodeFilter, referralCodeFilter, referralCodeFilter, referralCodeFilter, referralCodeFilter)

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
		err := rows.Scan(
			&row.Date,
			&row.ReferralCode,
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

