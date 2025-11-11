# Auth Middleware Guide - Token Verification & JWT Secret

## Overview
This guide explains how the authentication middleware verifies JWT tokens in the TucanBIT backend and where the JWT secret is configured.

---

## 🔐 JWT Secret Configuration

### Configuration Locations

The JWT secret can be configured in **three ways** (in order of priority):

1. **Environment Variable** (Highest Priority)
   ```bash
   export JWT_SECRET="your-secret-key-here"
   ```

2. **Config File: `app.jwt_secret`**
   ```yaml
   # config/config.yaml
   app:
     jwt_secret: "your-secret-key-here"
   ```

3. **Config File: `auth.jwt_secret`** (Fallback)
   ```yaml
   # config/config.yaml
   auth:
     jwt_secret: "tokensecrethere"  # Fallback for backward compatibility
   ```

### Current Configuration

Based on `config/config.yaml`:
```yaml
auth:
  jwt_secret: "tokensecrethere"
  otp_jwt_secret: "otpkwtsecretgoinghere"
```

### Environment Variable Binding

The system automatically binds the `JWT_SECRET` environment variable to both config keys:
- `app.jwt_secret`
- `auth.jwt_secret`

This is configured in `initiator/config.go`:
```go
viper.BindEnv("app.jwt_secret", "JWT_SECRET")
viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
```

---

## 🔍 How Token Verification Works

### Auth Middleware Location
**File:** `internal/handler/middleware/auth.go`

### Step-by-Step Verification Process

#### 1. **Get JWT Secret**
```go
key := viper.GetString("app.jwt_secret")
if key == "" {
    key = viper.GetString("auth.jwt_secret") // Fallback
}
if key == "" {
    // Error: JWT secret not configured
    c.Abort()
    return
}
jwtKey := []byte(key)
```

#### 2. **Extract Authorization Header**
```go
tokenString := c.GetHeader("Authorization")
if tokenString == "" {
    // Error: authorization header is missing
    c.Abort()
    return
}
```

#### 3. **Validate Bearer Token Format**
```go
// Must be: "Bearer <token>"
if len(tokenString) <= 7 || strings.ToUpper(tokenString[:7]) != "BEARER " {
    // Error: authorization format is Bearer <token>
    c.Abort()
    return
}
```

#### 4. **Extract Token from Header**
```go
tokenString = tokenString[7:] // Remove "Bearer " prefix
```

#### 5. **Parse and Verify JWT Token**
```go
claims := &dto.Claim{}
token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
    return jwtKey, nil
})
```

#### 6. **Validate Token**
```go
if err != nil || !token.Valid {
    // Error: invalid or expired token
    c.Abort()
    return
}
```

#### 7. **Set User Context**
```go
c.Set("user_id", claims.UserID)
c.Set("user-id", claims.UserID.String())
c.Set("is-verified", claims.IsVerified)
c.Set("email-verified", claims.EmailVerified)
c.Set("phone-verified", claims.PhoneVerified)
```

---

## 📋 Complete Auth Middleware Code

```go
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get JWT secret from config
		key := viper.GetString("app.jwt_secret")
		if key == "" {
			key = viper.GetString("auth.jwt_secret") // Fallback
		}
		if key == "" {
			err := fmt.Errorf("JWT secret not configured")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		jwtKey := []byte(key)
		
		// 2. Check if authorization header exists
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			err := fmt.Errorf("authorization header is missing")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		
		// 3. Check if it's bearer token format
		if len(tokenString) <= 7 || strings.ToUpper(tokenString[:7]) != "BEARER " {
			err := fmt.Errorf("authorization format is Bearer <token> ")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// 4. Extract token (remove "Bearer " prefix)
		tokenString = tokenString[7:]
		
		// 5. Parse and validate token
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			err := fmt.Errorf("invalid or expired token ")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		
		// 6. Set user context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user-id", claims.UserID.String())
		c.Set("is-verified", claims.IsVerified)
		c.Set("email-verified", claims.EmailVerified)
		c.Set("phone-verified", claims.PhoneVerified)
	}
}
```

---

## 🎯 JWT Claims Structure

The middleware expects tokens with the following claims structure:

```go
type Claim struct {
	UserID        uuid.UUID `json:"user_id"`
	IsVerified    bool      `json:"is_verified"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	jwt.StandardClaims
}
```

**StandardClaims includes:**
- `ExpiresAt` - Token expiration time
- `IssuedAt` - Token issue time
- `Issuer` - Token issuer
- `Subject` - Token subject

---

## 🚀 Usage in Routes

### Basic Usage
```go
router.GET("/api/protected", middleware.Auth(), handler.ProtectedHandler)
```

### With Additional Middleware
```go
router.GET("/api/verified-only", 
    middleware.Auth(), 
    middleware.RequireVerification(), 
    handler.VerifiedHandler)
```

### Accessing User Info in Handlers
```go
func ProtectedHandler(c *gin.Context) {
    userID, _ := c.Get("user-id")        // string
    userIDUUID, _ := c.Get("user_id")   // uuid.UUID
    isVerified, _ := c.Get("is-verified") // bool
    emailVerified, _ := c.Get("email-verified") // bool
    phoneVerified, _ := c.Get("phone-verified") // bool
    
    // Use the values...
}
```

---

## 🔧 Additional Middleware Functions

### RequireVerification()
Ensures the user account is verified:
```go
middleware.RequireVerification()
```

### RequireEmailVerification()
Ensures the user's email is verified:
```go
middleware.RequireEmailVerification()
```

### RequirePhoneVerification()
Ensures the user's phone is verified:
```go
middleware.RequirePhoneVerification()
```

---

## 📝 Example Request

### Valid Request
```bash
curl -X GET http://localhost:8094/api/protected \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Invalid Requests
```bash
# Missing header
curl -X GET http://localhost:8094/api/protected

# Wrong format
curl -X GET http://localhost:8094/api/protected \
  -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Invalid token
curl -X GET http://localhost:8094/api/protected \
  -H "Authorization: Bearer invalid-token-here"
```

---

## ⚙️ Configuration Priority

1. **Environment Variable** (`JWT_SECRET`) - Highest priority
2. **Config File** (`app.jwt_secret`)
3. **Config File** (`auth.jwt_secret`) - Fallback

---

## 🔒 Security Best Practices

1. **Never commit JWT secrets to version control**
   - Use environment variables in production
   - Use `.env` files` for local development (add to `.gitignore`)

2. **Use strong, random secrets**
   ```bash
   # Generate a secure random secret
   openssl rand -base64 32
   ```

3. **Rotate secrets regularly**
   - Change JWT secret periodically
   - Invalidate old tokens when rotating

4. **Use HTTPS in production**
   - Always use HTTPS to protect tokens in transit

---

## 🐛 Troubleshooting

### Error: "JWT secret not configured"
- **Solution:** Set `JWT_SECRET` environment variable or configure in `config.yaml`

### Error: "authorization header is missing"
- **Solution:** Ensure the request includes `Authorization: Bearer <token>` header

### Error: "invalid or expired token"
- **Solution:** 
  - Check if token is expired
  - Verify token was signed with the correct JWT secret
  - Ensure token format is correct

### Token verification fails but token looks valid
- **Check:** Ensure the JWT secret used to sign the token matches the one in config
- **Check:** Verify token hasn't expired
- **Check:** Ensure token claims structure matches `dto.Claim`

---

## 📚 Related Files

- **Middleware:** `internal/handler/middleware/auth.go`
- **Claims DTO:** `internal/constant/dto/auth.go`
- **Config:** `config/config.yaml`
- **Config Loader:** `initiator/config.go`

---

## 🔗 Other Auth Middleware

The codebase also includes:
- `SportsAuth()` - For sports service authentication
- `AddsAuth()` - For ads service authentication
- `LotteryAuth()` - For lottery service authentication
- `LotteryUserAuth()` - For lottery user authentication

These use the same JWT secret from `auth.jwt_secret` but may use different headers or claim structures.


