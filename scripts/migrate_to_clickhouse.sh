#!/bin/bash

# Data Migration Script: PostgreSQL to ClickHouse
# This script migrates all historical casino data to ClickHouse

set -e

echo " Starting PostgreSQL to ClickHouse Data Migration..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Database connection details
PG_HOST="localhost"
PG_PORT="5433"
PG_DB="tucanbit"
PG_USER="tucanbit"
PG_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"

CH_HOST="localhost"
CH_PORT="9000"
CH_DB="tucanbit_analytics"
CH_USER="tucanbit"
CH_PASSWORD="tucanbit_clickhouse_password"

# Check if ClickHouse is ready
print_status "Checking ClickHouse connection..."
if ! docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 1" > /dev/null 2>&1; then
    print_error "ClickHouse is not ready. Please start ClickHouse first."
    exit 1
fi
print_success "ClickHouse is ready!"

# Initialize ClickHouse schema if not exists
print_status "Initializing ClickHouse schema..."
# Drop materialized view temporarily to avoid issues during migration
docker exec -i tucanbit-clickhouse clickhouse-client --multiquery < clickhouse/schema.sql
docker exec tucanbit-clickhouse clickhouse-client --query "DROP VIEW IF EXISTS daily_user_stats_mv"
print_success "Schema initialized!"

# Function to migrate users/registrations
migrate_users() {
    print_status "Migrating user registrations..."
    
    # Export users from PostgreSQL
    PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
    COPY (
        SELECT 
            id::text as user_id,
            'registration' as transaction_type,
            0 as amount,
            'USD' as currency,
            'completed' as status,
            NULL as game_id,
            NULL as game_name,
            NULL as provider,
            NULL as session_id,
            NULL as round_id,
            NULL as bet_amount,
            NULL as win_amount,
            NULL as net_result,
            0 as balance_before,
            0 as balance_after,
            'registration' as payment_method,
            id::text as external_transaction_id,
            json_build_object('registration_method', 'direct', 'email', email, 'phone', phone_number)::text as metadata,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
        FROM users
        WHERE created_at IS NOT NULL
    ) TO STDOUT WITH CSV HEADER" > /tmp/users_export.csv

    # Import to ClickHouse
    docker exec -i tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < /tmp/users_export.csv
    
    print_success "User registrations migrated!"
}

# Function to migrate balances
migrate_balances() {
    print_status "Migrating balance transactions..."
    
    # Export balance logs from PostgreSQL
    PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
    COPY (
        SELECT 
            id::text as user_id,
            CASE 
                WHEN component = 'real_money' THEN 'deposit'
                WHEN component = 'bonus_money' THEN 'bonus'
                ELSE 'other'
            END as transaction_type,
            change_units as amount,
            'USD' as currency,
            'completed' as status,
            NULL as game_id,
            NULL as game_name,
            NULL as provider,
            NULL as session_id,
            NULL as round_id,
            NULL as bet_amount,
            NULL as win_amount,
            change_units as net_result,
            balance_after_units as balance_before,
            balance_after_units as balance_after,
            'balance_log' as payment_method,
            transaction_id as external_transaction_id,
            json_build_object('balance_log', true, 'description', description)::text as metadata,
            to_char(timestamp, 'YYYY-MM-DD HH24:MI:SS') as created_at,
            to_char(timestamp, 'YYYY-MM-DD HH24:MI:SS') as updated_at
        FROM balance_logs
        WHERE timestamp IS NOT NULL
    ) TO STDOUT WITH CSV HEADER" > /tmp/balances_export.csv

    # Import to ClickHouse
    docker exec -i tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < /tmp/balances_export.csv
    
    print_success "Balance transactions migrated!"
}

# Function to migrate GrooveTech transactions
migrate_groove_transactions() {
    print_status "Migrating GrooveTech transactions..."
    
    # Export GrooveTech transactions from PostgreSQL
    PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
    COPY (
        SELECT 
            account_id as user_id,
            CASE 
                WHEN type = 'wager' THEN 'groove_bet'
                WHEN type = 'result' THEN 'groove_win'
                WHEN type = 'rollback' THEN 'refund'
                ELSE 'groove_bet'
            END as transaction_type,
            bet_amount as amount,
            'USD' as currency,
            'completed' as status,
            game_id,
            NULL as game_name,
            'GrooveTech' as provider,
            game_session_id as session_id,
            round_id,
            bet_amount,
            CASE WHEN type = 'result' THEN bet_amount ELSE NULL END as win_amount,
            CASE 
                WHEN type = 'wager' THEN -bet_amount
                WHEN type = 'result' THEN bet_amount
                ELSE bet_amount
            END as net_result,
            0 as balance_before,
            0 as balance_after,
            'groove' as payment_method,
            account_transaction_id as external_transaction_id,
            json_build_object('groove_transaction', true, 'device', device, 'status', status)::text as metadata,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
        FROM groove_transactions
        WHERE created_at IS NOT NULL
    ) TO STDOUT WITH CSV HEADER" > /tmp/groove_export.csv

    # Import to ClickHouse
    docker exec -i tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < /tmp/groove_export.csv
    
    print_success "GrooveTech transactions migrated!"
}

# Function to create balance snapshots
create_balance_snapshots() {
    print_status "Creating balance snapshots..."
    
    # Export balance snapshots
    PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
    COPY (
        SELECT 
            user_id,
            change_units as balance,
            'USD' as currency,
            to_char(timestamp, 'YYYY-MM-DD HH24:MI:SS') as snapshot_time,
            id::text as transaction_id,
            CASE 
                WHEN component = 'real_money' THEN 'deposit'
                WHEN component = 'bonus_money' THEN 'bonus'
                ELSE 'other'
            END as transaction_type
        FROM balance_logs
        WHERE timestamp IS NOT NULL
        ORDER BY user_id, timestamp
    ) TO STDOUT WITH CSV HEADER" > /tmp/balance_snapshots.csv

    # Import to ClickHouse
    docker exec -i tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.balance_snapshots FORMAT CSV" < /tmp/balance_snapshots.csv
    
    print_success "Balance snapshots created!"
}

# Function to create session analytics
create_session_analytics() {
    print_status "Creating session analytics..."
    
    # Export session data
    PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
    COPY (
        SELECT 
            DISTINCT ON (user_id, DATE(created_at))
            user_id,
            DATE(created_at) as date,
            SUM(CASE WHEN transaction_type = 'deposit' THEN amount ELSE 0 END) as total_deposits,
            SUM(CASE WHEN transaction_type = 'withdrawal' THEN amount ELSE 0 END) as total_withdrawals,
            SUM(CASE WHEN transaction_type IN ('bet', 'groove_bet') THEN amount ELSE 0 END) as total_bets,
            SUM(CASE WHEN transaction_type IN ('win', 'groove_win') THEN amount ELSE 0 END) as total_wins,
            SUM(CASE WHEN transaction_type = 'bonus' THEN amount ELSE 0 END) as total_bonuses,
            SUM(CASE WHEN transaction_type = 'cashback' THEN amount ELSE 0 END) as total_cashback,
            COUNT(*) as transaction_count,
            COUNT(DISTINCT game_id) as unique_games_played,
            COUNT(DISTINCT session_id) as session_count,
            AVG(CASE WHEN transaction_type IN ('bet', 'groove_bet') THEN amount END) as avg_bet_amount,
            MAX(CASE WHEN transaction_type IN ('bet', 'groove_bet') THEN amount END) as max_bet_amount,
            MIN(CASE WHEN transaction_type IN ('bet', 'groove_bet') THEN amount END) as min_bet_amount,
            MAX(to_char(created_at, 'YYYY-MM-DD HH24:MI:SS')) as last_activity
        FROM (
            SELECT user_id::text as user_id, 
                   CASE 
                       WHEN component = 'real_money' THEN 'deposit'
                       WHEN component = 'bonus_money' THEN 'bonus'
                       ELSE 'other'
                   END as transaction_type, 
                   change_units as amount, 
                   NULL as game_id, 
                   NULL as session_id, 
                   timestamp as created_at
            FROM balance_logs
            UNION ALL
            SELECT account_id as user_id, 
                   CASE WHEN type = 'wager' THEN 'groove_bet' ELSE 'groove_win' END as transaction_type,
                   bet_amount as amount, game_id, game_session_id as session_id, created_at
            FROM groove_transactions
        ) all_transactions
        GROUP BY user_id, DATE(created_at)
    ) TO STDOUT WITH CSV HEADER" > /tmp/user_analytics.csv

    # Import to ClickHouse
    docker exec -i tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.user_analytics FORMAT CSV" < /tmp/user_analytics.csv
    
    print_success "Session analytics created!"
}

# Main migration process
main() {
    print_status "Starting comprehensive data migration..."
    
    # Create temporary directory
    mkdir -p /tmp/migration
    
    # Run migrations
    migrate_users
    migrate_balances
    migrate_groove_transactions
    create_balance_snapshots
    create_session_analytics
    
    # Clean up temporary files
    rm -f /tmp/*_export.csv /tmp/*_snapshots.csv /tmp/user_analytics.csv
    
    # Recreate materialized view after migration
    print_status "Recreating materialized views..."
    docker exec tucanbit-clickhouse clickhouse-client --query "
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
    GROUP BY user_id, date;"
    
    print_success "Materialized views recreated!"
    
    print_success "Data migration completed successfully!"
    
    # Show migration summary
    print_status "Migration Summary:"
    docker exec tucanbit-clickhouse clickhouse-client --query "
    SELECT 
        'transactions' as table_name,
        count(*) as record_count
    FROM tucanbit_analytics.transactions
    UNION ALL
    SELECT 
        'balance_snapshots' as table_name,
        count(*) as record_count
    FROM tucanbit_analytics.balance_snapshots
    UNION ALL
    SELECT 
        'user_analytics' as table_name,
        count(*) as record_count
    FROM tucanbit_analytics.user_analytics"
    
    print_success "All casino data has been migrated to ClickHouse!"
    print_status "Next steps:"
    echo "1. Set up real-time sync in your application"
    echo "2. Configure analytics APIs"
    echo "3. Build your analytics dashboard"
}

# Run migration
main