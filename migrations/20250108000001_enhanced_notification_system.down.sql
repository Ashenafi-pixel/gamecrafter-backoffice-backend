-- Enhanced Notification System Migration Rollback
-- This migration removes the enhanced notification system tables and types

-- 1. Drop triggers
DROP TRIGGER IF EXISTS update_message_campaigns_updated_at ON message_campaigns;

-- 2. Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 3. Drop indexes
DROP INDEX IF EXISTS idx_user_activity_log_created_at;
DROP INDEX IF EXISTS idx_user_activity_log_activity_type;
DROP INDEX IF EXISTS idx_user_activity_log_user_id;
DROP INDEX IF EXISTS idx_campaign_recipients_notification_id;
DROP INDEX IF EXISTS idx_campaign_recipients_status;
DROP INDEX IF EXISTS idx_campaign_recipients_user_id;
DROP INDEX IF EXISTS idx_campaign_recipients_campaign_id;
DROP INDEX IF EXISTS idx_message_segments_segment_type;
DROP INDEX IF EXISTS idx_message_segments_campaign_id;
DROP INDEX IF EXISTS idx_message_campaigns_message_type;
DROP INDEX IF EXISTS idx_message_campaigns_scheduled_at;
DROP INDEX IF EXISTS idx_message_campaigns_status;
DROP INDEX IF EXISTS idx_message_campaigns_created_by;

-- 4. Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS user_activity_log;
DROP TABLE IF EXISTS campaign_recipients;
DROP TABLE IF EXISTS message_segments;
DROP TABLE IF EXISTS message_campaigns;

-- 5. Drop notification_type enum (only if no other tables reference it)
-- Note: This is commented out as it might be referenced by other tables
-- DROP TYPE IF EXISTS notification_type;

-- Note: We don't remove the created_by column from user_notifications as it might be used elsewhere
-- Note: We don't remove the notification_type enum as it might be referenced by other tables
