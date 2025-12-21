-- Migration: Remove Affiliate Report Permission
-- This migration removes the permission for viewing the affiliate report

DELETE FROM permissions WHERE name = 'view affiliate report';

