-- name: CreateTemp :one
INSERT INTO temp (id, user_id, data, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
RETURNING id, user_id, data, created_at, updated_at;

-- name: GetTempByUserID :one
SELECT id, user_id, data, created_at, updated_at
FROM temp
WHERE user_id = $1;

-- name: DeleteTempByID :exec
DELETE FROM temp
WHERE id = $1;