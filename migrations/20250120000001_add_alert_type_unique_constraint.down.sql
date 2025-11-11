-- Remove unique constraint on alert_type
DROP INDEX IF EXISTS idx_alert_configurations_type_unique_active;

