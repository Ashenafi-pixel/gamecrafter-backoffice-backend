-- name: CreateWalletConnection :one
INSERT INTO crypto_wallet_connections (
    user_id, wallet_type, wallet_address, wallet_chain, wallet_name, wallet_icon_url
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetWalletConnectionByAddress :one
SELECT * FROM crypto_wallet_connections 
WHERE wallet_address = $1 AND wallet_type = $2;

-- name: GetWalletConnectionByID :one
SELECT * FROM crypto_wallet_connections 
WHERE id = $1;

-- name: GetUserWalletConnections :many
SELECT * FROM crypto_wallet_connections 
WHERE user_id = $1 
ORDER BY last_used_at DESC;

-- name: UpdateWalletConnection :one
UPDATE crypto_wallet_connections 
SET 
    wallet_name = $2,
    wallet_icon_url = $3,
    is_verified = $4,
    verification_signature = $5,
    verification_message = $6,
    verification_timestamp = $7,
    last_used_at = $8,
    updated_at = NOW()
WHERE id = $1 
RETURNING *;

-- name: DeleteWalletConnection :exec
DELETE FROM crypto_wallet_connections 
WHERE id = $1;

-- name: CreateWalletChallenge :one
INSERT INTO crypto_wallet_challenges (
    wallet_address, wallet_type, challenge_message, challenge_nonce, expires_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetWalletChallenge :one
SELECT * FROM crypto_wallet_challenges 
WHERE wallet_address = $1 AND wallet_type = $2 AND challenge_nonce = $3 AND expires_at > NOW() AND is_used = false;

-- name: MarkChallengeAsUsed :exec
UPDATE crypto_wallet_challenges 
SET is_used = true 
WHERE id = $1;

-- name: CleanExpiredChallenges :exec
DELETE FROM crypto_wallet_challenges 
WHERE expires_at < NOW();

-- name: CreateWalletAuthLog :one
INSERT INTO crypto_wallet_auth_logs (
    wallet_address, wallet_type, action, ip_address, user_agent, success, error_message, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetUserWallets :many
SELECT 
    cwc.id as connection_id,
    cwc.wallet_type,
    cwc.wallet_address,
    cwc.wallet_chain,
    cwc.wallet_name,
    cwc.is_verified,
    CASE WHEN u.primary_wallet_address = cwc.wallet_address THEN true ELSE false END as is_primary,
    cwc.last_used_at,
    cwc.created_at as connected_at
FROM crypto_wallet_connections cwc
WHERE cwc.user_id = $1
ORDER BY cwc.last_used_at DESC;

-- name: GetWalletConnectionWithUser :one
SELECT 
    cwc.id as connection_id,
    cwc.user_id,
    cwc.wallet_type,
    cwc.wallet_address,
    cwc.wallet_chain,
    cwc.wallet_name,
    cwc.is_verified,
    u.phone_number as user_phone,
    u.email as user_email,
    u.first_name as user_first_name,
    u.last_name as user_last_name,
    cwc.last_used_at,
    cwc.created_at as connected_at
FROM crypto_wallet_connections cwc
JOIN users u ON cwc.user_id = u.id
WHERE cwc.wallet_address = $1 AND cwc.wallet_type = $2;

-- name: SetPrimaryWallet :exec
UPDATE users 
SET primary_wallet_address = $2, wallet_verification_status = 'verified'
WHERE id = $1;

-- name: GetUserByWalletAddress :one
SELECT u.* FROM users u
JOIN crypto_wallet_connections cwc ON u.id = cwc.user_id
WHERE cwc.wallet_address = $1 AND cwc.wallet_type = $2;

-- name: CheckWalletExists :one
SELECT EXISTS(
    SELECT 1 FROM crypto_wallet_connections 
    WHERE wallet_address = $1 AND wallet_type = $2
) as exists;

-- name: GetWalletAuthLogs :many
SELECT * FROM crypto_wallet_auth_logs 
WHERE wallet_address = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountUserWallets :one
SELECT COUNT(*) FROM crypto_wallet_connections 
WHERE user_id = $1; 