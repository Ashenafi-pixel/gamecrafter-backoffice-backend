#!/bin/bash

# Script to sync all pages from dev to prod and assign them to admin users
# This will:
# 1. Export all pages from dev database
# 2. Import them into prod database (with conflict handling)
# 3. Assign all pages to all admin users in prod

SSH_KEY="$HOME/.ssh/smiss/tucanbit-dev.pem"
SSH_HOST="admin@18.231.108.116"
DEV_DB_HOST="rds-dev.internal.tucanbit.local"
PROD_DB_HOST="rds-prod.internal.tucanbit.local"
DB_USER="tucanbit"
DB_NAME="tucanbit"

echo "=========================================="
echo "Syncing Pages from Dev to Prod"
echo "=========================================="
echo ""

# Step 1: Export pages from dev
echo "📥 Step 1: Exporting pages from dev database..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $DEV_DB_HOST -U $DB_USER -d $DB_NAME -t -c \"
SELECT 
    'INSERT INTO pages (id, path, label, parent_id, icon, created_at, updated_at) VALUES (' ||
    quote_literal(id::text) || ', ' ||
    quote_literal(path) || ', ' ||
    quote_literal(label) || ', ' ||
    COALESCE(quote_literal(parent_id::text), 'NULL') || ', ' ||
    COALESCE(quote_literal(icon), 'NULL') || ', ' ||
    quote_literal(created_at::text) || ', ' ||
    quote_literal(updated_at::text) || ') ON CONFLICT (path) DO UPDATE SET label = EXCLUDED.label, parent_id = EXCLUDED.parent_id, icon = EXCLUDED.icon, updated_at = EXCLUDED.updated_at;'
FROM pages
ORDER BY 
    CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END,
    path;
\"" > /tmp/pages_export.sql

if [ $? -ne 0 ]; then
    echo "❌ Failed to export pages from dev"
    exit 1
fi

echo "✅ Pages exported successfully"
echo ""

# Step 2: Import pages into prod
echo "📤 Step 2: Importing pages into prod database..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $PROD_DB_HOST -U $DB_USER -d $DB_NAME -f /tmp/pages_export.sql"

if [ $? -eq 0 ]; then
    echo "✅ Pages imported successfully"
else
    echo "❌ Failed to import pages into prod"
    exit 1
fi

echo ""

# Step 3: Assign all pages to all admin users
echo "🔐 Step 3: Assigning all pages to admin users..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $PROD_DB_HOST -U $DB_USER -d $DB_NAME -c \"
INSERT INTO user_allowed_pages (user_id, page_id)
SELECT DISTINCT
    u.id as user_id,
    p.id as page_id
FROM users u
CROSS JOIN pages p
WHERE u.is_admin = true
  AND u.user_type = 'ADMIN'
ON CONFLICT (user_id, page_id) DO NOTHING;
\""

if [ $? -eq 0 ]; then
    echo "✅ Pages assigned to admin users successfully"
else
    echo "❌ Failed to assign pages to admin users"
    exit 1
fi

echo ""

# Step 4: Verification
echo "📊 Step 4: Verification..."
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $PROD_DB_HOST -U $DB_USER -d $DB_NAME -c \"
SELECT 
    'Total Pages' as metric,
    COUNT(*)::text as value
FROM pages
UNION ALL
SELECT 
    'Admin Users' as metric,
    COUNT(*)::text as value
FROM users
WHERE is_admin = true AND user_type = 'ADMIN'
UNION ALL
SELECT 
    'Page Assignments' as metric,
    COUNT(*)::text as value
FROM user_allowed_pages uap
JOIN users u ON u.id = uap.user_id
WHERE u.is_admin = true AND u.user_type = 'ADMIN';
\""

echo ""
echo "=========================================="
echo "✅ Pages sync completed successfully!"
echo "=========================================="

