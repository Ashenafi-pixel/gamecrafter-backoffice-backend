-- Migration: Seed Essential Data
-- This migration seeds essential configuration data and admin user
-- Generated from existing production database

-- ============================================================================
-- 1. BRANDS
-- ============================================================================
INSERT INTO brands (id, name, code, domain, is_active, created_at, updated_at) VALUES
('00000000-0000-0000-0000-000000000001', 'Game Crafter', 'game_crafter', 'gamecrafter.io', true, '2025-11-09 14:42:46.068983+00', '2025-11-09 14:42:46.068983+00'),
('00000000-0000-0000-0000-000000000002', 'Game Crafter Production', 'game-crafter-prod', 'prod.gamecrafter.io', true, '2025-11-09 14:31:19.123839+00', '2025-11-09 14:31:19.123839+00'),
('00000000-0000-0000-0000-000000000003', 'Game Crafter Staging', 'game-crafter-staging', 'staging.gamecrafter.io', true, '2025-11-09 14:31:19.123839+00', '2025-11-09 14:31:19.123839+00'),
('00000000-0000-0000-0000-000000000004', 'Game Crafter Development', 'game-crafter-dev', 'dev.gamecrafter.io', true, '2025-11-09 14:31:19.123839+00', '2025-11-09 14:31:19.123839+00')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. CURRENCY CONFIGURATION
-- ============================================================================
INSERT INTO currency_config (currency_code, currency_name, currency_type, decimal_places, smallest_unit_name, is_active, created_at) VALUES
('BNB', 'Binance Coin', 'crypto', 18, 'Jager', true, '2025-09-28 12:29:28.119755+00'),
('BTC', 'Bitcoin', 'crypto', 8, 'Satoshi', true, '2025-09-08 05:49:27.13437+00'),
('ETH', 'Ethereum', 'crypto', 18, 'Wei', true, '2025-09-08 05:49:27.13437+00'),
('LTC', 'Litecoin', 'crypto', 8, 'Photon', true, '2025-09-08 05:49:27.13437+00'),
('MATIC', 'Polygon', 'crypto', 18, 'Wei', true, '2025-09-28 12:29:28.119755+00'),
('SOL', 'Solana', 'crypto', 9, 'Lamport', true, '2025-09-28 12:29:28.119755+00'),
('USD', 'US Dollar', 'fiat', 2, 'Cent', true, '2025-09-08 05:49:27.13437+00'),
('USDC', 'USD Coin', 'crypto', 6, 'Micro-USDC', true, '2025-09-08 05:49:27.13437+00'),
('USDT', 'Tether', 'crypto', 6, 'Micro-USDT', true, '2025-09-08 05:49:27.13437+00')
ON CONFLICT (currency_code) DO NOTHING;

-- ============================================================================
-- 3. SUPPORTED CHAINS
-- ============================================================================
INSERT INTO supported_chains (id, chain_id, chain_name, protocol, is_testnet, native_currency, processor, status, created_at, updated_at) VALUES
('5b59d0f8-a734-410c-a4ad-57db3f1a3030', 'bsc-mainnet', 'Binance Smart Chain', 'BEP-20', false, 'BNB', 'internal', 'inactive', '2025-09-27 00:24:20.267191+00', '2025-09-27 00:24:20.267191+00'),
('c5928463-7eea-4321-aedb-080e4bc6e4ab', 'btc-mainnet', 'Bitcoin Mainnet', 'BTC', false, 'BTC', 'internal', 'active', '2025-10-10 17:17:52.030092+00', '2025-10-10 17:17:52.030092+00'),
('b92c3dd8-1c13-4261-8edf-60dbac1332a4', 'eth-mainnet', 'Ethereum Mainnet', 'ETHEREUM', false, 'ETH', 'internal', 'active', '2025-09-27 00:24:20.267191+00', '2025-11-24 02:41:17.190808+00'),
('f55ea66e-dd3b-47e1-b65d-da6d5a9d868d', 'eth-testnet', 'Ethereum Sepolia Testnet', 'SEPOLIA', true, 'ETH', 'internal', 'inactive', '2025-10-02 13:03:08.408204+00', '2025-10-23 09:55:07.236599+00'),
('c09d319e-b346-492e-9ae4-94dae13d6637', 'polygon-mainnet', 'Polygon Mainnet', 'ERC-20', false, 'MATIC', 'internal', 'inactive', '2025-09-27 00:24:20.267191+00', '2025-09-27 00:24:20.267191+00'),
('ba59641d-0c33-4638-8436-473ba296cae7', 'sol-mainnet', 'Solana Mainnet', 'SOL', false, 'SOL', 'internal', 'active', '2025-09-27 00:24:20.267191+00', '2025-10-21 08:44:05.628783+00'),
('a2d042ac-1581-49bc-85c0-792a2afe1985', 'sol-testnet', 'Solana Testnet', 'SOL', true, 'SOL', 'internal', 'inactive', '2025-09-27 00:24:20.267191+00', '2025-10-20 03:43:50.230718+00')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 4. ROLES
-- ============================================================================
INSERT INTO roles (id, name, description) VALUES
('6d9325c3-ea8c-47c1-ba2b-285d1f7667bb', 'super', ''),
('33dbb86c-e306-4d1d-b7df-cdf556e1ae32', 'admin role', ''),
('3ffeac56-e266-40e2-b2fd-09c2083b0415', 'manager', '')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 5. ADMIN ACTIVITY CATEGORIES
-- ============================================================================
INSERT INTO admin_activity_categories (id, name, description, color, icon, is_active, created_at) VALUES
('22a33143-2c9c-4eb4-8225-109913318385', 'user_management', 'User account management activities', '#3B82F6', 'users', true, '2025-10-16 13:36:49.713497+00'),
('e9cb1fa3-5d4f-423e-91a1-ce486310ffe7', 'financial', 'Financial transactions and balance management', '#10B981', 'dollar-sign', true, '2025-10-16 13:36:49.713497+00'),
('359862e4-119c-4aa1-9aed-816405c37a92', 'security', 'Security-related actions and access control', '#EF4444', 'shield', true, '2025-10-16 13:36:49.713497+00'),
('8e1c8087-de15-4182-bd9f-7242b4b27d76', 'system', 'System configuration and maintenance', '#8B5CF6', 'settings', true, '2025-10-16 13:36:49.713497+00'),
('14236593-1e5c-46c7-a827-288580311240', 'withdrawal', 'Withdrawal management and processing', '#F59E0B', 'arrow-up', true, '2025-10-16 13:36:49.713497+00'),
('ffbc13e7-9c44-4cee-bc04-e4494c059b91', 'game_management', 'Game configuration and management', '#EC4899', 'gamepad', true, '2025-10-16 13:36:49.713497+00'),
('66573253-b0aa-4069-aea5-d3cd620d28dc', 'reports', 'Report generation and analytics', '#06B6D4', 'chart-bar', true, '2025-10-16 13:36:49.713497+00'),
('eafce5b2-e78b-4f5d-a956-e9b022b48039', 'notifications', 'Notification and communication management', '#84CC16', 'bell', true, '2025-10-16 13:36:49.713497+00')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 6. SYSTEM CONFIGURATION
-- ============================================================================
INSERT INTO system_config (config_key, brand_id, config_value) VALUES
('cumulative_kyc_transaction_limit', NULL, '{"enabled": false, "usd_amount_cents": 300}'),
('deposit_margin_percent', NULL, '{"percent": 10}'),
('global_withdrawal_limits', NULL, '{"enabled": true, "max_amount_cents": 100, "min_amount_cents": 1}'),
('require_kyc_on_first_withdrawal', NULL, '{"enabled": false}'),
('withdrawal_limit_validation_enabled', NULL, '{"enabled": false}'),
('withdrawal_margin_percent', NULL, '{"percent": 10}')
ON CONFLICT (config_key, brand_id) DO NOTHING;

INSERT INTO system_config (config_key, brand_id, config_value, description) VALUES
('welcome_bonus_channel_settings', '00000000-0000-0000-0000-000000000001', '{"channels":[]}', 'Welcome bonus channel settings'),
('welcome_bonus_channel_settings', '00000000-0000-0000-0000-000000000002', '{"channels":[]}', 'Welcome bonus channel settings'),
('welcome_bonus_channel_settings', '00000000-0000-0000-0000-000000000003', '{"channels":[]}', 'Welcome bonus channel settings'),
('welcome_bonus_channel_settings', '00000000-0000-0000-0000-000000000004', '{"channels":[]}', 'Welcome bonus channel settings')
ON CONFLICT (config_key, brand_id) DO NOTHING;

-- Welcome bonus settings per brand (with anti-abuse IP configuration)
INSERT INTO system_config (config_key, brand_id, config_value, description) VALUES
('welcome_bonus_settings',
 '00000000-0000-0000-0000-000000000001',
 '{"type":"fixed","enabled":false,"fixed_enabled":false,"percentage_enabled":false,"ip_restriction_enabled":true,"allow_multiple_bonuses_per_ip":false,"fixed_amount":0.0,"percentage":0.0,"max_deposit_amount":0.0,"max_bonus_percentage":90.0}',
 'Welcome bonus settings'),
('welcome_bonus_settings',
 '00000000-0000-0000-0000-000000000002',
 '{"type":"fixed","enabled":false,"fixed_enabled":false,"percentage_enabled":false,"ip_restriction_enabled":true,"allow_multiple_bonuses_per_ip":false,"fixed_amount":0.0,"percentage":0.0,"max_deposit_amount":0.0,"max_bonus_percentage":90.0}',
 'Welcome bonus settings'),
('welcome_bonus_settings',
 '00000000-0000-0000-0000-000000000003',
 '{"type":"fixed","enabled":false,"fixed_enabled":false,"percentage_enabled":false,"ip_restriction_enabled":true,"allow_multiple_bonuses_per_ip":false,"fixed_amount":0.0,"percentage":0.0,"max_deposit_amount":0.0,"max_bonus_percentage":90.0}',
 'Welcome bonus settings'),
('welcome_bonus_settings',
 '00000000-0000-0000-0000-000000000004',
 '{"type":"fixed","enabled":false,"fixed_enabled":false,"percentage_enabled":false,"ip_restriction_enabled":true,"allow_multiple_bonuses_per_ip":false,"fixed_amount":0.0,"percentage":0.0,"max_deposit_amount":0.0,"max_bonus_percentage":90.0}',
 'Welcome bonus settings')
ON CONFLICT (config_key, brand_id) DO UPDATE
SET config_value = EXCLUDED.config_value;

-- ============================================================================
-- 7. CASHBACK TIERS
-- ============================================================================
INSERT INTO cashback_tiers (id, tier_name, tier_level, min_ggr_required, cashback_percentage, is_active, created_at, updated_at) VALUES
('5e5cbd7f-7b93-489a-9a01-5a9c0f1e2c94', 'Bronze', 1, 0.00000000, 10.00, true, '2025-09-15 11:41:40.757764+00', '2025-10-27 12:54:52.699738+00'),
('0d24c973-97b9-415f-b05a-4a80e7f707ef', 'Iron', 2, 1000.00000000, 15.00, true, '2025-09-15 11:41:40.757764+00', '2025-10-27 12:55:51.009072+00'),
('d38233b1-920b-462e-bc90-ca805218eaf0', 'Steel', 3, 5000.00000000, 20.00, true, '2025-09-15 11:41:40.757764+00', '2025-10-27 17:43:25.87166+00'),
('4366770a-5a6e-4536-9d4e-4be6dc1cbb38', 'Gold', 4, 15000.00000000, 25.00, true, '2025-09-15 11:41:40.757764+00', '2025-10-27 12:56:59.776631+00'),
('46272b27-9db1-4e8c-a4e3-dff375396e6b', 'Diamond', 5, 50000.00000000, 30.00, true, '2025-09-15 11:41:40.757764+00', '2025-10-28 13:07:53.049571+00'),
('1fca45b0-c959-4751-9e41-4452737b76c4', 'Crystal', 6, 5000000.00000000, 35.00, true, '2025-10-27 12:57:35.24333+00', '2025-10-28 16:09:27.797707+00'),
('9045328b-06c1-4024-96d9-df133e2826d4', 'Emerald', 7, 20000000.00000000, 40.00, true, '2025-10-27 12:58:44.941397+00', '2025-10-27 12:58:44.941397+00'),
('54f70fb2-8be0-4d70-97c3-d057510fb3d5', 'Royal', 8, 30000000.00000000, 43.00, true, '2025-10-27 12:59:27.07961+00', '2025-10-27 13:00:59.156202+00'),
('472448d0-d1c5-4eec-9366-d2cd1e27e7ff', 'Crown', 9, 40000000.00000000, 47.00, true, '2025-10-27 13:00:07.453504+00', '2025-10-27 13:00:07.453504+00'),
('11d4fc75-e4b1-4b02-8531-e395283f35e8', 'Cosmic', 10, 50000000.00000000, 50.00, true, '2025-10-27 13:00:45.759066+00', '2025-10-27 13:00:45.759066+00')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 8. ADMIN USER: kirubel_gizaw32
-- ============================================================================
-- Password hash: $2a$12$OHjbqF3r4wXcXHfoXEaqMuYP0hlruk.RD.PxEHv7YvD.z14Tfiy06
-- This is a bcrypt hash. The original password is not stored for security.
INSERT INTO users (
    id,
    username,
    email,
    password,
    is_admin,
    status,
    kyc_status,
    default_currency,
    profile,
    first_name,
    last_name,
    date_of_birth,
    source,
    is_email_verified,
    referal_code,
    street_address,
    country,
    state,
    city,
    postal_code,
    created_at
) VALUES (
    '1dba1be4-e7d6-4d99-88cd-604456da0b70',
    'kirubel_gizaw32',
    'kirubel.tech23@gmail.com',
    '$2a$12$OHjbqF3r4wXcXHfoXEaqMuYP0hlruk.RD.PxEHv7YvD.z14Tfiy06',
    true,
    'ACTIVE',
    'PENDING',
    'USD',
    '',
    '',
    '',
    '',
    '',
    false,
    '',
    '',
    '',
    '',
    '',
    '',
    '2025-10-26 21:47:49.893027+00'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 9. USER ROLE ASSIGNMENT
-- ============================================================================
INSERT INTO user_roles (user_id, role_id) VALUES
('1dba1be4-e7d6-4d99-88cd-604456da0b70', '33dbb86c-e306-4d1d-b7df-cdf556e1ae32')
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================================================
-- 10. USER BALANCE (Initial balance for admin user)
-- ============================================================================
INSERT INTO balances (id, user_id, currency_code, amount_cents, amount_units, reserved_cents, reserved_units, updated_at, brand_id) VALUES
('4a368094-09d7-4c10-98e8-cc5ae43e9836', '1dba1be4-e7d6-4d99-88cd-604456da0b70', 'USD', 549650, 5496.500000000000000000, 0, 0.000000000000000000, '2025-11-11 07:29:35.516367+00', '00000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 11. PERMISSIONS (204 permissions)
-- ============================================================================
-- NOTE: To generate this section, run:
--   python3 scripts/generate_seed_data.py | grep -A 1000 "11. PERMISSIONS" | grep -v "^-- End"
-- Then append the output here.
-- The permissions data will be inserted here when generated from the database.

-- ============================================================================
-- 12. ROLE PERMISSIONS (203 role-permission mappings for admin role)
-- ============================================================================
-- NOTE: To generate this section, run:
--   python3 scripts/generate_seed_data.py | grep -A 1000 "12. ROLE PERMISSIONS" | grep -v "^-- End"
-- Then append the output here.
-- The role_permissions data will be inserted here when generated from the database.

-- ============================================================================
-- 13. PAGES (31 pages)
-- ============================================================================
-- NOTE: To generate this section, run:
--   python3 scripts/generate_seed_data.py | grep -A 1000 "13. PAGES" | grep -v "^-- End"
-- Then append the output here.
-- The pages data will be inserted here when generated from the database.

-- ============================================================================
-- 14. ADMIN ACTIVITY ACTIONS (24 actions)
-- ============================================================================
-- NOTE: To generate this section, run:
--   python3 scripts/generate_seed_data.py | grep -A 1000 "14. ADMIN ACTIVITY ACTIONS" | grep -v "^-- End"
-- Then append the output here.
-- The admin_activity_actions data will be inserted here when generated from the database.

