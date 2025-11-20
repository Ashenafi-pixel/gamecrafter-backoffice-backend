package page

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	customErrors "github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

type pageHandler struct {
	pageModule module.Page
	log        *zap.Logger
}

func Init(pageModule module.Page, log *zap.Logger) PageHandler {
	return &pageHandler{
		pageModule: pageModule,
		log:        log,
	}
}

type PageHandler interface {
	GetAllPages(c *gin.Context)
	GetUserAllowedPages(c *gin.Context)
	UpdateUserAllowedPages(c *gin.Context)
}

// GetAllPages gets all available pages
//
//	@Summary		GetAllPages
//	@Description	Get all available pages in the system
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"pages"
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/pages [get]
func (p *pageHandler) GetAllPages(c *gin.Context) {
	pages, err := p.pageModule.GetAllPages(c)
	if err != nil {
		p.log.Error("failed to get all pages", zap.Error(err))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"pages": pages,
	})
}

// GetUserAllowedPages gets allowed pages for a specific user
//
//	@Summary		GetUserAllowedPages
//	@Description	Get all allowed pages for a specific user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200		{object}	map[string]interface{}	"pages"
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/users/{id}/pages [get]
func (p *pageHandler) GetUserAllowedPages(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, "invalid user ID")
		_ = c.Error(err)
		return
	}

	pages, err := p.pageModule.GetUserAllowedPages(c, userID)
	if err != nil {
		p.log.Error("failed to get user allowed pages", zap.Error(err), zap.String("user_id", userID.String()))
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"pages": pages,
	})
}

// UpdateUserAllowedPages updates allowed pages for a specific user
//
//	@Summary		UpdateUserAllowedPages
//	@Description	Update allowed pages for a specific user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"User ID"
//	@Param			request	body		dto.AssignPagesToUserReq	true	"Update pages request"
//	@Success		200		{object}	response.SuccessResponse
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/users/{id}/pages [put]
func (p *pageHandler) UpdateUserAllowedPages(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, "invalid user ID")
		_ = c.Error(err)
		return
	}

	var req dto.AssignPagesToUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Ensure userID in request matches path parameter
	req.UserID = userID

	// Replace all pages for the user (removes old ones and adds new ones)
	err = p.pageModule.ReplaceUserPages(c, userID, req.PageIDs)
	if err != nil {
		p.log.Error("failed to update user allowed pages", zap.Error(err), zap.String("user_id", userID.String()))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, map[string]string{
		"message": "User pages updated successfully",
	})
}

