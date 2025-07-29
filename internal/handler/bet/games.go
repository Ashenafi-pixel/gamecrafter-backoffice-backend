package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
)

// UpdateGame  allow admin to update games.
//	@Summary		UpdateGame
//	@Description	UpdateGame allow admin to update games.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Game	true	"get football round price  Request"
//	@Success		200	{object}	dto.Game
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games [PUT]
func (b *bet) UpdateGame(c *gin.Context) {
	var req dto.Game
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UpdateGame(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetGames  get games
//	@Summary		GetGames
//	@Description	GetGames allow user  to get games active games
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetGamesResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games [get]
func (b *bet) GetGames(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetGames(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)

}

// DisableAllGames  allow admin to disable all games from RGS.
//	@Summary		DisableAllGames
//	@Description	DisableAllGames allow admin to disable all games.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.UpdateFootballBetPriceRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games/disable [post]
func (b *bet) DisableAllGames(c *gin.Context) {
	resp, err := b.betModule.DisableAllGames(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetAvailableGames  get available games which are not active.
//	@Summary		GetAvailableGames
//	@Description	GetAvailableGames allow user to get available games which are not active.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	[]dto.Game
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games/available [get]
func (b *bet) GetAvailableGames(c *gin.Context) {
	resp, err := b.betModule.ListInactiveGames(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteGame allow admin to delete game
//	@Summary		DeleteGame
//	@Description	DeleteGame allow admin to delete game
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Game	true	"delete game  Request"
//	@Success		200	{object}	dto.DeleteResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games [DELETE]
func (b *bet) DeleteGame(c *gin.Context) {
	var req dto.Game
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.DeleteGame(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// AddGame allow admin to add game
//	@Summary		AddGame
//	@Description	AddGame allow admin to add game
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Game	true	"add game  Request"
//	@Success		200	{object}	dto.Game
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games [POST]
func (b *bet) AddGame(c *gin.Context) {
	var game dto.Game
	if err := c.ShouldBind(&game); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.AddGame(c, game)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateGameStatus allow admin to update game status
//	@Summary		UpdateGameStatus
//	@Description	UpdateGameStatus allow admin to update game status
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Game	true	"update game status  Request"
//	@Success		200	{object}	dto.Game
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/games/status [PUT]
func (b *bet) UpdateGameStatus(c *gin.Context) {
	var req dto.Game
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.UpdateGameStatus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
