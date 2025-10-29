-- name: CreateRollDaDice :one 
INSERT INTO roll_da_dice(user_id,bet_amount,won_amount,crash_point,user_guessed_start_point,user_guessed_end_point, timestamp,won_status,multiplier,status)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *;

-- name: GetRollDaDiceHistory :many 
WITH roll_da_dice_data AS (
    SELECT *
    FROM roll_da_dice
    WHERE user_id = $1 
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM roll_da_dice_data
     WHERE user_id = $1
)
SELECT c.*, r.total_rows
FROM roll_da_dice_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetRollDaDiceByID :one 
SELECT * FROM roll_da_dice where id = $1;