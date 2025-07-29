package bet

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type bet struct {
	betModule module.Bet
	log       *zap.Logger
}

func Init(betModule module.Bet, log *zap.Logger) handler.Bet {
	return &bet{
		betModule: betModule,
		log:       log,
	}
}

// GetOpenRound get open round for bet.
//	@Summary		GetOpenRound
//	@Description	Get allow users to get open round
//	@Tags			bet
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.OpenRoundRes
//	@Router			/api/game/round [get]
func (b *bet) GetOpenRound(c *gin.Context) {
	betRound, err := b.betModule.GetOpenRound(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, betRound)
}

// PlaceBet Place bet for user.
//	@Summary		PlaceBet
//	@Description	PlaceBet allow user to bet for open round
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			placeBetReq	body		dto.PlaceBetReq	true	"place bet  Request"
//	@Success		200			{object}	dto.PlaceBetRes
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/game/place-bet [post]
func (b *bet) PlaceBet(c *gin.Context) {
	userID := c.GetString("user-id")
	var placeBetReq dto.PlaceBetReq
	if err := c.ShouldBind(&placeBetReq); err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	placeBetReq.UserID = userIDParsed
	placedBet, err := b.betModule.PlaceBet(c, placeBetReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, placedBet)
}

// CashOut cashout bet for user.
//	@Summary		CashOut
//	@Description	CashOut allow user to cashout in progress bets
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			cashoutReq	body		dto.CashOutReq	true	"cashout  Request"
//	@Success		200			{object}	dto.CashOutRes
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/game/cash-out [post]
func (b *bet) CashOut(c *gin.Context) {
	userID := c.GetString("user-id")
	var cashoutReq dto.CashOutReq
	if err := c.ShouldBind(&cashoutReq); err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	cashoutReq.UserID = userIDParsed
	_, err = b.betModule.CashOut(c, cashoutReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, nil)
}

// GetBetHistory Get Bet History.
//	@Summary		GetBetHistory
//	@Description	Retrieve user bets based on user_id (opetional).
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			user-id			query		string	false	"User ID (UUID, optional)"
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per-page		query		string	true	"per-page type (required)"
//	@Success		200				{object}	dto.BetHistoryResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/game/history [get]
func (b *bet) GetBetHistory(c *gin.Context) {
	betHistory := dto.GetBetHistoryReq{}
	userID := c.Query("user-id")
	page := c.Query("page")
	perpage := c.Query("per-page")
	if perpage == "" || page == "" {
		err := fmt.Errorf("page and per_page query required")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	pageParsed, err := strconv.Atoi(page)
	if err != nil {
		err := fmt.Errorf("unable to convert page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	perPageParsed, err := strconv.Atoi(perpage)
	if err != nil {
		err := fmt.Errorf("unable to convert per_page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if userID != "" {
		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			b.log.Error(err.Error(), zap.Any("userID", userID))
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
		betHistory.UserID = userIDParsed
	}
	betHistory.Page = pageParsed
	betHistory.PerPage = perPageParsed
	betHistoryRes, err := b.betModule.GetBetHistory(c, betHistory)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, betHistoryRes)
}

// CashOut CancelBet bet for user.
//	@Summary		CancelBet
//	@Description	CancelBet allow user to cancel bet which is not started
//	@Tags			bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			cancelBetReq	body		dto.CancelBetReq	true	"cancel  Request"
//	@Success		200				{object}	dto.CancelBetResp
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/game/cancel [post]
func (b *bet) CancelBet(c *gin.Context) {
	var cancelBetReq dto.CancelBetReq
	if err := c.ShouldBind(&cancelBetReq); err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error())
		err = errors.ErrUnExpectedError.Wrap(err, err.Error())
		_ = c.Error(err)
	}
	cancelBetReq.UserID = userIDParsed
	cancelRes, err := b.betModule.CancelBet(c, cancelBetReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, cancelRes)
}

// GetLeaders Get Bet Leaders.
//	@Summary		GetLeaders
//	@Description	Retrieve bet leaders .
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Tags			bet
//	@Produce		json
//	@Success		200	{object}	dto.LeadersResp
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/api/game/leaders [get]
func (b *bet) GetLeaders(c *gin.Context) {
	leaders, err := b.betModule.GetLeaders(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, leaders)
}

// GetMyBetHistory Get User Bet History.
//	@Summary		GetMyBetHistory
//	@Description	Retrieve users bet history.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per-page		query		string	true	"per-page type (required)"
//	@Success		200				{object}	dto.BetHistoryResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/user/game/history [get]
func (b *bet) GetMyBetHistory(c *gin.Context) {
	userID := c.GetString("user-id")
	var cashoutReq dto.CashOutReq
	if err := c.ShouldBind(&cashoutReq); err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	betHistory := dto.GetBetHistoryReq{}
	page := c.Query("page")
	perpage := c.Query("per-page")
	if perpage == "" || page == "" {
		err := fmt.Errorf("page and per_page query required")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	pageParsed, err := strconv.Atoi(page)
	if err != nil {
		err := fmt.Errorf("unable to convert page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	perPageParsed, err := strconv.Atoi(perpage)
	if err != nil {
		err := fmt.Errorf("unable to convert per_page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	betHistory.UserID = userIDParsed
	betHistory.Page = pageParsed
	betHistory.PerPage = perPageParsed
	betHistoryRes, err := b.betModule.GetBetHistory(c, betHistory)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, betHistoryRes)
}

// GetAllFailedRounds Get failed rounds.
//	@Summary		GetAllFailedRounds
//	@Description	Retrieve failed rounds.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page (required)"
//	@Param			per_page		query		string	true	"per-page  (required)"
//	@Param			status			query		string	true	"status (required ,(which are refund status) failed or completed)"
//	@Success		200				{object}	dto.GetFailedRoundsRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/failed/rounds [get]
func (b *bet) GetAllFailedRounds(c *gin.Context) {
	userID := c.GetString("user-id")
	var cashoutReq dto.CashOutReq
	if err := c.ShouldBind(&cashoutReq); err != nil {
		b.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	_, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	var req dto.GetFailedRoundsReq

	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if req.PerPage == 0 || req.Page == 0 {
		err := fmt.Errorf("invalid page and per_page query can not be empty")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetFailedRounds(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// ManualRefundFailedRound manual refund failed  bet for user.
//	@Summary		ManualRefundFailedRound
//	@Description	ManualRefundFailedRound allow admin  to refund failed rounds if automatic refund is not work
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.ManualRefundFailedRoundsReq	true	"manual refund of failed rounds  Request"
//	@Success		200	{object}	dto.ManualRefundFailedRoundsRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/failed/rounds [post]
func (b *bet) ManualRefundFailedRound(c *gin.Context) {
	var req dto.ManualRefundFailedRoundsReq
	if err := c.ShouldBind(&req); err != nil {
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

	req.AdminID = userIDParsed
	resp, err := b.betModule.ManualRefundFailedRounds(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateBetIcon update bet icon
//	@Summary		UpdateProfilePicture
//	@Description	Allows a user to upload and update their profile picture.
//	@Tags			Admin
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			picture			formData	file	true	"Profile picture file (max size 8MB)"
//	@Success		200				{object}	string	"Profile picture URL"
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		413				{object}	response.ErrorResponse
//	@Router			/api/admin/bets/icons [POST]
func (b *bet) UpdateBetIcon(c *gin.Context) {
	file, header, err := c.Request.FormFile("picture")
	if err != nil {
		b.log.Error("Failed to retrieve file", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	defer file.Close()

	const maxFileSize = 8 * 1024 * 1024
	if header.Size > maxFileSize {
		err := errors.ErrInvalidUserInput.New("File size exceeds the 8 MB limit")
		b.log.Warn("File too large", zap.Int64("fileSize", header.Size))
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UploadBetIcons(c, file, header)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
