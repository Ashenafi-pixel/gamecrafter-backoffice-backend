package report

import (
	"context"
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

func (r *report) GetProviderPerformance(ctx context.Context, req dto.ProviderPerformanceReportReq, userBrandIDs []uuid.UUID) (dto.ProviderPerformanceReportRes, error) {
	var res dto.ProviderPerformanceReportRes

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
		if parsed, err := time.Parse("2006-01-02", *req.DateFrom); err == nil {
			dateFrom = &parsed
		}
	}
	if req.DateTo != nil && *req.DateTo != "" {
		if parsed, err := time.Parse("2006-01-02", *req.DateTo); err == nil {
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

	if req.BrandID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.brand_id = $%d", argIndex))
		args = append(args, *req.BrandID)
		argIndex++
	}

	if req.Currency != nil && *req.Currency != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(u.default_currency, 'USD') = $%d", argIndex))
		args = append(args, *req.Currency)
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

	chConn, hasClickHouse := r.getClickHouseConn()
	gamingMetricsByGame := make(map[string]struct {
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

		userFilter := ""
		chArgs := []interface{}{}
		if len(userIDs) > 0 {
			placeholders := make([]string, len(userIDs))
			for i := range userIDs {
				placeholders[i] = "?"
				chArgs = append(chArgs, userIDs[i].String())
			}
			userFilter = " AND user_id IN (" + strings.Join(placeholders, ",") + ")"
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

		chRows, err := chConn.Query(ctx, chQuery, chArgs...)
		if err == nil {
			defer chRows.Close()
			for chRows.Next() {
				var gameIDStr string
				var totalWageredStr, totalWonStr string
				var totalBets, totalRounds, uniquePlayers, bigWinsCount uint64
				var highestWinStr, highestMultiplierStr string

				if err := chRows.Scan(&gameIDStr, &totalWageredStr, &totalWonStr, &totalBets, &totalRounds, &uniquePlayers, &highestWinStr, &highestMultiplierStr, &bigWinsCount); err == nil {
					if gameIDStr == "" {
						continue
					}
					totalWagered, _ := decimal.NewFromString(totalWageredStr)
					totalWon, _ := decimal.NewFromString(totalWonStr)
					highestWin, _ := decimal.NewFromString(highestWinStr)
					highestMultiplier, _ := decimal.NewFromString(highestMultiplierStr)

					gamingMetricsByGame[gameIDStr] = struct {
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

	gameDetailsQuery := `
		SELECT DISTINCT
			g.game_id,
			COALESCE(g.provider, 'Unknown') as provider,
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

	gameFilters := []string{}
	gameArgs := []interface{}{}
	gameArgIndex := 1

	if req.Provider != nil && *req.Provider != "" {
		gameFilters = append(gameFilters, fmt.Sprintf("g.provider = $%d", gameArgIndex))
		gameArgs = append(gameArgs, *req.Provider)
		gameArgIndex++
	}

	if len(gameFilters) > 0 {
		gameDetailsQuery += " AND " + strings.Join(gameFilters, " AND ")
	}

	gameDetailsMap := make(map[string]struct {
		Provider  string
		RTP       *decimal.Decimal
		HouseEdge decimal.Decimal
	})

	rows, err := r.db.GetPool().Query(ctx, gameDetailsQuery, gameArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var gameID, provider string
			var houseEdge decimal.Decimal
			var rtp interface{}

			if err := rows.Scan(&gameID, &provider, &houseEdge, &rtp); err == nil {
				var rtpVal *decimal.Decimal
				if rtp != nil {
					if rtpStr, ok := rtp.(string); ok && rtpStr != "" {
						if parsed, err := decimal.NewFromString(rtpStr); err == nil {
							rtpVal = &parsed
						}
					} else if rtpDec, ok := rtp.(decimal.Decimal); ok {
						rtpVal = &rtpDec
					}
				}

				gameDetailsMap[gameID] = struct {
					Provider  string
					RTP       *decimal.Decimal
					HouseEdge decimal.Decimal
				}{
					Provider:  provider,
					RTP:       rtpVal,
					HouseEdge: houseEdge,
				}
			}
		}
	} else {
		r.log.Warn("Failed to query PostgreSQL for game details", zap.Error(err))
	}

	providerMetricsMap := make(map[string]struct {
		TotalGames        int64
		TotalBets         int64
		TotalRounds       int64
		UniquePlayersSum  int64
		TotalStake        decimal.Decimal
		TotalWin          decimal.Decimal
		HighestWin        decimal.Decimal
		HighestMultiplier decimal.Decimal
		BigWinsCount      int64
		GameRTPData       []struct {
			RTP     decimal.Decimal
			Wagered decimal.Decimal
		}
		TotalWageredForRTP decimal.Decimal
	})

	for gameID, gamingMetrics := range gamingMetricsByGame {
		gameDetails, hasDetails := gameDetailsMap[gameID]
		if !hasDetails {
			gameDetails = struct {
				Provider  string
				RTP       *decimal.Decimal
				HouseEdge decimal.Decimal
			}{
				Provider:  "Unknown",
				RTP:       nil,
				HouseEdge: decimal.Zero,
			}
		}

		provider := gameDetails.Provider
		if provider == "" {
			provider = "Unknown"
		}

		pm, exists := providerMetricsMap[provider]
		if !exists {
			pm = struct {
				TotalGames        int64
				TotalBets         int64
				TotalRounds       int64
				UniquePlayersSum  int64
				TotalStake        decimal.Decimal
				TotalWin          decimal.Decimal
				HighestWin        decimal.Decimal
				HighestMultiplier decimal.Decimal
				BigWinsCount      int64
				GameRTPData       []struct {
					RTP     decimal.Decimal
					Wagered decimal.Decimal
				}
				TotalWageredForRTP decimal.Decimal
			}{}
		}

		pm.TotalGames++
		pm.TotalBets += gamingMetrics.TotalBets
		pm.TotalRounds += gamingMetrics.TotalRounds
		pm.UniquePlayersSum += gamingMetrics.UniquePlayers
		pm.TotalStake = pm.TotalStake.Add(gamingMetrics.TotalWagered)
		pm.TotalWin = pm.TotalWin.Add(gamingMetrics.TotalWon)
		if gamingMetrics.HighestWin.GreaterThan(pm.HighestWin) {
			pm.HighestWin = gamingMetrics.HighestWin
		}
		if gamingMetrics.HighestMultiplier.GreaterThan(pm.HighestMultiplier) {
			pm.HighestMultiplier = gamingMetrics.HighestMultiplier
		}
		pm.BigWinsCount += gamingMetrics.BigWinsCount
		if gameDetails.RTP != nil {
			pm.GameRTPData = append(pm.GameRTPData, struct {
				RTP     decimal.Decimal
				Wagered decimal.Decimal
			}{
				RTP:     *gameDetails.RTP,
				Wagered: gamingMetrics.TotalWagered,
			})
			pm.TotalWageredForRTP = pm.TotalWageredForRTP.Add(gamingMetrics.TotalWagered)
		}

		providerMetricsMap[provider] = pm
	}

	rakebackQuery := fmt.Sprintf(`
		WITH user_provider_stakes AS (
			SELECT 
				COALESCE(g.provider, 'Unknown') as provider,
				ga.user_id,
				SUM(CASE WHEN gt.type = 'wager' THEN ABS(gt.amount) ELSE 0 END) as user_stake_in_provider
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			LEFT JOIN games g ON gt.game_id = g.game_id
			WHERE gt.type IN ('wager', 'result')
				AND gt.created_at >= $%d
				AND gt.created_at <= $%d
				%s
			GROUP BY COALESCE(g.provider, 'Unknown'), ga.user_id
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
		SELECT provider, rakeback_earned
		FROM provider_rakeback
	`, argIndex, argIndex+1, whereClause, argIndex+2, argIndex+3, whereClause)

	rakebackArgs := append(args, *dateFrom, *dateTo, *dateFrom, *dateTo)
	rakebackMap := make(map[string]decimal.Decimal)

	rakebackRows, err := r.db.GetPool().Query(ctx, rakebackQuery, rakebackArgs...)
	if err == nil {
		defer rakebackRows.Close()
		for rakebackRows.Next() {
			var provider string
			var rakeback decimal.Decimal
			if err := rakebackRows.Scan(&provider, &rakeback); err == nil {
				rakebackMap[provider] = rakeback
			}
		}
	} else {
		r.log.Warn("Failed to query PostgreSQL for rakeback", zap.Error(err))
	}

	var metrics []dto.ProviderPerformanceMetric
	for provider, pm := range providerMetricsMap {
		rakeback := rakebackMap[provider]

		effectiveRTP := decimal.Zero
		if len(pm.GameRTPData) > 0 {
			if pm.TotalWageredForRTP.GreaterThan(decimal.Zero) {
				totalWeightedRTP := decimal.Zero
				for _, gameData := range pm.GameRTPData {
					weightedContribution := gameData.RTP.Mul(gameData.Wagered)
					totalWeightedRTP = totalWeightedRTP.Add(weightedContribution)
				}
				effectiveRTP = totalWeightedRTP.Div(pm.TotalWageredForRTP)
			} else {
				sumRTP := decimal.Zero
				for _, gameData := range pm.GameRTPData {
					sumRTP = sumRTP.Add(gameData.RTP)
				}
				effectiveRTP = sumRTP.Div(decimal.NewFromInt(int64(len(pm.GameRTPData))))
			}
		} else if pm.TotalStake.GreaterThan(decimal.Zero) {
			effectiveRTP = pm.TotalWin.Div(pm.TotalStake).Mul(decimal.NewFromInt(100))
		}

		avgBetAmount := decimal.Zero
		if pm.TotalBets > 0 {
			avgBetAmount = pm.TotalStake.Div(decimal.NewFromInt(pm.TotalBets))
		}

		metric := dto.ProviderPerformanceMetric{
			Provider:          provider,
			TotalGames:        pm.TotalGames,
			TotalBets:         pm.TotalBets,
			TotalRounds:       pm.TotalRounds,
			UniquePlayers:     pm.UniquePlayersSum,
			TotalStake:        pm.TotalStake,
			TotalWin:          pm.TotalWin,
			GGR:               pm.TotalStake.Sub(pm.TotalWin),
			NGR:               decimal.Zero,
			EffectiveRTP:      effectiveRTP,
			AvgBetAmount:      avgBetAmount,
			RakebackEarned:    rakeback,
			BigWinsCount:      pm.BigWinsCount,
			HighestWin:        pm.HighestWin,
			HighestMultiplier: pm.HighestMultiplier,
		}

		metrics = append(metrics, metric)
	}

	r.sortProviderPerformanceMetrics(metrics, req.SortBy, req.SortOrder)

	total := int64(len(metrics))
	startIdx := (req.Page - 1) * req.PerPage
	endIdx := startIdx + req.PerPage
	if startIdx > len(metrics) {
		metrics = []dto.ProviderPerformanceMetric{}
	} else if endIdx > len(metrics) {
		metrics = metrics[startIdx:]
	} else {
		metrics = metrics[startIdx:endIdx]
	}

	var summary dto.GamePerformanceSummary
	for _, metric := range metrics {
		summary.TotalBets += metric.TotalBets
		summary.TotalUniquePlayers += metric.UniquePlayers
		summary.TotalWagered = summary.TotalWagered.Add(metric.TotalStake)
		summary.TotalGGR = summary.TotalGGR.Add(metric.GGR)
		summary.TotalRakeback = summary.TotalRakeback.Add(metric.RakebackEarned)
	}

	if summary.TotalWagered.GreaterThan(decimal.Zero) {
		totalWin := decimal.Zero
		for _, metric := range metrics {
			totalWin = totalWin.Add(metric.TotalWin)
		}
		summary.AverageRTP = totalWin.Div(summary.TotalWagered).Mul(decimal.NewFromInt(100))
	}

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

func (r *report) sortProviderPerformanceMetrics(metrics []dto.ProviderPerformanceMetric, sortBy *string, sortOrder *string) {
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
