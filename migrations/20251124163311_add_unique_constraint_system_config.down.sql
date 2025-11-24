-- Rollback: Remove unique constraint on (brand_id, config_key) from system_config table

-- Drop the unique constraint
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_brand_id_config_key_unique;

-- Restore the old unique constraint on config_key only (if needed)
-- Note: This is commented out as it may not be desired after migration
-- ALTER TABLE system_config 
-- ADD CONSTRAINT system_config_config_key_key UNIQUE (config_key);

