-- name: SaveAddsService :one
INSERT INTO adds_services (name, description, status, created_by, service_id, service_secret, service_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetAddsServices :many
select * from adds_services where deleted_at is null ORDER BY created_at DESC LIMIT $1 OFFSET $2 ;

-- name: GetAddsServiceByServiceID :one
SELECT id, name, description, service_id, service_secret, status, created_at, updated_at, created_by, service_url FROM adds_services WHERE service_id = $1 AND deleted_at is null;

-- name: GetAddsServiceByID :one
SELECT id, name, description, service_id, service_secret, status, created_at, updated_at, created_by, service_url FROM adds_services WHERE id = $1 AND deleted_at is null;

-- name: UpdateAddsServiceStatus :one
UPDATE adds_services SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at is null RETURNING *; 