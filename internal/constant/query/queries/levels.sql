-- name: CreateLevel :one 
INSERT INTO levels (level, created_by, type)
VALUES ($1, $2,$3) RETURNING *;

-- name: GetLevels :many 
WITH level_data AS (
    SELECT *
    FROM levels
    WHERE deleted_at IS NULL and type = $1
), row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM level_data
)
SELECT l.*, r.total_rows
FROM level_data l
CROSS JOIN row_count r
ORDER BY l.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateLevel :one
UPDATE levels
SET level = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: DeleteLevel :one
UPDATE levels
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetLevelById :one
SELECT id, level, created_at, updated_at, deleted_at, created_by
FROM levels 
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateLevelRequirement :one
INSERT INTO level_requirements (level_id, type, value, created_by)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetLevelRequirementsByLevelID :many
WITH requirement_data AS (
    SELECT *
    FROM level_requirements
    WHERE deleted_at IS NULL AND level_id = $1
), row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM requirement_data
)
SELECT l.*, r.total_rows
FROM requirement_data l
CROSS JOIN row_count r
ORDER BY l.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLevelRequirements :many
WITH requirement_data AS (
    SELECT *
    FROM level_requirements
    WHERE deleted_at IS NULL
), row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM requirement_data
)
SELECT l.*, r.total_rows
FROM requirement_data l
CROSS JOIN row_count r
ORDER BY l.created_at DESC
LIMIT $1 OFFSET $2;


-- name: UpdateLevelRequirement :one
UPDATE level_requirements
SET type = $1, value = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
RETURNING *;

-- name: DeleteLevelRequirement :one
UPDATE level_requirements
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: GetAallRequirementsByLevelID :many
WITH requirement_data AS (
    SELECT *
    FROM level_requirements
    WHERE deleted_at IS NULL AND level_id = $1
), row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM requirement_data
)
SELECT l.*, r.total_rows
FROM requirement_data l
CROSS JOIN row_count r
ORDER BY l.created_at DESC;

-- name: GetAllLevels :many
WITH level_data AS (
    SELECT *
    FROM levels
    WHERE deleted_at IS NULL and type = $1
), row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM level_data
)
SELECT l.*, r.total_rows
FROM level_data l
CROSS JOIN row_count r
ORDER BY l.created_at DESC; 

-- name: CalculateUserBets :one 
SELECT COALESCE(SUM(change_amount), 0)::decimal AS total_bet_amount
FROM balance_logs
WHERE user_id = $1
AND operational_type_id = (SELECT id FROM operational_types WHERE name = 'place_bet' LIMIT 1)
AND currency = 'P';
-- name: AddFakeBalanceLog :one
INSERT INTO balance_logs (user_id, change_amount, currency, operational_type_id,component)
VALUES ($1, $2, $3, (SELECT id FROM operational_types WHERE name = 'place_bet'), 'real_money')
RETURNING *;

-- name: CalculateSquadBets :one
SELECT COALESCE(SUM(change_amount), 0)::decimal AS total_bet_amount
FROM balance_logs
WHERE user_id in (SELECT user_id FROM squads_memebers WHERE squad_id = $1 AND deleted_at IS NULL) OR user_id in (SELECT owner FROM squads WHERE id = $1 AND deleted_at IS NULL)
AND operational_type_id = (SELECT id FROM operational_types WHERE name = 'place_bet' LIMIT 1)
AND currency = 'P' and timestamp >= (SELECT created_at FROM squads WHERE id = $1 AND deleted_at IS NULL);

-- name: GetUserSquads :many
SELECT s.id as squad_id from squads s JOIN squads_memebers sm ON s.id = sm.squad_id
WHERE sm.user_id = $1 AND sm.deleted_at IS NULL;

-- name: GetAllSquadMembersBySquadId :many
WITH squad_members AS (
    SELECT 
        COUNT(*) OVER () AS total,
        sm.id,
        sm.squad_id,
        sm.user_id,
        sm.created_at,
        sm.updated_at,
        sm.deleted_at,
        u.username,
        u.first_name,
        u.last_name,
        u.phone_number
    FROM squads_memebers sm
    JOIN users u ON sm.user_id = u.id
    WHERE sm.squad_id = $1 AND sm.deleted_at IS NULL
)
SELECT *
FROM squad_members
ORDER BY created_at DESC;

-- name: GetSquadIDByOwnerID :one
select id from squads where owner = $1 and deleted_at is null;