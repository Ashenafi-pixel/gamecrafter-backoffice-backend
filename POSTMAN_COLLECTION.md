# Postman Collection - Brand & Player Management API

## Base URL
```
http://localhost:8080
```
*Replace with your actual server URL*

## Authentication
All endpoints require authentication. Include the authorization token in headers:
```
Authorization: Bearer <your_token_here>
```

---

## Brand Management Endpoints

### 1. Create Brand
**POST** `/api/admin/brands`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Body (JSON):**
```json
{
  "name": "Premium Gaming",
  "code": "PG001",
  "domain": "premiumgaming.com",
  "is_active": true,
  "description": "Premium gaming brand for high-end players",
  "signature": "premium-signature-key",
  "webhook_url": "https://premiumgaming.com/webhook",
  "integration_type": "API",
  "api_url": "https://api.premiumgaming.com"
}
```

**Minimal Request:**
```json
{
  "name": "Test Brand",
  "code": "TEST001",
  "is_active": true
}
```

**Expected Response (201 Created):**
```json
{
  "id": 100000,
  "name": "Premium Gaming",
  "code": "PG001",
  "domain": "premiumgaming.com",
  "is_active": true,
  "created_at": "2026-02-18T19:50:41.618Z",
  "updated_at": "2026-02-18T19:50:41.618Z"
}
```

---

### 2. Get All Brands
**GET** `/api/admin/brands?page=1&per_page=10&search=premium&is_active=true`

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (required): Page number (min: 1)
- `per_page` (required): Items per page (min: 1, max: 100)
- `search` (optional): Search term for name or code
- `is_active` (optional): Filter by active status (true/false)

**Example URLs:**
```
GET /api/admin/brands?page=1&per_page=10
GET /api/admin/brands?page=1&per_page=20&search=gaming
GET /api/admin/brands?page=1&per_page=10&is_active=true
```

**Expected Response (200 OK):**
```json
{
  "brands": [
    {
      "id": 100000,
      "name": "Premium Gaming",
      "code": "PG001",
      "domain": "premiumgaming.com",
      "is_active": true,
      "webhook_url": "https://premiumgaming.com/webhook",
      "api_url": "https://api.premiumgaming.com",
      "created_at": "2026-02-18T19:50:41.618Z",
      "updated_at": "2026-02-18T19:50:41.618Z"
    }
  ],
  "total_count": 1,
  "total_pages": 1,
  "current_page": 1,
  "per_page": 10
}
```

---

### 3. Get Brand by ID
**GET** `/api/admin/brands/100000`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Brand ID (6-digit numeric, e.g., 100000)

**Expected Response (200 OK):**
```json
{
  "id": 100000,
  "name": "Premium Gaming",
  "code": "PG001",
  "domain": "premiumgaming.com",
  "is_active": true,
  "webhook_url": "https://premiumgaming.com/webhook",
  "api_url": "https://api.premiumgaming.com",
  "created_at": "2026-02-18T19:50:41.618Z",
  "updated_at": "2026-02-18T19:50:41.618Z"
}
```

---

### 4. Update Brand
**PATCH** `/api/admin/brands/100000`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Body (JSON) - All fields optional:**
```json
{
  "name": "Premium Gaming Updated",
  "code": "PG001",
  "domain": "newdomain.com",
  "webhook_url": "https://newdomain.com/webhook",
  "api_url": "https://api.newdomain.com",
  "is_active": false
}
```

**Partial Update Example:**
```json
{
  "is_active": false
}
```

**Expected Response (200 OK):**
```json
{
  "id": 100000,
  "name": "Premium Gaming Updated",
  "code": "PG001",
  "domain": "newdomain.com",
  "webhook_url": "https://newdomain.com/webhook",
  "api_url": "https://api.newdomain.com",
  "is_active": false,
  "created_at": "2026-02-18T19:50:41.618Z",
  "updated_at": "2026-02-18T20:00:00.000Z"
}
```

---

### 5. Delete Brand
**DELETE** `/api/admin/brands/100000`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Brand ID (6-digit numeric)

**Expected Response (204 No Content)**

---

### 6. Add Allowed Origin
**POST** `/api/admin/brands/:id/allowed-origins`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Brand ID (e.g. 100000)

**Body (JSON):**
```json
{
  "origin": "https://game.example.com"
}
```

**Expected Response (201 Created):**
```json
{
  "id": 1,
  "brand_id": 100000,
  "origin": "https://game.example.com",
  "created_at": "2026-02-22T12:00:00Z"
}
```

---

### 7. Get Allowed Origins (List)
**GET** `/api/admin/brands/:id/allowed-origins`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Brand ID

**Expected Response (200 OK):**
```json
{
  "origins": [
    {
      "id": 1,
      "brand_id": 100000,
      "origin": "https://game.example.com",
      "created_at": "2026-02-22T12:00:00Z"
    }
  ]
}
```

---

### 8. Delete Allowed Origin
**DELETE** `/api/admin/brands/:id/allowed-origins/:originId`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Brand ID
- `originId`: Allowed origin row id (from List response)

**Example:** `DELETE /api/admin/brands/100000/allowed-origins/1`

**Expected Response (204 No Content)**

---

## Player Management Endpoints

### 1. Create Player
**POST** `/api/admin/player-management`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Body (JSON):**
```json
{
  "email": "player@example.com",
  "username": "player123",
  "password": "SecurePass123!",
  "phone": "+1234567890",
  "first_name": "John",
  "last_name": "Doe",
  "default_currency": "USD",
  "brand": "Premium Gaming",
  "date_of_birth": "1990-01-15T00:00:00Z",
  "country": "United States",
  "state": "California",
  "street_address": "123 Main St",
  "postal_code": "90210",
  "test_account": false,
  "enable_withdrawal_limit": true,
  "brand_id": 100000
}
```

**Minimal Request:**
```json
{
  "email": "player@example.com",
  "username": "player123",
  "password": "SecurePass123!",
  "default_currency": "USD",
  "date_of_birth": "1990-01-15T00:00:00Z",
  "country": "United States"
}
```

**Expected Response (201 Created):**
```json
{
  "id": 1,
  "email": "player@example.com",
  "username": "player123",
  "created_at": "2026-02-18T19:50:41.618Z"
}
```

---

### 2. Get All Players
**GET** `/api/admin/player-management?page=1&per_page=10&search=player&brand_id=100000&country=United States&test_account=false`

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (required): Page number (min: 1)
- `per_page` (required): Items per page (min: 1, max: 100)
- `search` (optional): Search term for username or email
- `brand_id` (optional): Filter by brand ID (6-digit numeric as string, e.g., "100000")
- `country` (optional): Filter by country (partial match)
- `test_account` (optional): Filter by test account status (true/false)
- `sort_by` (optional): Sort by field - `email`, `username`, `created_at`, `updated_at`, `date_of_birth`, `country` (default: `created_at`)
- `sort_order` (optional): Sort order - `asc` or `desc` (default: `desc`)

**Example URLs:**
```
GET /api/admin/player-management?page=1&per_page=10
GET /api/admin/player-management?page=1&per_page=20&search=john
GET /api/admin/player-management?page=1&per_page=10&brand_id=100000
GET /api/admin/player-management?page=1&per_page=10&country=United States&test_account=false
GET /api/admin/player-management?page=1&per_page=10&sort_by=email&sort_order=asc
GET /api/admin/player-management?page=1&per_page=10&sort_by=created_at&sort_order=desc
```

**Expected Response (200 OK):**
```json
{
  "players": [
    {
      "id": 1,
      "email": "player@example.com",
      "username": "player123",
      "phone": "+1234567890",
      "first_name": "John",
      "last_name": "Doe",
      "default_currency": "USD",
      "brand": "Premium Gaming",
      "date_of_birth": "1990-01-15T00:00:00Z",
      "country": "United States",
      "state": "California",
      "street_address": "123 Main St",
      "postal_code": "90210",
      "test_account": false,
      "enable_withdrawal_limit": true,
      "brand_id": 100000,
      "created_at": "2026-02-18T19:50:41.618Z",
      "updated_at": "2026-02-18T19:50:41.618Z"
    }
  ],
  "total_count": 1,
  "total_pages": 1,
  "current_page": 1,
  "per_page": 10
}
```

---

### 3. Get Player by ID
**GET** `/api/admin/player-management/1`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Player ID (integer)

**Expected Response (200 OK):**
```json
{
  "player": {
    "id": 1,
    "email": "player@example.com",
    "username": "player123",
    "phone": "+1234567890",
    "first_name": "John",
    "last_name": "Doe",
    "default_currency": "USD",
    "brand": "Premium Gaming",
    "date_of_birth": "1990-01-15T00:00:00Z",
    "country": "United States",
    "state": "California",
    "street_address": "123 Main St",
    "postal_code": "90210",
    "test_account": false,
    "enable_withdrawal_limit": true,
    "brand_id": 100000,
    "created_at": "2026-02-18T19:50:41.618Z",
    "updated_at": "2026-02-18T19:50:41.618Z"
  }
}
```

---

### 4. Update Player
**PATCH** `/api/admin/player-management/1`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Body (JSON) - All fields optional except ID:**
```json
{
  "email": "newemail@example.com",
  "username": "newusername",
  "phone": "+9876543210",
  "first_name": "Jane",
  "last_name": "Smith",
  "default_currency": "EUR",
  "brand": "New Brand",
  "date_of_birth": "1992-05-20T00:00:00Z",
  "country": "Canada",
  "state": "Ontario",
  "street_address": "456 Oak Ave",
  "postal_code": "M5H 2N2",
  "test_account": true,
  "enable_withdrawal_limit": false,
  "brand_id": 100001
}
```

**Partial Update Example:**
```json
{
  "email": "updated@example.com",
  "phone": "+1111111111"
}
```

**Expected Response (200 OK):**
```json
{
  "player": {
    "id": 1,
    "email": "newemail@example.com",
    "username": "newusername",
    "phone": "+9876543210",
    "first_name": "Jane",
    "last_name": "Smith",
    "default_currency": "EUR",
    "brand": "New Brand",
    "date_of_birth": "1992-05-20T00:00:00Z",
    "country": "Canada",
    "state": "Ontario",
    "street_address": "456 Oak Ave",
    "postal_code": "M5H 2N2",
    "test_account": true,
    "enable_withdrawal_limit": false,
    "brand_id": 100001,
    "created_at": "2026-02-18T19:50:41.618Z",
    "updated_at": "2026-02-18T20:00:00.000Z"
  }
}
```

---

### 5. Delete Player
**DELETE** `/api/admin/player-management/1`

**Headers:**
```
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Player ID (integer)

**Expected Response (204 No Content)**

---

## Error Responses

All endpoints may return the following error responses:

**400 Bad Request:**
```json
{
  "error": "Invalid input data",
  "message": "validation failed: email is required"
}
```

**401 Unauthorized:**
```json
{
  "error": "Unauthorized",
  "message": "authentication required"
}
```

**404 Not Found:**
```json
{
  "error": "Resource not found",
  "message": "brand not found with ID: 999999"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error",
  "message": "unable to create brand"
}
```

---

## Notes

1. **Brand IDs**: Now use 6-digit numeric IDs (100000-999999) instead of UUIDs
2. **Player Brand ID**: References brand using the 6-digit numeric ID
3. **Date Format**: Use ISO 8601 format for dates (e.g., "1990-01-15T00:00:00Z")
4. **Pagination**: All list endpoints require `page` and `per_page` parameters
5. **Authentication**: All endpoints require a valid Bearer token in the Authorization header

