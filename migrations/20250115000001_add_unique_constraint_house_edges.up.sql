-- Add unique constraint to prevent duplicate game_id in house_edges
-- This ensures each game can only have one active house edge configuration

-- First, remove any existing duplicates (keep the most recent one)
WITH duplicates AS (
  SELECT 
    id,
    ROW_NUMBER() OVER (
      PARTITION BY game_id 
      ORDER BY created_at DESC
    ) as rn
  FROM house_edges
  WHERE is_active = true
)
DELETE FROM house_edges 
WHERE id IN (
  SELECT id FROM duplicates WHERE rn > 1
);

-- Add unique constraint for active house edges
CREATE UNIQUE INDEX IF NOT EXISTS idx_house_edges_unique_active_game 
ON house_edges (game_id) 
WHERE is_active = true;

-- Add comment explaining the constraint
COMMENT ON INDEX idx_house_edges_unique_active_game IS 'Ensures only one active house edge configuration per game';
