-- name: CreateBalance :one 
INSERT INTO balances(user_id,currency,real_money,bonus_money,updated_at) VALUES 
($1,$2,$3,$4,$5) RETURNING *;

-- name: UpdateBalance :one 
UPDATE balances set currency = $1,real_money = $2,bonus_money=$3,updated_at=$4 where user_id = $5
RETURNING *;

-- name: UpdateRealMoney :one 
UPDATE balances set real_money = $1,updated_at=$2 where user_id = $3 and currency = $4
RETURNING *;

-- name: UpdateBonusMoney :one 
UPDATE balances set bonus_money = $1,updated_at=$2 where user_id = $3 and currency = $4
RETURNING *;

-- name: UpdatePoints :one 
UPDATE balances set points = $1,updated_at=$2 where user_id = $3 and currency = $4
RETURNING *;

-- name: LockBalance :one
SELECT * FROM balances 
WHERE user_id = $1 and currency = $2
FOR UPDATE;

-- name: GetUserBalanaceByUserIDAndCurrency :one 
SELECT * FROM balances where user_id = $1 and currency=$2;

-- name: GetUserBalancesByUserID :many 
SELECT * FROM  balances where user_id = $1;

-- name: SaveBalanceLogs :one 
INSERT INTO balance_logs (user_id,component,currency,change_amount,operational_group_id,operational_type_id,description,TIMESTAMP,balance_after_update,transaction_id,status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING *;

-- name: UpdatePointByUserID :one 
UPDATE balances set real_money = $1 where user_id = $2 and currency = $3 RETURNING *;


-- name: BalanceExist :one
SELECT EXISTS (
    SELECT 1 
    FROM balances 
    WHERE user_id = $1 AND currency = $2
) AS exists;
