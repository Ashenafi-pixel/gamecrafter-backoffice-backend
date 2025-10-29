-- Drop index on type
DROP INDEX IF EXISTS idx_squads_type;

-- Remove type column from squads table
ALTER TABLE squads DROP COLUMN IF EXISTS type;

-- Drop squad_type enum
DROP TYPE IF EXISTS squad_type; 