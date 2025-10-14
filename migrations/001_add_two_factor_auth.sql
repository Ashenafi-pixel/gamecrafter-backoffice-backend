-- Migration: Add Two-Factor Authentication Support
-- Created: 2025-10-08
-- Description: Adds 2FA tables and columns for TOTP-based authentication

BEGIN;

-- Create user_2fa_settings table
CREATE TABLE user_2fa_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret_key VARCHAR(255) NOT NULL, -- Base32 encoded secret
    backup_codes TEXT[],              -- Array of backup codes
    is_enabled BOOLEAN DEFAULT FALSE,
    enabled_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Create user_2fa_attempts table for rate limiting and audit
CREATE TABLE user_2fa_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_type VARCHAR(50) NOT NULL, -- 'setup', 'verify', 'disable'
    is_successful BOOLEAN NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add 2FA columns to users table
ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN two_factor_setup_at TIMESTAMP;

-- Create indexes for performance
CREATE INDEX idx_user_2fa_settings_user_id ON user_2fa_settings(user_id);
CREATE INDEX idx_user_2fa_settings_enabled ON user_2fa_settings(is_enabled);
CREATE INDEX idx_user_2fa_attempts_user_id ON user_2fa_attempts(user_id);
CREATE INDEX idx_user_2fa_attempts_created_at ON user_2fa_attempts(created_at);
CREATE INDEX idx_user_2fa_attempts_type ON user_2fa_attempts(attempt_type);
CREATE INDEX idx_users_two_factor_enabled ON users(two_factor_enabled);

-- Add comments for documentation
COMMENT ON TABLE user_2fa_settings IS 'Stores 2FA settings and secrets for users';
COMMENT ON TABLE user_2fa_attempts IS 'Audit log for 2FA attempts and rate limiting';
COMMENT ON COLUMN user_2fa_settings.secret_key IS 'Base32 encoded TOTP secret key';
COMMENT ON COLUMN user_2fa_settings.backup_codes IS 'Array of single-use backup codes';
COMMENT ON COLUMN user_2fa_attempts.attempt_type IS 'Type of attempt: setup, verify, disable';
COMMENT ON COLUMN users.two_factor_enabled IS 'Whether user has 2FA enabled';
COMMENT ON COLUMN users.two_factor_setup_at IS 'When user completed 2FA setup';

COMMIT;
