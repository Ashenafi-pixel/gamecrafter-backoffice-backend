package withdrawal_management

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage/withdrawal_management"
	"go.uber.org/zap"
)

type WithdrawalManagementHandler struct {
	withdrawalManagement *withdrawal_management.WithdrawalManagement
	log                  *zap.Logger
}

func NewWithdrawalManagementHandler(db *persistencedb.PersistenceDB, log *zap.Logger) *WithdrawalManagementHandler {
	return &WithdrawalManagementHandler{
		withdrawalManagement: withdrawal_management.NewWithdrawalManagement(db, log),
		log:                  log,
	}
}

// GetPausedWithdrawals retrieves all paused withdrawals with pagination
func (h *WithdrawalManagementHandler) GetPausedWithdrawals(ctx *gin.Context) {
	h.log.Info("Getting paused withdrawals")

	// Parse pagination parameters
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get paused withdrawals
	withdrawals, total, err := h.withdrawalManagement.GetPausedWithdrawals(ctx.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("Failed to get paused withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get paused withdrawals",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"withdrawals": withdrawals,
			"pagination": gin.H{
				"total":  total,
				"limit":  limit,
				"offset": offset,
			},
		},
	})
}

// PauseWithdrawal pauses a specific withdrawal
func (h *WithdrawalManagementHandler) PauseWithdrawal(ctx *gin.Context) {
	withdrawalID := ctx.Param("id")
	if withdrawalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var req struct {
		Reason           string   `json:"reason" binding:"required"`
		RequiresReview   bool     `json:"requires_review"`
		ThresholdType    *string  `json:"threshold_type,omitempty"`
		ThresholdValue   *float64 `json:"threshold_value,omitempty"`
		Notes            *string  `json:"notes,omitempty"`
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

	err := h.withdrawalManagement.PauseWithdrawal(
		ctx.Request.Context(),
		withdrawalID,
		req.Reason,
		&adminUUID,
		req.RequiresReview,
		req.ThresholdType,
		req.ThresholdValue,
	)
	if err != nil {
		h.log.Error("Failed to pause withdrawal", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to pause withdrawal",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal paused successfully",
	})
}

// UnpauseWithdrawal unpauses a specific withdrawal
func (h *WithdrawalManagementHandler) UnpauseWithdrawal(ctx *gin.Context) {
	withdrawalID := ctx.Param("id")
	if withdrawalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Withdrawal ID is required",
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

	err := h.withdrawalManagement.UnpauseWithdrawal(ctx.Request.Context(), withdrawalID, &adminUUID)
	if err != nil {
		h.log.Error("Failed to unpause withdrawal", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unpause withdrawal",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal unpaused successfully",
	})
}

// ApproveWithdrawal approves a paused withdrawal
func (h *WithdrawalManagementHandler) ApproveWithdrawal(ctx *gin.Context) {
	withdrawalID := ctx.Param("id")
	if withdrawalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var req struct {
		Notes *string `json:"notes,omitempty"`
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

	err := h.withdrawalManagement.ApproveWithdrawal(ctx.Request.Context(), withdrawalID, adminUUID, req.Notes)
	if err != nil {
		h.log.Error("Failed to approve withdrawal", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to approve withdrawal",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal approved successfully",
	})
}

// RejectWithdrawal rejects a paused withdrawal
func (h *WithdrawalManagementHandler) RejectWithdrawal(ctx *gin.Context) {
	withdrawalID := ctx.Param("id")
	if withdrawalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var req struct {
		Notes *string `json:"notes,omitempty"`
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

	err := h.withdrawalManagement.RejectWithdrawal(ctx.Request.Context(), withdrawalID, adminUUID, req.Notes)
	if err != nil {
		h.log.Error("Failed to reject withdrawal", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to reject withdrawal",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdrawal rejected successfully",
	})
}

// GetWithdrawalPauseStats retrieves statistics about paused withdrawals
func (h *WithdrawalManagementHandler) GetWithdrawalPauseStats(ctx *gin.Context) {
	h.log.Info("Getting withdrawal pause stats")

	stats, err := h.withdrawalManagement.GetWithdrawalPauseStats(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get withdrawal pause stats", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal pause stats",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
