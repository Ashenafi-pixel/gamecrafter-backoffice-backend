-- Migration: Drop user_2fa_methods table
-- Created: 2025-09-16
-- Description: Removes user_2fa_methods table

BEGIN;

-- Drop indexes first
DROP INDEX IF EXISTS idx_user_2fa_methods_user_id;
DROP INDEX IF EXISTS idx_user_2fa_methods_method;
DROP INDEX IF EXISTS idx_user_2fa_methods_enabled_at;

-- Drop table
DROP TABLE IF EXISTS user_2fa_methods;

COMMIT;
