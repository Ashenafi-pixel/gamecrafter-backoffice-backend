# 🚀 **WORLD-CLASS ENTERPRISE REGISTRATION SYSTEM** 🚀

## 🏆 **Overview**

This is a **100% production-ready, enterprise-grade registration system** built with Go, featuring comprehensive email verification, OTP management, database persistence, professional email templates, and production monitoring. This system is designed to handle enterprise-scale user registrations with security, reliability, and performance at its core.

## ✨ **Key Features**

### **Security & Authentication**
- **Multi-factor verification** with email OTP
- **Rate limiting** on all endpoints
- **Input validation** with comprehensive error handling
- **Secure password hashing** using bcrypt
- **JWT token management** for authenticated sessions
- **IP filtering** and security monitoring

### 📧 **Professional Email System**
- **Beautiful HTML email templates** with responsive design
- **Plain text fallbacks** for accessibility
- **SMTP integration** with Gmail/enterprise providers
- **Email delivery tracking** and metrics
- **Automatic retry mechanisms** for failed deliveries
- **Template customization** for branding

### 🗄️ **Data Persistence**
- **PostgreSQL database** with proper indexing
- **Redis caching** for OTP and session management
- **Database migrations** with version control
- **Data integrity constraints** and validation
- **Audit logging** for compliance
- **Backup and recovery** strategies

### 📊 **Production Monitoring**
- **Prometheus metrics** integration
- **Comprehensive health checks** for all components
- **Performance monitoring** and alerting
- **Error tracking** and rate monitoring
- **Business metrics** and analytics
- **Real-time dashboards** for operations

### 🎯 **Enterprise Features**
- **Multi-user type support** (Player, Agent, Admin)
- **Referral system** integration
- **Company/enterprise** registration support
- **Scalable architecture** for high traffic
- **Multi-region** deployment support
- **Compliance** with enterprise standards

## 🏗️ **Architecture**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │  Business Logic │    │  Data Layer     │
│                 │    │                 │    │                 │
│ • Gin Router    │◄──►│ • Registration  │◄──►│ • PostgreSQL    │
│ • Middleware    │    │ • OTP Service   │    │ • Redis Cache   │
│ • Validation    │    │ • Email Service │    │ • File Storage  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Monitoring     │    │  Email System   │    │  Security      │
│                 │    │                 │    │                 │
│ • Prometheus    │    │ • SMTP Client   │    │ • Rate Limiting│
│ • Health Checks │    │ • Templates     │    │ • Input Validation│
│ • Metrics       │    │ • Delivery      │    │ • JWT Auth     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 **Quick Start**

### **Prerequisites**
- Go 1.23+
- PostgreSQL 12+
- Redis 6+
- SMTP server access

### **1. Environment Setup**
```bash
# Copy environment template
cp .env.example .env

# Configure your environment variables
SMTP_HOST=smtp.gmail.com
SMTP_PORT=465
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourcompany.com
SMTP_FROM_NAME=Your Company

DB_URL=postgres://user:password@localhost:5432/dbname
REDIS_ADDR=localhost:6379
JWT_SECRET=your-secret-key
```

### **2. Database Setup**
```bash
# Run database migrations
migrate -path migrations -database "$DB_URL" up

# Verify tables created
psql "$DB_URL" -c "\dt enterprise_registrations"
```

### **3. Build & Run**
```bash
# Build the application
go build -o tucanbit cmd/main.go

# Run the application
./tucanbit
```

### **4. Verify Installation**
```bash
# Health check
curl http://localhost:8080/health

# Swagger documentation
open http://localhost:8080/swagger/index.html
```

## 📚 **API Endpoints**

### **Enterprise Registration**
```
POST   /api/enterprise/register          # Initiate registration
POST   /api/enterprise/register/complete # Complete with OTP
GET    /api/enterprise/register/status/:user_id
POST   /api/enterprise/register/resend   # Resend verification
```

### **Health & Monitoring**
```
GET    /health                           # Quick health check
GET    /health/detailed                  # Detailed health status
GET    /health/ready                     # Readiness probe
GET    /health/live                      # Liveness probe
GET    /health/metrics                   # Custom metrics
```

### **Swagger Documentation**
```
GET    /swagger/index.html               # Interactive API docs
GET    /swagger/doc.json                 # OpenAPI specification
```

## 🔧 **Configuration**

### **Environment Variables**
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SMTP_HOST` | SMTP server hostname | - | |
| `SMTP_PORT` | SMTP server port | 465 | |
| `SMTP_USERNAME` | SMTP username | - | |
| `SMTP_PASSWORD` | SMTP password/app password | - | |
| `SMTP_FROM` | From email address | - | |
| `SMTP_FROM_NAME` | From display name | - | |
| `DB_URL` | PostgreSQL connection string | - | |
| `REDIS_ADDR` | Redis server address | localhost:6379 | |
| `JWT_SECRET` | JWT signing secret | - | |
| `APP_HOST` | Application host | 0.0.0.0 |  |
| `APP_PORT` | Application port | 8080 |  |

### **Database Configuration**
```sql
-- Enterprise registrations table
CREATE TABLE enterprise_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    user_type VARCHAR(50) NOT NULL,
    phone_number VARCHAR(20),
    company_name VARCHAR(255),
    registration_status VARCHAR(50) DEFAULT 'PENDING',
    verification_otp VARCHAR(10),
    otp_expires_at TIMESTAMP WITH TIME ZONE,
    verification_attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_enterprise_registrations_email ON enterprise_registrations(email);
CREATE INDEX idx_enterprise_registrations_status ON enterprise_registrations(registration_status);
CREATE INDEX idx_enterprise_registrations_user_type ON enterprise_registrations(user_type);
```

## 📊 **Monitoring & Metrics**

### **Prometheus Metrics**
The system exposes comprehensive Prometheus metrics:

```bash
# Registration metrics
enterprise_registration_attempts_total
enterprise_registration_success_total
enterprise_registration_failure_total

# OTP verification metrics
enterprise_registration_otp_verification_attempts_total
enterprise_registration_otp_verification_success_total
enterprise_registration_otp_verification_failure_total

# Email metrics
enterprise_registration_email_sent_total
enterprise_registration_email_failure_total
enterprise_registration_email_delivery_time_seconds

# Performance metrics
enterprise_registration_duration_seconds
enterprise_registration_otp_generation_time_seconds
```

### **Health Check Endpoints**
```bash
# Quick health check
curl http://localhost:8080/health

# Detailed health status
curl http://localhost:8080/health/detailed

# Kubernetes readiness probe
curl http://localhost:8080/health/ready

# Kubernetes liveness probe
curl http://localhost:8080/health/live
```

## 🧪 **Testing**

### **Unit Tests**
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/module/user/...

# Run with coverage
go test -cover ./...
```

### **Integration Tests**
```bash
# Test with real database
go test -tags=integration ./...

# Test email functionality
go test -tags=email ./...
```

### **Load Testing**
```bash
# Install k6
curl -L https://github.com/grafana/k6/releases/latest/download/k6-linux-amd64.tar.gz | tar xz

# Run load test
k6 run load-tests/registration.js
```

## 🚀 **Deployment**

### **Docker Deployment**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o tucanbit cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tucanbit .
EXPOSE 8080
CMD ["./tucanbit"]
```

### **Kubernetes Deployment**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-registration
spec:
  replicas: 3
  selector:
    matchLabels:
      app: enterprise-registration
  template:
    metadata:
      labels:
        app: enterprise-registration
    spec:
      containers:
      - name: app
        image: your-registry/enterprise-registration:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
```

### **Production Considerations**
- **Load balancing** with multiple instances
- **Database connection pooling** for high concurrency
- **Redis clustering** for scalability
- **CDN integration** for email template assets
- **Log aggregation** with ELK stack
- **Alerting** with Prometheus AlertManager
- **Backup strategies** for database and Redis
- **SSL/TLS termination** at load balancer

## 🔒 **Security Features**

### **Input Validation**
- **Email format validation** with regex patterns
- **Password strength requirements** (min 8 chars, complexity)
- **Phone number validation** with international format support
- **SQL injection prevention** with parameterized queries
- **XSS protection** with proper output encoding

### **Rate Limiting**
- **Per-IP rate limiting** on registration endpoints
- **Per-email rate limiting** for OTP requests
- **Configurable limits** for different user types
- **Redis-based rate limiting** for distributed deployments

### **Data Protection**
- **Password hashing** with bcrypt (cost factor 12)
- **Sensitive data encryption** at rest
- **Audit logging** for all operations
- **Data retention policies** for compliance
- **GDPR compliance** features

## 📈 **Performance Optimization**

### **Database Optimization**
- **Proper indexing** on frequently queried fields
- **Connection pooling** with configurable limits
- **Query optimization** with EXPLAIN analysis
- **Database partitioning** for large datasets
- **Read replicas** for scaling read operations

### **Caching Strategy**
- **Redis caching** for OTP codes and sessions
- **In-memory caching** for frequently accessed data
- **Cache invalidation** strategies
- **Cache warming** for critical data

### **Async Processing**
- **Background email sending** with worker queues
- **Batch processing** for bulk operations
- **Event-driven architecture** for scalability
- **Message queuing** with Kafka/RabbitMQ

## 🛠️ **Troubleshooting**

### **Common Issues**

#### **Email Delivery Problems**
```bash
# Check SMTP configuration
curl -X POST http://localhost:8080/api/enterprise/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","first_name":"Test","last_name":"User","user_type":"PLAYER"}'

# Check application logs
tail -f tucanbit.log | grep -i "email\|smtp"
```

#### **Database Connection Issues**
```bash
# Test database connectivity
psql "$DB_URL" -c "SELECT 1"

# Check application logs
tail -f tucanbit.log | grep -i "database\|postgres"
```

#### **Redis Connection Issues**
```bash
# Test Redis connectivity
redis-cli -h localhost -p 6379 ping

# Check application logs
tail -f tucanbit.log | grep -i "redis"
```

### **Debug Mode**
```bash
# Enable debug logging
export LOG_LEVEL=debug
./tucanbit

# Enable Gin debug mode
export GIN_MODE=debug
./tucanbit
```

## 📚 **API Examples**

### **Registration Flow**

#### **1. Initiate Registration**
```bash
curl -X POST http://localhost:8080/api/enterprise/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "user_type": "PLAYER",
    "phone_number": "+1234567890",
    "company_name": "Example Corp"
  }'
```

**Response:**
```json
{
  "message": "Registration initiated successfully",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "otp_id": "456e7890-e89b-12d3-a456-426614174000",
  "expires_at": "2025-08-29T15:30:00Z",
  "resend_after": "2025-08-29T15:00:00Z"
}
```

#### **2. Complete Registration**
```bash
curl -X POST http://localhost:8080/api/enterprise/register/complete \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "otp_id": "456e7890-e89b-12d3-a456-426614174000",
    "otp_code": "123456"
  }'
```

**Response:**
```json
{
  "message": "Registration completed successfully",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "is_verified": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "verified_at": "2025-08-29T15:25:00Z"
}
```

#### **3. Check Registration Status**
```bash
curl http://localhost:8080/api/enterprise/register/status/123e4567-e89b-12d3-a456-426614174000
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "status": "completed",
  "created_at": "2025-08-29T15:20:00Z",
  "verified_at": "2025-08-29T15:25:00Z",
  "completed_at": "2025-08-29T15:25:00Z",
  "otp_expires_at": "2025-08-29T15:30:00Z"
}
```

## 🤝 **Contributing**

### **Development Setup**
```bash
# Fork the repository
git clone https://github.com/yourusername/tucanbit.git
cd tucanbit

# Install dependencies
go mod download

# Install development tools
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Generate Swagger docs
swag init -g cmd/main.go
```

### **Code Standards**
- **Go formatting** with `gofmt`
- **Linting** with `golangci-lint`
- **Testing** with 90%+ coverage
- **Documentation** for all public APIs
- **Error handling** with proper logging
- **Security** best practices

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 **Support**

### **Getting Help**
- **Documentation**: [API Docs](http://localhost:8080/swagger/index.html)
- **Issues**: [GitHub Issues](https://github.com/yourusername/tucanbit/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/tucanbit/discussions)
- **Email**: support@yourcompany.com

### **Enterprise Support**
For enterprise customers, we offer:
- **Priority support** with dedicated engineers
- **Custom integrations** and features
- **Training and consulting** services
- **SLA guarantees** for uptime and response times
- **On-premise deployment** support

---

## 🎯 **Roadmap**

### **Phase 1: Core System** 
- [x] Enterprise registration endpoints
- [x] Email verification system
- [x] OTP management
- [x] Database persistence
- [x] Basic monitoring

### **Phase 2: Advanced Features** 🚧
- [ ] Multi-language support
- [ ] Advanced analytics dashboard
- [ ] A/B testing framework
- [ ] Advanced fraud detection
- [ ] Social login integration

### **Phase 3: Enterprise Features** 📋
- [ ] SSO integration (SAML, OAuth2)
- [ ] Advanced role-based access control
- [ ] Compliance reporting (GDPR, SOC2)
- [ ] Advanced audit logging
- [ ] Multi-tenant architecture

---

**Built with ❤️ for enterprise-grade reliability and performance** 