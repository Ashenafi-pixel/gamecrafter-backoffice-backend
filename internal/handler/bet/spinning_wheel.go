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

// GetSpinningWheelPrice  get spinning wheel  price .
//
//	@Summary		GetSpinningWheelPrice
//	@Description	GetSpinningWheelPrice allow user  to get spinning wheel price
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetSpinningWheelPrice
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/spinningwheels/price [get]
func (b *bet) GetSpinningWheelPrice(c *gin.Context) {
	resp, err := b.betModule.GetSpinningWheelPrice(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PlaceSpinningWheelBet  place spinning wheel bet.
//
//	@Summary		PlaceSpinningWheelBet
//	@Description	PlaceSpinningWheelBet allow player to place spinning wheel bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.PlaceSpinningWheelResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/spinningwheels [post]
func (b *bet) PlaceSpinningWheelBet(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.PlaceSpinningWheelBet(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSpinningWheelUserBetHistory  get spinning wheel bets.
//
//	@Summary		GetSpinningWheelUserBetHistory
//	@Description	GetSpinningWheelUserBetHistory allow user  to get spinning wheel bets.
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetSpinningWheelHistoryResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/spinningwheels [get]
func (b *bet) GetSpinningWheelUserBetHistory(c *gin.Context) {
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
	resp, err := b.betModule.GetSpinningWheelUserBetHistory(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateSpinningWheelMysteries  create spinning wheel mysteries.
//
//	@Summary		CreateSpinningWheelMysteries
//	@Description	CreateSpinningWheelMysteries allow admin  to create spinning wheel mysteries.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateSpinningWheelMysteryReq	true	"create spinning wheel mysteries  Request"
//	@Success		200	{object}	dto.CreateSpinningWheelMysteryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/mysteries [post]
func (b *bet) CreateSpinningWheelMysteries(c *gin.Context) {
	var req dto.CreateSpinningWheelMysteryReq
	if err := c.ShouldBindJSON(&req); err != nil {
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
	req.CreatedBy = userIDParsed
	resp, err := b.betModule.CreateSpinningWheelMystery(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSpinningWheelMysteries  get spinning wheel mysteries.
//
//	@Summary		GetSpinningWheelMysteries
//	@Description	GetSpinningWheelMysteries allow user  to get spinning wheel mysteries.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetSpinningWheelMysteryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/mysteries [get]
func (b *bet) GetSpinningWheelMysteries(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetSpinningWheelMystery(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateSpinningWheelMystery  update spinning wheel mystery.
//
//	@Summary		UpdateSpinningWheelMystery
//	@Description	UpdateSpinningWheelMystery allow admin  to update spinning wheel mystery.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateSpinningWheelMysteryReq	true	"update spinning wheel mystery  Request"
//	@Success		200	{object}	dto.UpdateSpinningWheelMysteryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/mysteries [put]
func (b *bet) UpdateSpinningWheelMystery(c *gin.Context) {
	var req dto.UpdateSpinningWheelMysteryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UpdateSpinningWheelMystery(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteSpinningWheelMystery  delete spinning wheel mystery.
//
//	@Summary		DeleteSpinningWheelMystery
//	@Description	DeleteSpinningWheelMystery allow admin  to delete spinning wheel mystery.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.DeleteReq	true	"delete spinning wheel mystery  Request"
//	@Success		200	{object}	nil
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/mysteries [delete]
func (b *bet) DeleteSpinningWheelMystery(c *gin.Context) {
	var req dto.DeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	err := b.betModule.DeleteSpinningWheelMystery(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, nil)
}

// CreateSpinningWheelConfig  create spinning wheel config.
//
//	@Summary		CreateSpinningWheelConfig
//	@Description	CreateSpinningWheelConfig allow admin  to create spinning wheel config.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateSpinningWheelConfigReq	true	"create spinning wheel config  Request"
//	@Success		200	{object}	dto.CreateSpinningWheelConfigRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/config [post]
func (b *bet) CreateSpinningWheelConfig(c *gin.Context) {
	var req dto.CreateSpinningWheelConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
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
	req.CreatedBy = userIDParsed
	resp, err := b.betModule.CreateSpinningWheelConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSpinningWheelConfig  get spinning wheel config.
//
//	@Summary		GetSpinningWheelConfig
//	@Description	GetSpinningWheelConfig allow user  to get spinning wheel config
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetSpinningWheelConfigResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/spinningwheels/configs [get]
func (b *bet) GetSpinningWheelConfigs(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetSpinningWheelConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateSpinningWheelConfig  update spinning wheel config.
//
//	@Summary		UpdateSpinningWheelConfig
//	@Description	UpdateSpinningWheelConfig allow admin  to update spinning wheel config.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateSpinningWheelConfigReq	true	"update spinning wheel config  Request"
//	@Success		200	{object}	dto.UpdateSpinningWheelConfigRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/config [put]
func (b *bet) UpdateSpinningWheelConfig(c *gin.Context) {
	var req dto.UpdateSpinningWheelConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.UpdateSpinningWheelConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteSpinningWheelConfig  delete spinning wheel config.
//
//	@Summary		DeleteSpinningWheelConfig
//	@Description	DeleteSpinningWheelConfig allow admin  to delete spinning wheel config.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.DeleteReq	true	"delete spinning wheel config  Request"
//	@Success		200	{object}	nil
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/spinningwheels/config [delete]
func (b *bet) DeleteSpinningWheelConfig(c *gin.Context) {
	var req dto.DeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	err := b.betModule.DeleteSpinningWheelConfig(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, nil)
}
