package report

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	analyticsStorage "github.com/tucanbit/internal/storage/analytics"
	"go.uber.org/zap"
)

func (r *report) getClickHouseConn() (driver.Conn, bool) {
	if r.analyticsStorage == nil {
		return nil, false
	}
	if impl, ok := r.analyticsStorage.(*analyticsStorage.AnalyticsStorageImpl); ok {
		if impl != nil {
			chClient := impl.GetClickHouseClient()
			if chClient != nil {
				return chClient.GetConnection(), true
			}
		}
	}
	return nil, false
}

func (r *report) GetBigWinners(ctx context.Context, req dto.BigWinnersReportReq, userBrandIDs []uuid.UUID) (dto.BigWinnersReportRes, error) {
	var res dto.BigWinnersReportRes

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}
	if req.MinWinThreshold == nil {
		defaultThreshold := 100.0
		req.MinWinThreshold = &defaultThreshold
	}
	if req.SortBy == nil {
		defaultSort := "date"
		req.SortBy = &defaultSort
	}
	if req.SortOrder == nil {
		defaultOrder := "desc"
		req.SortOrder = &defaultOrder
	}

	var dateFrom, dateTo *time.Time
	if req.DateFrom != nil && *req.DateFrom != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05", *req.DateFrom)
		if err != nil {
			parsed, err = time.Parse("2006-01-02T15:04", *req.DateFrom)
		}
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
			parsed, err = time.Parse("2006-01-02T15:04", *req.DateTo)
		}
		if err != nil {
			parsed, err = time.Parse("2006-01-02", *req.DateTo)
		}
		if err == nil {
			if len(*req.DateTo) == 10 {
				endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				dateTo = &endOfDay
			} else {
				dateTo = &parsed
			}
		}
	}

	if dateFrom == nil && dateTo == nil {
		now := time.Now()
		dateTo = &now
		yesterday := now.Add(-24 * time.Hour)
		dateFrom = &yesterday
	}

	threshold := decimal.NewFromFloat(*req.MinWinThreshold)

	chConn, hasClickHouse := r.getClickHouseConn()
	var gamingWins []dto.BigWinner

	if hasClickHouse && dateFrom != nil && dateTo != nil {
		startDateStr := dateFrom.Format("2006-01-02 15:04:05")
		endDateStr := dateTo.Format("2006-01-02 15:04:05")

		baseWhereConditions := []string{
			"t.transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')",
			"((t.transaction_type IN ('groove_bet', 'groove_win') AND (t.status = 'completed' OR (t.status = 'pending' AND t.bet_amount IS NOT NULL AND t.win_amount IS NOT NULL))) OR (t.transaction_type NOT IN ('groove_bet', 'groove_win') AND (t.status = 'completed' OR t.status IS NULL)))",
			fmt.Sprintf("t.created_at >= '%s' AND t.created_at <= '%s'", startDateStr, endDateStr),
		}

		if req.GameProvider != nil && *req.GameProvider != "" {
			baseWhereConditions = append(baseWhereConditions, fmt.Sprintf("t.provider = '%s'", *req.GameProvider))
		}

		if req.GameID != nil && *req.GameID != "" {
			baseWhereConditions = append(baseWhereConditions, fmt.Sprintf("t.game_id = '%s'", *req.GameID))
		}

		if req.PlayerSearch != nil && *req.PlayerSearch != "" {
			baseWhereConditions = append(baseWhereConditions, fmt.Sprintf("t.user_id = '%s'", *req.PlayerSearch))
		}

		if len(userBrandIDs) > 0 {
			brandIDs := make([]string, len(userBrandIDs))
			for i, brandID := range userBrandIDs {
				brandIDs[i] = fmt.Sprintf("'%s'", brandID.String())
			}
			baseWhereConditions = append(baseWhereConditions, fmt.Sprintf("(t.brand_id = '' OR t.brand_id IN (%s))", strings.Join(brandIDs, ",")))
		}

		if req.BrandID != nil {
			baseWhereConditions = append(baseWhereConditions, fmt.Sprintf("t.brand_id = '%s'", req.BrandID.String()))
		}

		chQuery := fmt.Sprintf(`
			WITH round_wins AS (
				SELECT 
					COALESCE(t.round_id, t.session_id, toString(t.id)) as round_id,
					t.user_id,
					t.game_id,
					t.session_id,
					t.provider as game_provider,
					t.game_name,
					t.brand_id,
					max(t.created_at) as created_at,
					sum(CASE 
						WHEN t.transaction_type IN ('groove_bet', 'groove_win') AND t.bet_amount IS NOT NULL THEN t.bet_amount
						WHEN t.transaction_type = 'groove_bet' THEN abs(t.amount)
						ELSE 0
					END) as stake_amount,
					sum(CASE 
						WHEN t.transaction_type IN ('groove_bet', 'groove_win') THEN COALESCE(t.win_amount, 0)
						WHEN t.transaction_type IN ('win', 'groove_win') THEN COALESCE(t.win_amount, abs(t.amount))
						ELSE 0
					END) as win_amount,
					toString(any(t.id)) as id
				FROM tucanbit_analytics.transactions t
				WHERE %s
				GROUP BY round_id, t.user_id, t.game_id, t.session_id, t.provider, t.game_name, t.brand_id
			)
			SELECT 
				id,
				created_at,
				user_id,
				game_id,
				session_id,
				stake_amount,
				win_amount,
				game_provider,
				game_name,
				brand_id
			FROM round_wins
			WHERE win_amount > 0 AND win_amount >= toDecimal64(%.8f, 8)
		`, strings.Join(baseWhereConditions, " AND "), threshold.InexactFloat64())

		chArgs := []interface{}{}
		rows, err := chConn.Query(ctx, chQuery, chArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var win dto.BigWinner
				var idStr, userIDStr, gameIDStr, sessionIDStr, gameProviderStr, gameNameStr, brandIDStr string
				var stakeAmountStr, winAmountStr string
				var createdAt time.Time

				err := rows.Scan(
					&idStr,
					&createdAt,
					&userIDStr,
					&gameIDStr,
					&sessionIDStr,
					&stakeAmountStr,
					&winAmountStr,
					&gameProviderStr,
					&gameNameStr,
					&brandIDStr,
				)
				if err != nil {
					r.log.Warn("Failed to scan ClickHouse big win row", zap.Error(err))
					continue
				}

				if idStr != "" {
					if parsedID, err := uuid.Parse(idStr); err == nil {
						win.ID = parsedID
					} else {
						win.ID = uuid.New()
					}
				} else {
					win.ID = uuid.New()
				}

				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					continue
				}
				win.PlayerID = userID
				win.DateTime = createdAt

				if gameIDStr != "" {
					win.GameID = &gameIDStr
				}
				if sessionIDStr != "" {
					win.SessionID = &sessionIDStr
				}
				if gameProviderStr != "" {
					win.GameProvider = &gameProviderStr
				}
				if gameNameStr != "" {
					win.GameName = &gameNameStr
				}

				stakeAmount, err := decimal.NewFromString(stakeAmountStr)
				if err == nil {
					win.StakeAmount = stakeAmount
				}

				winAmount, err := decimal.NewFromString(winAmountStr)
				if err == nil {
					win.WinAmount = winAmount
					win.NetWin = winAmount.Sub(stakeAmount)
					if !stakeAmount.IsZero() {
						multiplier := winAmount.Div(stakeAmount)
						win.WinMultiplier = &multiplier
					}
				}

				if brandIDStr != "" && brandIDStr != "00000000-0000-0000-0000-000000000001" {
					brandID, err := uuid.Parse(brandIDStr)
					if err == nil {
						win.BrandID = &brandID
					}
				}

				win.BetType = "cash"
				win.Currency = "USD"
				win.BetSource = "clickhouse_gaming"

				gamingWins = append(gamingWins, win)
			}
		} else {
			r.log.Warn("Failed to query ClickHouse for gaming wins", zap.Error(err))
		}
	}

	// All gaming data comes from ClickHouse only - no PostgreSQL queries needed
	// User details will be fetched separately and merged in Go
	allWinners := gamingWins

	userMap := make(map[uuid.UUID]struct {
		Username  string
		Email     *string
		BrandID   *uuid.UUID
		BrandName *string
		Country   *string
		IsTest    bool
	})

	if len(allWinners) > 0 {
		userIDs := make([]uuid.UUID, 0)
		for _, win := range allWinners {
			userIDs = append(userIDs, win.PlayerID)
		}

		if len(userIDs) > 0 {
			userPlaceholders := make([]string, len(userIDs))
			userArgs := make([]interface{}, len(userIDs))
			for i, userID := range userIDs {
				userPlaceholders[i] = fmt.Sprintf("$%d", i+1)
				userArgs[i] = userID
			}

			userQuery := fmt.Sprintf(`
			SELECT 
					u.id,
				u.username,
				u.email,
				u.brand_id,
				b.name as brand_name,
				u.country,
					u.is_test_account
				FROM users u
			LEFT JOIN brands b ON u.brand_id = b.id
				WHERE u.id IN (%s)
			`, strings.Join(userPlaceholders, ","))

			userRows, err := r.db.GetPool().Query(ctx, userQuery, userArgs...)
			if err == nil {
				defer userRows.Close()
				for userRows.Next() {
					var userID uuid.UUID
					var username string
					var email sql.NullString
					var brandID uuid.NullUUID
					var brandName sql.NullString
					var country sql.NullString
					var isTest bool

					if err := userRows.Scan(&userID, &username, &email, &brandID, &brandName, &country, &isTest); err == nil {
						var emailPtr *string
						if email.Valid {
							emailPtr = &email.String
						}
						var brandIDPtr *uuid.UUID
						if brandID.Valid {
							brandIDPtr = &brandID.UUID
						}
						var brandNamePtr *string
						if brandName.Valid {
							brandNamePtr = &brandName.String
						}
						var countryPtr *string
						if country.Valid {
							countryPtr = &country.String
						}

						userMap[userID] = struct {
							Username  string
							Email     *string
							BrandID   *uuid.UUID
							BrandName *string
							Country   *string
							IsTest    bool
						}{
							Username:  username,
							Email:     emailPtr,
							BrandID:   brandIDPtr,
							BrandName: brandNamePtr,
							Country:   countryPtr,
							IsTest:    isTest,
						}
					}
				}
			}
		}
	}

	gameMap := make(map[string]struct {
		GameName  *string
		Provider  *string
		SessionID *string
	})

	if len(allWinners) > 0 {
		gameIDs := make([]string, 0)
		sessionIDs := make([]string, 0)
		for _, win := range allWinners {
			if win.GameID != nil && *win.GameID != "" {
				gameIDs = append(gameIDs, *win.GameID)
			}
			if win.SessionID != nil && *win.SessionID != "" {
				sessionIDs = append(sessionIDs, *win.SessionID)
			}
		}

		if len(gameIDs) > 0 {
			gamePlaceholders := make([]string, len(gameIDs))
			gameArgs := make([]interface{}, len(gameIDs))
			for i, gameID := range gameIDs {
				gamePlaceholders[i] = fmt.Sprintf("$%d", i+1)
				gameArgs[i] = gameID
			}

			gameQuery := fmt.Sprintf(`
			SELECT 
					g.game_id,
					g.name as game_name,
					g.provider
				FROM games g
				WHERE g.game_id IN (%s)
			`, strings.Join(gamePlaceholders, ","))

			gameRows, err := r.db.GetPool().Query(ctx, gameQuery, gameArgs...)
			if err == nil {
				defer gameRows.Close()
				for gameRows.Next() {
					var gameID string
					var gameName sql.NullString
					var provider sql.NullString

					if err := gameRows.Scan(&gameID, &gameName, &provider); err == nil {
						var gameNamePtr *string
						if gameName.Valid {
							gameNamePtr = &gameName.String
						}
						var providerPtr *string
						if provider.Valid {
							providerPtr = &provider.String
						}

						gameMap[gameID] = struct {
							GameName  *string
							Provider  *string
							SessionID *string
						}{
							GameName: gameNamePtr,
							Provider: providerPtr,
						}
					}
				}
			}
		}

		if len(sessionIDs) > 0 {
			sessionPlaceholders := make([]string, len(sessionIDs))
			sessionArgs := make([]interface{}, len(sessionIDs))
			for i, sessionID := range sessionIDs {
				sessionPlaceholders[i] = fmt.Sprintf("$%d", i+1)
				sessionArgs[i] = sessionID
			}

			sessionQuery := fmt.Sprintf(`
			SELECT 
					gs.session_id,
					gs.game_id
				FROM game_sessions gs
				WHERE gs.session_id IN (%s)
			`, strings.Join(sessionPlaceholders, ","))

			sessionRows, err := r.db.GetPool().Query(ctx, sessionQuery, sessionArgs...)
			if err == nil {
				defer sessionRows.Close()
				for sessionRows.Next() {
					var sessionID, gameID string

					if err := sessionRows.Scan(&sessionID, &gameID); err == nil {
						if gameInfo, exists := gameMap[gameID]; exists {
							gameInfo.SessionID = &sessionID
							gameMap[gameID] = gameInfo
						}
					}
				}
			}
		}
	}

	filteredWinners := make([]dto.BigWinner, 0)
	for i := range allWinners {
		if userInfo, exists := userMap[allWinners[i].PlayerID]; exists {
			allWinners[i].Username = userInfo.Username
			allWinners[i].Email = userInfo.Email
			allWinners[i].BrandID = userInfo.BrandID
			allWinners[i].BrandName = userInfo.BrandName
			allWinners[i].Country = userInfo.Country

			if req.IsTestAccount != nil {
				if *req.IsTestAccount != userInfo.IsTest {
					continue
				}
			}
		}

		if allWinners[i].GameID != nil {
			if gameInfo, exists := gameMap[*allWinners[i].GameID]; exists {
				if gameInfo.GameName != nil && allWinners[i].GameName == nil {
					allWinners[i].GameName = gameInfo.GameName
				}
				if gameInfo.Provider != nil && allWinners[i].GameProvider == nil {
					allWinners[i].GameProvider = gameInfo.Provider
				}
			}
		}

		filteredWinners = append(filteredWinners, allWinners[i])
	}
	allWinners = filteredWinners

	if req.SortBy != nil {
		switch *req.SortBy {
		case "win_amount":
			if *req.SortOrder == "asc" {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].WinAmount.GreaterThan(allWinners[j].WinAmount) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].WinAmount.LessThan(allWinners[j].WinAmount) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			}
		case "net_win":
			if *req.SortOrder == "asc" {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].NetWin.GreaterThan(allWinners[j].NetWin) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].NetWin.LessThan(allWinners[j].NetWin) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			}
		case "date":
			if *req.SortOrder == "asc" {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].DateTime.After(allWinners[j].DateTime) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			} else {
				for i := 0; i < len(allWinners)-1; i++ {
					for j := i + 1; j < len(allWinners); j++ {
						if allWinners[i].DateTime.Before(allWinners[j].DateTime) {
							allWinners[i], allWinners[j] = allWinners[j], allWinners[i]
						}
					}
				}
			}
		}
	}

	total := int64(len(allWinners))
	startIdx := (req.Page - 1) * req.PerPage
	endIdx := startIdx + req.PerPage
	if startIdx > len(allWinners) {
		startIdx = len(allWinners)
	}
	if endIdx > len(allWinners) {
		endIdx = len(allWinners)
	}

	var winners []dto.BigWinner
	if startIdx < len(allWinners) {
		winners = allWinners[startIdx:endIdx]
	} else {
		winners = make([]dto.BigWinner, 0)
	}

	var summary dto.BigWinnersSummary
	summary.Count = total
	for _, winner := range allWinners {
		summary.TotalWins = summary.TotalWins.Add(winner.WinAmount)
		summary.TotalNetWins = summary.TotalNetWins.Add(winner.NetWin)
		summary.TotalStakes = summary.TotalStakes.Add(winner.StakeAmount)
	}

	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	res.Message = "Big winners retrieved successfully"
	res.Data = winners
	res.Total = total
	res.TotalPages = totalPages
	res.Page = req.Page
	res.PerPage = req.PerPage
	res.Summary = summary

	return res, nil
}
