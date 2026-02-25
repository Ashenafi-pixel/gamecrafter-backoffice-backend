package brand

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type brand struct {
	log        *zap.Logger
	brandModule module.Brand
}

func Init(brandModule module.Brand, log *zap.Logger) handler.Brand {
	return &brand{
		brandModule: brandModule,
		log:         log,
	}
}

// CreateBrand Create Brand.
//
//	@Summary		CreateBrand
//	@Description	CreateBrand creates a new brand
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			brandReq	body		dto.CreateBrandReq	true	"Create Brand Request"
//	@Success		201			{object}	dto.CreateBrandRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/brands [post]
func (b *brand) CreateBrand(ctx *gin.Context) {
	var req dto.CreateBrandReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	brand, err := b.brandModule.CreateBrand(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, brand)
}

// GetBrandByID Get Brand by ID.
//
//	@Summary		GetBrandByID
//	@Description	GetBrandByID retrieves a brand by its ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Brand ID"
//	@Success		200	{object}	dto.Brand
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/brands/{id} [get]
func (b *brand) GetBrandByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format")
		_ = ctx.Error(err)
		return
	}

	brand, err := b.brandModule.GetBrandByID(ctx, int32(id))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, brand)
}

// GetBrands Get Brands.
//
//	@Summary		GetBrands
//	@Description	GetBrands retrieves a paginated list of brands
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	true	"Page number (required)"
//	@Param			per-page	query		string	true	"Items per page (required)"
//	@Param			search		query		string	false	"Search term"
//	@Param			is_active	query		bool	false	"Filter by active status"
//	@Success		200			{object}	dto.GetBrandsRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/brands [get]
func (b *brand) GetBrands(ctx *gin.Context) {
	var req dto.GetBrandsReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(
			err,
			"invalid query parameters",
		)
		_ = ctx.Error(err)
		return
	}

	resp, err := b.brandModule.GetBrands(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// UpdateBrand Update Brand.
//
//	@Summary		UpdateBrand
//	@Description	UpdateBrand updates an existing brand
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Brand ID"
//	@Param			brandReq	body		dto.UpdateBrandReq	true	"Update Brand Request"
//	@Success		200			{object}	dto.UpdateBrandRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/brands/{id} [patch]
func (b *brand) UpdateBrand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format")
		_ = ctx.Error(err)
		return
	}
	var req dto.UpdateBrandReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.ID = int32(id)

	brand, err := b.brandModule.UpdateBrand(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, brand)
}

// DeleteBrand Delete Brand.
//
//	@Summary		DeleteBrand
//	@Description	DeleteBrand deletes a brand by ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Brand ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/brands/{id} [delete]
func (b *brand) DeleteBrand(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format")
		_ = ctx.Error(err)
		return
	}

	err = b.brandModule.DeleteBrand(ctx, int32(id))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}

func parseBrandID(c *gin.Context) (int32, bool) {
	idStr := c.Param("brandId")
	if idStr == "" {
		idStr = c.Param("id")
	}
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format"))
		return 0, false
	}
	return int32(id), true
}

// ChangeBrandStatus updates only is_active for a brand.
func (b *brand) ChangeBrandStatus(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	var req dto.ChangeBrandStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
		return
	}
	if err := b.brandModule.ChangeBrandStatus(ctx, brandID, req); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{"is_active": req.IsActive})
}

// CreateBrandCredential creates API credentials for a brand; returns client_secret once.
func (b *brand) CreateBrandCredential(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	var req dto.CreateBrandCredentialReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
		return
	}
	cred, secret, err := b.brandModule.CreateBrandCredential(ctx, brandID, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	cred.ClientSecret = secret
	response.SendSuccessResponse(ctx, http.StatusCreated, cred)
}

// RotateBrandCredential rotates the secret; returns new client_secret once.
func (b *brand) RotateBrandCredential(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	credIDStr := ctx.Param("credentialId")
	credID, err := strconv.ParseInt(credIDStr, 10, 32)
	if err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid credential ID format"))
		return
	}
	res, err := b.brandModule.RotateBrandCredential(ctx, brandID, int32(credID))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, res)
}

// GetBrandCredentialByID returns a credential (without secret).
func (b *brand) GetBrandCredentialByID(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	credIDStr := ctx.Param("credentialId")
	credID, err := strconv.ParseInt(credIDStr, 10, 32)
	if err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid credential ID format"))
		return
	}
	cred, found, err := b.brandModule.GetBrandCredentialByID(ctx, brandID, int32(credID))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if !found {
		_ = ctx.Error(errors.ErrResourceNotFound.New("credential not found"))
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, cred)
}

// AddBrandAllowedOrigin adds an allowed origin for the brand.
func (b *brand) AddBrandAllowedOrigin(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	var req dto.AddBrandAllowedOriginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
		return
	}
	res, err := b.brandModule.AddBrandAllowedOrigin(ctx, brandID, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusCreated, res)
}

// RemoveBrandAllowedOrigin removes an allowed origin by ID.
func (b *brand) RemoveBrandAllowedOrigin(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	originIDStr := ctx.Param("originId")
	originID, err := strconv.ParseInt(originIDStr, 10, 32)
	if err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, "invalid origin ID format"))
		return
	}
	if err := b.brandModule.RemoveBrandAllowedOrigin(ctx, brandID, int32(originID)); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}

// ListBrandAllowedOrigins returns all allowed origins for the brand.
func (b *brand) ListBrandAllowedOrigins(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	res, err := b.brandModule.ListBrandAllowedOrigins(ctx, brandID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, res)
}

// GetBrandFeatureFlags returns feature flags for the brand.
func (b *brand) GetBrandFeatureFlags(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	res, err := b.brandModule.GetBrandFeatureFlags(ctx, brandID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, res)
}

// UpdateBrandFeatureFlags updates feature flags for the brand.
func (b *brand) UpdateBrandFeatureFlags(ctx *gin.Context) {
	brandID, ok := parseBrandID(ctx)
	if !ok {
		return
	}
	var req dto.UpdateBrandFeatureFlagsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(errors.ErrInvalidUserInput.Wrap(err, err.Error()))
		return
	}
	if err := b.brandModule.UpdateBrandFeatureFlags(ctx, brandID, req); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{"flags": req.Flags})
}

