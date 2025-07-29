-- name: GetConfigByName :one 
SELECT * FROM configs where name = $1;

-- name: CreateConfig :one
INSERT INTO configs (name,value) values ($1,$2) RETURNING *;

-- name: UpdateConfigs :one 
UPDATE configs set value = $1 where id = $2 RETURNING *;

-- name: UpdateConfigByName :one 
UPDATE configs set value = $1 where name = $2 RETURNING *;

-- name: GetScratchCardConfigs :many 
SELECT name,value, id FROM configs where name in ('scratch_car','scratch_dollar','scratch_crawn','scratch_cent','scratch_diamond','scratch_cup');

-- name: UpdateScratchCardConfigById :one
UPDATE configs set name = $1, value = $2 where id = $3 RETURNING *;