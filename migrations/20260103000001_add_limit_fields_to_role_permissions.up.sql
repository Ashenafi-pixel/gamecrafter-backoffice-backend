-- Add limit_type and limit_period columns to role_permissions table
ALTER TABLE role_permissions 
ADD COLUMN IF NOT EXISTS limit_type VARCHAR(20),
ADD COLUMN IF NOT EXISTS limit_period INTEGER;

-- Add check constraint for limit_type if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'role_permissions_limit_type_check'
    ) THEN
        ALTER TABLE role_permissions 
        ADD CONSTRAINT role_permissions_limit_type_check 
        CHECK (limit_type IS NULL OR limit_type IN ('daily', 'weekly', 'monthly'));
    END IF;
END $$;

