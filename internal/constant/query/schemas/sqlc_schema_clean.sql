-- SQLC Schema file - Clean table definitions for code generation

-- Users table with all current columns
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(20),
    phone_number VARCHAR(15),
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    default_currency VARCHAR(3),
    profile VARCHAR DEFAULT ''::VARCHAR,
    email VARCHAR DEFAULT ''::VARCHAR,
    first_name VARCHAR DEFAULT ''::VARCHAR,
    last_name VARCHAR DEFAULT ''::VARCHAR,
    date_of_birth VARCHAR DEFAULT ''::VARCHAR,
    source VARCHAR DEFAULT ''::VARCHAR,
    is_email_verified BOOLEAN DEFAULT false,
    referal_code VARCHAR DEFAULT ''::VARCHAR,
    street_address VARCHAR NOT NULL DEFAULT ''::VARCHAR,
    country VARCHAR NOT NULL DEFAULT ''::VARCHAR,
    state VARCHAR NOT NULL DEFAULT ''::VARCHAR,
    city VARCHAR NOT NULL DEFAULT ''::VARCHAR,
    postal_code VARCHAR NOT NULL DEFAULT ''::VARCHAR,
    kyc_status VARCHAR NOT NULL DEFAULT 'PENDING'::VARCHAR,
    created_by UUID,
    is_admin BOOLEAN,
    status VARCHAR DEFAULT 'ACTIVE'::VARCHAR,
    referal_type VARCHAR(255),
    refered_by_code VARCHAR(255),
    user_type VARCHAR(255) DEFAULT 'PLAYER'::VARCHAR,
    primary_wallet_address VARCHAR(255),
    wallet_verification_status VARCHAR(50) DEFAULT 'none',
    is_test_account BOOLEAN DEFAULT true,
    two_factor_enabled BOOLEAN,
    two_factor_setup_at TIMESTAMP WITH TIME ZONE
);

-- Balances table with current structure
CREATE TABLE balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency_code VARCHAR(10) NOT NULL,
    amount_cents BIGINT DEFAULT 0,
    amount_units NUMERIC(36,18) DEFAULT 0,
    reserved_cents BIGINT DEFAULT 0,
    reserved_units NUMERIC(36,18) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    currency VARCHAR(10) DEFAULT 'USD'::VARCHAR,
    UNIQUE(user_id, currency_code)
);

-- Crypto wallet connections table
CREATE TABLE crypto_wallet_connections (
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
    UNIQUE(user_id, wallet_address, wallet_chain)
);

-- Crypto wallet challenges table
CREATE TABLE crypto_wallet_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(50) NOT NULL,
    challenge_message TEXT NOT NULL,
    challenge_nonce VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Crypto wallet auth logs table
CREATE TABLE crypto_wallet_auth_logs (
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

-- Currency config table
CREATE TABLE currency_config (
    currency_code VARCHAR(10) PRIMARY KEY,
    currency_name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimal_places INTEGER DEFAULT 2,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
