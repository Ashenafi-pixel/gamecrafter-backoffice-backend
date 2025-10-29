#!/bin/bash

# TucanBIT Database Migration Script
# This script helps you sync with team database changes

set -e

# Configuration
DB_HOST="localhost"
DB_PORT="5433"
DB_NAME="tucanbit"
DB_USER="tucanbit"
DB_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"
MIGRATIONS_DIR="./migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo -e "${BLUE}🔄 TucanBIT Database Migration Tool${NC}"
echo "=================================="

# Check if migrate tool is installed
check_migrate_tool() {
    if ! command -v migrate &> /dev/null; then
        echo -e "${YELLOW}⚠️  golang-migrate tool not found. Installing...${NC}"
        
        # Detect OS and install migrate
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            if command -v brew &> /dev/null; then
                brew install golang-migrate
            else
                echo -e "${RED}❌ Homebrew not found. Please install golang-migrate manually:${NC}"
                echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
                exit 1
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
            sudo mv migrate /usr/local/bin/
        else
            echo -e "${RED}❌ Unsupported OS. Please install golang-migrate manually.${NC}"
            exit 1
        fi
        
        echo -e "${GREEN}✅ golang-migrate installed successfully!${NC}"
    fi
}

# Check database connection
check_database_connection() {
    echo -e "${BLUE}🔍 Checking database connection...${NC}"
    
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${RED}❌ Cannot connect to database!${NC}"
        echo -e "${YELLOW}💡 Make sure:${NC}"
        echo "   1. SSH tunnel is running: ssh -fN -L 5433:localhost:5433 ubuntu@13.48.56.1317 -i ~/Developer/Upwork/Tucanbit/Tucanbit/TucanBIT.pem"
        echo "   2. Database is accessible on localhost:5433"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Database connection successful!${NC}"
}

# Show current migration status
show_status() {
    echo -e "${BLUE}📊 Current Migration Status:${NC}"
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
    echo ""
}

# Run migrations up
migrate_up() {
    echo -e "${BLUE}⬆️  Running migrations up...${NC}"
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
    echo -e "${GREEN}✅ Migrations completed!${NC}"
}

# Run migrations down
migrate_down() {
    echo -e "${YELLOW}⬇️  Running migrations down...${NC}"
    read -p "How many steps down? (default: 1): " steps
    steps=${steps:-1}
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down "$steps"
    echo -e "${GREEN}✅ Migrations rolled back!${NC}"
}

# Force migration version
force_version() {
    echo -e "${YELLOW}🔧 Force migration version...${NC}"
    read -p "Enter version number: " version
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" force "$version"
    echo -e "${GREEN}✅ Migration version forced to $version${NC}"
}

# Create new migration
create_migration() {
    echo -e "${BLUE}📝 Create new migration...${NC}"
    read -p "Enter migration name: " name
    migrate create -ext sql -dir "$MIGRATIONS_DIR" -seq "$name"
    echo -e "${GREEN}✅ Migration files created!${NC}"
    echo -e "${YELLOW}💡 Edit the .up.sql and .down.sql files in $MIGRATIONS_DIR${NC}"
}

# Show help
show_help() {
    echo -e "${BLUE}📖 Available Commands:${NC}"
    echo "  status     - Show current migration status"
    echo "  up         - Run all pending migrations"
    echo "  down       - Rollback migrations"
    echo "  force      - Force migration version"
    echo "  create     - Create new migration"
    echo "  help       - Show this help"
    echo ""
    echo -e "${YELLOW}💡 Team Sync Workflow:${NC}"
    echo "  1. When team makes DB changes, they'll create migration files"
    echo "  2. Pull latest code: git pull"
    echo "  3. Run migrations: ./migrate.sh up"
    echo "  4. Your local DB will be synced with team changes"
}

# Main script logic
main() {
    check_migrate_tool
    check_database_connection
    
    case "${1:-help}" in
        "status")
            show_status
            ;;
        "up")
            show_status
            migrate_up
            show_status
            ;;
        "down")
            show_status
            migrate_down
            show_status
            ;;
        "force")
            show_status
            force_version
            show_status
            ;;
        "create")
            create_migration
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Run main function with all arguments
main "$@"
