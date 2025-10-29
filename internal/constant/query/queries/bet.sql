-- name: SaveBetRound :one 
INSERT INTO rounds (status,crash_point,created_at)
VALUES ($1,$2,$3)
RETURNING *;

-- name: GetBetRoundsByStatus :many 
SELECT * FROM rounds where status = $1;

-- name: UpdateRoundStatusByID :one 
UPDATE rounds SET status = $1 where  id = $2
RETURNING *;

-- name: CloseRoundByID :one 
UPDATE rounds SET status = 'closed',closed_at=$1 where id = $2 
RETURNING *;

-- name: GetBetRoundByID :one 
SELECT * FROM rounds where id = $1;

-- name: PlaceBet :one
INSERT INTO bets (user_id,round_id,amount,currency,client_transaction_id,timestamp)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: GetUserBetByUserIDAndRoundID :many 
SELECT * FROM bets where user_id = $1 and round_id = $2 and status = $3;

-- name: CashOut :one 
UPDATE bets SET cash_out_multiplier = $1,payout=$2,timestamp=$3 where id = $4
RETURNING *;

-- name: ReverseCashOut :one 
UPDATE bets SET cash_out_multiplier = NULL,payout=NULL,timestamp=$1 where id = $2
RETURNING *;

-- name: GetBetHistoryByUserID :many
SELECT 
    rounds.id AS round_id,
    rounds.crash_point, 
    json_agg(
        json_build_object(
            'user_id', bets.user_id,
            'bet_amount', bets.amount,
            'cash_out_multiplier', bets.cash_out_multiplier,
            'payout', bets.payout,
            'currency', bets.currency,
           'timestamp', TO_CHAR(bets.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.US') || 'Z'
        )
    ) AS bets,
    COUNT(rounds.id) OVER() AS total
FROM bets
JOIN rounds ON bets.round_id = rounds.id
WHERE bets.user_id = $1 AND rounds.status = 'closed'
GROUP BY rounds.id, rounds.crash_point
ORDER BY rounds.closed_at
OFFSET $2 LIMIT $3;

-- name: GetBetHistory :many
SELECT 
    rounds.id AS round_id,
    rounds.crash_point, 
    json_agg(
        json_build_object(
            'user_id', bets.user_id,
            'bet_amount', bets.amount,
            'cash_out_multiplier', bets.cash_out_multiplier,
            'payout', bets.payout,
            'currency', bets.currency,
            'timestamp', TO_CHAR(bets.timestamp, 'YYYY-MM-DD"T"HH24:MI:SS.US') || 'Z'
        )
    ) AS bets,COUNT(rounds.id) OVER() AS total
FROM 
    bets 
JOIN 
    rounds 
ON 
    bets.round_id = rounds.id
WHERE 
    rounds.status = 'closed'
GROUP BY 
    rounds.id, rounds.crash_point
OFFSET 
    $1 
LIMIT 
    $2;

-- name: UpdateBetStatus :one 
UPDATE bets set status = $1 where id = $2 RETURNING *;

-- name: GetLeaders :many
SELECT 
    users.username,
    profile,
    COALESCE((SUM(bets.payout) - SUM(bets.amount)), 0)::decimal AS total_cash_out,
    COUNT(*) OVER () AS total_players
FROM 
    users
LEFT JOIN 
    bets
ON 
    users.id = bets.user_id
GROUP BY 
    users.id
ORDER BY 
    total_cash_out DESC LIMIT 15;

-- name: GetFailedRounds :many
SELECT rounds.*,b.user_id,b.amount,b.id as bet_id,b.currency FROM rounds join bets b on rounds.id = b.round_id where (rounds.status='open' or rounds.status = 'in_progress') and b.status='ACTIVE' AND rounds.status != 'failed';

-- name: SaveFailedBetsLogAuto :one 
INSERT INTO failed_bet_logs(user_id,round_id,bet_id,status,manual,created_at,transaction_id,admin_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: SaveFailedBetsLogManual :one 
INSERT INTO failed_bet_logs(user_id,round_id,bet_id,status,admin_id,manual,created_at,transaction_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: GetUserFundByUserIDAndRoundID :one
SELECT * FROM failed_bet_logs where user_id = $1 and round_id =$2 and status ='COMPLTED';


-- name: GetAllFailedRounds :many
SELECT r.id as round_id,r.crash_point, r.status as round_status,r.created_at as round_created_at,b.id as bet_id,b.amount,b.currency,b.client_transaction_id as bet_transaction_id,b.timestamp as bet_timestamp,fbl.id as failed_bet_id, fbl.manual as is_manual,fbl.status refund_status,fbl.created_at as refund_at,fbl.transaction_id as refund_transaction_id,us.* FROM rounds r join bets b on b.round_id = r.id join failed_bet_logs fbl on fbl.user_id = b.user_id join users us on us.id = b.user_id where r.status='failed' ORDER BY r.created_at limit $1 offset $2;

-- name: GetUnRefundedFaildedRouns :many
SELECT r.id as round_id,r.crash_point, r.status as round_status,r.created_at as round_created_at,b.id as bet_id,b.amount,b.currency,b.client_transaction_id as bet_transaction_id,b.timestamp as bet_timestamp,us.* FROM rounds r join bets b on b.round_id = r.id join users us on us.id = b.user_id where r.status='failed'  and (( select  count(id) from failed_bet_logs fbl where fbl.round_id = b.round_id) = 0) ORDER BY r.created_at limit $1 offset $2;

-- name: GetUserActiveBetWithRound :one 
SELECT * FROM bets where user_id =  $1 and round_id = $2  and status ='ACTIVE';

-- name: SavePlinkoBet :one 
INSERT INTO plinko (user_id,bet_amount,drop_path,multiplier,win_amount,finalPosition,timestamp)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: GetPlinkoBetHistoryByUserID :many 
select * from plinko where user_id =$1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3;

-- name: CountPlinkoBetHistoryByUserID :one 
SELECT count(id) as total from plinko where user_id = $1; 

-- name: PlinkoGameState :one 
SELECT 
    COUNT(id) AS total_games, 
    SUM(bet_amount)::decimal AS total_wagered, 
    SUM(win_amount)::decimal AS total_win, 
    (SUM(win_amount) - SUM(bet_amount))::decimal AS net_profit, 
    AVG(multiplier)::decimal AS average_multiplier 
FROM plinko 
WHERE user_id = $1;
-- name: GetUserHighestPlinkoBet :one 
WITH max_bet AS (
    SELECT MAX(win_amount) AS max_bet
    FROM plinko where user_id = $1
)
SELECT p.*
FROM plinko p, max_bet m
WHERE p.win_amount = m.max_bet and p.user_id = $1 limit 1 ;