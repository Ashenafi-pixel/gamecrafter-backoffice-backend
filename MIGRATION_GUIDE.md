# 🚀 TucanBIT Database Migration System

This document explains how to handle database migrations when working with the TucanBIT backend, especially when using the server database.

## 📋 Overview

The TucanBIT backend uses **golang-migrate** for database schema management. When working with the server database, you need to be careful not to run migrations that could affect the production data.

## 🔧 Migration Tools

### 1. Migration Script (`migrate.sh`)

The main migration script provides these commands:

```bash
# Check migration status
./migrate.sh status

# Run all pending migrations
./migrate.sh up

# Rollback migrations
./migrate.sh down

# Force migration version (use with caution)
./migrate.sh force

# Create new migration
./migrate.sh create

# Show help
./migrate.sh help
```

### 2. Makefile Commands

```bash
# Run migrations
make migrate-up

# Run backend with local database
make run-local

# Run backend with server database (recommended)
make run-server-db
```

## 🌐 Working with Server Database

### Prerequisites

1. **SSH Tunnel**: Ensure your SSH tunnel is running:
   ```bash
   ssh -fN -L 5433:localhost:5433 ubuntu@13.48.56.1317 -i ~/Developer/Upwork/Tucanbit/Tucanbit/TucanBIT.pem
   ```

2. **Environment Variables**: The backend automatically detects server database usage and skips permission initialization.

### Running Backend with Server Database

```bash
# Option 1: Use the script directly
./run-server-db.sh

# Option 2: Use Makefile
make run-server-db

# Option 3: Manual setup
export SKIP_PERMISSION_INIT=true
go run cmd/main.go
```

## 🔄 Team Sync Workflow

When your team makes database changes:

### 1. **Pull Latest Code**
```bash
git pull origin develop
```

### 2. **Check Migration Status**
```bash
./migrate.sh status
```

### 3. **Run Migrations** (if needed)
```bash
./migrate.sh up
```

### 4. **Start Backend**
```bash
make run-server-db
```

## 📁 Migration Files

Migration files are located in `migrations/` directory:
- `*.up.sql` - Forward migration (creates/modifies tables)
- `*.down.sql` - Rollback migration (undoes changes)

### Example Migration File Structure:
```
migrations/
├── 20241120115831_users.up.sql
├── 20241120115831_users.down.sql
├── 20241120120000_cashback_system.up.sql
└── 20241120120000_cashback_system.down.sql
```

## ⚠️ Important Notes

### When Using Server Database:
- ✅ **DO**: Use `make run-server-db` or `./run-server-db.sh`
- ✅ **DO**: Check migration status before running
- ❌ **DON'T**: Run migrations without team approval
- ❌ **DON'T**: Use `make run-local` (uses local database)

### When Using Local Database:
- ✅ **DO**: Use `make run-local` for development
- ✅ **DO**: Run migrations freely for testing
- ✅ **DO**: Create new migrations as needed

## 🛠️ Creating New Migrations

### 1. Create Migration Files
```bash
./migrate.sh create add_new_feature
```

This creates:
- `migrations/000001_add_new_feature.up.sql`
- `migrations/000001_add_new_feature.down.sql`

### 2. Edit Migration Files

**Up Migration** (`*.up.sql`):
```sql
-- Add new table
CREATE TABLE new_feature (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add new column
ALTER TABLE users ADD COLUMN new_field VARCHAR(100);
```

**Down Migration** (`*.down.sql`):
```sql
-- Remove new table
DROP TABLE IF EXISTS new_feature;

-- Remove new column
ALTER TABLE users DROP COLUMN IF EXISTS new_field;
```

### 3. Test Migrations
```bash
# Test locally first
make run-local
./migrate.sh up
./migrate.sh down  # Test rollback
```

## 🔍 Troubleshooting

### Common Issues:

1. **"Permission denied" errors**:
   - Solution: Use `make run-server-db` instead of `make run-local`

2. **"Migration already applied"**:
   - Solution: Check status with `./migrate.sh status`

3. **"SSH tunnel not detected"**:
   - Solution: Start SSH tunnel first

4. **"Database connection failed"**:
   - Solution: Verify SSH tunnel and database credentials

### Debug Commands:
```bash
# Check SSH tunnel
nc -z localhost 5433 && echo "Tunnel OK" || echo "Tunnel FAILED"

# Check database connection
PGPASSWORD="5kj0YmV5FKKpU9D50B7yH5A" psql -h localhost -p 5433 -U tucanbit -d tucanbit -c "SELECT 1;"

# Check migration status
./migrate.sh status
```

## 📞 Team Communication

When you need to run migrations:

1. **Notify the team** in your communication channel
2. **Describe the changes** you're about to make
3. **Wait for approval** before running migrations
4. **Test locally first** if possible

## 🎯 Best Practices

1. **Always test migrations locally** before applying to server
2. **Create both up and down migrations** for every change
3. **Use descriptive migration names**
4. **Keep migrations small and focused**
5. **Document breaking changes**
6. **Coordinate with team** for production migrations

---

## 📚 Additional Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Team Communication Guidelines](#) (link to your team docs)
