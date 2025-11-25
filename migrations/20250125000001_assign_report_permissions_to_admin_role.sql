-- SQL Script to Assign Report Permissions to Admin Role
-- This assigns the 5 new report permissions to the admin role
-- Permissions: get player metrics report, get country report, get game performance report, get provider performance report, get game players report

-- ============================================
-- STEP 1: Ensure permissions exist (they should be created by InitPermissions)
-- ============================================

-- The permissions should already exist from InitPermissions() call
-- But we'll verify they exist first

-- ============================================
-- STEP 2: Assign report permissions to admin role
-- ============================================

-- Option 1: Assign to role named 'admin role' (if it exists)
-- First, remove any existing assignments to avoid duplicates
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name = 'admin role')
  AND permission_id IN (
    SELECT id FROM permissions WHERE name IN (
      'get player metrics report',
      'get player transactions report',
      'get country report',
      'get game performance report',
      'get game players report',
      'get provider performance report'
    )
  );

-- Then insert the permissions
INSERT INTO role_permissions (role_id, permission_id, value)
SELECT 
    r.id as role_id,
    p.id as permission_id,
    NULL as value  -- NULL = unlimited
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin role'  -- Production uses 'admin role' as the role name
  AND p.name IN (
    'get player metrics report',
    'get player transactions report',
    'get country report',
    'get game performance report',
    'get game players report',
    'get provider performance report'
  )
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp 
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Option 2: Assign to all roles (uncomment if needed)
-- DELETE FROM role_permissions
-- WHERE role_id IN (SELECT id FROM roles WHERE name != 'super')
--   AND permission_id IN (
--     SELECT id FROM permissions WHERE name IN (
--       'get player metrics report',
--       'get player transactions report',
--       'get country report',
--       'get game performance report',
--       'get game players report',
--       'get provider performance report'
--     )
--   );
-- 
-- INSERT INTO role_permissions (role_id, permission_id, value)
-- SELECT 
--     r.id as role_id,
--     p.id as permission_id,
--     NULL as value
-- FROM roles r
-- CROSS JOIN permissions p
-- WHERE r.name != 'super'  -- Exclude super admin (has all permissions)
--   AND p.name IN (
--     'get player metrics report',
--     'get player transactions report',
--     'get country report',
--     'get game performance report',
--     'get game players report',
--     'get provider performance report'
--   )
--   AND NOT EXISTS (
--     SELECT 1 FROM role_permissions rp 
--     WHERE rp.role_id = r.id AND rp.permission_id = p.id
--   );

-- ============================================
-- VERIFICATION: Check if permissions were assigned
-- ============================================

-- 1. Check if permissions exist
SELECT 
    'Permissions Check' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) = 6 THEN '✓ All 6 permissions exist'
        ELSE '✗ Missing permissions - Expected 6, Found ' || COUNT(*)::text
    END as status
FROM permissions
WHERE name IN (
    'get player metrics report',
    'get player transactions report',
    'get country report',
    'get game performance report',
    'get game players report',
    'get provider performance report'
);

-- 2. Check if permissions are assigned to admin role
SELECT 
    r.name as role_name,
    p.name as permission_name,
    rp.value,
    CASE 
        WHEN rp.role_id IS NOT NULL THEN '✓ Assigned'
        ELSE '✗ NOT Assigned'
    END as assignment_status
FROM roles r
CROSS JOIN permissions p
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND p.id = rp.permission_id
WHERE r.name = 'admin role'  -- Production uses 'admin role' as the role name
  AND p.name IN (
    'get player metrics report',
    'get player transactions report',
    'get country report',
    'get game performance report',
    'get game players report',
    'get provider performance report'
  )
ORDER BY p.name;

-- 3. List all roles and their report permissions
SELECT 
    r.name as role_name,
    COUNT(DISTINCT rp.permission_id) as report_permissions_count,
    STRING_AGG(DISTINCT p.name, ', ' ORDER BY p.name) as assigned_permissions
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.name LIKE '%report%'
WHERE r.name != 'super'
GROUP BY r.id, r.name
ORDER BY r.name;

-- ============================================
-- MANUAL ASSIGNMENT: If you need to assign to a specific role by ID
-- ============================================

-- Uncomment and replace ROLE_ID with your actual role ID
-- DELETE FROM role_permissions
-- WHERE role_id = 'ROLE_ID_HERE'::uuid
--   AND permission_id IN (
--     SELECT id FROM permissions WHERE name IN (
--       'get player metrics report',
--       'get player transactions report',
--       'get country report',
--       'get game performance report',
--       'get game players report',
--       'get provider performance report'
--     )
--   );
-- 
-- INSERT INTO role_permissions (role_id, permission_id, value)
-- SELECT 
--     'ROLE_ID_HERE'::uuid,  -- Replace with your role ID
--     p.id,
--     NULL
-- FROM permissions p
-- WHERE p.name IN (
--     'get player metrics report',
--     'get player transactions report',
--     'get country report',
--     'get game performance report',
--     'get game players report',
--     'get provider performance report'
-- )
--   AND NOT EXISTS (
--     SELECT 1 FROM role_permissions rp 
--     WHERE rp.role_id = 'ROLE_ID_HERE'::uuid AND rp.permission_id = p.id
--   );

