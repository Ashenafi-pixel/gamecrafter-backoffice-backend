-- name: CreateRole :one 
INSERT INTO roles (name) values ($1) RETURNING *;

-- name: GetPermissionByID :one 
SELECT * from permissions where id = $1;

-- name: AssignPermissionToRole :one 
INSERT INTO role_permissions (role_id,permission_id,value) values ($1,$2,$3) RETURNING *;

-- name: GetRoleByName :one 
SELECT * FROM roles where name = $1;

-- name: RemoveRole :exec 
DELETE FROM roles where id = $1;

-- name: RemoveRolesPermissions :exec 
DELETE FROM role_permissions where id = $1;

-- name: RemoveRolesPermissionByRoleID :exec 
DELETE FROM role_permissions where role_id = $1;

-- name: GetPermissions :many 
SELECT * FROM permissions where true limit $1 offset $2;

-- name: GetRoles :many 
SELECT * FROM roles where true limit $1 offset $2;

-- name: GetRolePermissions :many 
SELECT * FROM permissions where id in (select permission_id from role_permissions where role_id = $1);

-- name: GetRoleByID :one 
SELECT * FROM roles where id = $1;

-- name: GetRolePermissionsForRole :many
SELECT * FROM role_permissions where role_id = $1;

-- name: GetAdminFundingLimit :one
SELECT COALESCE(MAX(rp.value), NULL) as max_funding_limit
FROM user_roles ur
JOIN role_permissions rp ON ur.role_id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE ur.user_id = $1 AND p.name = 'manual funding';

-- name: AddRoleToUser :one
INSERT INTO user_roles (user_id,role_id) VALUES ( $1,$2) RETURNING *;

-- name: GetUserRoles :many
SELECT * FROM user_roles where user_id = $1;

-- name: GetUserRoleByUserIDandRoleID :one 
SELECT * FROM user_roles where role_id = $1 and user_id = $2;

-- name: RemoveRoleFromUserRoles :exec 
DELETE FROM user_roles where role_id = $1;

-- name: RevokeUserRole :exec 
DELETE FROM  user_roles where user_id = $1 and role_id = $2;

-- name: RemoveAllUserRolesExceptSuper :exec
DELETE FROM user_roles 
WHERE user_id = $1 
AND role_id NOT IN (SELECT id FROM roles WHERE name = 'super');

-- name: GetRoleUsers :many 
SELECT u.* from user_roles ur join users u on ur.user_id = u.id where role_id = $1;

-- name: GetSupperAdmin :one
SELECT * FROM roles where name = 'super';

-- name: AddSupperAdminCasbinRule :one 
insert into casbin_rule (ptype,v0,v1,v2) values ('p',$1,'*','*') RETURNING *;

-- name: GetPermissonByName :one 
SELECT * FROM permissions where name = $1;