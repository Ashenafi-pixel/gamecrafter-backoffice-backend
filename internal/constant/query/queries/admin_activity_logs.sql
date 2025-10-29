-- Admin Activity Logs SQL Queries

-- name: CreateAdminActivityLog :one
INSERT INTO admin_activity_logs (
    admin_user_id, action, resource_type, resource_id, description, 
    details, ip_address, user_agent, session_id, severity, category
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetAdminActivityLogs :many
WITH admin_info AS (
    SELECT 
        u.id,
        u.username,
        u.email
    FROM users u
    WHERE u.id = ANY($1::uuid[])
),
logs_data AS (
    SELECT 
        aal.id,
        aal.admin_user_id,
        aal.action,
        aal.resource_type,
        aal.resource_id,
        aal.description,
        aal.details,
        aal.ip_address,
        aal.user_agent,
        aal.session_id,
        aal.severity,
        aal.category,
        aal.created_at,
        aal.updated_at
    FROM admin_activity_logs aal
    WHERE 
        ($2::uuid IS NULL OR aal.admin_user_id = $2) AND
        ($3::text = '' OR aal.action = $3) AND
        ($4::text = '' OR aal.resource_type = $4) AND
        ($5::uuid IS NULL OR aal.resource_id = $5) AND
        ($6::text = '' OR aal.category = $6) AND
        ($7::text = '' OR aal.severity = $7) AND
        ($8::timestamp IS NULL OR aal.created_at >= $8) AND
        ($9::timestamp IS NULL OR aal.created_at <= $9) AND
        ($10::text = '' OR aal.description ILIKE '%' || $10 || '%')
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM logs_data
)
SELECT 
    l.id,
    l.admin_user_id,
    l.action,
    l.resource_type,
    l.resource_id,
    l.description,
    l.details,
    l.ip_address,
    l.user_agent,
    l.session_id,
    l.severity,
    l.category,
    l.created_at,
    l.updated_at,
    COALESCE(ai.username, 'Unknown') AS admin_username,
    COALESCE(ai.email, '') AS admin_email,
    r.total_rows
FROM logs_data l
CROSS JOIN row_count r
LEFT JOIN admin_info ai ON ai.id = l.admin_user_id
ORDER BY 
    CASE WHEN $11 = 'created_at' AND $12 = 'asc' THEN l.created_at END ASC,
    CASE WHEN $11 = 'created_at' AND $12 = 'desc' THEN l.created_at END DESC,
    CASE WHEN $11 = 'action' AND $12 = 'asc' THEN l.action END ASC,
    CASE WHEN $11 = 'action' AND $12 = 'desc' THEN l.action END DESC,
    CASE WHEN $11 = 'category' AND $12 = 'asc' THEN l.category END ASC,
    CASE WHEN $11 = 'category' AND $12 = 'desc' THEN l.category END DESC,
    CASE WHEN $11 = 'severity' AND $12 = 'asc' THEN l.severity END ASC,
    CASE WHEN $11 = 'severity' AND $12 = 'desc' THEN l.severity END DESC,
    l.created_at DESC
LIMIT $13 OFFSET $14;

-- name: GetAdminActivityLogByID :one
SELECT 
    aal.id,
    aal.admin_user_id,
    aal.action,
    aal.resource_type,
    aal.resource_id,
    aal.description,
    aal.details,
    aal.ip_address,
    aal.user_agent,
    aal.session_id,
    aal.severity,
    aal.category,
    aal.created_at,
    aal.updated_at,
    u.username AS admin_username,
    u.email AS admin_email
FROM admin_activity_logs aal
LEFT JOIN users u ON u.id = aal.admin_user_id
WHERE aal.id = $1;

-- name: GetAdminActivityStats :one
WITH activity_stats AS (
    SELECT 
        COUNT(*) as total_activities,
        COUNT(CASE WHEN category = 'user_management' THEN 1 END) as user_management_count,
        COUNT(CASE WHEN category = 'financial' THEN 1 END) as financial_count,
        COUNT(CASE WHEN category = 'security' THEN 1 END) as security_count,
        COUNT(CASE WHEN category = 'system' THEN 1 END) as system_count,
        COUNT(CASE WHEN category = 'withdrawal' THEN 1 END) as withdrawal_count,
        COUNT(CASE WHEN category = 'game_management' THEN 1 END) as game_management_count,
        COUNT(CASE WHEN category = 'reports' THEN 1 END) as reports_count,
        COUNT(CASE WHEN category = 'notifications' THEN 1 END) as notifications_count,
        COUNT(CASE WHEN severity = 'low' THEN 1 END) as low_severity_count,
        COUNT(CASE WHEN severity = 'info' THEN 1 END) as info_severity_count,
        COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning_severity_count,
        COUNT(CASE WHEN severity = 'error' THEN 1 END) as error_severity_count,
        COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_severity_count
    FROM admin_activity_logs aal
    WHERE 
        ($1::timestamp IS NULL OR aal.created_at >= $1) AND
        ($2::timestamp IS NULL OR aal.created_at <= $2)
),
recent_activities AS (
    SELECT 
        aal.id,
        aal.admin_user_id,
        aal.action,
        aal.resource_type,
        aal.resource_id,
        aal.description,
        aal.details,
        aal.ip_address,
        aal.user_agent,
        aal.session_id,
        aal.severity,
        aal.category,
        aal.created_at,
        aal.updated_at,
        u.username AS admin_username,
        u.email AS admin_email
    FROM admin_activity_logs aal
    LEFT JOIN users u ON u.id = aal.admin_user_id
    WHERE 
        ($1::timestamp IS NULL OR aal.created_at >= $1) AND
        ($2::timestamp IS NULL OR aal.created_at <= $2)
    ORDER BY aal.created_at DESC
    LIMIT 10
),
top_admins AS (
    SELECT 
        aal.admin_user_id,
        u.username AS admin_username,
        u.email AS admin_email,
        COUNT(*) as activity_count
    FROM admin_activity_logs aal
    LEFT JOIN users u ON u.id = aal.admin_user_id
    WHERE 
        ($1::timestamp IS NULL OR aal.created_at >= $1) AND
        ($2::timestamp IS NULL OR aal.created_at <= $2)
    GROUP BY aal.admin_user_id, u.username, u.email
    ORDER BY activity_count DESC
    LIMIT 10
)
SELECT 
    s.total_activities,
    s.user_management_count,
    s.financial_count,
    s.security_count,
    s.system_count,
    s.withdrawal_count,
    s.game_management_count,
    s.reports_count,
    s.notifications_count,
    s.low_severity_count,
    s.info_severity_count,
    s.warning_severity_count,
    s.error_severity_count,
    s.critical_severity_count
FROM activity_stats s;

-- name: GetAdminActivityCategories :many
SELECT 
    id,
    name,
    description,
    color,
    icon,
    is_active,
    created_at
FROM admin_activity_categories
WHERE is_active = true
ORDER BY name;

-- name: GetAdminActivityActions :many
SELECT 
    aaa.id,
    aaa.name,
    aaa.description,
    aaa.category_id,
    aaa.is_active,
    aaa.created_at,
    aac.name as category_name,
    aac.color as category_color,
    aac.icon as category_icon
FROM admin_activity_actions aaa
LEFT JOIN admin_activity_categories aac ON aac.id = aaa.category_id
WHERE aaa.is_active = true
ORDER BY aac.name, aaa.name;

-- name: GetAdminActivityActionsByCategory :many
SELECT 
    aaa.id,
    aaa.name,
    aaa.description,
    aaa.category_id,
    aaa.is_active,
    aaa.created_at,
    aac.name as category_name,
    aac.color as category_color,
    aac.icon as category_icon
FROM admin_activity_actions aaa
LEFT JOIN admin_activity_categories aac ON aac.id = aaa.category_id
WHERE aaa.is_active = true AND aac.name = $1
ORDER BY aaa.name;

-- name: DeleteAdminActivityLog :exec
DELETE FROM admin_activity_logs WHERE id = $1;

-- name: DeleteAdminActivityLogsByAdmin :exec
DELETE FROM admin_activity_logs WHERE admin_user_id = $1;

-- name: DeleteOldAdminActivityLogs :exec
DELETE FROM admin_activity_logs WHERE created_at < $1;
