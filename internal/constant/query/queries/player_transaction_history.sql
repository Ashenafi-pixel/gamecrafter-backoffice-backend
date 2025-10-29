-- name: GetPlayerTransactionHistory :many
-- Unified query for all transaction types (GrooveTech, Sports, General)
WITH all_transactions AS (
    -- GrooveTech transactions
    SELECT 
        gt.id,
        gt.transaction_id,
        gt.account_id,
        gt.session_id,
        gt.type,
        gt.amount,
        gt.currency,
        gt.status,
        gt.created_at,
        gt.metadata::text,
        'gaming' as category,
        -- Extract game information from metadata
        (gt.metadata->>'game_id')::text as game_id,
        (gt.metadata->>'game_name')::text as game_name,
        (gt.metadata->>'round_id')::text as round_id,
        (gt.metadata->>'provider')::text as provider,
        (gt.metadata->>'device')::text as device,
        -- Null fields for other transaction types
        NULL::text as bet_reference_num,
        NULL::text as game_reference,
        NULL::text as bet_mode,
        NULL::text as description,
        NULL::numeric as potential_win,
        NULL::numeric as actual_win,
        NULL::numeric as odds,
        NULL::timestamp as placed_at,
        NULL::timestamp as settled_at,
        NULL::text as client_transaction_id,
        NULL::numeric as cash_out_multiplier,
        NULL::numeric as payout,
        NULL::numeric as house_edge
    FROM groove_transactions gt
    JOIN groove_accounts ga ON gt.account_id = ga.account_id
    WHERE ga.user_id = $1
        AND ($2::text IS NULL OR gt.account_id = $2)
        AND ($3::text IS NULL OR gt.type = $3)
        AND ($4::text IS NULL OR gt.status = $4)
        AND ($5::date IS NULL OR gt.created_at::date >= $5)
        AND ($6::date IS NULL OR gt.created_at::date <= $6)
        AND ($7::text IS NULL OR 'gaming' = $7)

    UNION ALL

    -- Sports betting transactions
    SELECT 
        sb.id,
        sb.transaction_id,
        NULL::text as account_id,
        NULL::text as session_id,
        'sport_bet' as type,
        sb.bet_amount as amount,
        sb.currency,
        sb.status,
        sb.created_at,
        sb.bet_details::text as metadata,
        'sports' as category,
        -- Null fields for gaming
        NULL::text as game_id,
        NULL::text as game_name,
        NULL::text as round_id,
        NULL::text as provider,
        NULL::text as device,
        -- Sports betting fields
        sb.bet_reference_num,
        sb.game_reference,
        sb.bet_mode,
        sb.description,
        sb.potential_win,
        sb.actual_win,
        sb.odds,
        sb.placed_at,
        sb.settled_at,
        -- Null fields for general betting
        NULL::text as client_transaction_id,
        NULL::numeric as cash_out_multiplier,
        NULL::numeric as payout,
        NULL::numeric as house_edge
    FROM sport_bets sb
    WHERE sb.user_id = $1
        AND ($2::text IS NULL OR sb.transaction_id = $2)
        AND ($3::text IS NULL OR 'sport_bet' = $3)
        AND ($4::text IS NULL OR sb.status = $4)
        AND ($5::date IS NULL OR sb.created_at::date >= $5)
        AND ($6::date IS NULL OR sb.created_at::date <= $6)
        AND ($7::text IS NULL OR 'sports' = $7)

    UNION ALL

    -- General betting transactions
    SELECT 
        b.id,
        b.client_transaction_id as transaction_id,
        NULL::text as account_id,
        NULL::text as session_id,
        'bet' as type,
        b.amount,
        b.currency,
        b.status,
        COALESCE(b.timestamp, NOW()) as created_at,
        NULL::text as metadata,
        'general' as category,
        -- Null fields for gaming
        NULL::text as game_id,
        NULL::text as game_name,
        NULL::text as round_id,
        NULL::text as provider,
        NULL::text as device,
        -- Null fields for sports betting
        NULL::text as bet_reference_num,
        NULL::text as game_reference,
        NULL::text as bet_mode,
        NULL::text as description,
        NULL::numeric as potential_win,
        NULL::numeric as actual_win,
        NULL::numeric as odds,
        NULL::timestamp as placed_at,
        NULL::timestamp as settled_at,
        -- General betting fields
        b.client_transaction_id,
        b.cash_out_multiplier,
        b.payout,
        b.house_edge
    FROM bets b
    WHERE b.user_id = $1
        AND ($2::text IS NULL OR b.client_transaction_id = $2)
        AND ($3::text IS NULL OR 'bet' = $3)
        AND ($4::text IS NULL OR b.status = $4)
        AND ($5::date IS NULL OR COALESCE(b.timestamp, NOW())::date >= $5)
        AND ($6::date IS NULL OR COALESCE(b.timestamp, NOW())::date <= $6)
        AND ($7::text IS NULL OR 'general' = $7)
)
SELECT * FROM all_transactions
ORDER BY created_at DESC
LIMIT $8 OFFSET $9;

-- name: GetPlayerTransactionHistoryCount :one
-- Unified count query for all transaction types
WITH all_transactions AS (
    -- GrooveTech transactions
    SELECT gt.id
    FROM groove_transactions gt
    JOIN groove_accounts ga ON gt.account_id = ga.account_id
    WHERE ga.user_id = $1
        AND ($2::text IS NULL OR gt.account_id = $2)
        AND ($3::text IS NULL OR gt.type = $3)
        AND ($4::text IS NULL OR gt.status = $4)
        AND ($5::date IS NULL OR gt.created_at::date >= $5)
        AND ($6::date IS NULL OR gt.created_at::date <= $6)
        AND ($7::text IS NULL OR 'gaming' = $7)

    UNION ALL

    -- Sports betting transactions
    SELECT sb.id
    FROM sport_bets sb
    WHERE sb.user_id = $1
        AND ($2::text IS NULL OR sb.transaction_id = $2)
        AND ($3::text IS NULL OR 'sport_bet' = $3)
        AND ($4::text IS NULL OR sb.status = $4)
        AND ($5::date IS NULL OR sb.created_at::date >= $5)
        AND ($6::date IS NULL OR sb.created_at::date <= $6)
        AND ($7::text IS NULL OR 'sports' = $7)

    UNION ALL

    -- General betting transactions
    SELECT b.id
    FROM bets b
    WHERE b.user_id = $1
        AND ($2::text IS NULL OR b.client_transaction_id = $2)
        AND ($3::text IS NULL OR 'bet' = $3)
        AND ($4::text IS NULL OR b.status = $4)
        AND ($5::date IS NULL OR COALESCE(b.timestamp, NOW())::date >= $5)
        AND ($6::date IS NULL OR COALESCE(b.timestamp, NOW())::date <= $6)
        AND ($7::text IS NULL OR 'general' = $7)
)
SELECT COUNT(*) as total FROM all_transactions;

-- name: GetPlayerTransactionSummary :one
-- Unified summary query for all transaction types
WITH all_transactions AS (
    -- GrooveTech transactions
    SELECT 
        gt.amount,
        gt.type,
        gt.created_at,
        CASE WHEN gt.type = 'result' AND gt.amount > 0 THEN 1 ELSE 0 END as is_win,
        CASE WHEN gt.type = 'wager' AND gt.amount < 0 THEN 1 ELSE 0 END as is_loss,
        CASE WHEN gt.type = 'wager' AND gt.amount < 0 THEN ABS(gt.amount) ELSE 0 END as wager_amount,
        CASE WHEN gt.type = 'result' AND gt.amount > 0 THEN gt.amount ELSE 0 END as win_amount
    FROM groove_transactions gt
    JOIN groove_accounts ga ON gt.account_id = ga.account_id
    WHERE ga.user_id = $1
        AND ($2::text IS NULL OR gt.account_id = $2)
        AND ($3::text IS NULL OR gt.type = $3)
        AND ($4::text IS NULL OR gt.status = $4)
        AND ($5::date IS NULL OR gt.created_at::date >= $5)
        AND ($6::date IS NULL OR gt.created_at::date <= $6)
        AND ($7::text IS NULL OR 'gaming' = $7)

    UNION ALL

    -- Sports betting transactions
    SELECT 
        sb.bet_amount as amount,
        'sport_bet' as type,
        sb.created_at,
        CASE WHEN sb.actual_win > 0 THEN 1 ELSE 0 END as is_win,
        CASE WHEN sb.actual_win = 0 OR sb.actual_win IS NULL THEN 1 ELSE 0 END as is_loss,
        sb.bet_amount as wager_amount,
        COALESCE(sb.actual_win, 0) as win_amount
    FROM sport_bets sb
    WHERE sb.user_id = $1
        AND ($2::text IS NULL OR sb.transaction_id = $2)
        AND ($3::text IS NULL OR 'sport_bet' = $3)
        AND ($4::text IS NULL OR sb.status = $4)
        AND ($5::date IS NULL OR sb.created_at::date >= $5)
        AND ($6::date IS NULL OR sb.created_at::date <= $6)
        AND ($7::text IS NULL OR 'sports' = $7)

    UNION ALL

    -- General betting transactions
    SELECT 
        b.amount,
        'bet' as type,
        COALESCE(b.timestamp, NOW()) as created_at,
        CASE WHEN b.payout > 0 THEN 1 ELSE 0 END as is_win,
        CASE WHEN b.payout = 0 OR b.payout IS NULL THEN 1 ELSE 0 END as is_loss,
        b.amount as wager_amount,
        COALESCE(b.payout, 0) as win_amount
    FROM bets b
    WHERE b.user_id = $1
        AND ($2::text IS NULL OR b.client_transaction_id = $2)
        AND ($3::text IS NULL OR 'bet' = $3)
        AND ($4::text IS NULL OR b.status = $4)
        AND ($5::date IS NULL OR COALESCE(b.timestamp, NOW())::date >= $5)
        AND ($6::date IS NULL OR COALESCE(b.timestamp, NOW())::date <= $6)
        AND ($7::text IS NULL OR 'general' = $7)
)
SELECT 
    $1 as user_id,
    COUNT(*) as transaction_count,
    COALESCE(SUM(wager_amount), 0) as total_wagers,
    COALESCE(SUM(win_amount), 0) as total_wins,
    COALESCE(SUM(wager_amount), 0) as total_losses,
    COALESCE(SUM(amount), 0) as net_result,
    SUM(is_win) as win_count,
    SUM(is_loss) as loss_count,
    COALESCE(AVG(wager_amount), 0) as average_bet,
    COALESCE(MAX(win_amount), 0) as max_win,
    COALESCE(MAX(wager_amount), 0) as max_loss,
    MIN(created_at) as first_transaction,
    MAX(created_at) as last_transaction
FROM all_transactions;

-- name: GetPlayerTransactionHistoryByAccountID :many
SELECT 
    gt.id,
    gt.transaction_id,
    gt.account_id,
    gt.session_id,
    gt.type,
    gt.amount,
    gt.currency,
    gt.status,
    gt.created_at,
    gt.metadata,
    -- Extract game information from metadata
    (gt.metadata->>'game_id')::text as game_id,
    (gt.metadata->>'game_name')::text as game_name,
    (gt.metadata->>'round_id')::text as round_id,
    (gt.metadata->>'provider')::text as provider,
    (gt.metadata->>'device')::text as device
FROM groove_transactions gt
WHERE gt.account_id = $1
    AND ($2::text IS NULL OR gt.type = $2)
    AND ($3::text IS NULL OR gt.status = $3)
    AND ($4::date IS NULL OR gt.created_at::date >= $4)
    AND ($5::date IS NULL OR gt.created_at::date <= $5)
ORDER BY gt.created_at DESC
LIMIT $6 OFFSET $7;

-- name: GetPlayerTransactionHistoryByAccountIDCount :one
SELECT COUNT(*) as total
FROM groove_transactions gt
WHERE gt.account_id = $1
    AND ($2::text IS NULL OR gt.type = $2)
    AND ($3::text IS NULL OR gt.status = $3)
    AND ($4::date IS NULL OR gt.created_at::date >= $4)
    AND ($5::date IS NULL OR gt.created_at::date <= $5);