-- Seed 20 players for TucanBIT pagination testing

-- Generate 20 test players with varied data
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
    kyc_status,
    street_address,
    city,
    state,
    country,
    postal_code,
    date_of_birth,
    created_at
)
VALUES 
    -- Player 1
    ('d8e491fe-491f-7416-b7f8-07c361f033e8'::uuid, 'player3', 'player3@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Alice', 'Johnson', '+1234567893', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '123 Main St', 'New York', 'NY', 'USA', '10001', '1990-05-15', NOW()),
    
    -- Player 2
    ('e9f5a2gf-5a2g-8527-c8g9-08d472g044f9'::uuid, 'player4', 'player4@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Bob', 'Williams', '+1234567894', 'USD', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '456 Oak Ave', 'Los Angeles', 'CA', 'USA', '90210', '1985-08-22', NOW()),
    
    -- Player 3
    ('f0g6b3hg-6b3h-9638-d9h0-09e583h055g0'::uuid, 'player5', 'player5@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Carol', 'Brown', '+1234567895', 'EUR', 'SUSPENDED', false, 'PLAYER', false, 'INACTIVE', '789 Pine Rd', 'London', 'England', 'UK', 'SW1A 1AA', '1992-12-03', NOW()),
    
    -- Player 4
    ('g1h7c4ih-7c4i-0749-e0i1-00f694i066h1'::uuid, 'player6', 'player6@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'David', 'Jones', '+1234567896', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '321 Elm St', 'Chicago', 'IL', 'USA', '60601', '1988-03-18', NOW()),
    
    -- Player 5
    ('h2i8d5ji-8d5j-1850-f1j2-01g705j077i2'::uuid, 'player7', 'player7@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Eva', 'Garcia', '+1234567897', 'USD', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '654 Maple Dr', 'Houston', 'TX', 'USA', '77001', '1995-07-25', NOW()),
    
    -- Player 6
    ('i3j9e6kj-9e6k-2961-g2k3-02h816k088j3'::uuid, 'player8', 'player8@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Frank', 'Miller', '+1234567898', 'EUR', 'INACTIVE', false, 'PLAYER', false, 'INACTIVE', '987 Cedar Ln', 'Berlin', 'Berlin', 'Germany', '10115', '1983-11-12', NOW()),
    
    -- Player 7
    ('j4k0f7lk-0f7l-3072-h3l4-03i927l099k4'::uuid, 'player9', 'player9@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Grace', 'Davis', '+1234567899', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '147 Birch St', 'Phoenix', 'AZ', 'USA', '85001', '1991-09-08', NOW()),
    
    -- Player 8
    ('k5l1g8ml-1g8m-4183-i4m5-04j038m000l5'::uuid, 'player10', 'player10@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Henry', 'Rodriguez', '+1234567900', 'USD', 'SUSPENDED', false, 'PLAYER', true, 'PENDING', '258 Spruce Ave', 'Philadelphia', 'PA', 'USA', '19101', '1987-01-30', NOW()),
    
    -- Player 9
    ('l6m2h9nm-2h9n-5294-j5n6-05k149n011m6'::uuid, 'player11', 'player11@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Ivy', 'Martinez', '+1234567901', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '369 Willow Rd', 'Paris', 'Ile-de-France', 'France', '75001', '1994-06-14', NOW()),
    
    -- Player 10
    ('m7n3i0on-3i0o-6305-k6o7-06l250o022n7'::uuid, 'player12', 'player12@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Jack', 'Hernandez', '+1234567902', 'USD', 'ACTIVE', false, 'PLAYER', false, 'PENDING', '741 Poplar St', 'San Antonio', 'TX', 'USA', '78201', '1989-04-07', NOW()),
    
    -- Player 11
    ('n8o4j1po-4j1p-7416-l7p8-07m361p033o8'::uuid, 'player13', 'player13@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Kate', 'Lopez', '+1234567903', 'USD', 'INACTIVE', false, 'PLAYER', true, 'INACTIVE', '852 Ash Dr', 'San Diego', 'CA', 'USA', '92101', '1993-10-19', NOW()),
    
    -- Player 12
    ('o9p5k2qp-5k2q-8527-m8q9-08n472q044p9'::uuid, 'player14', 'player14@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Leo', 'Gonzalez', '+1234567904', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '963 Hickory Ln', 'Madrid', 'Madrid', 'Spain', '28001', '1986-02-26', NOW()),
    
    -- Player 13
    ('p0q6l3rq-6l3r-9638-n9r0-09o583r055q0'::uuid, 'player15', 'player15@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Maya', 'Wilson', '+1234567905', 'USD', 'SUSPENDED', false, 'PLAYER', false, 'PENDING', '174 Walnut St', 'Dallas', 'TX', 'USA', '75201', '1990-12-11', NOW()),
    
    -- Player 14
    ('q1r7m4sr-7m4s-0749-o0s1-00p694s066r1'::uuid, 'player16', 'player16@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Noah', 'Anderson', '+1234567906', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '285 Cherry Ave', 'San Jose', 'CA', 'USA', '95101', '1984-08-05', NOW()),
    
    -- Player 15
    ('r2s8n5ts-8n5t-1850-p1t2-01q705t077s2'::uuid, 'player17', 'player17@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Olivia', 'Thomas', '+1234567907', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '396 Chestnut Rd', 'Rome', 'Lazio', 'Italy', '00100', '1997-05-23', NOW()),
    
    -- Player 16
    ('s3t9o6ut-9o6u-2961-q2u3-02r816u088t3'::uuid, 'player18', 'player18@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Paul', 'Taylor', '+1234567908', 'USD', 'INACTIVE', false, 'PLAYER', false, 'INACTIVE', '407 Sycamore Dr', 'Austin', 'TX', 'USA', '73301', '1981-11-16', NOW()),
    
    -- Player 17
    ('t4u0p7vu-0p7v-3072-r3v4-03s927v099u4'::uuid, 'player19', 'player19@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Quinn', 'Moore', '+1234567909', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '518 Magnolia Ln', 'Jacksonville', 'FL', 'USA', '32201', '1996-03-09', NOW()),
    
    -- Player 18
    ('u5v1q8wv-1q8w-4183-s4w5-04t038w000v5'::uuid, 'player20', 'player20@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Ruby', 'Jackson', '+1234567910', 'EUR', 'SUSPENDED', false, 'PLAYER', true, 'PENDING', '629 Dogwood St', 'Columbus', 'OH', 'USA', '43201', '1988-07-31', NOW()),
    
    -- Player 19
    ('v6w2r9xw-2r9x-5294-t5x6-05u149x011w6'::uuid, 'player21', 'player21@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Sam', 'Martin', '+1234567911', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '730 Redwood Ave', 'Charlotte', 'NC', 'USA', '28201', '1992-01-13', NOW()),
    
    -- Player 20
    ('w7x3s0yx-3s0y-6305-u6y7-06v250y022x7'::uuid, 'player22', 'player22@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Tina', 'Lee', '+1234567912', 'USD', 'ACTIVE', false, 'PLAYER', false, 'PENDING', '841 Sequoia Rd', 'Seattle', 'WA', 'USA', '98101', '1985-09-27', NOW())

ON CONFLICT (id) DO NOTHING;

-- Create balances for all 20 players
INSERT INTO balances (user_id, currency, real_money, bonus_money, points, updated_at)
VALUES 
    ('d8e491fe-491f-7416-b7f8-07c361f033e8'::uuid, 'USD', 2500.00, 250.00, 250, NOW()),
    ('e9f5a2gf-5a2g-8527-c8g9-08d472g044f9'::uuid, 'USD', 1800.00, 180.00, 180, NOW()),
    ('f0g6b3hg-6b3h-9638-d9h0-09e583h055g0'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('g1h7c4ih-7c4i-0749-e0i1-00f694i066h1'::uuid, 'USD', 3200.00, 320.00, 320, NOW()),
    ('h2i8d5ji-8d5j-1850-f1j2-01g705j077i2'::uuid, 'USD', 1500.00, 150.00, 150, NOW()),
    ('i3j9e6kj-9e6k-2961-g2k3-02h816k088j3'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('j4k0f7lk-0f7l-3072-h3l4-03i927l099k4'::uuid, 'USD', 4100.00, 410.00, 410, NOW()),
    ('k5l1g8ml-1g8m-4183-i4m5-04j038m000l5'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('l6m2h9nm-2h9n-5294-j5n6-05k149n011m6'::uuid, 'EUR', 2800.00, 280.00, 280, NOW()),
    ('m7n3i0on-3i0o-6305-k6o7-06l250o022n7'::uuid, 'USD', 1900.00, 190.00, 190, NOW()),
    ('n8o4j1po-4j1p-7416-l7p8-07m361p033o8'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('o9p5k2qp-5k2q-8527-m8q9-08n472q044p9'::uuid, 'EUR', 3600.00, 360.00, 360, NOW()),
    ('p0q6l3rq-6l3r-9638-n9r0-09o583r055q0'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('q1r7m4sr-7m4s-0749-o0s1-00p694s066r1'::uuid, 'USD', 2700.00, 270.00, 270, NOW()),
    ('r2s8n5ts-8n5t-1850-p1t2-01q705t077s2'::uuid, 'EUR', 2200.00, 220.00, 220, NOW()),
    ('s3t9o6ut-9o6u-2961-q2u3-02r816u088t3'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('t4u0p7vu-0p7v-3072-r3v4-03s927v099u4'::uuid, 'USD', 3300.00, 330.00, 330, NOW()),
    ('u5v1q8wv-1q8w-4183-s4w5-04t038w000v5'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('v6w2r9xw-2r9x-5294-t5x6-05u149x011w6'::uuid, 'USD', 2900.00, 290.00, 290, NOW()),
    ('w7x3s0yx-3s0y-6305-u6y7-06v250y022x7'::uuid, 'USD', 1600.00, 160.00, 160, NOW())

ON CONFLICT DO NOTHING;

-- Verify the data
SELECT 'Players created:' as info;
SELECT COUNT(*) as total_players FROM users WHERE user_type = 'PLAYER';

SELECT 'Sample players:' as info;
SELECT id, username, email, first_name, last_name, default_currency, status, kyc_status, is_email_verified
FROM users 
WHERE user_type = 'PLAYER'
ORDER BY username
LIMIT 10;

SELECT 'Balances created:' as info;
SELECT COUNT(*) as total_balances FROM balances;

SELECT 'Sample balances:' as info;
SELECT b.user_id, u.username, b.currency, b.real_money, b.bonus_money, b.points
FROM balances b
JOIN users u ON b.user_id = u.id
WHERE u.user_type = 'PLAYER'
ORDER BY u.username
LIMIT 10;
