-- Create sport_bets table with transaction-style structure based on PlaceBetRequest DTO
CREATE TABLE sport_bets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    bet_amount DECIMAL(10,2) NOT NULL,
    bet_reference_num VARCHAR(255) NOT NULL,
    game_reference VARCHAR(255) NOT NULL,
    bet_mode VARCHAR(50) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    frontend_type VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sport_ids TEXT,
    site_id VARCHAR(255) NOT NULL,
    client_ip INET,
    affiliate_user_id VARCHAR(255),
    autorecharge VARCHAR(10),
    bet_details JSONB NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    potential_win DECIMAL(10,2),
    actual_win DECIMAL(10,2),
    odds DECIMAL(10,4),
    placed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create essential indexes only
CREATE INDEX idx_sport_bets_transaction_id ON sport_bets(transaction_id);
CREATE INDEX idx_sport_bets_user_id ON sport_bets(user_id);
CREATE INDEX idx_sport_bets_bet_status ON sport_bets(status);
CREATE INDEX idx_sport_bets_placed_at ON sport_bets(placed_at);
