-- SQL Script to Assign Pages to Specific Users
-- Use this if you want to assign pages to specific users instead of all admins

-- ============================================
-- OPTION 1: Assign all pages to specific users by username
-- ============================================

-- Replace 'ashuadmin' and 'superadmin' with your actual usernames
INSERT INTO user_allowed_pages (user_id, page_id)
SELECT 
    u.id,
    p.id
FROM users u
CROSS JOIN pages p
WHERE u.username IN ('ashuadmin', 'superadmin', 'curlyceeeadmin')
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- OPTION 2: Assign all pages to specific users by user_id
-- ============================================

-- Replace the UUIDs with your actual user IDs
INSERT INTO user_allowed_pages (user_id, page_id)
SELECT 
    u.id,
    p.id
FROM users u
CROSS JOIN pages p
WHERE u.id IN (
    '25aa15d1-d89f-4244-a022-ec0642ce1f20',  -- Replace with actual user_id
    'f239b036-a547-46df-846f-995b6c7b55a4',  -- Replace with actual user_id
    '5a8328c7-d51b-4187-b45c-b1beea7b41ff'   -- Replace with actual user_id
)
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- OPTION 3: Assign specific pages to a specific user
-- ============================================

-- Example: Assign only dashboard and players pages to a user
INSERT INTO user_allowed_pages (user_id, page_id)
SELECT 
    u.id,
    p.id
FROM users u
CROSS JOIN pages p
WHERE u.username = 'ashuadmin'  -- Replace with actual username
  AND p.path IN ('/dashboard', '/players')
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- OPTION 4: Remove all pages from a user (before reassigning)
-- ============================================

-- Uncomment to remove all page assignments from a user
-- DELETE FROM user_allowed_pages 
-- WHERE user_id = (SELECT id FROM users WHERE username = 'ashuadmin');

-- ============================================
-- OPTION 5: Assign pages based on user role
-- ============================================

-- Assign all pages to users with 'super' role
INSERT INTO user_allowed_pages (user_id, page_id)
SELECT DISTINCT
    u.id,
    p.id
FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
INNER JOIN roles r ON ur.role_id = r.id
CROSS JOIN pages p
WHERE r.name = 'super'  -- Replace with your role name
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- VERIFICATION: Check user's assigned pages
-- ============================================

-- Check pages assigned to a specific user
SELECT 
    u.username,
    u.email,
    p.path,
    p.label,
    p.parent_id
FROM users u
INNER JOIN user_allowed_pages uap ON u.id = uap.user_id
INNER JOIN pages p ON uap.page_id = p.id
WHERE u.username = 'ashuadmin'  -- Replace with actual username
ORDER BY p.path;

-- Count pages per user
SELECT 
    u.username,
    u.email,
    COUNT(uap.page_id) as total_pages,
    COUNT(CASE WHEN p.parent_id IS NULL THEN 1 END) as parent_pages,
    COUNT(CASE WHEN p.parent_id IS NOT NULL THEN 1 END) as child_pages
FROM users u
LEFT JOIN user_allowed_pages uap ON u.id = uap.user_id
LEFT JOIN pages p ON uap.page_id = p.id
WHERE u.is_admin = true
GROUP BY u.id, u.username, u.email
ORDER BY total_pages DESC;

