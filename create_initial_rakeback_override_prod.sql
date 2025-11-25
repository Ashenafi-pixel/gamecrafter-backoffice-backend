-- Create initial rakeback override record in production
-- This is needed for the toggle function to work
-- Run this on the production database

-- Insert a default inactive record if none exists
INSERT INTO global_rakeback_override (
    is_active,
    rakeback_percentage,
    start_time,
    end_time,
    created_by,
    updated_by
)
SELECT 
    false,  -- is_active: start as inactive
    0.00,   -- rakeback_percentage: default 0%
    NULL,   -- start_time: no start time
    NULL,   -- end_time: no end time
    NULL,   -- created_by: system created
    NULL    -- updated_by: system created
WHERE NOT EXISTS (
    SELECT 1 FROM global_rakeback_override
);

-- Verify the record was created
SELECT 
    id,
    is_active,
    rakeback_percentage,
    start_time,
    end_time,
    created_at,
    updated_at,
    'Initial record created' as status
FROM global_rakeback_override
ORDER BY created_at DESC
LIMIT 1;

