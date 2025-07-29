CREATE TABLE risk_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    
    system_limits_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    system_max_daily_airtime_conversion BIGINT NOT NULL DEFAULT 0,
    system_max_weekly_airtime_conversion BIGINT NOT NULL DEFAULT 0,
    system_max_monthly_airtime_conversion BIGINT NOT NULL DEFAULT 0,

    player_limits_enabled BOOLEAN NOT NULL DEFAULT FALSE, 
    player_max_daily_airtime_conversion INT NOT NULL DEFAULT 0,
    player_max_weekly_airtime_conversion INT NOT NULL DEFAULT 0,
    player_max_monthly_airtime_conversion INT NOT NULL DEFAULT 0,
    player_min_airtime_conversion_amount INT NOT NULL DEFAULT 0,
    player_conversion_cooldown_hours INT NOT NULL DEFAULT 0,

    kyc_required_above_amount INT NOT NULL DEFAULT 0,
    kyc_verification_timeout_hours INT NOT NULL DEFAULT 0,
    kyc_allow_partial BOOLEAN NOT NULL DEFAULT FALSE,

    fraud_max_login_attempts SMALLINT NOT NULL DEFAULT 5,
    fraud_login_lockout_duration_minutes INT NOT NULL DEFAULT 0,

    alert_admins_on_trigger BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT singleton_check CHECK (id = 1)
);