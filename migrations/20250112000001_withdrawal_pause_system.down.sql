-- Rollback migration for withdrawal pause system
-- This will remove all pause-related tables and columns

-- Drop indexes
DROP INDEX IF EXISTS idx_withdrawals_paused_at;
DROP INDEX IF EXISTS idx_withdrawals_requires_manual_review;
DROP INDEX IF EXISTS idx_withdrawals_is_paused;
DROP INDEX IF EXISTS idx_withdrawal_pause_logs_action_taken;
DROP INDEX IF EXISTS idx_withdrawal_pause_logs_paused_at;
DROP INDEX IF EXISTS idx_withdrawal_pause_logs_withdrawal_id;

-- Remove columns from withdrawals table
ALTER TABLE withdrawals DROP COLUMN IF EXISTS requires_manual_review;
ALTER TABLE withdrawals DROP COLUMN IF EXISTS paused_at;
ALTER TABLE withdrawals DROP COLUMN IF EXISTS pause_reason;
ALTER TABLE withdrawals DROP COLUMN IF EXISTS is_paused;

-- Drop tables
DROP TABLE IF EXISTS withdrawal_pause_logs;
DROP TABLE IF EXISTS withdrawal_thresholds;
DROP TABLE IF EXISTS withdrawal_pause_settings;





