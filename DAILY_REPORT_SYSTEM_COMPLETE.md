# 📊 TucanBIT Daily Report System - Complete Implementation

## 🎯 Overview

The TucanBIT Daily Report System provides automated daily analytics reports sent via email to configured recipients. The system includes both manual API endpoints and automatic cronjob scheduling.

## 📧 Configured Recipients

The following email addresses are configured to receive daily reports:
- **ashenafialemu27@gmail.com**
- **johsjones612@gmail.com**

## ⏰ Automatic Scheduling

- **Schedule**: Every day at **23:59 UTC** (end of day)
- **Frequency**: Daily
- **Timezone**: UTC
- **Report Period**: Previous day's data

## 🚀 API Endpoints

### 1. Send Configured Daily Report
**Endpoint**: `POST /analytics/daily-report/send-configured`

Send daily report to configured recipients (ashenafialemu27@gmail.com, johsjones612@gmail.com).

```bash
curl -X POST http://localhost:8080/analytics/daily-report/send-configured \
  -H "Content-Type: application/json" \
  -d '{"date": "2025-01-15"}'
```

**Request Body** (optional):
```json
{
  "date": "2025-01-15"  // Optional, defaults to yesterday
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "message": "Configured daily report email sent successfully",
    "date": "2025-01-15",
    "recipients_count": 2,
    "recipients": ["ashenafialemu27@gmail.com", "johsjones612@gmail.com"],
    "note": "Email sent to configured recipients"
  }
}
```

### 2. Send Test Daily Report
**Endpoint**: `POST /analytics/daily-report/test`

Send a test daily report to configured recipients for verification.

```bash
curl -X POST http://localhost:8080/analytics/daily-report/test \
  -H "Content-Type: application/json"
```

**Response**:
```json
{
  "success": true,
  "data": {
    "message": "Test daily report sent successfully",
    "date": "2025-01-15",
    "recipients_count": 2,
    "recipients": ["ashenafialemu27@gmail.com", "johsjones612@gmail.com"],
    "note": "Test email sent to configured recipients"
  }
}
```

### 3. Get Cronjob Status
**Endpoint**: `GET /analytics/daily-report/cronjob-status`

Check the status of the automatic daily report cronjob service.

```bash
curl -X GET http://localhost:8080/analytics/daily-report/cronjob-status
```

**Response**:
```json
{
  "success": true,
  "data": {
    "status": "running",
    "is_running": true,
    "message": "Daily report cronjob service status",
    "configured_recipients": ["ashenafialemu27@gmail.com", "johsjones612@gmail.com"],
    "schedule": "23:59 UTC (end of day)",
    "next_run": "Tomorrow at 23:59 UTC"
  }
}
```

### 4. Send Yesterday's Report
**Endpoint**: `POST /analytics/daily-report/yesterday`

Send yesterday's report to specified recipients.

```bash
curl -X POST http://localhost:8080/analytics/daily-report/yesterday \
  -H "Content-Type: application/json" \
  -d '{"recipients": ["ashenafialemu27@gmail.com", "johsjones612@gmail.com"]}'
```

### 5. Get Daily Report Data
**Endpoint**: `GET /analytics/reports/daily?date=YYYY-MM-DD`

Retrieve daily report data for a specific date.

```bash
curl -X GET "http://localhost:8080/analytics/reports/daily?date=2025-01-15"
```

## 🔧 System Architecture

### Components

1. **DailyReportCronjobService**: Handles automatic scheduling using cron
2. **DailyReportService**: Generates and sends email reports
3. **Analytics Handler**: Provides API endpoints for manual operations
4. **Email Service**: Sends HTML-formatted reports

### Configuration

The system is configured in `config/production.yaml`:

```yaml
# Daily Report Configuration
daily_reports:
  enabled: true
  recipients:
    - "ashenafialemu27@gmail.com"
    - "johsjones612@gmail.com"
  schedule:
    time: "23:59"  # End of day (11:59 PM)
    timezone: "UTC"
  email:
    subject: "TucanBIT Daily Analytics Report - {{.Date}}"
    template: "daily_report"
```

## 📊 Report Content

The daily reports include:

### Financial Overview
- Total Deposits (count and amount)
- Total Withdrawals (count and amount)
- Net Revenue
- GGR (Gross Gaming Revenue)

### User Activity
- New Users
- Active Users
- Total Bets
- Total Wins
- Active Games

### Top Games
- Most played games
- Revenue by game
- Player engagement metrics

### Top Players
- Highest activity users
- Transaction patterns
- Game diversity

## 🚀 Quick Start

### 1. Test the System

Run the test script to verify all endpoints:

```bash
./test_daily_report_api.sh
```

### 2. Send Manual Report

Send a test report immediately:

```bash
curl -X POST http://localhost:8080/analytics/daily-report/test \
  -H "Content-Type: application/json"
```

### 3. Check Cronjob Status

Verify the automatic scheduler is running:

```bash
curl -X GET http://localhost:8080/analytics/daily-report/cronjob-status
```

## 🔍 Monitoring

### Logs

The system logs all activities:

- **Cronjob startup**: "Daily report cronjob service started successfully"
- **Report generation**: "Generating daily report for date: YYYY-MM-DD"
- **Email sending**: "Daily report email sent successfully"
- **Errors**: Detailed error messages with context

### Status Monitoring

Use the cronjob status endpoint to monitor:
- Service running status
- Next scheduled run
- Configured recipients
- Schedule information

## 🛠️ Troubleshooting

### Common Issues

1. **Cronjob not running**
   - Check if ClickHouse is available
   - Verify email service is initialized
   - Check logs for initialization errors

2. **Emails not received**
   - Verify email service configuration
   - Check SMTP settings in config
   - Test with manual endpoint first

3. **No data in reports**
   - Ensure ClickHouse has data
   - Check date range
   - Verify analytics integration

### Debug Commands

```bash
# Check cronjob status
curl -X GET http://localhost:8080/analytics/daily-report/cronjob-status

# Send test report
curl -X POST http://localhost:8080/analytics/daily-report/test

# Get report data
curl -X GET "http://localhost:8080/analytics/reports/daily?date=$(date -d yesterday +%Y-%m-%d)"
```

## 📈 Future Enhancements

- **Multiple timezone support**
- **Custom report templates**
- **Additional recipients management**
- **Report scheduling flexibility**
- **Email delivery tracking**

## ✅ Implementation Status

- ✅ **Email Configuration**: Recipients configured in production.yaml
- ✅ **Manual API**: Send configured daily report endpoint
- ✅ **Automatic Cronjob**: Daily scheduling at 23:59 UTC
- ✅ **Test Endpoints**: Test report and status checking
- ✅ **Error Handling**: Comprehensive error management
- ✅ **Logging**: Detailed activity logging
- ✅ **Documentation**: Complete API documentation

The TucanBIT Daily Report System is now fully operational and will automatically send daily analytics reports to the configured recipients every day at 23:59 UTC! 🎉