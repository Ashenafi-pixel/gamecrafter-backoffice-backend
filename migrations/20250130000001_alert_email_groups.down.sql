-- Drop alert email groups tables
DROP TRIGGER IF EXISTS trigger_update_alert_email_groups_updated_at ON alert_email_groups;
DROP FUNCTION IF EXISTS update_alert_email_groups_updated_at();
DROP INDEX IF EXISTS idx_alert_configurations_email_group_ids;
DROP INDEX IF EXISTS idx_alert_email_group_members_email;
DROP INDEX IF EXISTS idx_alert_email_group_members_group_id;
DROP INDEX IF EXISTS idx_alert_email_groups_created_by;
DROP INDEX IF EXISTS idx_alert_email_groups_name;
ALTER TABLE alert_configurations DROP COLUMN IF EXISTS email_group_ids;
DROP TABLE IF EXISTS alert_email_group_members;
DROP TABLE IF EXISTS alert_email_groups;

