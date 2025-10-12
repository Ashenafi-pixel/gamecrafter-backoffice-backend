-- Fix campaign recipient counts to reflect actual player count (excluding admins)
-- This script updates existing campaigns that have hardcoded recipient counts

-- Get the actual player count (excluding admins)
DO $$
DECLARE
    actual_player_count INTEGER;
BEGIN
    -- Get the actual count of players in the database (excluding admins)
    SELECT COUNT(*) INTO actual_player_count 
    FROM users 
    WHERE is_admin IS NOT TRUE AND user_type = 'PLAYER';
    
    -- Update all campaigns that have segments with 'all_users' type
    -- and set their total_recipients to the actual player count
    UPDATE message_campaigns 
    SET total_recipients = actual_player_count
    WHERE id IN (
        SELECT DISTINCT ms.campaign_id 
        FROM message_segments ms 
        WHERE ms.segment_type = 'all_users'
    );
    
    -- Update the user_count in message_segments table for 'all_users' segments
    UPDATE message_segments 
    SET user_count = actual_player_count
    WHERE segment_type = 'all_users';
    
    -- Log the update
    RAISE NOTICE 'Updated campaigns with actual player count (excluding admins): %', actual_player_count;
END $$;

-- Show user breakdown for verification
SELECT 
    'Total Users' as user_category,
    COUNT(*) as count
FROM users
UNION ALL
SELECT 
    'Admin Users' as user_category,
    COUNT(*) as count
FROM users 
WHERE is_admin = TRUE
UNION ALL
SELECT 
    'Player Users' as user_category,
    COUNT(*) as count
FROM users 
WHERE is_admin IS NOT TRUE AND user_type = 'PLAYER'
UNION ALL
SELECT 
    'Other User Types' as user_category,
    COUNT(*) as count
FROM users 
WHERE is_admin IS NOT TRUE AND user_type != 'PLAYER';

-- Show the updated results
SELECT 
    mc.id,
    mc.title,
    mc.total_recipients,
    ms.segment_type,
    ms.user_count
FROM message_campaigns mc
LEFT JOIN message_segments ms ON mc.id = ms.campaign_id
ORDER BY mc.created_at DESC;
