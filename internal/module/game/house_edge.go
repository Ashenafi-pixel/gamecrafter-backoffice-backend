package game

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/storage/game"
	"go.uber.org/zap"
)

type HouseEdgeService struct {
	houseEdgeStorage game.HouseEdgeStorage
	logger           *zap.Logger
}

func NewHouseEdgeService(houseEdgeStorage game.HouseEdgeStorage, logger *zap.Logger) *HouseEdgeService {
	return &HouseEdgeService{
		houseEdgeStorage: houseEdgeStorage,
		logger:           logger,
	}
}

func (s *HouseEdgeService) CreateHouseEdge(ctx *gin.Context, req dto.GameHouseEdgeRequest) (*dto.GameHouseEdgeResponse, error) {
	s.logger.Info("Creating new house edge",
		zap.String("game_type", req.GameType),
		zap.String("game_variant", getStringValue(req.GameVariant)),
		zap.String("house_edge", req.HouseEdge.String()))

	// Set effective_from to now if not provided
	effectiveFrom := time.Now()
	if req.EffectiveFrom != nil {
		effectiveFrom = *req.EffectiveFrom
	}

	// Convert DTO to storage model
	houseEdgeModel := &game.GameHouseEdge{
		GameType:       req.GameType,
		GameVariant:    req.GameVariant,
		HouseEdge:      req.HouseEdge,
		MinBet:         req.MinBet,
		MaxBet:         req.MaxBet,
		IsActive:       req.IsActive,
		EffectiveFrom:  effectiveFrom,
		EffectiveUntil: req.EffectiveUntil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	createdHouseEdge, err := s.houseEdgeStorage.CreateHouseEdge(ctx, houseEdgeModel)
	if err != nil {
		s.logger.Error("Failed to create house edge", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create house edge")
	}

	// Convert storage model to response DTO
	response := s.convertToHouseEdgeResponse(createdHouseEdge)

	s.logger.Info("House edge created successfully",
		zap.String("house_edge_id", createdHouseEdge.ID.String()),
		zap.String("game_type", createdHouseEdge.GameType))

	return response, nil
}

func (s *HouseEdgeService) GetHouseEdgeByID(ctx *gin.Context, id uuid.UUID) (*dto.GameHouseEdgeResponse, error) {
	s.logger.Info("Getting house edge by ID", zap.String("id", id.String()))

	houseEdge, err := s.houseEdgeStorage.GetHouseEdgeByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get house edge by ID", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edge")
	}

	if houseEdge == nil {
		return nil, errors.ErrResourceNotFound.New("house edge not found")
	}

	response := s.convertToHouseEdgeResponse(houseEdge)
	return response, nil
}

func (s *HouseEdgeService) GetHouseEdges(ctx *gin.Context, params dto.GameHouseEdgeQueryParams) (*dto.GameHouseEdgeListResponse, error) {
	s.logger.Info("Getting house edges with filters",
		zap.Int("page", params.Page),
		zap.Int("per_page", params.PerPage),
		zap.String("game_type", params.GameType),
		zap.String("game_variant", params.GameVariant))

	houseEdges, totalCount, err := s.houseEdgeStorage.GetHouseEdges(ctx, game.HouseEdgeQueryParams{
		Page:        params.Page,
		PerPage:     params.PerPage,
		GameType:    params.GameType,
		GameVariant: params.GameVariant,
		IsActive:    params.IsActive,
		SortBy:      params.SortBy,
		SortOrder:   params.SortOrder,
	})
	if err != nil {
		s.logger.Error("Failed to get house edges", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edges")
	}

	// Convert to response DTOs
	houseEdgeResponses := make([]dto.GameHouseEdgeResponse, len(houseEdges))
	for i, houseEdge := range houseEdges {
		houseEdgeResponses[i] = *s.convertToHouseEdgeResponse(&houseEdge)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.PerPage)))

	response := &dto.GameHouseEdgeListResponse{
		HouseEdges: houseEdgeResponses,
		TotalCount: totalCount,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	s.logger.Info("House edges retrieved successfully",
		zap.Int("count", len(houseEdges)),
		zap.Int("total_count", totalCount))

	return response, nil
}

func (s *HouseEdgeService) UpdateHouseEdge(ctx *gin.Context, id uuid.UUID, req dto.GameHouseEdgeRequest) (*dto.GameHouseEdgeResponse, error) {
	s.logger.Info("Updating house edge", zap.String("id", id.String()))

	// Check if house edge exists
	existingHouseEdge, err := s.houseEdgeStorage.GetHouseEdgeByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get house edge for update", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edge")
	}

	if existingHouseEdge == nil {
		return nil, errors.ErrResourceNotFound.New("house edge not found")
	}

	// Update fields
	existingHouseEdge.GameType = req.GameType
	existingHouseEdge.GameVariant = req.GameVariant
	existingHouseEdge.HouseEdge = req.HouseEdge
	existingHouseEdge.MinBet = req.MinBet
	existingHouseEdge.MaxBet = req.MaxBet
	existingHouseEdge.IsActive = req.IsActive

	if req.EffectiveFrom != nil {
		existingHouseEdge.EffectiveFrom = *req.EffectiveFrom
	}
	if req.EffectiveUntil != nil {
		existingHouseEdge.EffectiveUntil = req.EffectiveUntil
	}

	existingHouseEdge.UpdatedAt = time.Now()

	updatedHouseEdge, err := s.houseEdgeStorage.UpdateHouseEdge(ctx, existingHouseEdge)
	if err != nil {
		s.logger.Error("Failed to update house edge", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update house edge")
	}

	response := s.convertToHouseEdgeResponse(updatedHouseEdge)

	s.logger.Info("House edge updated successfully",
		zap.String("house_edge_id", updatedHouseEdge.ID.String()),
		zap.String("game_type", updatedHouseEdge.GameType))

	return response, nil
}

func (s *HouseEdgeService) DeleteHouseEdge(ctx *gin.Context, id uuid.UUID) error {
	s.logger.Info("Deleting house edge", zap.String("id", id.String()))

	// Check if house edge exists
	existingHouseEdge, err := s.houseEdgeStorage.GetHouseEdgeByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get house edge for deletion", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to get house edge")
	}

	if existingHouseEdge == nil {
		return errors.ErrResourceNotFound.New("house edge not found")
	}

	err = s.houseEdgeStorage.DeleteHouseEdge(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete house edge", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to delete house edge")
	}

	s.logger.Info("House edge deleted successfully",
		zap.String("house_edge_id", id.String()),
		zap.String("game_type", existingHouseEdge.GameType))

	return nil
}

func (s *HouseEdgeService) BulkUpdateHouseEdgeStatus(ctx *gin.Context, req dto.BulkUpdateHouseEdgeRequest) error {
	s.logger.Info("Bulk updating house edge status",
		zap.Bool("is_active", req.IsActive),
		zap.Int("count", len(req.HouseEdgeIDs)))

	err := s.houseEdgeStorage.BulkUpdateHouseEdgeStatus(ctx, req.HouseEdgeIDs, req.IsActive)
	if err != nil {
		s.logger.Error("Failed to bulk update house edge status", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to bulk update house edge status")
	}

	s.logger.Info("House edge statuses updated successfully",
		zap.Bool("is_active", req.IsActive),
		zap.Int("count", len(req.HouseEdgeIDs)))

	return nil
}

func (s *HouseEdgeService) GetHouseEdgeStats(ctx *gin.Context) (*dto.HouseEdgeStats, error) {
	s.logger.Info("Getting house edge statistics")

	stats, err := s.houseEdgeStorage.GetHouseEdgeStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get house edge statistics", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edge statistics")
	}

	response := &dto.HouseEdgeStats{
		TotalHouseEdges:    stats.TotalHouseEdges,
		ActiveHouseEdges:   stats.ActiveHouseEdges,
		InactiveHouseEdges: stats.InactiveHouseEdges,
		UniqueGameTypes:    stats.UniqueGameTypes,
		UniqueGameVariants: stats.UniqueGameVariants,
	}

	s.logger.Info("House edge statistics retrieved successfully",
		zap.Int("total_house_edges", stats.TotalHouseEdges),
		zap.Int("active_house_edges", stats.ActiveHouseEdges))

	return response, nil
}

func (s *HouseEdgeService) GetHouseEdgesByGameType(ctx *gin.Context, gameType string) ([]dto.GameHouseEdgeResponse, error) {
	s.logger.Info("Getting house edges by game type", zap.String("game_type", gameType))

	houseEdges, err := s.houseEdgeStorage.GetHouseEdgesByGameType(ctx, gameType)
	if err != nil {
		s.logger.Error("Failed to get house edges by game type", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edges by game type")
	}

	// Convert to response DTOs
	responses := make([]dto.GameHouseEdgeResponse, len(houseEdges))
	for i, houseEdge := range houseEdges {
		responses[i] = *s.convertToHouseEdgeResponse(&houseEdge)
	}

	s.logger.Info("House edges by game type retrieved successfully",
		zap.String("game_type", gameType),
		zap.Int("count", len(houseEdges)))

	return responses, nil
}

func (s *HouseEdgeService) GetHouseEdgesByGameVariant(ctx *gin.Context, gameType, gameVariant string) ([]dto.GameHouseEdgeResponse, error) {
	s.logger.Info("Getting house edges by game variant",
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant))

	houseEdges, err := s.houseEdgeStorage.GetHouseEdgesByGameVariant(ctx, gameType, gameVariant)
	if err != nil {
		s.logger.Error("Failed to get house edges by game variant", zap.Error(err))
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get house edges by game variant")
	}

	// Convert to response DTOs
	responses := make([]dto.GameHouseEdgeResponse, len(houseEdges))
	for i, houseEdge := range houseEdges {
		responses[i] = *s.convertToHouseEdgeResponse(&houseEdge)
	}

	s.logger.Info("House edges by game variant retrieved successfully",
		zap.String("game_type", gameType),
		zap.String("game_variant", gameVariant),
		zap.Int("count", len(houseEdges)))

	return responses, nil
}

func (s *HouseEdgeService) convertToHouseEdgeResponse(houseEdge *game.GameHouseEdge) *dto.GameHouseEdgeResponse {
	// Calculate house edge percentage
	houseEdgePercent := houseEdge.HouseEdge.Mul(decimal.NewFromInt(100)).StringFixed(2) + "%"

	return &dto.GameHouseEdgeResponse{
		ID:               houseEdge.ID,
		GameType:         houseEdge.GameType,
		GameVariant:      houseEdge.GameVariant,
		HouseEdge:        houseEdge.HouseEdge,
		HouseEdgePercent: houseEdgePercent,
		MinBet:           houseEdge.MinBet,
		MaxBet:           houseEdge.MaxBet,
		IsActive:         houseEdge.IsActive,
		EffectiveFrom:    houseEdge.EffectiveFrom,
		EffectiveUntil:   houseEdge.EffectiveUntil,
		CreatedAt:        houseEdge.CreatedAt,
		UpdatedAt:        houseEdge.UpdatedAt,
	}
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
