package company

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

type company struct {
	log           *zap.Logger
	companyModule module.Company
}

func Init(companyModule module.Company, log *zap.Logger) handler.Company {
	return &company{
		companyModule: companyModule,
		log:           log,
	}
}

// CreateCompany Create Company.
//
//	@Summary		CreateCompany
//	@Description	CreateCompany creates a new company
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			companyReq	body		dto.CreateCompanyReq	true	"Create Company Request"
//	@Success		201			{object}	dto.CreateCompanyRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/companies [post]
func (c *company) CreateCompany(ctx *gin.Context) {
	var req dto.CreateCompanyReq
	userID := ctx.GetString("user-id")

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		c.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.CreatedBy = userIDParsed

	company, err := c.companyModule.CreateCompany(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, company)
}

// GetCompanyByID Get Company by ID.
//
//	@Summary		GetCompanyByID
//	@Description	GetCompanyByID retrieves a company by its ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		200	{object}	dto.Company
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/companies/{id} [get]
func (c *company) GetCompanyByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid company ID format")
		_ = ctx.Error(err)
		return
	}

	company, err := c.companyModule.GetCompanyByID(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, company)
}

// GetCompanies Get Companies.
//
//	@Summary		GetCompanies
//	@Description	GetCompanies retrieves a paginated list of companies
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	true	"Page number (required)"
//	@Param			per-page	query		string	true	"Items per page (required)"
//	@Success		200			{object}	dto.GetCompaniesRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/companies [get]
func (c *company) GetCompanies(ctx *gin.Context) {
	var req dto.GetCompaniesReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(
			err,
			"invalid query parameters",
		)
		_ = ctx.Error(err)
		return
	}

	resp, err := c.companyModule.GetCompanies(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}

// UpdateCompany Update Company.
//
//	@Summary		UpdateCompany
//	@Description	UpdateCompany updates an existing company
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			companyReq	body		dto.UpdateCompanyReq	true	"Update Company Request"
//	@Success		200			{object}	dto.UpdateCompanyRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/companies/{id} [patch]
func (c *company) UpdateCompany(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid company ID format")
		_ = ctx.Error(err)
		return
	}
	var req dto.UpdateCompanyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	req.ID = id

	company, err := c.companyModule.UpdateCompany(ctx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, company)
}

// DeleteCompany Delete Company.
//
//	@Summary		DeleteCompany
//	@Description	DeleteCompany soft deletes a company by ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/admin/companies/{id} [delete]
func (c *company) DeleteCompany(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid company ID format")
		_ = ctx.Error(err)
		return
	}

	err = c.companyModule.DeleteCompany(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusNoContent, nil)
}

// AddIP Add IP Address to Company.
//
//	@Summary		AddIP
//	@Description	AddIP adds an IP address to a company's allowed list
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			addIPReq	body		dto.AddCompanyIPReq	true	"Add IP Request"
//	@Success		200			{object}	dto.UpdateCompanyRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/companies/{id}/add-ip [patch]
func (c *company) AddIP(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "invalid company ID format")
		_ = ctx.Error(err)
		return
	}

	var req dto.AddCompanyIPReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = ctx.Error(err)
		return
	}

	resp, err := c.companyModule.AddIP(ctx, id, req.IpAddr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, resp)
}
