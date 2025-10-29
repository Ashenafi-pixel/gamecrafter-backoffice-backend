-- Create alert system tables
CREATE TYPE alert_type AS ENUM (
    'bets_count_less',
    'bets_count_more',
    'bets_amount_less',
    'bets_amount_more',
    'deposits_total_less',
    'deposits_total_more',
    'deposits_type_less',
    'deposits_type_more',
    'withdrawals_total_less',
    'withdrawals_total_more',
    'withdrawals_type_less',
    'withdrawals_type_more',
    'ggr_total_less',
    'ggr_total_more',
    'ggr_single_more'
);

CREATE TYPE alert_status AS ENUM (
    'active',
    'inactive',
    'triggered'
);

-- Create a new enum for specific currency codes since existing currency_type only has 'fiat' and 'crypto'
CREATE TYPE alert_currency_code AS ENUM (
    'USD',
    'BTC',
    'ETH',
    'SOL',
    'USDT',
    'USDC'
);

-- Alert configurations table
CREATE TABLE alert_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    alert_type alert_type NOT NULL,
    status alert_status DEFAULT 'active',
    
    -- Criteria parameters
    threshold_amount DECIMAL(20,8) NOT NULL, -- Amount threshold in USD
    time_window_minutes INTEGER NOT NULL, -- Time window in minutes
    currency_code alert_currency_code, -- For type-specific alerts (deposits/withdrawals)
    
    -- Notification settings
    email_notifications BOOLEAN DEFAULT false,
    webhook_url TEXT,
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id)
);

-- Alert triggers table (stores when alerts were triggered)
CREATE TABLE alert_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_configuration_id UUID NOT NULL REFERENCES alert_configurations(id) ON DELETE CASCADE,
    
    -- Trigger details
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    trigger_value DECIMAL(20,8) NOT NULL, -- Actual value that triggered the alert
    threshold_value DECIMAL(20,8) NOT NULL, -- Threshold that was exceeded
    
    -- Related data
    user_id UUID REFERENCES users(id),
    transaction_id VARCHAR(255),
    amount_usd DECIMAL(20,8),
    currency_code alert_currency_code,
    
    -- Additional context
    context_data JSONB, -- Store additional context like game details, etc.
    
    -- Status
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_alert_configurations_type ON alert_configurations(alert_type);
CREATE INDEX idx_alert_configurations_status ON alert_configurations(status);
CREATE INDEX idx_alert_triggers_config_id ON alert_triggers(alert_configuration_id);
CREATE INDEX idx_alert_triggers_triggered_at ON alert_triggers(triggered_at);
CREATE INDEX idx_alert_triggers_user_id ON alert_triggers(user_id);
CREATE INDEX idx_alert_triggers_acknowledged ON alert_triggers(acknowledged);

-- Insert default alert configurations
INSERT INTO alert_configurations (name, description, alert_type, threshold_amount, time_window_minutes, currency_code, created_by) VALUES
('Low Bet Activity', 'Alert when bet count is below threshold', 'bets_count_less', 10, 60, 'USD', (SELECT id FROM users WHERE is_admin = true LIMIT 1)),
('High Bet Activity', 'Alert when bet count exceeds threshold', 'bets_count_more', 100, 60, 'USD', (SELECT id FROM users WHERE is_admin = true LIMIT 1)),
('Low Deposit Volume', 'Alert when total deposits are below threshold', 'deposits_total_less', 1000, 60, 'USD', (SELECT id FROM users WHERE is_admin = true LIMIT 1)),
('High Withdrawal Volume', 'Alert when total withdrawals exceed threshold', 'withdrawals_total_more', 5000, 60, 'USD', (SELECT id FROM users WHERE is_admin = true LIMIT 1)),
('High GGR Single Transaction', 'Alert when single transaction GGR exceeds threshold', 'ggr_single_more', 10000, 0, 'USD', (SELECT id FROM users WHERE is_admin = true LIMIT 1));
