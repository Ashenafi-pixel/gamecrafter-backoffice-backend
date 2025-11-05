-- Migration: Add withdrawal_limit_enabled column to user_limits table
-- This allows storing whether the withdrawal limit is enabled separately from the limit amount

ALTER TABLE user_limits
ADD COLUMN withdrawal_limit_enabled BOOLEAN NOT NULL DEFAULT true;

-- Update existing records to have the limit enabled
UPDATE user_limits
SET withdrawal_limit_enabled = true
WHERE limit_type = 'withdrawal' AND withdrawal_limit_enabled IS NULL;

COMMENT ON COLUMN user_limits.withdrawal_limit_enabled IS 'Whether the withdrawal limit is currently enabled for this user';

