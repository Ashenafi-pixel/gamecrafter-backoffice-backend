# 🚀 Complete ClickHouse Migration Guide for TucanBIT Casino

This guide provides step-by-step instructions to migrate all your casino data from PostgreSQL to ClickHouse and set up real-time synchronization.

## 📋 What We're Migrating

### ✅ **All Registration Info**
- User registrations with timestamps
- Registration methods (email, phone, social)
- User profile data
- Verification status

### ✅ **All Session Info** 
- Game session start/end times
- Session duration
- Device information (mobile, desktop)
- IP addresses and user agents
- Session outcomes

### ✅ **All Bets/Wagering Info**
- Bet amounts and timestamps
- Game IDs and names
- Player IDs
- Win/loss amounts
- Round IDs
- Balance before/after each transaction
- GrooveTech transactions

## 🛠️ Step 1: Start ClickHouse

```bash
# Start ClickHouse container
docker-compose -f docker-compose.clickhouse.yaml up -d

# Wait for ClickHouse to be ready
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 'Ready!'"
```

## 📊 Step 2: Initialize Schema

```bash
# Initialize ClickHouse schema
docker exec -i tucanbit-clickhouse clickhouse-client --multiquery < clickhouse/schema.sql
```

## 🔄 Step 3: Migrate Historical Data

```bash
# Run the migration script
./scripts/migrate_to_clickhouse.sh
```

This script will:
- ✅ Export all users from PostgreSQL
- ✅ Export all balance transactions
- ✅ Export all GrooveTech transactions
- ✅ Create balance snapshots
- ✅ Generate user analytics
- ✅ Import everything to ClickHouse

## 🔌 Step 4: Integrate Real-Time Sync

Add these hooks to your existing services:

### In User Registration Service:
```go
// After successful user registration
analyticsIntegration.OnUserRegistration(ctx, userID, map[string]interface{}{
    "registration_method": "email",
    "email": user.Email,
    "phone": user.PhoneNumber,
})
```

### In Balance Service:
```go
// After balance changes
analyticsIntegration.OnBalanceChange(ctx, userID, amount, "deposit", transactionID)
```

### In GrooveTech Service:
```go
// After GrooveTech transactions
analyticsIntegration.OnGrooveTransaction(ctx, grooveTx, "wager")
```

### In Game Service:
```go
// When game session starts
analyticsIntegration.OnGameSessionStart(ctx, userID, gameID, sessionID)

// When bet is placed
analyticsIntegration.OnBetPlaced(ctx, userID, gameID, sessionID, betAmount, roundID)

// When player wins
analyticsIntegration.OnWin(ctx, userID, gameID, sessionID, winAmount, roundID)

// When session ends
analyticsIntegration.OnGameSessionEnd(ctx, userID, gameID, sessionID, duration)
```

## 🎯 Step 5: Start Real-Time Sync Service

```go
// In your main application initialization
realtimeSync := analytics.NewRealtimeSyncService(syncService, analyticsStorage, logger)
realtimeSync.StartRealtimeSync(ctx)
```

## 📈 Step 6: Verify Data Migration

```bash
# Check transaction count
docker exec tucanbit-clickhouse clickhouse-client --query "
SELECT 
    transaction_type,
    count(*) as count,
    sum(amount) as total_amount
FROM tucanbit_analytics.transactions 
GROUP BY transaction_type"

# Check user analytics
docker exec tucanbit-clickhouse clickhouse-client --query "
SELECT 
    count(*) as total_users,
    sum(total_deposits) as total_deposits,
    sum(total_bets) as total_bets,
    sum(total_wins) as total_wins
FROM tucanbit_analytics.user_analytics"

# Check balance snapshots
docker exec tucanbit-clickhouse clickhouse-client --query "
SELECT 
    count(*) as snapshots,
    min(snapshot_time) as earliest,
    max(snapshot_time) as latest
FROM tucanbit_analytics.balance_snapshots"
```

## 🔍 Step 7: Test Analytics APIs

```bash
# Test real-time stats
curl "http://localhost:8080/analytics/realtime"

# Test user transactions
curl "http://localhost:8080/analytics/users/{user_id}/transactions?limit=10"

# Test user analytics
curl "http://localhost:8080/analytics/users/{user_id}/analytics"

# Test daily report
curl "http://localhost:8080/analytics/reports/daily?date=2024-01-15"
```

## 📊 Data Coverage Verification

### Registration Data ✅
```sql
SELECT 
    count(*) as registrations,
    count(DISTINCT user_id) as unique_users,
    min(created_at) as first_registration,
    max(created_at) as latest_registration
FROM tucanbit_analytics.transactions 
WHERE transaction_type = 'registration'
```

### Session Data ✅
```sql
SELECT 
    count(DISTINCT session_id) as total_sessions,
    count(DISTINCT user_id) as users_with_sessions,
    avg(duration_seconds) as avg_session_duration
FROM tucanbit_analytics.session_analytics
```

### Betting Data ✅
```sql
SELECT 
    count(*) as total_bets,
    sum(bet_amount) as total_bet_amount,
    sum(win_amount) as total_win_amount,
    count(DISTINCT user_id) as active_players,
    count(DISTINCT game_id) as games_played
FROM tucanbit_analytics.transactions 
WHERE transaction_type IN ('bet', 'groove_bet', 'win', 'groove_win')
```

## 🚨 Troubleshooting

### Migration Issues
```bash
# Check PostgreSQL connection
PGPASSWORD=password psql -h localhost -p 5433 -U postgres -d tucanbit -c "SELECT count(*) FROM users"

# Check ClickHouse connection
docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 1"

# Check migration logs
tail -f /tmp/migration.log
```

### Data Quality Issues
```bash
# Verify data integrity
docker exec tucanbit-clickhouse clickhouse-client --query "
SELECT 
    'transactions' as table_name,
    count(*) as records,
    min(created_at) as earliest,
    max(created_at) as latest
FROM tucanbit_analytics.transactions
UNION ALL
SELECT 
    'balance_snapshots' as table_name,
    count(*) as records,
    min(snapshot_time) as earliest,
    max(snapshot_time) as latest
FROM tucanbit_analytics.balance_snapshots"
```

## 🎯 Next Steps After Migration

1. **Build Analytics Dashboard** - Use the analytics APIs to create real-time dashboards
2. **Set Up Monitoring** - Monitor ClickHouse performance and sync status
3. **Optimize Queries** - Add indexes and optimize frequently used queries
4. **Data Retention** - Set up automated data retention policies
5. **Backup Strategy** - Implement ClickHouse backup procedures

## 📞 Support

If you encounter any issues:
- Check ClickHouse logs: `docker logs tucanbit-clickhouse`
- Check migration logs: `tail -f /tmp/migration.log`
- Verify data: Use the verification queries above
- Test APIs: Use the test endpoints provided

---

**🎰 Ou complete casino analytics system is ready! All registration, session, and betting data will be automatically synced to ClickHouse for powerful analytics and reporting.**