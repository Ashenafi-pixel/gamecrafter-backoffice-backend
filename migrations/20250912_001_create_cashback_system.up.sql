-- Cashback and Level System Migration
-- This creates a world-class cashback system with multiple tiers and sophisticated tracking

-- Cashback Tiers Configuration (create first since user_levels references it)
CREATE TABLE cashback_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_name VARCHAR(50) NOT NULL UNIQUE, -- Bronze, Silver, Gold, Platinum, Diamond
    tier_level INTEGER NOT NULL UNIQUE, -- 1, 2, 3, 4, 5
    min_ggr_required DECIMAL(20,8) NOT NULL, -- Minimum GGR to reach this tier
    cashback_percentage DECIMAL(5,2) NOT NULL, -- Cashback percentage (0.00-100.00)
    bonus_multiplier DECIMAL(3,2) DEFAULT 1.00, -- Bonus multiplier for this tier
    daily_cashback_limit DECIMAL(20,8) DEFAULT NULL, -- Daily cashback limit (NULL = unlimited)
    weekly_cashback_limit DECIMAL(20,8) DEFAULT NULL, -- Weekly cashback limit
    monthly_cashback_limit DECIMAL(20,8) DEFAULT NULL, -- Monthly cashback limit
    special_benefits JSONB DEFAULT '{}', -- Special benefits for this tier
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User Levels Table
CREATE TABLE user_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_level INTEGER NOT NULL DEFAULT 1,
    total_ggr DECIMAL(20,8) NOT NULL DEFAULT 0, -- Gross Gaming Revenue
    total_bets DECIMAL(20,8) NOT NULL DEFAULT 0, -- Total amount wagered
    total_wins DECIMAL(20,8) NOT NULL DEFAULT 0, -- Total winnings
    level_progress DECIMAL(5,2) NOT NULL DEFAULT 0, -- Progress to next level (0-100%)
    current_tier_id UUID REFERENCES cashback_tiers(id),
    last_level_up TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Cashback Earnings Tracking
CREATE TABLE cashback_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id UUID NOT NULL REFERENCES cashback_tiers(id),
    earning_type VARCHAR(20) NOT NULL, -- 'bet', 'bonus', 'promotion', 'referral'
    source_bet_id UUID REFERENCES bets(id), -- Reference to the bet that generated this earning
    ggr_amount DECIMAL(20,8) NOT NULL, -- GGR from this specific bet/activity
    cashback_rate DECIMAL(5,2) NOT NULL, -- Rate used for this earning
    earned_amount DECIMAL(20,8) NOT NULL, -- Amount earned
    claimed_amount DECIMAL(20,8) DEFAULT 0, -- Amount claimed
    available_amount DECIMAL(20,8) NOT NULL, -- Available to claim
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'available', 'claimed', 'expired'
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '30 days'), -- Auto-expire after 30 days
    claimed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Cashback Claims History
CREATE TABLE cashback_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    claim_amount DECIMAL(20,8) NOT NULL,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    transaction_id UUID REFERENCES balance_logs(id), -- Reference to balance transaction
    processing_fee DECIMAL(20,8) DEFAULT 0, -- Any processing fees
    net_amount DECIMAL(20,8) NOT NULL, -- Amount after fees
    claimed_earnings JSONB NOT NULL, -- Array of earning IDs that were claimed
    admin_notes TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Game House Edge Configuration
CREATE TABLE game_house_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_type VARCHAR(50) NOT NULL, -- 'plinko', 'crash', 'dice', 'blackjack', etc.
    game_variant VARCHAR(50), -- Specific variant if applicable
    house_edge DECIMAL(5,4) NOT NULL, -- House edge as decimal (0.02 = 2%)
    min_bet DECIMAL(20,8) DEFAULT 0,
    max_bet DECIMAL(20,8) DEFAULT NULL,
    is_active BOOLEAN DEFAULT true,
    effective_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(game_type, game_variant, effective_from)
);

-- Cashback Promotions (Special events, bonuses)
CREATE TABLE cashback_promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_name VARCHAR(100) NOT NULL,
    description TEXT,
    promotion_type VARCHAR(20) NOT NULL, -- 'boost', 'bonus', 'special'
    boost_multiplier DECIMAL(3,2) DEFAULT 1.00, -- Multiplier for cashback rate
    bonus_amount DECIMAL(20,8) DEFAULT 0, -- Fixed bonus amount
    min_bet_amount DECIMAL(20,8) DEFAULT 0,
    max_bonus_amount DECIMAL(20,8) DEFAULT NULL,
    target_tiers INTEGER[] DEFAULT NULL, -- Array of tier levels this applies to
    target_games VARCHAR(50)[] DEFAULT NULL, -- Array of game types this applies to
    is_active BOOLEAN DEFAULT true,
    starts_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ends_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_user_levels_user_id ON user_levels(user_id);
CREATE INDEX idx_user_levels_current_level ON user_levels(current_level);
CREATE INDEX idx_cashback_earnings_user_id ON cashback_earnings(user_id);
CREATE INDEX idx_cashback_earnings_status ON cashback_earnings(status);
CREATE INDEX idx_cashback_earnings_created_at ON cashback_earnings(created_at);
CREATE INDEX idx_cashback_earnings_expires_at ON cashback_earnings(expires_at);
CREATE INDEX idx_cashback_claims_user_id ON cashback_claims(user_id);
CREATE INDEX idx_cashback_claims_status ON cashback_claims(status);
CREATE INDEX idx_game_house_edges_game_type ON game_house_edges(game_type);
CREATE INDEX idx_cashback_promotions_active ON cashback_promotions(is_active, starts_at, ends_at);

-- Insert default cashback tiers
INSERT INTO cashback_tiers (tier_name, tier_level, min_ggr_required, cashback_percentage, bonus_multiplier, daily_cashback_limit, special_benefits) VALUES
('Bronze', 1, 0, 0.50, 1.00, 50.00, '{"priority_support": false, "exclusive_games": false, "withdrawal_limit": 1000}'),
('Silver', 2, 1000, 1.00, 1.10, 100.00, '{"priority_support": true, "exclusive_games": false, "withdrawal_limit": 2500}'),
('Gold', 3, 5000, 1.50, 1.25, 250.00, '{"priority_support": true, "exclusive_games": true, "withdrawal_limit": 5000}'),
('Platinum', 4, 15000, 2.00, 1.50, 500.00, '{"priority_support": true, "exclusive_games": true, "withdrawal_limit": 10000, "personal_manager": true}'),
('Diamond', 5, 50000, 2.50, 2.00, 1000.00, '{"priority_support": true, "exclusive_games": true, "withdrawal_limit": 25000, "personal_manager": true, "vip_events": true}');

-- Insert default game house edges
INSERT INTO game_house_edges (game_type, house_edge, min_bet, max_bet) VALUES
('plinko', 0.0200, 0.10, 1000.00),
('crash', 0.0100, 0.10, 1000.00),
('dice', 0.0100, 0.10, 1000.00),
('blackjack', 0.0048, 1.00, 1000.00),
('roulette', 0.0270, 1.00, 1000.00),
('baccarat', 0.0106, 1.00, 1000.00),
('poker', 0.0200, 1.00, 1000.00),
('slots', 0.0300, 0.10, 1000.00);

-- Add house_edge column to existing bets table if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bets' AND column_name = 'house_edge') THEN
        ALTER TABLE bets ADD COLUMN house_edge DECIMAL(5,4) DEFAULT 0.0200;
    END IF;
END $$;

-- Function to check and update user level
CREATE OR REPLACE FUNCTION check_and_update_user_level(p_user_id UUID)
RETURNS VOID AS $$
DECLARE
    current_tier RECORD;
    new_tier RECORD;
    user_stats RECORD;
BEGIN
    -- Get current user stats
    SELECT ul.*, ct.tier_name, ct.cashback_percentage
    INTO user_stats
    FROM user_levels ul
    LEFT JOIN cashback_tiers ct ON ul.current_tier_id = ct.id
    WHERE ul.user_id = p_user_id;
    
    -- Find the highest tier the user qualifies for
    SELECT ct.*
    INTO new_tier
    FROM cashback_tiers ct
    WHERE ct.is_active = true 
    AND ct.min_ggr_required <= user_stats.total_ggr
    ORDER BY ct.tier_level DESC
    LIMIT 1;
    
    -- Update user level if they qualify for a higher tier
    IF new_tier.id IS NOT NULL AND (user_stats.current_tier_id IS NULL OR new_tier.tier_level > user_stats.current_level) THEN
        UPDATE user_levels 
        SET 
            current_level = new_tier.tier_level,
            current_tier_id = new_tier.id,
            level_progress = CASE 
                WHEN new_tier.tier_level = 5 THEN 100.00 -- Diamond is max level
                ELSE LEAST(100.00, (user_stats.total_ggr / (
                    SELECT min_ggr_required 
                    FROM cashback_tiers 
                    WHERE tier_level = new_tier.tier_level + 1 
                    AND is_active = true
                )) * 100.00)
            END,
            last_level_up = NOW(),
            updated_at = NOW()
        WHERE user_id = p_user_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update user_levels when cashback_earnings are created
CREATE OR REPLACE FUNCTION update_user_level_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Update user level statistics
    UPDATE user_levels 
    SET 
        total_ggr = total_ggr + NEW.ggr_amount,
        updated_at = NOW()
    WHERE user_id = NEW.user_id;
    
    -- Check if user should level up
    PERFORM check_and_update_user_level(NEW.user_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_level_stats
    AFTER INSERT ON cashback_earnings
    FOR EACH ROW
    EXECUTE FUNCTION update_user_level_stats();