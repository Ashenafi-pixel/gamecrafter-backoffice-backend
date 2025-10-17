-- Admin Activity Logs Migration Rollback
-- This migration removes the admin activity logging system

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_admin_activity_logs_updated_at ON admin_activity_logs;

-- Drop function
DROP FUNCTION IF EXISTS update_admin_activity_logs_updated_at();

-- Drop tables in reverse order
DROP TABLE IF EXISTS admin_activity_actions;
DROP TABLE IF EXISTS admin_activity_categories;
DROP TABLE IF EXISTS admin_activity_logs;
