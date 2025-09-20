# TucanBIT Cashback & User Level API Collection for Postman

## 📋 **API Collection Overview**

This document provides all the cashback, user level, and admin APIs with correct request bodies for updating the Postman collection.

## 🔐 **Authentication**
All authenticated endpoints require:
```
Authorization: Bearer {{access_token}}
```

## 📚 **API Endpoints**

### 1. **Public Cashback APIs**

#### 1.1 Get Cashback Tiers
```
GET {{base_url}}/cashback/tiers
```
**Response Example:**
```json
[
  {
    "id": "199dae9d-f784-4e20-9b77-46d7f2112229",
    "tier_name": "Bronze",
    "tier_level": 1,
    "min_ggr_required": "0",
    "cashback_percentage": "0.5",
    "bonus_multiplier": "1",
    "daily_cashback_limit": "50"
  }
]
```

#### 1.2 Get House Edge by Game Type
```
GET {{base_url}}/cashback/house-edge?game_type=groovetech
```
**Response Example:**
```json
{
  "id": "c86a1677-bb0f-4cd2-970d-100350bba41c",
  "game_type": "groovetech",
  "house_edge": "0.02",
  "min_bet": "0.1",
  "max_bet": "1000",
  "is_active": true
}
```

### 2. **User Cashback APIs** (Authentication Required)

#### 2.1 Get User Cashback Summary
```
GET {{base_url}}/user/cashback
Authorization: Bearer {{access_token}}
```
**Response Example:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "current_tier": {
    "tier_name": "Bronze",
    "tier_level": 1,
    "cashback_percentage": "0.5"
  },
  "total_ggr": "2.4",
  "available_cashback": "0",
  "pending_cashback": "0",
  "total_claimed": "0"
}
```

#### 2.2 Claim Cashback
```
POST {{base_url}}/user/cashback/claim
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
  "amount": "5.50"
}
```
**Response Example:**
```json
{
  "claim_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": "5.50",
  "net_amount": "5.50",
  "processing_fee": "0",
  "status": "completed",
  "message": "Cashback claim processed successfully"
}
```

#### 2.3 Get User Cashback Earnings
```
GET {{base_url}}/user/cashback/earnings
Authorization: Bearer {{access_token}}
```

#### 2.4 Get User Cashback Claims
```
GET {{base_url}}/user/cashback/claims
Authorization: Bearer {{access_token}}
```

### 3. **Balance Synchronization APIs** (Authentication Required)

#### 3.1 Validate Balance Synchronization
```
GET {{base_url}}/user/balance/validate-sync
Authorization: Bearer {{access_token}}
```
**Response Example:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "main_balance": "255",
  "groove_balance": "255",
  "is_synchronized": true,
  "discrepancy": "0",
  "last_sync_time": "2025-09-16T23:13:35.226609+03:00",
  "last_validation_time": "2025-09-16T23:18:45.463734749+03:00"
}
```

#### 3.2 Reconcile Balances
```
POST {{base_url}}/user/balance/reconcile
Authorization: Bearer {{access_token}}
```
**Response Example:**
```json
{
  "message": "Balances reconciled successfully",
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5"
}
```

### 4. **Admin Cashback APIs** (Admin Authentication Required)

#### 4.1 Get Cashback Statistics
```
GET {{base_url}}/admin/cashback/stats
Authorization: Bearer {{admin_access_token}}
```

#### 4.2 Create Cashback Tier
```
POST {{base_url}}/admin/cashback/tiers
Authorization: Bearer {{admin_access_token}}
Content-Type: application/json

{
  "tier_name": "Platinum",
  "tier_level": 4,
  "min_ggr_required": "15000",
  "cashback_percentage": "2.0",
  "bonus_multiplier": "1.5",
  "daily_cashback_limit": "500",
  "weekly_cashback_limit": "3000",
  "monthly_cashback_limit": "10000",
  "special_benefits": {
    "priority_support": true,
    "exclusive_games": true
  },
  "is_active": true
}
```

#### 4.3 Update Cashback Tier
```
PUT {{base_url}}/admin/cashback/tiers/{{tier_id}}
Authorization: Bearer {{admin_access_token}}
Content-Type: application/json

{
  "tier_name": "Updated Platinum",
  "cashback_percentage": "2.5",
  "daily_cashback_limit": "600"
}
```

#### 4.4 Create Cashback Promotion
```
POST {{base_url}}/admin/cashback/promotions
Authorization: Bearer {{admin_access_token}}
Content-Type: application/json

{
  "promotion_name": "Weekend Boost",
  "description": "Double cashback on weekends",
  "boost_multiplier": "2.0",
  "start_date": "2025-09-20T00:00:00Z",
  "end_date": "2025-09-22T23:59:59Z",
  "min_tier_level": 1,
  "game_types": ["groovetech", "plinko"],
  "is_active": true
}
```

#### 4.5 Create House Edge Configuration
```
POST {{base_url}}/admin/cashback/house-edge
Authorization: Bearer {{admin_access_token}}
Content-Type: application/json

{
  "game_type": "blackjack",
  "game_variant": "european",
  "house_edge": "0.0048",
  "min_bet": "1.00",
  "max_bet": "1000.00",
  "is_active": true,
  "effective_from": "2025-09-17T00:00:00Z"
}
```

### 5. **Admin Dashboard APIs** (Admin Authentication Required)

#### 5.1 Get Dashboard Statistics
```
GET {{base_url}}/admin/cashback/dashboard
Authorization: Bearer {{admin_access_token}}
```

#### 5.2 Get Cashback Analytics
```
GET {{base_url}}/admin/cashback/dashboard/analytics?start_date=2025-09-01&end_date=2025-09-16&game_type=groovetech
Authorization: Bearer {{admin_access_token}}
```

#### 5.3 Get System Health
```
GET {{base_url}}/admin/cashback/dashboard/health
Authorization: Bearer {{admin_access_token}}
```

#### 5.4 Get User Cashback Details
```
GET {{base_url}}/admin/cashback/dashboard/users/{{user_id}}
Authorization: Bearer {{admin_access_token}}
```

#### 5.5 Process Manual Cashback
```
POST {{base_url}}/admin/cashback/dashboard/manual-cashback
Authorization: Bearer {{admin_access_token}}
Content-Type: application/json

{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "amount": "25.00",
  "reason": "Customer service adjustment",
  "game_type": "groovetech",
  "game_id": "82695"
}
```

## 🔧 **Postman Variables to Add**

```json
{
  "key": "admin_access_token",
  "value": "",
  "type": "string",
  "description": "Admin access token for admin APIs"
},
{
  "key": "tier_id",
  "value": "199dae9d-f784-4e20-9b77-46d7f2112229",
  "type": "string",
  "description": "Cashback tier ID for updates"
},
{
  "key": "claim_amount",
  "value": "5.50",
  "type": "string",
  "description": "Amount to claim for cashback"
}
```

## 🧪 **Test Scenarios**

### Complete Cashback Flow Test:
1. **Login** → Get access token
2. **Launch Game** → Get session ID
3. **Place Wager** → $25 bet
4. **Process Result** → $0 result (player loses)
5. **Check Cashback Summary** → Verify GGR increase
6. **Validate Balance Sync** → Confirm balances are synchronized
7. **Check House Edge** → Verify game-specific house edge
8. **Claim Cashback** → Claim available cashback
9. **Reconcile Balances** → Ensure balance consistency

### Admin Dashboard Test:
1. **Login as Admin** → Get admin access token
2. **Get Dashboard Stats** → View overall statistics
3. **Check System Health** → Verify system status
4. **View User Details** → Check specific user cashback
5. **Create Manual Cashback** → Process manual adjustment
6. **Update House Edge** → Modify game configuration

## 📝 **Notes**

- All amounts are in USD
- Timestamps are in ISO 8601 format
- User IDs are UUIDs
- House edge is represented as decimal (0.02 = 2%)
- Cashback percentage is decimal (0.5 = 0.5%)
- All endpoints return proper error responses with status codes