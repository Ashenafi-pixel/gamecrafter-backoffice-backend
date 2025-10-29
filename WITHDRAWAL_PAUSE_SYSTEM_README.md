# Withdrawal Pause System

This document explains how to use the withdrawal pause system implemented using the `system_config` table.

## Overview

The withdrawal pause system allows administrators to:
- Pause withdrawals globally or individually
- Set configurable thresholds for automatic pausing
- Require manual review for certain withdrawals
- Monitor and manage paused withdrawals through the back-office interface

## Database Configuration

The system uses the existing `system_config` table to store all configuration. No database schema changes are required.

### Configuration Keys

1. **`withdrawal_global_status`** - Global withdrawal enable/disable status
2. **`withdrawal_thresholds`** - Volume and transaction thresholds
3. **`withdrawal_manual_review`** - Manual review requirements
4. **`withdrawal_pause_reasons`** - Predefined pause reasons
5. **`withdrawal_paused_transactions`** - Currently paused withdrawals

## Setup Instructions

### 1. Initialize System Configuration

Run the initialization script to set up the withdrawal pause system:

```bash
psql -d tucanbit -f scripts/init_withdrawal_pause_system.sql
```

This will create the necessary configuration entries in the `system_config` table.

### 2. Backend Integration

The system includes:
- **System Config Storage** (`internal/storage/system_config/`)
- **Withdrawal Management** (`internal/storage/withdrawal_management/`)
- **Withdrawal Processor** (`internal/service/withdrawal_processor/`)
- **API Handlers** (`internal/handler/system_config/` and `internal/handler/withdrawal_management/`)

### 3. Frontend Components

The back-office includes:
- **WithdrawalDashboard** - Overview and quick controls
- **WithdrawalManagement** - Manage paused withdrawals
- **WithdrawalSettings** - Configure thresholds and settings

## API Endpoints

### System Configuration

- `GET /api/v1/system-config/withdrawal/global-status` - Get global status
- `PUT /api/v1/system-config/withdrawal/global-status` - Update global status
- `GET /api/v1/system-config/withdrawal/thresholds` - Get thresholds
- `PUT /api/v1/system-config/withdrawal/thresholds` - Update thresholds
- `GET /api/v1/system-config/withdrawal/manual-review` - Get manual review settings
- `PUT /api/v1/system-config/withdrawal/manual-review` - Update manual review settings
- `GET /api/v1/system-config/withdrawal/pause-reasons` - Get pause reasons

### Withdrawal Management

- `GET /api/v1/withdrawal-management/paused` - Get paused withdrawals
- `POST /api/v1/withdrawal-management/pause/:id` - Pause a withdrawal
- `POST /api/v1/withdrawal-management/unpause/:id` - Unpause a withdrawal
- `POST /api/v1/withdrawal-management/approve/:id` - Approve a withdrawal
- `POST /api/v1/withdrawal-management/reject/:id` - Reject a withdrawal
- `GET /api/v1/withdrawal-management/stats` - Get statistics

## Configuration Examples

### Global Status

```json
{
  "enabled": true,
  "reason": "System operational",
  "paused_by": null,
  "paused_at": null
}
```

### Thresholds

```json
{
  "hourly_volume": {
    "value": 50000,
    "currency": "USD",
    "active": true
  },
  "daily_volume": {
    "value": 1000000,
    "currency": "USD",
    "active": true
  },
  "single_transaction": {
    "value": 10000,
    "currency": "USD",
    "active": true
  },
  "user_daily": {
    "value": 5000,
    "currency": "USD",
    "active": true
  }
}
```

### Manual Review

```json
{
  "enabled": true,
  "threshold_amount": 5000,
  "currency": "USD",
  "require_kyc": true
}
```

## Usage Scenarios

### 1. Global Pause

To pause all withdrawals system-wide:

```bash
curl -X PUT http://localhost:8080/api/v1/system-config/withdrawal/global-status \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false,
    "reason": "Maintenance in progress"
  }'
```

### 2. Set Thresholds

To configure volume thresholds:

```bash
curl -X PUT http://localhost:8080/api/v1/system-config/withdrawal/thresholds \
  -H "Content-Type: application/json" \
  -d '{
    "hourly_volume": {
      "value": 100000,
      "currency": "USD",
      "active": true
    },
    "daily_volume": {
      "value": 2000000,
      "currency": "USD",
      "active": true
    }
  }'
```

### 3. Pause Individual Withdrawal

To pause a specific withdrawal:

```bash
curl -X POST http://localhost:8080/api/v1/withdrawal-management/pause/WD123456 \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Suspicious activity detected",
    "requires_review": true,
    "notes": "Manual review required"
  }'
```

### 4. Approve Withdrawal

To approve a paused withdrawal:

```bash
curl -X POST http://localhost:8080/api/v1/withdrawal-management/approve/WD123456 \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Approved after review"
  }'
```

## Integration with Withdrawal Processing

To integrate with your existing withdrawal processing:

1. **Check Global Status** - Before processing any withdrawal
2. **Check Thresholds** - Before processing individual withdrawals
3. **Pause if Needed** - Use the withdrawal processor service

Example integration:

```go
// In your withdrawal processing code
processor := withdrawal_processor.NewWithdrawalProcessor(db, log)

err := processor.ProcessWithdrawalRequest(
    ctx,
    withdrawalID,
    userID,
    amountCents,
    currency,
)

if err != nil {
    // Withdrawal was paused or blocked
    return err
}

// Continue with normal withdrawal processing
```

## Monitoring

### Statistics

The system provides real-time statistics:
- Total paused withdrawals
- Pending manual reviews
- Paused today/this hour
- Total paused amount

### Logs

All pause actions are logged with:
- Withdrawal ID
- Pause reason
- Admin who took action
- Timestamp
- Notes

## Security Considerations

1. **Admin Authentication** - All configuration changes require admin authentication
2. **Audit Trail** - All actions are logged with admin ID and timestamp
3. **Validation** - Input validation on all configuration values
4. **Rate Limiting** - Consider rate limiting for configuration changes

## Troubleshooting

### Common Issues

1. **Configuration Not Loading**
   - Check if `system_config` entries exist
   - Verify JSON format in `config_value` column
   - Check database connection

2. **Withdrawals Not Pausing**
   - Verify global status is disabled
   - Check threshold values and currency
   - Ensure withdrawal processor is integrated

3. **Frontend Not Updating**
   - Check API endpoints are accessible
   - Verify authentication tokens
   - Check browser console for errors

### Debug Queries

Check current configuration:

```sql
SELECT config_key, config_value, updated_at 
FROM system_config 
WHERE config_key LIKE 'withdrawal_%' 
ORDER BY config_key;
```

Check paused withdrawals:

```sql
SELECT config_value 
FROM system_config 
WHERE config_key = 'withdrawal_paused_transactions';
```

## Support

For issues or questions:
1. Check the logs for error messages
2. Verify configuration values in `system_config` table
3. Test API endpoints directly
4. Check frontend browser console for errors
