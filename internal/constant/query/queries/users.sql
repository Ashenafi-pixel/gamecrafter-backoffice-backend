-- name: CreateUser :one 
INSERT INTO users (username,phone_number,password,default_currency,email,source,referal_code,date_of_birth,created_by,is_admin,first_name,last_name,referal_type,refered_by_code,user_type,status,street_address,country,state,city,postal_code,kyc_status,profile) 
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24) RETURNING *;

-- name: GetUserByUserName :one 
SELECT * FROM users where username = $1;

-- name: GetUserByPhone :one 
SELECT * FROM users where phone_number = $1;

-- name: GetUserByID :one 
SELECT * FROM users where id = $1;

-- name: UpdateProfilePicuter :one 
UPDATE users set profile = $1  where id = $2 RETURNING *;

-- name: UpdatePassword :one 
Update users set password = $1 where id = $2 RETURNING *;

-- name: GetUserByEmail :one 
SELECT id, username, phone_number, password, created_at, default_currency, profile, email, first_name, last_name, date_of_birth, source, is_email_verified, referal_code, street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, status, referal_type, refered_by_code, user_type, primary_wallet_address, wallet_verification_status FROM users where email = $1;

-- name: GetUserByEmailFull :one 
SELECT id, username, phone_number, password, created_at, default_currency, profile, email, first_name, last_name, date_of_birth, source, is_email_verified, referal_code, street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, status, referal_type, refered_by_code, user_type, primary_wallet_address, wallet_verification_status FROM users where email = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: SaveOTP :exec 
INSERT INTO users_otp (user_id,otp,created_at)
VALUES ($1,$2,$3);

-- name: GetOTP :one 
SELECT * FROM users_otp where user_id = $1;

-- name: DeleteOTP :exec 
DELETE from users_otp where user_id = $1;

-- name: UpdateProfile :one 
UPDATE users set first_name=$1,last_name = $2,email=$3,date_of_birth=$4,phone_number=$5,username = $6,street_address = $7,city = $8,postal_code = $9,state = $10,country = $11,kyc_status=$12,status=$13,is_email_verified=$14,default_currency=$15,wallet_verification_status=$16 where id = $17
RETURNING *;

-- name: GetUsersByDepartmentNotificationTypes :many 
SELECT us.* 
FROM users us
JOIN departements_users dus ON us.id = dus.user_id
JOIN departments dp ON dp.id = dus.department_id
WHERE $1::VARCHAR = ANY(dp.notifications);

-- name: GetAllUsers :many 
WITH users_data AS (
    SELECT *
    FROM  users where default_currency  is not null AND user_type = 'PLAYER' AND is_admin = false
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM users_data
)
SELECT c.*, r.total_rows
FROM users_data c
CROSS JOIN row_count r
ORDER BY c.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAllUsersWithFilters :many 
WITH users_data AS (
    SELECT *
    FROM users 
    WHERE default_currency IS NOT NULL 
    AND user_type = 'PLAYER'
    AND is_admin = false
    AND (
        -- Simple OR search across username and email using single searchterm parameter
        ($1::text IS NULL OR $1 = '' OR $1 = '%%' OR username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
    )
    AND ($2::text[] IS NULL OR array_length($2, 1) IS NULL OR array_length($2, 1) = 0 OR status = ANY($2))
    AND ($3::text[] IS NULL OR array_length($3, 1) IS NULL OR array_length($3, 1) = 0 OR kyc_status = ANY($3))
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM users_data
)
SELECT c.*, r.total_rows
FROM users_data c
CROSS JOIN row_count r
ORDER BY c.created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetUserPointsByReferals :one 
SELECT real_money,user_id from balances where user_id = (select id from users where referal_code = $1 limit 1) and currency = $2;


-- name: GetUsersDoseNotHaveReferalCode :many 
select * from users where referal_code = '';

-- name: AddReferalCode :exec 
UPDATE users set referal_code = $1 where id = $2;

-- name: GetUserByReferalCode :one 
SELECT * FROM users where referal_code  = $1;

-- name: GetUserReferalUsersByUserID :many 
select description,change_amount,timestamp from balance_logs where user_id =$1 and currency = 'P';

-- name: GetUsersInArrayOfUserIDs :many 
SELECT username, created_at ,id
FROM users 
WHERE id = ANY($1::UUID[]);

-- name: GetAddminAssignedPoints :many 
SELECT 
    user_id,
    change_amount,
    timestamp,
    description,
	balance_after_update,
    transaction_id,
    COUNT(id) AS total
FROM 
    balance_logs
WHERE 
    description LIKE 'referal_point_admin_%'
GROUP BY 
timestamp,
    user_id,
    change_amount, 
    description,
	balance_after_update,
    transaction_id
    limit $1 offset $2;

-- name: UpdateUserStatus :one 
UPDATE users SET status = $1 WHERE id = $2 RETURNING *;

-- name: UpdateUserVerificationStatus :one 
UPDATE users SET is_email_verified = $1 WHERE id = $2 RETURNING *;

-- name: GetAdmins :many  
WITH UserData AS (
    SELECT 
        us.id AS user_id,
        us.username,
        us.phone_number,
        us.profile,
        us.status,
        us.email,
        us.first_name,
        us.last_name,
        us.date_of_birth,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'role_id', r.id,
                'name', r.name
            ) ORDER BY us.created_at DESC
        ) AS roles
    FROM users us
    JOIN user_roles ur ON ur.user_id = us.id
    JOIN roles r ON r.id = ur.role_id
    GROUP BY us.id, us.username, us.phone_number, us.profile, us.email, us.first_name, us.last_name, us.date_of_birth
),
TotalCount AS (
    SELECT COUNT(*) AS total
    FROM users us
    JOIN user_roles ur ON ur.user_id = us.id
)
SELECT 
    u.*,
    t.total
FROM UserData u
CROSS JOIN TotalCount t
ORDER BY u.user_id DESC
LIMIT $1
OFFSET $2;


-- name: GetAdminsByStatus :many  
WITH UserData AS (
    SELECT 
        us.id AS user_id,
        us.username,
        us.phone_number,
        us.profile,
        us.status,
        us.email,
        us.first_name,
        us.last_name,
        us.date_of_birth,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'role_id', r.id,
                'name', r.name
            ) ORDER BY us.created_at DESC
        ) AS roles
    FROM users us
    JOIN user_roles ur ON ur.user_id = us.id
    JOIN roles r ON r.id = ur.role_id where us.status = $1
    GROUP BY us.id, us.username, us.phone_number, us.profile, us.email, us.first_name, us.last_name, us.date_of_birth
),
TotalCount AS (
    SELECT COUNT(*) AS total
    FROM users us
    JOIN user_roles ur ON ur.user_id = us.id
)
SELECT 
    u.*,
    t.total
FROM UserData u
CROSS JOIN TotalCount t
ORDER BY u.user_id DESC
LIMIT $2
OFFSET $3;

-- name: GetAdminsByRole :many  
WITH admin_data AS (
    select * from user_roles ur join users us on ur.user_id = us.id where role_id = $1
),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM admin_data
)
SELECT c.*, r.total_rows
FROM admin_data c
CROSS JOIN row_count r
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAdminsByRoleAndStatus :many  
WITH admin_data AS (
    select * from user_roles ur join users us on ur.user_id = us.id where role_id = $1 and status = $2 ),
row_count AS (
    SELECT COUNT(*) AS total_rows
    FROM admin_data
)
SELECT c.*, r.total_rows
FROM admin_data c
CROSS JOIN row_count r
ORDER BY c.created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetUserEmailOrPhoneNumber :many
WITH user_data AS (
    SELECT *,count(*) OVER() AS total_rows
    FROM users
    WHERE (phone_number ILIKE '%' || $1 || '%' OR $1 IS NULL)
      AND (email ILIKE '%' || $2 || '%' OR $2 IS NULL)
)
SELECT c.* FROM user_data c
ORDER BY c.created_at DESC
LIMIT $3 OFFSET $4;
