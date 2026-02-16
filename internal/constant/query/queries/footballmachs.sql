-- name: CreateClub :one 
INSERT INTO clubs (club_name,status,timestamp) VALUES ($1,$2,$3) RETURNING *;

-- name: GetClubs :many 
WITH club_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM clubs
    WHERE true
)
SELECT *
FROM club_data
ORDER BY club_name ASC
LIMIT $1 OFFSET $2; 
-- name: UpdateClubNameByID :one 
UPDATE clubs SET club_name = $1 where id = $2 RETURNING *;

-- name: GetClubByID :one
SELECT * FROM clubs where id = $1;

-- name: CreateLeague :one
INSERT INTO leagues(league_name,status,timestamp) VALUES ($1,$2,$3) RETURNING *;

-- name: GetLeagues :many 
WITH league_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM leagues
    WHERE true
)
SELECT *
FROM league_data
ORDER BY league_name ASC
LIMIT $1 OFFSET $2;

-- name: GetLeagueByID :one
SELECT * FROM leagues where id = $1;

-- name: UpdateLeagueNameByID :one 
UPDATE leagues SET league_name = $1 where id = $2 RETURNING *;

-- name: AddFootballMatchs :one
INSERT INTO football_matchs (round_id,league,date,home_team,away_team,timestamp)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: CreateFootballRound :one 
INSERT INTO football_match_rounds (status,timestamp) VALUES ($1,$2) RETURNING *;

-- name: GetFootballMatchRoundByID :one
SELECT * FROM football_match_rounds where id = $1;

-- name: GetFootballMatchByRoundStatus :many
select m.*,r.status from football_matchs m join football_match_rounds r on r.id = m.round_id where r.status=$1;


-- name: UpdateFootballmatchByRoundID :one 
UPDATE football_match_rounds set status = $1 where id = $2 RETURNING *;

-- name: UpdateFootballMatchsByID :one 
UPDATE football_matchs set status = $1 where id = $2 RETURNING *;

-- name: CloseFootballMatchRound :one
UPDATE football_matchs SET status = $1 , home_score = $2, away_score = $3, won = $4 where id = $5 RETURNING *;

-- name: CreateFootballMatchRound :one
INSERT INTO football_match_rounds (status,timestamp) VALUES ($1,$2) RETURNING *;

-- name: GetFootballMatchRound :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_match_rounds
    WHERE true
)
SELECT *
FROM match_data
ORDER BY timestamp DESC
LIMIT $1 OFFSET $2;

-- name: GetFootballMatchRoundByStatus :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_match_rounds
    WHERE  status = $1
)
SELECT *
FROM match_data 
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;


-- name: CreateFootballMatch :one
INSERT INTO football_matchs (round_id,league,date,home_team,away_team,timestamp)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: GetFootballMatchs :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_matchs
    WHERE true
)
SELECT * FROM match_data ORDER BY timestamp DESC LIMIT $1 OFFSET $2;

-- name: GetFootballMatchByStatus :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_matchs
    WHERE status = $1
)
SELECT * FROM match_data ORDER BY timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetFootballMatchByRoundID :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_matchs
    WHERE round_id = $1
)
SELECT * FROM match_data ORDER BY timestamp DESC LIMIT $2 OFFSET $3;

-- name: GetFootballMatchByID :one
SELECT * FROM football_matchs where id = $1;

-- name: GetFootballRoundMatchs :many
WITH match_data AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM football_matchs
    WHERE round_id = $1
)
SELECT * FROM match_data ORDER BY timestamp DESC LIMIT $2 OFFSET $3;

-- name: CreateFootballBet :one 
INSERT INTO users_football_matche_rounds (user_id,football_round_id,bet_amount,won_amount,timestamp) 
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: CreateFootballBetUserSelection :one 
INSERT INTO users_football_matches (match_id,selection,status,users_football_matche_round_id)
VALUES ($1,$2,$3,$4) RETURNING *;

-- name: GetFootballMatchesByRoundID :many 
SELECT * FROM football_matchs where round_id = $1;

-- name: UpdateUserFootballMatchStatusByMatchID :one 
UPDATE users_football_matches SET status = $1 where id = $2 RETURNING *;

-- name: GetUserFootballMatchSelectionsByMatchID :many 
SELECT * FROM users_football_matches where match_id = $1 ;

-- name: GetFootballMatchesByStatus :many 
SELECT * FROM football_matchs where status = $1 and round_id = $2;

-- name: GetAllFootBallMatchByRoundByRoundID :many 
SELECT * FROM users_football_matche_rounds where football_round_id = $1;

-- name: GetAllUserFootballBetByStatusAndRoundID :many 
SELECT * FROM users_football_matches where status = $1 and users_football_matche_round_id = $2;

-- name: UpdateUserFootballMatcheRoundsByID :exec 
UPDATE users_football_matche_rounds SET status = $1, won_status = $2 WHERE id = $3;

-- name: GetUserFootballBets :many 
WITH match_rounds AS (
    SELECT *,
           COUNT(*) OVER () as total_rows
    FROM users_football_matche_rounds ufmr where user_id = $1
)
SELECT 
    status AS round_status,
    id,
    bet_amount,
    won_amount AS winning_amount,
    currency,
    total_rows
FROM 
    match_rounds 
ORDER BY 
    timestamp DESC 
LIMIT $2 
OFFSET $3;

-- name: GetUserFootballBetMatchesForUserBet :many 
SELECT  * FROM users_football_matches where users_football_matche_round_id = $1;
