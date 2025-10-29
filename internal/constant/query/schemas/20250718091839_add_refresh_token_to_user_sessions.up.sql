ALTER TABLE user_sessions
ADD COLUMN refresh_token TEXT,
ADD COLUMN refresh_token_expires_at TIMESTAMP;