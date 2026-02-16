-- name: CreateCurrency :one 
INSERT INTO currencies(name) VALUES($1) RETURNING *;

-- name: GetAvailableCurrencies :many 
WITH currency_data AS (
    SELECT *
    FROM currencies where status ='ACTIVE'
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM currency_data
)
SELECT c.*, r.total_rows
FROM currency_data c
CROSS JOIN row_count r
ORDER BY c.name ASC
LIMIT $1 OFFSET $2;

-- name: UpdateCurrencyStatus :one 
UPDATE currencies set status = $1 where id =$2 RETURNING *;

-- name: GetCurrency :many 
WITH currency_data AS (
    SELECT *
    FROM currencies
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM currency_data
)
SELECT c.*, r.total_rows
FROM currency_data c
CROSS JOIN row_count r
ORDER BY c.name ASC
LIMIT $1 OFFSET $2;
