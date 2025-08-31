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

// SetCrytoKingsConfig update crypto kings game config
//
//	@Summary		SetCrytoKingsConfig
//	@Description	SetCrytoKingsConfig allow admin to update crypto kings config
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateCryptoKingsConfigReq	true	"update crypto kings config  Request"
//	@Success		200	{object}	dto.UpdateCryptokingsConfigRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/cryptokings/configs [post]
func (b *bet) SetCrytoKingsConfig(c *gin.Context) {
	var req dto.UpdateCryptoKingsConfigReq

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.SetCrytoKingsConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PlaceCryptoKingsBet place crypto kings bet.
//
//	@Summary		PlaceCryptoKingsBet
//	@Description	PlaceCryptoKingsBet allow player  to place crypto kings bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.PlaceCryptoKingsBetReq	true	"place crypto kings bet  Request"
//	@Success		200	{object}	dto.PlaceCryptoKingsBetRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/cryptokings [post]
func (b *bet) PlaceCryptoKingsBet(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.PlaceCryptoKingsBetReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.PlaceCryptoKingsBet(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetCryptoKingsBetHistory  get crypto kings bets.
//
//	@Summary		GetCryptoKingsBetHistory
//	@Description	GetCryptoKingsBetHistory allow user  to get crypto kings bets
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetCryptoKingsUserBetHistoryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/streetkings [get]
func (b *bet) GetCryptoKingsBetHistory(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetCryptoKingsBetHistory(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetCryptoKingsCurrentCryptoPrice  get currenc crypto king's crypto price.
//
//	@Summary		GetCryptoKingsCurrentCryptoPrice
//	@Description	GetCryptoKingsCurrentCryptoPrice allow user  to get currenc crypto king's crypto price
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetCryptoCurrencyPriceResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/cryptokings/price [get]
func (b *bet) GetCryptoKingsCurrentCryptoPrice(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetCryptoKingsCurrentCryptoPrice(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
