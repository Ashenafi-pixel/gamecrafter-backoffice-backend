#!/bin/bash

# Script to run migrations on production database
# This runs:
# 1. Migration to create and assign the 5 new report pages to admin users
# 2. Migration to assign report permissions to admin role

SSH_KEY="$HOME/.ssh/smiss/tucanbit-dev.pem"
SSH_HOST="admin@18.231.108.116"
DB_HOST="rds-prod.internal.tucanbit.local"
DB_USER="tucanbit"
DB_NAME="tucanbit"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATIONS_DIR="$(cd "$SCRIPT_DIR/../migrations" && pwd)"

echo "=========================================="
echo "Running Production Migrations"
echo "=========================================="
echo ""

# Migration 1: Create and assign report pages
echo "📄 Step 1: Creating and assigning report pages to admin users..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f -" < "$MIGRATIONS_DIR/20250125000000_assign_performance_reports_to_admins.sql"

if [ $? -eq 0 ]; then
    echo "✅ Report pages migration completed successfully!"
else
    echo "❌ Failed to run report pages migration"
    exit 1
fi

echo ""
echo "=========================================="
echo ""

# Migration 2: Assign report permissions to admin role
echo "🔐 Step 2: Assigning report permissions to admin role..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f -" < "$MIGRATIONS_DIR/20250125000001_assign_report_permissions_to_admin_role.sql"

if [ $? -eq 0 ]; then
    echo "✅ Report permissions migration completed successfully!"
else
    echo "❌ Failed to run report permissions migration"
    exit 1
fi

echo ""
echo "=========================================="
echo "✅ All migrations completed successfully!"
echo "=========================================="

