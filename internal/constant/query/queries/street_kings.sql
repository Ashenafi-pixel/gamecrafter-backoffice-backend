-- name: CreateStreetKingsGame :one 
INSERT INTO street_kings (user_id,version,bet_amount,crash_point,timestamp)
VALUES($1,$2,$3,$4,$5) RETURNING *;

-- name: GetStreetKingsGameByID :one
SELECT * FROM street_kings where id = $1;

-- name: GetStreetKingsGamesByUserIDAndVersion :many 
WITH crash_king_data AS (
    SELECT *
    FROM street_kings
    WHERE user_id = $1 AND version = $2 and status = 'CLOSED' 
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM crash_king_data
     WHERE user_id = $1 AND status = 'CLOSED'
)
SELECT c.*, r.total_rows
FROM crash_king_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $3 OFFSET $4;

-- name: CloseStreetKingGameByID :one 
UPDATE street_kings set status = $1 , won_amount = $2 ,cash_out_point = $3 
WHERE id = $4 RETURNING *;
