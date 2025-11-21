-- name: GetActiveGlobalRakebackOverride :one
SELECT id, is_active, rakeback_percentage, start_time, end_time, created_by, created_at, updated_by, updated_at
FROM global_rakeback_override
WHERE is_active = true
ORDER BY created_at DESC
LIMIT 1;

-- name: GetGlobalRakebackOverride :one
SELECT id, is_active, rakeback_percentage, start_time, end_time, created_by, created_at, updated_by, updated_at
FROM global_rakeback_override
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateGlobalRakebackOverride :one
INSERT INTO global_rakeback_override (is_active, rakeback_percentage, start_time, end_time, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, is_active, rakeback_percentage, start_time, end_time, created_by, created_at, updated_by, updated_at;

-- name: UpdateGlobalRakebackOverride :one
UPDATE global_rakeback_override
SET 
    is_active = $2,
    rakeback_percentage = $3,
    start_time = $4,
    end_time = $5,
    updated_by = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, is_active, rakeback_percentage, start_time, end_time, created_by, created_at, updated_by, updated_at;

-- name: DisableGlobalRakebackOverride :exec
UPDATE global_rakeback_override
SET 
    is_active = false,
    updated_by = $2,
    updated_at = NOW()
WHERE id = $1;

