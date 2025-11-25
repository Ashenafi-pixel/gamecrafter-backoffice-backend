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

// GetGamePerformance retrieves game-level aggregated performance metrics
func (r *report) GetGamePerformance(ctx context.Context, req dto.GamePerformanceReportReq, userBrandIDs []uuid.UUID) (dto.GamePerformanceReportRes, error) {
	var res dto.GamePerformanceReportRes

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

	// Game Provider filter
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

	// Category filter
	if req.Category != nil && *req.Category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *req.Category)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

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

	// Main query - aggregates metrics by game from multiple sources
	query := fmt.Sprintf(`
		WITH all_game_transactions AS (
			-- GrooveTech transactions
			SELECT 
				gt.game_id,
				COALESCE(gt.game_id, 'Unknown') as game_name,
				COALESCE(gt.game_id, 'GrooveTech') as provider,
				NULL::text as category,
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
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s

			UNION ALL

			-- General bets
			SELECT 
				COALESCE(b.game_id, b.game_name, 'Unknown') as game_id,
				COALESCE(b.game_name, b.game_id, 'Unknown') as game_name,
				COALESCE(b.provider, 'Unknown') as provider,
				NULL::text as category,
				b.user_id,
				COALESCE(b.payout, b.amount) as amount,
				CASE WHEN COALESCE(b.payout, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				COALESCE(b.timestamp, b.created_at, NOW()) as created_at,
				b.round_id::text as round_id,
				COALESCE(b.payout, 0) as win_amount,
				b.amount as bet_amount,
				CASE WHEN b.amount > 0 THEN COALESCE(b.payout, 0) / b.amount ELSE NULL END as multiplier
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				AND COALESCE(b.timestamp, b.created_at, NOW()) >= $%d
				AND COALESCE(b.timestamp, b.created_at, NOW()) <= $%d
				%s
		),
		game_metrics AS (
			SELECT 
				game_id,
				MAX(game_name) as game_name,
				MAX(provider) as provider,
				MAX(category) as category,
				COUNT(DISTINCT CASE WHEN transaction_type IN ('bet', 'wager') THEN round_id ELSE NULL END) as total_rounds,
				COUNT(CASE WHEN transaction_type IN ('bet', 'wager') THEN 1 ELSE NULL END) as total_bets,
				COUNT(DISTINCT user_id) as unique_players,
				COALESCE(SUM(CASE WHEN transaction_type IN ('bet', 'wager') THEN bet_amount ELSE 0 END), 0) as total_stake,
				COALESCE(SUM(CASE WHEN transaction_type = 'win' THEN win_amount ELSE 0 END), 0) as total_win,
				MAX(win_amount) as highest_win,
				MAX(multiplier) as highest_multiplier,
				COUNT(CASE WHEN win_amount > 1000 OR multiplier > 10 THEN 1 ELSE NULL END) as big_wins_count
			FROM all_game_transactions
			GROUP BY game_id
		),
		user_game_stakes AS (
			SELECT 
				game_id,
				user_id,
				SUM(CASE WHEN transaction_type IN ('bet', 'wager') THEN bet_amount ELSE 0 END) as user_stake_in_game
			FROM all_game_transactions
			GROUP BY game_id, user_id
		),
		user_total_stakes AS (
			SELECT 
				user_id,
				SUM(user_stake_in_game) as total_user_stake
			FROM user_game_stakes
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
		game_rakeback AS (
			SELECT 
				ugs.game_id,
				COALESCE(SUM(
					CASE 
						WHEN uts.total_user_stake > 0 
						THEN ur.total_rakeback * (ugs.user_stake_in_game / uts.total_user_stake)
						ELSE 0
					END
				), 0) as rakeback_earned
			FROM user_game_stakes ugs
			JOIN user_total_stakes uts ON ugs.user_id = uts.user_id
			LEFT JOIN user_rakeback ur ON ugs.user_id = ur.user_id
			GROUP BY ugs.game_id
		)
		SELECT 
			gm.game_id,
			gm.game_name,
			gm.provider,
			gm.category,
			gm.total_bets,
			gm.total_rounds,
			gm.unique_players,
			gm.total_stake,
			gm.total_win,
			gm.total_stake - gm.total_win as ggr,
			gm.total_stake - gm.total_win - COALESCE(gr.rakeback_earned, 0) as ngr,
			CASE 
				WHEN gm.total_stake > 0 THEN (gm.total_win / gm.total_stake) * 100
				ELSE 0
			END as effective_rtp,
			CASE 
				WHEN gm.total_bets > 0 THEN gm.total_stake / gm.total_bets
				ELSE 0
			END as avg_bet_amount,
			COALESCE(gr.rakeback_earned, 0) as rakeback_earned,
			gm.big_wins_count,
			COALESCE(gm.highest_win, 0) as highest_win,
			COALESCE(gm.highest_multiplier, 0) as highest_multiplier
		FROM game_metrics gm
		LEFT JOIN game_rakeback gr ON gm.game_id = gr.game_id
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, dateFromParam, dateToParam, whereClause, orderBy, argIndex+1, argIndex+2)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get game performance", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get game performance")
	}
	defer rows.Close()

	var metrics []dto.GamePerformanceMetric
	for rows.Next() {
		var metric dto.GamePerformanceMetric
		var category sql.NullString

		err := rows.Scan(
			&metric.GameID,
			&metric.GameName,
			&metric.Provider,
			&category,
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
			r.log.Error("failed to scan game performance metric row", zap.Error(err))
			continue
		}

		if category.Valid {
			metric.Category = &category.String
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating game performance rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating game performance rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH all_game_transactions AS (
			SELECT DISTINCT
				COALESCE(gt.game_id, 'Unknown') as game_id
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s

			UNION

			SELECT DISTINCT
				COALESCE(b.game_id, b.game_name, 'Unknown') as game_id
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				AND COALESCE(b.timestamp, b.created_at, NOW()) >= $%d
				AND COALESCE(b.timestamp, b.created_at, NOW()) <= $%d
				%s
		)
		SELECT COUNT(DISTINCT game_id) as total
		FROM all_game_transactions
	`, dateFromParam, dateToParam, whereClause, dateFromParam, dateToParam, whereClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get game performance count", zap.Error(err))
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

	res.Message = "Game performance metrics retrieved successfully"
	res.Data = metrics
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage
	res.Summary = summary

	return res, nil
}

// GetGamePlayers retrieves players for a specific game (drill-down)
func (r *report) GetGamePlayers(ctx context.Context, req dto.GamePlayersReq, userBrandIDs []uuid.UUID) (dto.GamePlayersRes, error) {
	var res dto.GamePlayersRes

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 50
	}
	if req.SortBy == nil {
		defaultSort := "total_stake"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}

	// Parse date ranges
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

	// Build WHERE conditions
	whereConditions := []string{}
	args := []interface{}{req.GameID}
	argIndex := 2

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

	// Currency filter
	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
		argIndex++
	}

	// Bet type filter
	if req.BetType != nil && *req.BetType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("bet_type = $%d", argIndex))
		args = append(args, *req.BetType)
		argIndex++
	}

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

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build HAVING clause for stake and net result filters
	havingConditions := []string{}
	if req.MinStake != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("total_stake >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinStake))
		argIndex++
	}
	if req.MaxStake != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("total_stake <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxStake))
		argIndex++
	}
	if req.MinNetResult != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("ngr >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinNetResult))
		argIndex++
	}
	if req.MaxNetResult != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("ngr <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxNetResult))
		argIndex++
	}

	havingClause := ""
	if len(havingConditions) > 0 {
		havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "total_stake DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "total_stake":
			orderBy = "total_stake"
		case "total_win":
			orderBy = "total_win"
		case "ngr":
			orderBy = "ngr"
		case "rounds":
			orderBy = "number_of_rounds"
		case "last_played":
			orderBy = "last_played"
		default:
			orderBy = "total_stake"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	// Main query
	query := fmt.Sprintf(`
		WITH all_game_transactions AS (
			-- GrooveTech transactions
			SELECT 
				ga.user_id,
				gt.game_id,
				gt.amount,
				gt.type as transaction_type,
				gt.created_at as transaction_date,
				gt.round_id,
				COALESCE(ga.currency, 'USD') as currency,
				'cash' as bet_type,
				CASE WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount ELSE NULL END as win_amount,
				CASE WHEN gt.type = 'wager' THEN ABS(gt.amount) ELSE NULL END as bet_amount
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			WHERE gt.game_id = $1
				AND gt.type IN ('wager', 'result')
				%s

			UNION ALL

			-- General bets
			SELECT 
				b.user_id,
				COALESCE(b.game_id, b.game_name, 'Unknown') as game_id,
				COALESCE(b.payout, b.amount) as amount,
				CASE WHEN COALESCE(b.payout, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				COALESCE(b.timestamp, b.created_at, NOW()) as transaction_date,
				b.round_id::text as round_id,
				b.currency,
				'cash' as bet_type,
				COALESCE(b.payout, 0) as win_amount,
				b.amount as bet_amount
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE COALESCE(b.game_id, b.game_name, 'Unknown') = $1
				AND (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				%s
		),
		player_metrics AS (
			SELECT 
				agt.user_id,
				MAX(agt.currency) as currency,
				COUNT(DISTINCT agt.round_id) as number_of_rounds,
				COALESCE(SUM(CASE WHEN agt.transaction_type IN ('bet', 'wager') THEN agt.bet_amount ELSE 0 END), 0) as total_stake,
				COALESCE(SUM(CASE WHEN agt.transaction_type = 'win' THEN agt.win_amount ELSE 0 END), 0) as total_win,
				MAX(agt.transaction_date) as last_played
			FROM all_game_transactions agt
			GROUP BY agt.user_id
		),
		player_rakeback AS (
			SELECT 
				t.user_id,
				COALESCE(SUM(CASE WHEN t.transaction_type = 'rakeback_earned' THEN t.amount ELSE 0 END), 0) as rakeback
			FROM transactions t
			JOIN users u ON t.user_id = u.id
			WHERE t.transaction_type = 'rakeback_earned'
				%s
			GROUP BY t.user_id
		)
		SELECT 
			u.id as player_id,
			u.username,
			u.email,
			pm.total_stake,
			pm.total_win,
			pm.total_stake - pm.total_win - COALESCE(pr.rakeback, 0) as ngr,
			COALESCE(pr.rakeback, 0) as rakeback,
			pm.number_of_rounds,
			pm.last_played,
			pm.currency
		FROM player_metrics pm
		JOIN users u ON pm.user_id = u.id
		LEFT JOIN player_rakeback pr ON pm.user_id = pr.user_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, whereClause, whereClause, havingClause, orderBy, argIndex+1, argIndex+2)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get game players", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get game players")
	}
	defer rows.Close()

	var players []dto.GamePlayer
	for rows.Next() {
		var player dto.GamePlayer
		var email sql.NullString

		err := rows.Scan(
			&player.PlayerID,
			&player.Username,
			&email,
			&player.TotalStake,
			&player.TotalWin,
			&player.NGR,
			&player.Rakeback,
			&player.NumberOfRounds,
			&player.LastPlayed,
			&player.Currency,
		)
		if err != nil {
			r.log.Error("failed to scan game player row", zap.Error(err))
			continue
		}

		if email.Valid {
			player.Email = &email.String
		}

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating game players rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating game players rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH all_game_transactions AS (
			SELECT DISTINCT ga.user_id
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			WHERE gt.game_id = $1
				AND gt.type IN ('wager', 'result')
				%s

			UNION

			SELECT DISTINCT b.user_id
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE COALESCE(b.game_id, b.game_name, 'Unknown') = $1
				AND (COALESCE(b.payout, 0) > 0 OR b.amount > 0)
				%s
		)
		SELECT COUNT(DISTINCT user_id) as total
		FROM all_game_transactions
	`, whereClause, whereClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get game players count", zap.Error(err))
		total = int64(len(players))
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Game players retrieved successfully"
	res.Data = players
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage

	return res, nil
}

