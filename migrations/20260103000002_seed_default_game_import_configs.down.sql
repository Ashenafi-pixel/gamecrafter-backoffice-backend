-- Remove default game import configurations (optional - you may want to keep them)
-- This will only remove configs that match the default values
-- Uncomment if you want to remove seeded configs on rollback

-- DELETE FROM game_import_config 
-- WHERE schedule_type = 'daily' 
--   AND schedule_cron IS NULL 
--   AND providers IS NULL 
--   AND directus_url IS NULL 
--   AND check_frequency_minutes = 15 
--   AND is_active = true;

-- For safety, we'll just leave a comment - actual deletion should be manual if needed

