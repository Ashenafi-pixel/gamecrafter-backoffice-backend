-- Rollback Rakeback Schedules Migration

-- Drop indexes
DROP INDEX IF EXISTS idx_rakeback_schedules_scheduler;
DROP INDEX IF EXISTS idx_rakeback_schedules_created_by;
DROP INDEX IF EXISTS idx_rakeback_schedules_scope_type;
DROP INDEX IF EXISTS idx_rakeback_schedules_end_time;
DROP INDEX IF EXISTS idx_rakeback_schedules_start_time;
DROP INDEX IF EXISTS idx_rakeback_schedules_status;

-- Drop the rakeback_schedules table
DROP TABLE IF EXISTS rakeback_schedules CASCADE;

