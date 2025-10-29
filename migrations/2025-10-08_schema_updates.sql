-- Database Migration Script for AWS Server
-- Date: 2025-10-08
-- Description: Apply all database schema changes made during development

-- =============================================
-- 1. GAMES TABLE SCHEMA UPDATES
-- =============================================

-- Add game_id column
ALTER TABLE games ADD COLUMN IF NOT EXISTS game_id VARCHAR(255);

-- Add internal_name column  
ALTER TABLE games ADD COLUMN IF NOT EXISTS internal_name VARCHAR(255);

-- Add provider column
ALTER TABLE games ADD COLUMN IF NOT EXISTS provider VARCHAR(255);

-- Add integration_partner column
ALTER TABLE games ADD COLUMN IF NOT EXISTS integration_partner VARCHAR(255);

-- Add name column (if not exists)
ALTER TABLE games ADD COLUMN IF NOT EXISTS name VARCHAR(255);

-- Create index on game_id for better performance
CREATE INDEX IF NOT EXISTS idx_games_game_id ON games(game_id);

-- =============================================
-- 2. GROOVE TRANSACTIONS TABLE SCHEMA UPDATES
-- =============================================

-- Add balance tracking columns
ALTER TABLE groove_transactions ADD COLUMN IF NOT EXISTS balance_before NUMERIC(20,8) DEFAULT 0;
ALTER TABLE groove_transactions ADD COLUMN IF NOT EXISTS balance_after NUMERIC(20,8) DEFAULT 0;

-- =============================================
-- 3. MANUAL GAME ADDITIONS
-- =============================================

-- Add specific game that was missing (Sweet Bonanza)
INSERT INTO games (game_id, name, internal_name, provider, integration_partner)
VALUES ('82695', 'Sweet Bonanza', 'sweet_bonanza', 'Pragmatic Play', 'groovetech')
ON CONFLICT (game_id) DO UPDATE SET
    name = EXCLUDED.name,
    internal_name = EXCLUDED.internal_name,
    provider = EXCLUDED.provider,
    integration_partner = EXCLUDED.integration_partner;

-- =============================================
-- 4. VERIFICATION QUERIES
-- =============================================

-- Check games table structure
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'games' 
ORDER BY ordinal_position;

-- Check groove_transactions table structure
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'groove_transactions' 
AND column_name IN ('balance_before', 'balance_after');

-- Check if Sweet Bonanza was added
SELECT * FROM games WHERE game_id = '82695';

-- Count total games
SELECT COUNT(*) as total_games FROM games;

-- Count groovetech games
SELECT COUNT(*) as groovetech_games FROM games WHERE integration_partner = 'groovetech';

-- =============================================
-- 5. COMMENTS
-- =============================================

COMMENT ON COLUMN games.game_id IS 'Unique game identifier from provider';
COMMENT ON COLUMN games.internal_name IS 'Internal game name used by the system';
COMMENT ON COLUMN games.provider IS 'Game provider (e.g., Pragmatic Play, Evolution)';
COMMENT ON COLUMN games.integration_partner IS 'Integration partner (e.g., groovetech)';
COMMENT ON COLUMN groove_transactions.balance_before IS 'User balance before transaction';
COMMENT ON COLUMN groove_transactions.balance_after IS 'User balance after transaction';

-- =============================================
-- MIGRATION COMPLETE
-- =============================================

-- Log migration completion
INSERT INTO migration_log (migration_name, applied_at, description) 
VALUES ('2025-10-08_schema_updates', NOW(), 'Added game and balance tracking columns')
ON CONFLICT (migration_name) DO UPDATE SET 
    applied_at = EXCLUDED.applied_at,
    description = EXCLUDED.description;

-- Note: If migration_log table doesn't exist, create it:
-- CREATE TABLE IF NOT EXISTS migration_log (
--     id SERIAL PRIMARY KEY,
--     migration_name VARCHAR(255) UNIQUE NOT NULL,
--     applied_at TIMESTAMP DEFAULT NOW(),
--     description TEXT
-- );
