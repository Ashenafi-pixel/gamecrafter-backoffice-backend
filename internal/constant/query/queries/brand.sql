-- Brand Management Queries

-- name: CreateBrand :one
INSERT INTO brands (name, code, domain, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, name, code, domain, is_active, created_at, updated_at;

-- name: GetBrands :many
SELECT 
    id, name, code, domain, is_active, created_at, updated_at,
    COUNT(*) OVER() AS total
FROM brands
WHERE 
    ($1::text IS NULL OR name ILIKE '%' || $1 || '%' OR code ILIKE '%' || $1 || '%') AND
    ($2::bool IS NULL OR is_active = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetBrandByID :one
SELECT id, name, code, domain, is_active, created_at, updated_at
FROM brands
WHERE id = $1;

-- name: UpdateBrand :one
UPDATE brands
SET 
    name = COALESCE($2, name),
    code = COALESCE($3, code),
    domain = COALESCE($4, domain),
    is_active = COALESCE($5, is_active),
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, code, domain, is_active, created_at, updated_at;

-- name: DeleteBrand :exec
DELETE FROM brands
WHERE id = $1;

