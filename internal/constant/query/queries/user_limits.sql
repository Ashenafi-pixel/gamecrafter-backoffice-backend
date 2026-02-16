-- name: UpsertUserWithdrawalLimit :exec
INSERT INTO user_limits (user_id, limit_type, daily_limit_cents)
VALUES ($1, 'withdrawal', $2)
ON CONFLICT (user_id, limit_type) 
DO UPDATE SET 
    daily_limit_cents = EXCLUDED.daily_limit_cents,
    updated_at = NOW();

-- name: DeleteUserWithdrawalLimit :exec
DELETE FROM user_limits 
WHERE user_id = $1 AND limit_type = 'withdrawal';

-- name: GetUserWithdrawalLimit :one
SELECT daily_limit_cents 
FROM user_limits 
WHERE user_id = $1 AND limit_type = 'withdrawal';

-- name: GetUserLimitsByUserID :many
SELECT * FROM user_limits WHERE user_id = $1;

