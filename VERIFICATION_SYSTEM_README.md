# Enterprise-Grade User Verification & Route Protection System

This document describes the implementation of a world-class user verification system that ensures phone numbers, emails, and other unique constraints are validated before sending emails, and allows unverified users to login but requires verification for access to protected resources.

## Architecture Overview

The system implements a multi-layered, enterprise-grade approach:
1. **Unique Constraint Validation**: Prevents duplicate registrations before email sending
2. **Multi-Tier Authentication**: Allows unverified users to login while restricting access
3. **Comprehensive Route Protection**: Enterprise-grade middleware for different verification levels
4. **Audit & Compliance**: Full audit logging and compliance tracking
5. **Rate Limiting & Security**: Production-ready security measures

## Core Components

### 1. Unique Constraint Validation Engine
- **Email uniqueness**: Prevents duplicate email registrations
- **Phone number uniqueness**: Prevents duplicate phone number registrations  
- **Username uniqueness**: Prevents duplicate username registrations
- **Referral code uniqueness**: Prevents duplicate referral codes
- **Real-time validation**: Checks constraints before any email is sent

### 2. Enhanced JWT Claims System
JWT tokens now include comprehensive verification status:
```json
{
  "user_id": "uuid",
  "is_verified": false,
  "email_verified": false,
  "phone_verified": false,
  "exp": 1234567890,
  "iat": 1234567890,
  "iss": "tucanbit",
  "sub": "uuid"
}
```

### 3. Enterprise-Grade Middleware System

#### VerificationMiddleware
- **RequireVerification()**: Full account verification required
- **RequireEmailVerification()**: Email verification required
- **RequirePhoneVerification()**: Phone verification required
- **RequirePartialVerification()**: At least one verification method
- **RequireBettingAccess()**: Betting-specific verification
- **RequireFinancialAccess()**: Financial transaction verification
- **RequireKYCAccess()**: KYC verification required

#### RouteProtector
- **Protect(level)**: Dynamic protection based on verification level
- **Comprehensive rate limiting**: Production-ready rate limiting
- **Audit logging**: Full compliance and security logging
- **Custom verification logic**: Extensible verification requirements

#### MiddlewareManager
- **Centralized management**: Single point of control for all middleware
- **Comprehensive middleware**: Combines all protection layers
- **Configuration management**: Centralized configuration control
- **Performance optimization**: Optimized middleware execution

## Implementation Details

### Storage Layer
Enterprise-grade unique constraint validation:
```go
CheckEmailExists(ctx context.Context, email string) (bool, error)
CheckPhoneExists(ctx context.Context, phone string) (bool, error)
CheckUsernameExists(ctx context.Context, username string) (bool, error)
ValidateUniqueConstraints(ctx context.Context, userRequest dto.User) error
```

### Module Layer
Business logic integration:
```go
CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
CheckUserExistsByPhoneNumber(ctx context.Context, phone string) (bool, error)
CheckUserExistsByUsername(ctx context.Context, username string) (bool, error)
```

### Registration Flow
1. **Unique Constraint Validation**: Real-time constraint checking
2. **Temporary Storage**: Redis-based temporary data storage
3. **Email Verification**: Secure OTP-based verification
4. **Account Creation**: User account creation after verification
5. **Status Tracking**: Comprehensive verification status tracking

## Enterprise Features

### Security & Compliance
- **Rate Limiting**: Configurable rate limiting per user/IP
- **Audit Logging**: Complete audit trail for compliance
- **IP Tracking**: Full IP address and user agent logging
- **Request ID Tracking**: Unique request identification
- **Timestamp Logging**: Precise timing for all operations

### Performance & Scalability
- **Middleware Optimization**: Efficient middleware execution
- **Context Management**: Optimized context passing
- **Error Handling**: Comprehensive error handling and logging
- **Resource Management**: Efficient resource utilization

### Extensibility
- **Custom Verification Logic**: Pluggable verification requirements
- **Configuration Management**: Centralized configuration control
- **Middleware Composition**: Flexible middleware combination
- **Custom Error Messages**: Configurable error responses

## Production Deployment

### Configuration
```go
config := &MiddlewareConfig{
    VerificationEnabled:      true,
    RequireEmailVerification: true,
    RequirePhoneVerification: true,
    RequireFullVerification:  true,
    RouteProtectionEnabled:   true,
    DefaultProtectionLevel:   FullVerification,
    RateLimitEnabled:         true,
    MaxAttempts:              5,
    WindowDuration:           15 * time.Minute,
    AuditLogEnabled:          true,
    AuditLogger:              logger,
}
```

### Initialization
```go
middlewareManager := NewMiddlewareManager(userModule, logger, config)
verificationMiddleware := middlewareManager.GetVerificationMiddleware()
routeProtector := middlewareManager.GetRouteProtector()
```

### Route Protection
```go
// Basic authentication (allows unverified users)
router.GET("/api/user/profile", middleware.Auth(), userHandler.GetProfile)

// Email verification required
router.POST("/api/user/change-email", 
    middlewareManager.GetVerificationMiddleware().RequireEmailVerification(), 
    userHandler.ChangeEmail)

// Full verification required
router.POST("/api/user/place-bet", 
    middlewareManager.GetRouteProtector().Protect(BettingAccess), 
    betHandler.PlaceBet)

// Comprehensive protection
router.POST("/api/user/financial-transaction", 
    middlewareManager.CreateComprehensiveMiddleware(FinancialAccess), 
    financialHandler.ProcessTransaction)
```

## Error Handling & Responses

### Verification Required (403)
```json
{
  "code": 403,
  "message": "Account verification required for betting activities"
}
```

### Rate Limit Exceeded (429)
```json
{
  "code": 429,
  "message": "Too many access attempts. Please try again later."
}
```

### Authentication Required (401)
```json
{
  "code": 401,
  "message": "User authentication required"
}
```

## Monitoring & Observability

### Audit Logs
- **User Actions**: Complete user action tracking
- **Verification Attempts**: All verification attempts logged
- **Access Patterns**: Route access pattern analysis
- **Security Events**: Security event monitoring
- **Compliance Tracking**: Regulatory compliance monitoring

### Metrics
- **Verification Success Rates**: Verification completion rates
- **Access Denial Rates**: Access denial tracking
- **Rate Limit Violations**: Rate limiting effectiveness
- **Performance Metrics**: Middleware performance tracking

## Security Considerations

### Rate Limiting
- **Per-User Limits**: Individual user rate limiting
- **Per-IP Limits**: IP-based rate limiting
- **Configurable Windows**: Adjustable time windows
- **Distributed Limiting**: Redis-based distributed rate limiting

### Audit & Compliance
- **Complete Logging**: Full request/response logging
- **User Tracking**: Complete user action tracking
- **IP Logging**: Full IP address logging
- **Timestamp Precision**: Microsecond precision timing
- **Request Correlation**: Request ID correlation

### Verification Security
- **OTP Expiration**: Configurable OTP expiration
- **Resend Limits**: Limited verification resend attempts
- **Verification Attempts**: Limited verification attempts
- **Account Lockout**: Temporary account lockout on failures

## Testing & Quality Assurance

### Test Scenarios
1. **Registration Validation**: Duplicate constraint testing
2. **Verification Flow**: Complete verification testing
3. **Access Control**: Route protection testing
4. **Rate Limiting**: Rate limiting effectiveness
5. **Error Handling**: Comprehensive error testing
6. **Audit Logging**: Logging accuracy testing

### Performance Testing
- **Middleware Performance**: Middleware execution time
- **Concurrent Users**: High-concurrency testing
- **Rate Limiting**: Rate limiting performance
- **Memory Usage**: Memory utilization testing

## Future Enhancements

### Advanced Security
- **Multi-Factor Authentication**: Enhanced MFA support
- **Biometric Verification**: Biometric verification integration
- **Blockchain Verification**: Blockchain-based verification
- **AI-Powered Fraud Detection**: Machine learning fraud detection

### Compliance & Governance
- **GDPR Compliance**: Enhanced privacy controls
- **SOC 2 Compliance**: Security compliance framework
- **Regulatory Reporting**: Automated compliance reporting
- **Data Retention**: Configurable data retention policies

### Performance & Scalability
- **Horizontal Scaling**: Multi-instance deployment
- **Load Balancing**: Intelligent load balancing
- **Caching Layers**: Multi-level caching
- **Database Optimization**: Query optimization and indexing

## Production Checklist

### Security
- [ ] Rate limiting configured and tested
- [ ] Audit logging enabled and verified
- [ ] Error messages sanitized
- [ ] IP filtering configured
- [ ] SSL/TLS enabled

### Performance
- [ ] Middleware performance tested
- [ ] Rate limiting performance verified
- [ ] Memory usage optimized
- [ ] Database queries optimized
- [ ] Caching configured

### Monitoring
- [ ] Health checks implemented
- [ ] Metrics collection enabled
- [ ] Alerting configured
- [ ] Log aggregation enabled
- [ ] Performance monitoring active

### Compliance
- [ ] Audit logs verified
- [ ] Data retention configured
- [ ] Privacy controls implemented
- [ ] Regulatory requirements met
- [ ] Compliance reporting enabled 