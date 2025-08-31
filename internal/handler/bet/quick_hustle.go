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

// PlaceQuickHustleBet place quick hustle bet.
//
//	@Summary		PlaceQuickHustleBet
//	@Description	PlaceQuickHustleBet allow player  to place quick hustle bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateQuickHustleBetReq	true	"place quick hustle bet  Request"
//	@Success		200	{object}	dto.CreateQuickHustelBetRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/quickhustles [post]
func (b *bet) PlaceQuickHustleBet(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.CreateQuickHustleBetReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed

	resp, err := b.betModule.PlaceQuickHustleBet(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// UserSelectCard send user guess for quick hustle.
//
//	@Summary		UserSelectCard
//	@Description	UserSelectCard allow player  to send user guess for quick hustle.
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.SelectQuickHustlePossibilityReq	true	"guess for quick hustle. req"
//	@Success		200	{object}	dto.CloseQuickHustleResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/quickhustles/selects [post]
func (b *bet) UserSelectCard(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.SelectQuickHustlePossibilityReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed

	resp, err := b.betModule.UserSelectCard(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetQuickHustleBetHistory  get quick hustle bet history.
//
//	@Summary		GetQuickHustleBetHistory
//	@Description	GetQuickHustleBetHistory allow user  to get quick hustle bet history.
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetQuickHustleResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/quickhustle/bets [get]
func (b *bet) GetQuickHustleBetHistory(c *gin.Context) {
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

	resp, err := b.betModule.GetQuickHustleBetHistory(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
