
-- name: CreateBanner :one
INSERT INTO banners (page, page_url, image_url, headline, tagline, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, page, page_url, image_url, headline, tagline, updated_at;

-- name: GetAllBanners :many
SELECT 
    id, page, page_url, image_url, headline, tagline, updated_at,
    COUNT(*) OVER() AS total
FROM banners
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;

-- name: GetBannerByPage :one
SELECT id, page, page_url, image_url, headline, tagline, updated_at
FROM banners
WHERE page = $1
ORDER BY updated_at DESC
LIMIT 1;

-- name: GetBannerByID :one
SELECT id, page, page_url, image_url, headline, tagline, updated_at
FROM banners
WHERE id = $1;

-- name: UpdateBanner :one
UPDATE banners
SET page_url = $2, image_url = $3, headline = $4, tagline = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, page, page_url, image_url, headline, tagline, updated_at;

-- name: DeleteBanner :exec
DELETE FROM banners
WHERE id = $1;