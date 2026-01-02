-- Remove directus_url, check_frequency_minutes, and last_check_at columns from game_import_config
ALTER TABLE game_import_config 
DROP COLUMN IF EXISTS directus_url,
DROP COLUMN IF EXISTS check_frequency_minutes,
DROP COLUMN IF EXISTS last_check_at;


