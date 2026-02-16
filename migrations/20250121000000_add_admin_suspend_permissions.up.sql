-- Add suspend and unsuspend admin user permissions
INSERT INTO permissions (name, description, requires_value) VALUES
('suspend admin user', 'Permission to suspend admin users', FALSE),
('unsuspend admin user', 'Permission to unsuspend admin users', FALSE)
ON CONFLICT (name) DO NOTHING;

