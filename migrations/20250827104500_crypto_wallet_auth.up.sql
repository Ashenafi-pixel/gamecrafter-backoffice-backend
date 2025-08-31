-- Create crypto wallet authentication tables integrated with existing JWT auth
-- Migration: 20250827104500_crypto_wallet_auth.up.sql

-- Table for storing crypto wallet connections
CREATE TABLE IF NOT EXISTS crypto_wallet_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_type VARCHAR(50) NOT NULL CHECK (wallet_type IN ('metamask', 'walletconnect', 'coinbase', 'phantom', 'trust', 'ledger')),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_chain VARCHAR(50) NOT NULL DEFAULT 'ethereum',
    wallet_name VARCHAR(255),
    wallet_icon_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    verification_signature TEXT,
    verification_message TEXT,
    verification_timestamp TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, wallet_address),
    UNIQUE(wallet_address, wallet_type)
);

-- Table for wallet verification challenges (for sign-in verification)
CREATE TABLE IF NOT EXISTS crypto_wallet_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(50) NOT NULL,
    challenge_message TEXT NOT NULL,
    challenge_nonce VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Table for wallet authentication logs
CREATE TABLE IF NOT EXISTS crypto_wallet_auth_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('connect', 'disconnect', 'login', 'verify', 'challenge')),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_crypto_wallet_connections_user_id ON crypto_wallet_connections(user_id);
CREATE INDEX idx_crypto_wallet_connections_wallet_address ON crypto_wallet_connections(wallet_address);
CREATE INDEX idx_crypto_wallet_connections_wallet_type ON crypto_wallet_connections(wallet_type);
CREATE INDEX idx_crypto_wallet_challenges_wallet_address ON crypto_wallet_challenges(wallet_address);
CREATE INDEX idx_crypto_wallet_challenges_expires_at ON crypto_wallet_challenges(expires_at);
CREATE INDEX idx_crypto_wallet_auth_logs_wallet_address ON crypto_wallet_auth_logs(wallet_address);
CREATE INDEX idx_crypto_wallet_auth_logs_created_at ON crypto_wallet_auth_logs(created_at);

-- Add crypto wallet fields to users table if they don't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'primary_wallet_address') THEN
        ALTER TABLE users ADD COLUMN primary_wallet_address VARCHAR(255);
        ALTER TABLE users ADD COLUMN wallet_verification_status VARCHAR(50) DEFAULT 'none' CHECK (wallet_verification_status IN ('none', 'pending', 'verified', 'failed'));
    END IF;
END $$;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_crypto_wallet_connections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_crypto_wallet_connections_updated_at
    BEFORE UPDATE ON crypto_wallet_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_crypto_wallet_connections_updated_at();

-- Create function to clean expired challenges
CREATE OR REPLACE FUNCTION clean_expired_crypto_wallet_challenges()
RETURNS void AS $$
BEGIN
    DELETE FROM crypto_wallet_challenges WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Create function to get user wallet info
CREATE OR REPLACE FUNCTION get_user_wallets(user_uuid UUID)
RETURNS TABLE (
    wallet_type VARCHAR(50),
    wallet_address VARCHAR(255),
    wallet_chain VARCHAR(50),
    wallet_name VARCHAR(255),
    is_verified BOOLEAN,
    last_used_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cwc.wallet_type,
        cwc.wallet_address,
        cwc.wallet_chain,
        cwc.wallet_name,
        cwc.is_verified,
        cwc.last_used_at
    FROM crypto_wallet_connections cwc
    WHERE cwc.user_id = user_uuid
    ORDER BY cwc.last_used_at DESC;
END;
$$ LANGUAGE plpgsql; 