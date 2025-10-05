-- Rollback migration for cashback system
DROP TRIGGER IF EXISTS trigger_update_user_level_stats ON cashback_earnings;
DROP FUNCTION IF EXISTS update_user_level_stats();
DROP FUNCTION IF EXISTS check_and_update_user_level(UUID);

-- Drop indexes
DROP INDEX IF EXISTS idx_cashback_promotions_active;
DROP INDEX IF EXISTS idx_game_house_edges_game_type;
DROP INDEX IF EXISTS idx_cashback_claims_status;
DROP INDEX IF EXISTS idx_cashback_claims_user_id;
DROP INDEX IF EXISTS idx_cashback_earnings_expires_at;
DROP INDEX IF EXISTS idx_cashback_earnings_created_at;
DROP INDEX IF EXISTS idx_cashback_earnings_status;
DROP INDEX IF EXISTS idx_cashback_earnings_user_id;
DROP INDEX IF EXISTS idx_user_levels_current_level;
DROP INDEX IF EXISTS idx_user_levels_user_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS cashback_promotions;
DROP TABLE IF EXISTS game_house_edges;
DROP TABLE IF EXISTS cashback_claims;
DROP TABLE IF EXISTS cashback_earnings;
DROP TABLE IF EXISTS cashback_tiers;
DROP TABLE IF EXISTS user_levels;

-- Remove house_edge column from bets table if it was added
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bets' AND column_name = 'house_edge') THEN
        ALTER TABLE bets DROP COLUMN house_edge;
    END IF;
END $$;