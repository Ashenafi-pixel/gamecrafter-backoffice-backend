-- name: CreateAgentReferralLink :one
INSERT INTO agent_referrals (
    request_id,
    callback_url
) VALUES (
    $1, $2
) RETURNING *;

-- name: UpdateAgentReferralWithConversion :one
UPDATE agent_referrals 
SET user_id = $2, conversion_type = $3, amount = $4, msisdn = $5
WHERE request_id = $1
RETURNING *;

-- name: GetAgentReferralByRequestID :one
SELECT * FROM agent_referrals 
WHERE request_id = $1;

-- name: GetAgentReferralsByRequestID :many
SELECT * FROM agent_referrals 
WHERE request_id = $1
ORDER BY converted_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAgentReferralsByRequestID :one
SELECT COUNT(*) FROM agent_referrals 
WHERE request_id = $1;

-- name: GetPendingCallbacks :many
SELECT * FROM agent_referrals
WHERE callback_sent = false 
AND callback_attempts < 3
AND user_id IS NOT NULL
AND callback_url IS NOT NULL
AND TRIM(callback_url) <> ''
ORDER BY converted_at ASC
LIMIT $1 OFFSET $2;

-- name: MarkCallbackSent :exec
UPDATE agent_referrals 
SET callback_sent = true 
WHERE id = $1;

-- name: IncrementCallbackAttempts :exec
UPDATE agent_referrals 
SET callback_attempts = callback_attempts + 1 
WHERE id = $1;

-- name: GetReferralStatsByRequestID :one
SELECT 
    COUNT(*) as total_conversions,
    SUM(amount) as total_amount,
    COUNT(DISTINCT user_id) as unique_users
FROM agent_referrals 
WHERE request_id = $1 AND user_id IS NOT NULL;

-- name: GetReferralStatsByConversionType :many
SELECT 
    conversion_type,
    COUNT(*) as total_conversions,
    SUM(amount) as total_amount
FROM agent_referrals 
WHERE request_id = $1 AND user_id IS NOT NULL
GROUP BY conversion_type;

-- name: CheckIfUserAlreadyConverted :one
SELECT EXISTS(
    SELECT 1 FROM agent_referrals 
    WHERE user_id = $1 AND request_id = $2
) as already_converted;

-- name: GetReferralsByUserID :many
SELECT * FROM agent_referrals 
WHERE user_id = $1
ORDER BY converted_at DESC
LIMIT $2 OFFSET $3;

-- name: CountReferralsByUserID :one
SELECT COUNT(*) FROM agent_referrals 
WHERE user_id = $1;

-- name: CreateAgentProvider :one
INSERT INTO agent_providers (
    client_id,
    client_secret,
    status,
    name,
    description,
    callback_url
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetAgentProviderByClientID :one
SELECT * FROM agent_providers WHERE client_id = $1 AND deleted_at IS NULL;

-- name: GetAgentProviderByID :one
SELECT * FROM agent_providers WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAgentProviders :many
SELECT * FROM agent_providers WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateAgentProviderStatus :exec
UPDATE agent_providers SET status = $2, updated_at = NOW() WHERE id = $1;
