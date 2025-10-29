-- name: CreateCompany :one
INSERT INTO company (
    site_name, support_email, support_phone, maintenance_mode,
    maximum_login_attempt, password_expiry, lockout_duration,
    require_two_factor_authentication, ip_list, created_by
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10
)
RETURNING
    id,
    site_name,
    support_email,
    support_phone,
    maintenance_mode,
    maximum_login_attempt,
    password_expiry,
    lockout_duration,
    require_two_factor_authentication,
    ip_list::text[] AS ip_list,
    created_by,
    created_at,
    updated_at,
    deleted_at;

-- name: GetCompanyByID :one
SELECT
    id,
    site_name,
    support_email,
    support_phone,
    maintenance_mode,
    maximum_login_attempt,
    password_expiry,
    lockout_duration,
    require_two_factor_authentication,
    ip_list::text[] AS ip_list,
    created_by,
    created_at,
    updated_at,
    deleted_at
FROM company
WHERE id = $1;

-- name: GetCompanies :many
SELECT
    id,
    site_name,
    support_email,
    support_phone,
    maintenance_mode,
    maximum_login_attempt,
    password_expiry,
    lockout_duration,
    require_two_factor_authentication,
    ip_list::text[] AS ip_list,
    created_by,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER() AS total
FROM
    company
WHERE
    deleted_at IS NULL
ORDER BY
    site_name
LIMIT $1 OFFSET $2;

-- name: UpdateCompany :one
UPDATE company SET
    site_name = COALESCE($2, site_name),
    support_email = COALESCE($3, support_email),
    support_phone = COALESCE($4, support_phone),
    maintenance_mode = COALESCE($5, maintenance_mode),
    maximum_login_attempt = COALESCE($6, maximum_login_attempt),
    password_expiry = COALESCE($7, password_expiry),
    lockout_duration = COALESCE($8, lockout_duration),
    require_two_factor_authentication = COALESCE($9, require_two_factor_authentication),
    ip_list = COALESCE($10, ip_list)
WHERE id = $1
RETURNING
    id,
    site_name,
    support_email,
    support_phone,
    maintenance_mode,
    maximum_login_attempt,
    password_expiry,
    lockout_duration,
    require_two_factor_authentication,
    ip_list::text[] AS ip_list,
    created_by,
    created_at,
    updated_at,
    deleted_at;

-- name: DeleteCompany :exec
UPDATE company
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: AddIPAddressToCompany :one
UPDATE company
SET ip_list = CASE
    WHEN array_position(ip_list, $2) IS NULL THEN array_append(ip_list, $2)
    ELSE ip_list
END
WHERE id = $1
RETURNING
    id,
    site_name,
    support_email,
    support_phone,
    maintenance_mode,
    maximum_login_attempt,
    password_expiry,
    lockout_duration,
    require_two_factor_authentication,
    ip_list::text[] AS ip_list,
    created_by,
    created_at,
    updated_at,
    deleted_at;