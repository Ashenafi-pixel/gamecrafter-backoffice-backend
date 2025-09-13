-- Create groove_transactions table for storing transaction data for idempotency
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_groove_transactions_transaction_id ON groove_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_account_id ON groove_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_user_id ON groove_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_game_session_id ON groove_transactions(game_session_id);
CREATE INDEX IF NOT EXISTS idx_groove_transactions_created_at ON groove_transactions(created_at);

-- Add foreign key constraint to users table
ALTER TABLE groove_transactions 
ADD CONSTRAINT fk_groove_transactions_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;