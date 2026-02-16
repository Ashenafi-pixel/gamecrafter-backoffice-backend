
-- name: CreateLootBox :one
INSERT INTO loot_box (type, prizeAmount, weight)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLootBoxByID :one
SELECT id,type,prizeAmount,weight FROM loot_box WHERE id = $1;

-- name: GetAllLootBoxes :many
SELECT id,type,prizeAmount,weight FROM loot_box;

-- name: DeleteLootBoxByID :exec
DELETE FROM loot_box WHERE id = $1;

-- name: GetLootBoxByType :many
SELECT id,type,prizeAmount,weight FROM loot_box WHERE type = $1;

-- name: UpdateLootBox :one
UPDATE loot_box SET
    type = $1,
    prizeAmount = $2,
    weight = $3,
    updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: PlaceLootBoxBet :one
INSERT INTO loot_box_place_bets (user_id,loot_box)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateLootBoxBet :one
UPDATE loot_box_place_bets SET
    user_selection = $1,
    wonStatus = $2,
    status = $3,
    updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: GetLootBoxBetsByUserID :many
WITH user_bets AS (
    SELECT id, user_id, user_selection, loot_box, wonStatus, status, created_at, updated_at
    FROM loot_box_place_bets
    WHERE user_id = $1 ORDER BY created_at DESC
)
SELECT id, user_id, user_selection, loot_box, wonStatus, status, created_at, updated_at
FROM user_bets LIMIT $2 OFFSET $3;

-- name: GetLootBoxBetByID :one
SELECT id, user_id, user_selection, loot_box, wonStatus, status, created_at, updated_at 
FROM loot_box_place_bets
WHERE id = $1 and status = 'pending';