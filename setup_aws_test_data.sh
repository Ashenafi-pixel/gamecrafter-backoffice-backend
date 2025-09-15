#!/bin/bash

# Setup test data for GrooveTech API testing on AWS
# This script creates the necessary test data to resolve "Account ID doesn't match session ID" error

echo "🚀 Setting up test data for GrooveTech API testing..."

# Database connection details (update these for your AWS setup)
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="tucanbit"
DB_USER="tucanbit"
DB_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"

# Test data values
USER_ID="a5e168fb-168e-4183-84c5-d49038ce00b5"
ACCOUNT_ID="a5e168fb-168e-4183-84c5-d49038ce00b5"
SESSION_ID="Tucan_8b607aa6-9e17-440e-a33c-d6b86ebc4c83"
GAME_ID="82695"

echo "📋 Test Data Details:"
echo "   User ID: $USER_ID"
echo "   Account ID: $ACCOUNT_ID"
echo "   Session ID: $SESSION_ID"
echo "   Game ID: $GAME_ID"
echo ""

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
    echo "❌ PostgreSQL is not running or not accessible"
    echo "   Please ensure your database is running and accessible"
    exit 1
fi

echo "✅ PostgreSQL is running and accessible"

# Create test data
echo "🔧 Creating test data..."

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF

-- Create a test user if it doesn't exist
INSERT INTO users (id, username, email, password_hash, status, created_at, updated_at)
VALUES (
    '$USER_ID'::uuid,
    'testuser',
    'test@tucanbit.com',
    '\$2a\$10\$example_hash_here',
    'active',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create a test balance for the user
INSERT INTO balances (id, user_id, currency_code, amount_units, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    '$USER_ID'::uuid,
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
    '$USER_ID'::uuid,
    '$ACCOUNT_ID',
    '$SESSION_ID',
    1000.00,
    'USD',
    'active',
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (account_id) DO UPDATE SET
    session_id = '$SESSION_ID',
    balance = 1000.00,
    last_activity = NOW(),
    updated_at = NOW();

-- Create a game session for the test user
INSERT INTO game_sessions (id, user_id, session_id, game_id, device_type, game_mode, home_url, exit_url, history_url, license_type, is_test_account, reality_check_elapsed, reality_check_interval, created_at, expires_at, is_active, last_activity)
VALUES (
    gen_random_uuid(),
    '$USER_ID'::uuid,
    '$SESSION_ID',
    '$GAME_ID',
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
    user_id = '$USER_ID'::uuid,
    game_id = '$GAME_ID',
    device_type = 'desktop',
    game_mode = 'real',
    expires_at = NOW() + INTERVAL '2 hours',
    is_active = true,
    last_activity = NOW();

EOF

if [ $? -eq 0 ]; then
    echo "✅ Test data created successfully!"
else
    echo "❌ Failed to create test data"
    exit 1
fi

# Verify the data was created
echo "🔍 Verifying test data..."

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF

-- Check groove_accounts
SELECT 'groove_accounts' as table_name, user_id, account_id, session_id, balance, status
FROM groove_accounts 
WHERE account_id = '$ACCOUNT_ID';

-- Check game_sessions
SELECT 'game_sessions' as table_name, user_id::text, session_id, game_id, device_type, 
       CASE WHEN is_active THEN 'active' ELSE 'inactive' END as status
FROM game_sessions 
WHERE session_id = '$SESSION_ID';

-- Check balances
SELECT 'balances' as table_name, user_id::text, currency_code, amount_units::text, 'USD', 'active'
FROM balances 
WHERE user_id = '$USER_ID'::uuid 
AND currency_code = 'USD';

EOF

echo ""
echo "🎯 Test Data Setup Complete!"
echo ""
echo "📝 You can now test your GrooveTech API with:"
echo "   Account ID: $ACCOUNT_ID"
echo "   Session ID: $SESSION_ID"
echo "   Game ID: $GAME_ID"
echo ""
echo "🧪 Test URL:"
echo "   {{base_url}}/groove-official/wager?request=wager&accountid=$ACCOUNT_ID&gamesessionid=$SESSION_ID&device=desktop&gameid=$GAME_ID&apiversion=1.2&betamount=10.0&roundid=round_12312&transactionid=txn_45656434"
echo ""
echo "✅ The 'Account ID doesn't match session ID' error should now be resolved!"