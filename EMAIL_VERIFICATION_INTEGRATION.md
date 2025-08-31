# 🚀 TucanBIT Email Verification System - Integration Guide

## 📋 **Overview**

This guide explains how to integrate the enterprise-grade email verification system into your existing TucanBIT application. The system provides:

- ✅ **Professional email templates** with TucanBIT branding
- ✅ **Secure OTP generation** and validation
- ✅ **Redis-based storage** for scalability
- ✅ **Comprehensive API endpoints** with Swagger documentation
- ✅ **Professional logging** and error handling
- ✅ **Rate limiting** and security features

## 🏗️ **Architecture**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   OTP Module    │
│                 │    │                 │    │                 │
│ Registration    │───▶│ /api/register   │───▶│ Email Service   │
│ Form            │    │ /api/otp/*      │    │ OTP Storage    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   User Module   │    │   Redis Cache   │
                       │                 │    │                 │
                       │ Account         │    │ OTP Data       │
                       │ Creation        │    │ Session Data   │
                       └─────────────────┘    └─────────────────┘
```

## 🔧 **Installation Steps**

### **Step 1: Database Setup**

Run the database migration to create the OTP table:

```bash
# Apply the migration
psql -U postgres -d tucanbit -f migrations/20250128080000_create_otps_table.up.sql

# Verify the table was created
psql -U postgres -d tucanbit -c "\d otps"
```

### **Step 2: Environment Configuration**

1. Copy the configuration template:
```bash
cp config/email_verification.env .env
```

2. Update the `.env` file with your actual values:
```bash
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@tucanbit.com
SMTP_FROM_NAME=TucanBIT
SMTP_USE_TLS=true

# Redis Configuration
REDIS_URL=redis://localhost:6379
```

### **Step 3: Redis Setup**

Ensure Redis is running and accessible:

```bash
# Start Redis (if not already running)
redis-server

# Test Redis connection
redis-cli ping
# Should return: PONG
```

### **Step 4: Update Main Application**

Add the new modules to your main application:

```go
// In your main.go or initiator
import (
    "github.com/tucanbit/internal/module/otp"
    "github.com/tucanbit/internal/module/email"
    "github.com/tucanbit/internal/handler/user"
)

func main() {
    // Initialize email service
    emailConfig := email.LoadSMTPConfigFromEnv()
    emailService, err := email.NewEmailService(emailConfig, logger)
    if err != nil {
        log.Fatal("Failed to initialize email service:", err)
    }

    // Initialize OTP module
    otpStorage := otp.NewRedisOTP(redisClient, logger)
    otpModule := otp.NewOTPService(otpStorage, userModule, emailService, logger)

    // Initialize registration service
    registrationService := user.NewRegistrationService(
        userModule, 
        otpModule, 
        emailService, 
        redisClient, 
        logger,
    )

    // Add routes
    router.POST("/api/register", registrationService.InitiateUserRegistration)
    router.POST("/api/register/complete", registrationService.CompleteUserRegistration)
    router.POST("/api/register/resend-verification", registrationService.ResendVerificationEmail)
    
    // OTP routes
    router.POST("/api/otp/email-verification", otpHandler.CreateEmailVerification)
    router.POST("/api/otp/verify", otpHandler.VerifyOTP)
    router.POST("/api/otp/resend", otpHandler.ResendOTP)
    router.GET("/api/otp/:otp_id", otpHandler.GetOTPInfo)
    router.DELETE("/api/otp/:otp_id/invalidate", otpHandler.InvalidateOTP)
}
```

## 📧 **Email Configuration**

### **Gmail Setup (Recommended for Development)**

1. Enable 2-Factor Authentication on your Gmail account
2. Generate an App Password:
   - Go to Google Account settings
   - Security → 2-Step Verification → App passwords
   - Generate password for "Mail"
3. Use the generated password in your `.env` file

### **Production Email Services**

For production, consider using:
- **SendGrid**: Excellent deliverability, good pricing
- **AWS SES**: Cost-effective, high deliverability
- **Mailgun**: Developer-friendly, good API

## 🧪 **Testing the System**

### **Run the Test Script**

```bash
# Make the script executable
chmod +x test_email_verification.sh

# Run the tests
./test_email_verification.sh
```

### **Manual Testing**

1. **Test Registration Initiation**:
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "phone_number": "+1234567890",
    "first_name": "Test",
    "last_name": "User",
    "password": "SecurePass123!",
    "type": "PLAYER"
  }'
```

2. **Test OTP Creation**:
```bash
curl -X POST http://localhost:8080/api/otp/email-verification \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
```

3. **Check Email**: Look for the verification email in your inbox

## 🔐 **Security Features**

### **Rate Limiting**
- OTP requests: 10 per minute per email
- Registration attempts: 5 per minute per IP
- Verification attempts: 3 per OTP

### **OTP Security**
- 6-digit codes with 10-minute expiration
- One-time use only
- Secure random generation using crypto/rand

### **Data Protection**
- Passwords hashed using bcrypt
- JWT tokens with configurable expiration
- Redis data with automatic expiration

## 📊 **Monitoring and Logging**

### **Log Files**
- Application logs: `logs/app.log`
- Email verification logs: `logs/email_verification.log`
- OTP operation logs: `logs/otp.log`

### **Health Checks**
```bash
# Check system health
curl http://localhost:8080/health

# Check OTP system status
curl http://localhost:8080/api/admin/otp/stats
```

### **Metrics**
- OTP creation rate
- Verification success rate
- Email delivery status
- Response times

## 🚀 **Production Deployment**

### **Environment Variables**
```bash
# Production SMTP
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key

# Production Redis
REDIS_URL=redis://your-redis-cluster:6379
REDIS_PASSWORD=your-redis-password

# Security
JWT_SECRET=your-super-secure-production-jwt-secret
```

### **Scaling Considerations**
- Use Redis Cluster for high availability
- Implement email queue for high volume
- Use CDN for email template assets
- Monitor email delivery rates

## 🐛 **Troubleshooting**

### **Common Issues**

1. **Email Not Sending**:
   - Check SMTP credentials
   - Verify firewall settings
   - Check email service quotas

2. **OTP Not Working**:
   - Verify Redis connection
   - Check OTP expiration
   - Validate email format

3. **Registration Failing**:
   - Check database connection
   - Verify user table schema
   - Check validation rules

### **Debug Mode**
Enable debug logging:
```bash
LOG_LEVEL=debug
```

### **Support**
For technical support:
- Email: support@tucanbit.com
- Documentation: `/docs` endpoint
- Logs: Check application logs

## 📈 **Performance Optimization**

### **Caching Strategy**
- OTP data in Redis with TTL
- User session data in Redis
- Email templates in memory

### **Database Optimization**
- Indexes on frequently queried fields
- Connection pooling
- Query optimization

### **Email Optimization**
- Template caching
- Batch processing for high volume
- Delivery tracking

## 🔄 **API Reference**

### **Registration Endpoints**
- `POST /api/register` - Initiate registration
- `POST /api/register/complete` - Complete registration
- `POST /api/register/resend-verification` - Resend verification

### **OTP Endpoints**
- `POST /api/otp/email-verification` - Create OTP
- `POST /api/otp/verify` - Verify OTP
- `POST /api/otp/resend` - Resend OTP
- `GET /api/otp/:otp_id` - Get OTP info
- `DELETE /api/otp/:otp_id/invalidate` - Invalidate OTP

### **Admin Endpoints**
- `POST /api/admin/otp/cleanup` - Cleanup expired OTPs
- `GET /api/admin/otp/stats` - Get OTP statistics

## 🎯 **Next Steps**

1. **Complete Integration**: Wire up all modules in your main application
2. **Database Migration**: Run the OTP table creation script
3. **Configuration**: Set up your SMTP and Redis credentials
4. **Testing**: Run the comprehensive test script
5. **Deployment**: Deploy to staging environment
6. **Production**: Deploy to production with proper monitoring

## 📞 **Support and Maintenance**

- **Regular Updates**: Keep dependencies updated
- **Security Patches**: Monitor for security updates
- **Performance Monitoring**: Track system metrics
- **Backup Strategy**: Regular database and Redis backups

---

**🎉 Congratulations!** You now have a production-ready, enterprise-grade email verification system integrated into your TucanBIT application.

For additional support or questions, please contact the development team or refer to the API documentation at `/docs`. 