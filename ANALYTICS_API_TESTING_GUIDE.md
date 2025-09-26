# Analytics API Testing Guide for TucanBIT Casino

## 🎯 Overview
This guide provides comprehensive testing instructions for the ClickHouse Analytics APIs that have been implemented for the TucanBIT Casino platform.

## 🚀 Available Analytics APIs

### Base URL
All analytics APIs are available under: `http://localhost:8080/api/analytics/`

### Authentication
Most endpoints require authentication. Use the JWT token from login:
```bash
Authorization: Bearer <your_jwt_token>
```

## 📊 Available Endpoints

### 1. **User Transactions**
**GET** `/analytics/users/{user_id}/transactions`

**Description**: Retrieve user transactions with optional filters

**Parameters**:
- `user_id` (path): User UUID
- `date_from` (query): Start date (RFC3339 format)
- `date_to` (query): End date (RFC3339 format)
- `transaction_type` (query): Transaction type filter
- `game_id` (query): Game ID filter
- `status` (query): Transaction status filter
- `limit` (query): Limit results (default: 100)
- `offset` (query): Offset results (default: 0)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/transactions?limit=10&date_from=2025-01-01T00:00:00Z"
```

### 2. **User Analytics**
**GET** `/analytics/users/{user_id}/analytics`

**Description**: Get comprehensive analytics for a specific user

**Parameters**:
- `user_id` (path): User UUID
- `date_from` (query): Start date (RFC3339 format)
- `date_to` (query): End date (RFC3339 format)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/analytics"
```

### 3. **User Balance History**
**GET** `/analytics/users/{user_id}/balance-history`

**Description**: Get user balance history for the last N hours

**Parameters**:
- `user_id` (path): User UUID
- `hours` (query): Number of hours to look back (default: 24)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/balance-history?hours=48"
```

### 4. **Real-time Stats**
**GET** `/analytics/realtime/stats`

**Description**: Get real-time casino statistics (last hour)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/realtime/stats"
```

### 5. **Daily Report**
**GET** `/analytics/reports/daily`

**Description**: Get daily casino report

**Parameters**:
- `date` (query): Date in YYYY-MM-DD format (default: today)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/reports/daily?date=2025-01-24"
```

### 6. **Top Games**
**GET** `/analytics/reports/top-games`

**Description**: Get top performing games

**Parameters**:
- `limit` (query): Number of games to return (default: 10)
- `date_from` (query): Start date (RFC3339 format)
- `date_to` (query): End date (RFC3339 format)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/reports/top-games?limit=5"
```

### 7. **Top Players**
**GET** `/analytics/reports/top-players`

**Description**: Get top players by betting volume

**Parameters**:
- `limit` (query): Number of players to return (default: 10)
- `date_from` (query): Start date (RFC3339 format)
- `date_to` (query): End date (RFC3339 format)

**Example Request**:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/reports/top-players?limit=5"
```

## 🧪 Testing Steps

### Step 1: Start the Application
```bash
# Make sure ClickHouse is running
docker-compose -f docker-compose.clickhouse.yaml up -d

# Start the application
./tucanbit
```

### Step 2: Get Authentication Token
```bash
# Login to get JWT token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ashenafi_alemu",
    "password": "your_password"
  }'
```

### Step 3: Test Each Endpoint

#### Test Real-time Stats (No data required)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/realtime/stats"
```

**Expected Response**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_transactions": 0,
    "deposits_count": 0,
    "withdrawals_count": 0,
    "bets_count": 0,
    "wins_count": 0,
    "total_deposits": "0.00000000",
    "total_withdrawals": "0.00000000",
    "total_bets": "0.00000000",
    "total_wins": "0.00000000",
    "active_users": 0,
    "active_games": 0,
    "net_revenue": "0.00000000",
    "timestamp": "2025-01-24T14:30:00Z"
  }
}
```

#### Test User Analytics
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/users/a5e168fb-168e-4183-84c5-d49038ce00b5/analytics"
```

#### Test Daily Report
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/analytics/reports/daily"
```

## 📝 Sample Data for Testing

### Insert Test Transaction
```bash
# Insert a test transaction directly into ClickHouse
docker exec tucanbit-clickhouse clickhouse-client --query "
INSERT INTO tucanbit_analytics.transactions (
    id, user_id, transaction_type, amount, currency, status,
    balance_before, balance_after, payment_method, external_transaction_id,
    created_at, updated_at
) VALUES (
    'test-tx-1', 'a5e168fb-168e-4183-84c5-d49038ce00b5', 'deposit', 100.00, 'USD', 'completed',
    0, 100, 'credit_card', 'ext-tx-1', now(), now()
)"
```

### Insert Test Bet Transaction
```bash
docker exec tucanbit-clickhouse clickhouse-client --query "
INSERT INTO tucanbit_analytics.transactions (
    id, user_id, transaction_type, amount, currency, status,
    game_id, game_name, provider, bet_amount, net_result,
    balance_before, balance_after, payment_method, external_transaction_id,
    created_at, updated_at
) VALUES (
    'test-bet-1', 'a5e168fb-168e-4183-84c5-d49038ce00b5', 'bet', 10.00, 'USD', 'completed',
    'game-1', 'Test Game', 'Test Provider', 10.00, -10.00,
    100, 90, 'balance', 'bet-tx-1', now(), now()
)"
```

## 🔍 Response Format

All analytics APIs return responses in this format:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // Analytics data here
  },
  "meta": {
    "total": 0,
    "page": 1,
    "limit": 100,
    "offset": 0
  }
}
```

## 🚨 Error Handling

### Common Error Responses

**401 Unauthorized**:
```json
{
  "code": 401,
  "message": "unauthorized",
  "data": null
}
```

**400 Bad Request**:
```json
{
  "code": 400,
  "message": "invalid user ID format",
  "data": null
}
```

**500 Internal Server Error**:
```json
{
  "code": 500,
  "message": "internal server error",
  "data": null
}
```

## 🎯 What We Need Additionally

### 1. **Authentication Middleware**
- Ensure all analytics endpoints are properly protected
- Add role-based access control for sensitive analytics

### 2. **Rate Limiting**
- Implement rate limiting for analytics endpoints
- Prevent abuse of expensive analytical queries

### 3. **Caching**
- Add Redis caching for frequently accessed analytics
- Cache real-time stats for better performance

### 4. **Data Validation**
- Add input validation for date ranges
- Validate user permissions for user-specific analytics

### 5. **Monitoring**
- Add metrics for analytics API usage
- Monitor query performance

### 6. **Documentation**
- Generate Swagger documentation
- Add API versioning

## 🧪 Testing Checklist

- [ ] ClickHouse is running and accessible
- [ ] Application builds and starts successfully
- [ ] Authentication works with JWT tokens
- [ ] All analytics endpoints respond correctly
- [ ] Error handling works for invalid requests
- [ ] Data filtering works correctly
- [ ] Pagination works for large datasets
- [ ] Real-time stats update correctly
- [ ] User-specific analytics respect permissions

## 🚀 Next Steps

1. **Test with Real Data**: Run the migration script to populate ClickHouse with real data
2. **Performance Testing**: Test with large datasets
3. **Integration Testing**: Test with frontend integration
4. **Load Testing**: Test under high concurrent load
5. **Security Testing**: Test authentication and authorization

## 📞 Support

If you encounter issues:
1. Check ClickHouse logs: `docker logs tucanbit-clickhouse`
2. Check application logs: `tail -f tucanbit.log`
3. Verify ClickHouse connection: `docker exec tucanbit-clickhouse clickhouse-client --query "SELECT 1"`
4. Check database schema: `docker exec tucanbit-clickhouse clickhouse-client --query "SHOW TABLES FROM tucanbit_analytics"`