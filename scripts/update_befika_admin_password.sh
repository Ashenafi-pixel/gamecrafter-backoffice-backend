#!/bin/bash

# Script to update password for befika_admin user in production database
# Usage: ./update_befika_admin_password.sh

SSH_KEY="$HOME/.ssh/smiss/tucanbit-dev.pem"
SSH_HOST="admin@18.231.108.116"
DB_HOST="rds-prod.internal.tucanbit.local"
DB_USER="tucanbit"
DB_NAME="tucanbit"
SQL_FILE="update_befika_admin_password.sql"

echo "Updating password for befika_admin user..."

# Run the SQL script through SSH
ssh -i "$SSH_KEY" "$SSH_HOST" "psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f -" < "$(dirname "$0")/$SQL_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Password updated successfully!"
else
    echo "❌ Failed to update password"
    exit 1
fi

