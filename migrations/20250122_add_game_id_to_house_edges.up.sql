-- Add game_id column to game_house_edges table
ALTER TABLE game_house_edges 
ADD COLUMN game_id VARCHAR(255);

-- Add index for game_id for better performance
CREATE INDEX IF NOT EXISTS idx_game_house_edges_game_id ON game_house_edges(game_id);

-- Add comment to document the new column
COMMENT ON COLUMN game_house_edges.game_id IS 'The actual game ID from the game management system (e.g., 54512456)';
