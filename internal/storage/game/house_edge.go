package game

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// HouseEdgeStorageImpl implements HouseEdgeStorage interface
type HouseEdgeStorageImpl struct {
	db     *persistencedb.PersistenceDB
	logger *zap.Logger
}

func NewHouseEdgeStorage(db *persistencedb.PersistenceDB, logger *zap.Logger) HouseEdgeStorage {
	return &HouseEdgeStorageImpl{
		db:     db,
		logger: logger,
	}
}

func (s *HouseEdgeStorageImpl) CreateHouseEdge(ctx context.Context, houseEdge *GameHouseEdge) (*GameHouseEdge, error) {
	s.logger.Info("Creating house edge in database",
		zap.String("game_type", houseEdge.GameType),
		zap.String("game_variant", getStringValue(houseEdge.GameVariant)))

	query := `
		INSERT INTO game_house_edges (game_id, game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, game_id, game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until, created_at, updated_at`

	var createdHouseEdge GameHouseEdge
	err := s.db.GetPool().QueryRow(ctx, query,
		houseEdge.GameID, houseEdge.GameType, houseEdge.GameVariant, houseEdge.HouseEdge, houseEdge.MinBet, houseEdge.MaxBet,
		houseEdge.IsActive, houseEdge.EffectiveFrom, houseEdge.EffectiveUntil,
		houseEdge.CreatedAt, houseEdge.UpdatedAt).Scan(
		&createdHouseEdge.ID, &createdHouseEdge.GameID, &createdHouseEdge.GameType, &createdHouseEdge.GameVariant,
		&createdHouseEdge.HouseEdge, &createdHouseEdge.MinBet, &createdHouseEdge.MaxBet,
		&createdHouseEdge.IsActive, &createdHouseEdge.EffectiveFrom, &createdHouseEdge.EffectiveUntil,
		&createdHouseEdge.CreatedAt, &createdHouseEdge.UpdatedAt)

	if err != nil {
		s.logger.Error("Failed to create house edge", zap.Error(err))
		return nil, err
	}

	s.logger.Info("House edge created successfully", zap.String("id", createdHouseEdge.ID.String()))
	return &createdHouseEdge, nil
}

func (s *HouseEdgeStorageImpl) GetHouseEdgeByID(ctx context.Context, id uuid.UUID) (*GameHouseEdge, error) {
	s.logger.Info("Getting house edge by ID", zap.String("id", id.String()))

	query := `
		SELECT ghe.id, ghe.game_id, ghe.game_type, ghe.game_variant, ghe.house_edge, ghe.min_bet, ghe.max_bet, ghe.is_active, ghe.effective_from, ghe.effective_until, ghe.created_at, ghe.updated_at,
		       g.game_id as actual_game_id, g.name as actual_game_name
		FROM game_house_edges ghe
		LEFT JOIN games g ON ghe.game_id = g.game_id
		WHERE ghe.id = $1`

	var houseEdge GameHouseEdge
	var actualGameID, actualGameName *string
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&houseEdge.ID, &houseEdge.GameID, &houseEdge.GameType, &houseEdge.GameVariant,
		&houseEdge.HouseEdge, &houseEdge.MinBet, &houseEdge.MaxBet,
		&houseEdge.IsActive, &houseEdge.EffectiveFrom, &houseEdge.EffectiveUntil,
		&houseEdge.CreatedAt, &houseEdge.UpdatedAt,
		&actualGameID, &actualGameName)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		s.logger.Error("Failed to get house edge by ID", zap.Error(err))
		return nil, err
	}

	// Use actual game data from games table if available, otherwise use stored values
	if actualGameID != nil && *actualGameID != "" {
		houseEdge.GameID = actualGameID
	}
	if actualGameName != nil && *actualGameName != "" {
		houseEdge.GameName = actualGameName
	}

	return &houseEdge, nil
}

func (s *HouseEdgeStorageImpl) GetHouseEdges(ctx context.Context, params HouseEdgeQueryParams) ([]GameHouseEdge, int, error) {
	s.logger.Info("Getting house edges with filters", zap.Any("params", params))

	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.GameType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_type = $%d", argIndex))
		args = append(args, params.GameType)
		argIndex++
	}

	if params.GameVariant != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_variant = $%d", argIndex))
		args = append(args, params.GameVariant)
		argIndex++
	}

	if params.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *params.IsActive)
		argIndex++
	}

	// Add search functionality
	if params.Search != "" {
		searchCondition := fmt.Sprintf("(game_type ILIKE $%d OR game_variant ILIKE $%d)", argIndex, argIndex+1)
		whereConditions = append(whereConditions, searchCondition)
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if params.GameID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("game_id = $%d", argIndex))
		args = append(args, params.GameID)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM game_house_edges %s", whereClause)
	var totalCount int
	err := s.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		s.logger.Error("Failed to count house edges", zap.Error(err))
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := fmt.Sprintf("ORDER BY %s %s", params.SortBy, strings.ToUpper(params.SortOrder))

	// Build LIMIT and OFFSET
	offset := (params.Page - 1) * params.PerPage
	limitOffset := fmt.Sprintf("LIMIT %d OFFSET %d", params.PerPage, offset)

	// Execute query
	query := fmt.Sprintf(`
		SELECT ghe.id, ghe.game_id, ghe.game_type, ghe.game_variant, ghe.house_edge, ghe.min_bet, ghe.max_bet, ghe.is_active, ghe.effective_from, ghe.effective_until, ghe.created_at, ghe.updated_at,
		       g.game_id as actual_game_id, g.name as actual_game_name
		FROM game_house_edges ghe
		LEFT JOIN games g ON ghe.game_id = g.game_id
		%s %s %s`, whereClause, orderBy, limitOffset)

	rows, err := s.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to query house edges", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var houseEdges []GameHouseEdge
	for rows.Next() {
		var houseEdge GameHouseEdge
		var actualGameID, actualGameName *string
		err := rows.Scan(
			&houseEdge.ID, &houseEdge.GameID, &houseEdge.GameType, &houseEdge.GameVariant,
			&houseEdge.HouseEdge, &houseEdge.MinBet, &houseEdge.MaxBet,
			&houseEdge.IsActive, &houseEdge.EffectiveFrom, &houseEdge.EffectiveUntil,
			&houseEdge.CreatedAt, &houseEdge.UpdatedAt,
			&actualGameID, &actualGameName)
		if err != nil {
			s.logger.Error("Failed to scan house edge", zap.Error(err))
			return nil, 0, err
		}

		// Use actual game data from games table if available, otherwise use stored values
		if actualGameID != nil && *actualGameID != "" {
			houseEdge.GameID = actualGameID
		}
		if actualGameName != nil && *actualGameName != "" {
			houseEdge.GameName = actualGameName
		}

		houseEdges = append(houseEdges, houseEdge)
	}

	s.logger.Info("House edges retrieved successfully", zap.Int("count", len(houseEdges)))
	return houseEdges, totalCount, nil
}

func (s *HouseEdgeStorageImpl) UpdateHouseEdge(ctx context.Context, houseEdge *GameHouseEdge) (*GameHouseEdge, error) {
	s.logger.Info("Updating house edge", zap.String("id", houseEdge.ID.String()))

	query := `
		UPDATE game_house_edges 
		SET game_id = $1, game_type = $2, game_variant = $3, house_edge = $4, min_bet = $5, max_bet = $6, 
		    is_active = $7, effective_from = $8, effective_until = $9, updated_at = $10
		WHERE id = $11
		RETURNING id, game_id, game_type, game_variant, house_edge, min_bet, max_bet, is_active, effective_from, effective_until, created_at, updated_at`

	var updatedHouseEdge GameHouseEdge
	err := s.db.GetPool().QueryRow(ctx, query,
		houseEdge.GameID, houseEdge.GameType, houseEdge.GameVariant, houseEdge.HouseEdge, houseEdge.MinBet, houseEdge.MaxBet,
		houseEdge.IsActive, houseEdge.EffectiveFrom, houseEdge.EffectiveUntil,
		houseEdge.UpdatedAt, houseEdge.ID).Scan(
		&updatedHouseEdge.ID, &updatedHouseEdge.GameID, &updatedHouseEdge.GameType, &updatedHouseEdge.GameVariant,
		&updatedHouseEdge.HouseEdge, &updatedHouseEdge.MinBet, &updatedHouseEdge.MaxBet,
		&updatedHouseEdge.IsActive, &updatedHouseEdge.EffectiveFrom, &updatedHouseEdge.EffectiveUntil,
		&updatedHouseEdge.CreatedAt, &updatedHouseEdge.UpdatedAt)

	if err != nil {
		s.logger.Error("Failed to update house edge", zap.Error(err))
		return nil, err
	}

	s.logger.Info("House edge updated successfully", zap.String("id", updatedHouseEdge.ID.String()))
	return &updatedHouseEdge, nil
}

func (s *HouseEdgeStorageImpl) DeleteHouseEdge(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting house edge", zap.String("id", id.String()))

	query := "DELETE FROM game_house_edges WHERE id = $1"
	result, err := s.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		s.logger.Error("Failed to delete house edge", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Warn("No house edge found to delete", zap.String("id", id.String()))
		return fmt.Errorf("house edge not found")
	}

	s.logger.Info("House edge deleted successfully", zap.String("id", id.String()))
	return nil
}

func (s *HouseEdgeStorageImpl) BulkUpdateHouseEdgeStatus(ctx context.Context, houseEdgeIDs []uuid.UUID, isActive bool) error {
	s.logger.Info("Bulk updating house edge status", zap.Bool("is_active", isActive), zap.Int("count", len(houseEdgeIDs)))

	if len(houseEdgeIDs) == 0 {
		return nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(houseEdgeIDs))
	args := make([]interface{}, len(houseEdgeIDs)+1)

	for i, id := range houseEdgeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(houseEdgeIDs)] = isActive

	query := fmt.Sprintf("UPDATE game_house_edges SET is_active = $%d, updated_at = NOW() WHERE id IN (%s)",
		len(houseEdgeIDs)+1, strings.Join(placeholders, ","))

	result, err := s.db.GetPool().Exec(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to bulk update house edge status", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	s.logger.Info("Bulk house edge status update completed",
		zap.Bool("is_active", isActive),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

func (s *HouseEdgeStorageImpl) GetHouseEdgeStats(ctx context.Context) (*HouseEdgeStats, error) {
	s.logger.Info("Getting house edge statistics")

	query := `
		SELECT 
			COUNT(*) as total_house_edges,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_house_edges,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_house_edges,
			COUNT(DISTINCT game_type) as unique_game_types,
			COUNT(DISTINCT game_variant) as unique_game_variants
		FROM game_house_edges`

	var stats HouseEdgeStats
	err := s.db.GetPool().QueryRow(ctx, query).Scan(
		&stats.TotalHouseEdges, &stats.ActiveHouseEdges, &stats.InactiveHouseEdges,
		&stats.UniqueGameTypes, &stats.UniqueGameVariants)

	if err != nil {
		s.logger.Error("Failed to get house edge statistics", zap.Error(err))
		return nil, err
	}

	s.logger.Info("House edge statistics retrieved", zap.Any("stats", stats))
	return &stats, nil
}

func (s *HouseEdgeStorageImpl) GetHouseEdgesByGameType(ctx context.Context, gameType string) ([]GameHouseEdge, error) {
	s.logger.Info("Getting house edges by game type", zap.String("game_type", gameType))

	query := `
		SELECT ghe.id, ghe.game_id, ghe.game_type, ghe.game_variant, ghe.house_edge, ghe.min_bet, ghe.max_bet, ghe.is_active, ghe.effective_from, ghe.effective_until, ghe.created_at, ghe.updated_at,
		       g.game_id as actual_game_id, g.name as actual_game_name
		FROM game_house_edges ghe
		LEFT JOIN games g ON ghe.game_id = g.game_id
		WHERE ghe.game_type = $1 ORDER BY ghe.created_at DESC`

	rows, err := s.db.GetPool().Query(ctx, query, gameType)
	if err != nil {
		s.logger.Error("Failed to query house edges by game type", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var houseEdges []GameHouseEdge
	for rows.Next() {
		var houseEdge GameHouseEdge
		var actualGameID, actualGameName *string
		err := rows.Scan(
			&houseEdge.ID, &houseEdge.GameID, &houseEdge.GameType, &houseEdge.GameVariant,
			&houseEdge.HouseEdge, &houseEdge.MinBet, &houseEdge.MaxBet,
			&houseEdge.IsActive, &houseEdge.EffectiveFrom, &houseEdge.EffectiveUntil,
			&houseEdge.CreatedAt, &houseEdge.UpdatedAt,
			&actualGameID, &actualGameName)
		if err != nil {
			s.logger.Error("Failed to scan house edge", zap.Error(err))
			return nil, err
		}

		// Use actual game data from games table if available, otherwise use stored values
		if actualGameID != nil && *actualGameID != "" {
			houseEdge.GameID = actualGameID
		}
		if actualGameName != nil && *actualGameName != "" {
			houseEdge.GameName = actualGameName
		}

		houseEdges = append(houseEdges, houseEdge)
	}

	s.logger.Info("House edges by game type retrieved", zap.String("game_type", gameType), zap.Int("count", len(houseEdges)))
	return houseEdges, nil
}

func (s *HouseEdgeStorageImpl) GetHouseEdgesByGameVariant(ctx context.Context, gameType, gameVariant string) ([]GameHouseEdge, error) {
	s.logger.Info("Getting house edges by game variant",
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant))

	query := `
		SELECT ghe.id, ghe.game_id, ghe.game_type, ghe.game_variant, ghe.house_edge, ghe.min_bet, ghe.max_bet, ghe.is_active, ghe.effective_from, ghe.effective_until, ghe.created_at, ghe.updated_at,
		       g.game_id as actual_game_id, g.name as actual_game_name
		FROM game_house_edges ghe
		LEFT JOIN games g ON ghe.game_id = g.game_id
		WHERE ghe.game_type = $1 AND ghe.game_variant = $2 ORDER BY ghe.created_at DESC`

	rows, err := s.db.GetPool().Query(ctx, query, gameType, gameVariant)
	if err != nil {
		s.logger.Error("Failed to query house edges by game variant", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var houseEdges []GameHouseEdge
	for rows.Next() {
		var houseEdge GameHouseEdge
		var actualGameID, actualGameName *string
		err := rows.Scan(
			&houseEdge.ID, &houseEdge.GameID, &houseEdge.GameType, &houseEdge.GameVariant,
			&houseEdge.HouseEdge, &houseEdge.MinBet, &houseEdge.MaxBet,
			&houseEdge.IsActive, &houseEdge.EffectiveFrom, &houseEdge.EffectiveUntil,
			&houseEdge.CreatedAt, &houseEdge.UpdatedAt,
			&actualGameID, &actualGameName)
		if err != nil {
			s.logger.Error("Failed to scan house edge", zap.Error(err))
			return nil, err
		}

		// Use actual game data from games table if available, otherwise use stored values
		if actualGameID != nil && *actualGameID != "" {
			houseEdge.GameID = actualGameID
		}
		if actualGameName != nil && *actualGameName != "" {
			houseEdge.GameName = actualGameName
		}
		houseEdges = append(houseEdges, houseEdge)
	}

	s.logger.Info("House edges by game variant retrieved",
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant),
		zap.Int("count", len(houseEdges)))
	return houseEdges, nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
