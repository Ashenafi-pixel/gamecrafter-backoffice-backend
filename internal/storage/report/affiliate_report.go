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

	userReferralMap := make(map[string]string) // user_id -> referral_code

	userQueryArgs := []interface{}{}
	userQueryFilters := []string{"COALESCE(u.refered_by_code, '') != ''"}
	userArgIndex := 1

	if len(allowedReferralCodes) > 0 {
		placeholders := []string{}
		for _, code := range allowedReferralCodes {
			placeholders = append(placeholders, fmt.Sprintf("$%d", userArgIndex))
			userQueryArgs = append(userQueryArgs, code)
			userArgIndex++
		}
		userQueryFilters = append(userQueryFilters, fmt.Sprintf("COALESCE(u.refered_by_code, '') IN (%s)", strings.Join(placeholders, ",")))
	}

	if req.ReferralCode != nil && *req.ReferralCode != "" {
		userQueryFilters = append(userQueryFilters, fmt.Sprintf("COALESCE(u.refered_by_code, '') = $%d", userArgIndex))
		userQueryArgs = append(userQueryArgs, *req.ReferralCode)
		userArgIndex++
	}

	if req.IsTestAccount != nil {
		userQueryFilters = append(userQueryFilters, fmt.Sprintf("u.is_test_account = $%d", userArgIndex))
		userQueryArgs = append(userQueryArgs, *req.IsTestAccount)
		userArgIndex++
	}

	userQuery := fmt.Sprintf(`
		SELECT DISTINCT
			u.id::text as user_id,
			COALESCE(u.refered_by_code, 'N/A') as referral_code
		FROM users u
		WHERE %s
	`, strings.Join(userQueryFilters, " AND "))

	userRows, err := r.db.GetPool().Query(ctx, userQuery, userQueryArgs...)
	if err != nil {
		r.log.Warn("Failed to query users for referral map", zap.Error(err))
	} else {
		defer userRows.Close()
		for userRows.Next() {
			var userID, referralCode string
			if err := userRows.Scan(&userID, &referralCode); err == nil {
				userReferralMap[userID] = referralCode
			}
		}
		r.log.Debug("Built user referral map", zap.Int("user_count", len(userReferralMap)))
	}

	// Get deposits and withdrawals from ClickHouse
	chConn, hasClickHouse := r.getClickHouseConn()
	depositsByReferral := make(map[string]struct {
		UniqueDepositors int64
		DepositsUSD      decimal.Decimal
	})
	withdrawalsByReferral := make(map[string]decimal.Decimal)
	gamingMetricsByReferral := make(map[string]struct {
		TotalBets      int64
		TotalStake     decimal.Decimal
		TotalWin       decimal.Decimal
		RakebackEarned decimal.Decimal
	})

	if hasClickHouse && len(userReferralMap) > 0 {
		userIDStrings := make([]string, 0, len(userReferralMap))
		for userID := range userReferralMap {
			userIDStrings = append(userIDStrings, fmt.Sprintf("'%s'", userID))
		}
		userIDFilter := strings.Join(userIDStrings, ",")

		dateFromStr := dateFrom.Format("2006-01-02")
		dateToStr := dateTo.Format("2006-01-02")

		r.log.Debug("Querying ClickHouse for deposits",
			zap.Int("user_count", len(userReferralMap)),
			zap.String("date_from", dateFromStr),
			zap.String("date_to", dateToStr))

		// Query deposits from ClickHouse
		depositsQuery := fmt.Sprintf(`
			SELECT 
				toString(user_id) as user_id,
				toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as deposits_usd
			FROM tucanbit_financial.deposits
			WHERE toString(status) = 'verified'
				AND toDate(created_at) >= '%s'
				AND toDate(created_at) <= '%s'
				AND toString(user_id) IN (%s)
			GROUP BY user_id
		`, dateFromStr, dateToStr, userIDFilter)

		depRows, err := chConn.Query(ctx, depositsQuery)
		if err == nil {
			defer depRows.Close()
			uniqueDepositorsByReferral := make(map[string]map[string]bool)
			depositRowCount := 0
			for depRows.Next() {
				depositRowCount++
				var userIDStr, depositsUSDStr string
				if err := depRows.Scan(&userIDStr, &depositsUSDStr); err == nil {
					referralCode := userReferralMap[userIDStr]
					if referralCode == "" {
						referralCode = "N/A"
						r.log.Warn("User ID not found in referral map", zap.String("user_id", userIDStr))
					}
					depositsUSD, parseErr := decimal.NewFromString(depositsUSDStr)
					if parseErr != nil {
						r.log.Warn("Failed to parse deposits USD", zap.String("user_id", userIDStr), zap.String("value", depositsUSDStr), zap.Error(parseErr))
						continue
					}

					if existing, ok := depositsByReferral[referralCode]; ok {
						existing.DepositsUSD = existing.DepositsUSD.Add(depositsUSD)
						depositsByReferral[referralCode] = existing
					} else {
						depositsByReferral[referralCode] = struct {
							UniqueDepositors int64
							DepositsUSD      decimal.Decimal
						}{
							UniqueDepositors: 0,
							DepositsUSD:      depositsUSD,
						}
					}

					if uniqueDepositorsByReferral[referralCode] == nil {
						uniqueDepositorsByReferral[referralCode] = make(map[string]bool)
					}
					uniqueDepositorsByReferral[referralCode][userIDStr] = true
				} else {
					r.log.Warn("Failed to scan deposit row", zap.Error(err))
				}
			}
			// Count unique depositors per referral code
			for referralCode, userMap := range uniqueDepositorsByReferral {
				if existing, ok := depositsByReferral[referralCode]; ok {
					existing.UniqueDepositors = int64(len(userMap))
					depositsByReferral[referralCode] = existing
				}
			}
			r.log.Debug("ClickHouse deposits query completed",
				zap.Int("rows_processed", depositRowCount),
				zap.Int("referral_codes_with_deposits", len(depositsByReferral)))
		} else {
			r.log.Warn("Failed to query ClickHouse for deposits", zap.Error(err), zap.String("query", depositsQuery))
		}

		// Query withdrawals from ClickHouse
		withdrawalsQuery := fmt.Sprintf(`
			SELECT 
				toString(user_id) as user_id,
				toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as withdrawals_usd
			FROM tucanbit_financial.withdrawals
			WHERE toString(status) = 'completed'
				AND toDate(created_at) >= '%s'
				AND toDate(created_at) <= '%s'
				AND toString(user_id) IN (%s)
			GROUP BY user_id
		`, dateFromStr, dateToStr, userIDFilter)

		withRows, err := chConn.Query(ctx, withdrawalsQuery)
		if err == nil {
			defer withRows.Close()
			withdrawalRowCount := 0
			for withRows.Next() {
				withdrawalRowCount++
				var userIDStr, withdrawalsUSDStr string
				if err := withRows.Scan(&userIDStr, &withdrawalsUSDStr); err == nil {
					referralCode := userReferralMap[userIDStr]
					if referralCode == "" {
						referralCode = "N/A"
						r.log.Warn("User ID not found in referral map for withdrawals", zap.String("user_id", userIDStr))
					}
					withdrawalsUSD, parseErr := decimal.NewFromString(withdrawalsUSDStr)
					if parseErr != nil {
						r.log.Warn("Failed to parse withdrawals USD", zap.String("user_id", userIDStr), zap.String("value", withdrawalsUSDStr), zap.Error(parseErr))
						continue
					}

					if existing, ok := withdrawalsByReferral[referralCode]; ok {
						withdrawalsByReferral[referralCode] = existing.Add(withdrawalsUSD)
					} else {
						withdrawalsByReferral[referralCode] = withdrawalsUSD
					}
				} else {
					r.log.Warn("Failed to scan withdrawal row", zap.Error(err))
				}
			}
			r.log.Debug("ClickHouse withdrawals query completed",
				zap.Int("rows_processed", withdrawalRowCount),
				zap.Int("referral_codes_with_withdrawals", len(withdrawalsByReferral)))
		} else {
			r.log.Warn("Failed to query ClickHouse for withdrawals", zap.Error(err), zap.String("query", withdrawalsQuery))
		}

		// Query gaming metrics from ClickHouse
		dateFromStrGaming := dateFrom.Format("2006-01-02 15:04:05")
		dateToStrGaming := dateTo.Format("2006-01-02 15:04:05")
		gamingQuery := fmt.Sprintf(`
				SELECT 
					toString(user_id) as user_id,
					toString(toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8)) as total_wagered,
					toString(toDecimal64(sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8)) as total_won,
					toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as total_bets,
					toString(toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8)) as rakeback_earned
				FROM tucanbit_analytics.transactions
				WHERE created_at >= '%s'
					AND created_at <= '%s'
					AND toString(user_id) IN (%s)
					AND toString(status) = 'completed'
				GROUP BY user_id
			`, dateFromStrGaming, dateToStrGaming, userIDFilter)

		gamingRows, err := chConn.Query(ctx, gamingQuery)
		if err == nil {
			defer gamingRows.Close()
			gamingRowCount := 0
			for gamingRows.Next() {
				gamingRowCount++
				var userIDStr, totalWageredStr, totalWonStr, rakebackEarnedStr string
				var totalBets uint64
				if err := gamingRows.Scan(&userIDStr, &totalWageredStr, &totalWonStr, &totalBets, &rakebackEarnedStr); err == nil {
					referralCode := userReferralMap[userIDStr]
					if referralCode == "" {
						referralCode = "N/A"
						r.log.Warn("User ID not found in referral map for gaming metrics", zap.String("user_id", userIDStr))
					}
					totalWagered, parseErr1 := decimal.NewFromString(totalWageredStr)
					totalWon, parseErr2 := decimal.NewFromString(totalWonStr)
					rakebackEarned, parseErr3 := decimal.NewFromString(rakebackEarnedStr)
					if parseErr1 != nil || parseErr2 != nil || parseErr3 != nil {
						r.log.Warn("Failed to parse gaming metrics",
							zap.String("user_id", userIDStr),
							zap.Error(parseErr1),
							zap.Error(parseErr2),
							zap.Error(parseErr3))
						continue
					}

					if existing, ok := gamingMetricsByReferral[referralCode]; ok {
						existing.TotalBets += int64(totalBets)
						existing.TotalStake = existing.TotalStake.Add(totalWagered)
						existing.TotalWin = existing.TotalWin.Add(totalWon)
						existing.RakebackEarned = existing.RakebackEarned.Add(rakebackEarned)
						gamingMetricsByReferral[referralCode] = existing
					} else {
						gamingMetricsByReferral[referralCode] = struct {
							TotalBets      int64
							TotalStake     decimal.Decimal
							TotalWin       decimal.Decimal
							RakebackEarned decimal.Decimal
						}{
							TotalBets:      int64(totalBets),
							TotalStake:     totalWagered,
							TotalWin:       totalWon,
							RakebackEarned: rakebackEarned,
						}
					}
				} else {
					r.log.Warn("Failed to scan gaming metrics row", zap.Error(err))
				}
			}
			r.log.Debug("ClickHouse gaming metrics query completed",
				zap.Int("rows_processed", gamingRowCount),
				zap.Int("referral_codes_with_gaming", len(gamingMetricsByReferral)))
		} else {
			r.log.Warn("Failed to query ClickHouse for gaming metrics", zap.Error(err), zap.String("query", gamingQuery))
		}
	}

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
				0 as unique_depositors,
				0 as deposits_usd
			FROM users u
			WHERE 1=0
			GROUP BY COALESCE(u.refered_by_code, 'N/A')
		),
		user_withdrawals AS (
			SELECT 
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				0 as withdrawals_usd
			FROM users u
			WHERE 1=0
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
				0 as total_bets,
				0 as total_stake,
				0 as total_win,
				0 as rakeback_earned
			FROM users u
			WHERE 1=0
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
	`, referralCodeFilter, testAccountFilter, referralCodeFilter, testAccountFilter)

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

		if chDeposits, ok := depositsByReferral[row.ReferralCode]; ok {
			row.UniqueDepositors = int64(chDeposits.UniqueDepositors)
			row.DepositsUSD = chDeposits.DepositsUSD
			r.log.Debug("Merged ClickHouse deposits data",
				zap.String("referral_code", row.ReferralCode),
				zap.Int64("unique_depositors", row.UniqueDepositors),
				zap.String("deposits_usd", row.DepositsUSD.String()))
		}
		if chWithdrawals, ok := withdrawalsByReferral[row.ReferralCode]; ok {
			row.WithdrawalsUSD = chWithdrawals
			r.log.Debug("Merged ClickHouse withdrawals data",
				zap.String("referral_code", row.ReferralCode),
				zap.String("withdrawals_usd", row.WithdrawalsUSD.String()))
		}

		if chGaming, ok := gamingMetricsByReferral[row.ReferralCode]; ok {
			row.TotalBets = chGaming.TotalBets
			row.GGR = chGaming.TotalStake.Sub(chGaming.TotalWin)
			row.NGR = chGaming.TotalStake.Sub(chGaming.TotalWin).Sub(chGaming.RakebackEarned)
			r.log.Debug("Merged ClickHouse gaming metrics",
				zap.String("referral_code", row.ReferralCode),
				zap.Int64("total_bets", row.TotalBets),
				zap.String("ggr", row.GGR.String()),
				zap.String("ngr", row.NGR.String()))
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

	var registrationsQuery string
	var regArgs []interface{}

	if req.IsTestAccount != nil {
		registrationsQuery = `
			SELECT 
				u.id::text as user_id,
				COALESCE(u.username, '') as username,
				COALESCE(u.email, '') as email,
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				u.created_at::text as created_at
			FROM users u
			WHERE COALESCE(u.refered_by_code, '') != ''
				AND u.is_test_account = $1
			ORDER BY u.created_at DESC
			LIMIT 1000
		`
		regArgs = []interface{}{*req.IsTestAccount}
	} else {
		registrationsQuery = `
			SELECT 
				u.id::text as user_id,
				COALESCE(u.username, '') as username,
				COALESCE(u.email, '') as email,
				COALESCE(u.refered_by_code, 'N/A') as referral_code,
				u.created_at::text as created_at
			FROM users u
			WHERE COALESCE(u.refered_by_code, '') != ''
			ORDER BY u.created_at DESC
			LIMIT 1000
		`
		regArgs = []interface{}{}
	}

	regRows, err := r.db.GetPool().Query(ctx, registrationsQuery, regArgs...)
	if err != nil {
		r.log.Error("Failed to execute registrations query", zap.Error(err))
		summary.Registrations = []dto.AffiliateRegistration{}
	} else {
		defer regRows.Close()
		var registrations []dto.AffiliateRegistration
		for regRows.Next() {
			var reg dto.AffiliateRegistration
			err := regRows.Scan(
				&reg.UserID,
				&reg.Username,
				&reg.Email,
				&reg.ReferralCode,
				&reg.CreatedAt,
			)
			if err != nil {
				r.log.Error("Failed to scan registration row", zap.Error(err))
				continue
			}
			registrations = append(registrations, reg)
		}
		summary.Registrations = registrations
	}

	res.Data = data
	res.Summary = summary
	res.Message = "Affiliate report retrieved successfully"

	return res, nil
}

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

	// Get user IDs for this referral code
	userQueryArgs := []interface{}{*req.ReferralCode}
	userQuery := `
		SELECT DISTINCT u.id::text as user_id
		FROM users u
		WHERE COALESCE(u.refered_by_code, '') = $1
	`
	if req.IsTestAccount != nil {
		userQuery += fmt.Sprintf(" AND u.is_test_account = $%d", len(userQueryArgs)+1)
		userQueryArgs = append(userQueryArgs, *req.IsTestAccount)
	}
	userRows, err := r.db.GetPool().Query(ctx, userQuery, userQueryArgs...)
	userIDs := []string{}
	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var userID string
			if err := userRows.Scan(&userID); err == nil {
				userIDs = append(userIDs, userID)
			}
		}
	}

	// Get deposits and withdrawals from ClickHouse
	chConn, hasClickHouse := r.getClickHouseConn()
	depositsByUser := make(map[string]struct {
		DepositCount int64
		DepositsUSD  decimal.Decimal
	})
	withdrawalsByUser := make(map[string]decimal.Decimal)
	gamingMetricsByUser := make(map[string]struct {
		TotalBets      int64
		TotalStake     decimal.Decimal
		TotalWin       decimal.Decimal
		RakebackEarned decimal.Decimal
	})

	if hasClickHouse && len(userIDs) > 0 {
		userIDStrings := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			userIDStrings = append(userIDStrings, fmt.Sprintf("'%s'", userID))
		}
		userIDFilter := strings.Join(userIDStrings, ",")

		dateFromStr := dateFrom.Format("2006-01-02")
		dateToStr := dateTo.Format("2006-01-02")

		// Query deposits from ClickHouse
		depositsQuery := fmt.Sprintf(`
			SELECT 
				toString(user_id) as user_id,
				toUInt64(count()) as deposit_count,
				toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as deposits_usd
			FROM tucanbit_financial.deposits
			WHERE toString(status) = 'verified'
				AND toDate(created_at) >= '%s'
				AND toDate(created_at) <= '%s'
				AND toString(user_id) IN (%s)
			GROUP BY user_id
		`, dateFromStr, dateToStr, userIDFilter)

		depRows, err := chConn.Query(ctx, depositsQuery)
		if err == nil {
			defer depRows.Close()
			for depRows.Next() {
				var userIDStr, depositsUSDStr string
				var depositCount uint64
				if err := depRows.Scan(&userIDStr, &depositCount, &depositsUSDStr); err == nil {
					depositsUSD, _ := decimal.NewFromString(depositsUSDStr)
					depositsByUser[userIDStr] = struct {
						DepositCount int64
						DepositsUSD  decimal.Decimal
					}{
						DepositCount: int64(depositCount),
						DepositsUSD:  depositsUSD,
					}
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for deposits", zap.Error(err))
		}

		// Query withdrawals from ClickHouse
		withdrawalsQuery := fmt.Sprintf(`
			SELECT 
				toString(user_id) as user_id,
				toString(ifNull(toDecimal64(sum(usd_amount), 8), toDecimal64(0, 8))) as withdrawals_usd
			FROM tucanbit_financial.withdrawals
			WHERE toString(status) = 'completed'
				AND toDate(created_at) >= '%s'
				AND toDate(created_at) <= '%s'
				AND toString(user_id) IN (%s)
			GROUP BY user_id
		`, dateFromStr, dateToStr, userIDFilter)

		withRows, err := chConn.Query(ctx, withdrawalsQuery)
		if err == nil {
			defer withRows.Close()
			for withRows.Next() {
				var userIDStr, withdrawalsUSDStr string
				if err := withRows.Scan(&userIDStr, &withdrawalsUSDStr); err == nil {
					withdrawalsUSD, _ := decimal.NewFromString(withdrawalsUSDStr)
					withdrawalsByUser[userIDStr] = withdrawalsUSD
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for withdrawals", zap.Error(err))
		}

		// Query gaming metrics from ClickHouse (only status = 'completed')
		dateFromStrGaming := dateFrom.Format("2006-01-02 15:04:05")
		dateToStrGaming := dateTo.Format("2006-01-02 15:04:05")

		gamingQuery := fmt.Sprintf(`
			SELECT 
				toString(user_id) as user_id,
				toString(toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8)) as total_wagered,
				toString(toDecimal64(sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8)) as total_won,
				toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as total_bets,
				toString(toDecimal64(sumIf(amount, transaction_type = 'cashback_earning'), 8)) as rakeback_earned
			FROM tucanbit_analytics.transactions
			WHERE created_at >= '%s'
				AND created_at <= '%s'
				AND toString(user_id) IN (%s)
				AND toString(status) = 'completed'
			GROUP BY user_id
		`, dateFromStrGaming, dateToStrGaming, userIDFilter)

		gamingRows, err := chConn.Query(ctx, gamingQuery)
		if err == nil {
			defer gamingRows.Close()
			for gamingRows.Next() {
				var userIDStr, totalWageredStr, totalWonStr, rakebackEarnedStr string
				var totalBets uint64
				if err := gamingRows.Scan(&userIDStr, &totalWageredStr, &totalWonStr, &totalBets, &rakebackEarnedStr); err == nil {
					totalWagered, _ := decimal.NewFromString(totalWageredStr)
					totalWon, _ := decimal.NewFromString(totalWonStr)
					rakebackEarned, _ := decimal.NewFromString(rakebackEarnedStr)

					gamingMetricsByUser[userIDStr] = struct {
						TotalBets      int64
						TotalStake     decimal.Decimal
						TotalWin       decimal.Decimal
						RakebackEarned decimal.Decimal
					}{
						TotalBets:      int64(totalBets),
						TotalStake:     totalWagered,
						TotalWin:       totalWon,
						RakebackEarned: rakebackEarned,
					}
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for gaming metrics", zap.Error(err))
		}
	}

	// Query to get player-level metrics grouped by user
	query := fmt.Sprintf(`
		WITH user_deposits AS (
			SELECT 
				u.id as user_id,
				u.username,
				u.email,
				0 as deposit_count,
				0 as deposits_usd
			FROM users u
			WHERE 1=0
			GROUP BY u.id, u.username, u.email
		),
		user_withdrawals AS (
			SELECT 
				u.id as user_id,
				0 as withdrawals_usd
			FROM users u
			WHERE 1=0
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
				0 as total_bets,
				0 as total_stake,
				0 as total_win,
				0 as rakeback_earned
			FROM users u
			WHERE 1=0
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
	`, testAccountFilter, testAccountFilter)

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

		// Override deposits and withdrawals with ClickHouse data
		if chDeposits, ok := depositsByUser[row.PlayerID]; ok {
			if chDeposits.DepositCount > 0 {
				row.UniqueDepositors = 1
			}
			row.DepositsUSD = chDeposits.DepositsUSD
		}
		if chWithdrawals, ok := withdrawalsByUser[row.PlayerID]; ok {
			row.WithdrawalsUSD = chWithdrawals
		}

		// Override gaming metrics with ClickHouse data
		if chGaming, ok := gamingMetricsByUser[row.PlayerID]; ok {
			row.TotalBets = chGaming.TotalBets
			row.GGR = chGaming.TotalStake.Sub(chGaming.TotalWin)
			row.NGR = chGaming.TotalStake.Sub(chGaming.TotalWin).Sub(chGaming.RakebackEarned)
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
