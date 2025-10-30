package kyc

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage/admin_activity_logs"
	"github.com/tucanbit/internal/storage/kyc"
	"go.uber.org/zap"
)

type KYCHandler struct {
	kycStorage        kyc.KYCStorage
	adminActivityLogs admin_activity_logs.AdminActivityLogsStorage
	log               *zap.Logger
}

func NewKYCHandler(kycStorage kyc.KYCStorage, adminActivityLogs admin_activity_logs.AdminActivityLogsStorage, log *zap.Logger) *KYCHandler {
	return &KYCHandler{
		kycStorage:        kycStorage,
		adminActivityLogs: adminActivityLogs,
		log:               log,
	}
}

// logAdminActivity logs an admin activity
func (h *KYCHandler) logAdminActivity(ctx *gin.Context, action, resourceType, description string, details map[string]interface{}) {
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
		Category:     "kyc",
		IPAddress:    ctx.ClientIP(),
		UserAgent:    ctx.GetHeader("User-Agent"),
	}

	_, err := h.adminActivityLogs.CreateAdminActivityLog(ctx.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to log admin activity", zap.Error(err))
	}
}

// CreateKYCDocument manually creates a KYC document (for testing)
func (h *KYCHandler) CreateKYCDocument(ctx *gin.Context) {
	var req struct {
		UserID       uuid.UUID `json:"user_id" binding:"required"`
		DocumentType string    `json:"document_type" binding:"required"`
		FileUrl      string    `json:"file_url" binding:"required"`
		FileName     string    `json:"file_name" binding:"required"`
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

	// Get admin user ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	doc := &dto.KYCDocument{
		ID:           uuid.New(),
		UserID:       req.UserID,
		DocumentType: req.DocumentType,
		FileUrl:      req.FileUrl,
		FileName:     req.FileName,
		UploadDate:   time.Now(),
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdDoc, err := h.kycStorage.CreateDocument(ctx.Request.Context(), doc)
	if err != nil {
		h.log.Error("Failed to create KYC document", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create KYC document",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "create", "kyc_document", "Created KYC document", map[string]interface{}{
		"user_id":       req.UserID,
		"document_id":   createdDoc.ID,
		"document_type": req.DocumentType,
		"created_by":    adminUserID,
	})

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    createdDoc,
		"message": "KYC document created successfully",
	})
}

// GetKYCDocuments retrieves all KYC documents for a user
func (h *KYCHandler) GetKYCDocuments(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	documents, err := h.kycStorage.GetDocumentsByUserID(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get KYC documents", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get KYC documents",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    documents,
		"total":   len(documents),
		"message": "KYC documents retrieved successfully",
	})
}

// UpdateDocumentStatus updates the status of a KYC document
func (h *KYCHandler) UpdateDocumentStatus(ctx *gin.Context) {
	var req struct {
		DocumentID      uuid.UUID `json:"document_id" binding:"required"`
		Status          string    `json:"status" binding:"required"`
		RejectionReason *string   `json:"rejection_reason,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Get admin user ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid admin user ID",
		})
		return
	}

	err := h.kycStorage.UpdateDocumentStatus(ctx.Request.Context(), req.DocumentID, req.Status, req.RejectionReason, adminUUID)
	if err != nil {
		h.log.Error("Failed to update document status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update document status",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "kyc_document", "Updated KYC document status", map[string]interface{}{
		"document_id":      req.DocumentID,
		"status":           req.Status,
		"rejection_reason": req.RejectionReason,
		"updated_by":       adminUUID,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document status updated successfully",
	})
}

// UpdateUserKYCStatus updates the overall KYC status of a user
func (h *KYCHandler) UpdateUserKYCStatus(ctx *gin.Context) {
	var req struct {
		UserID    uuid.UUID `json:"user_id" binding:"required"`
		NewStatus string    `json:"new_status" binding:"required"`
		Reason    *string   `json:"reason,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Get admin user ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid admin user ID",
		})
		return
	}

	err := h.kycStorage.UpdateUserKYCStatus(ctx.Request.Context(), req.UserID, req.NewStatus, req.Reason, adminUUID)
	if err != nil {
		h.log.Error("Failed to update user KYC status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update user KYC status",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "kyc_status", "Updated user KYC status", map[string]interface{}{
		"user_id":    req.UserID,
		"new_status": req.NewStatus,
		"reason":     req.Reason,
		"updated_by": adminUUID,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User KYC status updated successfully",
	})
}

// GetUserKYCStatus retrieves the current KYC status of a user
func (h *KYCHandler) GetUserKYCStatus(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	status, err := h.kycStorage.GetUserKYCStatus(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get user KYC status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user KYC status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":    userID,
			"kyc_status": status,
		},
		"message": "User KYC status retrieved successfully",
	})
}

// GetWithdrawalBlockStatus returns whether the user's withdrawals are currently blocked
func (h *KYCHandler) GetWithdrawalBlockStatus(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	blocked, err := h.kycStorage.IsWithdrawalBlocked(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get withdrawal block status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get withdrawal block status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":    userID,
			"is_blocked": blocked,
		},
		"message": "Withdrawal block status retrieved successfully",
	})
}

// BlockUserWithdrawals blocks withdrawals for a user
func (h *KYCHandler) BlockUserWithdrawals(ctx *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Reason string    `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Get admin user ID from context
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid admin user ID",
		})
		return
	}

	err := h.kycStorage.BlockWithdrawals(ctx.Request.Context(), req.UserID, req.Reason, adminUUID)
	if err != nil {
		h.log.Error("Failed to block withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to block withdrawals",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activitỵ
	h.logAdminActivity(ctx, "block", "withdrawals", "Blocked user withdrawals", map[string]interface{}{
		"user_id":    req.UserID,
		"reason":     req.Reason,
		"blocked_by": adminUUID,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User withdrawals blocked successfully",
	})
}

// UnblockUserWithdrawals unblocks withdrawals for a user
func (h *KYCHandler) UnblockUserWithdrawals(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = h.kycStorage.UnblockWithdrawals(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to unblock withdrawals", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unblock withdrawals",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "unblock", "withdrawals", "Unblocked user withdrawals", map[string]interface{}{
		"user_id": userID,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User withdrawals unblocked successfully",
	})
}

// GetKYCSubmissions retrieves all KYC submissions for a user
func (h *KYCHandler) GetKYCSubmissions(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	submissions, err := h.kycStorage.GetSubmissionsByUserID(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get KYC submissions", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get KYC submissions",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    submissions,
		"total":   len(submissions),
		"message": "KYC submissions retrieved successfully",
	})
}

// GetStatusChanges retrieves all status changes for a user
func (h *KYCHandler) GetStatusChanges(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	changes, err := h.kycStorage.GetStatusChangesByUserID(ctx.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get status changes", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get status changes",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    changes,
		"total":   len(changes),
		"message": "Status changes retrieved successfully",
	})
}

// GetAllSubmissions returns paginated KYC submissions, optionally filtered by status
func (h *KYCHandler) GetAllSubmissions(ctx *gin.Context) {
	status := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	perPageStr := ctx.DefaultQuery("per_page", "20")

	page, _ := strconv.Atoi(pageStr)
	perPage, _ := strconv.Atoi(perPageStr)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	subs, total, err := h.kycStorage.GetAllSubmissions(ctx.Request.Context(), statusPtr, perPage, offset)
	if err != nil {
		h.log.Error("Failed to list KYC submissions", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	if subs == nil {
		subs = []dto.KYCSubmission{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":    subs,
			"page":     page,
			"per_page": perPage,
			"total":    total,
			"total_pages": func() int64 {
				if perPage == 0 {
					return 1
				}
				t := (total + int64(perPage) - 1) / int64(perPage)
				return t
			}(),
		},
		"message": "KYC submissions listed successfully",
	})
}
