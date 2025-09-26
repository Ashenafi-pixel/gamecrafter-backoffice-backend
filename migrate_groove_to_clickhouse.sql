-- Migrate existing GrooveTech transactions from PostgreSQL to ClickHouse
-- This script should be run to backfill existing data

-- First, let's see what data we have
SELECT 
    transaction_id,
    account_id,
    amount,
    type,
    created_at
FROM groove_transactions 
WHERE account_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'
ORDER BY created_at DESC;

-- Insert wager transactions
INSERT INTO tucanbit_analytics.transactions (
    id, user_id, transaction_type, amount, currency, status, 
    bet_amount, net_result, balance_before, balance_after,
    created_at, updated_at
)
SELECT 
    transaction_id as id,
    account_id as user_id,
    'groove_bet' as transaction_type,
    amount,
    'USD' as currency,
    'completed' as status,
    amount as bet_amount,
    -amount as net_result,
    0 as balance_before,
    0 as balance_after,
    created_at,
    created_at as updated_at
FROM groove_transactions 
WHERE account_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'
AND type = 'wager'
ON CONFLICT (id) DO NOTHING;

-- Insert result transactions
INSERT INTO tucanbit_analytics.transactions (
    id, user_id, transaction_type, amount, currency, status, 
    win_amount, net_result, balance_before, balance_after,
    created_at, updated_at
)
SELECT 
    transaction_id as id,
    account_id as user_id,
    'groove_win' as transaction_type,
    amount,
    'USD' as currency,
    'completed' as status,
    amount as win_amount,
    amount as net_result,
    0 as balance_before,
    0 as balance_after,
    created_at,
    created_at as updated_at
FROM groove_transactions 
WHERE account_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'
AND type = 'result'
ON CONFLICT (id) DO NOTHING;

-- Insert rollback transactions
INSERT INTO tucanbit_analytics.transactions (
    id, user_id, transaction_type, amount, currency, status, 
    net_result, balance_before, balance_after,
    created_at, updated_at
)
SELECT 
    transaction_id as id,
    account_id as user_id,
    'refund' as transaction_type,
    amount,
    'USD' as currency,
    'completed' as status,
    amount as net_result,
    0 as balance_before,
    0 as balance_after,
    created_at,
    created_at as updated_at
FROM groove_transactions 
WHERE account_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'
AND type = 'rollback'
ON CONFLICT (id) DO NOTHING;