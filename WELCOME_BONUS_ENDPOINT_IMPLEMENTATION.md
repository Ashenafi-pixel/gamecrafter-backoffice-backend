# Welcome Bonus Endpoint Implementation Documentation

## Overview

This document provides a comprehensive guide for implementing the `/analytics/users/{user_id}/welcome_bonus` endpoint in a backoffice codebase. The endpoint retrieves welcome bonus transactions for a specific user with pagination support.

**Endpoint:** `GET /analytics/users/{user_id}/welcome_bonus?limit=100&offset=0`

## Architecture Overview

The implementation follows a layered architecture pattern:

```
HTTP Request
    ↓
Route Handler (internal/glue/analytics/analytics.go)
    ↓
HTTP Handler (internal/handler/analytics/analytics.go)
    ↓
Storage Interface (internal/storage/analytics/analytics.go)
    ↓
Groove Storage Implementation (internal/storage/groove/groove.go)
    ↓
PostgreSQL Database (groove_transactions table)
```

## Database Schema

### Table: `groove_transactions`

The welcome bonus transactions are stored in the `groove_transactions` table with the following relevant columns:

```sql
CREATE TABLE public.groove_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id character varying(255) NOT NULL,
    account_id character varying(255) NOT NULL,
    user_id uuid,
    amount numeric(20,8) NOT NULL,
    currency character varying(10) DEFAULT 'USD' NOT NULL,
    type character varying(50) NOT NULL,  -- Must be 'welcome_bonus'
    status character varying(50) DEFAULT 'completed' NOT NULL,
    balance_before numeric(20,8) DEFAULT 0,
    balance_after numeric(20,8) DEFAULT 0,
    win_amount numeric(20,8) DEFAULT 0,
    net_result numeric(20,8) DEFAULT 0,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    brand_id uuid DEFAULT '00000000-0000-0000-0000-000000000001'::uuid NOT NULL
);
```

**Key Filter:** `type = 'welcome_bonus'`

## Implementation Details

### 1. DTO Definition

**File:** `internal/constant/dto/analytics.go`

```go
// WelcomeBonusTransaction represents a welcome bonus transaction
type WelcomeBonusTransaction struct {
    TransactionID   string                 `json:"transaction_id"`
    AccountID       string                 `json:"account_id"`
    Amount          decimal.Decimal         `json:"amount"`
    Currency        string                  `json:"currency"`
    Status          string                  `json:"status"`
    BalanceBefore   decimal.Decimal         `json:"balance_before"`
    BalanceAfter    decimal.Decimal         `json:"balance_after"`
    WinAmount       *decimal.Decimal        `json:"win_amount,omitempty"`
    NetResult       *decimal.Decimal        `json:"net_result,omitempty"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt       time.Time               `json:"created_at"`
}
```

### 2. Storage Interface

**File:** `internal/storage/analytics/analytics.go`

Add to the `AnalyticsStorage` interface:

```go
type AnalyticsStorage interface {
    // ... existing methods
    GetUserWelcomeBonus(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.WelcomeBonusTransaction, int, error)
}
```

Add interface for Groove Storage dependency:

```go
type GrooveStorageFetcher interface {
    GetUserWelcomeBonusTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.WelcomeBonusTransaction, int, error)
}
```

**Storage Implementation:**

```go
func (s *AnalyticsStorageImpl) GetUserWelcomeBonus(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.WelcomeBonusTransaction, int, error) {
    if s.grooveStorage == nil {
        return nil, 0, fmt.Errorf("groove storage not available")
    }
    return s.grooveStorage.GetUserWelcomeBonusTransactions(ctx, userID, limit, offset)
}
```

### 3. Groove Storage Implementation

**File:** `internal/storage/groove/groove.go`

Add method to `GrooveStorage` interface:

```go
type GrooveStorage interface {
    // ... existing methods
    GetUserWelcomeBonusTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.WelcomeBonusTransaction, int, error)
}
```

**Implementation:**

```go
// GetUserWelcomeBonusTransactions retrieves welcome bonus transactions for a user
func (s *GrooveStorageImpl) GetUserWelcomeBonusTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.WelcomeBonusTransaction, int, error) {
    s.logger.Info("Fetching user welcome bonus transactions",
        zap.String("user_id", userID.String()),
        zap.Int("limit", limit),
        zap.Int("offset", offset))

    // First, get the total count for pagination
    countQuery := `
        SELECT COUNT(*)
        FROM groove_transactions gt
        JOIN groove_accounts ga ON gt.account_id = ga.account_id
        JOIN users u ON ga.user_id = u.id
        WHERE ga.user_id = $1
            AND gt.type = 'welcome_bonus'
            AND gt.brand_id = u.brand_id
    `
    
    var totalCount int
    err := s.db.GetPool().QueryRow(ctx, countQuery, userID).Scan(&totalCount)
    if err != nil {
        s.logger.Error("Failed to count welcome bonus transactions", zap.Error(err))
        return nil, 0, fmt.Errorf("failed to count welcome bonus transactions: %w", err)
    }

    // Query welcome bonus transactions with pagination
    query := `
        SELECT 
            gt.transaction_id,
            gt.account_id,
            gt.amount,
            gt.currency,
            gt.status,
            COALESCE(gt.balance_before, 0) as balance_before,
            COALESCE(gt.balance_after, 0) as balance_after,
            gt.win_amount,
            gt.net_result,
            gt.metadata,
            gt.created_at
        FROM groove_transactions gt
        JOIN groove_accounts ga ON gt.account_id = ga.account_id
        JOIN users u ON ga.user_id = u.id
        WHERE ga.user_id = $1
            AND gt.type = 'welcome_bonus'
            AND gt.brand_id = u.brand_id
        ORDER BY gt.created_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := s.db.GetPool().Query(ctx, query, userID, limit, offset)
    if err != nil {
        s.logger.Error("Failed to fetch welcome bonus transactions", zap.Error(err))
        return nil, 0, fmt.Errorf("failed to fetch welcome bonus transactions: %w", err)
    }
    defer rows.Close()

    var transactions []*dto.WelcomeBonusTransaction
    for rows.Next() {
        var tx dto.WelcomeBonusTransaction
        var winAmount, netResult sql.NullString
        var metadataBytes []byte

        err := rows.Scan(
            &tx.TransactionID,
            &tx.AccountID,
            &tx.Amount,
            &tx.Currency,
            &tx.Status,
            &tx.BalanceBefore,
            &tx.BalanceAfter,
            &winAmount,
            &netResult,
            &metadataBytes,
            &tx.CreatedAt,
        )
        if err != nil {
            s.logger.Error("Failed to scan welcome bonus transaction", zap.Error(err))
            continue
        }

        // Parse optional fields
        if winAmount.Valid {
            if parsed, err := decimal.NewFromString(winAmount.String); err == nil {
                tx.WinAmount = &parsed
            }
        }

        if netResult.Valid {
            if parsed, err := decimal.NewFromString(netResult.String); err == nil {
                tx.NetResult = &parsed
            }
        }

        // Parse metadata JSON
        if len(metadataBytes) > 0 {
            var metadata map[string]interface{}
            if err := json.Unmarshal(metadataBytes, &metadata); err == nil {
                tx.Metadata = metadata
            }
        }

        transactions = append(transactions, &tx)
    }

    if err := rows.Err(); err != nil {
        s.logger.Error("Error iterating welcome bonus transaction rows", zap.Error(err))
        return nil, 0, fmt.Errorf("error iterating welcome bonus transaction rows: %w", err)
    }

    s.logger.Info("Successfully fetched welcome bonus transactions",
        zap.String("user_id", userID.String()),
        zap.Int("count", len(transactions)),
        zap.Int("total_count", totalCount))

    return transactions, totalCount, nil
}
```

### 4. Handler Implementation

**File:** `internal/handler/analytics/analytics.go`

Add method to `Analytics` interface in `internal/handler/handler.go`:

```go
type Analytics interface {
    // ... existing methods
    GetUserWelcomeBonus(c *gin.Context)
}
```

**Handler Implementation:**

```go
// GetUserWelcomeBonus Get user welcome bonus transactions
// @Summary Get user welcome bonus transactions
// @Description Retrieve welcome bonus transactions for a specific user with pagination
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 403 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/users/{user_id}/welcome_bonus [get]
func (a *analytics) GetUserWelcomeBonus(c *gin.Context) {
    // Parse user_id from path parameter
    userIDStr := c.Param("user_id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
            Success: false,
            Error:   "Invalid user ID format",
        })
        return
    }

    // Security: Verify user can access this user_id (prevents IDOR vulnerability)
    if err := a.checkUserAccess(c, userID); err != nil {
        a.logger.Warn("Unauthorized access attempt to user welcome bonus transactions",
            zap.String("requested_user_id", userID.String()),
            zap.Error(err))
        c.JSON(http.StatusForbidden, dto.AnalyticsResponse{
            Success: false,
            Error:   "Unauthorized: You can only access your own data",
        })
        return
    }

    // Parse pagination parameters
    limit := 100 // Default limit
    if limitStr := c.Query("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }

    offset := 0 // Default offset
    if offsetStr := c.Query("offset"); offsetStr != "" {
        if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        }
    }

    // Fetch welcome bonus transactions from storage
    welcomeBonusTransactions, totalCount, err := a.analyticsStorage.GetUserWelcomeBonus(c.Request.Context(), userID, limit, offset)
    if err != nil {
        a.logger.Error("Failed to get user welcome bonus transactions",
            zap.String("user_id", userID.String()),
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
            Success: false,
            Error:   "Failed to retrieve welcome bonus transactions",
        })
        return
    }

    // Calculate pagination metadata
    page := 1
    if limit > 0 {
        page = (offset / limit) + 1
    }
    pages := 1
    if limit > 0 {
        pages = (totalCount + limit - 1) / limit
    }

    // Return response
    c.JSON(http.StatusOK, dto.AnalyticsResponse{
        Success: true,
        Data:    welcomeBonusTransactions,
        Meta: &dto.Meta{
            Total:    totalCount,
            PageSize: limit,
            Page:     page,
            Pages:    pages,
        },
    })
}
```

**Security Check Implementation:**

The `checkUserAccess` method ensures users can only access their own data unless they are admins:

```go
// checkUserAccess verifies that the authenticated user can access the requested user_id
// Returns error if:
// - Authenticated user_id doesn't match requested user_id AND
// - Authenticated user is not an admin
// This prevents IDOR (Insecure Direct Object Reference) vulnerabilities
func (a *analytics) checkUserAccess(c *gin.Context, requestedUserID uuid.UUID) error {
    // Get authenticated user_id from context (set by Auth middleware)
    authenticatedUserIDStr, exists := c.Get("user-id")
    if !exists {
        return fmt.Errorf("user not authenticated")
    }

    authenticatedUserID, err := uuid.Parse(authenticatedUserIDStr.(string))
    if err != nil {
        return fmt.Errorf("invalid authenticated user ID: %w", err)
    }

    // If authenticated user is requesting their own data, allow access
    if authenticatedUserID == requestedUserID {
        return nil
    }

    // Check if authenticated user is an admin
    if a.pgPool == nil {
        return fmt.Errorf("postgres pool not available")
    }

    var isAdmin bool
    query := `
        SELECT is_admin 
        FROM users 
        WHERE id = $1
    `
    err = a.pgPool.QueryRow(c.Request.Context(), query, authenticatedUserID).Scan(&isAdmin)
    if err != nil {
        return fmt.Errorf("failed to check admin status: %w", err)
    }

    if !isAdmin {
        return fmt.Errorf("user is not authorized to access this resource")
    }

    return nil
}
```

### 5. Route Registration

**File:** `internal/glue/analytics/analytics.go`

Add route to the user analytics group:

```go
// User analytics routes (authentication required)
userAnalyticsGroup := grp.Group("/analytics")
userAnalyticsGroup.Use(middleware.Auth())
{
    // ... existing routes
    userAnalyticsGroup.GET("/users/:user_id/welcome_bonus", analyticsHandler.GetUserWelcomeBonus)
    // ... other routes
}
```

Also add to admin analytics group if needed:

```go
// Admin analytics routes (requires authentication)
adminAnalyticsGroup := grp.Group("/admin/analytics")
adminAnalyticsGroup.Use(middleware.Auth())
{
    // ... existing routes
    adminAnalyticsGroup.GET("/users/:user_id/welcome_bonus", analyticsHandler.GetUserWelcomeBonus)
    // ... other routes
}
```

### 6. Dependency Injection

**File:** `initiator/persistence.go`

Ensure `grooveStorage` is injected into `analyticsStorage`:

```go
Analytics: func() analyticsStorage.AnalyticsStorage {
    analyticsStorageInstance := analyticsStorage.NewAnalyticsStorage(
        clickhouseClient,
        gameSessionStorageInstance,
        gameAdapter,
        userAdapter,
        log,
    )
    
    // Inject groove storage for welcome bonus transactions
    if grooveStorageInstance != nil {
        if impl, ok := analyticsStorageInstance.(*analyticsStorage.AnalyticsStorageImpl); ok {
            impl.SetGrooveStorage(grooveStorageInstance)
        }
    }
    
    return analyticsStorageInstance
}(),
```

**File:** `initiator/handler.go`

Ensure `grooveStorage` is passed to analytics handler:

```go
analyticsHandler := analyticsHandler.Init(
    // ... other dependencies
    grooveStorage, // Add this dependency
    // ... other dependencies
)
```

## SQL Queries

### Count Query

```sql
SELECT COUNT(*)
FROM groove_transactions gt
JOIN groove_accounts ga ON gt.account_id = ga.account_id
JOIN users u ON ga.user_id = u.id
WHERE ga.user_id = $1
    AND gt.type = 'welcome_bonus'
    AND gt.brand_id = u.brand_id
```

**Parameters:**
- `$1`: `user_id` (UUID)

### Main Query

```sql
SELECT 
    gt.transaction_id,
    gt.account_id,
    gt.amount,
    gt.currency,
    gt.status,
    COALESCE(gt.balance_before, 0) as balance_before,
    COALESCE(gt.balance_after, 0) as balance_after,
    gt.win_amount,
    gt.net_result,
    gt.metadata,
    gt.created_at
FROM groove_transactions gt
JOIN groove_accounts ga ON gt.account_id = ga.account_id
JOIN users u ON ga.user_id = u.id
WHERE ga.user_id = $1
    AND gt.type = 'welcome_bonus'
    AND gt.brand_id = u.brand_id
ORDER BY gt.created_at DESC
LIMIT $2 OFFSET $3
```

**Parameters:**
- `$1`: `user_id` (UUID)
- `$2`: `limit` (integer)
- `$3`: `offset` (integer)

**Key Points:**
- Joins `groove_transactions` with `groove_accounts` to get user association
- Joins with `users` table for brand isolation
- Filters by `type = 'welcome_bonus'`
- Enforces brand isolation via `gt.brand_id = u.brand_id`
- Orders by `created_at DESC` (newest first)
- Supports pagination with `LIMIT` and `OFFSET`

## Response Format

### Success Response (200 OK)

```json
{
    "success": true,
    "data": [
        {
            "transaction_id": "3d3aeff9-b54f-4a23-af00-96b6e44026e8",
            "account_id": "account-123",
            "amount": "50.00",
            "currency": "USD",
            "status": "completed",
            "balance_before": "0.00",
            "balance_after": "50.00",
            "win_amount": null,
            "net_result": null,
            "metadata": {
                "bonus_type": "fixed",
                "brand_id": "00000000-0000-0000-0000-000000000001",
                "deposit_amount": "0"
            },
            "created_at": "2025-12-10T18:22:58.758282Z"
        }
    ],
    "meta": {
        "total": 1,
        "page_size": 100,
        "page": 1,
        "pages": 1
    }
}
```

### Error Responses

**400 Bad Request - Invalid User ID:**
```json
{
    "success": false,
    "error": "Invalid user ID format"
}
```

**403 Forbidden - Unauthorized Access:**
```json
{
    "success": false,
    "error": "Unauthorized: You can only access your own data"
}
```

**500 Internal Server Error:**
```json
{
    "success": false,
    "error": "Failed to retrieve welcome bonus transactions"
}
```

## Security Considerations

1. **Authentication Required:** The endpoint requires authentication via `middleware.Auth()`
2. **Authorization Check:** Users can only access their own welcome bonus transactions unless they are admins
3. **Brand Isolation:** Queries enforce brand isolation via `gt.brand_id = u.brand_id`
4. **IDOR Prevention:** The `checkUserAccess` method prevents Insecure Direct Object Reference vulnerabilities
5. **Input Validation:** User ID is validated as UUID format
6. **Pagination Limits:** Consider implementing maximum limit constraints to prevent resource exhaustion

## Testing

### Example cURL Request

```bash
curl -X GET "https://api.example.com/analytics/users/1c435552-d8f4-46bc-887e-2fe613bcf308/welcome_bonus?limit=100&offset=0" \
  -H "Authorization: Bearer <token>"
```

### Test Cases

1. **Valid Request:** User requests their own welcome bonus transactions
2. **Pagination:** Test with different limit and offset values
3. **Unauthorized Access:** Non-admin user tries to access another user's data
4. **Admin Access:** Admin user accesses any user's data
5. **Invalid User ID:** Request with malformed UUID
6. **Empty Results:** User with no welcome bonus transactions
7. **Large Dataset:** Test pagination with many transactions

## Integration Points

1. **Welcome Bonus Creation:** When a welcome bonus is applied, a transaction is created in `groove_transactions` with `type = 'welcome_bonus'`
2. **Tips Endpoint:** Welcome bonuses are also included in the `/analytics/users/{user_id}/tips` endpoint response as `transaction_type: "welcome_bonus"`
3. **Balance Updates:** Welcome bonus transactions include `balance_before` and `balance_after` for tracking balance changes

## Notes

- The endpoint follows the same architectural pattern as other analytics endpoints (e.g., `/rakeback`, `/tips`)
- Welcome bonus transactions are stored in the `groove_transactions` table, not a separate table
- The `metadata` field contains JSON with bonus-specific information (bonus_type, brand_id, deposit_amount)
- Brand isolation is critical for multi-tenant systems
- The endpoint supports pagination for efficient data retrieval

