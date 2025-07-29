-- name: CreateScratchCardsBet :one 
INSERT INTO scratch_cards(user_id,status,bet_amount,won_amount,won_status,timestamp)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: GetUserScratchCardBetHistories :many 
WITH scratch_cards_data AS (
    SELECT *
    FROM scratch_cards
    WHERE user_id = $1 and status = 'CLOSED' 
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM scratch_cards_data
     WHERE user_id = $1 AND status = 'CLOSED'
)
SELECT c.*, r.total_rows
FROM scratch_cards_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $2 OFFSET $3;