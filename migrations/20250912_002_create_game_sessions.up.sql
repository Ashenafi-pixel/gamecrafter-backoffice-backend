-- Create function for unique GrooveTech session ID generation
CREATE OR REPLACE FUNCTION generate_groove_session_id()
RETURNS TEXT AS $$
BEGIN
    RETURN 'Tucan_' || gen_random_uuid()::text;
END;
$$ LANGUAGE plpgsql;

-- Create game_sessions table for tracking GrooveTech game launches
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(64) UNIQUE DEFAULT generate_groove_session_id(),
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

-- Create indexes for performance
CREATE INDEX idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX idx_game_sessions_session_id ON game_sessions(session_id);
CREATE INDEX idx_game_sessions_game_id ON game_sessions(game_id);
CREATE INDEX idx_game_sessions_created_at ON game_sessions(created_at);
CREATE INDEX idx_game_sessions_active ON game_sessions(is_active);

-- Create function to clean up expired sessions
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

-- Create trigger to update last_activity on any update
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

-- Insert some sample data for testing
INSERT INTO game_sessions (user_id, game_id, device_type, game_mode, home_url, exit_url, license_type) 
VALUES 
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', '80102', 'desktop', 'real', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao'),
    ('a5e168fb-168e-4183-84c5-d49038ce00b5', '80103', 'mobile', 'demo', 'https://tucanbit.tv/games', 'https://tucanbit.tv/responsible-gaming', 'Curacao');