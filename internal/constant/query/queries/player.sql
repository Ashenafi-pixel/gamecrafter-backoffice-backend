-- Player Management Queries

-- name: CreatePlayer :one
INSERT INTO players (
    email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at;

-- name: GetPlayerByID :one
SELECT id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at
FROM players
WHERE id = $1;

-- name: GetPlayerByEmail :one
SELECT id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at
FROM players
WHERE email = $1;

-- name: GetPlayerByUsername :one
SELECT id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at
FROM players
WHERE username = $1;

-- name: GetPlayers :many
SELECT 
    id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at,
    COUNT(*) OVER() AS total
FROM players
WHERE 
    ($1::text IS NULL OR email ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%') AND
    ($2::int IS NULL OR brand_id = $2) AND
    ($3::text IS NULL OR country = $3) AND
    ($4::bool IS NULL OR test_account = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: UpdatePlayer :one
UPDATE players
SET 
    email = COALESCE($2, email),
    username = COALESCE($3, username),
    phone = COALESCE($4, phone),
    first_name = COALESCE($5, first_name),
    last_name = COALESCE($6, last_name),
    default_currency = COALESCE($7, default_currency),
    brand = COALESCE($8, brand),
    date_of_birth = COALESCE($9, date_of_birth),
    country = COALESCE($10, country),
    state = COALESCE($11, state),
    street_address = COALESCE($12, street_address),
    postal_code = COALESCE($13, postal_code),
    test_account = COALESCE($14, test_account),
    enable_withdrawal_limit = COALESCE($15, enable_withdrawal_limit),
    brand_id = COALESCE($16, brand_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at;

-- name: DeletePlayer :exec
DELETE FROM players
WHERE id = $1;

-- name: GetPlayersByBrandID :many
SELECT id, email, username, password, phone, first_name, last_name,
    default_currency, brand, date_of_birth, country, state,
    street_address, postal_code, test_account, enable_withdrawal_limit,
    brand_id, created_at, updated_at
FROM players
WHERE brand_id = $1
ORDER BY created_at DESC;

