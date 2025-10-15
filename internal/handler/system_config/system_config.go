package system_config

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/system_config"
	"go.uber.org/zap"
)

type SystemConfigHandler struct {
	systemConfigStorage *system_config.SystemConfig
	log                 *zap.Logger
}

func NewSystemConfigHandler(db *persistencedb.PersistenceDB, log *zap.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{
		systemConfigStorage: system_config.NewSystemConfig(db, log),
		log:                 log,
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
