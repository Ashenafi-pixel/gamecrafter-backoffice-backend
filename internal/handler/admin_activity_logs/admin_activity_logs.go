package admin_activity_logs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/admin_activity_logs"
	"go.uber.org/zap"
)

type AdminActivityLogsHandler interface {
	GetAdminActivityLogs(c *gin.Context)
	GetAdminActivityLogByID(c *gin.Context)
	GetAdminActivityStats(c *gin.Context)
	GetAdminActivityCategories(c *gin.Context)
	GetAdminActivityActions(c *gin.Context)
	GetAdminActivityActionsByCategory(c *gin.Context)
	DeleteAdminActivityLog(c *gin.Context)
	DeleteAdminActivityLogsByAdmin(c *gin.Context)
	DeleteOldAdminActivityLogs(c *gin.Context)
}

type adminActivityLogsHandler struct {
	module admin_activity_logs.AdminActivityLogsModule
	log    *zap.Logger
}

func NewAdminActivityLogsHandler(module admin_activity_logs.AdminActivityLogsModule, log *zap.Logger) AdminActivityLogsHandler {
	return &adminActivityLogsHandler{
		module: module,
		log:    log,
	}
}

// GetAdminActivityLogs handles GET /api/admin/activity-logs
func (h *adminActivityLogsHandler) GetAdminActivityLogs(c *gin.Context) {
	var req dto.GetAdminActivityLogsReq

	// Parse query parameters
	if adminUserIDStr := c.Query("admin_user_id"); adminUserIDStr != "" {
		if adminUserID, err := uuid.Parse(adminUserIDStr); err == nil {
			req.AdminUserID = &adminUserID
		}
	}

	req.Action = c.Query("action")
	req.ResourceType = c.Query("resource_type")

	if resourceIDStr := c.Query("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			req.ResourceID = &resourceID
		}
	}

	req.Category = c.Query("category")
	req.Severity = c.Query("severity")

	// Parse date filters
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			req.From = &from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			req.To = &to
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Page = page
	req.PerPage = perPage

	// Parse sorting
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// Parse search term
	req.Search = c.Query("search")

	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	// Get logs
	logs, err := h.module.GetAdminActivityLogs(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to get admin activity logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetAdminActivityLogByID handles GET /api/admin/activity-logs/:id
func (h *adminActivityLogsHandler) GetAdminActivityLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log ID",
		})
		return
	}

	log, err := h.module.GetAdminActivityLogByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get admin activity log by ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    log,
	})
}

// GetAdminActivityStats handles GET /api/admin/activity-logs/stats
func (h *adminActivityLogsHandler) GetAdminActivityStats(c *gin.Context) {
	var from, to *time.Time

	// Parse date filters
	if fromStr := c.Query("from"); fromStr != "" {
		if parsedFrom, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &parsedFrom
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsedTo, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &parsedTo
		}
	}

	stats, err := h.module.GetAdminActivityStats(c.Request.Context(), from, to)
	if err != nil {
		h.log.Error("Failed to get admin activity stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetAdminActivityCategories handles GET /api/admin/activity-logs/categories
func (h *adminActivityLogsHandler) GetAdminActivityCategories(c *gin.Context) {
	categories, err := h.module.GetAdminActivityCategories(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get admin activity categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// GetAdminActivityActions handles GET /api/admin/activity-logs/actions
func (h *adminActivityLogsHandler) GetAdminActivityActions(c *gin.Context) {
	actions, err := h.module.GetAdminActivityActions(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get admin activity actions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity actions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    actions,
	})
}

// GetAdminActivityActionsByCategory handles GET /api/admin/activity-logs/actions/:category
func (h *adminActivityLogsHandler) GetAdminActivityActionsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category parameter is required",
		})
		return
	}

	actions, err := h.module.GetAdminActivityActionsByCategory(c.Request.Context(), category)
	if err != nil {
		h.log.Error("Failed to get admin activity actions by category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin activity actions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    actions,
	})
}

// DeleteAdminActivityLog handles DELETE /api/admin/activity-logs/:id
func (h *adminActivityLogsHandler) DeleteAdminActivityLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log ID",
		})
		return
	}

	err = h.module.DeleteAdminActivityLog(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete admin activity log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete admin activity log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Admin activity log deleted successfully",
	})
}

// DeleteAdminActivityLogsByAdmin handles DELETE /api/admin/activity-logs/admin/:admin_id
func (h *adminActivityLogsHandler) DeleteAdminActivityLogsByAdmin(c *gin.Context) {
	adminIDStr := c.Param("admin_id")
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID",
		})
		return
	}

	err = h.module.DeleteAdminActivityLogsByAdmin(c.Request.Context(), adminID)
	if err != nil {
		h.log.Error("Failed to delete admin activity logs by admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete admin activity logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Admin activity logs deleted successfully",
	})
}

// DeleteOldAdminActivityLogs handles DELETE /api/admin/activity-logs/cleanup
func (h *adminActivityLogsHandler) DeleteOldAdminActivityLogs(c *gin.Context) {
	var req struct {
		Before string `json:"before" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	before, err := time.Parse(time.RFC3339, req.Before)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use RFC3339 format",
		})
		return
	}

	err = h.module.DeleteOldAdminActivityLogs(c.Request.Context(), before)
	if err != nil {
		h.log.Error("Failed to delete old admin activity logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete old admin activity logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Old admin activity logs deleted successfully",
	})
}
