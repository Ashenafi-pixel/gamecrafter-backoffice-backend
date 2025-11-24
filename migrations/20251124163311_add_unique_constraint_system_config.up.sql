-- Migration: Add unique constraint on (brand_id, config_key) to system_config table
-- This ensures each brand can only have one config per config_key
-- The backend code uses ON CONFLICT (brand_id, config_key) for upsert operations

-- Drop the old unique constraint on config_key only (if it exists)
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_config_key_key;

-- Drop the partial unique index if it exists (we don't need it anymore)
DROP INDEX IF EXISTS idx_system_config_global_unique;

-- Drop the existing (brand_id, config_key) constraint if it exists (to allow re-running)
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_brand_id_config_key_unique;

-- Add unique constraint on (brand_id, config_key)
-- This ensures each brand has only one config per config_key
ALTER TABLE system_config 
ADD CONSTRAINT system_config_brand_id_config_key_unique 
UNIQUE (brand_id, config_key);

-- Add comment to explain the constraint
COMMENT ON CONSTRAINT system_config_brand_id_config_key_unique ON system_config 
IS 'Ensures each brand can only have one config per config_key. Used for ON CONFLICT upsert operations.';

