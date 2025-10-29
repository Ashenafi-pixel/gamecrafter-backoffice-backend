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

// CreateRollDaDice place roll da dice bet.
//
//	@Summary		CreateRollDaDice
//	@Description	CreateRollDaDice allow player  to place roll da dice bet
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CreateRollDaDiceReq	true	"place roll da dice bet  Request"
//	@Success		200	{object}	dto.CreateRollDaDiceResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/rolldadices [post]
func (b *bet) CreateRollDaDice(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	var req dto.CreateRollDaDiceReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed
	resp, err := b.betModule.CreateRollDaDice(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetRollDaDiceHistory  get roll da dice bets.
//
//	@Summary		GetRollDaDiceHistory
//	@Description	GetRollDaDiceHistory allow user  to get roll da dice bets
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetRollDaDiceResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/rolldadices [get]
func (b *bet) GetRollDaDiceHistory(c *gin.Context) {
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
	}

	resp, err := b.betModule.GetRollDaDiceHistory(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
