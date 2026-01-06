-- Add directus_url, check_frequency_minutes, and last_check_at columns to game_import_config
ALTER TABLE game_import_config 
ADD COLUMN IF NOT EXISTS directus_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS check_frequency_minutes INTEGER DEFAULT 15,
ADD COLUMN IF NOT EXISTS last_check_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN game_import_config.directus_url IS 'Per-brand Directus API URL (optional, falls back to default)';
COMMENT ON COLUMN game_import_config.check_frequency_minutes IS 'How often to check if import is due (in minutes, default: 15)';
COMMENT ON COLUMN game_import_config.last_check_at IS 'Timestamp of last check for this brand (separate from last_run_at)';


