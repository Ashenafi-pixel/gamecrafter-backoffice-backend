-- name: CreateAirtimeUtiles :one 
INSERT INTO airtime_utilities (id,productName,billerName,amount,isAmountFixed,status,timestamp,price)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: GetAllUtilities :many 
SELECT * FROM airtime_utilities;

-- name: GetAvailableAirtime :many 
WITH airtime_utilities_data AS (
    SELECT 
        au.*, COALESCE(SUM(tr.cashout), 0)::decimal as total_redemptions, COALESCE(SUM(tr.amount), 0)::decimal as total_bucks_spent
    FROM airtime_utilities au 
    LEFT JOIN airtime_transactions tr 
        ON au.id = tr.utilityPackageId
    GROUP BY au.id,au.local_id
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM airtime_utilities_data
)
SELECT 
    c.*, 
    r.total_rows
FROM airtime_utilities_data c 
CROSS JOIN row_count r  
ORDER BY c.timestamp DESC LIMIT $1 OFFSET $2;

-- name: UpdateAirtimeUtilitiesStatus :one
UPDATE airtime_utilities SET status = $1 where local_id = $2 RETURNING *;

-- name: GetAirtimeUtilitiesByID :one
SELECT * FROM airtime_utilities where local_id = $1;

-- name: UpdateAirtimeUtilitiesPrice :one 
UPDATE airtime_utilities SET price = $1 where local_id = $2 RETURNING *;

-- name: GetActiveAvailableAirtime :many 
WITH airtime_utilities_data AS (
    SELECT *
    FROM airtime_utilities WHERE status = 'ACTIVE'
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM airtime_utilities_data
)
SELECT c.*, r.total_rows
FROM airtime_utilities_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $1 OFFSET $2;

-- name: SaveAirtimeTransactions :one 
INSERT INTO airtime_transactions (user_id,transaction_id,cashout,billerName,utilityPackageId,packageName,amount,status,timestamp)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING *;

-- name: GetUserAitimeTransactions :many 
WITH transaction_data AS (
    SELECT *
    FROM airtime_transactions where user_id = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM transaction_data
)
SELECT c.*, r.total_rows
FROM transaction_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetAllAitimeTransactions :many 
WITH transaction_data AS (
    SELECT *
    FROM airtime_transactions
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM transaction_data
)
SELECT c.*, r.total_rows
FROM transaction_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $1 OFFSET $2;

-- name: UpdateAirtimeUtilitiesAmount :one
UPDATE airtime_utilities SET amount = $1 where local_id = $2 RETURNING *;

-- name: GetAirtimeUtilitiesStats :one 
SELECT 
    COUNT(*)::int AS total,
    COALESCE(SUM(CASE WHEN au.status = 'ACTIVE' THEN 1 ELSE 0 END), 0)::int AS active_utilities,
    COALESCE(SUM(at.amount), 0)::decimal AS total_spend_bucks,
	COALESCE(SUM(at.cashout), 0)::decimal AS total_redemptions,
    COALESCE(SUM(CASE WHEN au.status = 'INACTIVE' THEN 1 ELSE 0 END), 0)::int AS inactive_utilities
FROM airtime_utilities au
LEFT JOIN airtime_transactions at ON au.id = at.utilityPackageId;