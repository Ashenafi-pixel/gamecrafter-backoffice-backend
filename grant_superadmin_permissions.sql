-- Grant Full Permissions to superadmin user
-- This script grants comprehensive permissions to the superadmin user
-- Run this script as a superuser (postgres) or database owner

-- ==============================================
-- 1. CREATE SUPERADMIN USER (if not exists)
-- ==============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'superadmin') THEN
        CREATE USER superadmin WITH PASSWORD 'SecurePass123';
        RAISE NOTICE 'User superadmin created successfully';
    ELSE
        RAISE NOTICE 'User superadmin already exists';
    END IF;
END
$$;

-- ==============================================
-- 2. DATABASE-LEVEL PERMISSIONS
-- ==============================================
-- Grant connection permission to the database
GRANT CONNECT ON DATABASE tucanbit TO superadmin;

-- Grant usage on public schema
GRANT USAGE ON SCHEMA public TO superadmin;

-- Grant create permission on public schema (for temporary tables, etc.)
GRANT CREATE ON SCHEMA public TO superadmin;

-- ==============================================
-- 3. TABLE-LEVEL PERMISSIONS (ALL TABLES)
-- ==============================================
-- Grant all privileges on all existing tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO superadmin;

-- Grant all privileges on all sequences
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO superadmin;

-- Grant all privileges on all functions
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO superadmin;

-- ==============================================
-- 4. FUTURE OBJECTS PERMISSIONS
-- ==============================================
-- Grant permissions on future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO superadmin;

-- Grant permissions on future sequences
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO superadmin;

-- Grant permissions on future functions
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO superadmin;

-- ==============================================
-- 5. ROLE PERMISSIONS
-- ==============================================
-- Make superadmin a superuser (highest level access)
ALTER USER superadmin WITH SUPERUSER;

-- Alternative: If you don't want superuser privileges, use CREATEDB and CREATEROLE
-- ALTER USER superadmin WITH CREATEDB CREATEROLE;

-- ==============================================
-- 6. ADDITIONAL PERMISSIONS
-- ==============================================
-- Grant permission to create databases
ALTER USER superadmin WITH CREATEDB;

-- Grant permission to create roles
ALTER USER superadmin WITH CREATEROLE;

-- Grant permission to login
ALTER USER superadmin WITH LOGIN;

-- ==============================================
-- 7. VERIFICATION QUERIES
-- ==============================================
-- Display user information
SELECT 
    rolname as username,
    rolsuper as is_superuser,
    rolcreatedb as can_create_db,
    rolcreaterole as can_create_role,
    rolcanlogin as can_login
FROM pg_roles 
WHERE rolname = 'superadmin';

-- Display table permissions
SELECT 
    schemaname,
    tablename,
    tableowner,
    hasinserts,
    hasselects,
    hasupdates,
    hasdeletes
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;

-- Display current user permissions
SELECT 
    grantee,
    table_schema,
    table_name,
    privilege_type
FROM information_schema.table_privileges 
WHERE grantee = 'superadmin'
ORDER BY table_schema, table_name, privilege_type;

-- ==============================================
-- 8. TEST CONNECTION
-- ==============================================
-- Test if superadmin can connect and perform operations
-- This will be executed when superadmin connects
DO $$
BEGIN
    RAISE NOTICE 'superadmin user has been granted full permissions';
    RAISE NOTICE 'Username: superadmin';
    RAISE NOTICE 'Password: SecurePass123';
    RAISE NOTICE 'Database: tucanbit';
    RAISE NOTICE 'Connection string: postgres://superadmin:SecurePass123@localhost:5433/tucanbit';
END
$$;
