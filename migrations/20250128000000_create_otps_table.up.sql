-- Create OTPs table for email verification
CREATE TABLE IF NOT EXISTS otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(10) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'email_verification',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_otps_email ON otps(email);
CREATE INDEX IF NOT EXISTS idx_otps_type ON otps(type);
CREATE INDEX IF NOT EXISTS idx_otps_status ON otps(status);
CREATE INDEX IF NOT EXISTS idx_otps_expires_at ON otps(expires_at);
CREATE INDEX IF NOT EXISTS idx_otps_created_at ON otps(created_at);

-- Create composite index for recent OTP lookups
CREATE INDEX IF NOT EXISTS idx_otps_email_type_created ON otps(email, type, created_at DESC);

-- Add email verification field to users table if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_email_verified') THEN
        ALTER TABLE users ADD COLUMN is_email_verified BOOLEAN DEFAULT FALSE;
    END IF;
END $$;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_otps_updated_at 
    BEFORE UPDATE ON otps 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment to table
COMMENT ON TABLE otps IS 'OTP table for email verification and password reset functionality';
COMMENT ON COLUMN otps.email IS 'Email address for OTP delivery';
COMMENT ON COLUMN otps.otp_code IS '6-digit OTP code';
COMMENT ON COLUMN otps.type IS 'Type of OTP: email_verification, password_reset, login';
COMMENT ON COLUMN otps.status IS 'Status: pending, verified, expired, used';
COMMENT ON COLUMN otps.expires_at IS 'When the OTP expires';
COMMENT ON COLUMN otps.verified_at IS 'When the OTP was verified'; 