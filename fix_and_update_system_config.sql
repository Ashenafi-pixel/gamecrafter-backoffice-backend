-- Script to fix system_config constraint and update all brands
-- Run this on the production database

-- Step 1: Check if the constraint exists
SELECT 
    conname as constraint_name,
    contype as constraint_type,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint
WHERE conrelid = 'system_config'::regclass
AND conname = 'system_config_brand_id_config_key_unique';

-- Step 2: List all brands
SELECT id, name, created_at 
FROM brands 
ORDER BY name;

-- Step 3: Check current system_config entries
SELECT 
    sc.id,
    sc.config_key,
    sc.brand_id,
    b.name as brand_name,
    sc.updated_at
FROM system_config sc
LEFT JOIN brands b ON sc.brand_id = b.id
ORDER BY sc.config_key, b.name;

-- Step 3a: Check for NULL brand_id values (these need to be handled)
SELECT config_key, COUNT(*) as count 
FROM system_config 
WHERE brand_id IS NULL 
GROUP BY config_key;

-- Step 3b: If there are NULL brand_id values, delete them or migrate them
-- Option: Delete NULL brand_id entries (uncomment if needed)
-- DELETE FROM system_config WHERE brand_id IS NULL;

-- Step 4: Drop old constraints/indexes if they exist
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_config_key_key;

DROP INDEX IF EXISTS idx_system_config_global_unique;

ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_brand_id_config_key_unique;

-- Step 5: Ensure brand_id is NOT NULL (if it's not already)
-- This is required for the unique constraint to work properly
-- Comment this out if brand_id already has NOT NULL constraint
-- ALTER TABLE system_config ALTER COLUMN brand_id SET NOT NULL;

-- Step 6: Create the unique constraint
ALTER TABLE system_config 
ADD CONSTRAINT system_config_brand_id_config_key_unique 
UNIQUE (brand_id, config_key);

-- Step 7: Verify the constraint was created
SELECT 
    conname as constraint_name,
    contype as constraint_type,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint
WHERE conrelid = 'system_config'::regclass
AND conname = 'system_config_brand_id_config_key_unique';

-- Step 8: Initialize/Update general_settings for all brands
-- Replace '5a8328c7-d51b-4187-b45c-b1beea7b41ff' with your admin user ID
WITH all_brands AS (
    SELECT id FROM brands
),
default_general_settings AS (
    SELECT 
        'general_settings'::varchar(100) as config_key,
        '{"site_name": "Global Site", "site_description": "Global Description", "support_email": "support@cryptocasino.com", "timezone": "UTC", "language": "en", "maintenance_mode": false, "registration_enabled": true, "demo_mode": false}'::jsonb as config_value,
        'General application settings'::text as description,
        '5a8328c7-d51b-4187-b45c-b1beea7b41ff'::uuid as updated_by
)
INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
SELECT 
    dgs.config_key,
    dgs.config_value,
    dgs.description,
    ab.id as brand_id,
    dgs.updated_by,
    NOW()
FROM all_brands ab
CROSS JOIN default_general_settings dgs
ON CONFLICT (brand_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

-- Step 9: Verify the updates
SELECT 
    sc.config_key,
    b.name as brand_name,
    sc.config_value,
    sc.updated_at
FROM system_config sc
LEFT JOIN brands b ON sc.brand_id = b.id
WHERE sc.config_key = 'general_settings'
ORDER BY b.name;

