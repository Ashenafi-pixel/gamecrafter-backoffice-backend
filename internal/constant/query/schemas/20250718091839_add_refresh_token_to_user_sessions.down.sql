ALTER TABLE user_sessions
DROP COLUMN IF EXISTS refresh_token,
DROP COLUMN IF EXISTS refresh_token_expires_at;