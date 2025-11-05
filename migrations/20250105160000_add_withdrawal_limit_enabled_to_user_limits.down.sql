-- Rollback: Remove withdrawal_limit_enabled column from user_limits table

ALTER TABLE user_limits
DROP COLUMN IF EXISTS withdrawal_limit_enabled;

