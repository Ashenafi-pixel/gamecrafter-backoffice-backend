-- Script to check and fix permissions for kirubel_gizaw32 user
-- This will check the user's role and add the missing "get big winners report" permission

-- ============================================
-- STEP 1: Check user and role information
-- ============================================

SELECT 
    u.id as user_id,
    u.username,
    u.email,
    u.is_admin,
    u.user_type,
    r.id as role_id,
    r.name as role_name
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.username = 'kirubel_gizaw32';

-- ============================================
-- STEP 2: Check if "get big winners report" permission exists
-- ============================================

SELECT 
    id,
    name,
    description
FROM permissions
WHERE name = 'get big winners report';

-- ============================================
-- STEP 3: Check current permissions for the user's role
-- ============================================

SELECT 
    r.name as role_name,
    p.name as permission_name,
    rp.value
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
WHERE u.username = 'kirubel_gizaw32'
  AND p.name LIKE '%report%'
ORDER BY p.name;

-- ============================================
-- STEP 4: Add "get big winners report" permission to the user's role(s)
-- ============================================

-- First, get the role(s) for the user
-- Then add the permission to all roles the user has

INSERT INTO role_permissions (role_id, permission_id, value)
SELECT DISTINCT
    r.id as role_id,
    p.id as permission_id,
    NULL::numeric as value  -- NULL = unlimited
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
CROSS JOIN permissions p
WHERE u.username = 'kirubel_gizaw32'
  AND p.name = 'get big winners report'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp 
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- ============================================
-- STEP 5: Verify the permission was added
-- ============================================

SELECT 
    r.name as role_name,
    p.name as permission_name,
    CASE 
        WHEN rp.role_id IS NOT NULL THEN '✓ Assigned'
        ELSE '✗ NOT Assigned'
    END as status
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
CROSS JOIN permissions p
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND p.id = rp.permission_id
WHERE u.username = 'kirubel_gizaw32'
  AND p.name = 'get big winners report';

