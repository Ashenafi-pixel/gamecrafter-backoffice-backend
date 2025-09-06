#!/bin/bash
# TucanBIT Database Export Script for Server
# Run this script on your Ubuntu server where PostgreSQL is running

# Set database connection parameters
DB_NAME="tucanbit"
DB_USER="tucanbit"
DB_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"
DB_HOST="localhost"
DB_PORT="5432"

# Create backup directory
mkdir -p /opt/tucanbit/backups
BACKUP_DIR="/opt/tucanbit/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo "Starting TucanBIT database export on server..."

# Set PGPASSWORD environment variable
export PGPASSWORD="$DB_PASSWORD"

# Export database schema only
echo "Exporting database schema..."
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --schema-only \
    --no-owner \
    --no-privileges \
    --file="$BACKUP_DIR/tucanbit_schema_$TIMESTAMP.sql"

# Export database data only
echo "Exporting database data..."
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --data-only \
    --no-owner \
    --no-privileges \
    --file="$BACKUP_DIR/tucanbit_data_$TIMESTAMP.sql"

# Export complete database (schema + data)
echo "Exporting complete database..."
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --no-owner \
    --no-privileges \
    --file="$BACKUP_DIR/tucanbit_complete_$TIMESTAMP.sql"

# Export as custom format (recommended for large databases)
echo "Exporting database in custom format..."
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --format=custom \
    --no-owner \
    --no-privileges \
    --file="$BACKUP_DIR/tucanbit_custom_$TIMESTAMP.dump"

# Create a compressed archive
echo "Creating compressed archive..."
tar -czf "$BACKUP_DIR/tucanbit_backup_$TIMESTAMP.tar.gz" -C "$BACKUP_DIR" \
    tucanbit_schema_$TIMESTAMP.sql \
    tucanbit_data_$TIMESTAMP.sql \
    tucanbit_complete_$TIMESTAMP.sql \
    tucanbit_custom_$TIMESTAMP.dump

echo "Database export completed successfully!"
echo "Backup files created in: $BACKUP_DIR"
echo "Files created:"
ls -la "$BACKUP_DIR/tucanbit_*_$TIMESTAMP.*"

echo ""
echo "To download the backup to your local machine:"
echo "scp -i TucanBIT.pem ubuntu@51.21.181.162:$BACKUP_DIR/tucanbit_backup_$TIMESTAMP.tar.gz ."
echo ""
echo "To restore the database later, use:"
echo "pg_restore -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME $BACKUP_DIR/tucanbit_custom_$TIMESTAMP.dump"
echo "or"
echo "psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < $BACKUP_DIR/tucanbit_complete_$TIMESTAMP.sql"
