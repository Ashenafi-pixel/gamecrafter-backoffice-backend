-- Sport Bets Queries based on PlaceBetRequest DTO

-- name: CreateSportBet :one
INSERT INTO sport_bets (
    transaction_id, bet_amount, bet_reference_num, game_reference, bet_mode, 
    description, frontend_type, sport_ids, site_id, 
    client_ip, autorecharge, bet_details, user_id, currency, 
    potential_win, actual_win, odds,status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,$18
) ON CONFLICT (transaction_id) DO UPDATE SET
    bet_amount = EXCLUDED.bet_amount,
    bet_reference_num = EXCLUDED.bet_reference_num,
    game_reference = EXCLUDED.game_reference,
    bet_mode = EXCLUDED.bet_mode,
    description = EXCLUDED.description,
    frontend_type = EXCLUDED.frontend_type,
    sport_ids = EXCLUDED.sport_ids,
    site_id = EXCLUDED.site_id,
    client_ip = EXCLUDED.client_ip,
    autorecharge = EXCLUDED.autorecharge,
    bet_details = EXCLUDED.bet_details,
    user_id = EXCLUDED.user_id,
    currency = EXCLUDED.currency,
    potential_win = EXCLUDED.potential_win,
    actual_win = EXCLUDED.actual_win,
    odds = EXCLUDED.odds,
    status=EXCLUDED.status,
    updated_at = NOW()
RETURNING *;

-- name: UpdateSportBetStatus :one
UPDATE sport_bets 
SET status = $2, 
    actual_win = $3, 
    settled_at = NOW(),
    updated_at = NOW()
WHERE transaction_id = $1 and user_id =$2 and settled_at is null
RETURNING *;

-- name: GetSportBet :one
SELECT * FROM sport_bets WHERE transaction_id = $1;

-- name: ListSportBets :many
SELECT * FROM sport_bets 
WHERE user_id = $1 
ORDER BY placed_at DESC 
LIMIT $2 OFFSET $3; 