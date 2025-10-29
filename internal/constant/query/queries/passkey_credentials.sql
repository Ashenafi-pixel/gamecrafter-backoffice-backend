-- name: CreatePasskeyCredential :one
INSERT INTO passkey_credentials (
    user_id,
    credential_id,
    raw_id,
    public_key,
    attestation_object,
    client_data_json,
    counter,
    name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPasskeyCredentialByID :one
SELECT * FROM passkey_credentials 
WHERE credential_id = $1 AND user_id = $2 AND is_active = true;

-- name: GetPasskeyCredentialsByUserID :many
SELECT id, credential_id, name, created_at, last_used_at, is_active 
FROM passkey_credentials 
WHERE user_id = $1 AND is_active = true 
ORDER BY created_at DESC;

-- name: UpdatePasskeyCredentialCounter :exec
UPDATE passkey_credentials 
SET counter = $1, last_used_at = NOW() 
WHERE credential_id = $2 AND user_id = $3;

-- name: DeletePasskeyCredential :exec
UPDATE passkey_credentials 
SET is_active = false 
WHERE credential_id = $1 AND user_id = $2;

-- name: GetPasskeyCredentialForVerification :one
SELECT * FROM passkey_credentials 
WHERE credential_id = $1 AND user_id = $2 AND is_active = true;

-- name: CheckPasskeyExists :one
SELECT EXISTS(
    SELECT 1 FROM passkey_credentials 
    WHERE user_id = $1 AND is_active = true
) as has_passkey;
