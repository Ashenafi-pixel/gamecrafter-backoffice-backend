-- Campaign Management Queries

-- name: CreateCampaign :one
INSERT INTO message_campaigns (
    title, message_type, subject, content, created_by, scheduled_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING
    id,
    title,
    message_type,
    subject,
    content,
    created_by,
    status,
    scheduled_at,
    sent_at,
    total_recipients,
    delivered_count,
    read_count,
    created_at,
    updated_at;

-- name: GetCampaigns :many
SELECT
    id,
    title,
    message_type,
    subject,
    content,
    created_by,
    status,
    scheduled_at,
    sent_at,
    total_recipients,
    delivered_count,
    read_count,
    created_at,
    updated_at,
    COUNT(*) OVER() AS total
FROM message_campaigns
WHERE 
    ($1::uuid IS NULL OR created_by = $1) AND
    ($2::text IS NULL OR status = $2) AND
    ($3::text IS NULL OR message_type = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetCampaignByID :one
SELECT
    id,
    title,
    message_type,
    subject,
    content,
    created_by,
    status,
    scheduled_at,
    sent_at,
    total_recipients,
    delivered_count,
    read_count,
    created_at,
    updated_at
FROM message_campaigns
WHERE id = $1;

-- name: UpdateCampaign :one
UPDATE message_campaigns
SET 
    title = COALESCE($2, title),
    subject = COALESCE($3, subject),
    content = COALESCE($4, content),
    scheduled_at = COALESCE($5, scheduled_at),
    status = COALESCE($6, status),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    title,
    message_type,
    subject,
    content,
    created_by,
    status,
    scheduled_at,
    sent_at,
    total_recipients,
    delivered_count,
    read_count,
    created_at,
    updated_at;

-- name: DeleteCampaign :exec
DELETE FROM message_campaigns WHERE id = $1;

-- name: UpdateCampaignStats :exec
UPDATE message_campaigns
SET 
    total_recipients = $2,
    delivered_count = $3,
    read_count = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkCampaignAsSent :exec
UPDATE message_campaigns
SET 
    status = 'sent',
    sent_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- Segment Queries

-- name: CreateSegment :one
INSERT INTO message_segments (
    campaign_id, segment_type, segment_name, criteria, csv_data, user_count
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING
    id,
    campaign_id,
    segment_type,
    segment_name,
    criteria,
    csv_data,
    user_count,
    created_at;

-- name: GetSegmentsByCampaign :many
SELECT
    id,
    campaign_id,
    segment_type,
    segment_name,
    criteria,
    csv_data,
    user_count,
    created_at
FROM message_segments
WHERE campaign_id = $1
ORDER BY created_at;

-- name: UpdateSegmentUserCount :exec
UPDATE message_segments
SET user_count = $2
WHERE id = $1;

-- Recipient Queries

-- name: CreateCampaignRecipient :one
INSERT INTO campaign_recipients (
    campaign_id, user_id, notification_id, status
) VALUES (
    $1, $2, $3, $4
) RETURNING
    id,
    campaign_id,
    user_id,
    notification_id,
    status,
    sent_at,
    delivered_at,
    read_at,
    error_message,
    created_at;

-- name: GetCampaignRecipients :many
SELECT
    cr.id,
    cr.campaign_id,
    cr.user_id,
    cr.notification_id,
    cr.status,
    cr.sent_at,
    cr.delivered_at,
    cr.read_at,
    cr.error_message,
    cr.created_at,
    u.username,
    u.email,
    COUNT(*) OVER() AS total
FROM campaign_recipients cr
JOIN users u ON cr.user_id = u.id
WHERE 
    cr.campaign_id = $1 AND
    ($2::text IS NULL OR cr.status = $2)
ORDER BY cr.created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateRecipientStatus :exec
UPDATE campaign_recipients
SET 
    status = $2,
    sent_at = CASE WHEN $2 = 'sent' THEN NOW() ELSE sent_at END,
    delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END,
    read_at = CASE WHEN $2 = 'read' THEN NOW() ELSE read_at END,
    error_message = CASE WHEN $2 = 'failed' THEN $3 ELSE error_message END
WHERE id = $1;

-- name: GetCampaignStats :one
SELECT
    campaign_id,
    COUNT(*) as total_recipients,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_count,
    COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered_count,
    COUNT(CASE WHEN status = 'read' THEN 1 END) as read_count,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
FROM campaign_recipients
WHERE campaign_id = $1
GROUP BY campaign_id;

-- User Segmentation Queries

-- name: GetUsersByCriteria :many
SELECT DISTINCT
    u.id,
    u.username,
    u.email,
    u.created_at,
    u.kyc_status,
    u.country,
    u.currency,
    w.balance
FROM users u
LEFT JOIN wallets w ON u.id = w.user_id AND w.currency = u.currency
WHERE 
    ($1::int IS NULL OR u.created_at >= NOW() - INTERVAL '%d days' * $1) AND
    ($2::text IS NULL OR u.kyc_status = $2) AND
    ($3::text IS NULL OR u.country = $3) AND
    ($4::text IS NULL OR u.currency = $4) AND
    ($5::decimal IS NULL OR w.balance >= $5) AND
    ($6::decimal IS NULL OR w.balance <= $6)
ORDER BY u.created_at DESC;

-- name: GetUsersByCSV :many
SELECT DISTINCT
    u.id,
    u.username,
    u.email,
    u.created_at,
    u.kyc_status,
    u.country,
    u.currency,
    w.balance
FROM users u
LEFT JOIN wallets w ON u.id = w.user_id AND w.currency = u.currency
WHERE u.username = ANY($1::text[])
ORDER BY u.created_at DESC;

-- name: GetUserActivityCount :one
SELECT COUNT(*)
FROM user_activity_log
WHERE user_id = $1 AND created_at >= $2;

-- name: LogUserActivity :exec
INSERT INTO user_activity_log (user_id, activity_type, activity_data)
VALUES ($1, $2, $3);

-- name: GetScheduledCampaigns :many
SELECT
    id,
    title,
    message_type,
    subject,
    content,
    created_by,
    status,
    scheduled_at,
    sent_at,
    total_recipients,
    delivered_count,
    read_count,
    created_at,
    updated_at
FROM message_campaigns
WHERE 
    status = 'scheduled' AND
    scheduled_at <= NOW()
ORDER BY scheduled_at;
