-- name: CreateSquad :one 
INSERT INTO squads (handle, owner, type, created_at) values ($1, $2, $3, $4)
RETURNING *;

-- name: GetSquadByhandle :one 
SELECT id, handle, owner, type, created_at, updated_at, deleted_at from squads WHERE handle = $1;

-- name: GetSquadById :one
SELECT id, handle, owner, type, created_at, updated_at, deleted_at from squads WHERE id = $1;

-- name: GetSquadByOwner :many
SELECT id, handle, owner, type, created_at, updated_at, deleted_at 
FROM squads 
WHERE owner IN (SELECT squad_id FROM squads_memebers WHERE user_id = $1 AND deleted_at IS NULL)
AND deleted_at IS NULL;

-- name: UpdateSquad :one
UPDATE squads SET handle = $1, type = $2, updated_at = $3
WHERE id = $4 RETURNING *;

-- name: GetSquadByUserID :one 
WITH squad_members AS (
    SELECT 
        s.id,
        s.handle,
        s.owner,
        s.type,
        s.created_at,
        s.updated_at,
        s.deleted_at,
        u.first_name,
        u.last_name,
        u.id as owner_id,
        u.phone_number
    FROM squads s
    JOIN squads_memebers sm ON s.id = sm.squad_id JOIN users u ON s.owner = u.id
    WHERE sm.user_id = $1 AND s.deleted_at IS NULL AND sm.deleted_at IS NULL
)
SELECT * 
FROM squad_members
ORDER BY created_at DESC;

-- name: GetSquadsByOwner :many
WITH squad_members AS (
    SELECT 
        COUNT(*) OVER () AS total,
        s.id,
        s.handle,
        s.owner,
        s.type,
        s.created_at,
        s.updated_at,
        s.deleted_at,
        u.first_name,
        u.last_name,
        u.phone_number,
        u.id as owner_id
    FROM squads s JOIN users u ON s.owner = u.id
    WHERE s.owner = $1 AND s.deleted_at IS NULL
)
SELECT * 
FROM squad_members
ORDER BY created_at DESC;

    
-- name: GetSquadsByType :many
SELECT id, handle, owner, type, created_at, updated_at, deleted_at from squads WHERE type = $1 AND deleted_at IS NULL;

-- name: DeleteSquad :exec
UPDATE squads SET deleted_at = $1, updated_at = $2
WHERE id = $3;

-- name: CreateSquadMember :one
INSERT INTO squads_memebers (squad_id, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSquadMembersBySquadId :many
WITH squad_memebers as (
    SELECT count(*) as total, sm.id, squad_id, user_id, sm.created_at, updated_at, deleted_at, us.first_name, us.last_name, us.phone_number from squads_memebers sm join users us on sm.user_id = us.id WHERE squad_id = $1 AND deleted_at IS NULL
    GROUP BY sm.id, squad_id, user_id, sm.created_at, updated_at, deleted_at, us.first_name, us.last_name, us.phone_number
)
SELECT * 
FROM squad_memebers
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetSquadMembersByUserId :many
SELECT sm.id, squad_id, user_id, sm.created_at, updated_at, deleted_at, us.first_name, us.last_name, us.phone_number from squads_memebers sm JOIN users us on sm.user_id = us.id  WHERE user_id = $1 AND deleted_at IS NULL;

-- name: DeleteSquadMember :exec
UPDATE squads_memebers SET deleted_at = $1, updated_at = $2
WHERE id = $3;

-- name: DeleteSquadMembersBySquadId :exec
UPDATE squads_memebers SET deleted_at = $1, updated_at = $2
WHERE squad_id = $3;

-- name: CreateSquadEarn :one
INSERT INTO squads_earns (squad_id, user_id, currency, earned, game_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetSquadEarnsBySquadId :many
WITH squadEarns AS (
    SELECT 
        COUNT(*) OVER () AS total, 
        id, 
        squad_id, 
        user_id, 
        currency, 
        earned, 
        game_id, 
        created_at, 
        updated_at, 
        deleted_at 
    FROM squads_earns 
    WHERE squad_id = $1 
    AND deleted_at IS NULL
)
SELECT * 
FROM squadEarns
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetSquadEarnsByUserIdAndSquadID :many
WITH squadEarns AS (
    SELECT 
        COUNT(*) OVER () AS total, 
        id, 
        squad_id, 
        user_id, 
        currency, 
        earned, 
        game_id, 
        created_at, 
        updated_at, 
        deleted_at 
    FROM squads_earns 
    WHERE squad_id = $1 and user_id = $2
    AND deleted_at IS NULL
)
SELECT * 
FROM squadEarns
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- name: GetSquadTotalEarnsBySquadID :one 
SELECT COALESCE(sum (earned), 0)::decimal from squads_earns WHERE squad_id = $1 AND deleted_at IS NULL;

-- name: GetUserEarnsForSquad :one
SELECT COALESCE(sum (earned), 0)::decimal from squads_earns WHERE squad_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetSquadMembersEarnings :many
WITH squad_members_earnings AS (
    SELECT 
        COUNT(*) OVER () AS total,
        sm.user_id,
        u.first_name,
        u.last_name,
        u.phone_number,
        COALESCE(SUM(se.earned), 0)::decimal AS total_earned,
        COUNT(se.id) AS total_games,
        MAX(se.created_at) AS last_earned_at
    FROM squads_memebers sm
    LEFT JOIN users u ON sm.user_id = u.id
    LEFT JOIN squads s ON sm.squad_id = s.id
    LEFT JOIN squads_earns se ON sm.user_id = se.user_id AND sm.squad_id = se.squad_id AND se.deleted_at IS NULL
    WHERE sm.squad_id = $1 
    AND s.owner = $2
    AND sm.deleted_at IS NULL
    GROUP BY sm.user_id, u.first_name, u.last_name, u.phone_number
)
SELECT * 
FROM squad_members_earnings
ORDER BY total_earned DESC
LIMIT $3 OFFSET $4;

-- name: CreateTournaments :one
INSERT INTO tournaments (rank, level, cumulative_points, rewards, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6) 
RETURNING *;

-- name: GetTournaments :many
WITH tournaments AS (
    SELECT 
        COUNT(*) OVER () AS total, 
        id, 
        rank, 
        level, 
        cumulative_points, 
        rewards, 
        created_at, 
        updated_at 
    FROM tournaments 
    WHERE deleted_at IS NULL
)
SELECT * 
FROM tournaments
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;

-- name: GetTornamentStyleRanking :many
WITH squadEarnings AS (
    SELECT 
        COUNT(*) OVER () AS total, 
        se.squad_id, 
        s.handle, 
        SUM(se.earned) AS total_earned,
        DATE_TRUNC('day', se.created_at) AS day
    FROM squads_earns se
    JOIN squads s ON se.squad_id = s.id
    WHERE se.created_at >= NOW() - INTERVAL '6 days'
    AND se.deleted_at IS NULL
    GROUP BY se.squad_id, s.handle, DATE_TRUNC('day', se.created_at)
)
SELECT 
    squad_id, 
    handle, 
    total,
    SUM(total_earned)::decimal AS total_earned,
    ARRAY_AGG(DISTINCT day) AS days,
    RANK() OVER (ORDER BY SUM(total_earned) DESC) AS rank
FROM squadEarnings
GROUP BY squad_id, handle, total
ORDER BY total_earned DESC
LIMIT $1 OFFSET $2;

-- name: GetTornamentStyles :many
SELECT id, rank, level, cumulative_points, rewards, created_at, updated_at 
FROM tournaments 
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetSquadEarningsAmountBySquadID :one
SELECT COALESCE(SUM(earned), 0)::decimal 
FROM squads_earns 
WHERE squad_id = $1 
AND deleted_at IS NULL; 

-- name: CreateTournamentClaim :one
INSERT INTO tournaments_claims (tournament_id, squad_id, claimed_at) 
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTournamentClaimBySquadID :one
SELECT id, tournament_id, squad_id, claimed_at 
FROM tournaments_claims 
WHERE squad_id = $1 
AND tournament_id = $2 
AND claimed_at IS NOT NULL
AND deleted_at IS NULL;

-- name: GetTournamentClaimByTournamentID :many
SELECT id, tournament_id, squad_id, claimed_at 
FROM tournaments_claims 
WHERE tournament_id = $1 
AND claimed_at IS NOT NULL
AND deleted_at IS NULL;

-- name: GetSquadMemberById :one
SELECT 
    sm.*, 
    s.handle AS squad_handle, 
    s.owner AS squad_owner
FROM 
    squads_memebers sm
JOIN 
    squads s 
ON 
    sm.squad_id = s.id
WHERE 
    sm.id = $1 
    AND sm.deleted_at IS NULL;


-- name: DeleteSquadMemberByID :exec
UPDATE squads_memebers SET deleted_at = now(), updated_at = now()
WHERE id = $1 and deleted_at IS NULL;

-- name: LeaveSquad :exec
UPDATE squads_memebers SET deleted_at = now(), updated_at = now()
WHERE user_id = $1 AND squad_id = $2 AND deleted_at IS NULL;

-- name: GetSquadByUserIDFromSquadMember :many
SELECT 
    s.id, 
    s.handle, 
    s.type, 
    s.created_at, 
    s.updated_at, 
    s.deleted_at,
    s.owner,
    sm.id AS member_id
FROM 
    squads s
JOIN 
    squads_memebers sm 
ON 
    s.id = sm.squad_id
WHERE 
    sm.user_id = $1 
    AND s.deleted_at IS NULL
    AND sm.deleted_at IS NULL;

-- name: AddWaitingSquadMember :one
INSERT INTO waiting_squad_members (user_id, squad_id, created_at) 
VALUES ($1, $2, $3)
RETURNING *;


-- name: DeleteWaitingSquadMember :exec
DELETE FROM waiting_squad_members 
WHERE id = $1 ;

-- name: DeleteWaitingSquadMembersBySquadID :exec
DELETE FROM waiting_squad_members 
WHERE squad_id = $1;

-- name: GetWaitingSquadMemberOwner :one
SELECT 
    wsm.id, 
    wsm.user_id, 
    wsm.squad_id,
    wsm.created_at,
    u.id AS owner_id,
    u.first_name,
    u.last_name,
    u.phone_number
FROM 
    waiting_squad_members wsm
JOIN 
    squads s 
ON 
    wsm.squad_id = s.squad_id
JOIN 
    users u 
ON 
    s.owner = u.id
WHERE 
    wsm.id = $1 and wsm.deleted_at IS NULL;


-- name: GetWaitingSquadMembersBySquadID :many
SELECT 
    COUNT(*) OVER () AS total,
    wsm.id,
    wsm.user_id,
    wsm.squad_id,
    wsm.created_at,
    u.first_name,
    u.last_name,
    u.phone_number
FROM 
    waiting_squad_members wsm
JOIN 
    users u 
ON 
    wsm.user_id = u.id
WHERE 
    wsm.squad_id = $1;

-- name: ApproveWaitingSquadMember :exec
WITH new_member AS (
    INSERT INTO squads_memebers (squad_id, user_id, created_at, updated_at)
    SELECT wsm.squad_id, wsm.user_id, NOW(), NOW()
    FROM waiting_squad_members wsm
    WHERE wsm.id = $1
    RETURNING squads_memebers.id, squad_id, user_id, created_at, updated_at
)
DELETE FROM waiting_squad_members 
WHERE waiting_squad_members.id = $1;
