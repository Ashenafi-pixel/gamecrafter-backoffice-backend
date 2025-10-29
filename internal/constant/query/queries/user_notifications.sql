-- name: InsertUserNotification :one
INSERT INTO user_notifications (
    user_id, title, content, type, metadata, read, delivered, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING
    id,
    user_id,
    title,
    content,
    type,
    metadata,
    read,
    delivered,
    created_by,
    read_at,
    created_at;

-- name: GetUserNotifications :many
SELECT
    id,
    user_id,
    title,
    content,
    type,
    metadata,
    read,
    delivered,
    created_by,
    read_at,
    created_at,
    COUNT(*) OVER() AS total
FROM
    user_notifications
WHERE
    user_id = $1
ORDER BY
    created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :exec
UPDATE user_notifications
SET read = TRUE, read_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :execrows
UPDATE user_notifications
SET read = TRUE, read_at = NOW()
WHERE user_id = $1 AND (read = FALSE OR read IS NULL);

-- name: DeleteNotification :exec
DELETE FROM user_notifications
WHERE id = $1 AND user_id = $2;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) FROM user_notifications
WHERE user_id = $1 AND read = FALSE;

-- name: GetAllUnreadNotificationCount :one
SELECT COUNT(*) FROM user_notifications
WHERE read = FALSE;

-- name: GetAllDeliveredNotificationCount :one
SELECT COUNT(*) FROM user_notifications
WHERE delivered = TRUE;

-- name: GetAllReadNotificationCount :one
SELECT COUNT(*) FROM user_notifications
WHERE read = TRUE;

-- name: GetAllNotifications :many
SELECT
    id,
    user_id,
    title,
    content,
    type,
    metadata,
    read,
    delivered,
    created_by,
    read_at,
    created_at,
    COUNT(*) OVER() AS total
FROM
    user_notifications
ORDER BY
    created_at DESC
LIMIT $1 OFFSET $2;
