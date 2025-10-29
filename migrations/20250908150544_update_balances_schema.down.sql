-- Rollback changes to balances table
-- Remove the new columns added in the up migration

ALTER TABLE balances 
DROP COLUMN IF EXISTS amount_cents,
DROP COLUMN IF EXISTS amount_units,
DROP COLUMN IF EXISTS reserved_cents,
DROP COLUMN IF EXISTS reserved_units;