-- Provider Management Queries

-- name: CreateProvider :one
INSERT INTO game_providers (name, code, status, created_at, updated_at)
VALUES ($1, $2, COALESCE($3, 'ACTIVE'), NOW(), NOW())
RETURNING id, name, code, status, created_at, updated_at;

