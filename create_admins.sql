-- Create 10 admin users for RBAC testing
INSERT INTO users (
    username,
    phone_number,
    password,
    email,
    first_name,
    last_name,
    default_currency,
    is_admin,
    user_type,
    status,
    kyc_status,
    street_address,
    country,
    state,
    city,
    postal_code
) VALUES 
    ('admin1', '+1234567890', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin1@tucanbit.com', 'Admin', 'One', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin2', '+1234567891', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin2@tucanbit.com', 'Admin', 'Two', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin3', '+1234567892', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin3@tucanbit.com', 'Admin', 'Three', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin4', '+1234567893', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin4@tucanbit.com', 'Admin', 'Four', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin5', '+1234567894', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin5@tucanbit.com', 'Admin', 'Five', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin6', '+1234567895', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin6@tucanbit.com', 'Admin', 'Six', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin7', '+1234567896', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin7@tucanbit.com', 'Admin', 'Seven', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin8', '+1234567897', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin8@tucanbit.com', 'Admin', 'Eight', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin9', '+1234567898', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin9@tucanbit.com', 'Admin', 'Nine', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000'),
    ('admin10', '+1234567899', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin10@tucanbit.com', 'Admin', 'Ten', 'ETB', true, 'PLAYER', 'ACTIVE', 'VERIFIED', '123 Admin St', 'Ethiopia', 'Addis Ababa', 'Addis Ababa', '1000');

-- Create balances for the admin users
INSERT INTO balances (user_id, currency, real_money, bonus_money, points)
SELECT 
    u.id,
    'ETB',
    0,
    0,
    0
FROM users u 
WHERE u.username IN ('admin1', 'admin2', 'admin3', 'admin4', 'admin5', 'admin6', 'admin7', 'admin8', 'admin9', 'admin10')
AND NOT EXISTS (
    SELECT 1 FROM balances b WHERE b.user_id = u.id AND b.currency = 'ETB'
);

-- Display the created admin users
SELECT 
    username,
    email,
    phone_number,
    first_name,
    last_name,
    is_admin,
    status,
    created_at
FROM users 
WHERE username IN ('admin1', 'admin2', 'admin3', 'admin4', 'admin5', 'admin6', 'admin7', 'admin8', 'admin9', 'admin10')
ORDER BY username;
