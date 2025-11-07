-- Rakeback Schedules Migration
-- This migration adds scheduled rakeback functionality with automatic activation/deactivation

-- Create rakeback_schedules table
CREATE TABLE IF NOT EXISTS rakeback_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    percentage DECIMAL(5,2) NOT NULL,
    scope_type VARCHAR(50) NOT NULL DEFAULT 'all',
    scope_value TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    activated_at TIMESTAMP WITH TIME ZONE,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_percentage CHECK (percentage >= 0.00 AND percentage <= 100.00),
    CONSTRAINT valid_scope_type CHECK (scope_type IN ('all', 'provider', 'game')),
    CONSTRAINT valid_status CHECK (status IN ('scheduled', 'active', 'completed', 'cancelled')),
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

-- Create indexes for performance
CREATE INDEX idx_rakeback_schedules_status ON rakeback_schedules(status);
CREATE INDEX idx_rakeback_schedules_start_time ON rakeback_schedules(start_time);
CREATE INDEX idx_rakeback_schedules_end_time ON rakeback_schedules(end_time);
CREATE INDEX idx_rakeback_schedules_scope_type ON rakeback_schedules(scope_type);
CREATE INDEX idx_rakeback_schedules_created_by ON rakeback_schedules(created_by);

-- Add composite index for scheduler queries
CREATE INDEX idx_rakeback_schedules_scheduler ON rakeback_schedules(status, start_time, end_time);

-- Add comments for documentation
COMMENT ON TABLE rakeback_schedules IS 
'Scheduled rakeback events with automatic activation/deactivation based on time windows';

COMMENT ON COLUMN rakeback_schedules.name IS 
'Display name for the schedule (e.g., "Weekend Boost", "Happy Hour")';

COMMENT ON COLUMN rakeback_schedules.start_time IS 
'When the rakeback override should automatically activate';

COMMENT ON COLUMN rakeback_schedules.end_time IS 
'When the rakeback override should automatically deactivate';

COMMENT ON COLUMN rakeback_schedules.percentage IS 
'Rakeback percentage (0-100%) to apply during the scheduled window';

COMMENT ON COLUMN rakeback_schedules.scope_type IS 
'Scope of the rakeback: all (all games), provider (specific provider), game (specific game)';

COMMENT ON COLUMN rakeback_schedules.scope_value IS 
'Provider name or game ID when scope_type is provider or game';

COMMENT ON COLUMN rakeback_schedules.status IS 
'Current status: scheduled (pending), active (running now), completed (finished), cancelled (manually cancelled)';

