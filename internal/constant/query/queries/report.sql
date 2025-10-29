-- name: GetPlayerCounts :one
SELECT
    COUNT(*) FILTER (WHERE default_currency IS NOT NULL) AS total_players,
    COUNT(*) FILTER (
        WHERE default_currency IS NOT NULL
          AND DATE(created_at) = $1
    ) AS new_players
FROM users;

-- name: GetBucksSpent :one
SELECT
    COALESCE(SUM(amount), 0)::int8 AS bucks_spent
FROM airtime_transactions
WHERE DATE(timestamp) = $1;