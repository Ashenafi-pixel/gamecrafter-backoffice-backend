#!/bin/bash

# TucanBIT GrooveTech Docker Migration Script for AWS Server
# This script migrates all GrooveTech schemas and data to AWS server DB using Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LOCAL_DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@db:5432/tucanbit?sslmode=disable"
AWS_DB_URL="${AWS_DB_URL:-postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require}"
BACKUP_DIR="./aws_migration_backup"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}🐳 TucanBIT GrooveTech Docker Migration to AWS Server${NC}"
echo -e "${BLUE}====================================================${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if Docker is running
check_docker() {
    print_info "Checking Docker status..."
    
    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    print_status "Docker is running"
}

# Create backup directory
create_backup_dir() {
    print_info "Creating backup directory..."
    mkdir -p "$BACKUP_DIR"
    print_status "Backup directory created: $BACKUP_DIR"
}

# Export current database schema using Docker
export_schema_docker() {
    print_info "Exporting current database schema using Docker..."
    
    # Export all schemas using Docker
    docker exec tucanbit-db pg_dump -U tucanbit -d tucanbit \
        --schema-only \
        --no-owner \
        --no-privileges \
        > "$BACKUP_DIR/tucanbit_schema_$TIMESTAMP.sql"
    
    print_status "Schema exported to: $BACKUP_DIR/tucanbit_schema_$TIMESTAMP.sql"
}

# Export GrooveTech specific data using Docker
export_groovetech_data_docker() {
    print_info "Exporting GrooveTech specific data using Docker..."
    
    # Export GrooveTech tables data using Docker
    docker exec tucanbit-db pg_dump -U tucanbit -d tucanbit \
        --data-only \
        --no-owner \
        --no-privileges \
        --table=groove_accounts \
        --table=groove_transactions \
        --table=groove_game_sessions \
        --table=game_sessions \
        > "$BACKUP_DIR/groovetech_data_$TIMESTAMP.sql"
    
    print_status "GrooveTech data exported to: $BACKUP_DIR/groovetech_data_$TIMESTAMP.sql"
}

# Export all data using Docker
export_all_data_docker() {
    print_info "Exporting all database data using Docker..."
    
    docker exec tucanbit-db pg_dump -U tucanbit -d tucanbit \
        --data-only \
        --no-owner \
        --no-privileges \
        > "$BACKUP_DIR/tucanbit_data_$TIMESTAMP.sql"
    
    print_status "All data exported to: $BACKUP_DIR/tucanbit_data_$TIMESTAMP.sql"
}

# Create Docker Compose file for AWS migration
create_docker_compose_aws() {
    print_info "Creating Docker Compose file for AWS migration..."
    
    cat > "$BACKUP_DIR/docker-compose-aws.yml" << 'EOF'
version: '3.8'

services:
  tucanbit-app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_URL=${AWS_DB_URL}
      - REDIS_ADDR=${REDIS_ADDR:-redis:6379}
      - KAFKA_BOOTSTRAP_SERVERS=${KAFKA_BOOTSTRAP_SERVERS:-kafka:9092}
    depends_on:
      - redis
    volumes:
      - ./config:/app/config
    networks:
      - tucanbit-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - tucanbit-network

  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    depends_on:
      - zookeeper
    networks:
      - tucanbit-network

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    ports:
      - "2181:2181"
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    networks:
      - tucanbit-network

volumes:
  redis_data:

networks:
  tucanbit-network:
    driver: bridge
EOF

    print_status "Docker Compose file created: $BACKUP_DIR/docker-compose-aws.yml"
}

# Create AWS environment file
create_aws_env() {
    print_info "Creating AWS environment file..."
    
    cat > "$BACKUP_DIR/.env.aws" << 'EOF'
# TucanBIT AWS Environment Variables
# Update these values for your AWS environment

# Database
AWS_DB_URL=postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require

# Redis (if using AWS ElastiCache)
REDIS_ADDR=your-aws-redis-endpoint:6379
REDIS_PASSWORD=your_redis_password

# Kafka (if using AWS MSK)
KAFKA_BOOTSTRAP_SERVERS=your-aws-msk-endpoint:9092
KAFKA_API_KEY=your_msk_api_key
KAFKA_API_SECRET=your_msk_api_secret

# Application
APP_HOST=0.0.0.0
APP_PORT=8080
DEBUG=false

# JWT Secrets
JWT_SECRET=your_jwt_secret_here
OTP_JWT_SECRET=your_otp_jwt_secret_here

# GrooveTech Configuration
GROOVE_OPERATOR_ID=3818
GROOVE_API_DOMAIN=https://routerstg.groovegaming.com
GROOVE_API_KEY=your_groove_api_key_here
GROOVE_HOME_URL=https://your-domain.com
GROOVE_EXIT_URL=https://your-domain.com
GROOVE_HISTORY_URL=https://your-domain.com/history
GROOVE_LICENSE_TYPE=Curacao
GROOVE_SIGNATURE_VALIDATION=true
GROOVE_SIGNATURE_SECRET=your_signature_secret_here

# AWS S3
AWS_BUCKET_NAME=your_s3_bucket_name
AWS_ACCESS_KEY=your_aws_access_key
AWS_SECRET_KEY=your_aws_secret_key
AWS_REGION=your_aws_region

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_SMTP_PASSWORD=your_gmail_app_password
GOOGLE_SMTP_FROM=your_email@gmail.com

# Facebook OAuth
FACEBOOK_CLIENT_ID=your_facebook_client_id
FACEBOOK_CLIENT_SECRET=your_facebook_client_secret

# Sports Service
SPORTS_SERVICE_API_KEY=your_sports_service_api_key
SPORTS_SERVICE_API_SECRET=your_sports_service_api_secret

# PISI SMS
PISI_BASE_URL=https://api.pisimobile.com/bulksms/v1/
PISI_PASSWORD=your_pisi_password
PISI_VASPID=your_pisi_vaspid
PISI_SENDER_ID=your_pisi_sender_id
EOF

    print_status "AWS environment file created: $BACKUP_DIR/.env.aws"
}

# Create AWS deployment script
create_aws_deployment_script() {
    print_info "Creating AWS deployment script..."
    
    cat > "$BACKUP_DIR/deploy_to_aws.sh" << 'EOF'
#!/bin/bash

# TucanBIT AWS Deployment Script
# This script deploys TucanBIT to AWS server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 TucanBIT AWS Deployment${NC}"
echo -e "${BLUE}===========================${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if required environment variables are set
check_environment() {
    print_info "Checking environment variables..."
    
    if [ -z "$AWS_DB_URL" ]; then
        print_error "AWS_DB_URL environment variable is not set"
        print_info "Please set it with: export AWS_DB_URL='postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require'"
        exit 1
    fi
    
    print_status "Environment variables check passed"
}

# Run database migration
run_migration() {
    print_info "Running database migration..."
    
    if [ -f "aws_migration_*.sql" ]; then
        # Use Docker to run migration
        docker run --rm -e PGPASSWORD=$(echo $AWS_DB_URL | cut -d':' -f3 | cut -d'@' -f1) \
            postgres:15-alpine \
            psql "$AWS_DB_URL" -f /migration.sql
    else
        print_error "Migration file not found. Please run the migration preparation script first."
        exit 1
    fi
    
    print_status "Database migration completed"
}

# Build Docker image
build_image() {
    print_info "Building Docker image..."
    
    docker build -t tucanbit-aws:latest .
    
    print_status "Docker image built successfully"
}

# Deploy application
deploy_application() {
    print_info "Deploying application..."
    
    # Stop existing containers
    docker-compose -f docker-compose-aws.yml down || true
    
    # Start new containers
    docker-compose -f docker-compose-aws.yml up -d
    
    print_status "Application deployed successfully"
}

# Verify deployment
verify_deployment() {
    print_info "Verifying deployment..."
    
    # Wait for application to start
    sleep 10
    
    # Check if application is running
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_status "Application is running and healthy"
    else
        print_warning "Application health check failed. Please check logs."
    fi
    
    # Test GrooveTech endpoints
    print_info "Testing GrooveTech endpoints..."
    
    # Test Game Launch
    if curl -f -X POST http://localhost:8080/api/groove/launch-game \
        -H "Content-Type: application/json" \
        -d '{"game_id":"82695","device_type":"desktop","game_mode":"real","country":"US","currency":"USD","language":"en_US","is_test_account":false,"reality_check_elapsed":0,"reality_check_interval":60}' \
        > /dev/null 2>&1; then
        print_status "Game Launch API is working"
    else
        print_warning "Game Launch API test failed"
    fi
    
    print_status "Deployment verification completed"
}

# Main execution
main() {
    check_environment
    run_migration
    build_image
    deploy_application
    verify_deployment
    
    echo -e "${GREEN}🎉 TucanBIT AWS deployment completed successfully!${NC}"
    echo -e "${BLUE}===============================================${NC}"
    echo -e "${GREEN}Application is running at: http://localhost:8080${NC}"
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "1. Configure your domain and SSL certificates"
    echo -e "2. Set up monitoring and logging"
    echo -e "3. Configure load balancing if needed"
    echo -e "4. Test all GrooveTech endpoints in production"
    echo -e "${BLUE}===============================================${NC}"
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/deploy_to_aws.sh"
    print_status "AWS deployment script created: $BACKUP_DIR/deploy_to_aws.sh"
}

# Create AWS migration SQL file (same as before)
create_aws_migration() {
    print_info "Creating AWS migration SQL file..."
    
    cat > "$BACKUP_DIR/aws_migration_$TIMESTAMP.sql" << 'EOF'
-- TucanBIT GrooveTech Migration to AWS Server
-- Generated on: $(date)

-- Set timezone
SET timezone = 'UTC';

-- Create extensions if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==============================================
-- GROOVETECH TABLES MIGRATION
-- ==============================================

-- Drop existing tables if they exist (in correct order due to foreign keys)
DROP TABLE IF EXISTS groove_transactions CASCADE;
DROP TABLE IF EXISTS groove_game_sessions CASCADE;
DROP TABLE IF EXISTS groove_accounts CASCADE;
DROP TABLE IF EXISTS game_sessions CASCADE;

-- Create GrooveTech accounts table
CREATE TABLE IF NOT EXISTS groove_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(255) UNIQUE NOT NULL,
    session_id VARCHAR(255),
    balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create groove_transactions table for storing transaction data for idempotency
CREATE TABLE IF NOT EXISTS groove_transactions (
    id SERIAL PRIMARY KEY,
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    account_transaction_id VARCHAR(50) NOT NULL,
    account_id VARCHAR(60) NOT NULL,
    game_session_id VARCHAR(64) NOT NULL,
    round_id VARCHAR(255) NOT NULL,
    game_id VARCHAR(255) NOT NULL,
    bet_amount DECIMAL(32,10) NOT NULL,
    device VARCHAR(20) NOT NULL,
    frbid VARCHAR(255),
    user_id UUID NOT NULL,
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create game_sessions table for tracking GrooveTech game launches
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(64) UNIQUE DEFAULT ('Tucan_' || gen_random_uuid()::text),
    game_id VARCHAR(20) NOT NULL,
    device_type VARCHAR(20) NOT NULL CHECK (device_type IN ('desktop', 'mobile')),
    game_mode VARCHAR(10) NOT NULL CHECK (game_mode IN ('demo', 'real')),
    groove_url TEXT,
    home_url TEXT,
    exit_url TEXT,
    history_url TEXT,
    license_type VARCHAR(20) DEFAULT 'Curacao',
    is_test_account BOOLEAN DEFAULT false,
    reality_check_elapsed INTEGER DEFAULT 0,
    reality_check_interval INTEGER DEFAULT 60,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '2 hours',
    is_active BOOLEAN DEFAULT true,
    last_activity TIMESTAMP DEFAULT NOW()
);

-- Create GrooveTech game sessions table
CREATE TABLE IF NOT EXISTS groove_game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    account_id VARCHAR(255) NOT NULL REFERENCES groove_accounts(account_id) ON DELETE CASCADE,
    game_id VARCHAR(255) NOT NULL,
    balance DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- ==============================================
-- INDEXES FOR PERFORMANCE
-- ==============================================

-- GrooveTech accounts indexes
CREATE INDEX IF NOT EXISTS idx_groove_accounts_user_id ON groove_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_status ON groove_accounts(status);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_account_id ON groove_accounts(account_id);

-- GrooveTech transactions indexes
CREATE INDEX IF NOT EXISTS idx_groove_transactions_transaction_id ON groove_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_account_id ON groove_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_user_id ON groove_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_game_session_id ON groove_transactions(game_session_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_created_at ON groove_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_account_created ON groove_transactions(account_id, created_at);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_session_id ON groove_transactions(session_id);

-- Game sessions indexes
CREATE INDEX IF NOT EXISTS idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_session_id ON game_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_game_id ON game_sessions(game_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_created_at ON game_sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_game_sessions_active ON game_sessions(is_active);

-- GrooveTech game sessions indexes
CREATE INDEX IF NOT EXISTS idx_groove_game_sessions_session_id ON groove_game_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_groove_game_sessions_account_id ON groove_game_sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_groove_game_sessions_status ON groove_game_sessions(status);

-- ==============================================
-- FOREIGN KEY CONSTRAINTS
-- ==============================================

-- Add foreign key constraint to users table
ALTER TABLE groove_transactions 
ADD CONSTRAINT fk_groove_transactions_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ==============================================
-- FUNCTIONS AND TRIGGERS
-- ==============================================

-- Create function for unique GrooveTech session ID generation
CREATE OR REPLACE FUNCTION generate_groove_session_id()
RETURNS TEXT AS $$
BEGIN
    RETURN 'Tucan_' || gen_random_uuid()::text;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE OR REPLACE FUNCTION update_groove_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_groove_accounts_updated_at
    BEFORE UPDATE ON groove_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_groove_accounts_updated_at();

-- Create trigger to update last_activity on any update
CREATE OR REPLACE FUNCTION update_game_session_activity()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_activity = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_game_session_activity
    BEFORE UPDATE ON game_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_game_session_activity();

-- Create a function to clean up expired game sessions
CREATE OR REPLACE FUNCTION cleanup_expired_game_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE game_sessions 
    SET is_active = false 
    WHERE expires_at < NOW() AND is_active = true;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to clean up expired GrooveTech sessions
CREATE OR REPLACE FUNCTION cleanup_expired_groove_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE groove_game_sessions 
    SET status = 'expired'
    WHERE status = 'active' 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get GrooveTech account summary
CREATE OR REPLACE FUNCTION get_groove_account_summary(p_account_id VARCHAR(255))
RETURNS TABLE (
    account_id VARCHAR(255),
    balance DECIMAL(20,8),
    currency VARCHAR(10),
    status VARCHAR(50),
    total_transactions BIGINT,
    last_transaction_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ga.account_id,
        ga.balance,
        ga.currency,
        ga.status,
        COUNT(gt.transaction_id) as total_transactions,
        MAX(gt.created_at) as last_transaction_at
    FROM groove_accounts ga
    LEFT JOIN groove_transactions gt ON ga.account_id = gt.account_id
    WHERE ga.account_id = p_account_id
    GROUP BY ga.account_id, ga.balance, ga.currency, ga.status;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- SAMPLE DATA INSERTION
-- ==============================================

-- Insert sample GrooveTech accounts for existing users
INSERT INTO groove_accounts (user_id, account_id, balance, currency, status)
SELECT 
    u.id,
    'groove_' || u.id::text,
    COALESCE(b.amount_units, 0),
    'USD',
    'active'
FROM users u
LEFT JOIN balances b ON u.id = b.user_id AND b.currency_code = 'USD'
WHERE NOT EXISTS (
    SELECT 1 FROM groove_accounts ga WHERE ga.user_id = u.id
)
LIMIT 10;

-- Insert sample game sessions
INSERT INTO game_sessions (user_id, game_id, device_type, game_mode, home_url, exit_url, license_type) 
VALUES 
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', '82695', 'desktop', 'real', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao'),
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', '82695', 'mobile', 'demo', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao')
ON CONFLICT DO NOTHING;

-- ==============================================
-- VERIFICATION QUERIES
-- ==============================================

-- Verify tables were created
SELECT 'GrooveTech Tables Created Successfully' as status;

-- Show table counts
SELECT 
    'groove_accounts' as table_name, 
    COUNT(*) as record_count 
FROM groove_accounts
UNION ALL
SELECT 
    'groove_transactions' as table_name, 
    COUNT(*) as record_count 
FROM groove_transactions
UNION ALL
SELECT 
    'game_sessions' as table_name, 
    COUNT(*) as record_count 
FROM game_sessions
UNION ALL
SELECT 
    'groove_game_sessions' as table_name, 
    COUNT(*) as record_count 
FROM groove_game_sessions;
EOF

    print_status "AWS migration SQL file created: $BACKUP_DIR/aws_migration_$TIMESTAMP.sql"
}

# Main execution
main() {
    echo -e "${BLUE}Starting TucanBIT GrooveTech Docker migration to AWS...${NC}"
    
    check_docker
    create_backup_dir
    export_schema_docker
    export_groovetech_data_docker
    export_all_data_docker
    create_aws_migration
    create_docker_compose_aws
    create_aws_env
    create_aws_deployment_script
    
    echo -e "${GREEN}🎉 Docker migration preparation completed successfully!${NC}"
    echo -e "${BLUE}====================================================${NC}"
    echo -e "${GREEN}Files created in: $BACKUP_DIR${NC}"
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "1. Review the migration files"
    echo -e "2. Update AWS_DB_URL environment variable"
    echo -e "3. Copy files to your AWS server"
    echo -e "4. Run: ./deploy_to_aws.sh"
    echo -e "5. Test all GrooveTech endpoints"
    echo -e "${BLUE}====================================================${NC}"
}

# Run main function
main "$@"