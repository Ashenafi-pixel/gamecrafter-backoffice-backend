package brand

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
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format")
		_ = ctx.Error(err)
		return
	}

	brand, err := b.brandModule.GetBrandByID(ctx, id)
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
	id, err := uuid.Parse(idStr)
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

	req.ID = id

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
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid brand ID format")
		_ = ctx.Error(err)
		return
	}

	err = b.brandModule.DeleteBrand(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}

