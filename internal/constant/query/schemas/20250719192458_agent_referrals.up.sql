CREATE TABLE IF NOT EXISTS agent_referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL UNIQUE,
    callback_url TEXT NOT NULL,
    user_id UUID,
    conversion_type VARCHAR(100),
    amount DECIMAL(20,8) DEFAULT 0,
    msisdn VARCHAR(20),
    converted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    callback_sent BOOLEAN DEFAULT FALSE,
    callback_attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_referrals_request_id ON agent_referrals(request_id);
CREATE INDEX IF NOT EXISTS idx_agent_referrals_user_id ON agent_referrals(user_id);
CREATE INDEX IF NOT EXISTS idx_agent_referrals_converted_at ON agent_referrals(converted_at);
CREATE INDEX IF NOT EXISTS idx_agent_referrals_callback_sent ON agent_referrals(callback_sent);
CREATE INDEX IF NOT EXISTS idx_agent_referrals_callback_attempts ON agent_referrals(callback_attempts);

ALTER TABLE agent_referrals 
ADD CONSTRAINT fk_agent_referrals_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL; 