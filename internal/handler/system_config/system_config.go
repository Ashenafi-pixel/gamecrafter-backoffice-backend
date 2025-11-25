package system_config

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/admin_activity_logs"
	"github.com/tucanbit/internal/storage/alert"
	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

type SystemConfigHandler struct {
	systemConfigStorage *system_config.SystemConfig
	adminActivityLogs   admin_activity_logs.AdminActivityLogsStorage
	alertStorage        alert.AlertStorage
	log                 *zap.Logger
}

func NewSystemConfigHandler(db *persistencedb.PersistenceDB, adminActivityLogs admin_activity_logs.AdminActivityLogsStorage, alertStorage alert.AlertStorage, log *zap.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{
		systemConfigStorage: system_config.NewSystemConfig(db, log),
		adminActivityLogs:   adminActivityLogs,
		alertStorage:        alertStorage,
		log:                 log,
	}
}

// logAdminActivity logs an admin activity
func (h *SystemConfigHandler) logAdminActivity(ctx *gin.Context, action, resourceType, description string, details map[string]interface{}) {
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Warn("No user_id found in context, skipping activity log")
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Warn("Invalid user_id type in context, skipping activity log")
		return
	}

	req := dto.CreateAdminActivityLogReq{
		AdminUserID:  adminUUID,
		Action:       action,
		ResourceType: resourceType,
		Description:  description,
		Details:      details,
		Severity:     "info",
		Category:     "system_config",
		IPAddress:    ctx.ClientIP(),
		UserAgent:    ctx.GetHeader("User-Agent"),
	}

	_, err := h.adminActivityLogs.CreateAdminActivityLog(ctx.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to log admin activity", zap.Error(err))
	}
}

// GetWithdrawalGlobalStatus retrieves current global withdrawal status
func (h *SystemConfigHandler) GetWithdrawalGlobalStatus(ctx *gin.Context) {
	h.log.Info("Getting withdrawal global status")

	status, err := h.systemConfigStorage.GetWithdrawalGlobalStatus(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get withdrawal global status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal global status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// UpdateWithdrawalGlobalStatus updates global withdrawal status
func (h *SystemConfigHandler) UpdateWithdrawalGlobalStatus(ctx *gin.Context) {
	var req struct {
		Enabled  bool    `json:"enabled"`
		Reason   *string `json:"reason,omitempty"`
		PausedBy *string `json:"paused_by,omitempty"`
		PausedAt *string `json:"paused_at,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get admin ID from context (assuming it's set by auth middleware)
	adminID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Admin authentication required",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid admin ID",
		})
		return
	}

	status := system_config.WithdrawalGlobalStatus{
		Enabled:  req.Enabled,
		Reason:   req.Reason,
		PausedBy: req.PausedBy,
		PausedAt: req.PausedAt,
	}

	err := h.systemConfigStorage.UpdateWithdrawalGlobalStatus(ctx.Request.Context(), status, adminUUID)
	if err != nil {
		h.log.Error("Failed to update withdrawal global status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update withdrawal global status",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	action := "disable_withdrawals"
	if req.Enabled {
		action = "enable_withdrawals"
	}

	details := map[string]interface{}{
		"enabled":   req.Enabled,
		"reason":    req.Reason,
		"paused_by": req.PausedBy,
		"paused_at": req.PausedAt,
	}

	description := "Withdrawals enabled"
	if !req.Enabled {
		description = "Withdrawals disabled"
		if req.Reason != nil {
			description += " - Reason: " + *req.Reason
		}
	}

	h.logAdminActivity(ctx, action, "withdrawal_settings", description, details)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal global status updated successfully",
		"data":    status,
	})
}

// GetWithdrawalThresholds retrieves current withdrawal thresholds
func (h *SystemConfigHandler) GetWithdrawalThresholds(ctx *gin.Context) {
	h.log.Info("Getting withdrawal thresholds")

	thresholds, err := h.systemConfigStorage.GetWithdrawalThresholds(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get withdrawal thresholds", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal thresholds",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    thresholds,
	})
}

// UpdateWithdrawalThresholds updates withdrawal thresholds
func (h *SystemConfigHandler) UpdateWithdrawalThresholds(ctx *gin.Context) {
	var req system_config.WithdrawalThresholds

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get admin ID from context
	adminID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Admin authentication required",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid admin ID",
		})
		return
	}

	err := h.systemConfigStorage.UpdateWithdrawalThresholds(ctx.Request.Context(), req, adminUUID)
	if err != nil {
		h.log.Error("Failed to update withdrawal thresholds", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update withdrawal thresholds",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	details := map[string]interface{}{
		"hourly_volume":      req.HourlyVolume,
		"daily_volume":       req.DailyVolume,
		"single_transaction": req.SingleTransaction,
		"user_daily":         req.UserDaily,
	}

	h.logAdminActivity(ctx, "update_withdrawal_thresholds", "withdrawal_settings", "Withdrawal thresholds updated", details)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal thresholds updated successfully",
		"data":    req,
	})
}

// GetWithdrawalManualReview retrieves manual review settings
func (h *SystemConfigHandler) GetWithdrawalManualReview(ctx *gin.Context) {
	h.log.Info("Getting withdrawal manual review settings")

	review, err := h.systemConfigStorage.GetWithdrawalManualReview(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get withdrawal manual review settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal manual review settings",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    review,
	})
}

// UpdateWithdrawalManualReview updates manual review settings
func (h *SystemConfigHandler) UpdateWithdrawalManualReview(ctx *gin.Context) {
	var req system_config.WithdrawalManualReview

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get admin ID from context
	adminID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Admin authentication required",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid admin ID",
		})
		return
	}

	err := h.systemConfigStorage.UpdateWithdrawalManualReview(ctx.Request.Context(), req, adminUUID)
	if err != nil {
		h.log.Error("Failed to update withdrawal manual review settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update withdrawal manual review settings",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	details := map[string]interface{}{
		"enabled":                req.Enabled,
		"threshold_cents":        req.ThresholdCents,
		"require_admin_approval": req.RequireAdminApproval,
	}

	h.logAdminActivity(ctx, "update_withdrawal_manual_review", "withdrawal_settings", "Withdrawal manual review settings updated", details)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal manual review settings updated successfully",
		"data":    req,
	})
}

// CheckWithdrawalAllowed checks if withdrawals are currently allowed
func (h *SystemConfigHandler) CheckWithdrawalAllowed(ctx *gin.Context) {
	h.log.Info("Checking if withdrawals are allowed")

	allowed, reason, err := h.systemConfigStorage.CheckWithdrawalAllowed(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to check withdrawal status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check withdrawal status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"allowed": allowed,
			"reason":  reason,
		},
	})
}

// CheckWithdrawalThresholds checks if amount exceeds thresholds
func (h *SystemConfigHandler) CheckWithdrawalThresholds(ctx *gin.Context) {
	amountStr := ctx.Query("amount")
	currency := ctx.Query("currency")
	thresholdType := ctx.Query("threshold_type")

	if amountStr == "" || currency == "" || thresholdType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "amount, currency, and threshold_type parameters are required",
		})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid amount format",
		})
		return
	}

	exceeds, reason, err := h.systemConfigStorage.CheckWithdrawalThresholds(ctx.Request.Context(), amount, currency, thresholdType)
	if err != nil {
		h.log.Error("Failed to check withdrawal thresholds", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check withdrawal thresholds",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"exceeds_threshold": exceeds,
			"reason":            reason,
		},
	})
}

// GetWithdrawalPauseReasons retrieves predefined pause reasons
func (h *SystemConfigHandler) GetWithdrawalPauseReasons(ctx *gin.Context) {
	h.log.Info("Getting withdrawal pause reasons")

	reasons, err := h.systemConfigStorage.GetWithdrawalPauseReasons(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get withdrawal pause reasons", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal pause reasons",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reasons,
	})
}

// Alert Management Methods

// GetAlertConfigurations gets all alert configurations
func (h *SystemConfigHandler) GetAlertConfigurations(ctx *gin.Context) {
	var req dto.GetAlertConfigurationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid query parameters",
			"error":   err.Error(),
		})
		return
	}

	configs, totalCount, err := h.alertStorage.GetAlertConfigurations(ctx.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to get alert configurations", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get alert configurations",
			"error":   err.Error(),
		})
		return
	}

	// Set default pagination values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	// Ensure data is always an array, never nil
	if configs == nil {
		configs = []dto.AlertConfiguration{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Alert configurations retrieved successfully",
		"data":        configs,
		"total_count": totalCount,
		"page":        page,
		"per_page":    perPage,
	})
}

// CreateAlertConfiguration creates a new alert configuration
func (h *SystemConfigHandler) CreateAlertConfiguration(ctx *gin.Context) {
	var req dto.CreateAlertConfigurationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	config, err := h.alertStorage.CreateAlertConfiguration(ctx.Request.Context(), &req, userUUID)
	if err != nil {
		h.log.Error("Failed to create alert configuration", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create alert configuration",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "create", "alert_configuration", "Created alert configuration", map[string]interface{}{
		"alert_type":          config.AlertType,
		"threshold_amount":    config.ThresholdAmount,
		"time_window_minutes": config.TimeWindowMinutes,
		"currency_code":       config.CurrencyCode,
		"email_notifications": config.EmailNotifications,
	})

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Alert configuration created successfully",
		"data":    config,
	})
}

// UpdateAlertConfiguration updates an alert configuration
func (h *SystemConfigHandler) UpdateAlertConfiguration(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid alert configuration ID",
			"error":   err.Error(),
		})
		return
	}

	var req dto.UpdateAlertConfigurationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	config, err := h.alertStorage.UpdateAlertConfiguration(ctx.Request.Context(), id, &req, userUUID)
	if err != nil {
		h.log.Error("Failed to update alert configuration", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update alert configuration",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "alert_configuration", "Updated alert configuration", map[string]interface{}{
		"alert_configuration_id": id,
		"updated_fields":         req,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alert configuration updated successfully",
		"data":    config,
	})
}

// DeleteAlertConfiguration deletes an alert configuration
func (h *SystemConfigHandler) DeleteAlertConfiguration(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid alert configuration ID",
			"error":   err.Error(),
		})
		return
	}

	err = h.alertStorage.DeleteAlertConfiguration(ctx.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete alert configuration", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete alert configuration",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "delete", "alert_configuration", "Deleted alert configuration", map[string]interface{}{
		"alert_configuration_id": id,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alert configuration deleted successfully",
	})
}

// GetAlertTriggers gets alert triggers with filtering
func (h *SystemConfigHandler) GetAlertTriggers(ctx *gin.Context) {
	var req dto.GetAlertTriggersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid query parameters",
			"error":   err.Error(),
		})
		return
	}

	triggers, totalCount, err := h.alertStorage.GetAlertTriggers(ctx.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to get alert triggers", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get alert triggers",
			"error":   err.Error(),
		})
		return
	}

	// Set default pagination values
	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	// Ensure data is always an array, never nil
	if triggers == nil {
		triggers = []dto.AlertTrigger{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Alert triggers retrieved successfully",
		"data":        triggers,
		"total_count": totalCount,
		"page":        page,
		"per_page":    perPage,
	})
}

// AcknowledgeAlert acknowledges an alert trigger
func (h *SystemConfigHandler) AcknowledgeAlert(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid alert trigger ID",
			"error":   err.Error(),
		})
		return
	}

	var req dto.AcknowledgeAlertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = h.alertStorage.AcknowledgeAlert(ctx.Request.Context(), id, userUUID)
	if err != nil {
		h.log.Error("Failed to acknowledge alert", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to acknowledge alert",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "acknowledge", "alert_trigger", "Acknowledged alert trigger", map[string]interface{}{
		"alert_trigger_id": id,
		"acknowledged":     req.Acknowledged,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alert acknowledged successfully",
	})
}

// Settings Management Methods

// extractBrandIDFromJSONBody extracts brand_id from JSON body and returns the body bytes for re-binding
func (h *SystemConfigHandler) extractBrandIDFromJSONBody(ctx *gin.Context) (*uuid.UUID, []byte, error) {
	// Read the raw body
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, nil, err
	}

	// Restore the request body so it can be read again if needed
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse JSON to extract brand_id
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
		return nil, bodyBytes, err
	}

	var brandID *uuid.UUID
	if brandIDVal, exists := jsonBody["brand_id"]; exists && brandIDVal != nil {
		if brandIDStr, ok := brandIDVal.(string); ok && brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	return brandID, bodyBytes, nil
}

// GetGeneralSettings retrieves general settings
func (h *SystemConfigHandler) GetGeneralSettings(ctx *gin.Context) {
	h.log.Info("Getting general settings")

	// Get brand_id from query params (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	settings, err := h.systemConfigStorage.GetGeneralSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get general settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "General settings retrieved successfully",
	})
}

// UpdateGeneralSettings updates general settings
func (h *SystemConfigHandler) UpdateGeneralSettings(ctx *gin.Context) {
	// Extract brand_id from JSON body if present
	brandID, bodyBytes, err := h.extractBrandIDFromJSONBody(ctx)
	if err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse JSON to remove brand_id and bind to GeneralSettings
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
		h.log.Error("Failed to parse JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Remove brand_id from jsonBody before binding to GeneralSettings
	delete(jsonBody, "brand_id")

	// Convert back to JSON and bind to GeneralSettings struct
	jsonBytes, err := json.Marshal(jsonBody)
	if err != nil {
		h.log.Error("Failed to marshal JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var req system_config.GeneralSettings
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		h.log.Error("Failed to unmarshal JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// If brand_id not in JSON body, check query params or form data
	if brandID == nil {
		if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		} else if brandIDStr := ctx.PostForm("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("No user_id found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.systemConfigStorage.UpdateGeneralSettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update general settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "general_settings", "Updated general settings", map[string]interface{}{
		"site_name":            req.SiteName,
		"maintenance_mode":     req.MaintenanceMode,
		"registration_enabled": req.RegistrationEnabled,
		"demo_mode":            req.DemoMode,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "General settings updated successfully",
	})
}

// GetPaymentSettings retrieves payment settings
func (h *SystemConfigHandler) GetPaymentSettings(ctx *gin.Context) {
	h.log.Info("Getting payment settings")

	// Get brand_id from query params (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	settings, err := h.systemConfigStorage.GetPaymentSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get payment settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Payment settings retrieved successfully",
	})
}

// UpdatePaymentSettings updates payment settings
func (h *SystemConfigHandler) UpdatePaymentSettings(ctx *gin.Context) {
	var req system_config.PaymentSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("No user_id found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get brand_id from query params or request body (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	} else if brandIDStr := ctx.PostForm("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	err := h.systemConfigStorage.UpdatePaymentSettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update payment settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "payment_settings", "Updated payment settings", map[string]interface{}{
		"min_deposit_btc":    req.MinDepositBTC,
		"max_deposit_btc":    req.MaxDepositBTC,
		"min_withdrawal_btc": req.MinWithdrawalBTC,
		"max_withdrawal_btc": req.MaxWithdrawalBTC,
		"kyc_required":       req.KycRequired,
		"kyc_threshold":      req.KycThreshold,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Payment settings updated successfully",
	})
}

// GetTipSettings retrieves tip settings
func (h *SystemConfigHandler) GetTipSettings(ctx *gin.Context) {
	h.log.Info("Getting tip settings")

	// Get brand_id from query params (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	settings, err := h.systemConfigStorage.GetTipSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get tip settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Tip settings retrieved successfully",
	})
}

// UpdateTipSettings updates tip settings
func (h *SystemConfigHandler) UpdateTipSettings(ctx *gin.Context) {
	var req system_config.TipSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("No user_id found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get brand_id from query params or request body (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	} else if brandIDStr := ctx.PostForm("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	err := h.systemConfigStorage.UpdateTipSettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update tip settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "tip_settings", "Updated tip settings", map[string]interface{}{
		"tip_transaction_fee_from_who": req.TipTransactionFeeFromWho,
		"transaction_fee":              req.TransactionFee,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Tip settings updated successfully",
	})
}

// GetSecuritySettings retrieves security settings
func (h *SystemConfigHandler) GetSecuritySettings(ctx *gin.Context) {
	h.log.Info("Getting security settings")

	// Get brand_id from query params (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	settings, err := h.systemConfigStorage.GetSecuritySettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get security settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Security settings retrieved successfully",
	})
}

// UpdateSecuritySettings updates security settings
func (h *SystemConfigHandler) UpdateSecuritySettings(ctx *gin.Context) {
	// Extract brand_id from JSON body if present
	brandID, bodyBytes, err := h.extractBrandIDFromJSONBody(ctx)
	if err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Parse JSON to remove brand_id and bind to SecuritySettings
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
		h.log.Error("Failed to parse JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Remove brand_id from jsonBody before binding to SecuritySettings
	delete(jsonBody, "brand_id")

	// Convert back to JSON and bind to SecuritySettings struct
	jsonBytes, err := json.Marshal(jsonBody)
	if err != nil {
		h.log.Error("Failed to marshal JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var req system_config.SecuritySettings
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		h.log.Error("Failed to unmarshal JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// If brand_id not in JSON body, check query params or form data
	if brandID == nil {
		if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		} else if brandIDStr := ctx.PostForm("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("No user_id found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.systemConfigStorage.UpdateSecuritySettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update security settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "security_settings", "Updated security settings", map[string]interface{}{
		"session_timeout":      req.SessionTimeout,
		"max_login_attempts":   req.MaxLoginAttempts,
		"two_factor_required":  req.TwoFactorRequired,
		"password_min_length":  req.PasswordMinLength,
		"ip_whitelist_enabled": req.IpWhitelistEnabled,
		"rate_limit_enabled":   req.RateLimitEnabled,
		"rate_limit_requests":  req.RateLimitRequests,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Security settings updated successfully",
	})
}

// GetGeoBlockingSettings retrieves geo blocking settings
func (h *SystemConfigHandler) GetGeoBlockingSettings(ctx *gin.Context) {
	h.log.Info("Getting geo blocking settings")

	// Get brand_id from query params (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	settings, err := h.systemConfigStorage.GetGeoBlockingSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get geo blocking settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Geo blocking settings retrieved successfully",
	})
}

// UpdateGeoBlockingSettings updates geo blocking settings
func (h *SystemConfigHandler) UpdateGeoBlockingSettings(ctx *gin.Context) {
	var req system_config.GeoBlockingSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		h.log.Error("No user_id found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		h.log.Error("Invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get brand_id from query params or request body (optional)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	} else if brandIDStr := ctx.PostForm("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	err := h.systemConfigStorage.UpdateGeoBlockingSettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update geo blocking settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "geo_blocking_settings", "Updated geo blocking settings", map[string]interface{}{
		"enable_geo_blocking": req.EnableGeoBlocking,
		"default_action":      req.DefaultAction,
		"vpn_detection":       req.VpnDetection,
		"proxy_detection":     req.ProxyDetection,
		"tor_blocking":        req.TorBlocking,
		"log_attempts":        req.LogAttempts,
		"blocked_countries":   len(req.BlockedCountries),
		"allowed_countries":   len(req.AllowedCountries),
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Geo blocking settings updated successfully",
	})
}
