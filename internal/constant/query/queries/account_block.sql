-- name: BlockAccount :one 
INSERT INTO account_block (user_id,blocked_by,duration,blocked_from,blocked_to,unblocked_at,reason,created_at,type,note) 
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *;

-- name: GetPermamentlyBlockedAccountByUserIdAndType :one 
SELECT * FROM account_block where user_id = $1 and type = $2 and duration = $3 and unblocked_at is null;

-- name: GetAccountBlockByUserID :many
SELECT * FROM account_block where user_id =$1 and unblocked_at is null;

-- name: UnlockAccount :one 
UPDATE account_block set unblocked_at = now() where id = $1 RETURNING *;

-- name: GetBlockedAllAccount :many 
   SELECT 
     ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    ORDER BY ab.id
	Limit $1 offset $2;


-- name: GetBlockedAllAccountByType :many 
   SELECT 
     ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE type = $1
    ORDER BY ab.id
	Limit $2 offset $3;

-- name: GetBlockedAllAccountByDuration :many 
   SELECT 
     ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE duration = $1
    ORDER BY ab.id
	Limit $2 offset $3;


-- name: GetBlockedAllAccountByTypeAndDuration :many 
   SELECT 
     ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE duration = $1 AND type = $2
    ORDER BY ab.id
	Limit $3 offset $4;

-- name: GetBlockedAllAccountByTypeAndDurationAndUserID :many 
   SELECT 
     ab.*, 
    acuser.id AS blocked_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE duration = $1 AND type = $2 and user_id = $3
    ORDER BY ab.id
	Limit $4 offset $5;

-- name: GetBlockedAccountByUserID :many 
   SELECT 
    ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE user_id = $1
    ORDER BY ab.id 
	Limit $2 offset $3;


-- name: GetBlockedAccountByUserIDAndType :many 
   SELECT 
    ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE user_id = $1 and type = $2
    ORDER BY ab.id 
	Limit $3 offset $4;

-- name: GetBlockedAccountByUserIDAndDuration :many 
   SELECT 
    ab.*, 
    acuser.id AS blockde_account_user_id, 
    acuser.phone_number AS blocked_account_user_phone, 
    acuser.email AS blocked_account_user_email,
    acuser.first_name AS blocked_account_user_first_name,
    acuser.last_name AS blocked_account_user_last_name,
    acuser.username AS blocked_account_user_username,

	blockerAcc.id AS blocker_account_user_id, 
    blockerAcc.phone_number AS blocker_account_user_phone, 
    blockerAcc.email AS blocker_account_user_email,
    blockerAcc.first_name AS blocker_account_user_first_name,
    blockerAcc.last_name AS blocker_account_user_last_name,
    blockerAcc.username AS blocker_account_user_username,
    COUNT(*) OVER() AS total

FROM 
    account_block ab
JOIN 
    users acuser 
    ON ab.user_id = acuser.id
JOIN 
    users blockerAcc 
    ON ab.blocked_by = blockerAcc.id
    WHERE user_id = $1 and duration = $2
    ORDER BY ab.id 
	Limit $3 offset $4;

-- name: CreateIPFilter :one
INSERT INTO ip_filters(created_by,start_ip,end_ip,type,created_at,description) VALUES(
    $1,$2,$3,$4,$5,$6
) RETURNING *;

-- name: GetIpFilterByType :many
SELECT *,us.id as user_id,us.first_name,us.last_name,us.email,count(*) as total FROM ip_filters ipf JOIN users us ON ipf.created_by = us.id where type = $1 GROUP BY ipf.id,us.id;

-- name: GetIpFilterByTypeWithLimitAndOffset :many 
SELECT *,us.id as user_id,us.first_name,us.last_name,us.email,count(*) as total  FROM ip_filters ipf JOIN users us ON ipf.created_by = us.id where type = $1 GROUP BY ipf.id,us.id limit $2 offset $3;

-- name: GetAllIpFilterWithLimitAndOffset :many
SELECT *,us.id as user_id,us.first_name,us.last_name,us.email,count(*) as total  FROM ip_filters ipf JOIN users us ON ipf.created_by = us.id where  true GROUP BY ipf.id,us.id limit $1 offset $2;

-- name: RemoveAccountBlock :exec 
DELETE FROM ip_filters where id =$1;

-- name: GetIpFiltersByID :one 
SELECT * FROM ip_filters where id = $1;

-- name: UpdateIPfilter :one 
UPDATE ip_filters  set description = $1,hits = $2 ,last_hit = $3 where id = $4 RETURNING *;