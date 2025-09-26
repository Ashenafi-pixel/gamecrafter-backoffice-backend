-- ClickHouse Schema for TucanBIT Casino Analytics
-- Simplified version without materialized views

-- Create database
CREATE DATABASE IF NOT EXISTS tucanbit_analytics;

USE tucanbit_analytics;

-- Main transactions table (all casino transactions)
CREATE TABLE IF NOT EXISTS transactions (
    id String,
    user_id String,
    transaction_type Enum8('deposit' = 1, 'withdrawal' = 2, 'bet' = 3, 'win' = 4, 'bonus' = 5, 'cashback' = 6, 'refund' = 7, 'groove_deposit' = 8, 'groove_withdrawal' = 9, 'groove_bet' = 10, 'groove_win' = 11, 'registration' = 12),
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

-- Balance snapshots table
CREATE TABLE IF NOT EXISTS balance_snapshots (
    user_id String,
    balance Decimal(20, 8),
    currency String DEFAULT 'USD',
    snapshot_time DateTime,
    transaction_id Nullable(String),
    transaction_type Nullable(String),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(snapshot_time)
ORDER BY (user_id, snapshot_time)
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
    transaction_count UInt32 DEFAULT 0,
    unique_games_played UInt32 DEFAULT 0,
    session_count UInt32 DEFAULT 0,
    avg_bet_amount Decimal(20, 8) DEFAULT 0,
    max_bet_amount Decimal(20, 8) DEFAULT 0,
    min_bet_amount Decimal(20, 8) DEFAULT 0,
    last_activity DateTime DEFAULT now(),
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (user_id, date)
SETTINGS index_granularity = 8192;

-- Game analytics table
CREATE TABLE IF NOT EXISTS game_analytics (
    game_id String,
    game_name Nullable(String),
    provider Nullable(String),
    date Date,
    total_bets Decimal(20, 8) DEFAULT 0,
    total_wins Decimal(20, 8) DEFAULT 0,
    total_players UInt32 DEFAULT 0,
    total_sessions UInt32 DEFAULT 0,
    total_rounds UInt32 DEFAULT 0,
    avg_bet_amount Decimal(20, 8) DEFAULT 0,
    max_bet_amount Decimal(20, 8) DEFAULT 0,
    min_bet_amount Decimal(20, 8) DEFAULT 0,
    rtp Decimal(5, 2) DEFAULT 0,
    volatility String DEFAULT 'medium',
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = MergeTree()
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
    end_time DateTime,
    duration_seconds Nullable(UInt32),
    total_bets Decimal(20, 8) DEFAULT 0,
    total_wins Decimal(20, 8) DEFAULT 0,
    net_result Decimal(20, 8) DEFAULT 0,
    bet_count UInt32 DEFAULT 0,
    win_count UInt32 DEFAULT 0,
    max_balance Decimal(20, 8) DEFAULT 0,
    min_balance Decimal(20, 8) DEFAULT 0,
    session_type String DEFAULT 'regular',
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (session_id, start_time)
SETTINGS index_granularity = 8192;