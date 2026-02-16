package system_config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/module/game_import"
	"github.com/tucanbit/internal/storage/admin_activity_logs"
	"github.com/tucanbit/internal/storage/alert"
	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

type SystemConfigHandler struct {
	systemConfigStorage *system_config.SystemConfig
	adminActivityLogs   admin_activity_logs.AdminActivityLogsStorage
	alertStorage        alert.AlertStorage
	gameImportService   GameImportServiceInterface
	log                 *zap.Logger
}

type GameImportServiceInterface interface {
	ImportGames(ctx context.Context, brandID uuid.UUID) (*game_import.ImportResult, error)
}

func NewSystemConfigHandler(db *persistencedb.PersistenceDB, adminActivityLogs admin_activity_logs.AdminActivityLogsStorage, alertStorage alert.AlertStorage, gameImportService GameImportServiceInterface, log *zap.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{
		systemConfigStorage: system_config.NewSystemConfig(db, log),
		adminActivityLogs:   adminActivityLogs,
		alertStorage:        alertStorage,
		gameImportService:   gameImportService,
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
	// Read JSON body to extract both brand_id and tip settings
	var jsonBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Extract brand_id from JSON body, query params, or form data (required)
	var brandID *uuid.UUID
	if brandIDVal, exists := jsonBody["brand_id"]; exists {
		if brandIDStr, ok := brandIDVal.(string); ok && brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	// If not in JSON body, check query params or form data
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

	// brand_id is required for tip settings
	if brandID == nil {
		h.log.Error("brand_id is required for tip settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for tip settings"})
		return
	}

	// Parse the TipSettings from JSON body
	var req system_config.TipSettings

	// Parse tip_transaction_fee_from_who
	if tipTransactionFeeFromWho, ok := jsonBody["tip_transaction_fee_from_who"].(string); ok && tipTransactionFeeFromWho != "" {
		req.TipTransactionFeeFromWho = tipTransactionFeeFromWho
	} else {
		h.log.Warn("tip_transaction_fee_from_who missing or invalid, using default")
		req.TipTransactionFeeFromWho = "sender" // Default value
	}

	// Parse transaction_fee - handle both float64 and int types from JSON
	if transactionFeeVal, exists := jsonBody["transaction_fee"]; exists {
		switch v := transactionFeeVal.(type) {
		case float64:
			req.TransactionFee = v
		case int:
			req.TransactionFee = float64(v)
		case int64:
			req.TransactionFee = float64(v)
		case float32:
			req.TransactionFee = float64(v)
		default:
			h.log.Warn("transaction_fee has invalid type, using 0.0", zap.Any("type", fmt.Sprintf("%T", v)))
			req.TransactionFee = 0.0
		}
	} else {
		h.log.Warn("transaction_fee missing, using default 0.0")
		req.TransactionFee = 0.0
	}

	h.log.Info("Parsed tip settings",
		zap.String("tip_transaction_fee_from_who", req.TipTransactionFeeFromWho),
		zap.Float64("transaction_fee", req.TransactionFee),
		zap.Any("brand_id", brandID))

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

// GetWelcomeBonusSettings retrieves welcome bonus settings
func (h *SystemConfigHandler) GetWelcomeBonusSettings(ctx *gin.Context) {
	h.log.Info("Getting welcome bonus settings")

	// Get brand_id from query params (required)
	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	// brand_id is required for welcome bonus settings
	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus settings"})
		return
	}

	settings, err := h.systemConfigStorage.GetWelcomeBonusSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get welcome bonus settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Welcome bonus settings retrieved successfully",
	})
}

// UpdateWelcomeBonusSettings updates welcome bonus settings
func (h *SystemConfigHandler) UpdateWelcomeBonusSettings(ctx *gin.Context) {
	// Read JSON body to extract both brand_id and welcome bonus settings
	var jsonBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Extract brand_id from JSON body, query params, or form data (required)
	var brandID *uuid.UUID
	if brandIDVal, exists := jsonBody["brand_id"]; exists {
		if brandIDStr, ok := brandIDVal.(string); ok && brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	// If not in JSON body, check query params or form data
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

	// brand_id is required for welcome bonus settings
	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus settings"})
		return
	}

	// Parse the WelcomeBonusSettings from JSON body
	var req system_config.WelcomeBonusSettings

	// Parse new fields: fixed_enabled and percentage_enabled
	if fixedEnabledVal, exists := jsonBody["fixed_enabled"]; exists {
		if fixedEnabled, ok := fixedEnabledVal.(bool); ok {
			req.FixedEnabled = fixedEnabled
		}
	}

	if percentageEnabledVal, exists := jsonBody["percentage_enabled"]; exists {
		if percentageEnabled, ok := percentageEnabledVal.(bool); ok {
			req.PercentageEnabled = percentageEnabled
		}
	}

	// Parse IP based restriction anti-abuse settings
	if ipRestrictionEnabledVal, exists := jsonBody["ip_restriction_enabled"]; exists {
		if ipRestrictionEnabled, ok := ipRestrictionEnabledVal.(bool); ok {
			req.IPRestrictionEnabled = ipRestrictionEnabled
		}
	}
	if allowMultiplePerIPVal, exists := jsonBody["allow_multiple_bonuses_per_ip"]; exists {
		if allowMultiplePerIP, ok := allowMultiplePerIPVal.(bool); ok {
			req.AllowMultipleBonusesPerIP = allowMultiplePerIP
		}
	}

	// Backward compatibility: parse old type and enabled fields
	if typeVal, ok := jsonBody["type"].(string); ok && typeVal != "" {
		if typeVal == "fixed" || typeVal == "percentage" {
			req.Type = typeVal
		}
	}

	if enabledVal, exists := jsonBody["enabled"]; exists {
		if enabled, ok := enabledVal.(bool); ok {
			req.Enabled = enabled
		}
	}

	// Helper function to parse float64 from JSON (handles int, int64, float32, float64)
	parseFloat := func(key string, defaultValue float64) float64 {
		if val, exists := jsonBody[key]; exists {
			switch v := val.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case float32:
				return float64(v)
			default:
				h.log.Warn(fmt.Sprintf("%s has invalid type, using default %.2f", key, defaultValue), zap.Any("type", fmt.Sprintf("%T", v)))
				return defaultValue
			}
		}
		return defaultValue
	}

	// Parse fixed_amount
	req.FixedAmount = parseFloat("fixed_amount", 0.0)

	// Parse percentage
	req.Percentage = parseFloat("percentage", 0.0)

	// Parse max_deposit_amount (with backward compatibility for min_deposit_amount)
	if _, exists := jsonBody["max_deposit_amount"]; exists {
		req.MaxDepositAmount = parseFloat("max_deposit_amount", 0.0)
	} else if _, exists := jsonBody["min_deposit_amount"]; exists {
		// Backward compatibility: accept min_deposit_amount and convert to max_deposit_amount
		req.MaxDepositAmount = parseFloat("min_deposit_amount", 0.0)
		h.log.Info("Migrated min_deposit_amount to max_deposit_amount in request", zap.Float64("value", req.MaxDepositAmount))
	} else {
		req.MaxDepositAmount = 0.0
	}

	// Parse max_bonus_percentage
	req.MaxBonusPercentage = parseFloat("max_bonus_percentage", 90.0)

	// Validate: max_bonus_percentage must be < 100 if percentage is enabled
	if req.PercentageEnabled {
		if req.MaxBonusPercentage >= 100 {
			h.log.Error("max_bonus_percentage must be less than 100")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "max_bonus_percentage must be less than 100 to prevent bonus from equaling deposit"})
			return
		}
	}

	h.log.Info("Parsed welcome bonus settings",
		zap.Bool("fixed_enabled", req.FixedEnabled),
		zap.Bool("percentage_enabled", req.PercentageEnabled),
		zap.Bool("ip_restriction_enabled", req.IPRestrictionEnabled),
		zap.Bool("allow_multiple_bonuses_per_ip", req.AllowMultipleBonusesPerIP),
		zap.Float64("fixed_amount", req.FixedAmount),
		zap.Float64("percentage", req.Percentage),
		zap.Float64("max_deposit_amount", req.MaxDepositAmount),
		zap.Float64("max_bonus_percentage", req.MaxBonusPercentage),
		zap.Any("brand_id", brandID))

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

	err := h.systemConfigStorage.UpdateWelcomeBonusSettings(ctx.Request.Context(), req, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update welcome bonus settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "welcome_bonus_settings", "Updated welcome bonus settings", map[string]interface{}{
		"fixed_enabled":                 req.FixedEnabled,
		"percentage_enabled":            req.PercentageEnabled,
		"ip_restriction_enabled":        req.IPRestrictionEnabled,
		"allow_multiple_bonuses_per_ip": req.AllowMultipleBonusesPerIP,
		"fixed_amount":                  req.FixedAmount,
		"percentage":                    req.Percentage,
		"max_deposit_amount":            req.MaxDepositAmount,
		"max_bonus_percentage":          req.MaxBonusPercentage,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Welcome bonus settings updated successfully",
	})
}

func (h *SystemConfigHandler) GetWelcomeBonusChannels(ctx *gin.Context) {
	h.log.Info("Getting welcome bonus channel settings")

	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus channel settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus channel settings"})
		return
	}

	settings, err := h.systemConfigStorage.GetWelcomeBonusChannelSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get welcome bonus channel settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "Welcome bonus channel settings retrieved successfully",
	})
}

func (h *SystemConfigHandler) CreateWelcomeBonusChannel(ctx *gin.Context) {
	var jsonBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	var brandID *uuid.UUID
	if brandIDVal, exists := jsonBody["brand_id"]; exists {
		if brandIDStr, ok := brandIDVal.(string); ok && brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	if brandID == nil {
		if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus channel settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus channel settings"})
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentSettings, err := h.systemConfigStorage.GetWelcomeBonusChannelSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get current channel settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	var newChannel system_config.WelcomeBonusChannelRule
	if channelVal, exists := jsonBody["channel"]; exists {
		if channel, ok := channelVal.(string); ok {
			newChannel.Channel = channel
		}
	}

	if patternsVal, exists := jsonBody["referrer_patterns"]; exists {
		if patterns, ok := patternsVal.([]interface{}); ok {
			newChannel.ReferrerPatterns = make([]string, 0, len(patterns))
			for _, p := range patterns {
				if pattern, ok := p.(string); ok {
					newChannel.ReferrerPatterns = append(newChannel.ReferrerPatterns, pattern)
				}
			}
		}
	}

	if enabledVal, exists := jsonBody["enabled"]; exists {
		if enabled, ok := enabledVal.(bool); ok {
			newChannel.Enabled = enabled
		}
	}

	if bonusTypeVal, exists := jsonBody["bonus_type"]; exists {
		if bonusType, ok := bonusTypeVal.(string); ok {
			newChannel.BonusType = bonusType
		}
	}

	if fixedAmountVal, exists := jsonBody["fixed_amount"]; exists {
		if fixedAmount, ok := fixedAmountVal.(float64); ok {
			newChannel.FixedAmount = fixedAmount
		}
	}

	if percentageVal, exists := jsonBody["percentage"]; exists {
		if percentage, ok := percentageVal.(float64); ok {
			newChannel.Percentage = percentage
		}
	}

	if maxBonusPercentageVal, exists := jsonBody["max_bonus_percentage"]; exists {
		if maxBonusPercentage, ok := maxBonusPercentageVal.(float64); ok {
			newChannel.MaxBonusPercentage = maxBonusPercentage
		}
	}

	if maxDepositAmountVal, exists := jsonBody["max_deposit_amount"]; exists {
		if maxDepositAmount, ok := maxDepositAmountVal.(float64); ok {
			newChannel.MaxDepositAmount = maxDepositAmount
		}
	}

	if inheritIPPolicyVal, exists := jsonBody["inherit_ip_policy"]; exists {
		if inheritIPPolicy, ok := inheritIPPolicyVal.(bool); ok {
			newChannel.InheritIPPolicy = inheritIPPolicy
		}
	} else {
		newChannel.InheritIPPolicy = true
	}

	if ipRestrictionEnabledVal, exists := jsonBody["ip_restriction_enabled"]; exists {
		if ipRestrictionEnabled, ok := ipRestrictionEnabledVal.(bool); ok {
			newChannel.IPRestrictionEnabled = ipRestrictionEnabled
		}
	}

	if allowMultipleBonusesPerIPVal, exists := jsonBody["allow_multiple_bonuses_per_ip"]; exists {
		if allowMultipleBonusesPerIP, ok := allowMultipleBonusesPerIPVal.(bool); ok {
			newChannel.AllowMultipleBonusesPerIP = allowMultipleBonusesPerIP
		}
	}

	newChannel.ID = uuid.New().String()

	currentSettings.Channels = append(currentSettings.Channels, newChannel)

	err = h.systemConfigStorage.UpdateWelcomeBonusChannelSettings(ctx.Request.Context(), currentSettings, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to create welcome bonus channel", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	h.logAdminActivity(ctx, "create", "welcome_bonus_channel", "Created welcome bonus channel", map[string]interface{}{
		"channel_id": newChannel.ID,
		"channel":    newChannel.Channel,
		"brand_id":   brandID.String(),
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    newChannel,
		"message": "Welcome bonus channel created successfully",
	})
}

func (h *SystemConfigHandler) UpdateWelcomeBonusChannel(ctx *gin.Context) {
	channelID := ctx.Param("id")
	if channelID == "" {
		h.log.Error("channel id is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channel id is required"})
		return
	}

	var jsonBody map[string]interface{}
	if err := ctx.ShouldBindJSON(&jsonBody); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	var brandID *uuid.UUID
	if brandIDVal, exists := jsonBody["brand_id"]; exists {
		if brandIDStr, ok := brandIDVal.(string); ok && brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	if brandID == nil {
		if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
			if parsed, err := uuid.Parse(brandIDStr); err == nil {
				brandID = &parsed
			}
		}
	}

	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus channel settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus channel settings"})
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentSettings, err := h.systemConfigStorage.GetWelcomeBonusChannelSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get current channel settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	var foundIndex = -1
	for i, channel := range currentSettings.Channels {
		if channel.ID == channelID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		h.log.Error("Channel not found", zap.String("channel_id", channelID))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	updatedChannel := currentSettings.Channels[foundIndex]

	if channelVal, exists := jsonBody["channel"]; exists {
		if channel, ok := channelVal.(string); ok {
			updatedChannel.Channel = channel
		}
	}

	if patternsVal, exists := jsonBody["referrer_patterns"]; exists {
		if patterns, ok := patternsVal.([]interface{}); ok {
			updatedChannel.ReferrerPatterns = make([]string, 0, len(patterns))
			for _, p := range patterns {
				if pattern, ok := p.(string); ok {
					updatedChannel.ReferrerPatterns = append(updatedChannel.ReferrerPatterns, pattern)
				}
			}
		}
	}

	if enabledVal, exists := jsonBody["enabled"]; exists {
		if enabled, ok := enabledVal.(bool); ok {
			updatedChannel.Enabled = enabled
		}
	}

	if bonusTypeVal, exists := jsonBody["bonus_type"]; exists {
		if bonusType, ok := bonusTypeVal.(string); ok {
			updatedChannel.BonusType = bonusType
		}
	}

	if fixedAmountVal, exists := jsonBody["fixed_amount"]; exists {
		if fixedAmount, ok := fixedAmountVal.(float64); ok {
			updatedChannel.FixedAmount = fixedAmount
		}
	}

	if percentageVal, exists := jsonBody["percentage"]; exists {
		if percentage, ok := percentageVal.(float64); ok {
			updatedChannel.Percentage = percentage
		}
	}

	if maxBonusPercentageVal, exists := jsonBody["max_bonus_percentage"]; exists {
		if maxBonusPercentage, ok := maxBonusPercentageVal.(float64); ok {
			updatedChannel.MaxBonusPercentage = maxBonusPercentage
		}
	}

	if maxDepositAmountVal, exists := jsonBody["max_deposit_amount"]; exists {
		if maxDepositAmount, ok := maxDepositAmountVal.(float64); ok {
			updatedChannel.MaxDepositAmount = maxDepositAmount
		}
	}

	if inheritIPPolicyVal, exists := jsonBody["inherit_ip_policy"]; exists {
		if inheritIPPolicy, ok := inheritIPPolicyVal.(bool); ok {
			updatedChannel.InheritIPPolicy = inheritIPPolicy
		}
	}

	if ipRestrictionEnabledVal, exists := jsonBody["ip_restriction_enabled"]; exists {
		if ipRestrictionEnabled, ok := ipRestrictionEnabledVal.(bool); ok {
			updatedChannel.IPRestrictionEnabled = ipRestrictionEnabled
		}
	}

	if allowMultipleBonusesPerIPVal, exists := jsonBody["allow_multiple_bonuses_per_ip"]; exists {
		if allowMultipleBonusesPerIP, ok := allowMultipleBonusesPerIPVal.(bool); ok {
			updatedChannel.AllowMultipleBonusesPerIP = allowMultipleBonusesPerIP
		}
	}

	currentSettings.Channels[foundIndex] = updatedChannel

	err = h.systemConfigStorage.UpdateWelcomeBonusChannelSettings(ctx.Request.Context(), currentSettings, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to update welcome bonus channel", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	h.logAdminActivity(ctx, "update", "welcome_bonus_channel", "Updated welcome bonus channel", map[string]interface{}{
		"channel_id": channelID,
		"brand_id":   brandID.String(),
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedChannel,
		"message": "Welcome bonus channel updated successfully",
	})
}

func (h *SystemConfigHandler) DeleteWelcomeBonusChannel(ctx *gin.Context) {
	channelID := ctx.Param("id")
	if channelID == "" {
		h.log.Error("channel id is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "channel id is required"})
		return
	}

	var brandID *uuid.UUID
	if brandIDStr := ctx.Query("brand_id"); brandIDStr != "" {
		if parsed, err := uuid.Parse(brandIDStr); err == nil {
			brandID = &parsed
		}
	}

	if brandID == nil {
		h.log.Error("brand_id is required for welcome bonus channel settings")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "brand_id is required for welcome bonus channel settings"})
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentSettings, err := h.systemConfigStorage.GetWelcomeBonusChannelSettings(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get current channel settings", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	var foundIndex = -1
	for i, channel := range currentSettings.Channels {
		if channel.ID == channelID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		h.log.Error("Channel not found", zap.String("channel_id", channelID))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	currentSettings.Channels = append(currentSettings.Channels[:foundIndex], currentSettings.Channels[foundIndex+1:]...)

	err = h.systemConfigStorage.UpdateWelcomeBonusChannelSettings(ctx.Request.Context(), currentSettings, adminUUID, brandID)
	if err != nil {
		h.log.Error("Failed to delete welcome bonus channel", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	h.logAdminActivity(ctx, "delete", "welcome_bonus_channel", "Deleted welcome bonus channel", map[string]interface{}{
		"channel_id": channelID,
		"brand_id":   brandID.String(),
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Welcome bonus channel deleted successfully",
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

// retrieves game import configuration for a brand
func (h *SystemConfigHandler) GetGameImportConfig(ctx *gin.Context) {
	brandIDStr := ctx.Query("brand_id")
	if brandIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "brand_id is required",
		})
		return
	}

	brandID, err := uuid.Parse(brandIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid brand_id format",
		})
		return
	}

	config, err := h.systemConfigStorage.GetGameImportConfig(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Failed to get game import config", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get game import config",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// updates game import configuration
func (h *SystemConfigHandler) UpdateGameImportConfig(ctx *gin.Context) {
	var req system_config.GameImportConfig
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if err := h.systemConfigStorage.UpdateGameImportConfig(ctx.Request.Context(), req); err != nil {
		h.log.Error("Failed to update game import config", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update game import config",
			"error":   err.Error(),
		})
		return
	}

	h.logAdminActivity(ctx, "update", "game_import_config", "Updated game import configuration", map[string]interface{}{
		"brand_id":      req.BrandID.String(),
		"schedule_type": req.ScheduleType,
		"is_active":     req.IsActive,
		"providers":     req.Providers,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
		"message": "Game import configuration updated successfully",
	})
}

// Manually triggers a game import
func (h *SystemConfigHandler) TriggerGameImport(ctx *gin.Context) {
	brandIDStr := ctx.Query("brand_id")
	if brandIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "brand_id is required",
		})
		return
	}

	brandID, err := uuid.Parse(brandIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid brand_id format",
		})
		return
	}

	if h.gameImportService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Game import service is not configured",
		})
		return
	}

	h.log.Info("Manual game import triggered", zap.String("brand_id", brandID.String()))

	result, err := h.gameImportService.ImportGames(ctx.Request.Context(), brandID)
	if err != nil {
		h.log.Error("Game import failed", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Game import failed",
			"error":   err.Error(),
		})
		return
	}

	h.logAdminActivity(ctx, "trigger", "game_import", "Manually triggered game import", map[string]interface{}{
		"brand_id":             brandID.String(),
		"games_imported":       result.GamesImported,
		"house_edges_imported": result.HouseEdgesImported,
		"total_fetched":        result.TotalFetched,
		"filtered":             result.Filtered,
		"duplicates_skipped":   result.DuplicatesSkipped,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Game import completed successfully",
	})
}
