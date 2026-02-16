-- Consolidated schema for SQLC generation
-- This file contains all table definitions for the application

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE,
    referral_code VARCHAR(50) UNIQUE,
    password VARCHAR(255),
    default_currency VARCHAR(10) DEFAULT 'ETB',
    profile_picture TEXT,
    date_of_birth DATE,
    source VARCHAR(100),
    street_address TEXT,
    country VARCHAR(100),
    state VARCHAR(100),
    city VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    created_by UUID,
    is_admin BOOLEAN DEFAULT false,
    postal_code VARCHAR(20),
    kyc_status VARCHAR(50) DEFAULT 'PENDING',
    type VARCHAR(50) DEFAULT 'PLAYER',
    referal_type VARCHAR(50),
    refered_by_code VARCHAR(50),
    agent_request_id VARCHAR(100),
    primary_wallet_address VARCHAR(255),
    wallet_verification_status VARCHAR(50) DEFAULT 'none',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Crypto wallet connections table
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

-- Crypto wallet challenges table
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

-- Crypto wallet auth logs table
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

-- Balances table
CREATE TABLE IF NOT EXISTS balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL,
    real_money DECIMAL(20,8) DEFAULT 0,
    bonus_money DECIMAL(20,8) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, currency)
);

-- Balance logs table
CREATE TABLE IF NOT EXISTS balance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance_id UUID NOT NULL REFERENCES balances(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    balance_before DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    description TEXT,
    reference_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User roles table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Role permissions table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_connections_user_id ON crypto_wallet_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_connections_wallet_address ON crypto_wallet_connections(wallet_address);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_connections_wallet_type ON crypto_wallet_connections(wallet_type);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_challenges_wallet_address ON crypto_wallet_challenges(wallet_address);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_challenges_expires_at ON crypto_wallet_challenges(expires_at);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_auth_logs_wallet_address ON crypto_wallet_auth_logs(wallet_address);
CREATE INDEX IF NOT EXISTS idx_crypto_wallet_auth_logs_created_at ON crypto_wallet_auth_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);
CREATE INDEX IF NOT EXISTS idx_balance_logs_user_id ON balance_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_balance_logs_balance_id ON balance_logs(balance_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id); 