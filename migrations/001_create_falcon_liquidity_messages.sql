-- Migration: Create falcon_liquidity_messages table for reconciliation and dispute resolution
-- This table stores all messages sent to Falcon Liquidity for audit and reconciliation purposes

CREATE TABLE IF NOT EXISTS falcon_liquidity_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id VARCHAR(255) NOT NULL UNIQUE, -- Unique identifier for the message
    transaction_id VARCHAR(255) NOT NULL, -- Original transaction ID from our system
    user_id UUID NOT NULL REFERENCES users(id),
    message_type VARCHAR(50) NOT NULL, -- 'casino' or 'sport'
    
    -- Message content (stored as JSONB for easy querying)
    message_data JSONB NOT NULL,
    
    -- Message metadata
    bet_amount DECIMAL(20,8) NOT NULL,
    payout_amount DECIMAL(20,8) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    game_name VARCHAR(255),
    game_id VARCHAR(100),
    house_edge DECIMAL(8,6), -- House edge percentage
    
    -- Falcon-specific data
    falcon_routing_key VARCHAR(255) NOT NULL,
    falcon_exchange VARCHAR(255) NOT NULL,
    falcon_queue VARCHAR(255) NOT NULL,
    
    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'failed', 'acknowledged'
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    
    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(100),
    
    -- Reconciliation data
    falcon_response JSONB, -- Response from Falcon (if any)
    reconciliation_status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'reconciled', 'disputed'
    reconciliation_notes TEXT,
    
    -- Indexes for performance
    CONSTRAINT falcon_messages_status_check CHECK (status IN ('pending', 'sent', 'failed', 'acknowledged')),
    CONSTRAINT falcon_messages_type_check CHECK (message_type IN ('casino', 'sport')),
    CONSTRAINT falcon_messages_reconciliation_check CHECK (reconciliation_status IN ('pending', 'reconciled', 'disputed'))
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_falcon_messages_user_id ON falcon_liquidity_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_transaction_id ON falcon_liquidity_messages(transaction_id);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_message_id ON falcon_liquidity_messages(message_id);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_status ON falcon_liquidity_messages(status);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_created_at ON falcon_liquidity_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_reconciliation_status ON falcon_liquidity_messages(reconciliation_status);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_message_type ON falcon_liquidity_messages(message_type);

-- Create a composite index for common queries
CREATE INDEX IF NOT EXISTS idx_falcon_messages_user_status ON falcon_liquidity_messages(user_id, status);
CREATE INDEX IF NOT EXISTS idx_falcon_messages_transaction_status ON falcon_liquidity_messages(transaction_id, status);

-- Add comments for documentation
COMMENT ON TABLE falcon_liquidity_messages IS 'Stores all messages sent to Falcon Liquidity for reconciliation and dispute resolution';
COMMENT ON COLUMN falcon_liquidity_messages.message_id IS 'Unique identifier for the message sent to Falcon';
COMMENT ON COLUMN falcon_liquidity_messages.transaction_id IS 'Original transaction ID from our system';
COMMENT ON COLUMN falcon_liquidity_messages.message_data IS 'Complete message data sent to Falcon (JSONB)';
COMMENT ON COLUMN falcon_liquidity_messages.status IS 'Current status of the message: pending, sent, failed, acknowledged';
COMMENT ON COLUMN falcon_liquidity_messages.reconciliation_status IS 'Reconciliation status: pending, reconciled, disputed';
COMMENT ON COLUMN falcon_liquidity_messages.falcon_response IS 'Response received from Falcon (if any)';
COMMENT ON COLUMN falcon_liquidity_messages.reconciliation_notes IS 'Notes for dispute resolution and reconciliation';
