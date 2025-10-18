package withdrawal_pause

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/platform"
	"github.com/tucanbit/internal/utils/response"
	"go.uber.org/zap"
)

type WithdrawalPauseHandler struct {
	withdrawalPauseStorage *WithdrawalPause
	log                    *zap.Logger
}

func NewWithdrawalPauseHandler(db *platform.Platform, log *zap.Logger) *WithdrawalPauseHandler {
	return &WithdrawalPauseHandler{
		withdrawalPauseStorage: NewWithdrawalPause(db, log),
		log:                    log,
	}
}

// GetWithdrawalPauseSettings retrieves current pause settings
func (w *WithdrawalPauseHandler) GetWithdrawalPauseSettings(ctx *gin.Context) {
	w.log.Info("Getting withdrawal pause settings")

	settings, err := w.withdrawalPauseStorage.GetWithdrawalPauseSettings(ctx.Request.Context())
	if err != nil {
		w.log.Error("Failed to get withdrawal pause settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get withdrawal pause settings",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, settings)
}

// UpdateWithdrawalPauseSettings updates global pause settings
func (w *WithdrawalPauseHandler) UpdateWithdrawalPauseSettings(ctx *gin.Context) {
	var req dto.UpdateWithdrawalPauseSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		w.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := ctx.Get("user_id")
	if !exists {
		w.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		w.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid admin ID",
		})
		return
	}

	err := w.withdrawalPauseStorage.UpdateWithdrawalPauseSettings(ctx.Request.Context(), req, adminUUID)
	if err != nil {
		w.log.Error("Failed to update withdrawal pause settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update withdrawal pause settings",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"message": "Withdrawal pause settings updated successfully",
	})
}

// GetWithdrawalThresholds retrieves all active thresholds
func (w *WithdrawalPauseHandler) GetWithdrawalThresholds(ctx *gin.Context) {
	w.log.Info("Getting withdrawal thresholds")

	thresholds, err := w.withdrawalPauseStorage.GetWithdrawalThresholds(ctx.Request.Context())
	if err != nil {
		w.log.Error("Failed to get withdrawal thresholds", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get withdrawal thresholds",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"thresholds": thresholds,
	})
}

// CreateWithdrawalThreshold creates a new threshold
func (w *WithdrawalPauseHandler) CreateWithdrawalThreshold(ctx *gin.Context) {
	var req dto.CreateWithdrawalThresholdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		w.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Get admin ID from context
	adminID, exists := ctx.Get("user_id")
	if !exists {
		w.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		w.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid admin ID",
		})
		return
	}

	threshold, err := w.withdrawalPauseStorage.CreateWithdrawalThreshold(ctx.Request.Context(), req, adminUUID)
	if err != nil {
		w.log.Error("Failed to create withdrawal threshold", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create withdrawal threshold",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusCreated, gin.H{
		"message":   "Withdrawal threshold created successfully",
		"threshold": threshold,
	})
}

// UpdateWithdrawalThreshold updates an existing threshold
func (w *WithdrawalPauseHandler) UpdateWithdrawalThreshold(ctx *gin.Context) {
	thresholdIDStr := ctx.Param("id")
	thresholdID, err := uuid.Parse(thresholdIDStr)
	if err != nil {
		w.log.Error("Invalid threshold ID", zap.String("id", thresholdIDStr), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid threshold ID",
		})
		return
	}

	var req dto.UpdateWithdrawalThresholdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		w.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	threshold, err := w.withdrawalPauseStorage.UpdateWithdrawalThreshold(ctx.Request.Context(), thresholdID, req)
	if err != nil {
		w.log.Error("Failed to update withdrawal threshold", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update withdrawal threshold",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"message":   "Withdrawal threshold updated successfully",
		"threshold": threshold,
	})
}

// DeleteWithdrawalThreshold deletes a threshold
func (w *WithdrawalPauseHandler) DeleteWithdrawalThreshold(ctx *gin.Context) {
	thresholdIDStr := ctx.Param("id")
	thresholdID, err := uuid.Parse(thresholdIDStr)
	if err != nil {
		w.log.Error("Invalid threshold ID", zap.String("id", thresholdIDStr), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid threshold ID",
		})
		return
	}

	err = w.withdrawalPauseStorage.DeleteWithdrawalThreshold(ctx.Request.Context(), thresholdID)
	if err != nil {
		w.log.Error("Failed to delete withdrawal threshold", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete withdrawal threshold",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"message": "Withdrawal threshold deleted successfully",
	})
}

// GetPausedWithdrawals retrieves paused withdrawals with pagination
func (w *WithdrawalPauseHandler) GetPausedWithdrawals(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "10"))

	var status *string
	if statusStr := ctx.Query("status"); statusStr != "" {
		status = &statusStr
	}

	var pauseReason *string
	if reasonStr := ctx.Query("pause_reason"); reasonStr != "" {
		pauseReason = &reasonStr
	}

	var userID *uuid.UUID
	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		if parsedUUID, err := uuid.Parse(userIDStr); err == nil {
			userID = &parsedUUID
		}
	}

	req := dto.GetPausedWithdrawalsRequest{
		Page:        page,
		PerPage:     perPage,
		Status:      status,
		PauseReason: pauseReason,
		UserID:      userID,
	}

	withdrawals, err := w.withdrawalPauseStorage.GetPausedWithdrawals(ctx.Request.Context(), req)
	if err != nil {
		w.log.Error("Failed to get paused withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get paused withdrawals",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, withdrawals)
}

// ApproveWithdrawal approves a paused withdrawal
func (w *WithdrawalPauseHandler) ApproveWithdrawal(ctx *gin.Context) {
	withdrawalIDStr := ctx.Param("id")
	withdrawalID, err := uuid.Parse(withdrawalIDStr)
	if err != nil {
		w.log.Error("Invalid withdrawal ID", zap.String("id", withdrawalIDStr), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid withdrawal ID",
		})
		return
	}

	var req dto.WithdrawalPauseActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		w.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Get admin ID from context
	adminID, exists := ctx.Get("user_id")
	if !exists {
		w.log.Error("Admin ID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminID.(uuid.UUID)
	if !ok {
		w.log.Error("Invalid admin ID type")
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid admin ID",
		})
		return
	}

	if req.Action == "approved" {
		err = w.withdrawalPauseStorage.ApproveWithdrawal(ctx.Request.Context(), withdrawalID, adminUUID, req.Notes)
		if err != nil {
			w.log.Error("Failed to approve withdrawal", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to approve withdrawal",
			})
			return
		}
	} else if req.Action == "rejected" {
		err = w.withdrawalPauseStorage.RejectWithdrawal(ctx.Request.Context(), withdrawalID, adminUUID, req.Notes)
		if err != nil {
			w.log.Error("Failed to reject withdrawal", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to reject withdrawal",
			})
			return
		}
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"message": "Withdrawal action completed successfully",
	})
}

// GetWithdrawalPauseStats retrieves statistics for the pause system
func (w *WithdrawalPauseHandler) GetWithdrawalPauseStats(ctx *gin.Context) {
	w.log.Info("Getting withdrawal pause stats")

	stats, err := w.withdrawalPauseStorage.GetWithdrawalPauseStats(ctx.Request.Context())
	if err != nil {
		w.log.Error("Failed to get withdrawal pause stats", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get withdrawal pause stats",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, stats)
}

// GetWithdrawalPauseStatus retrieves current pause status and metrics
func (w *WithdrawalPauseHandler) GetWithdrawalPauseStatus(ctx *gin.Context) {
	w.log.Info("Getting withdrawal pause status")

	// Get pause settings
	settings, err := w.withdrawalPauseStorage.GetWithdrawalPauseSettings(ctx.Request.Context())
	if err != nil {
		w.log.Error("Failed to get pause settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get pause settings",
		})
		return
	}

	// Get stats
	stats, err := w.withdrawalPauseStorage.GetWithdrawalPauseStats(ctx.Request.Context())
	if err != nil {
		w.log.Error("Failed to get pause stats", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get pause stats",
		})
		return
	}

	response.SendSuccessResponse(ctx, http.StatusOK, gin.H{
		"settings": settings,
		"stats":    stats,
	})
}






