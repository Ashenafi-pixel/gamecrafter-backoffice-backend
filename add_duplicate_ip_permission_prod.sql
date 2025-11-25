-- Add "get duplicate ip accounts report" permission to production database
-- Run this on the production database

-- Step 1: Insert the permission (if it doesn't exist)
INSERT INTO permissions (id, name, description)
SELECT 
    gen_random_uuid(),
    'get duplicate ip accounts report',
    'allow user to get duplicate IP accounts report'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE LOWER(name) = LOWER('get duplicate ip accounts report')
);

-- Step 2: Get the permission ID for verification
SELECT 
    id,
    name,
    description,
    'Permission exists' as status
FROM permissions 
WHERE LOWER(name) = LOWER('get duplicate ip accounts report');

-- Step 3: Assign permission to "admin role" (if it exists)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin role'
  AND LOWER(p.name) = LOWER('get duplicate ip accounts report')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp 
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Step 4: Assign permission to "super" role (if it exists)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id,
    p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'super'
  AND LOWER(p.name) = LOWER('get duplicate ip accounts report')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp 
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Step 5: Verify the permission was added and assigned
SELECT 
    p.name as permission_name,
    r.name as role_name,
    'Assigned' as status
FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN roles r ON rp.role_id = r.id
WHERE LOWER(p.name) = LOWER('get duplicate ip accounts report')
ORDER BY r.name;

-- Step 6: Show summary
SELECT 
    (SELECT COUNT(*) FROM permissions WHERE LOWER(name) = LOWER('get duplicate ip accounts report')) as permission_exists,
    (SELECT COUNT(*) FROM role_permissions rp 
     INNER JOIN permissions p ON rp.permission_id = p.id 
     WHERE LOWER(p.name) = LOWER('get duplicate ip accounts report')) as role_assignments;

