# AWS Server Deployment Checklist

## Pre-Deployment Setup

### 1. Database Connection Configuration
Update the database connection parameters in the migration files:

**For SQL migration (`migrations/2025-10-08_schema_updates.sql`):**
- No changes needed - uses standard PostgreSQL commands

**For Python import script (`scripts/import_games_aws.py`):**
```python
DB_CONFIG = {
    'host': 'your-aws-rds-endpoint.amazonaws.com',  # Your RDS endpoint
    'port': '5432',  # Standard PostgreSQL port
    'database': 'tucanbit',
    'user': 'your-db-username',  # Your DB username
    'password': 'your-db-password'  # Your DB password
}
```

### 2. Required Files
Ensure you have these files on your AWS server:
- `migrations/2025-10-08_schema_updates.sql`
- `scripts/import_games_aws.py`
- `game 20250930-905.csv` (the game data file)

## Deployment Steps

### Step 1: Database Schema Migration
```bash
# Connect to your PostgreSQL database
psql -h your-aws-rds-endpoint.amazonaws.com -U your-db-username -d tucanbit

# Run the migration script
\i migrations/2025-10-08_schema_updates.sql
```

### Step 2: Game Data Import
```bash
# Install required Python packages
pip install psycopg2-binary

# Run the import script
python3 scripts/import_games_aws.py

# Or specify a different CSV file
python3 scripts/import_games_aws.py /path/to/your/game_data.csv
```

### Step 3: ClickHouse Data Updates (if applicable)
If you have ClickHouse running on AWS, update existing transaction data:

```sql
-- Connect to ClickHouse
clickhouse-client --host your-clickhouse-host --port 8123

-- Update transactions with missing game information
ALTER TABLE tucanbit_analytics.transactions 
UPDATE game_id = '82695', game_name = 'Sweet Bonanza', provider = 'Pragmatic Play' 
WHERE session_id IN ('3818_9b092713-def5-48b8-9581-e7b513fc2c40', '3818_786691bc-30cf-453d-8bac-ff5cdfb233c4', '3818_bd53d624-a5d5-431b-9ef2-d34ffafd9421') 
AND (game_id IS NULL OR game_id = '');
```

### Step 4: Verification
Run these queries to verify the migration:

```sql
-- Check games table
SELECT COUNT(*) FROM games WHERE integration_partner = 'groovetech';
SELECT * FROM games WHERE game_id = '82695';

-- Check groove_transactions table
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'groove_transactions' 
AND column_name IN ('balance_before', 'balance_after');

-- Check recent transactions have balance data
SELECT id, balance_before, balance_after, amount 
FROM groove_transactions 
WHERE balance_before > 0 OR balance_after > 0
ORDER BY created_at DESC 
LIMIT 5;
```

### Step 5: Application Deployment
Deploy your updated application code:

```bash
# Build and deploy your Go application
go build -o tucanbit-server cmd/main.go
./tucanbit-server
```

### Step 6: API Testing
Test the analytics endpoints to ensure everything works:

```bash
# Test analytics endpoint
curl -X GET "http://your-server:8080/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/transactions?limit=5" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" | jq '.data[] | {id, game_id, game_name, provider, balance_before, balance_after}'
```

## Rollback Plan

If something goes wrong, you can rollback the changes:

### Rollback Database Schema
```sql
-- Remove balance columns (if needed)
ALTER TABLE groove_transactions DROP COLUMN IF EXISTS balance_before;
ALTER TABLE groove_transactions DROP COLUMN IF EXISTS balance_after;

-- Remove game columns (if needed)
ALTER TABLE games DROP COLUMN IF EXISTS game_id;
ALTER TABLE games DROP COLUMN IF EXISTS internal_name;
ALTER TABLE games DROP COLUMN IF EXISTS provider;
ALTER TABLE games DROP COLUMN IF EXISTS integration_partner;
ALTER TABLE games DROP COLUMN IF EXISTS name;

-- Drop index
DROP INDEX IF EXISTS idx_games_game_id;
```

### Rollback Game Data
```sql
-- Remove imported games (if needed)
DELETE FROM games WHERE integration_partner = 'groovetech';
```

## Monitoring

After deployment, monitor:

1. **Application logs** for any errors
2. **Database performance** with new columns
3. **Analytics API response times**
4. **Game data accuracy** in analytics responses

## Troubleshooting

### Common Issues:

1. **Connection refused**: Check RDS security groups and VPC settings
2. **Permission denied**: Ensure database user has ALTER TABLE permissions
3. **CSV file not found**: Verify file path and permissions
4. **Import errors**: Check CSV file format and encoding

### Debug Commands:
```bash
# Check database connection
psql -h your-aws-rds-endpoint.amazonaws.com -U your-db-username -d tucanbit -c "SELECT version();"

# Check table structure
psql -h your-aws-rds-endpoint.amazonaws.com -U your-db-username -d tucanbit -c "\d games"

# Check recent logs
tail -f /var/log/your-app.log
```

## Success Criteria

✅ Database schema updated successfully  
✅ Game data imported (2285+ games)  
✅ Balance tracking working in analytics  
✅ Game names showing in transaction history  
✅ No application errors in logs  
✅ Analytics API responding correctly  

## Contact

If you encounter issues during deployment, refer to:
- Database migration logs
- Application error logs  
- This deployment checklist
- The detailed migration documentation in `DATABASE_MIGRATIONS.md`
