-- Drop indexes
DROP INDEX IF EXISTS idx_game_management_created_at;
DROP INDEX IF EXISTS idx_game_management_enabled;
DROP INDEX IF EXISTS idx_game_management_provider;
DROP INDEX IF EXISTS idx_game_management_status;

DROP INDEX IF EXISTS idx_house_edge_created_at;
DROP INDEX IF EXISTS idx_house_edge_is_active;
DROP INDEX IF EXISTS idx_house_edge_game_variant;
DROP INDEX IF EXISTS idx_house_edge_game_type;

-- Drop tables
DROP TABLE IF EXISTS house_edge_management;
DROP TABLE IF EXISTS game_management;
