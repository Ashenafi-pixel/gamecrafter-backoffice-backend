# Local Development Setup Guide for Backoffice Backend

This guide will help you set up the backoffice-backend project locally to work against the dev ClickHouse and PostgreSQL databases.

## Prerequisites

1. **Git** - For cloning the repository
2. **Docker** - Version 20.10 or higher
3. **Docker Compose** - Version 2.0 or higher (usually included with Docker Desktop)

## Step 1: Clone the Project

```bash
git clone <repository-url>
cd backoffice-backend
```

## Step 2: Install Docker

### For macOS:
1. Download Docker Desktop from https://www.docker.com/products/docker-desktop
2. Install and start Docker Desktop
3. Verify installation:
   ```bash
   docker --version
   docker-compose --version
   ```

### For Linux (Ubuntu/Debian):
```bash
# Update package index
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io docker-compose

# Add your user to docker group (to run without sudo)
sudo usermod -aG docker $USER

# Log out and log back in for group changes to take effect
```

### For Windows:
1. Download Docker Desktop from https://www.docker.com/products/docker-desktop
2. Install and start Docker Desktop
3. Verify installation in PowerShell:
   ```powershell
   docker --version
   docker-compose --version
   ```

## Step 3: Pull Docker Images

From the root project directory, pull the required Docker images:

```bash
docker-compose pull
```

This will download all required images:
- PostgreSQL 13
- Redis 6.2
- ClickHouse 25.8
- Zookeeper 3.8.1
- Kafka 6.2.0
- Confluent Kafka tools

## Step 4: Start Infrastructure Services

Start all supporting services (PostgreSQL, ClickHouse, Redis, Kafka, Zookeeper):

```bash
docker-compose up -d db redis clickhouse zookeeper kafka create-kafka-topic
```

Wait for all services to be healthy (this may take 1-2 minutes):

```bash
# Check service status
docker-compose ps

# Wait until all services show as "healthy"
# You can also check logs:
docker-compose logs -f db
docker-compose logs -f clickhouse
```

## Step 5: Create PostgreSQL Tables

The application uses `go-migrate` to manage database migrations. Migrations run automatically when the app starts, but you can also run them manually.

### Option A: Automatic Migration (Recommended)
Migrations will run automatically when you start the backoffice-app container. The `docker-entrypoint.sh` script handles this.

### Option B: Manual Migration (If needed)

If you need to run migrations manually before starting the app:

```bash
# Connect to PostgreSQL container
docker exec -it tucanbit-db psql -U tucanbit -d tucanbit

# Or run migrations using migrate tool (if installed locally)
# First, install go-migrate:
# macOS: brew install golang-migrate
# Linux: See https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md

# Then run migrations:
export PGPASSWORD=5kj0YmV5FKKpU9D50B7yH5A
migrate -path ./migrations -database "postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable" up
```

### Verify Tables Created

```bash
docker exec tucanbit-db psql -U tucanbit -d tucanbit -c "\dt" | head -20
```

You should see tables like `users`, `transactions`, `roles`, `permissions`, etc.

## Step 6: Restore Data from Backup (If Available)

If you have backup files in a `backup/` folder or provided separately:

### PostgreSQL Backup Restore

```bash
# If you have a SQL dump file:
docker exec -i tucanbit-db psql -U tucanbit -d tucanbit < backup/postgres_dump.sql

# If you have a custom format backup:
docker exec -i tucanbit-db pg_restore -U tucanbit -d tucanbit < backup/postgres_backup.dump

# If backup is a .sql file:
cat backup/postgres_backup.sql | docker exec -i tucanbit-db psql -U tucanbit -d tucanbit
```

### ClickHouse Backup Restore

```bash
# If you have ClickHouse backup files:
# ClickHouse backups are typically in native format or SQL format

# For SQL format:
docker exec -i tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password < backup/clickhouse_backup.sql

# For native format (if available):
# Copy backup files to container and restore
docker cp backup/clickhouse_data/ tucanbit-clickhouse:/var/lib/clickhouse/backup/
# Then use clickhouse-backup tool or manual restore commands
```

**Note:** If no backup folder exists, the migrations will create empty tables with the correct schema. You can populate test data manually or connect to the dev database to export/import specific data.

## Step 7: Create ClickHouse Tables

ClickHouse tables are typically created by the application on first run, but you can also create them manually:

### Check if ClickHouse is accessible:

```bash
docker exec tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password --query "SHOW DATABASES"
```

### Create databases (if needed):

```bash
docker exec tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password --query "CREATE DATABASE IF NOT EXISTS tucanbit_analytics"
docker exec tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password --query "CREATE DATABASE IF NOT EXISTS tucanbit_financial"
```

### Create tables from SQL files (if backup folder contains ClickHouse schema):

```bash
# If you have ClickHouse schema files in backup folder:
docker exec -i tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password --multiquery < backup/clickhouse_schema.sql
```

**Note:** The application will create necessary ClickHouse tables automatically when it starts if they don't exist. Check the application logs for table creation messages.

## Step 8: Build and Run the Application

Build and start the backoffice-app:

```bash
docker-compose up -d --build --force-recreate backoffice-app
```

This command will:
- Build the Go application
- Run database migrations automatically
- Start the application on port 8094

### Monitor the startup:

```bash
# Watch the logs
docker-compose logs -f backoffice-app

# Check if the app is healthy
docker-compose ps backoffice-app
```

The application should be available at: `http://localhost:8094`

## Step 9: Verify Everything is Working

### Check all services are running:

```bash
docker-compose ps
```

All services should show as "Up" or "healthy".

### Test the API:

```bash
# Health check
curl http://localhost:8094/api/admin/health

# Should return a JSON response
```

### Check application logs:

```bash
docker-compose logs backoffice-app | tail -50
```

Look for:
- "All migrations completed successfully"
- "ClickHouse client initialized successfully"
- "Server listening on port : 8094"

## Troubleshooting

### Database Connection Issues

```bash
# Test PostgreSQL connection
docker exec tucanbit-db psql -U tucanbit -d tucanbit -c "SELECT version();"

# Test ClickHouse connection
docker exec tucanbit-clickhouse clickhouse-client --password=tucanbit_clickhouse_password --query "SELECT version()"
```

### Migration Issues

If migrations fail:

```bash
# Check migration status
docker exec tucanbit-db psql -U tucanbit -d tucanbit -c "SELECT * FROM schema_migrations;"

# Check for dirty migration state
docker exec tucanbit-db psql -U tucanbit -d tucanbit -c "SELECT version, dirty FROM schema_migrations;"
```

### Reset Everything (Fresh Start)

If you need to start completely fresh:

```bash
# Stop all containers
docker-compose down

# Remove volumes (WARNING: This deletes all data)
docker-compose down -v

# Start again from Step 4
docker-compose up -d db redis clickhouse zookeeper kafka create-kafka-topic
```

### Port Conflicts

If ports are already in use:

- PostgreSQL: 5433 (change in docker-compose.yaml if needed)
- Redis: 63790 (change in docker-compose.yaml if needed)
- ClickHouse: 8123, 9000 (change in docker-compose.yaml if needed)
- Application: 8094 (change in docker-compose.yaml if needed)

## Environment Variables

The application uses environment variables defined in `docker-compose.yaml`. Key variables:

- `DB_URL`: PostgreSQL connection string
- `CLICKHOUSE_HOST`: ClickHouse hostname
- `CLICKHOUSE_DATABASE`: ClickHouse database name
- `JWT_SECRET`: JWT signing secret
- `KAFKA_BOOTSTRAP_SERVER`: Kafka broker address

## Next Steps

1. **Configure your IDE** to connect to the local database:
   - Host: `localhost`
   - Port: `5433`
   - Database: `tucanbit`
   - Username: `tucanbit`
   - Password: `5kj0YmV5FKKpU9D50B7yH5A`

2. **Access ClickHouse**:
   - HTTP: `http://localhost:8123`
   - Native: `localhost:9000`
   - Username: `tucanbit`
   - Password: `tucanbit_clickhouse_password`

3. **Start developing!** The application will hot-reload on code changes if you're running it locally (outside Docker).

## Additional Resources

- Migration files are in: `./migrations/`
- Docker entrypoint script: `./scripts/docker-entrypoint.sh`
- Local migration script: `./migrate.sh` (for manual migrations)

## Getting Help

If you encounter issues:
1. Check the logs: `docker-compose logs [service-name]`
2. Verify all services are healthy: `docker-compose ps`
3. Check database connectivity using the troubleshooting commands above
4. Review the migration logs in the backoffice-app container
