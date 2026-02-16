-- name: CreateDepartment :one
INSERT INTO departments (name,notifications,created_at) VALUES ($1,$2,$3)
RETURNING *;

-- name: GetDepartementByName :one 
SELECT * FROM departments where name = $1;

-- name: GetDepartementByID :one 
SELECT * FROM departments where id = $1;

-- name: GetAllDepatments :many 
SELECT *,count(*) as total FROM departments where true GROUP BY id limit $1 offset $2 ;

-- name: UpdateDepartments :one
UPDATE departments SET name = $1,notifications = $2 WHERE id = $3 RETURNING *;

-- name: AssignUserToDepartment :one
INSERT INTO departements_users (user_id,department_id) values (
$1,$2
) RETURNING *;

-- name: GetUserDepartmentByID :many
SELECT dep.name, dep.notifications, dep.id AS department_id,
       usr.id AS user_id, usr.first_name, usr.last_name, usr.username, usr.email
FROM departments dep
JOIN departements_users dpu ON dep.id = dpu.department_id
JOIN users usr ON usr.id = dpu.user_id
WHERE usr.id = $1;
