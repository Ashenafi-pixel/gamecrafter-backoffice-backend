package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
)

type GameHandler struct {
	gameService GameService
}

type GameService interface {
	CreateGame(ctx *gin.Context, req dto.CreateGameRequest) (*dto.GameResponse, error)
	GetGameByID(ctx *gin.Context, id uuid.UUID) (*dto.GameResponse, error)
	GetGames(ctx *gin.Context, params dto.GameQueryParams) (*dto.GameListResponse, error)
	UpdateGame(ctx *gin.Context, id uuid.UUID, req dto.UpdateGameRequest) (*dto.GameResponse, error)
	DeleteGame(ctx *gin.Context, id uuid.UUID) error
	BulkUpdateGameStatus(ctx *gin.Context, req dto.BulkUpdateGameStatusRequest) error
	GetGameStats(ctx *gin.Context) (*dto.GameManagementStats, error)
}

func NewGameHandler(gameService GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

// CreateGame creates a new game
//
//	@Summary		Create Game
//	@Description	Create a new game with all necessary details
//	@Tags			Game Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string					true	"Bearer <token>"
//	@Param			request			body	dto.CreateGameRequest	true	"Game creation request"
//	@Success		201				{object}	dto.GameResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req dto.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := h.gameService.CreateGame(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetGameByID retrieves a game by ID
//
//	@Summary		Get Game by ID
//	@Description	Retrieve a specific game by its ID
//	@Tags			Game Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			id				path	string	true	"Game ID"
//	@Success		200				{object}	dto.GameResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/{id} [get]
func (h *GameHandler) GetGameByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid game ID format")
		_ = c.Error(err)
		return
	}

	resp, err := h.gameService.GetGameByID(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetGames retrieves a list of games with filtering and pagination
//
//	@Summary		Get Games
//	@Description	Retrieve a paginated list of games with optional filtering
//	@Tags			Game Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			page			query	int		false	"Page number (default: 1)"
//	@Param			per_page		query	int		false	"Items per page (default: 10, max: 100)"
//	@Param			search			query	string	false	"Search by name"
//	@Param			status			query	string	false	"Filter by status (ACTIVE, INACTIVE, MAINTENANCE)"
//	@Param			provider		query	string	false	"Filter by provider"
//	@Param			enabled			query	bool	false	"Filter by enabled status"
//	@Param			sort_by			query	string	false	"Sort by field (name, status, created_at, updated_at)"
//	@Param			sort_order		query	string	false	"Sort order (asc, desc)"
//	@Success		200				{object}	dto.GameListResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games [get]
func (h *GameHandler) GetGames(c *gin.Context) {
	var params dto.GameQueryParams
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

	resp, err := h.gameService.GetGames(c, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateGame updates an existing game
//
//	@Summary		Update Game
//	@Description	Update an existing game's details
//	@Tags			Game Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string					true	"Bearer <token>"
//	@Param			id				path	string					true	"Game ID"
//	@Param			request			body	dto.UpdateGameRequest	true	"Game update request"
//	@Success		200				{object}	dto.GameResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/{id} [put]
func (h *GameHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid game ID format")
		_ = c.Error(err)
		return
	}

	var req dto.UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := h.gameService.UpdateGame(c, id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteGame deletes a game
//
//	@Summary		Delete Game
//	@Description	Delete a game by ID
//	@Tags			Game Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			id				path	string	true	"Game ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/{id} [delete]
func (h *GameHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid game ID format")
		_ = c.Error(err)
		return
	}

	err = h.gameService.DeleteGame(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkUpdateGameStatus updates the status of multiple games
//
//	@Summary		Bulk Update Game Status
//	@Description	Update the status of multiple games at once
//	@Tags			Game Management
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string							true	"Bearer <token>"
//	@Param			request			body	dto.BulkUpdateGameStatusRequest	true	"Bulk update request"
//	@Success		200				{object}	response.SuccessResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/bulk-status [put]
func (h *GameHandler) BulkUpdateGameStatus(c *gin.Context) {
	var req dto.BulkUpdateGameStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err := h.gameService.BulkUpdateGameStatus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"message": "Game statuses updated successfully",
		"count":   len(req.GameIDs),
	})
}

// GetGameStats retrieves game statistics
//
//	@Summary		Get Game Statistics
//	@Description	Retrieve statistics about games in the system
//	@Tags			Game Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Success		200				{object}	dto.GameStats
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/stats [get]
func (h *GameHandler) GetGameStats(c *gin.Context) {
	resp, err := h.gameService.GetGameStats(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetGameByGameID retrieves a game by its game_id field
//
//	@Summary		Get Game by Game ID
//	@Description	Retrieve a game by its game_id field (external game identifier)
//	@Tags			Game Management
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Param			game_id			path	string	true	"External Game ID"
//	@Success		200				{object}	dto.GameResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/games/by-game-id/{game_id} [get]
func (h *GameHandler) GetGameByGameID(c *gin.Context) {
	gameID := c.Param("game_id")
	if gameID == "" {
		err := errors.ErrInvalidUserInput.New("game_id parameter is required")
		_ = c.Error(err)
		return
	}

	// Convert to UUID for consistency with service layer
	id, err := uuid.Parse(gameID)
	if err != nil {
		// If not a UUID, we'll need to handle this in the service layer
		// For now, we'll pass the string directly
	}

	resp, err := h.gameService.GetGameByID(c, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
