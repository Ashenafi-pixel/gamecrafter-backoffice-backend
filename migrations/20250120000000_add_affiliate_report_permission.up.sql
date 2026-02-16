-- Migration: Add Affiliate Report Permission
-- This migration adds the permission for viewing the affiliate report

INSERT INTO permissions (id, name, description, requires_value)
SELECT gen_random_uuid(), 'view affiliate report', 'Access to view affiliate report with daily metrics grouped by referral code', FALSE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'view affiliate report');

