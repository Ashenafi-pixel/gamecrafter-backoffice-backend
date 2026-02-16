-- name: CreateOperationalGroup :one 
 INSERT INTO operational_groups(name,description,created_at)
  values(
    $1,$2,$3
  ) RETURNING *;

-- name: GetOperationalGroups :many 
SELECT * FROM operational_groups;

-- name: GetOperationalGroupByName :one 
SELECT * from operational_groups where name = $1 limit 1;

-- name: CreateOperationalGroupType :one 
INSERT INTO operational_types (group_id,name,description,created_at) values (
 $1,$2,$3,$4
) RETURNING *;

-- name: GetOperationalTtypeByGoupIDandOperationalTypeName :one 
SELECT * FROM operational_types where group_id = $1 and name = $2 limit 1;

-- name: GetOperationalTypesByGroupID :many 
SELECT * FROM operational_types where group_id = $1;

-- name: GetOperationalGroupTypes :many 
SELECT  * FROM operational_types;

-- name: GetOperationalGroupByID :one 
SELECT * FROM operational_groups where id = $1;

-- name: GetOperationalGroupTypeByID :one 
SELECT * FROM operational_types where id = $1;