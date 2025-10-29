-- name: CreateQuickHustle :one 
INSERT INTO quick_hustles (user_id,bet_amount,first_card,timestamp)
VALUES ($1,$2,$3,$4) RETURNING *;

-- name: GetQuickHustelByID :one 
SELECT * FROM quick_hustles where id = $1;

-- name: CloseQuickHustleGame :one 
UPDATE  quick_hustles set user_guessed = $1,won_status = $2 , second_card = $3,won_amount = $4,status = $5 where id = $6
RETURNING *;

-- name: GetQuickHustleBetHistoy :many 
 WITH crash_king_data AS (
    SELECT *
    FROM quick_hustles
    WHERE user_id = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM crash_king_data  WHERE user_id = $1
)
SELECT c.*, r.total_rows
FROM crash_king_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $2 OFFSET $3;