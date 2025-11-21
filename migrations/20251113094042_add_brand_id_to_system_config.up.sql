-- Migration: Add brand_id to system_config table
-- This allows system configs to be brand-specific or global (NULL = global)

-- Add brand_id column to system_config table
ALTER TABLE system_config 
ADD COLUMN brand_id UUID REFERENCES brands(id) ON DELETE CASCADE;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_system_config_brand_id ON system_config(brand_id);

-- Add comment to explain the column
COMMENT ON COLUMN system_config.brand_id IS 'NULL = global config (applies to all brands), UUID = brand-specific config';

