-- Rollback crypto wallet authentication tables
-- Migration: 20250827104500_crypto_wallet_auth.down.sql

-- Drop functions first
DROP FUNCTION IF EXISTS clean_expired_crypto_wallet_challenges();
DROP FUNCTION IF EXISTS get_user_wallets(UUID);

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_crypto_wallet_connections_updated_at ON crypto_wallet_connections;

-- Drop functions
DROP FUNCTION IF EXISTS update_crypto_wallet_connections_updated_at();

-- Remove crypto wallet fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS primary_wallet_address;
ALTER TABLE users DROP COLUMN IF EXISTS wallet_verification_status;

-- Drop tables in reverse order (due to foreign key constraints)
DROP TABLE IF EXISTS crypto_wallet_auth_logs;
DROP TABLE IF EXISTS crypto_wallet_challenges;
DROP TABLE IF EXISTS crypto_wallet_connections; 