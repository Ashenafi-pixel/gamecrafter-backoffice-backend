-- Migration: Create user_2fa_methods table
-- Created: 2025-09-16
-- Description: Adds table to store enabled 2FA methods per user

BEGIN;

-- Create user_2fa_methods table
CREATE TABLE user_2fa_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL, -- 'totp', 'email_otp', 'sms_otp', 'biometric', 'backup_codes'
    enabled_at TIMESTAMP DEFAULT NOW(),
    disabled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, method)
);

-- Create indexes for performance
CREATE INDEX idx_user_2fa_methods_user_id ON user_2fa_methods(user_id);
CREATE INDEX idx_user_2fa_methods_method ON user_2fa_methods(method);
CREATE INDEX idx_user_2fa_methods_enabled_at ON user_2fa_methods(enabled_at);

-- Add comments for documentation
COMMENT ON TABLE user_2fa_methods IS 'Stores enabled 2FA methods for each user';
COMMENT ON COLUMN user_2fa_methods.method IS '2FA method type: totp, email_otp, sms_otp, biometric, backup_codes';
COMMENT ON COLUMN user_2fa_methods.enabled_at IS 'When the method was enabled (NULL if disabled)';
COMMENT ON COLUMN user_2fa_methods.disabled_at IS 'When the method was disabled (NULL if enabled)';

COMMIT;
