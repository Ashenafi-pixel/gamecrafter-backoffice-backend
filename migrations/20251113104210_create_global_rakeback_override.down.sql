-- Migration: Drop global_rakeback_override table

DROP INDEX IF EXISTS idx_global_rakeback_override_active;
DROP TABLE IF EXISTS global_rakeback_override;

