-- name: CreateLottery :one
INSERT INTO lotteries (id, name, price, min_selectable, max_selectable, draw_frequency, number_of_balls, description, status, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, 'active', NOW(), NOW())
RETURNING id, name, price, min_selectable, max_selectable, draw_frequency, number_of_balls, description, status, created_at, updated_at;

-- name: GetLotteryByID :one
SELECT id, name, price, min_selectable, max_selectable, draw_frequency, number_of_balls, description, status, created_at, updated_at
FROM lotteries
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateLotteryWinnersLog :one
INSERT INTO lottery_winners_logs (id, lottery_id, user_id, reward_id, won_amount, currency, ticket_number, status, created_at, updated_at,number_of_tickets)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, 'closed', NOW(), NOW(), $7)
RETURNING id, lottery_id, user_id, reward_id, won_amount, currency, ticket_number, status, created_at, updated_at;

-- name: CreateLotteryLog :one
INSERT INTO lottery_logs (id, lottery_id, lottery_reward_id, draw_numbers, prize, created_at, updated_at, uniq_identifier)
VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW(), $5)
RETURNING id, lottery_id, lottery_reward_id, draw_numbers, prize, created_at, updated_at;

-- name: GetAvailableLotteryService :one
SELECT id, client_id, client_secret, status, name, description, callback_url, created_at, updated_at, deleted_at
FROM lottery_services
WHERE status = 'active' AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLotteryLogsByUniqIdentifier :many
SELECT id, lottery_id, lottery_reward_id, draw_numbers, prize, created_at, updated_at, uniq_identifier
FROM lottery_logs
WHERE uniq_identifier = $1;