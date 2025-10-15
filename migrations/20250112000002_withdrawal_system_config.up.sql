-- Migration: Add withdrawal control to system_config table
-- This replaces the dedicated withdrawal_pause_settings table with system config entries

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
            "hourly_volume": {"value": 50000, "currency": "USD", "active": true},
            "daily_volume": {"value": 1000000, "currency": "USD", "active": true},
            "single_transaction": {"value": 10000, "currency": "USD", "active": true},
            "user_daily": {"value": 5000, "currency": "USD", "active": true}
        }',
        'Withdrawal threshold limits for automatic pausing based on volume and transaction size',
        (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
    ),
    (
        'withdrawal_manual_review',
        '{
            "enabled": true,
            "threshold_amount": 5000,
            "currency": "USD",
            "require_kyc": true
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
    )
ON CONFLICT (config_key) DO NOTHING;

-- Create indexes for better performance on system config queries
CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config(config_key);
CREATE INDEX IF NOT EXISTS idx_system_config_updated_at ON system_config(updated_at);

-- Add a function to easily get withdrawal status
CREATE OR REPLACE FUNCTION get_withdrawal_global_status()
RETURNS BOOLEAN AS $$
DECLARE
    status_config JSONB;
BEGIN
    SELECT config_value INTO status_config 
    FROM system_config 
    WHERE config_key = 'withdrawal_global_status' 
    LIMIT 1;
    
    IF status_config IS NULL THEN
        RETURN TRUE; -- Default to enabled if no config found
    END IF;
    
    RETURN COALESCE((status_config->>'enabled')::BOOLEAN, TRUE);
END;
$$ LANGUAGE plpgsql;

-- Add a function to easily update withdrawal status
CREATE OR REPLACE FUNCTION update_withdrawal_global_status(
    p_enabled BOOLEAN,
    p_reason TEXT DEFAULT NULL,
    p_paused_by UUID DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    new_config JSONB;
BEGIN
    new_config := jsonb_build_object(
        'enabled', p_enabled,
        'reason', COALESCE(p_reason, 'Status updated'),
        'paused_by', p_paused_by,
        'paused_at', CASE WHEN p_enabled = FALSE THEN NOW() ELSE NULL END
    );
    
    INSERT INTO system_config (config_key, config_value, updated_by)
    VALUES ('withdrawal_global_status', new_config, p_paused_by)
    ON CONFLICT (config_key) 
    DO UPDATE SET 
        config_value = new_config,
        updated_by = p_paused_by,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Add a function to get withdrawal thresholds
CREATE OR REPLACE FUNCTION get_withdrawal_thresholds()
RETURNS JSONB AS $$
DECLARE
    thresholds_config JSONB;
BEGIN
    SELECT config_value INTO thresholds_config 
    FROM system_config 
    WHERE config_key = 'withdrawal_thresholds' 
    LIMIT 1;
    
    RETURN COALESCE(thresholds_config, '{}'::JSONB);
END;
$$ LANGUAGE plpgsql;

-- Add a function to check if withdrawal should be paused based on thresholds
CREATE OR REPLACE FUNCTION check_withdrawal_thresholds(
    p_amount DECIMAL,
    p_currency VARCHAR(10) DEFAULT 'USD',
    p_threshold_type VARCHAR(50) DEFAULT 'single_transaction'
)
RETURNS BOOLEAN AS $$
DECLARE
    thresholds_config JSONB;
    threshold_value DECIMAL;
    threshold_active BOOLEAN;
BEGIN
    SELECT config_value INTO thresholds_config 
    FROM system_config 
    WHERE config_key = 'withdrawal_thresholds' 
    LIMIT 1;
    
    IF thresholds_config IS NULL THEN
        RETURN FALSE; -- No thresholds configured, don't pause
    END IF;
    
    -- Get threshold value and active status
    threshold_value := (thresholds_config->p_threshold_type->>'value')::DECIMAL;
    threshold_active := (thresholds_config->p_threshold_type->>'active')::BOOLEAN;
    
    -- Check if threshold is active and amount exceeds it
    RETURN threshold_active AND p_amount > threshold_value;
END;
$$ LANGUAGE plpgsql;
