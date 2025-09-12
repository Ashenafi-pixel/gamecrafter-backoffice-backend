-- Drop GrooveTech integration tables
-- This migration removes all GrooveTech related tables and functions

-- Drop functions first
DROP FUNCTION IF EXISTS get_groove_account_summary(VARCHAR(255));
DROP FUNCTION IF EXISTS cleanup_expired_groove_sessions();
DROP FUNCTION IF EXISTS update_groove_accounts_updated_at();

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_groove_accounts_updated_at ON groove_accounts;

-- Drop tables in reverse order (due to foreign key constraints)
DROP TABLE IF EXISTS groove_game_sessions;
DROP TABLE IF EXISTS groove_transactions;
DROP TABLE IF EXISTS groove_accounts;