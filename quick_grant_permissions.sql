-- Quick Permission Grant Script for superadmin
-- Run this as postgres superuser

-- Create user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'superadmin') THEN
        CREATE USER superadmin WITH PASSWORD 'SecurePass123';
    END IF;
END
$$;

-- Grant superuser privileges (highest level access)
ALTER USER superadmin WITH SUPERUSER CREATEDB CREATEROLE LOGIN;

-- Grant all permissions on current database
GRANT ALL PRIVILEGES ON DATABASE tucanbit TO superadmin;
GRANT ALL PRIVILEGES ON SCHEMA public TO superadmin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO superadmin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO superadmin;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO superadmin;

-- Grant permissions on future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO superadmin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO superadmin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO superadmin;

-- Verify permissions
SELECT 'superadmin permissions granted successfully' as status;
