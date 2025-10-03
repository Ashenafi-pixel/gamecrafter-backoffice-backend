-- Seed RBAC data for admin user
-- This script creates the necessary roles and permissions for the admin user

-- Insert permissions (if they don't exist)
INSERT INTO permissions (id, name, description) VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'get players', 'allow to get players'),
    ('550e8400-e29b-41d4-a716-446655440002', 'super', 'supper user has all permissions on the system')
ON CONFLICT (id) DO NOTHING;

-- Insert admin role (if it doesn't exist)
INSERT INTO roles (id, name, description) VALUES 
    ('660e8400-e29b-41d4-a716-446655440001', 'admin', 'Administrator role with full access')
ON CONFLICT (id) DO NOTHING;

-- Assign permissions to admin role
INSERT INTO role_permissions (role_id, permission_id) VALUES 
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001'), -- get players
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002')  -- super
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign admin role to the admin user
INSERT INTO user_roles (user_id, role_id) VALUES 
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', '660e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (user_id, role_id) DO NOTHING;

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
