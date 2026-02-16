-- Remove requires_value column from permissions table
ALTER TABLE permissions 
DROP COLUMN IF EXISTS requires_value;

