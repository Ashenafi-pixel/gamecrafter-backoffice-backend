-- name: GetWithdrawalPauseSettings :one
SELECT id, is_globally_paused, pause_reason, paused_by, paused_at, created_at, updated_at
FROM withdrawal_pause_settings
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateWithdrawalPauseSettings :exec
UPDATE withdrawal_pause_settings
SET 
    is_globally_paused = $2,
    pause_reason = $3,
    paused_by = $4,
    paused_at = CASE WHEN $2 = true THEN NOW() ELSE paused_at END,
    updated_at = NOW()
WHERE id = $1;

-- name: CreateWithdrawalPauseSettings :one
INSERT INTO withdrawal_pause_settings (is_globally_paused, pause_reason, paused_by)
VALUES ($1, $2, $3)
RETURNING id, is_globally_paused, pause_reason, paused_by, paused_at, created_at, updated_at;

-- name: GetWithdrawalThresholds :many
SELECT id, threshold_type, threshold_value, currency, is_active, created_by, created_at, updated_at
FROM withdrawal_thresholds
WHERE is_active = true
ORDER BY threshold_type, currency;

-- name: GetWithdrawalThresholdByType :one
SELECT id, threshold_type, threshold_value, currency, is_active, created_by, created_at, updated_at
FROM withdrawal_thresholds
WHERE threshold_type = $1 AND currency = $2 AND is_active = true;

-- name: CreateWithdrawalThreshold :one
INSERT INTO withdrawal_thresholds (threshold_type, threshold_value, currency, is_active, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, threshold_type, threshold_value, currency, is_active, created_by, created_at, updated_at;

-- name: UpdateWithdrawalThreshold :one
UPDATE withdrawal_thresholds
SET 
    threshold_value = $2,
    is_active = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, threshold_type, threshold_value, currency, is_active, created_by, created_at, updated_at;

-- name: DeleteWithdrawalThreshold :exec
DELETE FROM withdrawal_thresholds WHERE id = $1;

-- name: CreateWithdrawalPauseLog :one
INSERT INTO withdrawal_pause_logs (withdrawal_id, pause_reason, threshold_type, threshold_value, paused_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, withdrawal_id, pause_reason, threshold_type, threshold_value, paused_at, reviewed_by, reviewed_at, action_taken, notes, created_at;

-- name: UpdateWithdrawalPauseLog :one
UPDATE withdrawal_pause_logs
SET 
    reviewed_by = $2,
    reviewed_at = NOW(),
    action_taken = $3,
    notes = $4
WHERE id = $1
RETURNING id, withdrawal_id, pause_reason, threshold_type, threshold_value, paused_at, reviewed_by, reviewed_at, action_taken, notes, created_at;

-- name: GetPausedWithdrawals :many
SELECT 
    w.id, w.user_id, w.withdrawal_id, w.usd_amount_cents, w.crypto_currency, 
    w.status, w.is_paused, w.pause_reason, w.paused_at, w.requires_manual_review,
    w.created_at, w.updated_at,
    u.username, u.email,
    COUNT(*) OVER() AS total
FROM withdrawals w
LEFT JOIN users u ON w.user_id = u.id
WHERE w.is_paused = true
    AND ($1::text IS NULL OR w.status = $1)
    AND ($2::text IS NULL OR w.pause_reason = $2)
    AND ($3::uuid IS NULL OR w.user_id = $3)
ORDER BY w.paused_at DESC
LIMIT $4 OFFSET $5;

-- name: GetWithdrawalPauseLogs :many
SELECT id, withdrawal_id, pause_reason, threshold_type, threshold_value, paused_at, reviewed_by, reviewed_at, action_taken, notes, created_at
FROM withdrawal_pause_logs
WHERE withdrawal_id = $1
ORDER BY created_at DESC;

-- name: UpdateWithdrawalPauseStatus :exec
UPDATE withdrawals
SET 
    is_paused = $2,
    pause_reason = $3,
    paused_at = CASE WHEN $2 = true THEN NOW() ELSE NULL END,
    requires_manual_review = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: GetWithdrawalPauseStats :one
SELECT 
    COUNT(*) FILTER (WHERE w.is_paused = true AND DATE(w.created_at) = CURRENT_DATE) as total_paused_today,
    COUNT(*) FILTER (WHERE w.is_paused = true AND w.created_at >= NOW() - INTERVAL '1 hour') as total_paused_this_hour,
    COUNT(*) FILTER (WHERE w.is_paused = true AND w.requires_manual_review = true) as pending_review,
    COUNT(*) FILTER (WHERE wlp.action_taken = 'approved' AND DATE(wlp.reviewed_at) = CURRENT_DATE) as approved_today,
    COUNT(*) FILTER (WHERE wlp.action_taken = 'rejected' AND DATE(wlp.reviewed_at) = CURRENT_DATE) as rejected_today,
    COALESCE(AVG(EXTRACT(EPOCH FROM (wlp.reviewed_at - wlp.paused_at))/60), 0) as average_review_time_minutes
FROM withdrawals w
LEFT JOIN withdrawal_pause_logs wlp ON w.id = wlp.withdrawal_id;

-- name: GetHourlyWithdrawalVolume :one
SELECT COALESCE(SUM(usd_amount_cents), 0) as hourly_volume_cents
FROM withdrawals
WHERE created_at >= NOW() - INTERVAL '1 hour'
    AND status IN ('completed', 'processing');

-- name: GetDailyWithdrawalVolume :one
SELECT COALESCE(SUM(usd_amount_cents), 0) as daily_volume_cents
FROM withdrawals
WHERE created_at >= CURRENT_DATE
    AND status IN ('completed', 'processing');

-- name: GetUserDailyWithdrawalVolume :one
SELECT COALESCE(SUM(usd_amount_cents), 0) as user_daily_volume_cents
FROM withdrawals
WHERE user_id = $1
    AND created_at >= CURRENT_DATE
    AND status IN ('completed', 'processing');





