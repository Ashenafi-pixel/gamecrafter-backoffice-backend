-- SQL Script to Seed All Pages from Dev to Prod
-- This script creates all pages based on the page.go SeedPages function
-- and assigns them to all admin users

-- ============================================
-- STEP 1: Create Parent Pages (main menu items)
-- ============================================

INSERT INTO pages (path, label, parent_id, icon)
VALUES
    ('/dashboard', 'Dashboard', NULL, NULL),
    ('/reports', 'Reports', NULL, NULL),
    ('/players', 'Player Management', NULL, NULL),
    ('/notifications', 'Player Notifications', NULL, NULL),
    ('/kyc-management', 'KYC Management', NULL, NULL),
    ('/cashback', 'Rakeback', NULL, NULL),
    ('/admin/game-management', 'Game Management', NULL, NULL),
    ('/admin/brand-management', 'Brand Management', NULL, NULL),
    ('/admin/falcon-liquidity', 'Falcon Liquidity', NULL, NULL),
    ('/wallet', 'Wallet', NULL, NULL),
    ('/settings', 'Site Settings', NULL, NULL),
    ('/access-control', 'Back Office Settings', NULL, NULL),
    ('/admin/activity-logs', 'Admin Activity Logs', NULL, NULL),
    ('/admin/alerts', 'Alert Management', NULL, NULL)
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- ============================================
-- STEP 2: Create Child Pages (with parent references)
-- ============================================

-- Reports children
-- Note: /reports itself is a parent page, so we only insert actual children
INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    ('/reports/daily', 'Daily Report', '/reports'),
    ('/reports/big-winners', 'Big Winners', '/reports'),
    ('/reports/player-metrics', 'Player Metrics', '/reports'),
    ('/reports/country', 'Country Report', '/reports'),
    ('/reports/game-performance', 'Game Performance', '/reports'),
    ('/reports/provider-performance', 'Provider Performance', '/reports')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    parent_id = EXCLUDED.parent_id,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- Update /reports page label to "Analytics Dashboard" if it exists as a parent
UPDATE pages 
SET label = 'Analytics Dashboard', updated_at = NOW()
WHERE path = '/reports' AND parent_id IS NULL;

-- Rakeback children
-- Note: /cashback itself is a parent page, so we only insert actual children
INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    ('/admin/rakeback-override', 'Happy Hour', '/cashback')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    parent_id = EXCLUDED.parent_id,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- Create /cashback as "VIP Levels" child page if it doesn't exist as a parent
-- Actually, /cashback should remain a parent. Let's create a separate VIP Levels page
INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    '/cashback/vip-levels',
    'VIP Levels',
    parent.id as parent_id,
    NULL as icon
FROM pages parent
WHERE parent.path = '/cashback'
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    parent_id = EXCLUDED.parent_id,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- Wallet children
INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    ('/transactions/details', 'Transaction Details', '/wallet'),
    ('/transactions/withdrawals/dashboard', 'Withdrawal Dashboard', '/wallet'),
    ('/transactions/deposits', 'Deposit Management', '/wallet'),
    ('/transactions/manual-funds', 'Fund Management', '/wallet'),
    ('/wallet/management', 'Wallet Management', '/wallet'),
    ('/transactions/withdrawals', 'Withdrawal Management', '/wallet'),
    ('/transactions/withdrawals/settings', 'Withdrawal Settings', '/wallet')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    parent_id = EXCLUDED.parent_id,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- KYC Management children
INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    ('/kyc-risk', 'KYC Risk Management', '/kyc-management')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO UPDATE SET 
    label = EXCLUDED.label,
    parent_id = EXCLUDED.parent_id,
    icon = EXCLUDED.icon,
    updated_at = NOW();

-- ============================================
-- STEP 3: Assign All Pages to All Admin Users
-- ============================================

INSERT INTO user_allowed_pages (user_id, page_id)
SELECT DISTINCT
    u.id as user_id,
    p.id as page_id
FROM users u
CROSS JOIN pages p
WHERE u.is_admin = true
  AND u.user_type = 'ADMIN'
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- VERIFICATION: Check results
-- ============================================

-- Count total pages
SELECT 
    'Total Pages' as metric,
    COUNT(*)::text as value
FROM pages
UNION ALL
SELECT 
    'Parent Pages' as metric,
    COUNT(*)::text as value
FROM pages
WHERE parent_id IS NULL
UNION ALL
SELECT 
    'Child Pages' as metric,
    COUNT(*)::text as value
FROM pages
WHERE parent_id IS NOT NULL
UNION ALL
SELECT 
    'Admin Users' as metric,
    COUNT(*)::text as value
FROM users
WHERE is_admin = true AND user_type = 'ADMIN'
UNION ALL
SELECT 
    'Page Assignments' as metric,
    COUNT(*)::text as value
FROM user_allowed_pages uap
JOIN users u ON u.id = uap.user_id
WHERE u.is_admin = true AND u.user_type = 'ADMIN';

-- Show pages by category
SELECT 
    CASE 
        WHEN p.parent_id IS NULL THEN 'Parent'
        ELSE 'Child'
    END as page_type,
    p.path,
    p.label,
    parent.path as parent_path,
    parent.label as parent_label
FROM pages p
LEFT JOIN pages parent ON p.parent_id = parent.id
ORDER BY 
    CASE WHEN p.parent_id IS NULL THEN 0 ELSE 1 END,
    COALESCE(parent.path, p.path),
    p.path;

