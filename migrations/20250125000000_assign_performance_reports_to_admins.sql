-- SQL Script to Create and Assign the 5 New Performance Report Pages to All Admin Users
-- This creates and assigns: Big Winners, Player Metrics, Country Report, Game Performance, Provider Performance

-- ============================================
-- STEP 1: Create the 5 new report pages in the pages table
-- ============================================

INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    ('/reports/big-winners', 'Big Winners', '/reports'),
    ('/reports/player-metrics', 'Player Metrics', '/reports'),
    ('/reports/country', 'Country Report', '/reports'),
    ('/reports/game-performance', 'Game Performance', '/reports'),
    ('/reports/provider-performance', 'Provider Performance', '/reports')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 2: Assign the 5 new report pages to all admin users
-- ============================================

INSERT INTO user_allowed_pages (user_id, page_id)
SELECT DISTINCT
    u.id,
    p.id
FROM users u
CROSS JOIN pages p
WHERE u.is_admin = true
  AND p.path IN (
    '/reports/big-winners',
    '/reports/player-metrics',
    '/reports/country',
    '/reports/game-performance',
    '/reports/provider-performance'
  )
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- VERIFICATION: Check if pages were assigned
-- ============================================

-- Count how many admin users have each of the 5 new report pages
SELECT 
    p.path,
    p.label,
    COUNT(DISTINCT uap.user_id) as users_with_access
FROM pages p
LEFT JOIN user_allowed_pages uap ON p.id = uap.page_id
LEFT JOIN users u ON uap.user_id = u.id
WHERE p.path IN (
    '/reports/big-winners',
    '/reports/player-metrics',
    '/reports/country',
    '/reports/game-performance',
    '/reports/provider-performance'
  )
GROUP BY p.id, p.path, p.label
ORDER BY p.path;

-- Check all report pages assigned to admin users
SELECT 
    u.username,
    u.email,
    p.path,
    p.label
FROM users u
INNER JOIN user_allowed_pages uap ON u.id = uap.user_id
INNER JOIN pages p ON uap.page_id = p.id
WHERE u.is_admin = true
  AND p.path LIKE '/reports%'
ORDER BY u.username, p.path;

-- ============================================
-- VIEW ALL PAGES WITH PARENT INFORMATION
-- ============================================

SELECT 
    p.id,
    p.path,
    p.label,
    p.parent_id,
    parent.path as parent_path,
    parent.label as parent_label,
    p.icon,
    p.created_at,
    p.updated_at
FROM pages p
LEFT JOIN pages parent ON p.parent_id = parent.id
ORDER BY p.id ASC;

-- ============================================
-- DIAGNOSTIC: Check if pages exist and are assigned
-- ============================================

-- 1. Check if the 5 pages exist in the pages table
SELECT 
    'Pages Check' as check_type,
    COUNT(*) as count,
    CASE 
        WHEN COUNT(*) = 5 THEN '✓ All 5 pages exist'
        ELSE '✗ Missing pages - Expected 5, Found ' || COUNT(*)::text
    END as status
FROM pages
WHERE path IN (
    '/reports/big-winners',
    '/reports/player-metrics',
    '/reports/country',
    '/reports/game-performance',
    '/reports/provider-performance'
);

-- 2. Check if pages are assigned to any admin users
SELECT 
    'Assignment Check' as check_type,
    COUNT(DISTINCT uap.user_id) as admin_users_with_pages,
    COUNT(DISTINCT uap.page_id) as pages_assigned,
    CASE 
        WHEN COUNT(DISTINCT uap.page_id) = 5 THEN '✓ All 5 pages assigned'
        ELSE '✗ Not all pages assigned - Expected 5, Found ' || COUNT(DISTINCT uap.page_id)::text
    END as status
FROM user_allowed_pages uap
INNER JOIN pages p ON uap.page_id = p.id
INNER JOIN users u ON uap.user_id = u.id
WHERE u.is_admin = true
  AND p.path IN (
    '/reports/big-winners',
    '/reports/player-metrics',
    '/reports/country',
    '/reports/game-performance',
    '/reports/provider-performance'
  );

-- 3. Check specific user's allowed pages (replace USERNAME with your username)
-- SELECT 
--     u.username,
--     u.id as user_id,
--     p.path,
--     p.label,
--     CASE 
--         WHEN uap.user_id IS NOT NULL THEN '✓ Assigned'
--         ELSE '✗ NOT Assigned'
--     END as assignment_status
-- FROM users u
-- CROSS JOIN pages p
-- LEFT JOIN user_allowed_pages uap ON u.id = uap.user_id AND p.id = uap.page_id
-- WHERE u.username = 'YOUR_USERNAME_HERE'  -- Replace with your username
--   AND p.path IN (
--     '/reports/big-winners',
--     '/reports/player-metrics',
--     '/reports/country',
--     '/reports/game-performance',
--     '/reports/provider-performance'
--   )
-- ORDER BY p.path;

-- 4. Force assign pages to a specific user (if needed - uncomment and replace USER_ID)
-- INSERT INTO user_allowed_pages (user_id, page_id)
-- SELECT 
--     'YOUR_USER_ID_HERE'::uuid,  -- Replace with your user ID
--     p.id
-- FROM pages p
-- WHERE p.path IN (
--     '/reports/big-winners',
--     '/reports/player-metrics',
--     '/reports/country',
--     '/reports/game-performance',
--     '/reports/provider-performance'
-- )
-- ON CONFLICT (user_id, page_id) DO NOTHING;

