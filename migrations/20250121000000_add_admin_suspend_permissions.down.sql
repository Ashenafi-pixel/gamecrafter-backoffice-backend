-- Remove suspend and unsuspend admin user permissions
DELETE FROM permissions WHERE name IN ('suspend admin user', 'unsuspend admin user');

