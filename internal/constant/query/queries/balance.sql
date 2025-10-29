-- name: CreateBalance :one 
INSERT INTO balances(user_id,currency_code,amount_cents,amount_units,reserved_cents,reserved_units,updated_at) VALUES 
($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: UpdateBalance :one 
UPDATE balances set currency_code = $1,amount_units=$2,reserved_units=$3,reserved_cents=$4,updated_at=$5 where user_id = $6
RETURNING *;

-- name: UpdateAmountUnits :one 
UPDATE balances set amount_units = $1, reserved_units = $2, updated_at=$3 where user_id = $4 and currency_code = $5
RETURNING *;

-- name: UpdateReservedUnits :one 
UPDATE balances set reserved_cents = $1, updated_at=$2 where user_id = $3 and currency_code = $4
RETURNING *;

-- name: LockBalance :one
SELECT * FROM balances 
WHERE user_id = $1 and currency_code = $2
FOR UPDATE;

-- name: GetUserBalanaceByUserIDAndCurrency :one 
SELECT * FROM balances where user_id = $1 and currency_code=$2;

-- name: GetUserBalancesByUserID :many 
SELECT * FROM  balances where user_id = $1;


-- name: BalanceExist :one
SELECT EXISTS (
    SELECT 1 
    FROM balances 
    WHERE user_id = $1 AND currency_code = $2
) AS exists;
