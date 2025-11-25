-- Add global_rakeback_override table (Happy Hour) to Production Database
-- This script safely creates the table if it doesn't exist
-- Run this on the production database

-- Step 1: Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS global_rakeback_override (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_active BOOLEAN DEFAULT false,
    rakeback_percentage DECIMAL(5,2) DEFAULT 0.00,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_percentage CHECK (rakeback_percentage >= 0.00 AND rakeback_percentage <= 100.00)
);

-- Step 2: Create index if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_global_rakeback_override_active 
ON global_rakeback_override(is_active) WHERE is_active = true;

-- Step 3: Add table comments
COMMENT ON TABLE global_rakeback_override IS 
'Global rakeback override for marketing events (Happy Hour). When active, overrides all VIP-level rakeback percentages with a single global value.';

COMMENT ON COLUMN global_rakeback_override.is_active IS 
'When true, rakeback_percentage applies to all users regardless of VIP tier during the time window';

COMMENT ON COLUMN global_rakeback_override.rakeback_percentage IS 
'Global rakeback percentage (0.00-100.00) that overrides VIP tier rates when active';

COMMENT ON COLUMN global_rakeback_override.start_time IS 
'Start timestamp for the override period';

COMMENT ON COLUMN global_rakeback_override.end_time IS 
'End timestamp for the override period';

COMMENT ON COLUMN global_rakeback_override.created_by IS 
'Admin user ID who created the override';

COMMENT ON COLUMN global_rakeback_override.updated_by IS 
'Admin user ID who last updated the override';

-- Step 4: Verify the table was created
SELECT 
    'Table created successfully!' as status,
    COUNT(*) as row_count
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name = 'global_rakeback_override';

-- Step 5: Show table structure
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public' 
AND table_name = 'global_rakeback_override'
ORDER BY ordinal_position;

