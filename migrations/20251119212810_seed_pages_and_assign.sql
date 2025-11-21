-- SQL Script to Seed Pages and Assign to Admin Users
-- Run this after the migration 20251119212810_add_allowed_pages_system.up.sql

-- Step 1: Insert Parent Pages (sidebar main items)
-- Using INSERT ... ON CONFLICT to avoid duplicates if run multiple times

INSERT INTO pages (path, label, parent_id, icon) VALUES
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
ON CONFLICT (path) DO NOTHING;

-- Step 2: Insert Child Pages (sub-menu items and routes)
-- Using subqueries to get parent_id from parent pages

INSERT INTO pages (path, label, parent_id, icon)
SELECT 
    child.path,
    child.label,
    parent.id as parent_id,
    NULL as icon
FROM (VALUES
    -- Reports children
    ('/reports', 'Analytics Dashboard', '/reports'),
    ('/reports/transaction', 'Wallet Report', '/reports'),
    ('/reports/daily', 'Daily Report', '/reports'),
    ('/reports/game', 'Game Report', '/reports'),
    
    -- Rakeback children
    ('/cashback', 'VIP Levels', '/cashback'),
    ('/admin/rakeback-override', 'Happy Hour', '/cashback'),
    
    -- Wallet children
    ('/transactions/details', 'Transaction Details', '/wallet'),
    ('/transactions/withdrawals/dashboard', 'Withdrawal Dashboard', '/wallet'),
    ('/transactions/deposits', 'Deposit Management', '/wallet'),
    ('/transactions/manual-funds', 'Fund Management', '/wallet'),
    ('/wallet/management', 'Wallet Management', '/wallet'),
    ('/transactions/withdrawals', 'Withdrawal Management', '/wallet'),
    ('/transactions/withdrawals/settings', 'Withdrawal Settings', '/wallet'),
    
    -- KYC Management children
    ('/kyc-risk', 'KYC Risk Management', '/kyc-management')
) AS child(path, label, parent_path)
INNER JOIN pages parent ON parent.path = child.parent_path
ON CONFLICT (path) DO NOTHING;

-- Step 3: Assign all pages to all existing admin users
-- This query will insert all pages for each admin user

INSERT INTO user_allowed_pages (user_id, page_id)
SELECT DISTINCT
    u.id as user_id,
    p.id as page_id
FROM users u
CROSS JOIN pages p
WHERE u.is_admin = true 
  AND u.status = 'ACTIVE'
  AND u.user_type = 'ADMIN'
ON CONFLICT (user_id, page_id) DO NOTHING;

-- Verification Queries (optional - run to check results)

-- Count total pages
-- SELECT COUNT(*) as total_pages FROM pages;

-- Count pages by type (parent vs child)
-- SELECT 
--     CASE WHEN parent_id IS NULL THEN 'Parent' ELSE 'Child' END as page_type,
--     COUNT(*) as count
-- FROM pages
-- GROUP BY page_type;

-- Count pages assigned per admin user
-- SELECT 
--     u.username,
--     u.email,
--     COUNT(uap.page_id) as pages_assigned
-- FROM users u
-- LEFT JOIN user_allowed_pages uap ON u.id = uap.user_id
-- WHERE u.is_admin = true
-- GROUP BY u.id, u.username, u.email
-- ORDER BY pages_assigned DESC;

-- List all pages with their parent (if any)
-- SELECT 
--     p.path,
--     p.label,
--     parent.path as parent_path,
--     parent.label as parent_label
-- FROM pages p
-- LEFT JOIN pages parent ON p.parent_id = parent.id
-- ORDER BY parent.path NULLS FIRST, p.path;

