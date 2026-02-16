-- Remove limit_type and limit_period columns from role_permissions table
ALTER TABLE role_permissions 
DROP CONSTRAINT IF EXISTS role_permissions_limit_type_check,
DROP COLUMN IF EXISTS limit_type,
DROP COLUMN IF EXISTS limit_period;

