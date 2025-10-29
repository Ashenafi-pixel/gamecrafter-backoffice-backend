package banner

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

type banner struct {
	log          *zap.Logger
	bannerModule module.Banner
}

func Init(bannerModule module.Banner, log *zap.Logger) handler.Banner {
	return &banner{
		bannerModule: bannerModule,
		log:          log,
	}
}

// GetAllBannersPaginated Get All Banners with Pagination.
//
//	@Summary		GetAllBanners
//	@Description	GetAllBanners retrieves banners with pagination support
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		int		false	"Page number (default: 1)"
//	@Param			per_page		query		int		false	"Items per page (default: 10, max: 100)"
//	@Success		200				{object}	dto.GetBannersRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/banners [get]
func (b *banner) GetAllBanners(ctx *gin.Context) {
	var req dto.GetBannersReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid query parameters")
		_ = ctx.Error(err)
		return
	}

	banners, err := b.bannerModule.GetAllBanners(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, banners)
}

// GetBannerByPage Get Banner by Page.
//
//	@Summary		GetBannerByPage
//	@Description	GetBannerByPage retrieves a banner for a specific page
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		string	true	"Page name (required)"
//	@Success		200				{object}	dto.Banner
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Router			/api/admin/banners/display [get]
func (b *banner) GetBannerByPage(ctx *gin.Context) {
	var req dto.GetBannerReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid query parameters")
		_ = ctx.Error(err)
		return
	}

	banner, err := b.bannerModule.GetBannerByPage(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, banner)
}

// UpdateBanner Update Banner.
//
//	@Summary		UpdateBanner
//	@Description	UpdateBanner updates an existing banner
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token>"
//	@Param			id				path		string				true	"Banner ID"
//	@Param			banner			body		dto.UpdateBannerReq	true	"Update Banner Request"
//	@Success		200				{object}	dto.Banner
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Router			/api/admin/banners/{id} [patch]
func (b *banner) UpdateBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid banner ID format")
		_ = ctx.Error(err)
		return
	}

	var req dto.UpdateBannerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.ID = id

	banner, err := b.bannerModule.UpdateBanner(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, banner)
}

// CreateBanner Create Banner.
//
//	@Summary		CreateBanner
//	@Description	CreateBanner creates a new banner
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token>"
//	@Param			banner			body		dto.CreateBannerReq	true	"Create Banner Request"
//	@Success		201				{object}	dto.Banner
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		409				{object}	response.ErrorResponse
//	@Router			/api/admin/banners [post]
func (b *banner) CreateBanner(ctx *gin.Context) {
	var req dto.CreateBannerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	banner, err := b.bannerModule.CreateBanner(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, banner)
}

// DeleteBanner Delete a banner by ID.
//
//	@Summary		DeleteBanner
//	@Description	Delete a banner by its ID.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			id				path		string	true	"Banner ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Router			/api/admin/banners/{id} [DELETE]
func (b *banner) DeleteBanner(c *gin.Context) {
	bannerID := c.Param("id")
	bannerIDParsed, err := uuid.Parse(bannerID)
	if err != nil {
		b.log.Error("invalid banner ID", zap.Error(err), zap.String("bannerID", bannerID))
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid banner ID")
		_ = c.Error(err)
		return
	}

	err = b.bannerModule.DeleteBanner(c, bannerIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UploadBannerImage Upload a banner image.
//
//	@Summary		UploadBannerImage
//	@Description	Upload a banner image file to S3 storage.
//	@Tags			Admin
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			image			formData	file	true	"Banner image file (max size 8MB)"
//	@Success		200				{object}	dto.UploadBannerImageResp
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		413				{object}	response.ErrorResponse
//	@Router			/api/admin/banners/upload [POST]
func (b *banner) UploadBannerImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		b.log.Error("Failed to retrieve file", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to retrieve file")
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

	resp, err := b.bannerModule.UploadBannerImage(c, file, header)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
