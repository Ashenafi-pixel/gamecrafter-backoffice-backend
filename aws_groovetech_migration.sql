-- =====================================================
-- TucanBIT GrooveTech Database Migration Script
-- For AWS PostgreSQL Database
-- =====================================================
-- This script creates all necessary tables and functions
-- for GrooveTech API integration
-- =====================================================

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- 1. GROOVE ACCOUNTS TABLE
-- =====================================================
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

-- =====================================================
-- 2. GROOVE TRANSACTIONS TABLE (Updated Schema)
-- =====================================================
CREATE TABLE IF NOT EXISTS groove_transactions (
    id SERIAL PRIMARY KEY,
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    account_transaction_id VARCHAR(50) NOT NULL,
    account_id VARCHAR(60) NOT NULL,
    game_session_id VARCHAR(64) NOT NULL,
    round_id VARCHAR(255) NOT NULL,
    game_id VARCHAR(255) NOT NULL,
    bet_amount DECIMAL(32,10) NOT NULL,
    device VARCHAR(20) NOT NULL,
    frbid VARCHAR(255),
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================================================
-- 3. GAME SESSIONS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(64) UNIQUE NOT NULL,
    game_id VARCHAR(20) NOT NULL,
    device_type VARCHAR(20) NOT NULL CHECK (device_type IN ('desktop', 'mobile')),
    game_mode VARCHAR(10) NOT NULL CHECK (game_mode IN ('demo', 'real')),
    groove_url TEXT,
    home_url TEXT,
    exit_url TEXT,
    history_url TEXT,
    license_type VARCHAR(20) DEFAULT 'Curacao',
    is_test_account BOOLEAN DEFAULT false,
    reality_check_elapsed INTEGER DEFAULT 0,
    reality_check_interval INTEGER DEFAULT 60,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '2 hours',
    is_active BOOLEAN DEFAULT true,
    last_activity TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- 4. INDEXES FOR PERFORMANCE
-- =====================================================

-- Groove accounts indexes
CREATE INDEX IF NOT EXISTS idx_groove_accounts_user_id ON groove_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_account_id ON groove_accounts(account_id);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_status ON groove_accounts(status);
CREATE INDEX IF NOT EXISTS idx_groove_accounts_created_at ON groove_accounts(created_at);

-- Groove transactions indexes
CREATE INDEX IF NOT EXISTS idx_groove_transactions_transaction_id ON groove_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_account_id ON groove_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_user_id ON groove_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_game_session_id ON groove_transactions(game_session_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_created_at ON groove_transactions(created_at);

-- Game sessions indexes
CREATE INDEX IF NOT EXISTS idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_session_id ON game_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_game_id ON game_sessions(game_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_created_at ON game_sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_game_sessions_active ON game_sessions(is_active);

-- =====================================================
-- 5. FOREIGN KEY CONSTRAINTS
-- =====================================================

-- Add foreign key constraint for groove_transactions
ALTER TABLE groove_transactions 
ADD CONSTRAINT fk_groove_transactions_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- =====================================================
-- 6. FUNCTIONS
-- =====================================================

-- Function for unique GrooveTech session ID generation
CREATE OR REPLACE FUNCTION generate_groove_session_id()
RETURNS TEXT AS $$
BEGIN
    RETURN 'Tucan_' || gen_random_uuid()::text;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired game sessions
CREATE OR REPLACE FUNCTION cleanup_expired_game_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE game_sessions 
    SET is_active = false 
    WHERE expires_at < NOW() AND is_active = true;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired GrooveTech sessions
CREATE OR REPLACE FUNCTION cleanup_expired_groove_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE groove_accounts 
    SET status = 'expired'
    WHERE status = 'active' 
    AND last_activity < NOW() - INTERVAL '24 hours';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get GrooveTech account summary
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

-- =====================================================
-- 7. TRIGGERS
-- =====================================================

-- Trigger for updated_at on groove_accounts
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

-- Trigger for updated_at on groove_transactions
CREATE OR REPLACE FUNCTION update_groove_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_groove_transactions_updated_at
    BEFORE UPDATE ON groove_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_groove_transactions_updated_at();

-- Trigger to update last_activity on game_sessions
CREATE OR REPLACE FUNCTION update_game_session_activity()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_activity = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_game_session_activity
    BEFORE UPDATE ON game_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_game_session_activity();

-- =====================================================
-- 8. SAMPLE DATA (Optional - Remove if not needed)
-- =====================================================

-- Insert sample GrooveTech accounts for existing users
INSERT INTO groove_accounts (user_id, account_id, balance, currency, status)
SELECT 
    u.id,
    u.id::text, -- Use user ID as account ID for consistency
    COALESCE(b.amount_units, 0),
    'USD',
    'active'
FROM users u
LEFT JOIN balances b ON u.id = b.user_id AND b.currency_code = 'USD'
WHERE NOT EXISTS (
    SELECT 1 FROM groove_accounts ga WHERE ga.user_id = u.id
)
ON CONFLICT (account_id) DO NOTHING;

-- Insert sample game sessions
INSERT INTO game_sessions (user_id, session_id, game_id, device_type, game_mode, home_url, exit_url, license_type) 
VALUES 
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', 'Tucan_sample_session_001', '80102', 'desktop', 'real', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao'),
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', 'Tucan_sample_session_002', '80103', 'mobile', 'demo', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao')
ON CONFLICT (session_id) DO NOTHING;

-- =====================================================
-- 9. VERIFICATION QUERIES
-- =====================================================

-- Verify tables were created
SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE tablename IN ('groove_accounts', 'groove_transactions', 'game_sessions')
ORDER BY tablename;

-- Verify indexes were created
SELECT 
    indexname,
    tablename,
    indexdef
FROM pg_indexes 
WHERE tablename IN ('groove_accounts', 'groove_transactions', 'game_sessions')
ORDER BY tablename, indexname;

-- Verify functions were created
SELECT 
    proname as function_name,
    prosrc as function_body
FROM pg_proc 
WHERE proname IN ('generate_groove_session_id', 'cleanup_expired_game_sessions', 'cleanup_expired_groove_sessions', 'get_groove_account_summary')
ORDER BY proname;

-- =====================================================
-- MIGRATION COMPLETE
-- =====================================================
-- All GrooveTech tables, indexes, functions, and triggers
-- have been created successfully.
-- 
-- Next steps:
-- 1. Update your application config to point to AWS database
-- 2. Test the GrooveTech API endpoints
-- 3. Verify data consistency between balances and groove_accounts
-- =====================================================