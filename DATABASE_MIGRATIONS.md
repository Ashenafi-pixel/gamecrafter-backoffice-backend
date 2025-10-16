# Database Migration Summary

This document contains all the database schema changes made during the development session. Apply these migrations to your AWS server database to match the local development environment.

## Migration Date: 2025-10-08

### 1. Games Table Schema Updates

#### Add new columns to games table:
```sql
-- Add game_id column
ALTER TABLE games ADD COLUMN game_id VARCHAR(255);

-- Add internal_name column  
ALTER TABLE games ADD COLUMN internal_name VARCHAR(255);

-- Add provider column
ALTER TABLE games ADD COLUMN provider VARCHAR(255);

-- Add integration_partner column
ALTER TABLE games ADD COLUMN integration_partner VARCHAR(255);

-- Add name column (if not exists)
ALTER TABLE games ADD COLUMN name VARCHAR(255);
```

#### Create index on game_id for better performance:
```sql
CREATE INDEX IF NOT EXISTS idx_games_game_id ON games(game_id);
```

### 2. Groove Transactions Table Schema Updates

#### Add balance tracking columns:
```sql
-- Add balance_before column
ALTER TABLE groove_transactions ADD COLUMN balance_before NUMERIC(20,8) DEFAULT 0;

-- Add balance_after column
ALTER TABLE groove_transactions ADD COLUMN balance_after NUMERIC(20,8) DEFAULT 0;
```

### 3. Game Data Import

#### Import game data from CSV file:
The following Python script was used to import game data from `game 20250930-905.csv`:

```python
import psycopg2
import csv
import os

# Database connection parameters
DB_CONFIG = {
    'host': 'localhost',
    'port': '5433',  # Docker mapped port
    'database': 'tucanbit',
    'user': 'tucanbit',
    'password': 'tucanbit_password'  # From docker-compose.yaml
}

def import_games():
    conn = psycopg2.connect(**DB_CONFIG)
    cursor = conn.cursor()
    
    csv_file = 'game 20250930-905.csv'
    
    with open(csv_file, 'r', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        
        for row in reader:
            # Insert or update game record
            cursor.execute("""
                INSERT INTO games (game_id, internal_name, provider, integration_partner, name)
                VALUES (%s, %s, %s, %s, %s)
                ON CONFLICT (game_id) DO UPDATE SET
                    internal_name = EXCLUDED.internal_name,
                    provider = EXCLUDED.provider,
                    integration_partner = EXCLUDED.integration_partner,
                    name = EXCLUDED.name
            """, (
                row['game_id'],
                row['internal_name'], 
                row['provider'],
                'groovetech',  # Set integration_partner to groovetech
                row['name']
            ))
    
    conn.commit()
    cursor.close()
    conn.close()
    print("Game data imported successfully!")

if __name__ == "__main__":
    import_games()
```

#### Manual game addition:
```sql
-- Add specific game that was missing
INSERT INTO games (game_id, name, internal_name, provider, integration_partner)
VALUES ('82695', 'Sweet Bonanza', 'sweet_bonanza', 'Pragmatic Play', 'groovetech')
ON CONFLICT (game_id) DO UPDATE SET
    name = EXCLUDED.name,
    internal_name = EXCLUDED.internal_name,
    provider = EXCLUDED.provider,
    integration_partner = EXCLUDED.integration_partner;
```

### 4. ClickHouse Analytics Data Updates

#### Update existing transactions with game information:
```sql
-- Update transactions with missing game_id based on session_id lookup
ALTER TABLE tucanbit_analytics.transactions 
UPDATE game_id = '82695', game_name = 'Sweet Bonanza', provider = 'Pragmatic Play' 
WHERE session_id = '3818_c4e9a25a-78a6-4511-b0da-808ecc11f4a6' 
AND (game_id IS NULL OR game_id = '');

-- Update other sessions
ALTER TABLE tucanbit_analytics.transactions 
UPDATE game_id = '82695', game_name = 'Sweet Bonanza', provider = 'Pragmatic Play' 
WHERE session_id IN ('3818_9b092713-def5-48b8-9581-e7b513fc2c40', '3818_786691bc-30cf-453d-8bac-ff5cdfb233c4', '3818_bd53d624-a5d5-431b-9ef2-d34ffafd9421') 
AND (game_id IS NULL OR game_id = '');
```

### 5. Verification Queries

#### Check games table:
```sql
-- Verify games table structure
\d games

-- Check imported game data
SELECT COUNT(*) FROM games WHERE integration_partner = 'groovetech';

-- Check specific game
SELECT * FROM games WHERE game_id = '82695';
```

#### Check groove_transactions table:
```sql
-- Verify groove_transactions table structure
\d groove_transactions

-- Check balance columns exist
SELECT column_name, data_type FROM information_schema.columns 
WHERE table_name = 'groove_transactions' 
AND column_name IN ('balance_before', 'balance_after');
```

#### Check ClickHouse analytics:
```sql
-- Verify ClickHouse transactions have game information
SELECT game_id, game_name, provider, COUNT(*) 
FROM tucanbit_analytics.transactions 
WHERE game_id IS NOT NULL AND game_id != ''
GROUP BY game_id, game_name, provider;

-- Check balance information in analytics
SELECT id, balance_before, balance_after, amount 
FROM tucanbit_analytics.transactions 
WHERE balance_before > 0 OR balance_after > 0
LIMIT 5;
```

## Migration Order

1. **PostgreSQL Schema Updates** (games table, groove_transactions table)
2. **Game Data Import** (CSV import + manual additions)
3. **ClickHouse Data Updates** (update existing transactions)
4. **Verification** (run verification queries)

## Notes

- The `game 20250930-905.csv` file contains 2285 game records
- All games are imported with `integration_partner = 'groovetech'`
- Balance tracking is now available for all new GrooveTech transactions
- Existing ClickHouse data was updated to include game names and providers
- The analytics API now returns consistent game information for all transactions

## AWS Server Deployment

When deploying to AWS server:

1. Run the PostgreSQL migrations first
2. Import the game data using the Python script (update connection parameters)
3. Update ClickHouse data if needed
4. Verify all changes are applied correctly
5. Test the analytics endpoints to ensure game names and balance information are working

## Files Modified

- `internal/constant/dto/groove.go` - Added BalanceBefore/BalanceAfter to GrooveTransaction
- `internal/module/groove/groove.go` - Updated transaction processing to capture balances
- `internal/module/analytics/sync.go` - Updated analytics sync to use balance information
- `internal/storage/groove/groove.go` - Updated StoreTransaction to include balance columns
- `internal/storage/analytics/analytics.go` - Updated GetUserTransactions to handle NULL values
- `import_games.py` - Created script for game data import
