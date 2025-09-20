-- Create retryable_operations table for retry mechanism
CREATE TABLE IF NOT EXISTS retryable_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'retrying', 'failed', 'completed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_retryable_operations_user_id ON retryable_operations(user_id);
CREATE INDEX IF NOT EXISTS idx_retryable_operations_status ON retryable_operations(status);
CREATE INDEX IF NOT EXISTS idx_retryable_operations_next_retry_at ON retryable_operations(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_retryable_operations_type ON retryable_operations(type);
CREATE INDEX IF NOT EXISTS idx_retryable_operations_created_at ON retryable_operations(created_at);

-- Create composite index for failed operations query
CREATE INDEX IF NOT EXISTS idx_retryable_operations_failed ON retryable_operations(status, next_retry_at) 
WHERE status = 'failed';

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_retryable_operations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_retryable_operations_updated_at
    BEFORE UPDATE ON retryable_operations
    FOR EACH ROW
    EXECUTE FUNCTION update_retryable_operations_updated_at();

-- Add comments for documentation
COMMENT ON TABLE retryable_operations IS 'Stores operations that can be retried with exponential backoff';
COMMENT ON COLUMN retryable_operations.type IS 'Type of operation (process_bet_cashback, claim_cashback, update_user_level)';
COMMENT ON COLUMN retryable_operations.data IS 'JSON data containing operation parameters';
COMMENT ON COLUMN retryable_operations.attempts IS 'Number of retry attempts made';
COMMENT ON COLUMN retryable_operations.last_error IS 'Last error message from failed attempt';
COMMENT ON COLUMN retryable_operations.next_retry_at IS 'Timestamp when operation should be retried next';
COMMENT ON COLUMN retryable_operations.status IS 'Current status: pending, retrying, failed, completed';