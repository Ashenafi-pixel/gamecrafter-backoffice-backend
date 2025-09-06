#!/bin/bash
# TucanBIT Database Export Script using Docker
# This script exports the TucanBIT database from Docker container

# Set database connection parameters
DB_NAME="tucanbit"
DB_USER="tucanbit"
DB_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"
CONTAINER_NAME="tucanbit-db"

# Create backup directory
mkdir -p backups
BACKUP_DIR="$(pwd)/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo "Starting TucanBIT database export from Docker container..."

# Export database schema only
echo "Exporting database schema..."
docker exec $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "$BACKUP_DIR/tucanbit_schema_$TIMESTAMP.sql"

# Export database data only
echo "Exporting database data..."
docker exec $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME \
    --data-only \
    --no-owner \
    --no-privileges \
    > "$BACKUP_DIR/tucanbit_data_$TIMESTAMP.sql"

# Export complete database (schema + data)
echo "Exporting complete database..."
docker exec $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME \
    --no-owner \
    --no-privileges \
    > "$BACKUP_DIR/tucanbit_complete_$TIMESTAMP.sql"

# Export as custom format (recommended for large databases)
echo "Exporting database in custom format..."
docker exec $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME \
    --format=custom \
    --no-owner \
    --no-privileges \
    > "$BACKUP_DIR/tucanbit_custom_$TIMESTAMP.dump"

# Export specific tables (if needed)
echo "Exporting specific tables..."
docker exec $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME \
    --table=users \
    --table=games \
    --table=currencies \
    --table=company \
    --no-owner \
    --no-privileges \
    > "$BACKUP_DIR/tucanbit_tables_$TIMESTAMP.sql"

# Create a compressed archive
echo "Creating compressed archive..."
tar -czf "$BACKUP_DIR/tucanbit_backup_$TIMESTAMP.tar.gz" -C "$BACKUP_DIR" \
    tucanbit_schema_$TIMESTAMP.sql \
    tucanbit_data_$TIMESTAMP.sql \
    tucanbit_complete_$TIMESTAMP.sql \
    tucanbit_custom_$TIMESTAMP.dump \
    tucanbit_tables_$TIMESTAMP.sql

echo "Database export completed successfully!"
echo "Backup files created in: $BACKUP_DIR"
echo "Files created:"
ls -la "$BACKUP_DIR/tucanbit_*_$TIMESTAMP.*"

echo ""
echo "To restore the database later, use:"
echo "docker exec -i $CONTAINER_NAME pg_restore -U $DB_USER -d $DB_NAME < $BACKUP_DIR/tucanbit_custom_$TIMESTAMP.dump"
echo "or"
echo "docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < $BACKUP_DIR/tucanbit_complete_$TIMESTAMP.sql"
