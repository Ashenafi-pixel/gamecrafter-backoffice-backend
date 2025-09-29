# TucanBIT Daily Report Email System 📊

## Overview
The TucanBIT Daily Report Email System automatically generates and sends comprehensive daily analytics reports via email. It's designed to keep administrators informed about platform performance metrics on a daily basis.

## Features ✨

- **Automated Daily Reports**: Sends reports automatically every day at a configured time
- **Rich HTML Email**: Reports include beautiful charts, metrics, and detailed analytics
- **Multiple Recipients**: Support for sending to multiple email addresses
- **Flexible Scheduling**: Configurable time of day and timezone for sending
- **Manual Triggers**: Ability to send reports manually for specific dates
- **Historical Reports**: Send reports for previous dates or weekly summaries
- **Real-time Status**: Checks application availability before sending
- **Comprehensive Logging**: Full audit trail of email sending activities

## Email Credentials 📧
- **SMTP Host**: smtp.gmail.com
- **Port**: 465 (SSL)
- **Username**: ashenafialemu66@gmail.com  
- **Password**: dpfr mgcv tgnh skyo
- **From Address**: noreply@tucanbit.com

## Report Contents 📋

Each daily report includes:

### 📊 Key Metrics Cards
- **Total Transactions**: Count of all transactions for the day
- **Active Users**: Number of unique users who performed activities
- **New Users**: Number of new user registrations
- **Active Games**: Number of games played during the day

### 💰 Financial Overview Table
- Market Cap Revenue: Revenue from market transactions
- Total Deposits: User deposits processed
- Total Withdrawals: User withdrawals processed
- Total Bets: Amount wagered across all games
- Total Wins: Amount paid out to users
- **Net Revenue**: Calculated as Total Bets - Total Wins

### 🎮 Top Performing Games
- Game ranking by revenue and activity
- Player count and session statistics
- Average bet amounts and RTP calculations

### 👑 Top Players
- Player ranking by activity and spending
- Transaction counts and game diversity
- Average bet patterns and last activity

## API Endpoints 🌐

### Manual Email Triggering

#### Send Daily Report for Specific Date
```bash
POST /analytics/daily-report/send
Content-Type: application/json

{
  "date": "2025-01-15",
  "recipients": ["admin@example.com", "manager@example.com"]
}
```

#### Send Yesterday's Report
```bash
POST /analytics/daily-report/yesterday
Content-Type: application/json

{
  "recipients": ["ashenafialemu66@gmail.com"]
}
```

#### Send Last Week's Reports (7 days)
```bash
POST /analytics/daily-report/last-week
Content-Type: application/json

{
  "recipients": ["ashenafialemu66@gmail.com"]
}
```

#### Schedule Automatic Reports
```bash
POST /analytics/daily-report/schedule
Content-Type: application/json

{
  "recipients": ["ashenafialemu66@gmail.com"],
  "send_frequency": "daily",
  "time_of.day": 9,
  "timezone": "UTC",
  "auto_schedule": true
}
```

## Setup Instructions 🛠️

### 1. Automated Setup (Recommended)

Run the automated setup script:

```bash
# Make script executable
chmod +x scripts/setup_daily_report_cronjob.sh

# Run the setup script
sudo ./scripts/setup_daily_report_cronjob.sh
```

The script will:
- ✅ Check TucanBIT application status
- ✅ Test the email API endpoints
- ✅ Create a cron job script
- ✅ Setup automatic daily sending
- ✅ Configure log rotation
- ✅ Send a test report email

### 2. Manual Cron Setup

If you prefer manual setup:

#### Create Cron Script
```bash
sudo nano /usr/local/bin/tucanbit-daily-report
```

Add this content:
```bash
#!/bin/bash

# TucanBIT Daily Report Email Script
APP_URL="http://localhost:8080"
RECIPIENTS="ashenafialemu66@gmail.com"
LOG_FILE="/var/log/tucanbit_daily_reports.log"

# Get yesterday's date
REPORT_DATE=$(date -d "yesterday" "+%Y-%m-%d")

# Send daily report email
curl -s -X POST "$APP_URL/analytics/daily-report/send" \
    -H "Content-Type: application/json" \
    -d '{
        "date": "'$REPORT_DATE'",
        "recipients": ["'$RECIPIENTS'"]
    }' >> "$LOG_FILE" 2>&1

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Daily report sent for $REPORT_DATE" >> "$LOG_FILE"
```

Make it executable:
```bash
sudo chmod +x /usr/local/bin/tucanbit-daily-report
```

#### Add to Crontab
```bash
# Edit crontab
crontab -e

# Add this line to send report every day at 9:00 AM UTC
0 9 * * * /usr/local/bin/tucanbit-daily-report
```

#### Verify Cron Job
```bash
# Check if cron job is scheduled
crontab -l | grep tucanbit-daily-report
```

### 3. AWS EC2 Specific Setup

For AWS deployment:

#### Step 1: Deploy Code
```bash
# SSH into your EC2 instance
ssh ubuntu@your-ec2-ip

# Pull latest code
cd ~/tucanbit-backend
git pull origin develop

# Build and run the application
go build -o tucanbit ./cmd/main.go
./tucanbit
```

#### Step 2: Run Setup Script on AWS
```bash
# On AWS EC2 instance
cd ~/tucanbit-backend
sudo ./scripts/setup_daily_report_cronjob.sh
```

#### Step 3: Verify Application is Running
```bash
# Check if application is accessible
curl http://localhost:8080/analytics/realtime/stats

# Test daily report API
curl -X POST http://localhost:8080/analytics/daily-report/yesterday \
    -H "Content-Type: application/json" \
    -d '{"recipients": ["ashenafialemu66@gmail.com"]}'
```

## Testing 🧪

### Manual Test Commands

#### Test Real-time Stats (Health Check)
```bash
curl http://localhost:8080/analytics/realtime/stats
```

#### Test Send Yesterday's Report
```bash
curl -X POST http://localhost:8080/analytics/daily-report/yesterday \
    -H "Content-Type: application/json" \
    -d '{
        "recipients": ["ashenafialemu66@gmail.com"]
    }'
```

#### Test Send Specific Date Report
```bash
curl -X POST http://localhost:8080/analytics/daily-report/send \
    -H "Content-Type: application/json" \
    -d '{
        "date": "2025-09-26",
        "recipients": ["ashenafialemu66@gmail.com"]
    }'
```

#### Test Last Week Reports
```bash
curl -X POST http://localhost:8080/analytics/daily-report/last-week \
    -H "Content-Type: application/json" \
    -d '{
        "recipients": ["ashenafialemu66@gmail.com"]
    }'
```

### Expected Response Format
```json
{
  "success": true,
  "data": {
    "message": "Daily report email sent successfully",
    "date": "2025-09-26",
    "recipients_count": 1,
    "recipients": ["ashenafialemu66@gmail.com"]
  }
}
```

## Monitoring 📈

### View Logs
```bash
# Real-time log monitoring
tail -f /var/log/tucanbit_daily_reports.log

# View recent logs
grep "$(date '+%Y-%m-%d')" /var/log/tucanbit_daily_reports.log
```

### Check Cron Job Status
```bash
# List current cron jobs
crontab -l

# Check cron service status
sudo systemctl status cron

# View cron logs
sudo journalctl -u cron -f
```

### Application Health Check
```bash
# Check if TucanBIT is running
curl http://localhost:8080/analytics/realtime/stats

# Check port usage
netstat -tlnp | grep 8080

# Check application process
ps aux | grep tucanbit
```

## Troubleshooting 🔧

### Common Issues

#### Issue 1: Application Not Running
**Symptoms**: Email sending fails, HTTP 000 or 503 errors
**Solutions**:
```bash
# Check if TucanBIT is running
ps aux | grep tucanbit

# Kill existing processes and restart
pkill -f tucanbit
./tucanbit

# Run in background
nohup ./tucanbit > tucanbit.log 2>&1 &
```

#### Issue 2: SMTP Authentication Failed
**Symptoms**: Email service initialization fails
**Solutions**:
- Verify email credentials in `config/config`
- Check Gmail app passwords: https://support.google.com/accounts/answer/185833
- Ensure 2FA is enabled for Gmail account

#### Issue 3: ClickHouse Not Available
**Symptoms**: Daily reports return empty data
**Solutions**:
```bash
# Check ClickHouse status
docker ps | grep clickhouse

# Restart ClickHouse
docker restart tucanbit-clickhouse

# Check ClickHouse logs
docker logs tucanbit-clickhouse
```

#### Issue 4: Cron Job Not Running
**Symptoms**: No daily emails received
**Solutions**:
```bash
# Check cron service
sudo systemctl status cron
sudo systemctl start cron

# Test cron script manually
/usr/local/bin/tucanbit-daily-report

# Check cron logs for errors
sudo journalctl -u cron -n 50
```

#### Issue 5: No Data in Reports
**Symptoms**: Reports contain zero values or empty sections
**Solutions**:
- Verify ClickHouse has data: `docker exec tucanbit-clickhouse clickhouse-client --query "SELECT COUNT(*) FROM tucanbit_analytics.transactions;"`
- Check PostgreSQL data: `docker exec tucanbit-db psql -U tucanbit -d tucanbit -c "SELECT COUNT(*) FROM users;"`
- Ensure real-time sync is working: Check application logs for analytics sync errors

### Log Analysis Examples

#### Successful Email Send
```
[2025-09-26 09:00:01] Starting daily report generation for date: 2025-09-25
[2025-09-26 09:00:15] SUCCESS: Daily report email sent for 2025-09-25
[2025-09-26 09:00:15] Recipients: ashenafialemu66@gmail.com
[2025-09-26 09:00:15] Daily report cron job completed successfully
```

#### Failed Email Send
```
[2025-09-26 09:00:01] Starting daily report generation for date: 2025-09-25
[2025-09-26 09:00:05] ERROR: TucanBIT application not accessible (HTTP 503)
[2025-09-26 09:00:05] Daily report email NOT sent for 2025-09-25
```

## Configuration Options ⚙️

### Environment Variables
```bash
# SMTP Configuration
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="465"
export SMTP_USERNAME="ashenafialemu66@gmail.com"
export SMTP_PASSWORD="dpfr mgcv tgnh skyo"
export SMTP_FROM="noreply@tucanbit.com"
export SMTP_FROM_NAME="TucanBIT"

# Application Configuration
export APP_HOST="0.0.0.0"
export APP_PORT="8080"
```

### Cron Schedule Examples
```bash
# Send every day at 9:00 AM UTC
0 9 * * * /usr/local/bin/tucanbit-daily-report

# Send Monday to Friday at 8:00 AM UTC
0 8 * * 1-5 /usr/local/bin/tucanbit-daily-report

# Send every day at 6:00 PM UTC
0 18 * * * /usr/local/bin/tucanbit-daily-report

# Send twice daily (morning and evening)
0 9 * * * /usr/local/bin/tucanbit-daily-report
0 18 * * * /usr/local/bin/tucanbit-daily-report
```

## Email Template Customization 🎨

The email templates are located in:
- `internal/module/email/daily_report.go`: Main template logic
- `internal/module/email/email.go`: Base email template

### Customizing Styling
Edit the CSS in `GetDailyReportEmailTemplate()`:

```go
// Color scheme examples
metrics-grid .metric-card.blue { background: linear-gradient(135deg, #3498db, #2980b9); }
metrics-grid .metric-card.green { background: linear-gradient(135deg, #27ae60, #229954); }
metrics-grid .metric-card.purple { background: linear-gradient(135deg, #9b59b6, #8e44ad); }
```

### Adding New Metrics
Edit `generateDailyReportHTML()` in `daily_report.go`:

```go
// Add custom metric card
metric-card .custom-metric {
    // Your custom styling
}
```

## Security Best Practices 🔒

1. **Email Credentials**: Store SMTP credentials securely
2. **Access Control**: Restrict API access to authorized IPs if needed
3. **Rate Limiting**: Implement rate limiting for email endpoints
4. **Input Validation**: All email inputs are validated before processing
5. **Error Handling**: Secure error messages don't leak sensitive information

## Performance Considerations ⚡

1. **Batch Processing**: Reports are generated efficiently with optimized ClickHouse queries
2. **Email Batching**: Multiple recipients receive emails concurrently
3. **Resource Management**: Proper cleanup of email connections
4. **Logging Optimization**: Efficient logging to prevent disk space issues
5. **Cache Utilization**: ClickHouse performance optimizations

## Support & Maintenance 🔧

### Regular Maintenance Tasks

#### Weekly
- Review email delivery logs
- Check recipient feedback
- Monitor ClickHouse performance

#### Monthly  
- Analyze email open rates
- Review report content effectiveness
- Update recipient lists if needed

#### Quarterly
- Performance optimization review
- Security audit of email system
- Template design updates

### Getting Help

If you encounter issues:

1. **Check the logs**: `/var/log/tucanbit_daily_reports.log`
2. **Test manually**: Use the curl commands in the testing section
3. **Verify setup**: Confirm cron job and application status
4. **Review configuration**: Check email credentials and SMTP settings

---

## Quick Reference Card 🃏

### Essential Commands

```bash
# Test application
curl http://localhost:8080/analytics/realtime/stats

# Send test report
curl -X POST http://localhost:8080/analytics/daily-report/yesterday \
    -H "Content-Type: application/json" \
    -d '{"recipients": ["ashenafialemu66@gmail.com"]}'

# Check cron jobs
crontab -l | grep tucanbit

# View logs
tail -f /var/log/tucanbit_daily_reports.log

# Restart application
pkill -f tucanbit && ./tucanbit
```

### Troubleshooting Checklist

- [ ] TucanBIT application running on port 8080
- [ ] ClickHouse database accessible and populated
- [ ] SMTP credentials valid and working
- [ ] Cron service running
- [ ] Cron script executable
- [ ] Sufficient disk space for logs
- [ ] Internet connectivity for SMTP

---

**🎉 Congratulations! Your TucanBIT Daily Report Email System is now ready to keep you informed about your platform's performance automatically every day!**