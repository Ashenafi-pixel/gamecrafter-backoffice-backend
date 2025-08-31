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

// CreateStreetKingsGame place streetkings bet.
//
//	@Summary		CreateStreetKingsGame
//	@Description	CreateStreetKingsGame allow player  to place streetkings bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateCrashKingsReq	true	"place streetkings bet  Request"
//	@Success		200	{object}	dto.CreateStreetKingsResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/streetkings/bets [post]
func (b *bet) CreateStreetKingsGame(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.CreateCrashKingsReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.CreateStreetKingsGame(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// CashOutStreetKingsBet cashout streetkings bet.
//
//	@Summary		CashOutStreetKingsBet
//	@Description	CashOutStreetKingsBet allow user  cashout streetkings bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CashOutReq	true	"cashout streetkings bet  Request"
//	@Success		200	{object}	dto.StreetKingsCrashResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/streetkings/bets/cashout [post]
func (b *bet) CashOutStreetKingsBet(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.CashOutReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	req.UserID = userIDParsed
	resp, err := b.betModule.CashOutStreetKings(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetStreetkingHistory  get streetking bets.
//
//	@Summary		GetStreetkingHistory
//	@Description	GetStreetkingHistory allow user  to get streetking bets
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Param			version			query	string	true	"version type (required) v1 for streetking and v2 for streetking 2"
//	@Produce		json
//	@Success		200	{object}	dto.GetStreetkingHistoryRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/streetkings/bets [get]
func (b *bet) GetStreetkingHistory(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	var req dto.GetStreetkingHistoryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetStreetkingHistory(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
