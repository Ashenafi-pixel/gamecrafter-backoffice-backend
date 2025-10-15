# Withdrawal Viewing API Documentation

This document describes all the API endpoints for viewing and managing withdrawals.

## Base URL
```
http://localhost:8080
```

## Authentication
All endpoints require authentication. Include the admin token in the Authorization header:
```
Authorization: Bearer <your_admin_token>
```

## Endpoints

### 1. Get All Withdrawals
**GET** `/api/v1/withdrawals`

Retrieves all withdrawals with filtering and pagination.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Number of results per page (default: 20, max: 100) |
| `offset` | integer | No | Number of results to skip (default: 0) |
| `status` | string | No | Filter by status (pending, processing, completed, failed, cancelled) |
| `user_id` | UUID | No | Filter by user ID |
| `withdrawal_id` | string | No | Filter by withdrawal ID (partial match) |
| `username` | string | No | Filter by username (partial match) |
| `email` | string | No | Filter by email (partial match) |
| `start_date` | date | No | Filter by start date (YYYY-MM-DD) |
| `end_date` | date | No | Filter by end date (YYYY-MM-DD) |

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals?status=completed&limit=10&offset=0" \
  -H "Authorization: Bearer your_admin_token"
```

#### Example Response
```json
{
  "success": true,
  "data": {
    "withdrawals": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "user_id": "456e7890-e89b-12d3-a456-426614174001",
        "withdrawal_id": "WD123456789",
        "usd_amount_cents": 10000,
        "crypto_amount": "0.001",
        "currency_code": "BTC",
        "status": "completed",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:35:00Z",
        "username": "john_doe",
        "email": "john@example.com",
        "tx_hash": "0x1234567890abcdef..."
      }
    ],
    "pagination": {
      "limit": 10,
      "offset": 0
    }
  }
}
```

### 2. Get Withdrawal Statistics
**GET** `/api/v1/withdrawals/stats`

Retrieves withdrawal statistics.

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/stats" \
  -H "Authorization: Bearer your_admin_token"
```

#### Example Response
```json
{
  "success": true,
  "data": {
    "total_withdrawals": 1500,
    "pending_withdrawals": 25,
    "processing_withdrawals": 10,
    "completed_withdrawals": 1400,
    "failed_withdrawals": 50,
    "cancelled_withdrawals": 15,
    "today_withdrawals": 45,
    "hourly_withdrawals": 3,
    "total_amount_cents": 150000000,
    "today_amount_cents": 4500000,
    "hourly_amount_cents": 300000
  }
}
```

### 3. Get Withdrawal by ID
**GET** `/api/v1/withdrawals/id/{id}`

Retrieves a specific withdrawal by its database ID.

#### Path Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | The withdrawal's database ID |

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/id/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer your_admin_token"
```

### 4. Get Withdrawal by Withdrawal ID
**GET** `/api/v1/withdrawals/withdrawal-id/{withdrawal_id}`

Retrieves a specific withdrawal by its withdrawal ID.

#### Path Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `withdrawal_id` | string | Yes | The withdrawal's withdrawal ID |

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/withdrawal-id/WD123456789" \
  -H "Authorization: Bearer your_admin_token"
```

### 5. Get Withdrawals by User ID
**GET** `/api/v1/withdrawals/user/{user_id}`

Retrieves all withdrawals for a specific user.

#### Path Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | UUID | Yes | The user's ID |

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Number of results per page (default: 20, max: 100) |
| `offset` | integer | No | Number of results to skip (default: 0) |

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/user/456e7890-e89b-12d3-a456-426614174001?limit=10" \
  -H "Authorization: Bearer your_admin_token"
```

### 6. Get Withdrawals by Date Range
**GET** `/api/v1/withdrawals/date-range`

Retrieves withdrawals within a specific date range.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_date` | date | Yes | Start date (YYYY-MM-DD) |
| `end_date` | date | Yes | End date (YYYY-MM-DD) |
| `limit` | integer | No | Number of results per page (default: 20, max: 100) |
| `offset` | integer | No | Number of results to skip (default: 0) |

#### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/date-range?start_date=2024-01-01&end_date=2024-01-31&limit=50" \
  -H "Authorization: Bearer your_admin_token"
```

## Withdrawal Object Structure

```json
{
  "id": "string (UUID)",
  "user_id": "string (UUID)",
  "admin_id": "string (UUID, optional)",
  "chain_id": "string",
  "currency_code": "string",
  "protocol": "string",
  "withdrawal_id": "string",
  "usd_amount_cents": "integer",
  "crypto_amount": "string",
  "exchange_rate": "number",
  "fee_cents": "integer",
  "source_wallet_address": "string",
  "to_address": "string",
  "tx_hash": "string (optional)",
  "status": "string (pending|processing|completed|failed|cancelled)",
  "requires_admin_review": "boolean",
  "admin_review_deadline": "string (ISO 8601, optional)",
  "processed_by_system": "boolean",
  "amount_reserved_cents": "integer",
  "reservation_released": "boolean",
  "reservation_released_at": "string (ISO 8601, optional)",
  "metadata": "object (optional)",
  "error_message": "string (optional)",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)",
  "username": "string (optional)",
  "email": "string (optional)",
  "first_name": "string (optional)",
  "last_name": "string (optional)"
}
```

## Error Responses

All endpoints return errors in the following format:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

### Common HTTP Status Codes
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid token)
- `404` - Not Found (withdrawal not found)
- `500` - Internal Server Error

## Usage Examples

### 1. Get Recent Withdrawals
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals?limit=20&offset=0" \
  -H "Authorization: Bearer your_admin_token"
```

### 2. Get Pending Withdrawals
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals?status=pending" \
  -H "Authorization: Bearer your_admin_token"
```

### 3. Get Withdrawals for a Specific User
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/user/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer your_admin_token"
```

### 4. Get Withdrawals from Last Week
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals/date-range?start_date=2024-01-08&end_date=2024-01-15" \
  -H "Authorization: Bearer your_admin_token"
```

### 5. Search by Withdrawal ID
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals?withdrawal_id=WD123" \
  -H "Authorization: Bearer your_admin_token"
```

### 6. Search by Username
```bash
curl -X GET "http://localhost:8080/api/v1/withdrawals?username=john" \
  -H "Authorization: Bearer your_admin_token"
```

## Frontend Integration

The frontend components are located in:
- `back_office_FE/src/components/WithdrawalViewer/WithdrawalViewer.tsx` - Main withdrawal viewer
- `back_office_FE/src/components/WithdrawalManagement/WithdrawalManagement.tsx` - Paused withdrawals management
- `back_office_FE/src/components/WithdrawalDashboard/WithdrawalDashboard.tsx` - Dashboard overview

## Rate Limiting

All endpoints are subject to rate limiting. The default limits are:
- 100 requests per minute per IP
- 1000 requests per hour per authenticated user

## Security Notes

1. All endpoints require authentication
2. Sensitive data like wallet addresses are included in responses
3. Consider implementing additional access controls for production use
4. Log all withdrawal access for audit purposes
