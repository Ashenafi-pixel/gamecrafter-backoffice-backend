-- Seed default game import configurations for all active brands
-- This migration creates a default configuration for each brand that doesn't already have one

INSERT INTO game_import_config (
    brand_id,
    schedule_type,
    schedule_cron,
    providers,
    directus_url,
    check_frequency_minutes,
    is_active,
    created_at,
    updated_at
)
SELECT 
    b.id as brand_id,
    'daily' as schedule_type,                    -- Default: daily schedule
    NULL as schedule_cron,                        -- NULL for non-custom schedules
    NULL as providers,                            -- NULL = import all providers
    NULL as directus_url,                        -- NULL = use default from config
    15 as check_frequency_minutes,                -- Default: check every 15 minutes
    true as is_active,                            -- Default: enabled
    NOW() as created_at,
    NOW() as updated_at
FROM brands b
WHERE b.is_active = true                          -- Only for active brands
  AND NOT EXISTS (
      SELECT 1 
      FROM game_import_config gic 
      WHERE gic.brand_id = b.id
  )
ON CONFLICT (brand_id) DO NOTHING;                -- Skip if config already exists

COMMENT ON TABLE game_import_config IS 'Configuration for automated game import from Directus. Each brand has one configuration record.';

