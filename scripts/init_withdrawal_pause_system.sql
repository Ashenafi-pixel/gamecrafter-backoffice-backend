-- Script to initialize withdrawal pause system in system_config table
-- Run this script to set up the withdrawal pause functionality

-- Insert withdrawal control configurations into system_config table
INSERT INTO system_config (config_key, config_value, description, updated_by) 
VALUES 
    (
        'withdrawal_global_status',
        '{"enabled": true, "reason": "System initialized", "paused_by": null, "paused_at": null}',
        'Global withdrawal status control - enables/disables all withdrawals system-wide',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    ),
    (
        'withdrawal_thresholds',
        '{
            "hourly_volume": {"value": 50000, "currency": "USD", "enabled": true},
            "daily_volume": {"value": 1000000, "currency": "USD", "enabled": true},
            "single_transaction": {"value": 10000, "currency": "USD", "enabled": true},
            "user_daily": {"value": 5000, "currency": "USD", "enabled": true}
        }',
        'Withdrawal threshold limits for automatic pausing based on volume and transaction size',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    ),
    (
        'withdrawal_manual_review',
        '{
            "enabled": true,
            "threshold_cents": 500000,
            "require_admin_approval": true
        }',
        'Manual review requirements for withdrawals above specified thresholds',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    ),
    (
        'withdrawal_pause_reasons',
        '[
            "global_pause",
            "threshold_exceeded", 
            "manual_review_required",
            "suspicious_activity",
            "kyc_verification_pending",
            "maintenance_mode"
        ]',
        'Predefined reasons for pausing withdrawals',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    ),
    (
        'withdrawal_paused_transactions',
        '{}',
        'JSON object storing paused withdrawal IDs and their pause details',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    )
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    description = EXCLUDED.description,
    updated_at = NOW();

-- Display current withdrawal pause configuration
SELECT 
    config_key,
    config_value,
    description,
    created_at,
    updated_at
FROM system_config 
WHERE config_key LIKE 'withdrawal_%'
ORDER BY config_key;
