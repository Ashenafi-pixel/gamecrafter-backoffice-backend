-- Add unique constraint on alert_type to ensure only one alert configuration per type
-- This migration adds a unique constraint to prevent duplicate alert types

-- First, remove any duplicate alert configurations (keep the most recent one)
DELETE FROM alert_configurations a
USING alert_configurations b
WHERE a.id < b.id
  AND a.alert_type = b.alert_type
  AND a.status = 'active';

-- Add unique constraint on alert_type for active alerts
-- Note: We allow multiple inactive alerts but only one active alert per type
CREATE UNIQUE INDEX idx_alert_configurations_type_unique_active 
ON alert_configurations(alert_type) 
WHERE status = 'active';

-- Add a comment explaining the constraint
COMMENT ON INDEX idx_alert_configurations_type_unique_active IS 
'Ensures only one active alert configuration exists per alert_type';

