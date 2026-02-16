-- Add suspend and unsuspend admin user permissions
INSERT INTO permissions (name, description, requires_value)
SELECT 'suspend admin user', 'Permission to suspend admin users', FALSE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'suspend admin user');

INSERT INTO permissions (name, description, requires_value)
SELECT 'unsuspend admin user', 'Permission to unsuspend admin users', FALSE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'unsuspend admin user');

