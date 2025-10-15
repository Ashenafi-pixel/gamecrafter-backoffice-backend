-- name: CreateBalance :one 
INSERT INTO balances(user_id,currency_code,real_money,bonus_money,points,updated_at) VALUES 
($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: UpdateBalance :one 
UPDATE balances set currency_code = $1,real_money=$2,bonus_money=$3,points=$4,updated_at=$5 where user_id = $6
RETURNING *;

-- name: UpdateAmountUnits :one 
UPDATE balances set real_money = $1, bonus_money = $2, updated_at=$3 where user_id = $4 and currency_code = $5
RETURNING *;

-- name: UpdateReservedUnits :one 
UPDATE balances set points = $1, updated_at=$2 where user_id = $3 and currency_code = $4
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
