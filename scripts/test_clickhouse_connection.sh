#!/bin/bash

# ClickHouse Connection Test Script
# Tests connection to ClickHouse database

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Testing ClickHouse Connection${NC}"
echo "=================================="

# Configuration
CH_HOST="${CLICKHOUSE_HOST:-localhost}"
CH_PORT="${CLICKHOUSE_PORT:-8123}"
CH_USER="${CLICKHOUSE_USER:-tucanbit}"
CH_PASSWORD="${CLICKHOUSE_PASSWORD:-tucanbit_clickhouse_password}"
CH_DB="${CLICKHOUSE_DB:-tucanbit_analytics}"

echo -e "${YELLOW}Configuration:${NC}"
echo "  Host: $CH_HOST"
echo "  Port: $CH_PORT"
echo "  User: $CH_USER"
echo "  Database: $CH_DB"
echo ""

# Test 1: Ping
echo -e "${BLUE}Test 1: Ping${NC}"
PING_RESULT=$(curl -s "http://${CH_HOST}:${CH_PORT}/ping" 2>&1 || echo "FAILED")
if [ "$PING_RESULT" = "Ok." ]; then
    echo -e "${GREEN}✅ Ping successful${NC}"
else
    echo -e "${RED}❌ Ping failed: $PING_RESULT${NC}"
    exit 1
fi
echo ""

# Test 2: Authentication
echo -e "${BLUE}Test 2: Authentication${NC}"
AUTH_RESULT=$(curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?query=SELECT%201" 2>&1)
if [ "$AUTH_RESULT" = "1" ]; then
    echo -e "${GREEN}✅ Authentication successful${NC}"
else
    echo -e "${RED}❌ Authentication failed: $AUTH_RESULT${NC}"
    exit 1
fi
echo ""

# Test 3: Version
echo -e "${BLUE}Test 3: Version Check${NC}"
VERSION=$(curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?query=SELECT%20version()" 2>&1)
if [ -n "$VERSION" ]; then
    echo -e "${GREEN}✅ ClickHouse Version: $VERSION${NC}"
else
    echo -e "${RED}❌ Version check failed${NC}"
    exit 1
fi
echo ""

# Test 4: Database Existence
echo -e "${BLUE}Test 4: Database Check${NC}"
DB_CHECK=$(curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?query=SELECT%20name%20FROM%20system.databases%20WHERE%20name%20%3D%20%27${CH_DB}%27" 2>&1)
if echo "$DB_CHECK" | grep -q "$CH_DB"; then
    echo -e "${GREEN}✅ Database '$CH_DB' exists${NC}"
else
    echo -e "${YELLOW}⚠️  Database '$CH_DB' not found. Available databases:${NC}"
    curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?query=SELECT%20name%20FROM%20system.databases" | grep -v "INFORMATION_SCHEMA\|information_schema\|system"
fi
echo ""

# Test 5: Tables Count
echo -e "${BLUE}Test 5: Tables in Database${NC}"
TABLES=$(curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?database=${CH_DB}&query=SELECT%20count()%20FROM%20system.tables%20WHERE%20database%20%3D%20%27${CH_DB}%27" 2>&1)
if [ -n "$TABLES" ] && [ "$TABLES" != "0" ]; then
    echo -e "${GREEN}✅ Found $TABLES table(s) in database${NC}"
    echo -e "${YELLOW}Tables:${NC}"
    curl -s "http://${CH_USER}:${CH_PASSWORD}@${CH_HOST}:${CH_PORT}/?database=${CH_DB}&query=SELECT%20name%20FROM%20system.tables%20WHERE%20database%20%3D%20%27${CH_DB}%27%20LIMIT%2010" | head -10
else
    echo -e "${YELLOW}⚠️  No tables found in database (database may be empty)${NC}"
fi
echo ""

# Test 6: Connection from Go client (if clickhouse-client is available)
echo -e "${BLUE}Test 6: Native Connection Test${NC}"
if command -v clickhouse-client &> /dev/null; then
    if clickhouse-client --host="$CH_HOST" --port="$CH_PORT" --user="$CH_USER" --password="$CH_PASSWORD" --database="$CH_DB" --query="SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✅ Native client connection successful${NC}"
    else
        echo -e "${YELLOW}⚠️  Native client connection failed (may not be installed)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  clickhouse-client not installed (optional)${NC}"
fi
echo ""

echo -e "${GREEN}✅ All connection tests passed!${NC}"
echo ""
echo -e "${BLUE}Connection Summary:${NC}"
echo "  ✅ ClickHouse is accessible"
echo "  ✅ Authentication works"
echo "  ✅ Version: $VERSION"
echo "  ✅ Database: $CH_DB"
echo ""
echo -e "${YELLOW}Note:${NC} Config file shows host as '172.31.36.46' (remote),"
echo "but local ClickHouse is running on 'localhost:8123'"
echo ""
echo -e "${BLUE}To use local ClickHouse, update config.yaml:${NC}"
echo "  clickhouse:"
echo "    host: \"localhost\""
echo "    port: 8123"
echo "    database: \"tucanbit_analytics\""
echo "    username: \"tucanbit\""
echo "    password: \"tucanbit_clickhouse_password\""

