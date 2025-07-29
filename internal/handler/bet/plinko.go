package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"go.uber.org/zap"
)

// GetPlinkoGameConfig Get plinko game configs
//	@Summary		GetPlinkoGameConfig
//	@Description	Retrieve plinko game configs
//	@Tags			Bet
//	@Produce		json
//	@Success		200	{object}	dto.PlinkoGameConfig
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/plinko/config [get]
func (b *bet) GetPlinkoGameConfig(c *gin.Context) {
	resp, err := b.betModule.GetPlinkoGameConfig(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PlacePlinkoBe allow user to bet plinko game
//	@Summary		PlacePlinkoBe
//	@Description	PlacePlinkoBe allow players to bet to plinko game
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlacePlinkoGameReq	true	"place plinko game request"
//	@Success		200	{object}	dto.PlacePlinkoGameRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/plinko/drop [post]
func (b *bet) PlacePlinkoBet(c *gin.Context) {
	var req dto.PlacePlinkoGameReq
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed
	resp, err := b.betModule.PlacePlinkoGame(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)

}

// GetUserPlinkoBetHistory allow user to get plinko game history
//	@Summary		GetUserPlinkoBetHistory
//	@Description	GetUserPlinkoBetHistory allow players to get plinko bet history
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	false	"page (int, optional)"
//	@Param			per_page		query	string	false	"per_page (int, optional)"
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlacePlinkoGameReq	true	"place plinko game request"
//	@Success		200	{object}	dto.PlinkoBetHistoryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/plinko/history [GET]
func (b *bet) GetUserPlinkoBetHistory(c *gin.Context) {
	var req dto.PlinkoBetHistoryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	req.UserID = userIDParsed
	resp, err := b.betModule.GetMyPlinkoBetHistory(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)

}

// GetPlinkoGameStats allow user to get plinko game stats
//	@Summary		GetPlinkoGameStats
//	@Description	GetPlinkoGameStats allow players to get plinko bet stats
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlacePlinkoGameReq	true	"get plinko game stats request"
//	@Success		200	{object}	dto.PlinkoGameStatRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/plinko/stats [GET]
func (b *bet) GetPlinkoGameStats(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetPlinkoGameStats(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
