-- Add requires_value column to permissions table
ALTER TABLE permissions 
ADD COLUMN IF NOT EXISTS requires_value BOOLEAN DEFAULT FALSE;

-- Update existing permissions to have requires_value = false if null
UPDATE permissions 
SET requires_value = FALSE 
WHERE requires_value IS NULL;

