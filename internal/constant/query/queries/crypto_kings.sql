-- name: CreateCryptoKings :one 
insert into crypto_kings(
    user_id,
    status,
    bet_amount,
    won_amount,
    start_crypto_value,
    end_crypto_value,
    selected_end_second,
    selected_start_value,
    selected_end_value,
    won_status,
    type,
    timestamp
    )
 VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
 ) RETURNING *;

-- name: GetCrytoKingsBetHistoryByUserID :many 
 WITH crash_king_data AS (
    SELECT *
    FROM crypto_kings
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