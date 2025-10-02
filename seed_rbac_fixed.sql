-- Seed RBAC data for admin user
-- This script creates the necessary roles and permissions for the admin user

-- Insert permissions (if they don't exist)
INSERT INTO permissions (id, name, description) 
SELECT '550e8400-e29b-41d4-a716-446655440001', 'get players', 'allow to get players'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'get players');

INSERT INTO permissions (id, name, description) 
SELECT '550e8400-e29b-41d4-a716-446655440002', 'super', 'supper user has all permissions on the system'
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'super');

-- Insert admin role (if it doesn't exist)
INSERT INTO roles (id, name, description) 
SELECT '660e8400-e29b-41d4-a716-446655440001', 'admin', 'Administrator role with full access'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'admin');

-- Assign permissions to admin role
INSERT INTO role_permissions (role_id, permission_id) 
SELECT '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001'
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions 
    WHERE role_id = '660e8400-e29b-41d4-a716-446655440001' 
    AND permission_id = '550e8400-e29b-41d4-a716-446655440001'
);

INSERT INTO role_permissions (role_id, permission_id) 
SELECT '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002'
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions 
    WHERE role_id = '660e8400-e29b-41d4-a716-446655440001' 
    AND permission_id = '550e8400-e29b-41d4-a716-446655440002'
);

-- Assign admin role to the admin user
INSERT INTO user_roles (user_id, role_id) 
SELECT 'a5e168fb-168e-4183-84c5-d49038ce00b5', '660e8400-e29b-41d4-a716-446655440001'
WHERE NOT EXISTS (
    SELECT 1 FROM user_roles 
    WHERE user_id = 'a5e168fb-168e-4183-84c5-d49038ce00b5' 
    AND role_id = '660e8400-e29b-41d4-a716-446655440001'
);

-- Verify the data
SELECT 
    u.email,
    r.name as role_name,
    p.name as permission_name
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.id = 'a5e168fb-168e-4183-84c5-d49038ce00b5';
