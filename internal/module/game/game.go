package game

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/game"

	"go.uber.org/zap"
)

type GameService struct {
	gameStorage     game.GameStorage
	logger          *zap.Logger
	providerStorage storage.Provider
}

func NewGameService(gameStorage game.GameStorage, logger *zap.Logger, providerStorage storage.Provider) *GameService {
	return &GameService{
		gameStorage:     gameStorage,
		logger:          logger,
		providerStorage: providerStorage,
	}
}

func (s *GameService) CreateGame(ctx *gin.Context, req dto.CreateGameRequest) (*dto.GameResponse, error) {
	s.logger.Info("Creating new game", zap.String("name", req.Name))
	providorUUID, err := uuid.Parse(req.ProvidorID.String())
	if err != nil {
		s.logger.Error("Invalid provider ID", zap.String("providor_id", req.ProvidorID.String()), zap.Error(err))
		return nil, errors.ErrInvalidUserInput.Wrap(err, "invalid provider ID")
	}
	fmt.Println("Provoidor Id", providorUUID)
	provider, exists, err := s.providerStorage.GetProvidorByID(ctx, providorUUID)
	if err != nil {
		s.logger.Error("Failed to get provider by ID", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get provider")
	}
	if !exists {
		s.logger.Error("Provider does not exist", zap.String("provider_id", providorUUID.String()))
		return nil, errors.ErrResourceNotFound.New("provider does not exist")
	}
	s.logger.Info("Found provider", zap.String("provider_id", provider.ID.String()))
	fmt.Printf("Provider details: ID=%s, Name=%s\n", provider.ID.String(), provider.Name)

	// Convert DTO to storage model
	providorString := provider.ID.String()
	gameModel := &game.Game{
		Name:               req.Name,
		Status:             req.Status,
		Photo:              req.Photo,
		Price:              req.Price,
		Enabled:            req.Enabled,
		GameID:             req.GameID,
		InternalName:       req.InternalName,
		IntegrationPartner: req.IntegrationPartner,
		Provider:           &providorString,
		Timestamp:          time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	createdGame, err := s.gameStorage.CreateGame(ctx, gameModel)
	if err != nil {
		s.logger.Error("Failed to create game", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create game")
	}

	// Convert storage model to response DTO
	response := s.convertToGameResponse(createdGame)

	s.logger.Info("Game created successfully",
		zap.String("game_id", createdGame.ID.String()),
		zap.String("name", createdGame.Name))

	return response, nil
}

func (s *GameService) GetGameByID(ctx *gin.Context, id uuid.UUID) (*dto.GameResponse, error) {
	s.logger.Info("Getting game by ID", zap.String("id", id.String()))

	game, err := s.gameStorage.GetGameByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get game by ID", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game")
	}

	if game == nil {
		return nil, errors.ErrResourceNotFound.New("game not found")
	}

	response := s.convertToGameResponse(game)
	return response, nil
}

func (s *GameService) GetGames(ctx *gin.Context, params dto.GameQueryParams) (*dto.GameListResponse, error) {
	s.logger.Info("Getting games with filters",
		zap.Int("page", params.Page),
		zap.Int("per_page", params.PerPage),
		zap.String("search", params.Search),
		zap.String("status", params.Status))

	games, totalCount, err := s.gameStorage.GetGames(ctx, game.GameQueryParams{
		Page:      params.Page,
		PerPage:   params.PerPage,
		Search:    params.Search,
		Status:    params.Status,
		Provider:  params.Provider,
		GameID:    params.GameID,
		Enabled:   params.Enabled,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	})
	if err != nil {
		s.logger.Error("Failed to get games", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get games")
	}

	// Convert to response DTOs
	gameResponses := make([]dto.GameResponse, len(games))
	for i, game := range games {
		gameResponses[i] = *s.convertToGameResponse(&game)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.PerPage)))

	response := &dto.GameListResponse{
		Games:      gameResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	s.logger.Info("Games retrieved successfully",
		zap.Int("count", len(games)),
		zap.Int("total_count", totalCount))

	return response, nil
}

func (s *GameService) UpdateGame(ctx *gin.Context, id uuid.UUID, req dto.UpdateGameRequest) (*dto.GameResponse, error) {
	s.logger.Info("Updating game", zap.String("id", id.String()))

	// Check if game exists
	existingGame, err := s.gameStorage.GetGameByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get game for update", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game")
	}

	if existingGame == nil {
		return nil, errors.ErrResourceNotFound.New("game not found")
	}

	// Update fields if provided
	if req.Name != nil {
		existingGame.Name = *req.Name
	}
	if req.Status != nil {
		existingGame.Status = *req.Status
	}
	if req.Photo != nil {
		existingGame.Photo = req.Photo
	}
	if req.Price != nil {
		existingGame.Price = req.Price
	}
	if req.Enabled != nil {
		existingGame.Enabled = *req.Enabled
	}
	if req.GameID != nil {
		existingGame.GameID = req.GameID
	}
	if req.InternalName != nil {
		existingGame.InternalName = req.InternalName
	}
	if req.IntegrationPartner != nil {
		existingGame.IntegrationPartner = req.IntegrationPartner
	}
	if req.Provider != nil {
		existingGame.Provider = req.Provider
	}

	existingGame.UpdatedAt = time.Now()

	updatedGame, err := s.gameStorage.UpdateGame(ctx, existingGame)
	if err != nil {
		s.logger.Error("Failed to update game", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update game")
	}

	response := s.convertToGameResponse(updatedGame)

	s.logger.Info("Game updated successfully",
		zap.String("game_id", updatedGame.ID.String()),
		zap.String("name", updatedGame.Name))

	return response, nil
}

func (s *GameService) DeleteGame(ctx *gin.Context, id uuid.UUID) error {
	s.logger.Info("Deleting game", zap.String("id", id.String()))

	// Check if game exists
	existingGame, err := s.gameStorage.GetGameByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get game for deletion", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get game")
	}

	if existingGame == nil {
		return errors.ErrResourceNotFound.New("game not found")
	}

	err = s.gameStorage.DeleteGame(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete game", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to delete game")
	}

	s.logger.Info("Game deleted successfully",
		zap.String("game_id", id.String()),
		zap.String("name", existingGame.Name))

	return nil
}

func (s *GameService) BulkUpdateGameStatus(ctx *gin.Context, req dto.BulkUpdateGameStatusRequest) error {
	s.logger.Info("Bulk updating game status",
		zap.String("status", req.Status),
		zap.Int("count", len(req.GameIDs)))

	err := s.gameStorage.BulkUpdateGameStatus(ctx, req.GameIDs, req.Status)
	if err != nil {
		s.logger.Error("Failed to bulk update game status", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to bulk update game status")
	}

	s.logger.Info("Game statuses updated successfully",
		zap.String("status", req.Status),
		zap.Int("count", len(req.GameIDs)))

	return nil
}

func (s *GameService) GetGameStats(ctx *gin.Context) (*dto.GameManagementStats, error) {
	s.logger.Info("Getting game statistics")

	stats, err := s.gameStorage.GetGameStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get game statistics", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get game statistics")
	}

	response := &dto.GameManagementStats{
		TotalGames:       stats.TotalGames,
		ActiveGames:      stats.ActiveGames,
		InactiveGames:    stats.InactiveGames,
		MaintenanceGames: stats.MaintenanceGames,
		EnabledGames:     stats.EnabledGames,
		DisabledGames:    stats.DisabledGames,
	}

	s.logger.Info("Game statistics retrieved successfully",
		zap.Int("total_games", stats.TotalGames),
		zap.Int("active_games", stats.ActiveGames))

	return response, nil
}

func (s *GameService) convertToGameResponse(game *game.Game) *dto.GameResponse {
	var rtp *string
	if game.HouseEdge != nil {
		// rtp_percent = 100 - house_edge (no extra scaling)
		he := *game.HouseEdge
		r := decimal.NewFromInt(100).Sub(he).StringFixed(2)
		rtp = &r
	}

	return &dto.GameResponse{
		ID:                 game.ID,
		Name:               game.Name,
		Status:             game.Status,
		Timestamp:          game.Timestamp,
		Photo:              game.Photo,
		Price:              game.Price,
		Enabled:            game.Enabled,
		GameID:             game.GameID,
		InternalName:       game.InternalName,
		IntegrationPartner: game.IntegrationPartner,
		Provider:           game.Provider,
		CreatedAt:          game.CreatedAt,
		UpdatedAt:          game.UpdatedAt,
		RTPPercent:         rtp,
	}
}
