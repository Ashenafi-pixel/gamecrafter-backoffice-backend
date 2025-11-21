-- Simple SQL Script to Seed Pages and Assign to Admin Users
-- Alternative version with step-by-step queries

-- ============================================
-- STEP 1: Insert Parent Pages
-- ============================================

INSERT INTO pages (path, label, parent_id) VALUES
    ('/dashboard', 'Dashboard', NULL),
    ('/reports', 'Reports', NULL),
    ('/players', 'Player Management', NULL),
    ('/notifications', 'Player Notifications', NULL),
    ('/kyc-management', 'KYC Management', NULL),
    ('/cashback', 'Rakeback', NULL),
    ('/admin/game-management', 'Game Management', NULL),
    ('/admin/brand-management', 'Brand Management', NULL),
    ('/admin/falcon-liquidity', 'Falcon Liquidity', NULL),
    ('/wallet', 'Wallet', NULL),
    ('/settings', 'Site Settings', NULL),
    ('/access-control', 'Back Office Settings', NULL),
    ('/admin/activity-logs', 'Admin Activity Logs', NULL),
    ('/admin/alerts', 'Alert Management', NULL)
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 2: Insert Reports Child Pages
-- ============================================

INSERT INTO pages (path, label, parent_id)
SELECT '/reports', 'Analytics Dashboard', id FROM pages WHERE path = '/reports'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/reports/transaction', 'Wallet Report', id FROM pages WHERE path = '/reports'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/reports/daily', 'Daily Report', id FROM pages WHERE path = '/reports'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/reports/game', 'Game Report', id FROM pages WHERE path = '/reports'
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 3: Insert Rakeback Child Pages
-- ============================================

INSERT INTO pages (path, label, parent_id)
SELECT '/cashback', 'VIP Levels', id FROM pages WHERE path = '/cashback'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/admin/rakeback-override', 'Happy Hour', id FROM pages WHERE path = '/cashback'
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 4: Insert Wallet Child Pages
-- ============================================

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/details', 'Transaction Details', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/withdrawals/dashboard', 'Withdrawal Dashboard', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/deposits', 'Deposit Management', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/manual-funds', 'Fund Management', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/wallet/management', 'Wallet Management', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/withdrawals', 'Withdrawal Management', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

INSERT INTO pages (path, label, parent_id)
SELECT '/transactions/withdrawals/settings', 'Withdrawal Settings', id FROM pages WHERE path = '/wallet'
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 5: Insert KYC Management Child Pages
-- ============================================

INSERT INTO pages (path, label, parent_id)
SELECT '/kyc-risk', 'KYC Risk Management', id FROM pages WHERE path = '/kyc-management'
ON CONFLICT (path) DO NOTHING;

-- ============================================
-- STEP 6: Assign All Pages to All Admin Users
-- ============================================

INSERT INTO user_allowed_pages (user_id, page_id)
SELECT 
    u.id,
    p.id
FROM users u
CROSS JOIN pages p
WHERE u.is_admin = true 
  AND u.status = 'ACTIVE'
  AND u.user_type = 'ADMIN'
ON CONFLICT (user_id, page_id) DO NOTHING;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Check total pages created
SELECT COUNT(*) as total_pages FROM pages;

-- Check pages assigned per admin
SELECT 
    u.username,
    u.email,
    COUNT(uap.page_id) as pages_assigned
FROM users u
LEFT JOIN user_allowed_pages uap ON u.id = uap.user_id
WHERE u.is_admin = true
GROUP BY u.id, u.username, u.email
ORDER BY pages_assigned DESC;

-- List all pages with hierarchy
SELECT 
    p.path,
    p.label,
    COALESCE(parent.path, 'ROOT') as parent_path,
    COALESCE(parent.label, 'N/A') as parent_label
FROM pages p
LEFT JOIN pages parent ON p.parent_id = parent.id
ORDER BY parent.path NULLS FIRST, p.path;

