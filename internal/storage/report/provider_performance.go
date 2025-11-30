package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"go.uber.org/zap"
)

// GetProviderPerformance retrieves provider-level aggregated performance metrics
func (r *report) GetProviderPerformance(ctx context.Context, req dto.ProviderPerformanceReportReq, userBrandIDs []uuid.UUID) (dto.ProviderPerformanceReportRes, error) {
	var res dto.ProviderPerformanceReportRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}
	if req.SortBy == nil {
		defaultSort := "ggr"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
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

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Provider filter (applied after provider_metrics CTE)
	providerFilter := ""
	if req.Provider != nil && *req.Provider != "" {
		providerFilter = fmt.Sprintf("AND pm.provider = $%d", argIndex)
		args = append(args, *req.Provider)
		argIndex++
	}
	// Note: Category filter is not applicable for provider performance reports

	// Build ORDER BY clause
	orderBy := "ggr DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "ggr":
			orderBy = "ggr"
		case "ngr":
			orderBy = "ngr"
		case "most_played":
			orderBy = "total_bets"
		case "rtp":
			orderBy = "effective_rtp"
		case "bet_volume":
			orderBy = "total_stake"
		default:
			orderBy = "ggr"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	// Add date params
	dateFromParam := argIndex
	dateToParam := argIndex + 1
	if dateFrom != nil {
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		args = append(args, *dateTo)
		argIndex++
	}

	// Main query - aggregates metrics by provider (similar to game performance but grouped by provider)
	query := fmt.Sprintf(`
		WITH all_game_transactions AS (
			-- GrooveTech transactions
			SELECT 
				COALESCE(g.provider, 'GrooveTech') as provider,
				gt.game_id,
				ga.user_id,
				gt.amount,
				gt.type as transaction_type,
				gt.created_at,
				gt.round_id,
				CASE WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount ELSE NULL END as win_amount,
				CASE WHEN gt.type = 'wager' THEN ABS(gt.amount) ELSE NULL END as bet_amount,
				CASE WHEN gt.type = 'result' AND gt.amount > 0 AND EXISTS (
					SELECT 1 FROM groove_transactions gt2 
					WHERE gt2.round_id = gt.round_id 
					AND gt2.type = 'wager' 
					AND gt2.amount > 0
					LIMIT 1
				) THEN gt.amount / (
					SELECT ABS(gt3.amount) FROM groove_transactions gt3 
					WHERE gt3.round_id = gt.round_id 
					AND gt3.type = 'wager' 
					LIMIT 1
				) ELSE NULL END as multiplier
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			LEFT JOIN games g ON gt.game_id = g.game_id
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s

			UNION ALL

			-- General bets (crash game)
			SELECT 
				'Crash'::text as provider,
				'Crash Game'::text as game_id,
				b.user_id,
				COALESCE(b.payout, b.amount) as amount,
				CASE WHEN COALESCE(b.payout, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				COALESCE(b.timestamp, NOW()) as created_at,
				b.round_id::text as round_id,
				COALESCE(b.payout, 0) as win_amount,
				b.amount as bet_amount,
				CASE WHEN b.amount > 0 THEN COALESCE(b.payout, 0) / b.amount ELSE NULL END as multiplier
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				AND COALESCE(b.timestamp, NOW()) >= $%d
				AND COALESCE(b.timestamp, NOW()) <= $%d
				%s
		),
		provider_metrics AS (
			SELECT 
				provider,
				COUNT(DISTINCT game_id) as total_games,
				COUNT(DISTINCT CASE WHEN transaction_type IN ('bet', 'wager') THEN round_id ELSE NULL END) as total_rounds,
				COUNT(CASE WHEN transaction_type IN ('bet', 'wager') THEN 1 ELSE NULL END) as total_bets,
				COUNT(DISTINCT user_id) as unique_players,
				COALESCE(SUM(CASE WHEN transaction_type IN ('bet', 'wager') THEN bet_amount ELSE 0 END), 0) as total_stake,
				COALESCE(SUM(CASE WHEN transaction_type = 'win' THEN win_amount ELSE 0 END), 0) as total_win,
				MAX(win_amount) as highest_win,
				MAX(multiplier) as highest_multiplier,
				COUNT(CASE WHEN win_amount > 1000 OR multiplier > 10 THEN 1 ELSE NULL END) as big_wins_count
			FROM all_game_transactions
			GROUP BY provider
		),
		user_provider_stakes AS (
			SELECT 
				provider,
				user_id,
				SUM(CASE WHEN transaction_type IN ('bet', 'wager') THEN bet_amount ELSE 0 END) as user_stake_in_provider
			FROM all_game_transactions
			GROUP BY provider, user_id
		),
		user_total_stakes AS (
			SELECT 
				user_id,
				SUM(user_stake_in_provider) as total_user_stake
			FROM user_provider_stakes
			GROUP BY user_id
		),
		user_rakeback AS (
			SELECT 
				t.user_id,
				COALESCE(SUM(t.amount), 0) as total_rakeback
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'rakeback_earned'
				AND t.created_at >= $%d
				AND t.created_at <= $%d
				%s
			GROUP BY t.user_id
		),
		provider_rakeback AS (
			SELECT 
				ups.provider,
				COALESCE(SUM(
					CASE 
						WHEN uts.total_user_stake > 0 
						THEN ur.total_rakeback * (ups.user_stake_in_provider / uts.total_user_stake)
						ELSE 0
					END
				), 0) as rakeback_earned
			FROM user_provider_stakes ups
			JOIN user_total_stakes uts ON ups.user_id = uts.user_id
			LEFT JOIN user_rakeback ur ON ups.user_id = ur.user_id
			GROUP BY ups.provider
		)
		SELECT 
			pm.provider,
			pm.total_games,
			pm.total_bets,
			pm.total_rounds,
			pm.unique_players,
			pm.total_stake,
			pm.total_win,
			pm.total_stake - pm.total_win as ggr,
			pm.total_stake - pm.total_win - COALESCE(pr.rakeback_earned, 0) as ngr,
			CASE 
				WHEN pm.total_stake > 0 THEN (pm.total_win / pm.total_stake) * 100
				ELSE 0
			END as effective_rtp,
			CASE 
				WHEN pm.total_bets > 0 THEN pm.total_stake / pm.total_bets
				ELSE 0
			END as avg_bet_amount,
			COALESCE(pr.rakeback_earned, 0) as rakeback_earned,
			pm.big_wins_count,
			COALESCE(pm.highest_win, 0) as highest_win,
			COALESCE(pm.highest_multiplier, 0) as highest_multiplier
		FROM provider_metrics pm
		LEFT JOIN provider_rakeback pr ON pm.provider = pr.provider
		WHERE 1=1 %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, whereClause, providerFilter, orderBy, argIndex, argIndex+1)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get provider performance", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get provider performance")
	}
	defer rows.Close()

	var metrics []dto.ProviderPerformanceMetric
	for rows.Next() {
		var metric dto.ProviderPerformanceMetric

		err := rows.Scan(
			&metric.Provider,
			&metric.TotalGames,
			&metric.TotalBets,
			&metric.TotalRounds,
			&metric.UniquePlayers,
			&metric.TotalStake,
			&metric.TotalWin,
			&metric.GGR,
			&metric.NGR,
			&metric.EffectiveRTP,
			&metric.AvgBetAmount,
			&metric.RakebackEarned,
			&metric.BigWinsCount,
			&metric.HighestWin,
			&metric.HighestMultiplier,
		)
		if err != nil {
			r.log.Error("failed to scan provider performance metric row", zap.Error(err))
			continue
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating provider performance rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating provider performance rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH all_game_transactions AS (
			SELECT DISTINCT
				COALESCE(g.provider, 'GrooveTech') as provider
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			LEFT JOIN games g ON gt.game_id = g.game_id
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s

			UNION

			SELECT DISTINCT
				'Crash'::text as provider
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				AND COALESCE(b.timestamp, NOW()) >= $%d
				AND COALESCE(b.timestamp, NOW()) <= $%d
				%s
		)
		SELECT COUNT(DISTINCT provider) as total
		FROM all_game_transactions
		WHERE 1=1 %s
	`, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, whereClause, providerFilter)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get provider performance count", zap.Error(err))
		total = int64(len(metrics))
	}

	// Calculate summary
	var summary dto.GamePerformanceSummary
	for _, metric := range metrics {
		summary.TotalBets += metric.TotalBets
		summary.TotalUniquePlayers += metric.UniquePlayers
		summary.TotalWagered = summary.TotalWagered.Add(metric.TotalStake)
		summary.TotalGGR = summary.TotalGGR.Add(metric.GGR)
		summary.TotalRakeback = summary.TotalRakeback.Add(metric.RakebackEarned)
	}

	// Calculate average RTP
	if summary.TotalWagered.GreaterThan(decimal.Zero) {
		totalWin := decimal.Zero
		for _, metric := range metrics {
			totalWin = totalWin.Add(metric.TotalWin)
		}
		summary.AverageRTP = totalWin.Div(summary.TotalWagered).Mul(decimal.NewFromInt(100))
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Provider performance metrics retrieved successfully"
	res.Data = metrics
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage
	res.Summary = summary

	return res, nil
}

