-- Check if there are any levels configured in the database
-- Run these queries to verify the level system setup

-- 1. Check if levels table exists and has any data
SELECT 
    'levels' as table_name,
    COUNT(*) as record_count
FROM levels 
WHERE deleted_at IS NULL;

-- 2. Check if level_requirements table exists and has any data
SELECT 
    'level_requirements' as table_name,
    COUNT(*) as record_count
FROM level_requirements 
WHERE deleted_at IS NULL;

-- 3. Get all levels with their details
SELECT 
    id,
    level,
    created_at,
    created_by
FROM levels 
WHERE deleted_at IS NULL
ORDER BY level;

-- 4. Get all level requirements with their details
SELECT 
    lr.id,
    lr.level_id,
    l.level as level_number,
    lr.type,
    lr.value,
    lr.created_at
FROM level_requirements lr
JOIN levels l ON lr.level_id = l.id
WHERE lr.deleted_at IS NULL AND l.deleted_at IS NULL
ORDER BY l.level, lr.type;

-- 5. Check if operational_types table has 'place_bet' entry
SELECT 
    'operational_types' as table_name,
    COUNT(*) as record_count
FROM operational_types 
WHERE name = 'place_bet';

-- 6. Check if there are any balance_logs for a specific user (replace USER_ID with actual user ID)
-- SELECT 
--     'balance_logs' as table_name,
--     COUNT(*) as record_count,
--     SUM(change_amount) as total_bet_amount
-- FROM balance_logs 
-- WHERE user_id = 'USER_ID_HERE' 
-- AND operational_type_id = (SELECT id FROM operational_types WHERE name = 'place_bet' LIMIT 1)
-- AND currency = 'P';

-- 7. Check the structure of levels table
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'levels'
ORDER BY ordinal_position;

-- 8. Check the structure of level_requirements table
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'level_requirements'
ORDER BY ordinal_position; 