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

// GetBigWinners retrieves big winners report with filters
func (r *report) GetBigWinners(ctx context.Context, req dto.BigWinnersReportReq, userBrandIDs []uuid.UUID) (dto.BigWinnersReportRes, error) {
	var res dto.BigWinnersReportRes

	// Set defaults
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
			// Set to end of day
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &endOfDay
		}
	}

	// If no date range provided, default to last 24 hours
	if dateFrom == nil && dateTo == nil {
		now := time.Now()
		dateTo = &now
		yesterday := now.Add(-24 * time.Hour)
		dateFrom = &yesterday
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
		whereConditions = append(whereConditions, fmt.Sprintf("(brand_id IS NULL OR brand_id IN (%s))", strings.Join(brandPlaceholders, ",")))
	}

	// Additional brand filter from request
	if req.BrandID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("brand_id = $%d", argIndex))
		args = append(args, *req.BrandID)
		argIndex++
	}

	// Date range
	if dateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("date_time >= $%d", argIndex))
		args = append(args, *dateFrom)
		argIndex++
	}
	if dateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("date_time <= $%d", argIndex))
		args = append(args, *dateTo)
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

	// Player search (username, email, or user ID)
	if req.PlayerSearch != nil && *req.PlayerSearch != "" {
		searchTerm := "%" + *req.PlayerSearch + "%"
		whereConditions = append(whereConditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d OR player_id::text = $%d)", argIndex, argIndex, argIndex))
		args = append(args, searchTerm)
		argIndex++
	}

	// Minimum win threshold
	threshold := decimal.NewFromFloat(*req.MinWinThreshold)
	whereConditions = append(whereConditions, fmt.Sprintf("win_amount >= $%d", argIndex))
	args = append(args, threshold)
	argIndex++

	// Bet type filter
	if req.BetType != nil && *req.BetType != "" && *req.BetType != "both" {
		whereConditions = append(whereConditions, fmt.Sprintf("bet_type = $%d", argIndex))
		args = append(args, *req.BetType)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "date_time DESC"
	if req.SortBy != nil {
		switch *req.SortBy {
		case "win_amount":
			orderBy = "win_amount"
		case "net_win":
			orderBy = "net_win"
		case "multiplier":
			orderBy = "win_multiplier"
		case "date":
			orderBy = "date_time"
		default:
			orderBy = "date_time"
		}
		if req.SortOrder != nil {
			orderBy += " " + strings.ToUpper(*req.SortOrder)
		} else {
			orderBy += " DESC"
		}
	}

	// Build the main query using CTE to combine all bet sources
	query := fmt.Sprintf(`
		WITH all_big_wins AS (
			-- GrooveTech transactions (wins from result type)
			SELECT 
				gt.id::uuid as id,
				gt.created_at as date_time,
				ga.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				b.name as brand_name,
				gt.game_id as game_provider,
				gt.game_id,
				NULL::text as game_name,
				gt.transaction_id as bet_id,
				gt.round_id,
				COALESCE(ABS((SELECT amount FROM groove_transactions gt2 
					WHERE gt2.round_id = gt.round_id 
					AND gt2.type = 'wager' 
					AND gt2.account_id = gt.account_id 
					ORDER BY gt2.created_at DESC LIMIT 1)), 0) as stake_amount,
				gt.amount as win_amount,
				gt.amount - COALESCE(ABS((SELECT amount FROM groove_transactions gt2 
					WHERE gt2.round_id = gt.round_id 
					AND gt2.type = 'wager' 
					AND gt2.account_id = gt.account_id 
					ORDER BY gt2.created_at DESC LIMIT 1)), 0) as net_win,
				COALESCE(ga.currency, 'USD') as currency,
				CASE 
					WHEN COALESCE(ABS((SELECT amount FROM groove_transactions gt2 
						WHERE gt2.round_id = gt.round_id 
						AND gt2.type = 'wager' 
						AND gt2.account_id = gt.account_id 
						ORDER BY gt2.created_at DESC LIMIT 1)), 0) > 0 
					THEN gt.amount / ABS((SELECT amount FROM groove_transactions gt2 
						WHERE gt2.round_id = gt.round_id 
						AND gt2.type = 'wager' 
						AND gt2.account_id = gt.account_id 
						ORDER BY gt2.created_at DESC LIMIT 1))
					ELSE NULL
				END as win_multiplier,
				'cash' as bet_type,
				false as is_jackpot,
				NULL::text as jackpot_name,
				gt.game_session_id as session_id,
				u.country,
				'groove' as bet_source
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			LEFT JOIN brands b ON u.brand_id = b.id
			WHERE gt.type = 'result' 
				AND gt.amount > 0
				AND gt.status = 'completed'

			UNION ALL

			-- General bets table
			SELECT 
				b.id,
				COALESCE(b.timestamp, b.created_at, NOW()) as date_time,
				b.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				b.id::text as bet_id,
				b.round_id::text as round_id,
				b.amount as stake_amount,
				COALESCE(b.payout, 0) as win_amount,
				COALESCE(b.payout, 0) - b.amount as net_win,
				b.currency,
				CASE 
					WHEN b.amount > 0 THEN COALESCE(b.payout, 0) / b.amount
					ELSE NULL
				END as win_multiplier,
				'cash' as bet_type,
				false as is_jackpot,
				NULL::text as jackpot_name,
				NULL::text as session_id,
				u.country,
				'bets' as bet_source
			FROM bets b
			JOIN users u ON b.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE COALESCE(b.payout, 0) > 0
				AND COALESCE(b.status, 'completed') = 'completed'

			UNION ALL

			-- Sport bets
			SELECT 
				sb.id::uuid as id,
				sb.created_at as date_time,
				sb.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				sb.transaction_id as bet_id,
				sb.round_id::text as round_id,
				sb.bet_amount as stake_amount,
				COALESCE(sb.actual_win, 0) as win_amount,
				COALESCE(sb.actual_win, 0) - sb.bet_amount as net_win,
				sb.currency,
				CASE 
					WHEN sb.bet_amount > 0 THEN COALESCE(sb.actual_win, 0) / sb.bet_amount
					ELSE NULL
				END as win_multiplier,
				'cash' as bet_type,
				false as is_jackpot,
				NULL::text as jackpot_name,
				NULL::text as session_id,
				u.country,
				'sport_bets' as bet_source
			FROM sport_bets sb
			JOIN users u ON sb.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE COALESCE(sb.actual_win, 0) > 0
				AND COALESCE(sb.status, 'completed') = 'completed'

			UNION ALL

			-- Plinko bets
			SELECT 
				p.id::uuid as id,
				p.timestamp as date_time,
				p.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				'Plinko' as game_provider,
				'plinko' as game_id,
				'Plinko' as game_name,
				p.id::text as bet_id,
				NULL::text as round_id,
				p.bet_amount as stake_amount,
				p.win_amount,
				p.win_amount - p.bet_amount as net_win,
				COALESCE(u.default_currency, 'USD') as currency,
				CASE 
					WHEN p.bet_amount > 0 THEN p.win_amount / p.bet_amount
					ELSE NULL
				END as win_multiplier,
				'cash' as bet_type,
				false as is_jackpot,
				NULL::text as jackpot_name,
				NULL::text as session_id,
				u.country,
				'plinko' as bet_source
			FROM plinko p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE p.win_amount > 0
		)
		SELECT 
			id,
			date_time,
			player_id,
			username,
			email,
			brand_id,
			brand_name,
			game_provider,
			game_id,
			game_name,
			bet_id,
			round_id,
			stake_amount,
			win_amount,
			net_win,
			currency,
			win_multiplier,
			bet_type,
			is_jackpot,
			jackpot_name,
			session_id,
			country,
			bet_source
		FROM all_big_wins
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	// Add pagination args
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	// Execute query
	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to get big winners", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "failed to get big winners")
	}
	defer rows.Close()

	var winners []dto.BigWinner
	for rows.Next() {
		var winner dto.BigWinner
		var winMultiplier sql.NullString
		var email sql.NullString
		var brandID uuid.NullUUID
		var brandName sql.NullString
		var gameProvider sql.NullString
		var gameID sql.NullString
		var gameName sql.NullString
		var betID sql.NullString
		var roundID sql.NullString
		var jackpotName sql.NullString
		var sessionID sql.NullString
		var country sql.NullString

		err := rows.Scan(
			&winner.ID,
			&winner.DateTime,
			&winner.PlayerID,
			&winner.Username,
			&email,
			&brandID,
			&brandName,
			&gameProvider,
			&gameID,
			&gameName,
			&betID,
			&roundID,
			&winner.StakeAmount,
			&winner.WinAmount,
			&winner.NetWin,
			&winner.Currency,
			&winMultiplier,
			&winner.BetType,
			&winner.IsJackpot,
			&jackpotName,
			&sessionID,
			&country,
			&winner.BetSource,
		)
		if err != nil {
			r.log.Error("failed to scan big winner row", zap.Error(err))
			continue
		}

		if email.Valid {
			winner.Email = &email.String
		}
		if brandID.Valid {
			winner.BrandID = &brandID.UUID
		}
		if brandName.Valid {
			winner.BrandName = &brandName.String
		}
		if gameProvider.Valid {
			winner.GameProvider = &gameProvider.String
		}
		if gameID.Valid {
			winner.GameID = &gameID.String
		}
		if gameName.Valid {
			winner.GameName = &gameName.String
		}
		if betID.Valid {
			winner.BetID = &betID.String
		}
		if roundID.Valid {
			winner.RoundID = &roundID.String
		}
		if jackpotName.Valid {
			winner.JackpotName = &jackpotName.String
		}
		if sessionID.Valid {
			winner.SessionID = &sessionID.String
		}
		if country.Valid {
			winner.Country = &country.String
		}
		if winMultiplier.Valid {
			multiplier, err := decimal.NewFromString(winMultiplier.String)
			if err == nil {
				winner.WinMultiplier = &multiplier
			}
		}

		winners = append(winners, winner)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("error iterating big winners rows", zap.Error(err))
		return res, errors.ErrUnableToGet.Wrap(err, "error iterating big winners rows")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		WITH all_big_wins AS (
			-- Same CTE as main query
			SELECT 
				gt.id::uuid as id,
				gt.created_at as date_time,
				ga.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				b.name as brand_name,
				gt.game_id as game_provider,
				gt.game_id,
				NULL::text as game_name,
				gt.transaction_id as bet_id,
				gt.round_id,
				COALESCE(ABS((SELECT amount FROM groove_transactions gt2 
					WHERE gt2.round_id = gt.round_id 
					AND gt2.type = 'wager' 
					AND gt2.account_id = gt.account_id 
					ORDER BY gt2.created_at DESC LIMIT 1)), 0) as stake_amount,
				gt.amount as win_amount,
				gt.amount - COALESCE(ABS((SELECT amount FROM groove_transactions gt2 
					WHERE gt2.round_id = gt.round_id 
					AND gt2.type = 'wager' 
					AND gt2.account_id = gt.account_id 
					ORDER BY gt2.created_at DESC LIMIT 1)), 0) as net_win,
				COALESCE(ga.currency, 'USD') as currency,
				'cash' as bet_type,
				'groove' as bet_source
			FROM groove_transactions gt
			JOIN groove_accounts ga ON gt.account_id = ga.account_id
			JOIN users u ON ga.user_id = u.id
			LEFT JOIN brands b ON u.brand_id = b.id
			WHERE gt.type = 'result' 
				AND gt.amount > 0
				AND gt.status = 'completed'

			UNION ALL

			SELECT 
				b.id,
				COALESCE(b.timestamp, b.created_at, NOW()) as date_time,
				b.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				b.id::text as bet_id,
				b.round_id::text as round_id,
				b.amount as stake_amount,
				COALESCE(b.payout, 0) as win_amount,
				COALESCE(b.payout, 0) - b.amount as net_win,
				b.currency,
				'cash' as bet_type,
				'bets' as bet_source
			FROM bets b
			JOIN users u ON b.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE COALESCE(b.payout, 0) > 0
				AND COALESCE(b.status, 'completed') = 'completed'

			UNION ALL

			SELECT 
				sb.id::uuid as id,
				sb.created_at as date_time,
				sb.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				NULL::text as game_provider,
				NULL::text as game_id,
				NULL::text as game_name,
				sb.transaction_id as bet_id,
				sb.round_id::text as round_id,
				sb.bet_amount as stake_amount,
				COALESCE(sb.actual_win, 0) as win_amount,
				COALESCE(sb.actual_win, 0) - sb.bet_amount as net_win,
				sb.currency,
				'cash' as bet_type,
				'sport_bets' as bet_source
			FROM sport_bets sb
			JOIN users u ON sb.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE COALESCE(sb.actual_win, 0) > 0
				AND COALESCE(sb.status, 'completed') = 'completed'

			UNION ALL

			SELECT 
				p.id::uuid as id,
				p.timestamp as date_time,
				p.user_id as player_id,
				u.username,
				u.email,
				u.brand_id,
				brand.name as brand_name,
				'Plinko' as game_provider,
				'plinko' as game_id,
				'Plinko' as game_name,
				p.id::text as bet_id,
				NULL::text as round_id,
				p.bet_amount as stake_amount,
				p.win_amount,
				p.win_amount - p.bet_amount as net_win,
				COALESCE(u.default_currency, 'USD') as currency,
				'cash' as bet_type,
				'plinko' as bet_source
			FROM plinko p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN brands brand ON u.brand_id = brand.id
			WHERE p.win_amount > 0
		)
		SELECT COUNT(*) as total
		FROM all_big_wins
		%s
	`, whereClause)

	var total int64
	err = r.db.GetPool().QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		r.log.Error("failed to get big winners count", zap.Error(err))
		total = int64(len(winners)) // Fallback to current page count
	}

	// Calculate summary
	var summary dto.BigWinnersSummary
	summary.Count = total
	for _, winner := range winners {
		summary.TotalWins = summary.TotalWins.Add(winner.WinAmount)
		summary.TotalNetWins = summary.TotalNetWins.Add(winner.NetWin)
		summary.TotalStakes = summary.TotalStakes.Add(winner.StakeAmount)
	}

	// Calculate total pages
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

