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

	if dateFrom == nil && dateTo == nil {
		now := time.Now()
		dateTo = &now
		thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
		dateFrom = &thirtyDaysAgo
	}

	if dateFrom == nil || dateTo == nil {
		return res, errors.ErrInvalidUserInput.New("date_from and date_to are required")
	}

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

	// Currency filter
	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
		argIndex++
	}

	// Test account filter
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

	// Query ClickHouse for gaming metrics grouped by game_id
	chConn, hasClickHouse := r.getClickHouseConn()
	gamingMetricsMap := make(map[string]struct {
		TotalWagered      decimal.Decimal
		TotalWon          decimal.Decimal
		TotalBets         int64
		TotalRounds       int64
		UniquePlayers     int64
		HighestWin        decimal.Decimal
		HighestMultiplier decimal.Decimal
		BigWinsCount      int64
	})

	if hasClickHouse {
		startOfDay := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)
		startDateStr := startOfDay.Format("2006-01-02 15:04:05")
		endDateStr := endOfDay.Format("2006-01-02 15:04:05")

		// Build user filter for ClickHouse
		userFilter := ""
		chArgs := []interface{}{}
		if len(whereConditions) > 0 {
			userIDsQuery := fmt.Sprintf(`
				SELECT DISTINCT u.id
				FROM users u
				%s
			`, whereClause)

			var userIDs []uuid.UUID
			rows, err := r.db.GetPool().Query(ctx, userIDsQuery, args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var userID uuid.UUID
					if err := rows.Scan(&userID); err == nil {
						userIDs = append(userIDs, userID)
					}
				}
			}

			if len(userIDs) > 0 {
				placeholders := make([]string, len(userIDs))
				for i := range userIDs {
					placeholders[i] = "?"
					chArgs = append(chArgs, userIDs[i].String())
				}
				userFilter = " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
			}
		}

		chQuery := fmt.Sprintf(`
			SELECT 
				game_id,
				toDecimal64(sumIf(amount, transaction_type IN ('bet', 'groove_bet')), 8) as total_wagered,
				toDecimal64(sumIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8) as total_won,
				toUInt64(countIf(transaction_type IN ('bet', 'groove_bet'))) as total_bets,
				toUInt64(uniqExact(session_id)) as total_rounds,
				toUInt64(uniqExact(user_id)) as unique_players,
				toDecimal64(maxIf(COALESCE(win_amount, amount), transaction_type IN ('win', 'groove_win')), 8) as highest_win,
				toDecimal64(maxIf(
					CASE 
						WHEN transaction_type IN ('bet', 'groove_bet') AND amount > 0 THEN COALESCE(win_amount, amount) / amount
						ELSE 0
					END,
					transaction_type IN ('win', 'groove_win')
				), 8) as highest_multiplier,
				toUInt64(countIf(
					(transaction_type IN ('groove_bet', 'groove_win') AND win_amount IS NOT NULL AND win_amount > 1000) OR 
					(transaction_type IN ('win', 'groove_win') AND COALESCE(win_amount, amount) > 1000)
				)) as big_wins_count
			FROM tucanbit_analytics.transactions
			WHERE game_id IS NOT NULL 
				AND game_id != '' 
				AND amount > 0
				AND transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
				AND created_at >= '%s' AND created_at <= '%s'
				%s
			GROUP BY game_id
		`, startDateStr, endDateStr, userFilter)

		rows, err := chConn.Query(ctx, chQuery, chArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var gameIDStr string
				var totalWageredStr, totalWonStr string
				var totalBets, totalRounds, uniquePlayers, bigWinsCount uint64
				var highestWinStr, highestMultiplierStr string

				if err := rows.Scan(&gameIDStr, &totalWageredStr, &totalWonStr, &totalBets, &totalRounds, &uniquePlayers, &highestWinStr, &highestMultiplierStr, &bigWinsCount); err == nil {
					if gameIDStr == "" {
						continue
					}
					totalWagered, _ := decimal.NewFromString(totalWageredStr)
					totalWon, _ := decimal.NewFromString(totalWonStr)
					highestWin, _ := decimal.NewFromString(highestWinStr)
					highestMultiplier, _ := decimal.NewFromString(highestMultiplierStr)

					gamingMetricsMap[gameIDStr] = struct {
						TotalWagered      decimal.Decimal
						TotalWon          decimal.Decimal
						TotalBets         int64
						TotalRounds       int64
						UniquePlayers     int64
						HighestWin        decimal.Decimal
						HighestMultiplier decimal.Decimal
						BigWinsCount      int64
					}{
						TotalWagered:      totalWagered,
						TotalWon:          totalWon,
						TotalBets:         int64(totalBets),
						TotalRounds:       int64(totalRounds),
						UniquePlayers:     int64(uniquePlayers),
						HighestWin:        highestWin,
						HighestMultiplier: highestMultiplier,
						BigWinsCount:      int64(bigWinsCount),
					}
				}
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for gaming metrics", zap.Error(err))
		}
	}

	gameIDsFromClickHouse := make([]string, 0, len(gamingMetricsMap))
	for gameID := range gamingMetricsMap {
		if gameID != "" {
			gameIDsFromClickHouse = append(gameIDsFromClickHouse, gameID)
		}
	}

	gameDetailsQuery := `
		SELECT DISTINCT
			g.game_id,
			COALESCE(g.name, g.game_id, 'Unknown') as game_name,
			COALESCE(g.provider, 'Unknown') as provider,
			g.category,
			COALESCE(he.house_edge, 0) as house_edge,
			CASE 
				WHEN he.house_edge IS NOT NULL THEN 100 - he.house_edge
				ELSE NULL
			END as rtp
		FROM games g
		LEFT JOIN LATERAL (
			SELECT house_edge
			FROM game_house_edges
			WHERE is_active = true 
				AND game_id = g.game_id
				AND (effective_from IS NULL OR effective_from <= NOW())
				AND (effective_until IS NULL OR effective_until >= NOW())
			ORDER BY effective_from DESC, updated_at DESC
			LIMIT 1
		) he ON true
		WHERE g.game_id IS NOT NULL
	`

	gameDetailsArgs := []interface{}{}
	if len(gameIDsFromClickHouse) > 0 {
		gamePlaceholders := make([]string, len(gameIDsFromClickHouse))
		for i, gameID := range gameIDsFromClickHouse {
			gamePlaceholders[i] = fmt.Sprintf("$%d", i+1)
			gameDetailsArgs = append(gameDetailsArgs, gameID)
		}
		gameDetailsQuery += fmt.Sprintf(" AND g.game_id IN (%s)", strings.Join(gamePlaceholders, ","))
	} else {
		gameDetailsQuery += " AND 1=0"
	}

	// Add filters for game_id, provider, category
	gameFilters := []string{}
	gameArgIndex := len(gameDetailsArgs) + 1

	if req.GameID != nil && *req.GameID != "" {
		gameFilters = append(gameFilters, fmt.Sprintf("g.game_id = $%d", gameArgIndex))
		gameDetailsArgs = append(gameDetailsArgs, *req.GameID)
		gameArgIndex++
	}

	if req.GameProvider != nil && *req.GameProvider != "" {
		gameFilters = append(gameFilters, fmt.Sprintf("g.provider = $%d", gameArgIndex))
		gameDetailsArgs = append(gameDetailsArgs, *req.GameProvider)
		gameArgIndex++
	}

	if req.Category != nil && *req.Category != "" {
		gameFilters = append(gameFilters, fmt.Sprintf("g.category = $%d", gameArgIndex))
		gameDetailsArgs = append(gameDetailsArgs, *req.Category)
		gameArgIndex++
	}

	if len(gameFilters) > 0 {
		gameDetailsQuery += " AND " + strings.Join(gameFilters, " AND ")
	}

	gameDetailsMap := make(map[string]struct {
		GameName  string
		Provider  string
		Category  *string
		RTP       *decimal.Decimal
		HouseEdge decimal.Decimal
	})

	if len(gameIDsFromClickHouse) > 0 {
		rows, err := r.db.GetPool().Query(ctx, gameDetailsQuery, gameDetailsArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var gameID, gameName, provider string
				var category sql.NullString
				var houseEdge decimal.Decimal
				var rtp sql.NullString

				if err := rows.Scan(&gameID, &gameName, &provider, &category, &houseEdge, &rtp); err == nil {
					var rtpVal *decimal.Decimal
					if rtp.Valid && rtp.String != "" {
						if parsed, err := decimal.NewFromString(rtp.String); err == nil {
							rtpVal = &parsed
						}
					}

					var categoryVal *string
					if category.Valid {
						categoryVal = &category.String
					}

					gameDetailsMap[gameID] = struct {
						GameName  string
						Provider  string
						Category  *string
						RTP       *decimal.Decimal
						HouseEdge decimal.Decimal
					}{
						GameName:  gameName,
						Provider:  provider,
						Category:  categoryVal,
						RTP:       rtpVal,
						HouseEdge: houseEdge,
					}
				}
			}
		} else {
			r.log.Warn("Failed to query PostgreSQL for game details", zap.Error(err))
		}
	}

	rakebackQuery := fmt.Sprintf(`
		WITH user_game_stakes AS (
			SELECT 
				gt.game_id,
				ga.user_id,
				SUM(CASE WHEN gt.type = 'wager' THEN ABS(gt.amount) ELSE 0 END) as user_stake_in_game
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s
			GROUP BY gt.game_id, ga.user_id
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
		SELECT game_id, rakeback_earned
		FROM game_rakeback
	`, argIndex, argIndex+1, whereClause, argIndex+2, argIndex+3, whereClause)

	rakebackArgs := append(args, *dateFrom, *dateTo, *dateFrom, *dateTo)
	rakebackMap := make(map[string]decimal.Decimal)

	rakebackRows, err := r.db.GetPool().Query(ctx, rakebackQuery, rakebackArgs...)
	if err == nil {
		defer rakebackRows.Close()
		for rakebackRows.Next() {
			var gameID string
			var rakeback decimal.Decimal
			if err := rakebackRows.Scan(&gameID, &rakeback); err == nil {
				rakebackMap[gameID] = rakeback
			}
		}
	} else {
		r.log.Warn("Failed to query PostgreSQL for rakeback", zap.Error(err))
	}

	var metrics []dto.GamePerformanceMetric
	for gameID, gamingMetrics := range gamingMetricsMap {
		gameDetails, hasDetails := gameDetailsMap[gameID]
		if !hasDetails {
			gameDetails = struct {
				GameName  string
				Provider  string
				Category  *string
				RTP       *decimal.Decimal
				HouseEdge decimal.Decimal
			}{
				GameName:  gameID,
				Provider:  "Unknown",
				Category:  nil,
				RTP:       nil,
				HouseEdge: decimal.Zero,
			}
		}

		rakeback := rakebackMap[gameID]

		metric := dto.GamePerformanceMetric{
			GameID:            gameID,
			GameName:          gameDetails.GameName,
			Provider:          gameDetails.Provider,
			Category:          gameDetails.Category,
			TotalBets:         gamingMetrics.TotalBets,
			TotalRounds:       gamingMetrics.TotalRounds,
			UniquePlayers:     gamingMetrics.UniquePlayers,
			TotalStake:        gamingMetrics.TotalWagered,
			TotalWin:          gamingMetrics.TotalWon,
			GGR:               gamingMetrics.TotalWagered.Sub(gamingMetrics.TotalWon),
			NGR:               decimal.Zero, // Management only wants GGR, not NGR
			EffectiveRTP:      decimal.Zero,
			AvgBetAmount:      decimal.Zero,
			RakebackEarned:    rakeback,
			BigWinsCount:      gamingMetrics.BigWinsCount,
			HighestWin:        gamingMetrics.HighestWin,
			HighestMultiplier: gamingMetrics.HighestMultiplier,
		}

		if gameDetails.RTP != nil {
			metric.EffectiveRTP = *gameDetails.RTP
		} else if gamingMetrics.TotalWagered.GreaterThan(decimal.Zero) {
			metric.EffectiveRTP = gamingMetrics.TotalWon.Div(gamingMetrics.TotalWagered).Mul(decimal.NewFromInt(100))
		}

		if gamingMetrics.TotalBets > 0 {
			metric.AvgBetAmount = gamingMetrics.TotalWagered.Div(decimal.NewFromInt(gamingMetrics.TotalBets))
		}

		metrics = append(metrics, metric)
	}

	// Sort metrics
	r.sortGamePerformanceMetrics(metrics, req.SortBy, req.SortOrder)

	// Apply pagination
	total := int64(len(metrics))
	startIdx := (req.Page - 1) * req.PerPage
	endIdx := startIdx + req.PerPage
	if startIdx > len(metrics) {
		metrics = []dto.GamePerformanceMetric{}
	} else if endIdx > len(metrics) {
		metrics = metrics[startIdx:]
	} else {
		metrics = metrics[startIdx:endIdx]
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

func (r *report) sortGamePerformanceMetrics(metrics []dto.GamePerformanceMetric, sortBy *string, sortOrder *string) {
	if len(metrics) == 0 {
		return
	}

	order := "desc"
	if sortOrder != nil {
		order = strings.ToLower(*sortOrder)
	}

	sortField := "ggr"
	if sortBy != nil {
		sortField = strings.ToLower(*sortBy)
	}

	sort.Slice(metrics, func(i, j int) bool {
		var less bool
		switch sortField {
		case "ggr":
			less = metrics[i].GGR.LessThan(metrics[j].GGR)
		case "ngr":
			less = metrics[i].NGR.LessThan(metrics[j].NGR)
		case "most_played":
			less = metrics[i].TotalBets < metrics[j].TotalBets
		case "rtp":
			less = metrics[i].EffectiveRTP.LessThan(metrics[j].EffectiveRTP)
		case "bet_volume":
			less = metrics[i].TotalStake.LessThan(metrics[j].TotalStake)
		default:
			less = metrics[i].GGR.LessThan(metrics[j].GGR)
		}
		if order == "asc" {
			return less
		}
		return !less
	})
}

func (r *report) GetGamePlayers(ctx context.Context, req dto.GamePlayersReq, userBrandIDs []uuid.UUID) (dto.GamePlayersRes, error) {
	var res dto.GamePlayersRes

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

	whereConditions := []string{}
	args := []interface{}{req.GameID}
	argIndex := 2
	if len(userBrandIDs) > 0 {
		brandPlaceholders := []string{}
		for _, brandID := range userBrandIDs {
			brandPlaceholders = append(brandPlaceholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, brandID)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(u.brand_id IS NULL OR u.brand_id IN (%s))", strings.Join(brandPlaceholders, ",")))
	}

	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
		argIndex++
	}

	if req.BetType != nil && *req.BetType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("bet_type = $%d", argIndex))
		args = append(args, *req.BetType)
		argIndex++
	}

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
		havingConditions = append(havingConditions, fmt.Sprintf("ggr >= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MinNetResult))
		argIndex++
	}
	if req.MaxNetResult != nil {
		havingConditions = append(havingConditions, fmt.Sprintf("ggr <= $%d", argIndex))
		args = append(args, decimal.NewFromFloat(*req.MaxNetResult))
		argIndex++
	}

	havingClause := ""
	if len(havingConditions) > 0 {
		havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
	}

	orderBy := "total_stake DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "total_stake":
			orderBy = "total_stake"
		case "total_win":
			orderBy = "total_win"
		case "ngr":
			orderBy = "ggr"
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

	query := fmt.Sprintf(`
		WITH all_game_transactions AS (
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

			SELECT 
				b.user_id,
				'Crash Game'::text as game_id,
				COALESCE(b.payout, b.amount) as amount,
				CASE WHEN COALESCE(b.payout, 0) > 0 THEN 'win' ELSE 'bet' END as transaction_type,
				COALESCE(b.timestamp, NOW()) as transaction_date,
				b.round_id::text as round_id,
				b.currency,
				'cash' as bet_type,
				COALESCE(b.payout, 0) as win_amount,
				b.amount as bet_amount
			FROM bets b
			JOIN users u ON b.user_id = u.id
			WHERE 'Crash Game' = $1
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
			pm.total_stake - pm.total_win as ggr,
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

	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

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

		var ggr decimal.Decimal
		err := rows.Scan(
			&player.PlayerID,
			&player.Username,
			&email,
			&player.TotalStake,
			&player.TotalWin,
			&ggr,
			&player.Rakeback,
			&player.NumberOfRounds,
			&player.LastPlayed,
			&player.Currency,
		)
		player.NGR = decimal.Zero
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
			WHERE 'Crash Game' = $1
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
