-- Migration: Create user_2fa_otps table for storing OTP codes
-- This table stores temporary OTP codes for email and SMS 2FA methods

CREATE TABLE IF NOT EXISTS user_2fa_otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    method VARCHAR(50) NOT NULL, -- 'email_otp' or 'sms_otp'
    otp_code VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure one OTP per user per method at a time
    UNIQUE(user_id, method)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_user_2fa_otps_user_method ON user_2fa_otps(user_id, method);
CREATE INDEX IF NOT EXISTS idx_user_2fa_otps_expires_at ON user_2fa_otps(expires_at);

-- Add foreign key constraint if user_2fa_settings table exists
-- ALTER TABLE user_2fa_otps ADD CONSTRAINT fk_user_2fa_otps_user_id 
--     FOREIGN KEY (user_id) REFERENCES user_2fa_settings(user_id) ON DELETE CASCADE;

-- Add comment
COMMENT ON TABLE user_2fa_otps IS 'Stores temporary OTP codes for 2FA email and SMS methods';
COMMENT ON COLUMN user_2fa_otps.method IS '2FA method: email_otp or sms_otp';
COMMENT ON COLUMN user_2fa_otps.otp_code IS 'The actual OTP code sent to user';
COMMENT ON COLUMN user_2fa_otps.expires_at IS 'When the OTP expires (typically 10 minutes)';
