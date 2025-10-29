-- Drop OTPs table and related objects
DROP TRIGGER IF EXISTS update_otps_updated_at ON otps;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Remove email verification field from users table if it exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_email_verified') THEN
        ALTER TABLE users DROP COLUMN is_email_verified;
    END IF;
END $$;

-- Drop OTPs table
DROP TABLE IF EXISTS otps CASCADE; 