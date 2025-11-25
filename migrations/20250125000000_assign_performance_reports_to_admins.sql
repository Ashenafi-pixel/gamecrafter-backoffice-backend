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

