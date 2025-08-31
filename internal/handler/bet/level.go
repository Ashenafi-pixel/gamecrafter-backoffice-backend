package bet

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"go.uber.org/zap"
)

// CreateLevel  create level.
//
//	@Summary		CreateLevel
//	@Description	CreateLevel allows the creation of a new level
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Level	true	"create level  Request"
//	@Success		200	{object}	dto.Level
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/levels [post]
func (b *bet) CreateLevel(c *gin.Context) {
	var level dto.Level
	if err := c.ShouldBindJSON(&level); err != nil {
		b.log.Error("Failed to bind JSON", zap.Error(err))
		_ = c.Error(err)
		return
	}
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	level.CreatedBy = userIDParsed
	resp, err := b.betModule.CreateLevel(context.Background(), level)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetLevels  get levels.
//
//	@Summary		GetLevels
//	@Description	GetLevels allows the creation of a new level
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.GetRequest	true	"get levels  Request"
//	@Success		200	{object}	dto.GetLevelResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/levels [get]
func (b *bet) GetLevels(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		b.log.Error("Failed to bind query", zap.Error(err))
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetLevels(context.Background(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetLevelByID  get level by id.
//
//	@Summary		GetLevelByID
//	@Description	GetLevelByID allows the retrieval of a level
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Param			req	body	dto.LevelRequirements	true	"create level requirements  Request"
//	@Produce		json
//	@Success		200	{object}	dto.LevelRequirements
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/levels/requirements [post]
func (b *bet) CreateLevelRequirements(c *gin.Context) {
	var req dto.LevelRequirements
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "failed to bind JSON")
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	req.CreatedBy = userIDParsed

	resp, err := b.betModule.CreateLevelReqirements(context.Background(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateLevelRequirement  update level requirement.
//
//	@Summary		UpdateLevelRequirement
//	@Description	UpdateLevelRequirement allows the update of a level requirement
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateLevelRequirementReq	true	"update level requirement  Request"
//	@Success		200	{object}	dto.LevelRequirement
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/levels/requirements [post]
func (b *bet) UpdateLevelRequirement(c *gin.Context) {
	var req dto.UpdateLevelRequirementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		b.log.Error("Failed to bind JSON", zap.Error(err))
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.UpdateLevelRequirement(context.Background(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetUserLevel  get user level.
//
//	@Summary		GetUserLevel
//	@Description	GetUserLevel allows the retrieval of a user level
//	@Tags			User
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.GetUserLevelResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/users/level [get]
func (b *bet) GetUserLevel(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		b.log.Error("User ID is required")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Message: "User ID is required"})
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error("Invalid User ID format", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, "Invalid User ID format")
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetUserLevel(context.Background(), parsedUserID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "Failed to get user level"})
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// AddFakeBalanceLog  add fake balance log.
//
//	@Summary		AddFakeBalanceLog
//	@Description	AddFakeBalanceLog allows the addition of a fake balance log
//	@Tags			Test
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.AddFakeBalanceLogReq	true	"add fake balance log  Request"
//	@Success		200	{object}	dto.FakeBalanceLogResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/test/fake/transaction [post]
func (b *bet) AddFakeBalanceLog(c *gin.Context) {
	var req dto.AddFakeBalanceLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		b.log.Error("Failed to bind JSON", zap.Error(err))
		_ = c.Error(err)
		return
	}

	if req.UserID == uuid.Nil {
		err := errors.ErrInvalidUserInput.Wrap(nil, "invalid user ID")
		b.log.Error("Invalid user ID", zap.Error(err))
		_ = c.Error(err)
		return
	}

	if req.Amount.IsZero() {
		err := errors.ErrInvalidUserInput.Wrap(nil, "change amount cannot be zero")
		b.log.Error("Change amount cannot be zero", zap.Error(err))
		_ = c.Error(err)
		return
	}

	err := b.betModule.AddFakeBalanceLog(context.Background(), req.UserID, req.Amount, "P")
	if err != nil {
		b.log.Error("Failed to add fake balance log", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, dto.FakeBalanceLogResp{
		Message: "Fake balance log added successfully",
	})

}
