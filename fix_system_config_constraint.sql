-- SQL Queries to fix system_config unique constraint and update logic
-- IMPORTANT: All settings must exist for ALL brands (no global configs with brand_id IS NULL)
-- When updating, always use brand_id. When "global" is selected, update ALL brands.

-- Step 0: Check for existing global configs (brand_id IS NULL) that need to be migrated
SELECT config_key, COUNT(*) as count 
FROM system_config 
WHERE brand_id IS NULL 
GROUP BY config_key;

-- Step 0b: Migrate global configs to all brands
-- This will copy each global config to all existing brands
-- Run this AFTER creating the unique constraint
/*
WITH global_configs AS (
    SELECT config_key, config_value, description
    FROM system_config
    WHERE brand_id IS NULL
),
all_brands AS (
    SELECT id FROM brands
)
INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
SELECT 
    gc.config_key,
    gc.config_value,
    gc.description,
    ab.id as brand_id,
    (SELECT id FROM users WHERE is_admin = true LIMIT 1) as updated_by,
    NOW()
FROM global_configs gc
CROSS JOIN all_brands ab
ON CONFLICT (brand_id, config_key) DO NOTHING;

-- After migration, delete global configs
DELETE FROM system_config WHERE brand_id IS NULL;
*/

-- 1. Drop the existing unique constraint on config_key only (if it exists)
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_config_key_key;

-- 2. Drop the partial unique index if it exists (we don't need it anymore)
DROP INDEX IF EXISTS idx_system_config_global_unique;

-- 2a. Check for NULL brand_id values (these need to be migrated to all brands or deleted)
SELECT config_key, COUNT(*) as count 
FROM system_config 
WHERE brand_id IS NULL 
GROUP BY config_key;

-- 2b. IMPORTANT: Handle NULL brand_id values BEFORE creating the constraint
--     If you have NULL brand_id values, you MUST either:
--     Option A: Delete them (if they're not needed)
--     DELETE FROM system_config WHERE brand_id IS NULL;
--     Option B: Migrate them to all brands (uncomment and run Step 0b above first)
--     You cannot create the constraint if there are NULL values and brand_id is set to NOT NULL

-- 3. Drop the existing (brand_id, config_key) constraint if it exists (to avoid errors on re-run)
ALTER TABLE system_config 
DROP CONSTRAINT IF EXISTS system_config_brand_id_config_key_unique;

-- 3a. Ensure brand_id is NOT NULL (only run this AFTER handling NULL values in step 2b)
--     This will fail if there are any NULL brand_id values remaining
--     Comment this out if brand_id column already allows NULLs and you want to keep it that way
-- ALTER TABLE system_config 
-- ALTER COLUMN brand_id SET NOT NULL;

-- 4. Add unique constraint on (brand_id, config_key)
-- This ensures each brand has only one config per config_key
-- brand_id must NOT be NULL (all settings are brand-specific)
-- Note: The constraint ensures (brand_id, config_key) is unique
ALTER TABLE system_config 
ADD CONSTRAINT system_config_brand_id_config_key_unique 
UNIQUE (brand_id, config_key);

-- 4a. Verify the constraint was created successfully:
SELECT 
    conname as constraint_name,
    contype as constraint_type,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint
WHERE conrelid = 'system_config'::regclass
AND conname = 'system_config_brand_id_config_key_unique';

-- ============================================================================
-- IMPORTANT: The constraint must be created successfully before running any
-- queries that use ON CONFLICT (brand_id, config_key)
-- ============================================================================
-- Verify the constraint exists by running the query in step 4a above
-- Only proceed with the example queries below AFTER confirming the constraint exists
-- ============================================================================

-- ============================================================================
-- EXAMPLE QUERIES (DO NOT RUN UNTIL CONSTRAINT IS VERIFIED)
-- ============================================================================

-- 5. Example SQL for updating settings with brand_id (for specific brand):
-- This uses ON CONFLICT to update if exists, insert if not
-- Replace 'general_settings' with the actual config_key
-- Replace the JSON value with actual settings
-- Replace brand_id and updated_by with actual UUIDs

-- For specific brand (brand_id is NOT NULL):
-- NOTE: This query requires the constraint from step 4 to exist
/*
INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
VALUES (
    'general_settings',
    '{"site_name": "Example Site", "site_description": "Description"}',
    'General application settings',
    '00000000-0000-0000-0000-000000000004'::uuid,  -- Specific brand ID
    '5a8328c7-d51b-4187-b45c-b1beea7b41ff'::uuid,
    NOW()
)
ON CONFLICT (brand_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();
*/

-- 6. SQL to update ALL brands with the same settings (when "global" is selected):
-- This is the MAIN query to use when updating settings for all brands
-- It updates/creates the config for EVERY brand in the database
-- NOTE: This query requires the constraint from step 4 to exist
/*
WITH brand_ids AS (
    SELECT id FROM brands
),
settings_to_insert AS (
    SELECT 
        'general_settings'::varchar(100) as config_key,
        '{"site_name": "Global Site", "site_description": "Global Description"}'::jsonb as config_value,
        'General application settings'::text as description,
        '5a8328c7-d51b-4187-b45c-b1beea7b41ff'::uuid as updated_by
)
INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
SELECT 
    sti.config_key,
    sti.config_value,
    sti.description,
    bi.id as brand_id,  -- Each brand gets its own config entry
    sti.updated_by,
    NOW()
FROM brand_ids bi
CROSS JOIN settings_to_insert sti
ON CONFLICT (brand_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();
*/

-- 7. Example SQL for updating settings globally (when no brand_id is selected):
-- This updates ALL brands with the same settings
-- Use the query from section 6 above (update ALL brands)

-- 8. Initialize all settings for all brands (run this once to create all configs for all brands):
-- This ensures every brand has every config type, even if with default values
-- NOTE: This query requires the constraint from step 4 to exist
/*
WITH all_brands AS (
    SELECT id FROM brands
),
all_config_keys AS (
    SELECT DISTINCT config_key FROM system_config
    UNION
    SELECT 'general_settings'::varchar(100)
    UNION
    SELECT 'payment_settings'::varchar(100)
    UNION
    SELECT 'tip_settings'::varchar(100)
    UNION
    SELECT 'security_settings'::varchar(100)
    UNION
    SELECT 'geo_blocking_settings'::varchar(100)
),
default_configs AS (
    SELECT 
        'general_settings'::varchar(100) as config_key,
        '{"site_name": "Default Site", "site_description": "Default Description", "support_email": "support@example.com", "timezone": "UTC", "language": "en", "maintenance_mode": false, "registration_enabled": true, "demo_mode": false}'::jsonb as config_value,
        'General application settings'::text as description
    UNION ALL
    SELECT 
        'payment_settings'::varchar(100),
        '{"min_deposit_btc": 0.001, "max_deposit_btc": 10.0, "min_withdrawal_btc": 0.001, "max_withdrawal_btc": 10.0, "kyc_required": false, "kyc_threshold": 1000}'::jsonb,
        'Payment processing settings'::text
    UNION ALL
    SELECT 
        'tip_settings'::varchar(100),
        '{"tip_transaction_fee_from_who": "sender", "transaction_fee": 0}'::jsonb,
        'Tip transaction fee settings'::text
    UNION ALL
    SELECT 
        'security_settings'::varchar(100),
        '{"session_timeout": 1800, "max_login_attempts": 5, "lockout_duration": 300, "two_factor_required": false, "password_min_length": 8, "password_require_special": true, "ip_whitelist_enabled": false, "rate_limit_enabled": true, "rate_limit_requests": 100}'::jsonb,
        'Security and authentication settings'::text
    UNION ALL
    SELECT 
        'geo_blocking_settings'::varchar(100),
        '{"enable_geo_blocking": false, "default_action": "allow", "vpn_detection": false, "proxy_detection": false, "tor_blocking": false, "log_attempts": true, "blocked_countries": [], "allowed_countries": []}'::jsonb,
        'Geo blocking and location-based access control settings'::text
)
INSERT INTO system_config (config_key, config_value, description, brand_id, updated_by, updated_at)
SELECT 
    dc.config_key,
    dc.config_value,
    dc.description,
    ab.id as brand_id,
    '5a8328c7-d51b-4187-b45c-b1beea7b41ff'::uuid as updated_by,
    NOW()
FROM all_brands ab
CROSS JOIN default_configs dc
ON CONFLICT (brand_id, config_key) DO NOTHING;  -- Don't overwrite existing configs
*/

-- 9. Query to get config for a SPECIFIC brand (always use brand_id):
SELECT * FROM system_config 
WHERE config_key = 'general_settings' 
AND brand_id = '00000000-0000-0000-0000-000000000004'::uuid;

-- 10. Query to get all configs for a specific brand:
SELECT * FROM system_config 
WHERE brand_id = '00000000-0000-0000-0000-000000000004'::uuid
ORDER BY config_key;

-- 11. Query to check if a brand has all required configs:
SELECT 
    b.id as brand_id,
    b.name as brand_name,
    COUNT(DISTINCT sc.config_key) as config_count,
    CASE 
        WHEN COUNT(DISTINCT sc.config_key) >= 5 THEN 'Complete'
        ELSE 'Incomplete'
    END as status
FROM brands b
LEFT JOIN system_config sc ON sc.brand_id = b.id
GROUP BY b.id, b.name
ORDER BY b.name;

-- 12. Query to list all brands and their config status:
SELECT 
    b.id,
    b.name,
    sc.config_key,
    CASE WHEN sc.id IS NOT NULL THEN 'Exists' ELSE 'Missing' END as status
FROM brands b
CROSS JOIN (
    SELECT DISTINCT config_key 
    FROM system_config
    WHERE config_key IN ('general_settings', 'payment_settings', 'tip_settings', 'security_settings', 'geo_blocking_settings')
) required_configs
LEFT JOIN system_config sc ON sc.brand_id = b.id AND sc.config_key = required_configs.config_key
ORDER BY b.name, required_configs.config_key;

