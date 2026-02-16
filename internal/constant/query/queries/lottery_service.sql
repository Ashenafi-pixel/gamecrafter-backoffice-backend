-- name: CreateLotteryService :one
INSERT INTO lottery_services (id, name, description,client_id,client_secret ,created_at, updated_at,callback_url)
VALUES (gen_random_uuid(), $1, $2,$3,$4 ,NOW(), NOW(),$5)
RETURNING id, name, description, created_at, updated_at;

-- name: GetLotteryServiceByID :one
SELECT id, name, description, created_at, updated_at,callback_url, client_id, client_secret
FROM lottery_services
WHERE id = $1;

-- name: GetLotteryServiceByClientID :one
SELECT id, name, description,client_id,client_id, created_at, updated_at,callback_url
FROM lottery_services
WHERE client_id = $1 AND deleted_at IS NULL;

-- name: UpdateLotteryService :one
UPDATE lottery_services
SET name = $1, description = $2, updated_at = NOW()
WHERE id = $3
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteLotteryService :exec
UPDATE lottery_services
SET deleted_at = NOW()
WHERE id = $1;

-- name: ListLotteryServices :many
SELECT id, name, description, created_at, updated_at
FROM lottery_services
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateLotteryServiceStatus :one
UPDATE lottery_services
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, status, updated_at;

