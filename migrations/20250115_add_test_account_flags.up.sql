-- Add test account flags to support test/live account separation
-- This migration adds is_test_account flags to users and related tables

-- Add is_test_account field to users table (default true for all existing users)
ALTER TABLE users ADD COLUMN is_test_account BOOLEAN DEFAULT true;

-- Add is_test_transaction field to groove_transactions table
ALTER TABLE groove_transactions ADD COLUMN is_test_transaction BOOLEAN DEFAULT true;

-- Add is_test_game_session field to groove_game_sessions table  
ALTER TABLE groove_game_sessions ADD COLUMN is_test_game_session BOOLEAN DEFAULT true;

-- Add is_test_transaction field to bets table
ALTER TABLE bets ADD COLUMN is_test_transaction BOOLEAN DEFAULT true;

-- Add is_test_transaction field to transactions table
ALTER TABLE transactions ADD COLUMN is_test_transaction BOOLEAN DEFAULT true;

-- Create indexes for better performance on test flags
CREATE INDEX IF NOT EXISTS idx_users_is_test_account ON users(is_test_account);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_is_test ON groove_transactions(is_test_transaction);
CREATE INDEX IF NOT EXISTS idx_groove_game_sessions_is_test ON groove_game_sessions(is_test_game_session);
CREATE INDEX IF NOT EXISTS idx_bets_is_test_transaction ON bets(is_test_transaction);
CREATE INDEX IF NOT EXISTS idx_transactions_is_test_transaction ON transactions(is_test_transaction);