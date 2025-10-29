-- Migration: Create enterprise_registrations table
-- This table stores enterprise registration data with proper indexing and constraints

CREATE TABLE IF NOT EXISTS enterprise_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    user_type VARCHAR(50) NOT NULL CHECK (user_type IN ('PLAYER', 'AGENT', 'ADMIN')),
    phone_number VARCHAR(20),
    company_name VARCHAR(255),
    registration_status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (registration_status IN ('PENDING', 'VERIFIED', 'COMPLETED', 'REJECTED')),
    verification_otp VARCHAR(10),
    otp_expires_at TIMESTAMP WITH TIME ZONE,
    verification_attempts INTEGER DEFAULT 0,
    max_verification_attempts INTEGER DEFAULT 3,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    phone_verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT unique_email_registration UNIQUE (email),
    CONSTRAINT unique_user_registration UNIQUE (user_id),
    CONSTRAINT valid_phone_number CHECK (phone_number ~ '^\+[1-9]\d{1,14}$')
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_enterprise_registrations_email ON enterprise_registrations(email);
CREATE INDEX IF NOT EXISTS idx_enterprise_registrations_user_id ON enterprise_registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_enterprise_registrations_status ON enterprise_registrations(registration_status);
CREATE INDEX IF NOT EXISTS idx_enterprise_registrations_created_at ON enterprise_registrations(created_at);
CREATE INDEX IF NOT EXISTS idx_enterprise_registrations_user_type ON enterprise_registrations(user_type);

-- Create a trigger to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_enterprise_registrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_enterprise_registrations_updated_at
    BEFORE UPDATE ON enterprise_registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_enterprise_registrations_updated_at();

-- Create a view for enterprise registration statistics
CREATE OR REPLACE VIEW enterprise_registration_stats AS
SELECT 
    registration_status,
    user_type,
    COUNT(*) as count,
    DATE_TRUNC('day', created_at) as date
FROM enterprise_registrations
GROUP BY registration_status, user_type, DATE_TRUNC('day', created_at);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON enterprise_registrations TO tucanbit;
GRANT SELECT ON enterprise_registration_stats TO tucanbit; 