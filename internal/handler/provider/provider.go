package provider

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type provider struct {
	log            *zap.Logger
	providerModule module.Provider
}

func Init(providerModule module.Provider, log *zap.Logger) handler.Provider {
	return &provider{
		providerModule: providerModule,
		log:            log,
	}
}

// CreateProvider Create Provider.
//
//	@Summary		CreateProvider
//	@Description	CreateProvider creates a new game provider
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateProviderRequest	true	"Create Provider Request"
//	@Success		201		{object}	dto.GameProvider
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/admin/providers [post]
func (p *provider) CreateProvider(ctx *gin.Context) {
	var req dto.CreateProviderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	provider, err := p.providerModule.CreateProvider(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, provider)
}
func (p *provider) GetAllProviders(ctx *gin.Context) {
	providers, err := p.providerModule.GetAllProviders(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, providers)
}
func (p *provider) UpdateProvider(ctx *gin.Context) {
	var req dto.UpdateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}
	id := ctx.Param("id")
	parsedID, err := uuid.Parse(id)
	req.ID = parsedID
	provider, err := p.providerModule.UpdateProvider(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, provider)
}

func (p *provider) DeleteProvider(ctx *gin.Context) {
	id := ctx.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid provider ID")
		_ = ctx.Error(err)
		return
	}
	err = p.providerModule.DeleteProvider(ctx, parsedID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{"message": "provider deleted successfully"})
}
