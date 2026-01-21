package game

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// Game represents a game in the database
type Game struct {
	ID                 uuid.UUID        `db:"id"`
	Name               string           `db:"name"`
	Status             string           `db:"status"`
	Timestamp          time.Time        `db:"timestamp"`
	Photo              *string          `db:"photo"`
	Price              *string          `db:"price"`
	Enabled            bool             `db:"enabled"`
	GameID             *string          `db:"game_id"`
	InternalName       *string          `db:"internal_name"`
	IntegrationPartner *string          `db:"integration_partner"`
	Provider           *string          `db:"provider"`
	CreatedAt          time.Time        `db:"created_at"`
	UpdatedAt          time.Time        `db:"updated_at"`
	HouseEdge          *decimal.Decimal `db:"house_edge"`
}

// GameHouseEdge represents a house edge configuration
type GameHouseEdge struct {
	ID             uuid.UUID        `db:"id"`
	GameID         *string          `db:"game_id"`
	GameName       *string          `db:"game_name"`
	GameType       string           `db:"game_type"`
	GameVariant    *string          `db:"game_variant"`
	HouseEdge      decimal.Decimal  `db:"house_edge"`
	MinBet         decimal.Decimal  `db:"min_bet"`
	MaxBet         *decimal.Decimal `db:"max_bet"`
	IsActive       bool             `db:"is_active"`
	EffectiveFrom  time.Time        `db:"effective_from"`
	EffectiveUntil *time.Time       `db:"effective_until"`
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at"`
}

// GameQueryParams represents query parameters for games
type GameQueryParams struct {
	Page      int
	PerPage   int
	Search    string
	Status    string
	Provider  string
	GameID    string // Filter by game_id
	Enabled   *bool
	SortBy    string
	SortOrder string
}

// HouseEdgeQueryParams represents query parameters for house edges
type HouseEdgeQueryParams struct {
	Page        int
	PerPage     int
	Search      string // Search by game_id and game name
	GameID      string // Specific game ID search
	GameType    string
	GameVariant string
	IsActive    *bool
	SortBy      string
	SortOrder   string
}

// GameStats represents game statistics
type GameStats struct {
	TotalGames       int
	ActiveGames      int
	InactiveGames    int
	MaintenanceGames int
	EnabledGames     int
	DisabledGames    int
}

// HouseEdgeStats represents house edge statistics
type HouseEdgeStats struct {
	TotalHouseEdges    int
	ActiveHouseEdges   int
	InactiveHouseEdges int
	UniqueGameTypes    int
	UniqueGameVariants int
}

// GameStorage interface for game operations
type GameStorage interface {
	CreateGame(ctx context.Context, game *Game) (*Game, error)
	GetGameByID(ctx context.Context, id uuid.UUID) (*Game, error)
	GetGameByGameID(ctx context.Context, gameID string) (*Game, error)
	GetGames(ctx context.Context, params GameQueryParams) ([]Game, int, error)
	UpdateGame(ctx context.Context, game *Game) (*Game, error)
	DeleteGame(ctx context.Context, id uuid.UUID) error
	BulkUpdateGameStatus(ctx context.Context, gameIDs []uuid.UUID, status string) error
	GetGameStats(ctx context.Context) (*GameStats, error)
}

// HouseEdgeStorage interface for house edge operations
type HouseEdgeStorage interface {
	CreateHouseEdge(ctx context.Context, houseEdge *GameHouseEdge) (*GameHouseEdge, error)
	GetHouseEdgeByID(ctx context.Context, id uuid.UUID) (*GameHouseEdge, error)
	GetHouseEdges(ctx context.Context, params HouseEdgeQueryParams) ([]GameHouseEdge, int, error)
	UpdateHouseEdge(ctx context.Context, houseEdge *GameHouseEdge) (*GameHouseEdge, error)
	DeleteHouseEdge(ctx context.Context, id uuid.UUID) error
	BulkUpdateHouseEdgeStatus(ctx context.Context, houseEdgeIDs []uuid.UUID, isActive bool) error
	GetHouseEdgeStats(ctx context.Context) (*HouseEdgeStats, error)
	GetHouseEdgesByGameType(ctx context.Context, gameType string) ([]GameHouseEdge, error)
	GetHouseEdgesByGameVariant(ctx context.Context, gameType, gameVariant string) ([]GameHouseEdge, error)
}

// GameStorageImpl implements GameStorage interface
type GameStorageImpl struct {
	db     *persistencedb.PersistenceDB
	logger *zap.Logger
}

// NewGameStorage creates a new game storage instance
func NewGameStorage(db *persistencedb.PersistenceDB, logger *zap.Logger) GameStorage {
	return &GameStorageImpl{
		db:     db,
		logger: logger,
	}
}

// Game Storage Implementation

func (s *GameStorageImpl) CreateGame(ctx context.Context, game *Game) (*Game, error) {
	s.logger.Info("Creating game in database", zap.String("name", game.Name))

	query := `
		INSERT INTO games (name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at`

	var createdGame Game
	err := s.db.GetPool().QueryRow(ctx, query,
		game.Name, game.Status, game.Timestamp, game.Photo, game.Price, game.Enabled,
		game.GameID, game.InternalName, game.IntegrationPartner, game.Provider,
		game.CreatedAt, game.UpdatedAt).Scan(
		&createdGame.ID, &createdGame.Name, &createdGame.Status, &createdGame.Timestamp,
		&createdGame.Photo, &createdGame.Price, &createdGame.Enabled, &createdGame.GameID,
		&createdGame.InternalName, &createdGame.IntegrationPartner, &createdGame.Provider,
		&createdGame.CreatedAt, &createdGame.UpdatedAt)

	if err != nil {
		s.logger.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Game created successfully", zap.String("id", createdGame.ID.String()))
	return &createdGame, nil
}

func (s *GameStorageImpl) GetGameByID(ctx context.Context, id uuid.UUID) (*Game, error) {
	s.logger.Info("Getting game by ID", zap.String("id", id.String()))

	query := `
		SELECT id, name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at
		FROM games WHERE id = $1`

	var game Game
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&game.ID, &game.Name, &game.Status, &game.Timestamp,
		&game.Photo, &game.Price, &game.Enabled, &game.GameID,
		&game.InternalName, &game.IntegrationPartner, &game.Provider,
		&game.CreatedAt, &game.UpdatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		s.logger.Error("Failed to get game by ID", zap.Error(err))
		return nil, err
	}

	return &game, nil
}

func (s *GameStorageImpl) GetGameByGameID(ctx context.Context, gameID string) (*Game, error) {
	s.logger.Info("Getting game by game_id", zap.String("game_id", gameID))

	query := `
		SELECT id, name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at
		FROM games WHERE game_id = $1`

	var game Game
	err := s.db.GetPool().QueryRow(ctx, query, gameID).Scan(
		&game.ID, &game.Name, &game.Status, &game.Timestamp,
		&game.Photo, &game.Price, &game.Enabled, &game.GameID,
		&game.InternalName, &game.IntegrationPartner, &game.Provider,
		&game.CreatedAt, &game.UpdatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		s.logger.Error("Failed to get game by game_id", zap.Error(err))
		return nil, err
	}

	return &game, nil
}

func (s *GameStorageImpl) GetGames(ctx context.Context, params GameQueryParams) ([]Game, int, error) {
	s.logger.Info("Getting games with filters", zap.Any("params", params))

	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+params.Search+"%")
		argIndex++
	}

	if params.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.Provider != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, params.Provider)
		argIndex++
	}

	if params.GameID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_id = $%d", argIndex))
		args = append(args, params.GameID)
		argIndex++
	}

	if params.Enabled != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("enabled = $%d", argIndex))
		args = append(args, *params.Enabled)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM games %s", whereClause)
	var totalCount int
	err := s.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		s.logger.Error("Failed to count games", zap.Error(err))
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := fmt.Sprintf("ORDER BY %s %s", params.SortBy, strings.ToUpper(params.SortOrder))

	// Build LIMIT and OFFSET
	offset := (params.Page - 1) * params.PerPage
	limitOffset := fmt.Sprintf("LIMIT %d OFFSET %d", params.PerPage, offset)

	// Execute query
	query := fmt.Sprintf(`
        SELECT g.id, g.name, g.status, g.timestamp, g.photo, g.price, g.enabled, g.game_id, g.internal_name, g.integration_partner, g.provider, g.created_at, g.updated_at,
               he.house_edge
        FROM games g
        LEFT JOIN LATERAL (
            SELECT house_edge
            FROM game_house_edges
            WHERE is_active = true AND game_id = g.game_id
            ORDER BY effective_from DESC, updated_at DESC
            LIMIT 1
        ) he ON true
        %s %s %s`, whereClause, orderBy, limitOffset)

	rows, err := s.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to query games", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&game.ID, &game.Name, &game.Status, &game.Timestamp,
			&game.Photo, &game.Price, &game.Enabled, &game.GameID,
			&game.InternalName, &game.IntegrationPartner, &game.Provider,
			&game.CreatedAt, &game.UpdatedAt, &game.HouseEdge)
		if err != nil {
			s.logger.Error("Failed to scan game", zap.Error(err))
			return nil, 0, err
		}
		games = append(games, game)
	}

	s.logger.Info("Games retrieved successfully", zap.Int("count", len(games)))
	return games, totalCount, nil
}

func (s *GameStorageImpl) UpdateGame(ctx context.Context, game *Game) (*Game, error) {
	s.logger.Info("Updating game", zap.String("id", game.ID.String()))

	query := `
		UPDATE games 
		SET name = $1, status = $2, timestamp = $3, photo = $4, price = $5, enabled = $6, 
		    game_id = $7, internal_name = $8, integration_partner = $9, provider = $10, updated_at = $11
		WHERE id = $12
		RETURNING id, name, status, timestamp, photo, price, enabled, game_id, internal_name, integration_partner, provider, created_at, updated_at`

	var updatedGame Game
	err := s.db.GetPool().QueryRow(ctx, query,
		game.Name, game.Status, game.Timestamp, game.Photo, game.Price, game.Enabled,
		game.GameID, game.InternalName, game.IntegrationPartner, game.Provider,
		game.UpdatedAt, game.ID).Scan(
		&updatedGame.ID, &updatedGame.Name, &updatedGame.Status, &updatedGame.Timestamp,
		&updatedGame.Photo, &updatedGame.Price, &updatedGame.Enabled, &updatedGame.GameID,
		&updatedGame.InternalName, &updatedGame.IntegrationPartner, &updatedGame.Provider,
		&updatedGame.CreatedAt, &updatedGame.UpdatedAt)

	if err != nil {
		s.logger.Error("Failed to update game", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Game updated successfully", zap.String("id", updatedGame.ID.String()))
	return &updatedGame, nil
}

func (s *GameStorageImpl) DeleteGame(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting game", zap.String("id", id.String()))

	query := "DELETE FROM games WHERE id = $1"
	result, err := s.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		s.logger.Error("Failed to delete game", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Warn("No game found to delete", zap.String("id", id.String()))
		return fmt.Errorf("game not found")
	}

	s.logger.Info("Game deleted successfully", zap.String("id", id.String()))
	return nil
}

func (s *GameStorageImpl) BulkUpdateGameStatus(ctx context.Context, gameIDs []uuid.UUID, status string) error {
	s.logger.Info("Bulk updating game status", zap.String("status", status), zap.Int("count", len(gameIDs)))

	if len(gameIDs) == 0 {
		return nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(gameIDs))
	args := make([]interface{}, len(gameIDs)+1)

	for i, id := range gameIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(gameIDs)] = status

	query := fmt.Sprintf("UPDATE games SET status = $%d, updated_at = NOW() WHERE id IN (%s)",
		len(gameIDs)+1, strings.Join(placeholders, ","))

	result, err := s.db.GetPool().Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to bulk update game status", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	s.logger.Info("Bulk game status update completed",
		zap.String("status", status),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

func (s *GameStorageImpl) GetGameStats(ctx context.Context) (*GameStats, error) {
	s.logger.Info("Getting game statistics")

	query := `
		SELECT 
			COUNT(*) as total_games,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active_games,
			COUNT(CASE WHEN status = 'INACTIVE' THEN 1 END) as inactive_games,
			COUNT(CASE WHEN status = 'MAINTENANCE' THEN 1 END) as maintenance_games,
			COUNT(CASE WHEN enabled = true THEN 1 END) as enabled_games,
			COUNT(CASE WHEN enabled = false THEN 1 END) as disabled_games
		FROM games`

	var stats GameStats
	err := s.db.GetPool().QueryRow(ctx, query).Scan(
		&stats.TotalGames, &stats.ActiveGames, &stats.InactiveGames,
		&stats.MaintenanceGames, &stats.EnabledGames, &stats.DisabledGames)

	if err != nil {
		s.logger.Error("Failed to get game statistics", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Game statistics retrieved", zap.Any("stats", stats))
	return &stats, nil
}
