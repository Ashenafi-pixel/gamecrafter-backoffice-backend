package authz

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type authz struct {
	log         *zap.Logger
	authzModule module.Authz
}

func Init(log *zap.Logger, authzModule module.Authz) handler.Authz {
	return &authz{
		log:         log,
		authzModule: authzModule,
	}
}

// GetPermissions Get list of permissions.
//	@Summary		GetPermissions
//	@Description	GetPermissions Retrieve list of permissions.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per-page		query		string	true	"per-page type (required)"
//	@Success		200				{object}	[]dto.Permissions
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/permissions [get]
func (a *authz) GetPermissions(c *gin.Context) {
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
	resp, err := a.authzModule.GetPermissions(c, dto.GetPermissionReq{Page: pageParsed, PerPage: perPageParsed})
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)

}

// CreateRole allow user to create role.
//	@Summary		CreateRole
//	@Description	CreateRole allow user to create role
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			createRoleReq	body		dto.CreateRoleReq	true	"create role   Request"
//	@Success		200				{object}	dto.Role
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/roles [post]
func (a *authz) CreateRole(c *gin.Context) {
	var createRoleReq dto.CreateRoleReq
	if err := c.ShouldBind(&createRoleReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.authzModule.CreateRole(c, createRoleReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetRoles Get list of roles.
//	@Summary		GetRoles
//	@Description	GetRoles Retrieve list of roles.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per-page		query		string	true	"per-page type (required)"
//	@Success		200				{object}	[]dto.Role
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/roles [get]
func (a *authz) GetRoles(c *gin.Context) {
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

	res, err := a.authzModule.GetRoles(c, dto.GetRoleReq{Page: pageParsed, PerPage: perPageParsed})
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// UpdateRolePermissions allow user to update role permissions.
//	@Summary		UpdateRolePermissions
//	@Description	UpdateRolePermissions allow user to update role permissions
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			rolePermissions	body		dto.UpdatePermissionToRoleReq	true	"update role permissions   Request"
//	@Success		200				{object}	dto.UpdatePermissionToRoleRes
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/roles [PATCH]
func (a *authz) UpdateRolePermissions(c *gin.Context) {
	var rolePermissions dto.UpdatePermissionToRoleReq
	if err := c.ShouldBind(&rolePermissions); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.authzModule.UpdatePermissionsOfRole(c, rolePermissions)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// RemoveRole allow user to remove role .
//	@Summary		RemoveRole
//	@Description	RemoveRole allow user to remove role
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			removeReq	body	dto.RemoveRoleReq	true	"remove role Request"
//	@Success		200
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/roles [DELETE]
func (a *authz) RemoveRole(c *gin.Context) {
	var removeReq dto.RemoveRoleReq
	if err := c.ShouldBind(&removeReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	err := a.authzModule.RemoveRole(c, removeReq.RoleID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, nil)
}

// AssignRoleToUser allow user to assign role to the user.
//	@Summary		AssignRoleToUser
//	@Description	AssignRoleToUser allow assign role to user
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			assignRoleReq	body		dto.AssignRoleToUserReq	true	"assign role to user Request"
//	@Success		200				{object}	dto.UpdatePermissionToRoleRes
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users/roles [POST]
func (a *authz) AssignRoleToUser(c *gin.Context) {
	var assignRoleReq dto.AssignRoleToUserReq
	if err := c.ShouldBind(&assignRoleReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := a.authzModule.AssignRoleToUser(c, assignRoleReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RevokeUserRole allow user to revoke role from the user.
//	@Summary		RevokeUserRole
//	@Description	RevokeUserRole allow revoke role from user
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			userRoleReq	body	dto.UserRole	true	"revoke role from user Request"
//	@Success		200
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/users/roles [POST]
func (a *authz) RevokeUserRole(c *gin.Context) {
	var userRoleReq dto.UserRole
	if err := c.ShouldBind(&userRoleReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	err := a.authzModule.RevokeUserRole(c, userRoleReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, nil)
}

// GetRoleUsers Get list of users of role.
//	@Summary		GetRoleUsers
//	@Description	GetRoleUsers Retrieve list of users of role.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	[]dto.Role
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/roles/:id/users [get]
func (a *authz) GetRoleUsers(c *gin.Context) {
	roleID := c.Param("id")
	roleIDParse, err := uuid.Parse(roleID)
	if err != nil {
		err := fmt.Errorf("unable to convert role id to uuid")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.authzModule.GetRoleUsers(c, roleIDParse)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetRoleUsers Get list of user roles.
//	@Summary		GetRoleUsers
//	@Description	GetRoleUsers Retrieve list of user roles.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.UserRolesRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users/:id/roles [get]
func (a *authz) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		err := fmt.Errorf("unable to convert user id to uuid")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := a.authzModule.GetUserRoles(c, userIDParsed)

	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
