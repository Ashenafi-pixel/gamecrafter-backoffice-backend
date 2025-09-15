# TucanBIT GrooveTech AWS Migration Guide

## 🚀 Overview

This guide provides comprehensive instructions for migrating your TucanBIT GrooveTech integration from local development to AWS production server. The migration includes all database schemas, data, and configurations necessary for the GrooveTech APIs to function properly.

## 📋 Prerequisites

### AWS Services Required

1. **Amazon RDS PostgreSQL**
   - PostgreSQL 13+ instance
   - Multi-AZ deployment for production
   - Automated backups enabled
   - Security groups configured

2. **Amazon ElastiCache Redis** (Optional)
   - Redis cluster for caching
   - Subnet group configured

3. **Amazon MSK Kafka** (Optional)
   - Kafka cluster for event streaming
   - VPC configuration

4. **Amazon S3**
   - Bucket for file storage
   - IAM policies configured

5. **EC2 Instance**
   - Ubuntu 20.04+ or Amazon Linux 2
   - Docker and Docker Compose installed
   - Security groups allowing HTTP/HTTPS traffic

## 🛠️ Migration Methods

### Method 1: Direct Migration (Recommended for Production)

```bash
# Run the migration preparation script
./migrate_to_aws.sh

# Set your AWS database URL
export AWS_DB_URL="postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require"

# Run the migration
psql "$AWS_DB_URL" -f aws_migration_backup/aws_migration_TIMESTAMP.sql
```

### Method 2: Docker Migration (Recommended for Development/Testing)

```bash
# Run the Docker migration preparation script
./docker_migrate_to_aws.sh

# Set your AWS database URL
export AWS_DB_URL="postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require"

# Deploy using Docker
cd aws_migration_backup
./deploy_to_aws.sh
```

## 📁 Migration Files Generated

The migration scripts create the following files in `aws_migration_backup/`:

### Database Files
- `tucanbit_schema_TIMESTAMP.sql` - Complete database schema
- `groovetech_data_TIMESTAMP.sql` - GrooveTech specific data
- `tucanbit_data_TIMESTAMP.sql` - All database data
- `aws_migration_TIMESTAMP.sql` - Complete AWS migration script

### Configuration Files
- `aws_config_template.yaml` - Application configuration template
- `.env.aws` - Environment variables template
- `docker-compose-aws.yml` - Docker Compose configuration

### Deployment Files
- `deploy_to_aws.sh` - Automated deployment script
- `AWS_DEPLOYMENT_INSTRUCTIONS.md` - Detailed deployment guide

## 🗄️ Database Schema Migration

### GrooveTech Tables Created

1. **groove_accounts**
   - Stores GrooveTech player accounts
   - Links to existing users table
   - Tracks balances and status

2. **groove_transactions**
   - Stores all GrooveTech transactions
   - Implements idempotency
   - Tracks transaction history

3. **game_sessions**
   - Tracks active game sessions
   - Manages session expiration
   - Links to GrooveTech accounts

4. **groove_game_sessions**
   - GrooveTech specific session data
   - Game-specific information
   - Session management

### Indexes and Performance

- **Primary indexes** on all foreign keys
- **Composite indexes** for common queries
- **Time-based indexes** for transaction history
- **Status indexes** for filtering

### Functions and Triggers

- **Session ID generation** functions
- **Automatic timestamp updates**
- **Session cleanup** functions
- **Account summary** functions

## ⚙️ Configuration Migration

### Required Configuration Updates

1. **Database Connection**
   ```yaml
   db:
     url: postgres://username:password@your-aws-rds-endpoint:5432/tucanbit?sslmode=require
   ```

2. **GrooveTech Settings**
   ```yaml
   groove:
     operator_id: "3818"
     api_domain: "https://routerstg.groovegaming.com"
     api_key: "your_groove_api_key_here"
     home_url: "https://your-domain.com"
     exit_url: "https://your-domain.com"
     history_url: "https://your-domain.com/history"
     license_type: "Curacao"
     signature_validation: true
     signature_secret: "your_signature_secret_here"
   ```

3. **AWS Services**
   ```yaml
   aws:
     bucket:
       name: your_s3_bucket_name
       accessKey: your_aws_access_key
       secretAccessKey: your_aws_secret_key
       region: your_aws_region
   ```

## 🚀 Deployment Steps

### Step 1: Prepare AWS Environment

1. **Create RDS PostgreSQL Instance**
   ```bash
   # Example RDS creation (use AWS CLI or Console)
   aws rds create-db-instance \
     --db-instance-identifier tucanbit-db \
     --db-instance-class db.t3.micro \
     --engine postgres \
     --engine-version 13.7 \
     --master-username tucanbit \
     --master-user-password YourSecurePassword \
     --allocated-storage 20 \
     --vpc-security-group-ids sg-xxxxxxxxx
   ```

2. **Create ElastiCache Redis Cluster** (Optional)
   ```bash
   aws elasticache create-cache-cluster \
     --cache-cluster-id tucanbit-redis \
     --cache-node-type cache.t3.micro \
     --engine redis \
     --num-cache-nodes 1
   ```

3. **Create S3 Bucket**
   ```bash
   aws s3 mb s3://tucanbit-files
   ```

### Step 2: Run Migration

1. **Export current data**
   ```bash
   ./migrate_to_aws.sh
   ```

2. **Set AWS database URL**
   ```bash
   export AWS_DB_URL="postgres://tucanbit:YourSecurePassword@tucanbit-db.xxxxxxxxx.us-east-1.rds.amazonaws.com:5432/tucanbit?sslmode=require"
   ```

3. **Run migration**
   ```bash
   psql "$AWS_DB_URL" -f aws_migration_backup/aws_migration_TIMESTAMP.sql
   ```

### Step 3: Deploy Application

1. **Update configuration**
   ```bash
   cp aws_migration_backup/aws_config_template.yaml config/production.yaml
   # Edit production.yaml with your AWS credentials
   ```

2. **Deploy with Docker**
   ```bash
   cd aws_migration_backup
   ./deploy_to_aws.sh
   ```

## 🧪 Testing and Verification

### Health Check
```bash
curl -f http://your-domain.com/health
```

### GrooveTech API Tests

1. **Game Launch**
   ```bash
   curl -X POST "https://your-domain.com/api/groove/launch-game" \
     -H "Content-Type: application/json" \
     -d '{
       "game_id": "82695",
       "device_type": "desktop",
       "game_mode": "real",
       "country": "US",
       "currency": "USD",
       "language": "en_US",
       "is_test_account": false,
       "reality_check_elapsed": 0,
       "reality_check_interval": 60
     }'
   ```

2. **Get Account**
   ```bash
   curl "https://your-domain.com/groove-official/getaccount?request=getaccount&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&nogsgameid=82695&apiversion=1.2"
   ```

3. **Get Balance**
   ```bash
   curl "https://your-domain.com/groove-official/balance?request=getbalance&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&nogsgameid=82695&apiversion=1.2"
   ```

4. **Wager Transaction**
   ```bash
   curl "https://your-domain.com/groove-official/wager?request=wager&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&gameid=82695&apiversion=1.2&betamount=10.0&roundid=test_round&transactionid=test_tx"
   ```

5. **Result Transaction**
   ```bash
   curl "https://your-domain.com/groove-official/result?request=result&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&gameid=82695&apiversion=1.2&result=15.0&roundid=test_round&transactionid=test_tx&gamestatus=completed"
   ```

6. **Rollback Transaction**
   ```bash
   curl "https://your-domain.com/groove-official/rollback?request=rollback&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&gameid=82695&apiversion=1.2&rollbackamount=10.0&roundid=test_round&transactionid=test_tx"
   ```

7. **Jackpot Transaction**
   ```bash
   curl "https://your-domain.com/groove-official/jackpot?request=jackpot&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&gameid=82695&apiversion=1.2&amount=50.0&roundid=jackpot_round&transactionid=jackpot_tx&gamestatus=completed"
   ```

8. **Rollback On Result**
   ```bash
   curl "https://your-domain.com/groove-official/reversewin?request=reversewin&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&gameid=82695&apiversion=1.2&amount=10.0&roundid=reverse_round&transactionid=reverse_tx"
   ```

9. **Rollback On Rollback**
   ```bash
   curl "https://your-domain.com/groove-official/rollbackrollback?request=rollbackrollback&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=Tucan_362a6ddd-eaf0-41f2-9a69-e64757c50cd7&device=desktop&gameid=82695&apiversion=1.2&rollbackAmount=10.0&roundid=rollback_round&transactionid=rollback_tx"
   ```

## 🔒 Security Considerations

### Database Security
- Use SSL/TLS connections (`sslmode=require`)
- Implement proper IAM roles
- Use VPC security groups
- Enable encryption at rest

### Application Security
- Use environment variables for secrets
- Implement proper logging
- Use HTTPS in production
- Enable signature validation

### API Security
- Use strong signature secrets
- Implement rate limiting
- Validate all input parameters
- Monitor for suspicious activity

## 📊 Monitoring and Maintenance

### Database Monitoring
- Monitor connection counts
- Track query performance
- Set up alerts for failures
- Monitor disk usage

### Application Monitoring
- Monitor API response times
- Track error rates
- Set up health checks
- Monitor memory usage

### Maintenance Tasks
- Regular database backups
- Cleanup expired sessions
- Monitor transaction logs
- Update security patches

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Check security group settings
   - Verify connection string
   - Check SSL certificate
   - Verify credentials

2. **API Issues**
   - Verify signature validation
   - Check parameter validation
   - Review logs for errors
   - Test with Postman collection

3. **Performance Issues**
   - Check database indexes
   - Monitor query performance
   - Review connection pooling
   - Check resource utilization

### Log Analysis
```bash
# Check application logs
docker logs tucanbit-app

# Check database logs
aws logs describe-log-groups --log-group-name-prefix /aws/rds/tucanbit-db

# Check system logs
journalctl -u tucanbit -f
```

## 📞 Support

For issues or questions:
1. Check application logs
2. Review database logs
3. Test with provided Postman collection
4. Contact development team

## 🎯 Success Criteria

Migration is successful when:
- ✅ All GrooveTech tables are created
- ✅ All indexes are in place
- ✅ All functions and triggers work
- ✅ Application starts without errors
- ✅ All GrooveTech APIs respond correctly
- ✅ Database connections are stable
- ✅ Performance is acceptable
- ✅ Security measures are in place

---

**Note**: This migration guide assumes you have the necessary AWS permissions and resources. Always test in a staging environment before deploying to production.