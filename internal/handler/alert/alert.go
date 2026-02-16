package alert

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/alert"
	"go.uber.org/zap"
)

type AlertHandler interface {
	CreateAlertConfiguration(c *gin.Context)
	GetAlertConfiguration(c *gin.Context)
	GetAlertConfigurations(c *gin.Context)
	UpdateAlertConfiguration(c *gin.Context)
	DeleteAlertConfiguration(c *gin.Context)
	GetAlertTriggers(c *gin.Context)
	GetAlertTrigger(c *gin.Context)
	AcknowledgeAlert(c *gin.Context)
	TestTriggerAlert(c *gin.Context) // Test endpoint to manually trigger alerts
}

type alertHandler struct {
	alertStorage      alert.AlertStorage
	emailGroupStorage alert.AlertEmailGroupStorage
	emailService      alert.AlertEmailSender
	log               *zap.Logger
}

// emailServiceAdapter adapts EmailService to AlertEmailSender interface
type emailServiceAdapter struct {
	emailService interface {
		SendAlertEmail(ctx context.Context, to []string, alertConfig interface{}, trigger interface{}) error
	}
}

func (a *emailServiceAdapter) SendAlertEmail(ctx context.Context, to []string, alertConfig *dto.AlertConfiguration, trigger *dto.AlertTrigger) error {
	return a.emailService.SendAlertEmail(ctx, to, alertConfig, trigger)
}

func NewAlertHandler(alertStorage alert.AlertStorage, emailGroupStorage alert.AlertEmailGroupStorage, emailService interface {
	SendAlertEmail(ctx context.Context, to []string, alertConfig interface{}, trigger interface{}) error
}, log *zap.Logger) AlertHandler {
	return &alertHandler{
		alertStorage:      alertStorage,
		emailGroupStorage: emailGroupStorage,
		emailService:      &emailServiceAdapter{emailService: emailService},
		log:               log,
	}
}

// CreateAlertConfiguration creates a new alert configuration
func (h *alertHandler) CreateAlertConfiguration(c *gin.Context) {
	var req dto.CreateAlertConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.AlertConfigurationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	config, err := h.alertStorage.CreateAlertConfiguration(c.Request.Context(), &req, userUUID)
	if err != nil {
		h.log.Error("Failed to create alert configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Failed to create alert configuration",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.AlertConfigurationResponse{
		Success: true,
		Message: "Alert configuration created successfully",
		Data:    config,
	})
}

// GetAlertConfiguration gets a single alert configuration by ID
func (h *alertHandler) GetAlertConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid alert configuration ID",
			Error:   err.Error(),
		})
		return
	}

	config, err := h.alertStorage.GetAlertConfigurationByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get alert configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Failed to get alert configuration",
			Error:   err.Error(),
		})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Alert configuration not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertConfigurationResponse{
		Success: true,
		Message: "Alert configuration retrieved successfully",
		Data:    config,
	})
}

// GetAlertConfigurations gets alert configurations with filtering and pagination
func (h *alertHandler) GetAlertConfigurations(c *gin.Context) {
	var req dto.GetAlertConfigurationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationsResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	configs, totalCount, err := h.alertStorage.GetAlertConfigurations(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to get alert configurations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationsResponse{
			Success: false,
			Message: "Failed to get alert configurations",
			Error:   err.Error(),
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

	c.JSON(http.StatusOK, dto.AlertConfigurationsResponse{
		Success:    true,
		Message:    "Alert configurations retrieved successfully",
		Data:       configs,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
	})
}

// UpdateAlertConfiguration updates an alert configuration
func (h *alertHandler) UpdateAlertConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid alert configuration ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateAlertConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.AlertConfigurationResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	config, err := h.alertStorage.UpdateAlertConfiguration(c.Request.Context(), id, &req, userUUID)
	if err != nil {
		h.log.Error("Failed to update alert configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Failed to update alert configuration",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertConfigurationResponse{
		Success: true,
		Message: "Alert configuration updated successfully",
		Data:    config,
	})
}

// DeleteAlertConfiguration deletes an alert configuration
func (h *alertHandler) DeleteAlertConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Invalid alert configuration ID",
			Error:   err.Error(),
		})
		return
	}

	err = h.alertStorage.DeleteAlertConfiguration(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete alert configuration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertConfigurationResponse{
			Success: false,
			Message: "Failed to delete alert configuration",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertConfigurationResponse{
		Success: true,
		Message: "Alert configuration deleted successfully",
	})
}

// GetAlertTriggers gets alert triggers with filtering and pagination
func (h *alertHandler) GetAlertTriggers(c *gin.Context) {
	var req dto.GetAlertTriggersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Error("Invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertTriggersResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	triggers, totalCount, err := h.alertStorage.GetAlertTriggers(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Failed to get alert triggers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertTriggersResponse{
			Success: false,
			Message: "Failed to get alert triggers",
			Error:   err.Error(),
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

	c.JSON(http.StatusOK, dto.AlertTriggersResponse{
		Success:    true,
		Message:    "Alert triggers retrieved successfully",
		Data:       triggers,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
	})
}

// GetAlertTrigger gets a single alert trigger by ID
func (h *alertHandler) GetAlertTrigger(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertTriggerResponse{
			Success: false,
			Message: "Invalid alert trigger ID",
			Error:   err.Error(),
		})
		return
	}

	trigger, err := h.alertStorage.GetAlertTriggerByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get alert trigger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertTriggerResponse{
			Success: false,
			Message: "Failed to get alert trigger",
			Error:   err.Error(),
		})
		return
	}

	if trigger == nil {
		c.JSON(http.StatusNotFound, dto.AlertTriggerResponse{
			Success: false,
			Message: "Alert trigger not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertTriggerResponse{
		Success: true,
		Message: "Alert trigger retrieved successfully",
		Data:    trigger,
	})
}

// AcknowledgeAlert acknowledges an alert trigger
func (h *alertHandler) AcknowledgeAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AlertTriggerResponse{
			Success: false,
			Message: "Invalid alert trigger ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.AcknowledgeAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertTriggerResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.AlertTriggerResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.AlertTriggerResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	err = h.alertStorage.AcknowledgeAlert(c.Request.Context(), id, userUUID)
	if err != nil {
		h.log.Error("Failed to acknowledge alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertTriggerResponse{
			Success: false,
			Message: "Failed to acknowledge alert",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AlertTriggerResponse{
		Success: true,
		Message: "Alert acknowledged successfully",
	})
}

// TestTriggerAlert manually creates a trigger for testing email sending
// This is a test endpoint to verify alert email functionality
func (h *alertHandler) TestTriggerAlert(c *gin.Context) {
	var req struct {
		AlertConfigurationID uuid.UUID  `json:"alert_configuration_id" binding:"required"`
		TriggerValue         float64    `json:"trigger_value" binding:"required"`
		UserID               *uuid.UUID `json:"user_id"`
		TransactionID        *string    `json:"transaction_id"`
		AmountUSD            *float64   `json:"amount_usd"`
		CurrencyCode         *string    `json:"currency_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.AlertTriggerResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get the alert configuration
	alertConfig, err := h.alertStorage.GetAlertConfigurationByID(c.Request.Context(), req.AlertConfigurationID)
	if err != nil {
		h.log.Error("Failed to get alert configuration", zap.Error(err))
		c.JSON(http.StatusNotFound, dto.AlertTriggerResponse{
			Success: false,
			Message: "Alert configuration not found",
			Error:   err.Error(),
		})
		return
	}

	if alertConfig == nil {
		c.JSON(http.StatusNotFound, dto.AlertTriggerResponse{
			Success: false,
			Message: "Alert configuration not found",
		})
		return
	}

	// Create a test trigger
	trigger := &dto.AlertTrigger{
		AlertConfigurationID: req.AlertConfigurationID,
		TriggeredAt:          time.Now(),
		TriggerValue:         req.TriggerValue,
		ThresholdValue:       alertConfig.ThresholdAmount,
		UserID:               req.UserID,
		TransactionID:        req.TransactionID,
		AmountUSD:            req.AmountUSD,
	}

	if req.CurrencyCode != nil {
		currency := dto.CurrencyType(*req.CurrencyCode)
		trigger.CurrencyCode = &currency
	}

	// Save the trigger
	err = h.alertStorage.CreateAlertTrigger(c.Request.Context(), trigger)
	if err != nil {
		h.log.Error("Failed to create alert trigger", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AlertTriggerResponse{
			Success: false,
			Message: "Failed to create alert trigger",
			Error:   err.Error(),
		})
		return
	}

	// Send emails to assigned groups
	err = alert.SendAlertEmailsToGroups(
		c.Request.Context(),
		h.emailGroupStorage,
		h.emailService,
		alertConfig,
		trigger,
		h.log,
	)

	if err != nil {
		h.log.Error("Failed to send alert emails", zap.Error(err))
		// Still return success for trigger creation, but log the email error
		c.JSON(http.StatusOK, dto.AlertTriggerResponse{
			Success: true,
			Message: "Alert trigger created successfully, but failed to send emails. Check logs for details.",
			Data:    trigger,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.AlertTriggerResponse{
		Success: true,
		Message: "Alert trigger created and emails sent successfully",
		Data:    trigger,
	})
}
