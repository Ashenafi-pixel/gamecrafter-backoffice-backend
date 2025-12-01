# Generating Additional Seed Data

This guide explains how to generate the additional seed data (Permissions, Role Permissions, Pages, and Admin Activity Actions) and add it to the seed migration.

## Prerequisites

1. Database connection must be available (SSH tunnel or direct connection)
2. Python 3 installed
3. PostgreSQL client tools (`psql`) installed

## Steps

### 1. Ensure Database Connection

Make sure you can connect to the database:
```bash
PGPASSWORD="5kj0YmV5FKKpU9D50B7yH5A" psql -h localhost -p 5433 -U tucanbit -d tucanbit -c "SELECT 1;"
```

### 2. Generate Seed Data

Run the generation script:
```bash
cd /Users/oza/Developer/Upwork/Tucanbit/Tucanbit/tucanbit-back
python3 scripts/generate_seed_data.py > /tmp/additional_seed_data.sql
```

### 3. Review Generated Data

Check the generated file:
```bash
cat /tmp/additional_seed_data.sql
```

### 4. Add to Seed Migration

You have two options:

#### Option A: Append to Existing Migration

Extract each section and add it to the appropriate place in `migrations/20251201213100_seed_essential_data.up.sql`:

```bash
# Extract Permissions section
grep -A 1000 "^-- ============================================================================$" /tmp/additional_seed_data.sql | grep -A 1000 "11. PERMISSIONS" | head -250 > /tmp/permissions.sql

# Extract Role Permissions section
grep -A 1000 "^-- ============================================================================$" /tmp/additional_seed_data.sql | grep -A 1000 "12. ROLE PERMISSIONS" | head -250 > /tmp/role_permissions.sql

# Extract Pages section
grep -A 1000 "^-- ============================================================================$" /tmp/additional_seed_data.sql | grep -A 1000 "13. PAGES" | head -400 > /tmp/pages.sql

# Extract Admin Activity Actions section
grep -A 1000 "^-- ============================================================================$" /tmp/additional_seed_data.sql | grep -A 1000 "14. ADMIN ACTIVITY ACTIONS" | head -300 > /tmp/actions.sql
```

Then manually copy each section to replace the placeholder comments in the migration file.

#### Option B: Create Separate Migration

Create a new migration file that includes all the additional data:

```bash
cd /Users/oza/Developer/Upwork/Tucanbit/Tucanbit/tucanbit-back
migrate create -ext sql -dir ./migrations -seq seed_additional_data
```

Then copy the generated SQL to the new migration file.

### 5. Update Down Migration

After adding the data, update the down migration file (`20251201213100_seed_essential_data.down.sql`) to include DELETE statements for the new data. The script will generate IDs that need to be added to the DELETE statements.

## Troubleshooting

### Connection Issues

If you get connection errors:
1. Check if SSH tunnel is running (if using tunnel)
2. Verify database credentials in `scripts/generate_seed_data.py`
3. Test connection manually with `psql`

### Missing Columns

If the script fails due to missing columns:
1. Check the actual table structure: `\d table_name` in psql
2. Update the script queries to match your schema
3. The script automatically detects if `resource` and `action` columns exist in permissions table

### Empty Results

If no data is returned:
1. Verify the data exists in the database
2. Check if you're querying the correct database
3. Verify role IDs match (especially for role_permissions)

## Notes

- The script uses `ON CONFLICT DO NOTHING` to prevent errors if data already exists
- All UUIDs are preserved from the source database
- The script handles NULL values appropriately
- Generated SQL is PostgreSQL-compatible

