package department

import (
	"fmt"
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

type department struct {
	departmentModule module.Departements
	log              *zap.Logger
}

func Init(departmentModule module.Departements, log *zap.Logger) handler.Departements {
	return &department{
		departmentModule: departmentModule,
		log:              log,
	}
}

// CreateDepartement Create Department.
//
//	@Summary		CreateDepartement
//	@Description	CreateDepartement Create Department
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			deps	body		dto.CreateDepartementReq	true	"Create Department Request"
//	@Success		200		{object}	dto.CreateDepartementRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/departments [post]
func (d *department) CreateDepartement(c *gin.Context) {
	var deps dto.CreateDepartementReq

	if err := c.ShouldBind(&deps); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := d.departmentModule.CreateDepartement(c, deps)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetDepartement Get Department.
//
//	@Summary		GetDepartement
//	@Description	GetDepartement get Department
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	true	"page type (required)"
//	@Param			per-page	query		string	true	"per-page type (required)"
//	@Success		200			{object}	dto.GetDepartementsRes
//	@Failure		400			{object}	response.ErrorResponse
//	@Router			/api/admin/departments [GET]
func (d *department) GetDepartement(c *gin.Context) {
	var depReq dto.GetDepartementsReq
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
	depReq.Page = pageParsed
	depReq.PerPage = perPageParsed
	resp, err := d.departmentModule.GetDepartments(c, depReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateDepartment Update Department.
//
//	@Summary		UpdateDepartment
//	@Description	UpdateDepartment Update Department
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			depReq	body		dto.UpdateDepartment	true	"update Department Request"
//	@Success		200		{object}	dto.UpdateDepartment
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/departments [PATCH]
func (d *department) UpdateDepartment(c *gin.Context) {
	var depReq dto.UpdateDepartment

	if err := c.ShouldBind(&depReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := d.departmentModule.UpdateDepartment(c, depReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// AssignUserToDepartment Update Department.
//
//	@Summary		AssignUserToDepartment
//	@Description	AssignUserToDepartment Update Department
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			depReq	body		dto.AssignDepartmentToUserReq	true	"assign Department Request"
//	@Success		200		{object}	dto.AssignDepartmentToUserResp
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/admin/departments/assign [POST]
func (d *department) AssignUserToDepartment(c *gin.Context) {
	var assignDep dto.AssignDepartmentToUserReq

	if err := c.ShouldBind(&assignDep); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := d.departmentModule.AssignUserToDepartment(c, assignDep)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}
