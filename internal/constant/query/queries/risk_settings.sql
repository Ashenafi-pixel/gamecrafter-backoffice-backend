-- risk_settings.sql - Queries for the singleton risk settings table

-- name: GetRiskSettings :one
SELECT * FROM risk_settings
WHERE id = 1;

-- name: CreateOrUpdateRiskSettings :one
INSERT INTO risk_settings (
    id,
    system_limits_enabled,
    system_max_daily_airtime_conversion,
    system_max_weekly_airtime_conversion,
    system_max_monthly_airtime_conversion,
    player_limits_enabled,
    player_max_daily_airtime_conversion,
    player_max_weekly_airtime_conversion,
    player_max_monthly_airtime_conversion,
    player_min_airtime_conversion_amount,
    player_conversion_cooldown_hours,
    kyc_required_above_amount,
    kyc_verification_timeout_hours,
    kyc_allow_partial,
    fraud_max_login_attempts,
    fraud_login_lockout_duration_minutes,
    alert_admins_on_trigger
) VALUES (
    1, -- The required ID for the singleton row
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
ON CONFLICT (id) DO UPDATE SET
    system_limits_enabled = EXCLUDED.system_limits_enabled,
    system_max_daily_airtime_conversion = EXCLUDED.system_max_daily_airtime_conversion,
    system_max_weekly_airtime_conversion = EXCLUDED.system_max_weekly_airtime_conversion,
    system_max_monthly_airtime_conversion = EXCLUDED.system_max_monthly_airtime_conversion,
    player_limits_enabled = EXCLUDED.player_limits_enabled,
    player_max_daily_airtime_conversion = EXCLUDED.player_max_daily_airtime_conversion,
    player_max_weekly_airtime_conversion = EXCLUDED.player_max_weekly_airtime_conversion,
    player_max_monthly_airtime_conversion = EXCLUDED.player_max_monthly_airtime_conversion,
    player_min_airtime_conversion_amount = EXCLUDED.player_min_airtime_conversion_amount,
    player_conversion_cooldown_hours = EXCLUDED.player_conversion_cooldown_hours,
    kyc_required_above_amount = EXCLUDED.kyc_required_above_amount,
    kyc_verification_timeout_hours = EXCLUDED.kyc_verification_timeout_hours,
    kyc_allow_partial = EXCLUDED.kyc_allow_partial,
    fraud_max_login_attempts = EXCLUDED.fraud_max_login_attempts,
    fraud_login_lockout_duration_minutes = EXCLUDED.fraud_login_lockout_duration_minutes,
    alert_admins_on_trigger = EXCLUDED.alert_admins_on_trigger,
    updated_at = NOW()
RETURNING *;