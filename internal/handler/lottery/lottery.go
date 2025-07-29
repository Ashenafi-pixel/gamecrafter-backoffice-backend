package lottery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type lottery struct {
	log           *zap.Logger
	lotteryModule module.Lottery
}

func Init(lotteryModule module.Lottery, log *zap.Logger) handler.Lottery {
	return &lottery{
		log:           log,
		lotteryModule: lotteryModule,
	}
}

// CreateLotteryService handles the creation of a lottery service
//
//	@Summary		Create Lottery Service
//	@Description	Create a new lottery service with the provided details
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer token for authentication"
//	@Param			request			body		dto.CreateLotteryServiceReq	true	"Create Lottery Service Request"
//	@Success		201				{object}	dto.CreateLotteryServiceRes
//	@Failure		400				{object}	response.ErrorResponse	"Invalid input"
//	@Failure		500				{object}	response.ErrorResponse	"Internal server error"
//	@Router			/admin/lottery/service [post]
func (l *lottery) CreateLotteryService(c *gin.Context) {
	var req dto.CreateLotteryServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		l.log.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	resp, err := l.lotteryModule.CreateLotteryService(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	resp.Message = "Lottery service created successfully"
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// CreateLotteryRequest handles the creation of a lottery request
//
//	@Summary		Create Lottery Request
//	@Description	Create a new lottery request with the provided details
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer token for authentication"
//	@Param			request			body		dto.LotteryRequestCreate	true	"Create Lottery Request"
//	@Success		201				{object}	dto.LotteryRequestCreate
//	@Failure		400				{object}	response.ErrorResponse	"Invalid input"
//	@Failure		500				{object}	response.ErrorResponse	"Internal server error"
//	@Router			/admin/lottery/request [post]
func (l *lottery) CreateLotteryRequest(c *gin.Context) {
	var req dto.LotteryRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		l.log.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	resp, err := l.lotteryModule.CreateLotteryRequest(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// CheckUserBalanceAndDeductBalance handles the verification and deduction of user balance for lottery
//
//	@Summary		Verify and Deduct User Balance for Lottery
//	@Description	Verify user balance and deduct the amount for lottery participation
//	@Tags			Lottery
//	@Accept			json
//	@Produce		json
//	@Param			x-user-token	header		string									true	"Bearer token for authentication"
//	@Param			request			body		dto.LotteryVerifyAndDeductBalanceReq	true	"Verify and Deduct User Balance Request"
//	@Success		200				{object}	dto.LotteryVerifyAndDeductBalanceRes
//	@Failure		400				{object}	response.ErrorResponse	"Invalid input"
//	@Failure		500				{object}	response.ErrorResponse	"Internal server error"
//	@Router			/lottery/verify/deduct/balance [post]
func (l *lottery) CheckUserBalanceAndDeductBalance(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		l.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.LotteryVerifyAndDeductBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		l.log.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	req.UserID = userIDParsed
	resp, err := l.lotteryModule.CheckUserBalanceAndDeductBalance(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
