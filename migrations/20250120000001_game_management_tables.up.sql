-- Create game_management table
CREATE TABLE IF NOT EXISTS game_management (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE', 'MAINTENANCE')),
    photo TEXT,
    price DECIMAL(10,2) DEFAULT 0.00,
    enabled BOOLEAN DEFAULT true,
    game_id VARCHAR(255) UNIQUE NOT NULL,
    internal_name VARCHAR(255) NOT NULL,
    integration_partner VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create house_edge_management table
CREATE TABLE IF NOT EXISTS house_edge_management (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_type VARCHAR(100) NOT NULL,
    game_variant VARCHAR(100) NOT NULL,
    house_edge DECIMAL(5,4) NOT NULL CHECK (house_edge >= 0 AND house_edge <= 1),
    min_bet DECIMAL(10,2) NOT NULL CHECK (min_bet >= 0),
    max_bet DECIMAL(10,2) NOT NULL CHECK (max_bet >= 0),
    is_active BOOLEAN DEFAULT true,
    effective_from TIMESTAMP WITH TIME ZONE NOT NULL,
    effective_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (max_bet >= min_bet),
    CHECK (effective_until > effective_from)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_game_management_status ON game_management(status);
CREATE INDEX IF NOT EXISTS idx_game_management_provider ON game_management(provider);
CREATE INDEX IF NOT EXISTS idx_game_management_enabled ON game_management(enabled);
CREATE INDEX IF NOT EXISTS idx_game_management_created_at ON game_management(created_at);

CREATE INDEX IF NOT EXISTS idx_house_edge_game_type ON house_edge_management(game_type);
CREATE INDEX IF NOT EXISTS idx_house_edge_game_variant ON house_edge_management(game_variant);
CREATE INDEX IF NOT EXISTS idx_house_edge_is_active ON house_edge_management(is_active);
CREATE INDEX IF NOT EXISTS idx_house_edge_created_at ON house_edge_management(created_at);

-- Insert some sample data for game_management
INSERT INTO game_management (name, status, photo, price, enabled, game_id, internal_name, integration_partner, provider) VALUES
('Classic Slots', 'ACTIVE', 'https://example.com/classic-slots.jpg', 0.00, true, 'classic_slots_001', 'classic_slots', 'groove', 'groove_gaming'),
('Premium Blackjack', 'ACTIVE', 'https://example.com/premium-blackjack.jpg', 5.00, true, 'premium_blackjack_001', 'premium_blackjack', 'groove', 'groove_gaming'),
('Roulette Pro', 'INACTIVE', 'https://example.com/roulette-pro.jpg', 10.00, false, 'roulette_pro_001', 'roulette_pro', 'groove', 'groove_gaming'),
('Poker Championship', 'MAINTENANCE', 'https://example.com/poker-championship.jpg', 15.00, true, 'poker_championship_001', 'poker_championship', 'groove', 'groove_gaming'),
('Baccarat Elite', 'ACTIVE', 'https://example.com/baccarat-elite.jpg', 0.00, true, 'baccarat_elite_001', 'baccarat_elite', 'groove', 'groove_gaming')
ON CONFLICT (game_id) DO NOTHING;

-- Insert some sample data for house_edge_management
INSERT INTO house_edge_management (game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until) VALUES
('slot', 'classic', 0.05, 1.00, 1000.00, true, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z'),
('slot', 'premium', 0.03, 5.00, 5000.00, true, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z'),
('table', 'blackjack', 0.02, 10.00, 2000.00, true, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z'),
('table', 'roulette', 0.027, 5.00, 1000.00, true, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z'),
('live', 'baccarat', 0.01, 25.00, 10000.00, true, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z'),
('live', 'poker', 0.03, 50.00, 5000.00, false, '2024-01-01T00:00:00Z', '2024-12-31T23:59:59Z')
ON CONFLICT DO NOTHING;
