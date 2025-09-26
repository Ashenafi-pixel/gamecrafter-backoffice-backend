-- ClickHouse Schema for TucanBIT Casino Analytics
-- This schema is optimized for analytical queries and reporting

-- Create database
CREATE DATABASE IF NOT EXISTS tucanbit_analytics;

USE tucanbit_analytics;

-- Main transactions table (all casino transactions)
CREATE TABLE IF NOT EXISTS transactions (
    id String,
    user_id String,
    transaction_type Enum8('deposit' = 1, 'withdrawal' = 2, 'bet' = 3, 'win' = 4, 'bonus' = 5, 'cashback' = 6, 'refund' = 7, 'groove_deposit' = 8, 'groove_withdrawal' = 9, 'groove_bet' = 10, 'groove_win' = 11),
    amount Decimal(20, 8),
    currency String DEFAULT 'USD',
    status Enum8('pending' = 1, 'completed' = 2, 'failed' = 3, 'cancelled' = 4),
    game_id Nullable(String),
    game_name Nullable(String),
    provider Nullable(String),
    session_id Nullable(String),
    round_id Nullable(String),
    bet_amount Nullable(Decimal(20, 8)),
    win_amount Nullable(Decimal(20, 8)),
    net_result Nullable(Decimal(20, 8)),
    balance_before Decimal(20, 8),
    balance_after Decimal(20, 8),
    payment_method Nullable(String),
    external_transaction_id Nullable(String),
    metadata Nullable(String), -- JSON metadata
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now(),
    date Date MATERIALIZED toDate(created_at),
    hour UInt8 MATERIALIZED toHour(created_at),
    day_of_week UInt8 MATERIALIZED toDayOfWeek(created_at),
    month UInt8 MATERIALIZED toMonth(created_at),
    year UInt16 MATERIALIZED toYear(created_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at, transaction_type)
SETTINGS index_granularity = 8192;

-- User analytics table
CREATE TABLE IF NOT EXISTS user_analytics (
    user_id String,
    date Date,
    total_deposits Decimal(20, 8) DEFAULT 0,
    total_withdrawals Decimal(20, 8) DEFAULT 0,
    total_bets Decimal(20, 8) DEFAULT 0,
    total_wins Decimal(20, 8) DEFAULT 0,
    total_bonuses Decimal(20, 8) DEFAULT 0,
    total_cashback Decimal(20, 8) DEFAULT 0,
    net_loss Decimal(20, 8) DEFAULT 0,
    transaction_count UInt32 DEFAULT 0,
    unique_games_played UInt16 DEFAULT 0,
    session_count UInt16 DEFAULT 0,
    avg_bet_amount Decimal(20, 8) DEFAULT 0,
    max_bet_amount Decimal(20, 8) DEFAULT 0,
    min_bet_amount Decimal(20, 8) DEFAULT 0,
    last_activity DateTime,
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (user_id, date)
SETTINGS index_granularity = 8192;

-- Game analytics table
CREATE TABLE IF NOT EXISTS game_analytics (
    game_id String,
    game_name String,
    provider String,
    date Date,
    total_bets Decimal(20, 8) DEFAULT 0,
    total_wins Decimal(20, 8) DEFAULT 0,
    total_players UInt32 DEFAULT 0,
    total_sessions UInt32 DEFAULT 0,
    total_rounds UInt32 DEFAULT 0,
    avg_bet_amount Decimal(20, 8) DEFAULT 0,
    max_bet_amount Decimal(20, 8) DEFAULT 0,
    min_bet_amount Decimal(20, 8) DEFAULT 0,
    rtp Decimal(5, 4) DEFAULT 0, -- Return to Player percentage
    volatility Enum8('low' = 1, 'medium' = 2, 'high' = 3) DEFAULT 'medium',
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (game_id, date)
SETTINGS index_granularity = 8192;

-- Session analytics table
CREATE TABLE IF NOT EXISTS session_analytics (
    session_id String,
    user_id String,
    game_id Nullable(String),
    game_name Nullable(String),
    provider Nullable(String),
    start_time DateTime,
    end_time Nullable(DateTime),
    duration_seconds Nullable(UInt32),
    total_bets Decimal(20, 8) DEFAULT 0,
    total_wins Decimal(20, 8) DEFAULT 0,
    net_result Decimal(20, 8) DEFAULT 0,
    bet_count UInt32 DEFAULT 0,
    win_count UInt32 DEFAULT 0,
    max_balance Decimal(20, 8) DEFAULT 0,
    min_balance Decimal(20, 8) DEFAULT 0,
    session_type Enum8('regular' = 1, 'bonus' = 2, 'free_play' = 3) DEFAULT 'regular',
    device_type Nullable(String),
    ip_address Nullable(String),
    user_agent Nullable(String),
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now(),
    date Date MATERIALIZED toDate(start_time)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (user_id, start_time, session_id)
SETTINGS index_granularity = 8192;

-- Real-time balance tracking
CREATE TABLE IF NOT EXISTS balance_snapshots (
    user_id String,
    balance Decimal(20, 8),
    currency String DEFAULT 'USD',
    snapshot_time DateTime DEFAULT now(),
    transaction_id Nullable(String),
    transaction_type Nullable(String),
    created_at DateTime DEFAULT now(),
    date Date MATERIALIZED toDate(snapshot_time),
    hour UInt8 MATERIALIZED toHour(snapshot_time)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(snapshot_time)
ORDER BY (user_id, snapshot_time)
SETTINGS index_granularity = 8192;

-- GrooveTech specific analytics
CREATE TABLE IF NOT EXISTS groove_analytics (
    account_id String,
    game_id String,
    session_id String,
    transaction_id String,
    transaction_type Enum8('wager' = 1, 'result' = 2, 'rollback' = 3),
    amount Decimal(20, 8),
    balance_before Decimal(20, 8),
    balance_after Decimal(20, 8),
    game_status Enum8('active' = 1, 'completed' = 2, 'cancelled' = 3),
    round_id Nullable(String),
    external_transaction_id Nullable(String),
    created_at DateTime DEFAULT now(),
    date Date MATERIALIZED toDate(created_at)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (account_id, created_at, transaction_type)
SETTINGS index_granularity = 8192;

-- Create materialized views for real-time aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_user_stats_mv
TO daily_user_stats
AS SELECT
    user_id,
    date,
    sumIf(amount, transaction_type = 'deposit') as total_deposits,
    sumIf(amount, transaction_type = 'withdrawal') as total_withdrawals,
    sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
    sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
    sumIf(amount, transaction_type = 'bonus') as total_bonuses,
    sumIf(amount, transaction_type = 'cashback') as total_cashback,
    count() as transaction_count,
    uniqExact(game_id) as unique_games_played,
    uniqExact(session_id) as session_count,
    avgIf(amount, transaction_type IN ('bet', 'groove_bet')) as avg_bet_amount,
    maxIf(amount, transaction_type IN ('bet', 'groove_bet')) as max_bet_amount,
    minIf(amount, transaction_type IN ('bet', 'groove_bet')) as min_bet_amount,
    max(created_at) as last_activity
FROM transactions
GROUP BY user_id, date;

-- Create the target table for the materialized view
CREATE TABLE IF NOT EXISTS daily_user_stats (
    user_id String,
    date Date,
    total_deposits Decimal(20, 8),
    total_withdrawals Decimal(20, 8),
    total_bets Decimal(20, 8),
    total_wins Decimal(20, 8),
    total_bonuses Decimal(20, 8),
    total_cashback Decimal(20, 8),
    transaction_count UInt32,
    unique_games_played UInt16,
    session_count UInt16,
    avg_bet_amount Decimal(20, 8),
    max_bet_amount Decimal(20, 8),
    min_bet_amount Decimal(20, 8),
    last_activity DateTime
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (user_id, date)
SETTINGS index_granularity = 8192;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions (user_id, date) TYPE minmax GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_transactions_type_date ON transactions (transaction_type, date) TYPE minmax GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_transactions_game_date ON transactions (game_id, date) TYPE minmax GRANULARITY 1;
CREATE INDEX IF NOT EXISTS idx_sessions_user_start ON session_analytics (user_id, start_time) TYPE minmax GRANULARITY 1;