-- Add "get duplicate ip accounts report" permission and assign it to admin role
-- Run this SQL script to add the permission and assign it to the admin role

-- Step 1: Insert the permission (if it doesn't exist)
INSERT INTO permissions (id, name, description)
SELECT 
    gen_random_uuid(),
    'get duplicate ip accounts report',
    'allow user to get duplicate IP accounts report'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE LOWER(name) = LOWER('get duplicate ip accounts report')
);

-- Step 2: Get the permission ID
-- (You can check this with: SELECT id, name FROM permissions WHERE name = 'get duplicate ip accounts report';)

-- Step 3: Assign permission to "admin role" (replace the role_id with your actual admin role ID)
-- First, find the admin role ID:
-- SELECT id, name FROM roles WHERE name = 'admin role';

-- Then assign the permission to the admin role:
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin role'
  AND p.name = 'get duplicate ip accounts report'
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp 
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Step 4: Also assign to "super" role (if it exists)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'super'
  AND p.name = 'get duplicate ip accounts report'
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp 
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Verify the permission was added and assigned
SELECT 
    p.name as permission_name,
    r.name as role_name,
    rp.role_id,
    rp.permission_id
FROM permissions p
LEFT JOIN role_permissions rp ON p.id = rp.permission_id
LEFT JOIN roles r ON rp.role_id = r.id
WHERE p.name = 'get duplicate ip accounts report';

