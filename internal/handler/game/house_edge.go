package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
)

type HouseEdgeHandler struct {
	houseEdgeService HouseEdgeService
}

type HouseEdgeService interface {
	CreateHouseEdge(ctx *gin.Context, req dto.GameHouseEdgeRequest) (*dto.GameHouseEdgeResponse, error)
	GetHouseEdgeByID(ctx *gin.Context, id uuid.UUID) (*dto.GameHouseEdgeResponse, error)
	GetHouseEdges(ctx *gin.Context, params dto.GameHouseEdgeQueryParams) (*dto.GameHouseEdgeListResponse, error)
	UpdateHouseEdge(ctx *gin.Context, id uuid.UUID, req dto.GameHouseEdgeRequest) (*dto.GameHouseEdgeResponse, error)
	DeleteHouseEdge(ctx *gin.Context, id uuid.UUID) error
	BulkUpdateHouseEdgeStatus(ctx *gin.Context, req dto.BulkUpdateHouseEdgeRequest) error
	GetHouseEdgeStats(ctx *gin.Context) (*dto.HouseEdgeStats, error)
	GetHouseEdgesByGameType(ctx *gin.Context, gameType string) ([]dto.GameHouseEdgeResponse, error)
	GetHouseEdgesByGameVariant(ctx *gin.Context, gameType, gameVariant string) ([]dto.GameHouseEdgeResponse, error)
}

func NewHouseEdgeHandler(houseEdgeService HouseEdgeService) *HouseEdgeHandler {
	return &HouseEdgeHandler{
		houseEdgeService: houseEdgeService,
	}
}

// CreateHouseEdge creates a new house edge configuration
//
//	@Summary		Create House Edge
//	@Description	Create a new house edge configuration for a game
//	@Tags			House Edge Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string						true	"Bearer <token>"
//	@Param			request			body	dto.GameHouseEdgeRequest	true	"House edge creation request"
//	@Success		201				{object}	dto.GameHouseEdgeResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges [post]
func (h *HouseEdgeHandler) CreateHouseEdge(c *gin.Context) {
	var req dto.GameHouseEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := h.houseEdgeService.CreateHouseEdge(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetHouseEdgeByID retrieves a house edge configuration by ID
//
//	@Summary		Get House Edge by ID
//	@Description	Retrieve a specific house edge configuration by its ID
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			id				path	string	true	"House Edge ID"
//	@Success		200				{object}	dto.GameHouseEdgeResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/{id} [get]
func (h *HouseEdgeHandler) GetHouseEdgeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid house edge ID format")
		_ = c.Error(err)
		return
	}

	resp, err := h.houseEdgeService.GetHouseEdgeByID(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetHouseEdges retrieves a list of house edge configurations with filtering and pagination
//
//	@Summary		Get House Edges
//	@Description	Retrieve a paginated list of house edge configurations with optional filtering
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			page			query	int		false	"Page number (default: 1)"
//	@Param			per_page		query	int		false	"Items per page (default: 10, max: 100)"
//	@Param			search			query	string	false	"Search by game_id or game_name"
//	@Param			game_id			query	string	false	"Filter by specific game ID"
//	@Param			game_type		query	string	false	"Filter by game type"
//	@Param			game_variant	query	string	false	"Filter by game variant"
//	@Param			is_active		query	bool	false	"Filter by active status"
//	@Param			sort_by			query	string	false	"Sort by field (game_type, game_variant, house_edge, created_at, updated_at)"
//	@Param			sort_order		query	string	false	"Sort order (asc, desc)"
//	@Success		200				{object}	dto.GameHouseEdgeListResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges [get]
func (h *HouseEdgeHandler) GetHouseEdges(c *gin.Context) {
	var params dto.GameHouseEdgeQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 {
		params.PerPage = 10
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	resp, err := h.houseEdgeService.GetHouseEdges(c, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateHouseEdge updates an existing house edge configuration
//
//	@Summary		Update House Edge
//	@Description	Update an existing house edge configuration
//	@Tags			House Edge Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string						true	"Bearer <token>"
//	@Param			id				path	string						true	"House Edge ID"
//	@Param			request			body	dto.GameHouseEdgeRequest	true	"House edge update request"
//	@Success		200				{object}	dto.GameHouseEdgeResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/{id} [put]
func (h *HouseEdgeHandler) UpdateHouseEdge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid house edge ID format")
		_ = c.Error(err)
		return
	}

	var req dto.GameHouseEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := h.houseEdgeService.UpdateHouseEdge(c, id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteHouseEdge deletes a house edge configuration
//
//	@Summary		Delete House Edge
//	@Description	Delete a house edge configuration by ID
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			id				path	string	true	"House Edge ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/{id} [delete]
func (h *HouseEdgeHandler) DeleteHouseEdge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid house edge ID format")
		_ = c.Error(err)
		return
	}

	err = h.houseEdgeService.DeleteHouseEdge(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkUpdateHouseEdgeStatus updates the status of multiple house edge configurations
//
//	@Summary		Bulk Update House Edge Status
//	@Description	Update the active status of multiple house edge configurations at once
//	@Tags			House Edge Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string							true	"Bearer <token>"
//	@Param			request			body	dto.BulkUpdateHouseEdgeRequest	true	"Bulk update request"
//	@Success		200				{object}	response.SuccessResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/bulk-status [put]
func (h *HouseEdgeHandler) BulkUpdateHouseEdgeStatus(c *gin.Context) {
	var req dto.BulkUpdateHouseEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err := h.houseEdgeService.BulkUpdateHouseEdgeStatus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"message": "House edge statuses updated successfully",
		"count":   len(req.HouseEdgeIDs),
	})
}

// GetHouseEdgeStats retrieves house edge statistics
//
//	@Summary		Get House Edge Statistics
//	@Description	Retrieve statistics about house edge configurations in the system
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Success		200				{object}	dto.HouseEdgeStats
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/stats [get]
func (h *HouseEdgeHandler) GetHouseEdgeStats(c *gin.Context) {
	resp, err := h.houseEdgeService.GetHouseEdgeStats(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetHouseEdgesByGameType retrieves house edge configurations by game type
//
//	@Summary		Get House Edges by Game Type
//	@Description	Retrieve all house edge configurations for a specific game type
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			game_type		path	string	true	"Game Type"
//	@Success		200				{object}	[]dto.GameHouseEdgeResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/by-game-type/{game_type} [get]
func (h *HouseEdgeHandler) GetHouseEdgesByGameType(c *gin.Context) {
	gameType := c.Param("game_type")
	if gameType == "" {
		err := errors.ErrInvalidUserInput.New("game_type parameter is required")
		_ = c.Error(err)
		return
	}

	resp, err := h.houseEdgeService.GetHouseEdgesByGameType(c, gameType)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetHouseEdgesByGameVariant retrieves house edge configurations by game type and variant
//
//	@Summary		Get House Edges by Game Variant
//	@Description	Retrieve all house edge configurations for a specific game type and variant
//	@Tags			House Edge Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			game_type		path	string	true	"Game Type"
//	@Param			game_variant	path	string	true	"Game Variant"
//	@Success		200				{object}	[]dto.GameHouseEdgeResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/house-edges/by-game-variant/{game_type}/{game_variant} [get]
func (h *HouseEdgeHandler) GetHouseEdgesByGameVariant(c *gin.Context) {
	gameType := c.Param("game_type")
	gameVariant := c.Param("game_variant")

	if gameType == "" {
		err := errors.ErrInvalidUserInput.New("game_type parameter is required")
		_ = c.Error(err)
		return
	}
	if gameVariant == "" {
		err := errors.ErrInvalidUserInput.New("game_variant parameter is required")
		_ = c.Error(err)
		return
	}

	resp, err := h.houseEdgeService.GetHouseEdgesByGameVariant(c, gameType, gameVariant)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
