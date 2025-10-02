-- Seed test users for TucanBIT

-- Create test user 1 (admin)
INSERT INTO users (
    id, 
    username, 
    email, 
    password, 
    first_name, 
    last_name, 
    phone_number, 
    default_currency, 
    status, 
    is_admin, 
    user_type,
    is_email_verified,
    created_at
)
VALUES (
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'admin',
    'admin@tucanbit.com',
    '$2a$10$YourBcryptHashedPasswordHere',
    'Admin',
    'User',
    '+1234567890',
    'USD',
    'ACTIVE',
    true,
    'ADMIN',
    true,
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create test user 2 (regular player)
INSERT INTO users (
    id, 
    username, 
    email, 
    password, 
    first_name, 
    last_name, 
    phone_number, 
    default_currency, 
    status, 
    is_admin, 
    user_type,
    is_email_verified,
    created_at
)
VALUES (
    'b6e279fc-279f-5294-95d6-e5a149df11c6'::uuid,
    'player1',
    'player1@tucanbit.com',
    '$2a$10$YourBcryptHashedPasswordHere',
    'John',
    'Doe',
    '+1234567891',
    'USD',
    'ACTIVE',
    false,
    'PLAYER',
    true,
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create test user 3 (another player)
INSERT INTO users (
    id, 
    username, 
    email, 
    password, 
    first_name, 
    last_name, 
    phone_number, 
    default_currency, 
    status, 
    is_admin, 
    user_type,
    is_email_verified,
    created_at
)
VALUES (
    'c7e380fd-380e-6305-a6e7-f6b250ef22d7'::uuid,
    'player2',
    'player2@tucanbit.com',
    '$2a$10$YourBcryptHashedPasswordHere',
    'Jane',
    'Smith',
    '+1234567892',
    'EUR',
    'ACTIVE',
    false,
    'PLAYER',
    true,
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create balances for test users
INSERT INTO balances (user_id, currency, real_money, bonus_money, points, updated_at)
VALUES 
    ('a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid, 'USD', 10000.00, 500.00, 1000, NOW()),
    ('b6e279fc-279f-5294-95d6-e5a149df11c6'::uuid, 'USD', 1000.00, 100.00, 100, NOW()),
    ('c7e380fd-380e-6305-a6e7-f6b250ef22d7'::uuid, 'EUR', 500.00, 50.00, 50, NOW())
ON CONFLICT DO NOTHING;

-- Verify the data
SELECT 'Users created:' as info;
SELECT id, username, email, first_name, last_name, default_currency, status, user_type, is_admin
FROM users 
WHERE id IN (
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'b6e279fc-279f-5294-95d6-e5a149df11c6'::uuid,
    'c7e380fd-380e-6305-a6e7-f6b250ef22d7'::uuid
)
ORDER BY username;

SELECT 'Balances created:' as info;
SELECT b.user_id, u.username, b.currency, b.real_money, b.bonus_money, b.points
FROM balances b
JOIN users u ON b.user_id = u.id
WHERE b.user_id IN (
    'a5e168fb-168e-4183-84c5-d49038ce00b5'::uuid,
    'b6e279fc-279f-5294-95d6-e5a149df11c6'::uuid,
    'c7e380fd-380e-6305-a6e7-f6b250ef22d7'::uuid
)
ORDER BY u.username;

