-- name: CreateSpinningWheel :one
INSERT INTO spinning_wheels (user_id,status,bet_amount,timestamp,won_amount,won_status,type)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: GetSpinningWheelUserHistory :many 
WITH spinning_wheels_data AS (
    SELECT *
    FROM spinning_wheels
    WHERE user_id = $1 and status = 'CLOSED' 
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM spinning_wheels_data
     WHERE user_id = $1 AND status = 'CLOSED'
)
SELECT c.*, r.total_rows
FROM spinning_wheels_data c
CROSS JOIN row_count r
ORDER BY c.timestamp DESC
LIMIT $2 OFFSET $3;


-- name: CreateSpinningWheelMysteries :one 
INSERT INTO spinning_wheel_mysteries (name,amount,type,status,frequency,created_by,created_at,icon) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: GetSpinningWheelMysteries :many
WITH spinning_wheel_mysteries_data AS (
    SELECT *
    FROM spinning_wheel_mysteries
    WHERE deleted_at is null
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM spinning_wheel_mysteries_data
)
SELECT s.id, s.name, s.amount, s.type, s.frequency,s.status, r.total_rows,s.icon
FROM spinning_wheel_mysteries_data s
CROSS JOIN row_count r
ORDER BY s.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetSpinningWheelConfig :many
WITH spinning_wheel_configs_data AS (
    SELECT *
    FROM spinning_wheel_configs where deleted_at is null
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM spinning_wheel_configs_data where deleted_at is null
)
SELECT s.name, s.amount, s.type, s.frequency, r.total_rows
FROM spinning_wheel_configs_data s
CROSS JOIN row_count r
ORDER BY s.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateSpinningWheelMystery :one 
UPDATE spinning_wheel_mysteries set status = $1,updated_at = now(),frequency = $3, amount = $4,name = $5,type = $6 where id = $2 RETURNING *;

-- name: DeleteSpinningWheelMystery :exec 
Update spinning_wheel_mysteries set deleted_at = now() where id = $1;

-- name: CreateSpinningWheelConfig :one 
INSERT INTO spinning_wheel_configs (name,amount,type,frequency,created_by,created_at,icon, color) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: DeleteSpinningWheelConfig :exec 
Update spinning_wheel_configs set deleted_at = now() where id = $1;

-- name: GetSpinningWheelConfigs :many
WITH spinning_wheel_configs_data AS (
    SELECT *
    FROM spinning_wheel_configs where deleted_at is null
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM spinning_wheel_configs_data  where deleted_at is null
)
SELECT s.id, s.name, s.amount, s.type, s.frequency, r.total_rows,s.icon, s.color
FROM spinning_wheel_configs_data s
CROSS JOIN row_count r
ORDER BY s.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAllSpinningWheelConfigs :many
SELECT id,name,amount,type,frequency,status,icon FROM spinning_wheel_configs where deleted_at is null;

-- name: GetAllSpinningWheelMysteries :many
SELECT id,name,amount,type,frequency,status,icon FROM spinning_wheel_mysteries where deleted_at is null;

-- name: UpdateSpinningWheelConfig :one 
UPDATE spinning_wheel_configs set status = $1,updated_at = now(),frequency = $3, amount = $4,name = $5,type = $6 where id = $2 RETURNING *;
