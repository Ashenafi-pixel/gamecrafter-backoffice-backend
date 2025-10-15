-- name: GetSystemConfigValue :one
SELECT config_value FROM system_config WHERE config_key = $1;

-- name: UpdateSystemConfigValue :exec
UPDATE system_config 
SET config_value = $2, updated_by = $3, updated_at = NOW()
WHERE config_key = $1;

-- name: CreateSystemConfig :one
INSERT INTO system_config (config_key, config_value, description, updated_by)
VALUES ($1, $2, $3, $4)
RETURNING id, config_key, config_value, description, updated_by, created_at, updated_at;

-- name: GetSystemConfig :one
SELECT id, config_key, config_value, description, updated_by, created_at, updated_at
FROM system_config 
WHERE config_key = $1;

-- name: ListSystemConfigs :many
SELECT id, config_key, config_value, description, updated_by, created_at, updated_at
FROM system_config
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteSystemConfig :exec
DELETE FROM system_config WHERE config_key = $1;

-- name: GetWithdrawalsByIDs :many
SELECT 
    w.id, w.user_id, w.withdrawal_id, w.usd_amount_cents, w.crypto_amount, 
    w.currency_code, w.status, w.created_at, w.updated_at,
    u.username, u.email
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.withdrawal_id = ANY($1::text[]);
