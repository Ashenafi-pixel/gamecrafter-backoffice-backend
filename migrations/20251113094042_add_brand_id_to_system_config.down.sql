-- Migration: Remove brand_id from system_config table

-- Drop index
DROP INDEX IF EXISTS idx_system_config_brand_id;

-- Remove brand_id column
ALTER TABLE system_config DROP COLUMN IF EXISTS brand_id;
