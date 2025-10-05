-- Create GrooveTech integration tables
-- This migration creates tables for GrooveTech game provider integration

-- GrooveTech accounts table
CREATE TABLE IF NOT EXISTS groove_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(255) UNIQUE NOT NULL,
    session_id VARCHAR(255),
    balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- GrooveTech transactions table
CREATE TABLE IF NOT EXISTS groove_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    account_id VARCHAR(255) NOT NULL REFERENCES groove_accounts(account_id) ON DELETE CASCADE,
    session_id VARCHAR(255),
    amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    type VARCHAR(50) NOT NULL, -- 'debit', 'credit', 'bet', 'win'
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB
);

-- GrooveTech game sessions table
CREATE TABLE IF NOT EXISTS groove_game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    account_id VARCHAR(255) NOT NULL REFERENCES groove_accounts(account_id) ON DELETE CASCADE,
    game_id VARCHAR(255) NOT NULL,
    balance DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_groove_accounts_user_id ON groove_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_status ON groove_accounts(status);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_account_created ON groove_transactions(account_id, created_at);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_session_id ON groove_transactions(session_id);

-- Create triggers for updated_at
CREATE OR REPLACE FUNCTION update_groove_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_groove_accounts_updated_at
    BEFORE UPDATE ON groove_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_groove_accounts_updated_at();

-- Insert some sample data for testing
INSERT INTO groove_accounts (user_id, account_id, balance, currency, status)
SELECT 
    u.id,
    'groove_' || u.id::text,
    COALESCE(b.amount_units, 0),
    'USD',
    'active'
FROM users u
LEFT JOIN balances b ON u.id = b.user_id AND b.currency_code = 'USD'
WHERE NOT EXISTS (
    SELECT 1 FROM groove_accounts ga WHERE ga.user_id = u.id
)
LIMIT 10;

-- Create a function to clean up expired game sessions
CREATE OR REPLACE FUNCTION cleanup_expired_groove_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE groove_game_sessions 
    SET status = 'expired'
    WHERE status = 'active' 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get GrooveTech account summary
CREATE OR REPLACE FUNCTION get_groove_account_summary(p_account_id VARCHAR(255))
RETURNS TABLE (
    account_id VARCHAR(255),
    balance DECIMAL(20,8),
    currency VARCHAR(10),
    status VARCHAR(50),
    total_transactions BIGINT,
    last_transaction_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ga.account_id,
        ga.balance,
        ga.currency,
        ga.status,
        COUNT(gt.transaction_id) as total_transactions,
        MAX(gt.created_at) as last_transaction_at
    FROM groove_accounts ga
    LEFT JOIN groove_transactions gt ON ga.account_id = gt.account_id
    WHERE ga.account_id = p_account_id
    GROUP BY ga.account_id, ga.balance, ga.currency, ga.status;
END;
$$ LANGUAGE plpgsql;