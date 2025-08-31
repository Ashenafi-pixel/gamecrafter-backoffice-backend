package agent

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

type agent struct {
	log         *zap.Logger
	agentModule module.Agent
}

func Init(agentModule module.Agent, log *zap.Logger) handler.Agent {
	return &agent{
		agentModule: agentModule,
		log:         log,
	}
}

// CreateAgentReferralLink handles the creation of agent referral links.
//
//	@Summary		Create Agent Referral Link
//	@Description	Create a new agent referral link
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateAgentReferralLinkReq	true	"Create Agent Referral Link Request"
//	@Success		200		{object}	dto.CreateAgentReferralLinkRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Router			/api/agent/links [post]
func (a *agent) CreateAgentReferralLink(c *gin.Context) {
	var req dto.CreateAgentReferralLinkReq

	// Extract client_id and secret from headers
	clientID := c.GetHeader("X-Provider-ID")
	providerSecret := c.GetHeader("X-Provider-Secret")
	if clientID == "" || providerSecret == "" {
		err := errors.ErrInvalidUserInput.New("Provider client_id and secret are required in headers")
		_ = c.Error(err)
		return
	}

	provider, err := a.agentModule.ValidateAgentProviderCredentials(c, clientID, providerSecret)
	if err != nil {
		a.log.Error("invalid provider credentials", zap.Error(err))
		_ = c.Error(err)
		return
	}
	if provider.Status != "active" {
		err := errors.ErrAcessError.New("provider is not active")
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.log.Error("invalid request body", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Set ProviderID from validated provider record
	req.ProviderID = provider.ID

	if req.RequestID == "" {
		err := errors.ErrInvalidUserInput.New("Request ID is required")
		_ = c.Error(err)
		return
	}

	res, err := a.agentModule.CreateAgentReferralLink(c, req)
	if err != nil {
		a.log.Error("unable to create agent referral link", zap.Error(err), zap.Any("request", req))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetAgentReferrals handles getting agent referrals with pagination.
//
//	@Summary		Get Agent Referrals
//	@Description	Get agent referrals with pagination
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer token for authentication"
//	@Param			request_id		query		string	true	"Request ID"
//	@Param			page			query		int		false	"Page number"
//	@Param			limit			query		int		false	"Page size"
//	@Success		200				{object}	dto.GetAgentReferralsRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/agent/referrals [get]
func (a *agent) GetAgentReferrals(c *gin.Context) {
	var req dto.GetAgentReferralsReq

	if err := c.ShouldBindQuery(&req); err != nil {
		a.log.Error("invalid query parameters", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	res, err := a.agentModule.GetAgentReferralsByRequestID(c, req)
	if err != nil {
		a.log.Error("unable to get agent referrals", zap.Error(err), zap.String("requestID", req.RequestID))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetReferralStats handles getting referral statistics.
//
//	@Summary		Get Agent Referral Statistics
//	@Description	Get statistics for agent referrals
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer token for authentication"
//	@Param			request_id		query		string	true	"Request ID"
//	@Success		200				{object}	dto.GetReferralStatsRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Router			/api/admin/agent/stats [get]
func (a *agent) GetReferralStats(c *gin.Context) {
	var req dto.GetReferralStatsReq

	// Bind query parameters to DTO
	if err := c.ShouldBindQuery(&req); err != nil {
		a.log.Error("invalid query parameters", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Call module layer
	res, err := a.agentModule.GetReferralStatsByRequestID(c, req)
	if err != nil {
		a.log.Error("unable to get referral stats", zap.Error(err), zap.String("requestID", req.RequestID))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// CreateAgentProvider handles the creation of agent providers (admin only)
//
//	@Summary		Create Agent Provider
//	@Description	Create a new agent provider (admin only)
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer token for authentication"
//	@Param			request			body		dto.CreateAgentProviderReq	true	"Create Agent Provider Request"
//	@Success		201				{object}	dto.CreateAgentProviderRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		403				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/agent/providers [post]
func (a *agent) CreateAgentProvider(c *gin.Context) {
	var req dto.CreateAgentProviderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		a.log.Error("invalid request body", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	res, err := a.agentModule.CreateAgentProvider(c, req)
	if err != nil {
		a.log.Error("failed to create agent provider", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, res)
}
