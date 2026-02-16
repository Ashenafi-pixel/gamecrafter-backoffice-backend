-- Migration: Add Affiliate Report Permission
-- This migration adds the permission for viewing the affiliate report

INSERT INTO permissions (id, name, description, requires_value) VALUES
(gen_random_uuid(), 'view affiliate report', 'Access to view affiliate report with daily metrics grouped by referral code', FALSE)
ON CONFLICT (name) DO NOTHING;

