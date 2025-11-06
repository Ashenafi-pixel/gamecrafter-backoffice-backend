-- Rollback Global Rakeback Override Migration
-- This migration removes the global rakeback override feature

-- Drop the index
DROP INDEX IF EXISTS idx_global_rakeback_override_enabled;

-- Drop the global rakeback override table
DROP TABLE IF EXISTS global_rakeback_override CASCADE;

-- Note: This rollback will permanently delete all global rakeback override
-- configuration and audit history. Make sure to backup data before running
-- this migration if you need to preserve the history.

