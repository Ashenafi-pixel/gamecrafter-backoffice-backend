package adds

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type adds struct {
	module module.Adds
	logger *zap.Logger
}

func Init(module module.Adds, logger *zap.Logger) handler.Adds {
	return &adds{module: module, logger: logger}
}

// SignIn authenticates the adds service
//
//	@Summary		AddsServiceSignIn
//	@Description	Authenticate adds service and get access token
//	@Tags			Adds Service
//	@Accept			json
//	@Produce		json
//	@Param			authReq	body		dto.AddSignInReq	true	"Adds service authentication request"
//	@Success		200		{object}	dto.AddSignInRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/adds/signin [post]
func (a *adds) SignIn(c *gin.Context) {
	var req dto.AddSignInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	token, err := a.module.SignIn(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, token)
}

// UpdateBalance updates user balance from adds service
//
//	@Summary		AddsServiceUpdateBalance
//	@Description	Update user balance from adds service
//	@Tags			Adds Service
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Bearer <token>"
//	@Param			balanceReq		body		dto.AddUpdateBalanceReq	true	"Balance update request"
//	@Success		200				{object}	dto.AddUpdateBalanceRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/adds/balance/update [post]
func (a *adds) UpdateBalance(c *gin.Context) {
	var req dto.AddUpdateBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}

	resp, err := a.module.UpdateBalance(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// SaveAddsService creates a new adds service (admin only)
//
//	@Summary		SaveAddsService
//	@Description	Create a new adds service configuration
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer <token>"
//	@Param			serviceReq		body		dto.CreateAddsServiceReq	true	"Create adds service request"
//	@Success		201				{object}	dto.CreateAddsServiceRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/adds/services [post]
func (a *adds) SaveAddsService(c *gin.Context) {
	var req dto.CreateAddsServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("error binding request", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding request")
		_ = c.Error(err)
		return
	}
	service, err := a.module.SaveAddsService(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, service)
}

// GetAddsServices retrieves all adds services (admin only)
//
//	@Summary		GetAddsServices
//	@Description	Get all adds services with pagination
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		int		false	"Page number"
//	@Param			per_page		query		int		false	"Items per page"
//	@Success		200				{object}	dto.GetAddsServicesRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/adds/services [get]
func (a *adds) GetAddsServices(c *gin.Context) {
	var req dto.GetAddServicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		a.logger.Error("error binding query parameters", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "error binding query parameters")
		_ = c.Error(err)
		return
	}

	// Set default values if not provided
	if req.Page < 0 {
		req.Page = 1
	}
	if req.PerPage < 0 {
		req.PerPage = 10
	}

	services, err := a.module.GetAddsServices(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, services)
}
