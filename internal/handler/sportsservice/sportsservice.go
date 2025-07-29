package sportsservice

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

type sportsservice struct {
	module module.SportsService
	logger *zap.Logger
}

func Init(module module.SportsService, logger *zap.Logger) handler.SportsService {
	return &sportsservice{module: module, logger: logger}
}

// SignIn authenticates the sports service
//
//	@Summary		SportsServiceSignIn
//	@Description	Authenticate sports service and get access token
//	@Tags			Sports Service
//	@Accept			json
//	@Produce		json
//	@Param			authReq	body		dto.SportsServiceSignInReq	true	"Sports service authentication request"
//	@Success		200		{object}	dto.SportsServiceSignInRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/sports/signin [post]
func (s *sportsservice) SignIn(c *gin.Context) {
	var req dto.SportsServiceSignInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	token, err := s.module.SignIn(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, token)
}

// PlaceBet places a bet on a sports event
//
//	@Summary		SportsServicePlaceBet
//	@Description	Place a bet on a sports event
//	@Tags			Sports Service
//	@Accept			json
//	@Produce		json
//	@Param			betReq	body		dto.PlaceBetRequest	true	"Sports service place bet request"
//	@Success		200		{object}	dto.PlaceBetResponse
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		403		{object}	response.ErrorResponse
//	@Router			/api/sports/placebet [post]
func (s *sportsservice) PlaceBet(c *gin.Context) {
	var req dto.PlaceBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed

	res, err := s.module.PlaceBet(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AwardWinnings awards winnings to a user
//
//	@Summary		SportsServiceAwardWinnings
//	@Description	Award winnings to a user
//	@Tags			Sports Service
//	@Accept			json
//	@Produce		json
//	@Param			winningsReq	body		dto.SportsServiceAwardWinningsReq	true	"Sports service award winnings request"
//	@Success		200			{object}	dto.SportsServiceAwardWinningsRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		401			{object}	response.ErrorResponse
//	@Failure		403			{object}	response.ErrorResponse
//	@Router			/api/sports/awardwinnings [post]
func (s *sportsservice) AwardWinnings(c *gin.Context) {
	var req dto.SportsServiceAwardWinningsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	req.ExternalUserID = c.GetString("user-id")
	res, err := s.module.AwardWinnings(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}
