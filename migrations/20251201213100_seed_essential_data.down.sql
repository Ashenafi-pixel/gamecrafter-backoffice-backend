-- Migration: Seed Essential Data (Rollback)
-- This migration removes the seeded essential data

-- Remove user balance
DELETE FROM balances WHERE id = '4a368094-09d7-4c10-98e8-cc5ae43e9836';

-- Remove user role assignment
DELETE FROM user_roles WHERE user_id = '1dba1be4-e7d6-4d99-88cd-604456da0b70' AND role_id = '33dbb86c-e306-4d1d-b7df-cdf556e1ae32';

-- Remove admin user
DELETE FROM users WHERE id = '1dba1be4-e7d6-4d99-88cd-604456da0b70';

-- Remove cashback tiers
DELETE FROM cashback_tiers WHERE id IN (
    '5e5cbd7f-7b93-489a-9a01-5a9c0f1e2c94',
    '0d24c973-97b9-415f-b05a-4a80e7f707ef',
    'd38233b1-920b-462e-bc90-ca805218eaf0',
    '4366770a-5a6e-4536-9d4e-4be6dc1cbb38',
    '46272b27-9db1-4e8c-a4e3-dff375396e6b',
    '1fca45b0-c959-4751-9e41-4452737b76c4',
    '9045328b-06c1-4024-96d9-df133e2826d4',
    '54f70fb2-8be0-4d70-97c3-d057510fb3d5',
    '472448d0-d1c5-4eec-9366-d2cd1e27e7ff',
    '11d4fc75-e4b1-4b02-8531-e395283f35e8'
);

-- Remove system configuration
DELETE FROM system_config WHERE config_key IN (
    'cumulative_kyc_transaction_limit',
    'deposit_margin_percent',
    'global_withdrawal_limits',
    'require_kyc_on_first_withdrawal',
    'withdrawal_limit_validation_enabled',
    'withdrawal_margin_percent'
) AND brand_id IS NULL;

DELETE FROM system_config WHERE config_key = 'welcome_bonus_channel_settings' AND brand_id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004'
);

-- Remove welcome bonus settings per brand
DELETE FROM system_config
WHERE config_key = 'welcome_bonus_settings'
  AND brand_id IN (
      '00000000-0000-0000-0000-000000000001',
      '00000000-0000-0000-0000-000000000002',
      '00000000-0000-0000-0000-000000000003',
      '00000000-0000-0000-0000-000000000004'
  );

-- Remove admin activity categories
DELETE FROM admin_activity_categories WHERE id IN (
    '22a33143-2c9c-4eb4-8225-109913318385',
    'e9cb1fa3-5d4f-423e-91a1-ce486310ffe7',
    '359862e4-119c-4aa1-9aed-816405c37a92',
    '8e1c8087-de15-4182-bd9f-7242b4b27d76',
    '14236593-1e5c-46c7-a827-288580311240',
    'ffbc13e7-9c44-4cee-bc04-e4494c059b91',
    '66573253-b0aa-4069-aea5-d3cd620d28dc',
    'eafce5b2-e78b-4f5d-a956-e9b022b48039'
);

-- Remove roles
DELETE FROM roles WHERE id IN (
    '6d9325c3-ea8c-47c1-ba2b-285d1f7667bb',
    '33dbb86c-e306-4d1d-b7df-cdf556e1ae32',
    '3ffeac56-e266-40e2-b2fd-09c2083b0415'
);

-- Remove supported chains
DELETE FROM supported_chains WHERE id IN (
    '5b59d0f8-a734-410c-a4ad-57db3f1a3030',
    'c5928463-7eea-4321-aedb-080e4bc6e4ab',
    'b92c3dd8-1c13-4261-8edf-60dbac1332a4',
    'f55ea66e-dd3b-47e1-b65d-da6d5a9d868d',
    'c09d319e-b346-492e-9ae4-94dae13d6637',
    'ba59641d-0c33-4638-8436-473ba296cae7',
    'a2d042ac-1581-49bc-85c0-792a2afe1985'
);

-- Remove currency configuration
DELETE FROM currency_config WHERE currency_code IN (
    'BNB', 'BTC', 'ETH', 'LTC', 'MATIC', 'SOL', 'USD', 'USDC', 'USDT'
);

-- Remove brands
DELETE FROM brands WHERE id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004'
);

-- Remove admin activity actions (will be populated when data is generated)
-- DELETE FROM admin_activity_actions WHERE id IN (...);

-- Remove pages (will be populated when data is generated)
-- DELETE FROM pages WHERE id IN (...);

-- Remove role permissions for admin role (will be populated when data is generated)
-- DELETE FROM role_permissions WHERE role_id = '33dbb86c-e306-4d1d-b7df-cdf556e1ae32';

-- Remove permissions (will be populated when data is generated)
-- DELETE FROM permissions WHERE id IN (...);

