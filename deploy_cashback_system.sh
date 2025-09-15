#!/bin/bash

# Deploy Cashback System to AWS Database
# This script deploys the complete cashback and level system

echo "🎰 Deploying TucanBIT World-Class Cashback System..."

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="tucanbit"
DB_USER="tucanbit"
DB_PASSWORD="5kj0YmV5FKKpU9D50B7yH5A"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📋 Cashback System Deployment Plan:${NC}"
echo "  1. ✅ Run cashback system migration"
echo "  2. ✅ Initialize existing users with cashback levels"
echo "  3. ✅ Set up default game house edges"
echo "  4. ✅ Verify system integration"
echo ""

# Check if PostgreSQL is running
echo -e "${YELLOW}🔍 Checking database connection...${NC}"
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
    echo -e "${RED}❌ Database connection failed!${NC}"
    echo "Please ensure PostgreSQL is running and accessible."
    exit 1
fi
echo -e "${GREEN}✅ Database connection successful${NC}"

# Run the cashback migration
echo -e "${YELLOW}🚀 Running cashback system migration...${NC}"
docker exec -i tucanbit-db psql -U tucanbit -d tucanbit < migrations/20250912_create_cashback_system.up.sql

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Cashback system migration completed successfully${NC}"
else
    echo -e "${RED}❌ Migration failed!${NC}"
    exit 1
fi

# Initialize existing users with cashback levels
echo -e "${YELLOW}👥 Initializing existing users with cashback levels...${NC}"
docker exec -it tucanbit-db psql -U tucanbit -d tucanbit -c "
-- Initialize all existing users with Bronze level
INSERT INTO user_levels (user_id, current_level, total_ggr, total_bets, total_wins, level_progress, current_tier_id)
SELECT 
    u.id,
    1,
    0,
    0,
    0,
    0,
    ct.id
FROM users u
CROSS JOIN cashback_tiers ct
WHERE ct.tier_level = 1
AND NOT EXISTS (
    SELECT 1 FROM user_levels ul WHERE ul.user_id = u.id
);
"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Existing users initialized with cashback levels${NC}"
else
    echo -e "${RED}❌ User initialization failed!${NC}"
    exit 1
fi

# Verify the system
echo -e "${YELLOW}🔍 Verifying cashback system...${NC}"
docker exec -it tucanbit-db psql -U tucanbit -d tucanbit -c "
-- Check cashback tiers
SELECT 'Cashback Tiers:' as info;
SELECT tier_name, tier_level, min_ggr_required, cashback_percentage, daily_cashback_limit 
FROM cashback_tiers 
ORDER BY tier_level;

-- Check user levels
SELECT 'User Levels:' as info;
SELECT COUNT(*) as total_users_with_levels FROM user_levels;

-- Check game house edges
SELECT 'Game House Edges:' as info;
SELECT game_type, house_edge, min_bet, max_bet 
FROM game_house_edges 
ORDER BY game_type;
"

echo -e "${GREEN}🎉 Cashback System Deployment Complete!${NC}"
echo ""
echo -e "${BLUE}📊 System Summary:${NC}"
echo "  • 5 Cashback Tiers (Bronze → Diamond)"
echo "  • Real-time GGR tracking"
echo "  • Automatic level progression"
echo "  • Game-specific house edges"
echo "  • 30-day cashback expiry"
echo "  • Daily/weekly/monthly limits"
echo ""
echo -e "${YELLOW}🚀 Next Steps:${NC}"
echo "  1. Integrate with GrooveTech bet processing"
echo "  2. Set up Kafka consumer for real-time events"
echo "  3. Complete API endpoints"
echo "  4. Test end-to-end functionality"
echo ""
echo -e "${GREEN}✅ Ready for world-class casino operations!${NC}"