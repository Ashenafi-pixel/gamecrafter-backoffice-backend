package player

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type player struct {
	log          *zap.Logger
	playerModule module.Player
}

func Init(playerModule module.Player, log *zap.Logger) handler.Player {
	return &player{
		playerModule: playerModule,
		log:          log,
	}
}

// CreatePlayer creates a new player.
//
//	@Summary		CreatePlayer
//	@Description	CreatePlayer creates a new player
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			playerReq	body		dto.CreatePlayerReq	true	"Create Player Request"
//	@Success		201			{object}	dto.CreatePlayerRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/player-management [post]
func (p *player) CreatePlayer(ctx *gin.Context) {
	var req dto.CreatePlayerReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	player, err := p.playerModule.CreatePlayer(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, player)
}

// GetPlayerByID gets a player by ID.
//
//	@Summary		GetPlayerByID
//	@Description	GetPlayerByID retrieves a player by its ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Player ID"
//	@Success		200	{object}	dto.GetPlayerRes
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/player-management/{id} [get]
func (p *player) GetPlayerByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid player ID format")
		_ = ctx.Error(err)
		return
	}

	player, err := p.playerModule.GetPlayerByID(ctx, int32(id))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, dto.GetPlayerRes{Player: player})
}

// GetPlayers gets a list of players.
//
//	@Summary		GetPlayers
//	@Description	GetPlayers retrieves a paginated list of players
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page			query		string	false	"Page number (default: 1)"
//	@Param			per-page		query		string	false	"Items per page (default: 10)"
//	@Param			search			query		string	false	"Search term (searches email and username)"
//	@Param			brand_id		query		string	false	"Filter by brand ID (6-digit numeric)"
//	@Param			country			query		string	false	"Filter by country"
//	@Param			test_account	query		bool	false	"Filter by test account status"
//	@Param			sort_by			query		string	false	"Sort by field (email, username, created_at, updated_at, date_of_birth, country)"
//	@Param			sort_order		query		string	false	"Sort order (asc, desc)"
//	@Success		200				{object}	dto.GetPlayersRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/player-management [get]
func (p *player) GetPlayers(ctx *gin.Context) {
	var req dto.GetPlayersReqs

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(
			err,
			"invalid query parameters",
		)
		_ = ctx.Error(err)
		return
	}

	// Set defaults if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	resp, err := p.playerModule.GetPlayers(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// UpdatePlayer updates a player.
//
//	@Summary		UpdatePlayer
//	@Description	UpdatePlayer updates an existing player
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string				true	"Player ID"
//	@Param			playerReq	body		dto.UpdatePlayerReq	true	"Update Player Request"
//	@Success		200			{object}	dto.UpdatePlayerRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/player-management/{id} [patch]
func (p *player) UpdatePlayer(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid player ID format")
		_ = ctx.Error(err)
		return
	}
	var req dto.UpdatePlayerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.ID = int32(id)

	player, err := p.playerModule.UpdatePlayer(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, player)
}

// DeletePlayer deletes a player.
//
//	@Summary		DeletePlayer
//	@Description	DeletePlayer deletes a player by ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Player ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/player-management/{id} [delete]
func (p *player) DeletePlayer(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid player ID format")
		_ = ctx.Error(err)
		return
	}

	err = p.playerModule.DeletePlayer(ctx, int32(id))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}
