-- name: CreateLoginAttemptsLog :one 
INSERT INTO login_attempts(user_id,ip_address,success,attempt_time,user_agent)
VALUES (
    $1,$2,$3,$4,$5
) RETURNING *;

-- name: GetLoginAttempts :many 
SELECT * FROM login_attempts OFFSET $1 LIMIT $2 ;

-- name: GetLoginAttemptsByUserID :many 
SELECT * from login_attempts where user_id = $1 OFFSET $2 LIMIT $3;

-- name: CreateUserSessions :one 
INSERT INTO user_sessions (user_id,token,expires_at,ip_address,user_agent,created_at,refresh_token,refresh_token_expires_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: DeleteBalanceLog :exec 
DELETE FROM bets where id = $1;

-- name: CreateSystemLog :one 
INSERT INTO logs (user_id,module,detail,ip_address,timestamp) VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: GetSystemLogs :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $1 OFFSET $2;

-- name: GetSystemLogsByModule :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs where module = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetSystemLogsByUserID :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs where logs.user_id = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetSystemLogsByStartData :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs where timestamp > $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetSystemLogsByEndData :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs where timestamp < $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetSystemLogsByStartAndEndData :many 
WITH logs_data AS (
    SELECT id, user_id, module, ip_address, timestamp, detail
    FROM logs where timestamp > $1 and timestamp < $2
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data 
),
roles_data AS (
    SELECT 
        ur.user_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'name', rs.name,
                    'role_id', rs.id
                )
            ) FILTER (WHERE rs.id IS NOT NULL),
            '[]'::json
        ) AS roles
    FROM user_roles ur
    LEFT JOIN roles rs ON rs.id = ur.role_id
    GROUP BY ur.user_id
)
SELECT 
    c.id,
    c.user_id,
    c.module,
    c.ip_address,
    c.timestamp,
    c.detail,
    r.total_rows,
    COALESCE(rd.roles, '[]'::json) AS roles
FROM logs_data c 
CROSS JOIN row_count r
LEFT JOIN roles_data rd ON rd.user_id = c.user_id
ORDER BY c.timestamp DESC LIMIT $3 OFFSET $4;  

-- name: GetAvailableModule :many
SELECT DISTINCT module FROM logs;

-- name: GetUserSessionByRefreshToken :one
SELECT * FROM user_sessions
WHERE refresh_token = $1
  AND refresh_token_expires_at > NOW();

-- name: UpdateUserSessionRefreshToken :exec
UPDATE user_sessions
SET refresh_token = $2,
    refresh_token_expires_at = $3
WHERE id = $1;

-- name: InvalidateOldUserSessions :exec
UPDATE user_sessions
SET expires_at = NOW()
WHERE user_id = $1 AND id != $2;

-- name: GetSessionsExpiringSoon :many
SELECT * FROM user_sessions
WHERE refresh_token_expires_at > NOW()
  AND refresh_token_expires_at <= $1;

-- name: InvalidateAllUserSessions :exec
UPDATE user_sessions
SET expires_at = NOW(),
    refresh_token_expires_at = NOW()
WHERE user_id = $1;