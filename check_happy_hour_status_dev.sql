-- Script to check Happy Hour status in DEV database
-- Run this on the dev database to see current happy hour configuration

-- Check if table exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'global_rakeback_override'
        ) THEN 'Table exists'
        ELSE 'Table does NOT exist'
    END as table_status;

-- If table exists, show current configuration
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
        THEN 'ACTIVE (Happy Hour Enabled)'
        WHEN is_active = true AND start_time > NOW()
        THEN 'SCHEDULED (Will activate in future)'
        WHEN is_active = true AND end_time <= NOW()
        THEN 'EXPIRED (End time passed)'
        ELSE 'INACTIVE (Happy Hour Disabled)'
    END as current_status,
    CASE 
        WHEN start_time IS NOT NULL AND end_time IS NOT NULL
        THEN end_time - start_time
        ELSE NULL
    END as duration
FROM global_rakeback_override
ORDER BY created_at DESC
LIMIT 5;

-- Show all active overrides
SELECT 
    id,
    is_active,
    rakeback_percentage,
    start_time,
    end_time,
    'Currently Active' as status
FROM global_rakeback_override
WHERE is_active = true
AND (start_time IS NULL OR start_time <= NOW())
AND (end_time IS NULL OR end_time > NOW())
ORDER BY created_at DESC;

