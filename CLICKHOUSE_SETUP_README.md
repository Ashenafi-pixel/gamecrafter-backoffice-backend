# 🎯 ClickHouse Analytics Setup for TucanBIT Casino

This document provides a comprehensive guide to setting up ClickHouse as a data warehouse backbone for your casino site, enabling powerful analytics and reporting without impacting your main PostgreSQL database.

## 📋 Overview

ClickHouse is perfect for casino analytics because it's designed for:
- **Fast aggregations** on large datasets
- **Time-series data** analysis
- **Complex analytical queries** without affecting OLTP performance
- **Real-time analytics** and reporting

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   ClickHouse    │    │   Frontend      │
│   (OLTP)        │───▶│   (OLAP)        │◀───│   Analytics     │
│   Main DB       │    │   Analytics     │    │   Dashboard     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### 1. Run the Setup Script

```bash
# Make the script executable and run it
chmod +x scripts/setup_clickhouse.sh
./scripts/setup_clickhouse.sh
```

This script will:
- ✅ Start ClickHouse with Docker
- ✅ Initialize the database schema
- ✅ Create all necessary tables and indexes
- ✅ Test the connection

### 2. Verify Installation

```bash
# Check if ClickHouse is running
docker ps | grep clickhouse

# Test connection
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 'Hello ClickHouse!' as message"
```

## 📊 Database Schema

### Core Tables

#### 1. **transactions** - Main transaction table
```sql
- id: String (Transaction ID)
- user_id: String (User UUID)
- transaction_type: Enum (deposit, withdrawal, bet, win, bonus, cashback, etc.)
- amount: Decimal(20,8) (Transaction amount)
- currency: String (Currency code)
- status: Enum (pending, completed, failed, cancelled)
- game_id: String (Game identifier)
- game_name: String (Game name)
- provider: String (Game provider)
- session_id: String (Game session)
- balance_before/after: Decimal(20,8) (Balance snapshots)
- created_at: DateTime (Timestamp)
```

#### 2. **user_analytics** - User performance metrics
```sql
- user_id: String
- date: Date
- total_deposits/withdrawals/bets/wins: Decimal(20,8)
- transaction_count: UInt32
- unique_games_played: UInt16
- avg_bet_amount: Decimal(20,8)
- last_activity: DateTime
```

#### 3. **game_analytics** - Game performance metrics
```sql
- game_id: String
- game_name: String
- provider: String
- total_bets/wins: Decimal(20,8)
- total_players/sessions: UInt32
- rtp: Decimal(5,4) (Return to Player %)
- volatility: Enum (low, medium, high)
```

#### 4. **session_analytics** - Session tracking
```sql
- session_id: String
- user_id: String
- start_time/end_time: DateTime
- duration_seconds: UInt32
- total_bets/wins: Decimal(20,8)
- device_type: String
- ip_address: String
```

## 🔌 API Endpoints

### User Analytics
```http
GET /analytics/users/{user_id}/transactions
GET /analytics/users/{user_id}/analytics
GET /analytics/users/{user_id}/balance-history
```

### Real-time Statistics
```http
GET /analytics/realtime
```

### Reports
```http
GET /analytics/reports/daily?date=2024-01-15
GET /analytics/games/top?limit=10
GET /analytics/players/top?limit=10
```

### Query Parameters
- `date_from`: Start date (RFC3339)
- `date_to`: End date (RFC3339)
- `transaction_type`: Filter by type
- `game_id`: Filter by game
- `limit`: Number of results
- `offset`: Pagination offset

## 📈 Example Queries

### 1. User Transaction History
```bash
curl "http://localhost:8080/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/transactions?limit=50"
```

### 2. Real-time Stats
```bash
curl "http://localhost:8080/analytics/realtime"
```

### 3. Daily Report
```bash
curl "http://localhost:8080/analytics/reports/daily?date=2024-01-15"
```

### 4. Top Games
```bash
curl "http://localhost:8080/analytics/games/top?limit=10&date_from=2024-01-01T00:00:00Z"
```

## 🔄 Data Synchronization

### Automatic Sync Points

The system automatically syncs data to ClickHouse when:

1. **User Transactions** (deposits, withdrawals, bets, wins)
2. **GrooveTech Transactions** (wager, result, rollback)
3. **Balance Changes** (real-time balance snapshots)
4. **Session Events** (game starts, ends, timeouts)

### Manual Sync

```go
// Example: Sync a transaction
syncService.SyncTransaction(ctx, &dto.AnalyticsTransaction{
    ID: "tx_123",
    UserID: userID,
    TransactionType: "deposit",
    Amount: decimal.NewFromFloat(100.0),
    // ... other fields
})
```

## 🛠️ Configuration

### Environment Variables

Add to your `.env` or config:

```yaml
clickhouse:
  host: "localhost"
  port: 9000
  database: "tucanbit_analytics"
  username: "tucanbit"
  password: "tucanbit_clickhouse_password"
  timeout: "30s"
```

### Docker Compose

The ClickHouse service is defined in `docker-compose.clickhouse.yaml`:

```yaml
services:
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    ports:
      - "8123:8123"  # HTTP interface
      - "9000:9000"  # Native TCP interface
    environment:
      CLICKHOUSE_DB: tucanbit_analytics
      CLICKHOUSE_USER: tucanbit
      CLICKHOUSE_PASSWORD: tucanbit_clickhouse_password
```

## 📊 Performance Optimization

### Indexes
- **Primary**: `(user_id, created_at, transaction_type)`
- **Secondary**: `(game_id, date)`, `(transaction_type, date)`
- **Covering**: `(user_id, start_time)` for sessions

### Partitioning
- **Monthly partitions** by `created_at` date
- **Automatic partition management** for data retention

### Materialized Views
- **Real-time aggregations** for user stats
- **Pre-computed metrics** for faster queries

## 🔍 Monitoring

### Health Check
```bash
# Check ClickHouse status
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 'OK' as status"

# View system metrics
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT * FROM system.metrics LIMIT 10"
```

### Logs
```bash
# View ClickHouse logs
docker logs tucanbit-clickhouse

# View application logs
tail -f tucanbit.log | grep -i clickhouse
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Connection Refused
```bash
# Check if ClickHouse is running
docker ps | grep clickhouse

# Restart if needed
docker-compose -f docker-compose.clickhouse.yaml restart
```

#### 2. Schema Errors
```bash
# Reinitialize schema
docker exec -i tucanbit-clickhouse clickhouse-client --multiquery < clickhouse/schema.sql
```

#### 3. Performance Issues
```bash
# Check table sizes
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT table, formatReadableSize(sum(bytes)) as size FROM system.parts GROUP BY table"

# Check query performance
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT * FROM system.query_log ORDER BY event_time DESC LIMIT 5"
```

## 📚 Advanced Usage

### Custom Queries

```sql
-- Top 10 players by net loss
SELECT 
    user_id,
    sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
    sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
    total_bets - total_wins as net_loss
FROM transactions
WHERE created_at >= now() - INTERVAL 30 DAY
GROUP BY user_id
ORDER BY net_loss DESC
LIMIT 10;

-- Game RTP analysis
SELECT 
    game_id,
    game_name,
    sumIf(amount, transaction_type IN ('bet', 'groove_bet')) as total_bets,
    sumIf(amount, transaction_type IN ('win', 'groove_win')) as total_wins,
    (total_wins / total_bets) * 100 as rtp_percentage
FROM transactions
WHERE created_at >= now() - INTERVAL 7 DAY
GROUP BY game_id, game_name
HAVING total_bets > 1000
ORDER BY rtp_percentage DESC;
```

### Data Retention

```sql
-- Delete old data (older than 1 year)
ALTER TABLE transactions DELETE WHERE created_at < now() - INTERVAL 1 YEAR;

-- Optimize tables
OPTIMIZE TABLE transactions FINAL;
```

## 🎯 Next Steps

1. **Start ClickHouse**: Run the setup script
2. **Update Configuration**: Add ClickHouse settings to your app config
3. **Integrate Sync**: Add sync calls to your transaction handlers
4. **Build Dashboard**: Create frontend analytics dashboard
5. **Monitor Performance**: Set up alerts and monitoring

## 📞 Support

For issues or questions:
- Check the logs: `docker logs tucanbit-clickhouse`
- Review the schema: `clickhouse/schema.sql`
- Test queries: Use the ClickHouse client
- Monitor performance: Check system tables

---

**🎰 Your casino analytics backbone is ready! ClickHouse will handle all your analytical workloads while keeping your PostgreSQL database optimized for transactions.**