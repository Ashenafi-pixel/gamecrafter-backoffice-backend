-- Rollback test account flags
-- This migration removes the test account flags added in the up migration

-- Remove indexes
DROP INDEX IF EXISTS idx_users_is_test_account;
DROP INDEX IF EXISTS idx_groove_transactions_is_test;
DROP INDEX IF EXISTS idx_groove_game_sessions_is_test;
DROP INDEX IF EXISTS idx_bets_is_test_transaction;
DROP INDEX IF EXISTS idx_transactions_is_test_transaction;

-- Remove columns
ALTER TABLE transactions DROP COLUMN IF EXISTS is_test_transaction;
ALTER TABLE bets DROP COLUMN IF EXISTS is_test_transaction;
ALTER TABLE groove_game_sessions DROP COLUMN IF EXISTS is_test_game_session;
ALTER TABLE groove_transactions DROP COLUMN IF EXISTS is_test_transaction;
ALTER TABLE users DROP COLUMN IF EXISTS is_test_account;