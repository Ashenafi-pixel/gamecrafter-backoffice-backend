-- Grant permissions on game_providers table to game_crafter_user
-- Run this as the postgres superuser or database owner

-- Grant SELECT, INSERT, UPDATE, DELETE on game_providers table
GRANT SELECT, INSERT, UPDATE, DELETE ON game_providers TO game_crafter_user;

-- Grant USAGE on the sequence (for auto-generated UUIDs)
GRANT USAGE ON SCHEMA public TO game_crafter_user;

-- If using gen_random_uuid() function, ensure the user has access
-- (This is usually available by default, but just in case)
GRANT EXECUTE ON FUNCTION gen_random_uuid() TO game_crafter_user;

-- Also grant permissions on related tables if needed
GRANT SELECT, INSERT, UPDATE, DELETE ON brand_providers TO game_crafter_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON brand_games TO game_crafter_user;

