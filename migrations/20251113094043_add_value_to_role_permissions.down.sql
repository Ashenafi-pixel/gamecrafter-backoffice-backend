-- Migration: Remove value column from role_permissions table

-- Drop index
DROP INDEX IF EXISTS idx_role_permissions_value;

-- Remove value column
ALTER TABLE role_permissions DROP COLUMN IF EXISTS value;
