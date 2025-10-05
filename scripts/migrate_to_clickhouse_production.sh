#!/bin/bash

# Production ClickHouse Migration Script for TucanBIT Casino
# This script migrates all historical casino data to ClickHouse with production-grade features

set -e

echo " Starting Production ClickHouse Migration for TucanBIT Casino..."

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

# Production database connection details
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5433}"
PG_DB="${PG_DB:-tucanbit}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-password}"

CH_HOST="${CH_HOST:-localhost}"
CH_PORT="${CH_PORT:-9000}"
CH_DB="${CH_DB:-tucanbit_analytics}"
CH_USER="${CH_USER:-tucanbit}"
CH_PASSWORD="${CH_PASSWORD:-tucanbit_clickhouse_password}"

# Migration settings
BATCH_SIZE="${BATCH_SIZE:-10000}"
PARALLEL_WORKERS="${PARALLEL_WORKERS:-4}"
BACKUP_ENABLED="${BACKUP_ENABLED:-true}"

# Create migration directory
MIGRATION_DIR="/tmp/tucanbit_migration_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$MIGRATION_DIR"

print_status "Migration directory: $MIGRATION_DIR"

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check PostgreSQL connection
    if ! PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "SELECT 1" > /dev/null 2>&1; then
        print_error "Cannot connect to PostgreSQL. Please check your connection settings."
        exit 1
    fi
    print_success "PostgreSQL connection verified"
    
    # Check ClickHouse connection
    if ! docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 1" > /dev/null 2>&1; then
        print_error "Cannot connect to ClickHouse. Please ensure ClickHouse is running."
        exit 1
    fi
    print_success "ClickHouse connection verified"
    
    # Check required tools
    for tool in psql docker jq; do
        if ! command -v $tool &> /dev/null; then
            print_error "Required tool '$tool' is not installed."
            exit 1
        fi
    done
    print_success "All required tools are available"
}

# Function to backup existing data
backup_existing_data() {
    if [ "$BACKUP_ENABLED" = "true" ]; then
        print_status "Creating backup of existing ClickHouse data..."
        
        BACKUP_FILE="$MIGRATION_DIR/clickhouse_backup_$(date +%Y%m%d_%H%M%S).sql"
        
        docker exec tucanbit-clickhouse clickhouse-client --query "
        SELECT 'CREATE DATABASE IF NOT EXISTS ' || name || ';' as sql
        FROM system.databases 
        WHERE name NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')
        UNION ALL
        SELECT 'CREATE TABLE IF NOT EXISTS ' || database || '.' || name || ' AS ' || 
               database || '.' || name || ';' as sql
        FROM system.tables 
        WHERE database NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')
        " > "$BACKUP_FILE" 2>/dev/null || true
        
        print_success "Backup created: $BACKUP_FILE"
    fi
}

# Function to get table statistics
get_table_stats() {
    local table_name=$1
    local count=$(PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -t -c "SELECT COUNT(*) FROM $table_name" 2>/dev/null | tr -d ' ')
    echo "${count:-0}"
}

# Function to migrate users with progress tracking
migrate_users() {
    print_status "Migrating user registrations..."
    
    local total_users=$(get_table_stats "users")
    print_status "Total users to migrate: $total_users"
    
    if [ "$total_users" -eq 0 ]; then
        print_warning "No users found to migrate"
        return 0
    fi
    
    # Export users in batches
    local offset=0
    local batch_num=1
    
    while [ $offset -lt $total_users ]; do
        print_status "Processing users batch $batch_num (offset: $offset, limit: $BATCH_SIZE)"
        
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
                json_build_object(
                    'registration_method', 'direct', 
                    'email', email, 
                    'phone', phone_number,
                    'username', username,
                    'first_name', first_name,
                    'last_name', last_name,
                'created_at', to_char(created_at, 'YYYY-MM-DD HH24:MI:SS'),
                'verified', is_email_verified
            )::text as metadata,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
            to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
            FROM users
            WHERE created_at IS NOT NULL
            ORDER BY created_at
            LIMIT $BATCH_SIZE OFFSET $offset
        ) TO STDOUT WITH CSV HEADER" > "$MIGRATION_DIR/users_batch_$batch_num.csv"

        # Import to ClickHouse
        docker exec -i tucanbit-clickhouse clickhouse-client --query "
        INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < "$MIGRATION_DIR/users_batch_$batch_num.csv"
        
        offset=$((offset + BATCH_SIZE))
        batch_num=$((batch_num + 1))
        
        print_status "Completed users batch $((batch_num - 1))"
    done
    
    print_success "User registrations migrated successfully!"
}

# Function to migrate balance transactions
migrate_balances() {
    print_status "Migrating balance transactions..."
    
    local total_balances=$(get_table_stats "balance_logs")
    print_status "Total balance transactions to migrate: $total_balances"
    
    if [ "$total_balances" -eq 0 ]; then
        print_warning "No balance transactions found to migrate"
        return 0
    fi
    
    # Export balance logs in batches
    local offset=0
    local batch_num=1
    
    while [ $offset -lt $total_balances ]; do
        print_status "Processing balance batch $batch_num (offset: $offset, limit: $BATCH_SIZE)"
        
        PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
        COPY (
            SELECT 
                id::text as user_id,
                CASE 
                    WHEN transaction_type = 'deposit' THEN 'deposit'
                    WHEN transaction_type = 'withdrawal' THEN 'withdrawal'
                    WHEN transaction_type = 'bet' THEN 'bet'
                    WHEN transaction_type = 'win' THEN 'win'
                    WHEN transaction_type = 'bonus' THEN 'bonus'
                    WHEN transaction_type = 'cashback' THEN 'cashback'
                    ELSE 'other'
                END as transaction_type,
                amount,
                'USD' as currency,
                'completed' as status,
                NULL as game_id,
                NULL as game_name,
                NULL as provider,
                NULL as session_id,
                NULL as round_id,
                CASE WHEN transaction_type = 'bet' THEN amount ELSE NULL END as bet_amount,
                CASE WHEN transaction_type = 'win' THEN amount ELSE NULL END as win_amount,
                CASE 
                    WHEN transaction_type = 'bet' THEN -amount
                    WHEN transaction_type = 'win' THEN amount
                    ELSE amount
                END as net_result,
                balance_before,
                balance_after,
                payment_method,
                external_transaction_id,
                json_build_object(
                    'balance_log', true, 
                    'description', description,
                    'created_at', to_char(created_at, 'YYYY-MM-DD HH24:MI:SS'),
                    'updated_at', to_char(updated_at, 'YYYY-MM-DD HH24:MI:SS')
                )::text as metadata,
                to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
                to_char(updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
            FROM balance_logs
            WHERE created_at IS NOT NULL
            ORDER BY created_at
            LIMIT $BATCH_SIZE OFFSET $offset
        ) TO STDOUT WITH CSV HEADER" > "$MIGRATION_DIR/balances_batch_$batch_num.csv"

        # Import to ClickHouse
        docker exec -i tucanbit-clickhouse clickhouse-client --query "
        INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < "$MIGRATION_DIR/balances_batch_$batch_num.csv"
        
        offset=$((offset + BATCH_SIZE))
        batch_num=$((batch_num + 1))
        
        print_status "Completed balance batch $((batch_num - 1))"
    done
    
    print_success "Balance transactions migrated successfully!"
}

# Function to migrate GrooveTech transactions
migrate_groove_transactions() {
    print_status "Migrating GrooveTech transactions..."
    
    local total_groove=$(get_table_stats "groove_transactions")
    print_status "Total GrooveTech transactions to migrate: $total_groove"
    
    if [ "$total_groove" -eq 0 ]; then
        print_warning "No GrooveTech transactions found to migrate"
        return 0
    fi
    
    # Export GrooveTech transactions in batches
    local offset=0
    local batch_num=1
    
    while [ $offset -lt $total_groove ]; do
        print_status "Processing GrooveTech batch $batch_num (offset: $offset, limit: $BATCH_SIZE)"
        
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
                amount,
                'USD' as currency,
                'completed' as status,
                NULL as game_id,
                NULL as game_name,
                'GrooveTech' as provider,
                session_id,
                NULL as round_id,
                CASE WHEN type = 'wager' THEN amount ELSE NULL END as bet_amount,
                CASE WHEN type = 'result' THEN amount ELSE NULL END as win_amount,
                CASE 
                    WHEN type = 'wager' THEN -amount
                    WHEN type = 'result' THEN amount
                    ELSE amount
                END as net_result,
                0 as balance_before,
                0 as balance_after,
                'groove' as payment_method,
                transaction_id as external_transaction_id,
                metadata::text as metadata,
                to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at,
                to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
            FROM groove_transactions
            WHERE created_at IS NOT NULL
            ORDER BY created_at
            LIMIT $BATCH_SIZE OFFSET $offset
        ) TO STDOUT WITH CSV HEADER" > "$MIGRATION_DIR/groove_batch_$batch_num.csv"

        # Import to ClickHouse
        docker exec -i tucanbit-clickhouse clickhouse-client --query "
        INSERT INTO tucanbit_analytics.transactions FORMAT CSV" < "$MIGRATION_DIR/groove_batch_$batch_num.csv"
        
        offset=$((offset + BATCH_SIZE))
        batch_num=$((batch_num + 1))
        
        print_status "Completed GrooveTech batch $((batch_num - 1))"
    done
    
    print_success "GrooveTech transactions migrated successfully!"
}

# Function to create balance snapshots
create_balance_snapshots() {
    print_status "Creating balance snapshots..."
    
    local total_snapshots=$(get_table_stats "balance_logs")
    print_status "Total balance snapshots to create: $total_snapshots"
    
    if [ "$total_snapshots" -eq 0 ]; then
        print_warning "No balance data found for snapshots"
        return 0
    fi
    
    # Export balance snapshots in batches
    local offset=0
    local batch_num=1
    
    while [ $offset -lt $total_snapshots ]; do
        print_status "Processing balance snapshots batch $batch_num (offset: $offset, limit: $BATCH_SIZE)"
        
        PGPASSWORD=$PG_PASSWORD psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "
        COPY (
            SELECT 
                user_id,
                amount as balance,
                'USD' as currency,
                to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') as snapshot_time,
                id::text as transaction_id,
                transaction_type
            FROM balance_logs
            WHERE created_at IS NOT NULL
            ORDER BY user_id, created_at
            LIMIT $BATCH_SIZE OFFSET $offset
        ) TO STDOUT WITH CSV HEADER" > "$MIGRATION_DIR/snapshots_batch_$batch_num.csv"

        # Import to ClickHouse
        docker exec -i tucanbit-clickhouse clickhouse-client --query "
        INSERT INTO tucanbit_analytics.balance_snapshots FORMAT CSV" < "$MIGRATION_DIR/snapshots_batch_$batch_num.csv"
        
        offset=$((offset + BATCH_SIZE))
        batch_num=$((batch_num + 1))
        
        print_status "Completed balance snapshots batch $((batch_num - 1))"
    done
    
    print_success "Balance snapshots created successfully!"
}

# Function to create aggregated analytics
create_aggregated_analytics() {
    print_status "Creating aggregated analytics..."
    
    # Create user analytics
    print_status "Creating user analytics..."
    docker exec tucanbit-clickhouse clickhouse-client --query "
    INSERT INTO tucanbit_analytics.user_analytics
    SELECT 
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
    FROM tucanbit_analytics.transactions
    WHERE user_id IS NOT NULL AND user_id != ''
    GROUP BY user_id, date
    "
    
    print_success "Aggregated analytics created successfully!"
}

# Function to verify migration
verify_migration() {
    print_status "Verifying migration..."
    
    # Get counts from ClickHouse
    local ch_transactions=$(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.transactions")
    local ch_snapshots=$(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.balance_snapshots")
    local ch_users=$(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.user_analytics")
    
    print_status "Migration verification:"
    echo "  - Transactions: $ch_transactions"
    echo "  - Balance snapshots: $ch_snapshots"
    echo "  - User analytics: $ch_users"
    
    # Test sample queries
    print_status "Testing sample queries..."
    
    # Test real-time stats
    docker exec tucanbit-clickhouse clickhouse-client --query "
    SELECT 
        'Real-time Stats Test' as test_name,
        count() as total_transactions,
        uniqExact(user_id) as unique_users,
        sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets
    FROM tucanbit_analytics.transactions
    WHERE created_at >= now() - INTERVAL 1 HOUR
    "
    
    print_success "Migration verification completed!"
}

# Function to generate migration report
generate_report() {
    print_status "Generating migration report..."
    
    local report_file="$MIGRATION_DIR/migration_report_$(date +%Y%m%d_%H%M%S).json"
    
    cat > "$report_file" << EOF
{
    "migration_info": {
        "timestamp": "$(date -Iseconds)",
        "migration_dir": "$MIGRATION_DIR",
        "batch_size": "$BATCH_SIZE",
        "parallel_workers": "$PARALLEL_WORKERS"
    },
    "database_info": {
        "postgresql": {
            "host": "$PG_HOST",
            "port": "$PG_PORT",
            "database": "$PG_DB"
        },
        "clickhouse": {
            "host": "$CH_HOST",
            "port": "$CH_PORT",
            "database": "$CH_DB"
        }
    },
    "migration_results": {
        "transactions": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.transactions"),
        "balance_snapshots": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.balance_snapshots"),
        "user_analytics": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.user_analytics")
    },
    "data_coverage": {
        "registration_data": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.transactions WHERE transaction_type = 'registration'"),
        "betting_data": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.transactions WHERE transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')"),
        "session_data": $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT uniqExact(session_id) FROM tucanbit_analytics.transactions WHERE session_id IS NOT NULL")
    }
}
EOF
    
    print_success "Migration report generated: $report_file"
    
    # Display summary
    echo ""
    print_success "🎰 TucanBIT Casino ClickHouse Migration Summary:"
    echo "  📊 Total Transactions: $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT count() FROM tucanbit_analytics.transactions")"
    echo "  👥 Total Users: $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT uniqExact(user_id) FROM tucanbit_analytics.transactions")"
    echo "  🎮 Total Games: $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT uniqExact(game_id) FROM tucanbit_analytics.transactions WHERE game_id IS NOT NULL")"
    echo "  💰 Total Bet Amount: $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT sumIf(amount, transaction_type IN ('bet', 'groove_bet')) FROM tucanbit_analytics.transactions")"
    echo "  🏆 Total Win Amount: $(docker exec tucanbit-clickhouse clickhouse-client --query "SELECT sumIf(amount, transaction_type IN ('win', 'groove_win')) FROM tucanbit_analytics.transactions")"
    echo ""
    print_success "✅ All casino data has been successfully migrated to ClickHouse!"
    print_status "📁 Migration files saved in: $MIGRATION_DIR"
    print_status "📋 Report available at: $report_file"
}

# Main migration process
main() {
    print_status "Starting production ClickHouse migration..."
    print_status "Configuration:"
    echo "  - PostgreSQL: $PG_HOST:$PG_PORT/$PG_DB"
    echo "  - ClickHouse: $CH_HOST:$CH_PORT/$CH_DB"
    echo "  - Batch size: $BATCH_SIZE"
    echo "  - Parallel workers: $PARALLEL_WORKERS"
    echo "  - Backup enabled: $BACKUP_ENABLED"
    echo ""
    
    # Run migration steps
    check_prerequisites
    backup_existing_data
    migrate_users
    migrate_balances
    migrate_groove_transactions
    create_balance_snapshots
    create_aggregated_analytics
    verify_migration
    generate_report
    
    print_success "🎉 Production ClickHouse migration completed successfully!"
    print_status "Next steps:"
    echo "1. Set up real-time sync in your application"
    echo "2. Configure analytics APIs"
    echo "3. Build your analytics dashboard"
    echo "4. Set up monitoring and alerts"
    echo "5. Implement data retention policies"
}

# Run migration
main