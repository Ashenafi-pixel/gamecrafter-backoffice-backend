package alert

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/alert"
	"go.uber.org/zap"
)

type AlertEmailGroupHandler interface {
	CreateEmailGroup(c *gin.Context)
	GetEmailGroup(c *gin.Context)
	GetAllEmailGroups(c *gin.Context)
	UpdateEmailGroup(c *gin.Context)
	DeleteEmailGroup(c *gin.Context)
}

type alertEmailGroupHandler struct {
	emailGroupStorage alert.AlertEmailGroupStorage
	log               *zap.Logger
}

func NewAlertEmailGroupHandler(emailGroupStorage alert.AlertEmailGroupStorage, log *zap.Logger) AlertEmailGroupHandler {
	return &alertEmailGroupHandler{
		emailGroupStorage: emailGroupStorage,
		log:               log,
	}
}

// CreateEmailGroup creates a new email group
func (h *alertEmailGroupHandler) CreateEmailGroup(c *gin.Context) {
	var req dto.CreateAlertEmailGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get admin user ID from context
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid admin user ID",
		})
		return
	}

	group, err := h.emailGroupStorage.CreateEmailGroup(c.Request.Context(), &req, adminUUID)
	if err != nil {
		h.log.Error("Failed to create email group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Failed to create email group",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertEmailGroupResponse{
		Success: true,
		Message: "Email group created successfully",
		Data:    group,
	})
}

// GetEmailGroup gets an email group by ID
func (h *alertEmailGroupHandler) GetEmailGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid group ID",
			Error:   err.Error(),
		})
		return
	}

	group, err := h.emailGroupStorage.GetEmailGroupByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get email group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Failed to get email group",
			Error:   err.Error(),
		})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Email group not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertEmailGroupResponse{
		Success: true,
		Message: "Email group retrieved successfully",
		Data:    group,
	})
}

// GetAllEmailGroups gets all email groups
func (h *alertEmailGroupHandler) GetAllEmailGroups(c *gin.Context) {
	groups, err := h.emailGroupStorage.GetAllEmailGroups(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get email groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertEmailGroupsResponse{
			Success: false,
			Message: "Failed to get email groups",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertEmailGroupsResponse{
		Success:    true,
		Message:    "Email groups retrieved successfully",
		Data:       groups,
		TotalCount: int64(len(groups)),
	})
}

// UpdateEmailGroup updates an email group
func (h *alertEmailGroupHandler) UpdateEmailGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid group ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateAlertEmailGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get admin user ID from context
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid admin user ID",
		})
		return
	}

	group, err := h.emailGroupStorage.UpdateEmailGroup(c.Request.Context(), id, &req, adminUUID)
	if err != nil {
		h.log.Error("Failed to update email group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Failed to update email group",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertEmailGroupResponse{
		Success: true,
		Message: "Email group updated successfully",
		Data:    group,
	})
}

// DeleteEmailGroup deletes an email group
func (h *alertEmailGroupHandler) DeleteEmailGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Invalid group ID",
			Error:   err.Error(),
		})
		return
	}

	err = h.emailGroupStorage.DeleteEmailGroup(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete email group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertEmailGroupResponse{
			Success: false,
			Message: "Failed to delete email group",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertEmailGroupResponse{
		Success: true,
		Message: "Email group deleted successfully",
	})
}
