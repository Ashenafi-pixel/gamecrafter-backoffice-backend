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
    ('08706e9a-1a80-464f-b95d-570eac67071a'::uuid, 'player3', 'player3@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Alice', 'Johnson', '+1234567893', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '123 Main St', 'New York', 'NY', 'USA', '10001', '1990-05-15', NOW()),
    ('be48c205-f531-4bd0-a1cf-9d79f0e4ee2b'::uuid, 'player4', 'player4@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Bob', 'Williams', '+1234567894', 'USD', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '456 Oak Ave', 'Los Angeles', 'CA', 'USA', '90210', '1985-08-22', NOW()),
    ('3bc14f80-b572-4b0e-95f0-dee924c58fcb'::uuid, 'player5', 'player5@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Carol', 'Brown', '+1234567895', 'EUR', 'SUSPENDED', false, 'PLAYER', false, 'INACTIVE', '789 Pine Rd', 'London', 'England', 'UK', 'SW1A 1AA', '1992-12-03', NOW()),
    ('79565f37-3213-4a18-81cb-2722933bad85'::uuid, 'player6', 'player6@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'David', 'Jones', '+1234567896', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '321 Elm St', 'Chicago', 'IL', 'USA', '60601', '1988-03-18', NOW()),
    ('1bef0394-162e-405d-8ea8-e2a4dad9aba7'::uuid, 'player7', 'player7@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Eva', 'Garcia', '+1234567897', 'USD', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '654 Maple Dr', 'Houston', 'TX', 'USA', '77001', '1995-07-25', NOW()),
    ('7810b953-7e43-402c-a981-fa1ec90f3af3'::uuid, 'player8', 'player8@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Frank', 'Miller', '+1234567898', 'EUR', 'INACTIVE', false, 'PLAYER', false, 'INACTIVE', '987 Cedar Ln', 'Berlin', 'Berlin', 'Germany', '10115', '1983-11-12', NOW()),
    ('417df173-d269-4711-b39c-6e7295264f64'::uuid, 'player9', 'player9@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Grace', 'Davis', '+1234567899', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '147 Birch St', 'Phoenix', 'AZ', 'USA', '85001', '1991-09-08', NOW()),
    ('00bedfd4-160b-4b00-a605-e028a7822bca'::uuid, 'player10', 'player10@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Henry', 'Rodriguez', '+1234567900', 'USD', 'SUSPENDED', false, 'PLAYER', true, 'PENDING', '258 Spruce Ave', 'Philadelphia', 'PA', 'USA', '19101', '1987-01-30', NOW()),
    ('6c750273-34f5-4df6-bd7a-ff6505221e4a'::uuid, 'player11', 'player11@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Ivy', 'Martinez', '+1234567901', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '369 Willow Rd', 'Paris', 'Ile-de-France', 'France', '75001', '1994-06-14', NOW()),
    ('36d0bb0e-87f3-4060-8357-3baf892e3799'::uuid, 'player12', 'player12@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Jack', 'Hernandez', '+1234567902', 'USD', 'ACTIVE', false, 'PLAYER', false, 'PENDING', '741 Poplar St', 'San Antonio', 'TX', 'USA', '78201', '1989-04-07', NOW()),
    ('b44c85da-3a07-47d5-b3df-70fc89344de9'::uuid, 'player13', 'player13@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Kate', 'Lopez', '+1234567903', 'USD', 'INACTIVE', false, 'PLAYER', true, 'INACTIVE', '852 Ash Dr', 'San Diego', 'CA', 'USA', '92101', '1993-10-19', NOW()),
    ('0493fd41-2f57-42d5-be1b-10f6d3481d31'::uuid, 'player14', 'player14@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Leo', 'Gonzalez', '+1234567904', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '963 Hickory Ln', 'Madrid', 'Madrid', 'Spain', '28001', '1986-02-26', NOW()),
    ('fb776af1-f495-454e-b516-6896e43467e0'::uuid, 'player15', 'player15@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Maya', 'Wilson', '+1234567905', 'USD', 'SUSPENDED', false, 'PLAYER', false, 'PENDING', '174 Walnut St', 'Dallas', 'TX', 'USA', '75201', '1990-12-11', NOW()),
    ('0911f68a-67fd-4824-9999-78da266709fc'::uuid, 'player16', 'player16@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Noah', 'Anderson', '+1234567906', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '285 Cherry Ave', 'San Jose', 'CA', 'USA', '95101', '1984-08-05', NOW()),
    ('8abfde70-1279-4890-868a-868aa1e2019e'::uuid, 'player17', 'player17@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Olivia', 'Thomas', '+1234567907', 'EUR', 'ACTIVE', false, 'PLAYER', true, 'PENDING', '396 Chestnut Rd', 'Rome', 'Lazio', 'Italy', '00100', '1997-05-23', NOW()),
    ('a327cf75-73df-4a36-ad0e-7f5f1dc3e9a7'::uuid, 'player18', 'player18@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Paul', 'Taylor', '+1234567908', 'USD', 'INACTIVE', false, 'PLAYER', false, 'INACTIVE', '407 Sycamore Dr', 'Austin', 'TX', 'USA', '73301', '1981-11-16', NOW()),
    ('5762caa1-75ee-4010-b468-28b004605ff9'::uuid, 'player19', 'player19@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Quinn', 'Moore', '+1234567909', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '518 Magnolia Ln', 'Jacksonville', 'FL', 'USA', '32201', '1996-03-09', NOW()),
    ('eb03d4ff-4cb5-4d7d-b07b-621533c6c705'::uuid, 'player20', 'player20@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Ruby', 'Jackson', '+1234567910', 'EUR', 'SUSPENDED', false, 'PLAYER', true, 'PENDING', '629 Dogwood St', 'Columbus', 'OH', 'USA', '43201', '1988-07-31', NOW()),
    ('16074cd2-e3b9-470c-bcf4-5ced75cd97fb'::uuid, 'player21', 'player21@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Sam', 'Martin', '+1234567911', 'USD', 'ACTIVE', false, 'PLAYER', true, 'VERIFIED', '730 Redwood Ave', 'Charlotte', 'NC', 'USA', '28201', '1992-01-13', NOW()),
    ('c40dfedb-4240-412b-881c-653b6c5c1477'::uuid, 'player22', 'player22@tucanbit.com', '$2a$10$YourBcryptHashedPasswordHere', 'Tina', 'Lee', '+1234567912', 'USD', 'ACTIVE', false, 'PLAYER', false, 'PENDING', '841 Sequoia Rd', 'Seattle', 'WA', 'USA', '98101', '1985-09-27', NOW())

ON CONFLICT (id) DO NOTHING;

-- Create balances for all 20 players
INSERT INTO balances (user_id, currency, real_money, bonus_money, points, updated_at)
VALUES 
    ('08706e9a-1a80-464f-b95d-570eac67071a'::uuid, 'USD', 2500.00, 250.00, 250, NOW()),
    ('be48c205-f531-4bd0-a1cf-9d79f0e4ee2b'::uuid, 'USD', 1800.00, 180.00, 180, NOW()),
    ('3bc14f80-b572-4b0e-95f0-dee924c58fcb'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('79565f37-3213-4a18-81cb-2722933bad85'::uuid, 'USD', 3200.00, 320.00, 320, NOW()),
    ('1bef0394-162e-405d-8ea8-e2a4dad9aba7'::uuid, 'USD', 1500.00, 150.00, 150, NOW()),
    ('7810b953-7e43-402c-a981-fa1ec90f3af3'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('417df173-d269-4711-b39c-6e7295264f64'::uuid, 'USD', 4100.00, 410.00, 410, NOW()),
    ('00bedfd4-160b-4b00-a605-e028a7822bca'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('6c750273-34f5-4df6-bd7a-ff6505221e4a'::uuid, 'EUR', 2800.00, 280.00, 280, NOW()),
    ('36d0bb0e-87f3-4060-8357-3baf892e3799'::uuid, 'USD', 1900.00, 190.00, 190, NOW()),
    ('b44c85da-3a07-47d5-b3df-70fc89344de9'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('0493fd41-2f57-42d5-be1b-10f6d3481d31'::uuid, 'EUR', 3600.00, 360.00, 360, NOW()),
    ('fb776af1-f495-454e-b516-6896e43467e0'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('0911f68a-67fd-4824-9999-78da266709fc'::uuid, 'USD', 2700.00, 270.00, 270, NOW()),
    ('8abfde70-1279-4890-868a-868aa1e2019e'::uuid, 'EUR', 2200.00, 220.00, 220, NOW()),
    ('a327cf75-73df-4a36-ad0e-7f5f1dc3e9a7'::uuid, 'USD', 0.00, 0.00, 0, NOW()),
    ('5762caa1-75ee-4010-b468-28b004605ff9'::uuid, 'USD', 3300.00, 330.00, 330, NOW()),
    ('eb03d4ff-4cb5-4d7d-b07b-621533c6c705'::uuid, 'EUR', 0.00, 0.00, 0, NOW()),
    ('16074cd2-e3b9-470c-bcf4-5ced75cd97fb'::uuid, 'USD', 2900.00, 290.00, 290, NOW()),
    ('c40dfedb-4240-412b-881c-653b6c5c1477'::uuid, 'USD', 1600.00, 160.00, 160, NOW())

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
