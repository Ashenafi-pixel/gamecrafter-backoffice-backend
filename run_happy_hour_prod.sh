#!/bin/bash

# Script to add global_rakeback_override table to production database
# Usage: ./run_happy_hour_prod.sh

# SSH connection details
SSH_HOST="tucanbit-prod-bastion"
SSH_USER="admin"
SSH_KEY="$HOME/.ssh/smiss/tucanbit-dev.pem"

# Database connection details
DB_HOST="rds-prod.internal.tucanbit.local"
DB_USER="tucanbit"
DB_NAME="tucanbit"

# SQL file path (relative to script location)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="$SCRIPT_DIR/add_happy_hour_to_prod.sql"

echo "Connecting to production database..."
echo "Host: $SSH_HOST"
echo "Database: $DB_HOST"
echo ""

# Option 1: If you're already on the bastion or have direct access
# Uncomment this line and comment out the SSH version below:
# psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $SQL_FILE

# Option 2: SSH tunnel approach (if needed)
# First, copy the SQL file to the bastion, then execute
scp -i $SSH_KEY $SQL_FILE $SSH_USER@$SSH_HOST:/tmp/add_happy_hour_to_prod.sql
ssh -i $SSH_KEY $SSH_USER@$SSH_HOST "psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f /tmp/add_happy_hour_to_prod.sql"

echo ""
echo "Done! Check the output above for any errors."

