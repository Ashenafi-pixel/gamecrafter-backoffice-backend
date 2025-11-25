-- Script to check and add global_rakeback_override table (Happy Hour) to production database
-- This script checks if the table exists and creates it if missing
-- Run this on the production database

-- Step 1: Check if table exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'global_rakeback_override'
    ) THEN
        -- Table doesn't exist, create it
        RAISE NOTICE 'Table global_rakeback_override does not exist. Creating it...';
        
        -- Create global_rakeback_override table
        CREATE TABLE global_rakeback_override (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            is_active BOOLEAN DEFAULT false,
            rakeback_percentage DECIMAL(5,2) DEFAULT 0.00, -- Up to 2 decimal places (e.g., 100.00, 12.5)
            start_time TIMESTAMP WITH TIME ZONE,
            end_time TIMESTAMP WITH TIME ZONE,
            created_by UUID REFERENCES users(id),
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_by UUID REFERENCES users(id),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            CONSTRAINT valid_percentage CHECK (rakeback_percentage >= 0.00 AND rakeback_percentage <= 100.00)
        );

        -- Create index for active override queries
        CREATE INDEX IF NOT EXISTS idx_global_rakeback_override_active 
        ON global_rakeback_override(is_active) WHERE is_active = true;

        -- Add comment to explain the table
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

        RAISE NOTICE 'Table global_rakeback_override created successfully!';
    ELSE
        RAISE NOTICE 'Table global_rakeback_override already exists. No action needed.';
        
        -- Check if table has the correct structure
        -- Verify required columns exist
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'global_rakeback_override' 
            AND column_name = 'is_active'
        ) THEN
            RAISE NOTICE 'WARNING: Column is_active is missing. You may need to add it manually.';
        END IF;
        
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'global_rakeback_override' 
            AND column_name = 'rakeback_percentage'
        ) THEN
            RAISE NOTICE 'WARNING: Column rakeback_percentage is missing. You may need to add it manually.';
        END IF;
    END IF;
END $$;

-- Step 2: Check current status in dev database (for reference)
-- This query shows if happy hour is currently enabled
-- Run this separately on dev database to check status
/*
SELECT 
    id,
    is_active,
    rakeback_percentage,
    start_time,
    end_time,
    created_by,
    created_at,
    updated_by,
    updated_at,
    CASE 
        WHEN is_active = true AND 
             (start_time IS NULL OR start_time <= NOW()) AND 
             (end_time IS NULL OR end_time > NOW()) 
        THEN 'ACTIVE'
        ELSE 'INACTIVE'
    END as status
FROM global_rakeback_override
ORDER BY created_at DESC
LIMIT 1;
*/

-- Step 3: Verify table structure
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public' 
AND table_name = 'global_rakeback_override'
ORDER BY ordinal_position;

