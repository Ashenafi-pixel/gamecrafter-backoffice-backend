package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"go.uber.org/zap"
)

// GetScratchGamePrice  get scratch cards  price and prize.
//
//	@Summary		GetScratchGamePrice
//	@Description	GetScratchGamePrice allow user  to get scratch cards price and prize
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetScratchCardRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/scratchcards/price [get]
func (b *bet) GetScratchGamePrice(c *gin.Context) {
	resp, err := b.betModule.GetScratchGamePrice(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PlaceScratchCardBet  place scratch cards bet.
//
//	@Summary		PlaceScratchCardBet
//	@Description	PlaceScratchCardBet allow player to place scratch cards bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.ScratchCard
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/scratchcards [post]
func (b *bet) PlaceScratchCardBet(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.PlaceScratchCardBet(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetUserScratchCardBetHistories  get scratch cards bets.
//
//	@Summary		GetUserScratchCardBetHistories
//	@Description	GetUserScratchCardBetHistories allow user  to get scratch cards bets.
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetScratchBetHistoriesResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/scratchcards [get]
func (b *bet) GetUserScratchCardBetHistories(c *gin.Context) {
	var req dto.GetRequest
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.GetUserScratchCardBetHistories(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetScratchCardsConfig  get scratch cards config.
//
//	@Summary		GetScratchCardsConfig
//	@Description	GetScratchCardsConfig allow user  to get scratch cards config
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetScratchCardConfigs
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/scratchcard/configs [get]
func (b *bet) GetScratchCardsConfig(c *gin.Context) {
	resp, err := b.betModule.GetScratchCardsConfig(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateScratchGameConfig  update scratch game config.
//
//	@Summary		UpdateScratchGameConfig
//	@Description	UpdateScratchGameConfig allow user  to update scratch game config
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.UpdateScratchGameConfigRequest
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/scratchcard/configs [put]
func (b *bet) UpdateScratchGameConfig(c *gin.Context) {
	var req dto.UpdateScratchGameConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UpdateScratchGameConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
