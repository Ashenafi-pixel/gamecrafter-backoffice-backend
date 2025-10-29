-- Drop indexes
DROP INDEX IF EXISTS idx_sport_bets_transaction_id;
DROP INDEX IF EXISTS idx_sport_bets_user_id;
DROP INDEX IF EXISTS idx_sport_bets_bet_status;
DROP INDEX IF EXISTS idx_sport_bets_placed_at;

-- Drop tables
DROP TABLE IF EXISTS sport_bets;
