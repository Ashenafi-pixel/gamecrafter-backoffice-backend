-- name: CreateGame :one
INSERT INTO games (id,name) VALUES ($1,$2) RETURNING *;

-- name: GetGames :many 
WITH games_data AS (
    SELECT *
    FROM games where status = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM games_data
)
SELECT c.*, r.total_rows
FROM games_data c
CROSS JOIN row_count r
ORDER BY c.name DESC
LIMIT $2 OFFSET $3;

-- name: UpdateGame :one 
UPDATE games set name = $1 , Status = $2 where id = $3
RETURNING *;

-- name: GetAllGames :many 
SELECT * FROM games ;

-- name: GetGameByID :one 
SELECT * FROM games where id = $1;

-- name: DeleteGame :exec
UPDATE games SET status = 'INACTIVE' WHERE id = $1;

-- name: AddGame :one 
UPDATE games SET status = 'ACTIVE' WHERE id = $1 RETURNING *;

-- name: ChangeEnableStatus :one 
UPDATE games SET enabled = $1 where id = $2 RETURNING *;
