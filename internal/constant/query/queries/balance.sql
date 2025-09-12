-- name: CreateBalance :one 
INSERT INTO balances(user_id,currency_code,amount_cents,amount_units,reserved_cents,reserved_units,updated_at) VALUES 
($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: UpdateBalance :one 
UPDATE balances set currency_code = $1,amount_cents=$2,amount_units=$3,reserved_cents=$4,reserved_units=$5,updated_at=$6 where user_id = $7
RETURNING *;

-- name: UpdateAmountUnits :one 
UPDATE balances set amount_units = $1, amount_cents = $2, updated_at=$3 where user_id = $4 and currency_code = $5
RETURNING *;

-- name: UpdateReservedUnits :one 
UPDATE balances set reserved_units = $1, reserved_cents = $2, updated_at=$3 where user_id = $4 and currency_code = $5
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
