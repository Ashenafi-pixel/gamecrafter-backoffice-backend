# Welcome Bonus Settings API Documentation

## Overview
This API allows you to manage welcome bonus configurations for users. The welcome bonus can be configured as either a fixed amount or a percentage-based bonus on first deposit.

**Base URL:** `{{base_url}}/api/admin/settings/welcome-bonus`

---

## Endpoints

### 1. Get Welcome Bonus Settings

Retrieves the current welcome bonus settings for a specific brand.

#### Request

**Method:** `GET`

**URL:** `/api/admin/settings/welcome-bonus`

**Headers:**
```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Query Parameters:**
- `brand_id` (required, string, UUID): The brand ID for which to retrieve settings

**Example Request:**
```
GET /api/admin/settings/welcome-bonus?brand_id=00000000-0000-0000-0000-000000000002
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "type": "fixed",
    "enabled": true,
    "fixed_amount": 50.0,
    "percentage": 0.0,
    "min_deposit_amount": 0.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings retrieved successfully"
}
```

**Error Response (400 Bad Request):**
```json
{
  "code": 400,
  "message": "brand_id is required for welcome bonus settings"
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "code": 500,
  "message": "Internal server error message"
}
```

---

### 2. Update Welcome Bonus Settings

Updates or creates welcome bonus settings for a specific brand.

#### Request

**Method:** `PUT`

**URL:** `/api/admin/settings/welcome-bonus`

**Headers:**
```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body:**
```json
{
  "brand_id": "00000000-0000-0000-0000-000000000002",
  "type": "fixed",
  "enabled": true,
  "fixed_amount": 50.0,
  "percentage": 0.0,
  "min_deposit_amount": 0.0,
  "max_bonus_percentage": 90.0
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `brand_id` | string (UUID) | Yes | The brand ID for which to update settings. Can also be passed as query parameter. |
| `type` | string | Yes | Bonus type: `"fixed"` or `"percentage"`. Only one type can be enabled at a time. |
| `enabled` | boolean | Yes | Whether the welcome bonus is enabled (`true`) or disabled (`false`). |
| `fixed_amount` | number | Yes | Fixed bonus amount (used when `type` is `"fixed"`). Set to `0.0` for percentage type. |
| `percentage` | number | Yes | Percentage value for deposit-based bonus (e.g., `50.0` means 50%). Used when `type` is `"percentage"`. Set to `0.0` for fixed type. |
| `min_deposit_amount` | number | Yes | Minimum deposit amount required to qualify for percentage-based bonus. Used when `type` is `"percentage"`. Set to `0.0` for fixed type. |
| `max_bonus_percentage` | number | Yes | Maximum bonus as percentage of deposit (default: `90.0`). Must be less than `100` to prevent bonus from equaling deposit. Used for percentage type. |

#### Response

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "type": "fixed",
    "enabled": true,
    "fixed_amount": 50.0,
    "percentage": 0.0,
    "min_deposit_amount": 0.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings updated successfully"
}
```

**Error Response (400 Bad Request):**
```json
{
  "code": 400,
  "message": "brand_id is required for welcome bonus settings"
}
```

**Error Response (400 Bad Request - Validation Error):**
```json
{
  "code": 400,
  "message": "max_bonus_percentage must be less than 100 to prevent bonus from equaling deposit"
}
```

**Error Response (401 Unauthorized):**
```json
{
  "code": 401,
  "message": "Unauthorized"
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "code": 500,
  "message": "Internal server error message"
}
```

---

## Configuration Examples

### Fixed Amount Bonus

**Request:**
```json
{
  "brand_id": "00000000-0000-0000-0000-000000000002",
  "type": "fixed",
  "enabled": true,
  "fixed_amount": 50.0,
  "percentage": 0.0,
  "min_deposit_amount": 0.0,
  "max_bonus_percentage": 90.0
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "type": "fixed",
    "enabled": true,
    "fixed_amount": 50.0,
    "percentage": 0.0,
    "min_deposit_amount": 0.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings updated successfully"
}
```

---

### Percentage-Based Bonus (50% Match)

**Request:**
```json
{
  "brand_id": "00000000-0000-0000-0000-000000000002",
  "type": "percentage",
  "enabled": true,
  "fixed_amount": 0.0,
  "percentage": 50.0,
  "min_deposit_amount": 100.0,
  "max_bonus_percentage": 90.0
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "type": "percentage",
    "enabled": true,
    "fixed_amount": 0.0,
    "percentage": 50.0,
    "min_deposit_amount": 100.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings updated successfully"
}
```

**How it works:**
- User deposits $200
- Bonus = $200 × 50% = $100
- But max bonus is capped at 90% of deposit = $180
- Final bonus = min($100, $180) = $100

---

### Percentage-Based Bonus (100% Match, Capped at 90%)

**Request:**
```json
{
  "brand_id": "00000000-0000-0000-0000-000000000002",
  "type": "percentage",
  "enabled": true,
  "fixed_amount": 0.0,
  "percentage": 100.0,
  "min_deposit_amount": 50.0,
  "max_bonus_percentage": 90.0
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "type": "percentage",
    "enabled": true,
    "fixed_amount": 0.0,
    "percentage": 100.0,
    "min_deposit_amount": 50.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings updated successfully"
}
```

**How it works:**
- User deposits $100
- Bonus = $100 × 100% = $100
- But max bonus is capped at 90% of deposit = $90
- Final bonus = min($100, $90) = $90 (anti-scam protection)

---

### Disable Welcome Bonus

**Request:**
```json
{
  "brand_id": "00000000-0000-0000-0000-000000000002",
  "type": "fixed",
  "enabled": false,
  "fixed_amount": 0.0,
  "percentage": 0.0,
  "min_deposit_amount": 0.0,
  "max_bonus_percentage": 90.0
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "type": "fixed",
    "enabled": false,
    "fixed_amount": 0.0,
    "percentage": 0.0,
    "min_deposit_amount": 0.0,
    "max_bonus_percentage": 90.0
  },
  "message": "Welcome bonus settings updated successfully"
}
```

---

## Important Notes

### 1. Mutually Exclusive Types
- Only one bonus type can be enabled at a time
- When `type` is `"fixed"`, set `percentage` and `min_deposit_amount` to `0.0`
- When `type` is `"percentage"`, set `fixed_amount` to `0.0`

### 2. Anti-Scam Protection
- `max_bonus_percentage` must always be less than `100`
- This prevents users from depositing and getting the same amount as bonus
- Example: If `max_bonus_percentage` is `90.0`, even a 100% match bonus will be capped at 90% of the deposit

### 3. Percentage Bonus Calculation
The actual bonus awarded is calculated as:
```
bonus = min(deposit × percentage / 100, deposit × max_bonus_percentage / 100)
```

### 4. Authentication
- All endpoints require a valid JWT token in the `Authorization` header
- Token format: `Bearer {JWT_TOKEN}`
- Token must be obtained through the login endpoint

### 5. Brand ID
- `brand_id` is required for all operations
- Can be provided in:
  - Request body (preferred for PUT requests)
  - Query parameter (for GET requests)
  - Form data (alternative for PUT requests)

---

## Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid input or validation error |
| 401 | Unauthorized - Missing or invalid JWT token |
| 500 | Internal Server Error - Server-side error |

---

## Default Values

When no configuration exists, the API returns default values:

```json
{
  "type": "fixed",
  "enabled": false,
  "fixed_amount": 0.0,
  "percentage": 0.0,
  "min_deposit_amount": 0.0,
  "max_bonus_percentage": 90.0
}
```


