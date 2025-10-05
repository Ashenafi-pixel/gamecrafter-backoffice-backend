-- Migration script to move data from PostgreSQL to ClickHouse
-- This script will be executed to populate the ClickHouse analytics tables

-- 1. Migrate groove_transactions data
INSERT INTO tucanbit_analytics.groove_transactions 
SELECT 
    id::String,
    transaction_id,
    account_id,
    session_id,
    amount,
    currency,
    type,
    status,
    created_at,
    metadata::String
FROM groove_transactions;

-- 2. Migrate transactions data (deposits/withdrawals)
INSERT INTO tucanbit_analytics.transactions
SELECT 
    id::String,
    user_id::String,
    amount,
    currency_code as currency,
    status,
    transaction_type,
    tx_hash,
    from_address,
    to_address,
    fee,
    exchange_rate,
    usd_amount_cents,
    processor::String,
    confirmations,
    block_number,
    block_hash,
    metadata::String,
    created_at,
    updated_at
FROM transactions;

-- 3. Migrate cashback_earnings data
INSERT INTO tucanbit_analytics.cashback_earnings
SELECT 
    id::String,
    user_id::String,
    earned_amount,
    currency,
    status,
    created_at
FROM cashback_earnings;

-- 4. Migrate cashback_claims data (if any)
INSERT INTO tucanbit_analytics.cashback_claims
SELECT 
    id::String,
    user_id::String,
    claim_amount,
    currency_code as currency,
    status,
    created_at
FROM cashback_claims;

-- 5. Migrate bets data (if any)
INSERT INTO tucanbit_analytics.bets
SELECT 
    id::String,
    user_id::String,
    amount,
    currency,
    status,
    timestamp as created_at
FROM bets;