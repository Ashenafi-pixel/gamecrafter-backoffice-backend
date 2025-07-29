-- Add type column to squads table
ALTER TABLE squads ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'Open';

-- Add index on type for better query performance
CREATE INDEX idx_squads_type ON squads(type); 