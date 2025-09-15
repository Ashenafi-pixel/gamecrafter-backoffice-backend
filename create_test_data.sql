-- Create test data for GrooveTech API testing
-- This script creates the necessary test data to resolve "Account ID doesn't match session ID" error

-- First, let's check if we have any existing users
-- If not, we'll create a test user

-- Create a test user if it doesn't exist
INSERT INTO users (id, username, email, password_hash, status, created_at, updated_at)
VALUES (
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'testuser',
    'test@tucanbit.com',
    '$2a$10$example_hash_here',
    'active',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create a test balance for the user
INSERT INTO balances (id, user_id, currency_code, amount_units, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'USD',
    1000.00,
    NOW(),
    NOW()
) ON CONFLICT (user_id, currency_code) DO UPDATE SET
    amount_units = 1000.00,
    updated_at = NOW();

-- Create a GrooveTech account for the test user
INSERT INTO groove_accounts (id, user_id, account_id, session_id, balance, currency, status, created_at, last_activity, updated_at)
VALUES (
    gen_random_uuid(),
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'a5e168fb-168e-4183-84c5-d49038ce00b5', -- Use user ID as account ID
    'Tucan_8b607aa6-9e17-440e-a33c-d6b86ebc4c83', -- The session ID from your test
    1000.00,
    'USD',
    'active',
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (account_id) DO UPDATE SET
    session_id = 'Tucan_8b607aa6-9e17-440e-a33c-d6b86ebc4c83',
    balance = 1000.00,
    last_activity = NOW(),
    updated_at = NOW();

-- Create a game session for the test user
INSERT INTO game_sessions (id, user_id, session_id, game_id, device_type, game_mode, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval, created_at, expires_at, is_active, last_activity)
VALUES (
    gen_random_uuid(),
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'Tucan_8b607aa6-9e17-440e-a33c-d6b86ebc4c83',
    '82695',
    'desktop',
    'real',
    'https://tucanbit.tv/games',
    'https://tucanbit.tv/responsible-gaming',
    'https://tucanbit.tv/history',
    'Curacao',
    false,
    0,
    60,
    NOW(),
    NOW() + INTERVAL '2 hours',
    true,
    NOW()
) ON CONFLICT (session_id) DO UPDATE SET
    user_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    game_id = '82695',
    device_type = 'desktop',
    game_mode = 'real',
    expires_at = NOW() + INTERVAL '2 hours',
    is_active = true,
    last_activity = NOW();

-- Verify the data was created correctly
SELECT 
    'groove_accounts' as table_name,
    user_id,
    account_id,
    session_id,
    balance,
    status
FROM groove_accounts 
WHERE account_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'

UNION ALL

SELECT 
    'game_sessions' as table_name,
    user_id::text,
    session_id,
    game_id,
    device_type,
    CASE WHEN is_active THEN 'active' ELSE 'inactive' END
FROM game_sessions 
WHERE session_id = 'Tucan_8b607aa6-9e17-440e-a33c-d6b86ebc4c83'

UNION ALL

SELECT 
    'balances' as table_name,
    user_id::text,
    currency_code,
    amount_units::text,
    'USD',
    'active'
FROM balances 
WHERE user_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid 
AND currency_code = 'USD';