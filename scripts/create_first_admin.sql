-- ============================================================================
-- CREATE FIRST ADMIN USER
-- ============================================================================
-- This script creates the first admin user with super admin privileges
-- 
-- INSTRUCTIONS:
-- 1. Generate a password hash using: go run scripts/generate_password_hash.go -password 'YourPassword123!'
-- 2. Replace 'YOUR_PASSWORD_HASH_HERE' below with the generated hash
-- 3. Customize the user details (username, email, first_name, last_name)
-- 4. Run this script in your PostgreSQL database
-- ============================================================================

DO $$
DECLARE
    v_user_id UUID := gen_random_uuid();
    v_super_role_id UUID;
    v_balance_id UUID := gen_random_uuid();
    v_super_permission_id UUID;
BEGIN
    -- ========================================================================
    -- STEP 1: Insert the admin user
    -- ========================================================================
    INSERT INTO users (
        id,
        username,
        email,
        password,
        is_admin,
        status,
        kyc_status,
        default_currency,
        first_name,
        last_name,
        is_email_verified,
        type,
        created_at,
        updated_at
    ) VALUES (
        v_user_id,
        'admin',                    -- CHANGE: Your desired username
        'admin@example.com',        -- CHANGE: Your email address
        'YOUR_PASSWORD_HASH_HERE',  -- REPLACE: Use script to generate hash
        true,                        -- is_admin = true
        'ACTIVE',                    -- status
        'PENDING',                   -- kyc_status
        'USD',                       -- default_currency
        'Admin',                     -- CHANGE: Your first name
        'User',                      -- CHANGE: Your last name
        true,                        -- is_email_verified
        'ADMIN',                     -- type
        NOW(),                       -- created_at
        NOW()                        -- updated_at
    )
    ON CONFLICT (id) DO NOTHING;

    -- ========================================================================
    -- STEP 2: Get or create super role
    -- ========================================================================
    -- Check if super role exists (from seed migration)
    SELECT id INTO v_super_role_id 
    FROM roles 
    WHERE name = 'super' 
    LIMIT 1;

    -- If super role doesn't exist, create it
    IF v_super_role_id IS NULL THEN
        v_super_role_id := gen_random_uuid();
        INSERT INTO roles (id, name, description)
        VALUES (v_super_role_id, 'super', 'Super Admin Role')
        ON CONFLICT (id) DO NOTHING;
        
        -- Get the super permission
        SELECT id INTO v_super_permission_id
        FROM permissions
        WHERE name = 'super'
        LIMIT 1;

        -- If super permission doesn't exist, create it
        IF v_super_permission_id IS NULL THEN
            v_super_permission_id := gen_random_uuid();
            INSERT INTO permissions (id, name, description, requires_value)
            VALUES (v_super_permission_id, 'super', 'Super Admin Permission', false)
            ON CONFLICT (id) DO NOTHING;
        END IF;

        -- Assign super permission to super role
        INSERT INTO role_permissions (role_id, permission_id, value)
        VALUES (v_super_role_id, v_super_permission_id, NULL)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END IF;

    -- ========================================================================
    -- STEP 3: Assign user to super role
    -- ========================================================================
    INSERT INTO user_roles (user_id, role_id)
    VALUES (v_user_id, v_super_role_id)
    ON CONFLICT (user_id, role_id) DO NOTHING;

    -- ========================================================================
    -- STEP 4: Create initial balance (optional)
    -- ========================================================================
    INSERT INTO balances (
        id,
        user_id,
        currency_code,
        amount_cents,
        amount_units,
        reserved_cents,
        reserved_units,
        brand_id,
        created_at,
        updated_at
    ) VALUES (
        v_balance_id,
        v_user_id,
        'USD',
        0,                                          -- Initial balance in cents
        0.00,                                       -- Initial balance in units
        0,
        0.00,
        '00000000-0000-0000-0000-000000000001',    -- Default brand ID
        NOW(),
        NOW()
    )
    ON CONFLICT (id) DO NOTHING;

    -- ========================================================================
    -- STEP 5: Output results
    -- ========================================================================
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Admin user created successfully!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'User ID: %', v_user_id;
    RAISE NOTICE 'Username: admin';  -- Update if changed
    RAISE NOTICE 'Email: admin@example.com';  -- Update if changed
    RAISE NOTICE 'Role: super';
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    RAISE NOTICE 'You can now login with:';
    RAISE NOTICE '  Email: admin@example.com';  -- Update if changed
    RAISE NOTICE '  Password: [Your password]';
    RAISE NOTICE '';
    RAISE NOTICE 'Login endpoint: POST /api/admin/login';
    RAISE NOTICE '========================================';

END $$;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================
-- Run these queries to verify the user was created correctly:

-- Check user details
SELECT 
    u.id,
    u.username,
    u.email,
    u.is_admin,
    u.status,
    u.type,
    r.name as role_name
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@example.com';  -- Update email if changed

-- Check role permissions
SELECT 
    r.name as role_name,
    p.name as permission_name,
    rp.value as permission_value
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'super';

-- Check balance
SELECT 
    b.id,
    b.user_id,
    u.email,
    b.currency_code,
    b.amount_units,
    b.reserved_units
FROM balances b
JOIN users u ON b.user_id = u.id
WHERE u.email = 'admin@example.com';  -- Update email if changed

