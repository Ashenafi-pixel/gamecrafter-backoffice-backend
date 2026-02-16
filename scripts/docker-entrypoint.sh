#!/bin/bash
set -e

echo "=== Waiting for services ==="
./wait-for-it.sh db:5432 -t 60 --
./wait-for-it.sh kafka:9092 -t 60 --
./wait-for-it.sh tucanbit-clickhouse:9000 -t 60 --
sleep 5

# Disable exit on error for migration (it's idempotent, errors are expected)
set +e

echo "=== Verifying database connection ==="
export PGPASSWORD=5kj0YmV5FKKpU9D50B7yH5A
DB_EXISTS=$(psql -h db -U tucanbit -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'tucanbit'" | xargs)

if [ -z "$DB_EXISTS" ]; then
    echo "⚠️  Warning: Database tucanbit does not exist"
    echo "Creating database tucanbit..."
    psql -h db -U tucanbit -d postgres -c "CREATE DATABASE tucanbit"
    echo "Database created successfully"
else
    echo "✅ Database tucanbit exists"
fi

echo "=== Running all database migrations with go-migrate ==="
MIGRATE_DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@db:5432/tucanbit?sslmode=disable"

# Check if migrate tool exists and is executable
if [ -f /usr/local/bin/migrate ] && [ -x /usr/local/bin/migrate ]; then
    echo "Using go-migrate to run all migrations from ./migrations directory"
    echo "Migration order (by timestamp):"
    echo "  1. sync_database_schema (20251101000000) - Creates users table + all base tables"
    echo "  2. affiliate_referal_track (20251107183541) - Needs users table ✅"
    echo "  3. alter affiliate_referal_track (20251112054541)"
    echo ""
    
    # Check for dirty migration state and clean it if needed
    echo "=== Checking migration state ==="
    
    # Check schema_migrations table directly for dirty state (more reliable)
    DIRTY_CHECK=$(psql -h db -U tucanbit -d tucanbit -tc "SELECT version, dirty FROM schema_migrations LIMIT 1;" 2>/dev/null)
    
    if [ -n "$DIRTY_CHECK" ]; then
        # Parse version and dirty flag (format: "version | dirty")
        MIGRATION_VERSION=$(echo "$DIRTY_CHECK" | awk '{print $1}' | tr -d ' ')
        IS_DIRTY=$(echo "$DIRTY_CHECK" | awk '{print $NF}' | tr -d ' ')
        
        if [ "$IS_DIRTY" = "t" ] || [ "$IS_DIRTY" = "true" ] || [ "$IS_DIRTY" = "1" ]; then
            echo "⚠️  WARNING: Database migration state is DIRTY (version: $MIGRATION_VERSION)"
            echo "   This usually means a migration was interrupted."
            echo "   Attempting to clean the dirty migration state..."
            
            # Check if critical tables exist (indicating migration completed)
            CRITICAL_TABLES_EXIST=$(psql -h db -U tucanbit -d tucanbit -tc "
                SELECT COUNT(*) FROM information_schema.tables 
                WHERE table_schema = 'public' 
                AND table_name IN ('users', 'player_self_protection_settings', 'schema_migrations')
            " 2>/dev/null | xargs)
            
            if [ "$CRITICAL_TABLES_EXIST" = "3" ] || [ "$CRITICAL_TABLES_EXIST" = "2" ]; then
                echo "   ✅ Critical tables exist - migration likely completed successfully"
                echo "   Force-setting migration version to $MIGRATION_VERSION to mark as clean..."
                
                # Force set the version to mark it as clean (not dirty)
                if /usr/local/bin/migrate -path ./migrations -database "$MIGRATE_DB_URL" force "$MIGRATION_VERSION" 2>/dev/null; then
                    echo "   ✅ Successfully cleaned dirty migration state"
                else
                    # Fallback: directly update the database
                    echo "   Attempting direct database update..."
                    psql -h db -U tucanbit -d tucanbit -c "UPDATE schema_migrations SET dirty = false WHERE version = $MIGRATION_VERSION;" 2>/dev/null
                    if [ $? -eq 0 ]; then
                        echo "   ✅ Successfully cleaned dirty state via direct update"
                    else
                        echo "   ❌ Failed to clean dirty migration state. Manual intervention required."
                        echo "   Run: UPDATE schema_migrations SET dirty = false WHERE version = $MIGRATION_VERSION;"
                    fi
                fi
            else
                echo "   ⚠️  Critical tables missing - migration may not have completed"
                echo "   Attempting to rollback to previous version..."
                
                # Try to rollback one version (if version > 0)
                if [ "$MIGRATION_VERSION" -gt 0 ] 2>/dev/null; then
                    PREV_VERSION=$((MIGRATION_VERSION - 1))
                    if /usr/local/bin/migrate -path ./migrations -database "$MIGRATE_DB_URL" force "$PREV_VERSION" 2>/dev/null; then
                        echo "   ✅ Rolled back to version $PREV_VERSION"
                    else
                        # Direct database update as fallback
                        psql -h db -U tucanbit -d tucanbit -c "UPDATE schema_migrations SET version = $PREV_VERSION, dirty = false;" 2>/dev/null
                        if [ $? -eq 0 ]; then
                            echo "   ✅ Rolled back to version $PREV_VERSION via direct update"
                        else
                            echo "   ❌ Could not automatically fix dirty state"
                            echo "   Manual intervention required. Check schema_migrations table."
                        fi
                    fi
                else
                    echo "   ⚠️  Cannot rollback from version 0. Resetting migration state..."
                    psql -h db -U tucanbit -d tucanbit -c "DELETE FROM schema_migrations;" 2>/dev/null
                    echo "   ✅ Migration state reset. Migrations will run from beginning."
                fi
            fi
        else
            echo "   ✅ Migration state is clean (version: $MIGRATION_VERSION)"
        fi
    else
        # Also try migrate version command as fallback
        MIGRATION_VERSION=$(/usr/local/bin/migrate -path ./migrations -database "$MIGRATE_DB_URL" version 2>&1)
        MIGRATION_VERSION_EXIT=$?
        
        if [ $MIGRATION_VERSION_EXIT -eq 0 ]; then
            echo "   ✅ Migration state is clean (version: $MIGRATION_VERSION)"
        elif echo "$MIGRATION_VERSION" | grep -q "dirty"; then
            echo "⚠️  WARNING: Database migration state is DIRTY"
            echo "   Attempting to extract version and clean..."
            STUCK_VERSION=$(echo "$MIGRATION_VERSION" | grep -oE '[0-9]+' | head -1)
            if [ -n "$STUCK_VERSION" ]; then
                echo "   Detected stuck version: $STUCK_VERSION"
                if /usr/local/bin/migrate -path ./migrations -database "$MIGRATE_DB_URL" force "$STUCK_VERSION" 2>/dev/null; then
                    echo "   ✅ Successfully cleaned dirty migration state"
                else
                    echo "   ❌ Could not automatically fix. Manual intervention required."
                fi
            fi
        else
            echo "   ℹ️  No migration state found - this is a fresh database"
        fi
    fi
    
    echo ""
    echo "=== Running migrations ==="
    
    # Run all migrations - go-migrate handles ordering by version number
    if /usr/local/bin/migrate -path ./migrations -database "$MIGRATE_DB_URL" up; then
        echo "✅ All migrations completed successfully"
    else
        MIGRATE_EXIT_CODE=$?
        # Exit code 0 = success, 1 = error, 2+ = no change needed
        if [ $MIGRATE_EXIT_CODE -eq 0 ]; then
            echo "✅ All migrations completed successfully"
        elif [ $MIGRATE_EXIT_CODE -eq 1 ]; then
            echo "⚠️  Migration error occurred (exit code: $MIGRATE_EXIT_CODE)"
            echo "   Check logs above for details"
        else
            echo "✅ Migrations up to date (exit code: $MIGRATE_EXIT_CODE)"
        fi
    fi
    
    echo ""
    echo "=== Verifying critical tables exist ==="
    
    # List of critical tables that should exist (from sync_database_schema migration)
    CRITICAL_TABLES=(
        "users"
        "player_self_protection_settings"
        "player_self_protection_activity_logs"
        "player_gaming_time_tracking"
        "player_excluded_games"
        "schema_migrations"
    )
    
    MISSING_TABLES=()
    
    # Check each critical table
    for table in "${CRITICAL_TABLES[@]}"; do
        TABLE_EXISTS=$(psql -h db -U tucanbit -d tucanbit -tc "
            SELECT COUNT(*) FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = '$table'
        " 2>/dev/null | xargs)
        
        if [ "$TABLE_EXISTS" != "1" ]; then
            echo "⚠️  Critical table missing: $table"
            MISSING_TABLES+=("$table")
        fi
    done
    
    # If any tables are missing, recreate them from the sync_database_schema migration
    if [ ${#MISSING_TABLES[@]} -gt 0 ]; then
        echo "⚠️  Found ${#MISSING_TABLES[@]} missing table(s). Attempting to recreate from migration..."
        
        SYNC_MIGRATION="./migrations/20251101000000_sync_database_schema.up.sql"
        if [ -f "$SYNC_MIGRATION" ]; then
            echo "   Running sync_database_schema migration to recreate missing tables..."
            # Run the migration file - CREATE TABLE IF NOT EXISTS will only create missing tables
            if psql -h db -U tucanbit -d tucanbit -f "$SYNC_MIGRATION" 2>/dev/null; then
                echo "   ✅ Sync migration executed successfully"
                
                # Verify tables were created
                RECREATED=0
                for table in "${MISSING_TABLES[@]}"; do
                    TABLE_EXISTS=$(psql -h db -U tucanbit -d tucanbit -tc "
                        SELECT COUNT(*) FROM information_schema.tables 
                        WHERE table_schema = 'public' AND table_name = '$table'
                    " 2>/dev/null | xargs)
                    
                    if [ "$TABLE_EXISTS" = "1" ]; then
                        echo "   ✅ Table '$table' recreated successfully"
                        RECREATED=$((RECREATED + 1))
                    else
                        echo "   ❌ Table '$table' still missing after migration"
                    fi
                done
                
                if [ $RECREATED -eq ${#MISSING_TABLES[@]} ]; then
                    echo "✅ All missing tables recreated successfully"
                else
                    echo "⚠️  Some tables could not be recreated. Manual intervention may be required."
                fi
            else
                echo "   ❌ Failed to run sync migration. Manual intervention required."
            fi
        else
            echo "   ❌ Sync migration file not found: $SYNC_MIGRATION"
        fi
    else
        echo "✅ All critical tables exist"
    fi
else
    echo "❌ ERROR: migrate tool not found in PATH"
    echo "   Cannot run migrations. Please ensure go-migrate is installed."
    echo "   Falling back to manual migration execution..."
    
    # Fallback: Try to run sync_database_schema manually
    SYNC_MIGRATION="./migrations/20251101000000_sync_database_schema.up.sql"
    if [ -f "$SYNC_MIGRATION" ]; then
        echo "Running schema sync migration manually..."
        psql -h db -U tucanbit -d tucanbit -f "$SYNC_MIGRATION" || echo "⚠️  Manual migration had errors"
    fi
fi

echo "=== Environment variables ==="
printenv | grep KAFKA || true

# Re-enable exit on error for application startup
set -e

echo "=== Starting application ==="
exec ./tucanbit

