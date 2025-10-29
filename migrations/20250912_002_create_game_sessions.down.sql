-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_game_session_activity ON game_sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_game_session_activity();

-- Drop cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_game_sessions();

-- Drop table
DROP TABLE IF EXISTS game_sessions;

-- Drop session ID generation function
DROP FUNCTION IF EXISTS generate_groove_session_id();