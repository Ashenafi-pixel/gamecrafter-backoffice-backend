-- Migration: Add value column to role_permissions table
-- This allows different roles to have different values for the same permission
-- Example: "manual funding" permission can have $100 for Role A, $500 for Role B, NULL (unlimited) for Role C

-- Add value column to role_permissions table
ALTER TABLE role_permissions 
ADD COLUMN value DECIMAL(20,8);

-- Add index for better query performance
CREATE INDEX IF NOT EXISTS idx_role_permissions_value ON role_permissions(value) WHERE value IS NOT NULL;

-- Add comment to explain the column
COMMENT ON COLUMN role_permissions.value IS 'Permission-specific value (e.g., funding limit). NULL = unlimited or not applicable';

