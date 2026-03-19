package kyc

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
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

// UploadKYCDocument handles file upload for KYC documents:
// - saves the file to a configured folder on disk
// - stores only the file path (URL) and name in kyc_documents.
func (h *KYCHandler) UploadKYCDocument(ctx *gin.Context) {
	userIDStr := ctx.PostForm("user_id")
	documentType := ctx.PostForm("document_type")

	if userIDStr == "" || documentType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "user_id and document_type are required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("Invalid user_id", zap.Error(err), zap.String("user_id", userIDStr))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user_id",
		})
		return
	}

	// Retrieve uploaded file
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		h.log.Error("Failed to read uploaded file", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "file is required",
		})
		return
	}
	defer file.Close()

	// Determine base upload directory (configurable)
	baseDir := viper.GetString("kyc.upload_dir")
	if baseDir == "" {
		baseDir = "./uploads/kyc"
	}

	// Build user-specific folder and destination path
	userDir := filepath.Join(baseDir, userID.String())
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		h.log.Error("Failed to create KYC upload directory", zap.Error(err), zap.String("dir", userDir))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to prepare upload directory",
		})
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = documentType + "_" + time.Now().Format("20060102150405")
	}
	destPath := filepath.Join(userDir, time.Now().Format("20060102150405_")+filename)

	// Save file to disk
	if err := ctx.SaveUploadedFile(header, destPath); err != nil {
		h.log.Error("Failed to save uploaded file", zap.Error(err), zap.String("path", destPath))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save uploaded file",
		})
		return
	}

	fileURL := destPath
	now := time.Now()
	doc := &dto.KYCDocument{
		ID:           uuid.New(),
		UserID:       userID,
		DocumentType: documentType,
		FileUrl:      fileURL,
		FileName:     filename,
		UploadDate:   now,
		Status:       dto.DocumentStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdDoc, err := h.kycStorage.CreateDocument(ctx.Request.Context(), doc)
	if err != nil {
		h.log.Error("Failed to create KYC document", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create KYC document",
		})
		return
	}

	// Log admin activity (if admin in context)
	if adminUserID, exists := ctx.Get("user_id"); exists {
		h.logAdminActivity(ctx, "upload", "kyc_document", "Uploaded KYC document", map[string]interface{}{
			"user_id":       userID,
			"document_id":   createdDoc.ID,
			"document_type": documentType,
			"file_url":      fileURL,
			"created_by":    adminUserID,
		})
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    createdDoc,
		"message": "KYC document uploaded successfully",
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

// GetOperatorKYCDocuments retrieves KYC documents for all users under an operator.
func (h *KYCHandler) GetOperatorKYCDocuments(ctx *gin.Context) {
	operatorIDStr := ctx.Param("operator_id")
	operatorID64, err := strconv.ParseInt(operatorIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid operator ID",
		})
		return
	}

	limit := 20
	offset := 0
	if v := ctx.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := ctx.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	documents, total, err := h.kycStorage.GetDocumentsByOperatorID(ctx.Request.Context(), int32(operatorID64), limit, offset)
	if err != nil {
		h.log.Error("Failed to get operator KYC documents", zap.Error(err), zap.Int32("operator_id", int32(operatorID64)))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get operator KYC documents",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    documents,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"message": "Operator KYC documents retrieved successfully",
	})
}

// UploadOperatorKYCDocument uploads a document for an operator (entity-level KYC).
func (h *KYCHandler) UploadOperatorKYCDocument(ctx *gin.Context) {
	operatorIDStr := ctx.Param("operator_id")
	operatorID64, err := strconv.ParseInt(operatorIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid operator ID"})
		return
	}

	documentType := ctx.PostForm("document_type")
	if documentType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "document_type is required"})
		return
	}

	_, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "file is required"})
		return
	}

	baseDir := viper.GetString("kyc.operator_upload_dir")
	if baseDir == "" {
		baseDir = "./uploads/operator_kyc"
	}
	operatorDir := filepath.Join(baseDir, operatorIDStr)
	if err := os.MkdirAll(operatorDir, 0o755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to prepare upload directory"})
		return
	}
	filename := header.Filename
	if filename == "" {
		filename = documentType + "_" + time.Now().Format("20060102150405")
	}
	destPath := filepath.Join(operatorDir, time.Now().Format("20060102150405_")+filename)
	if err := ctx.SaveUploadedFile(header, destPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save uploaded file"})
		return
	}

	now := time.Now()
	doc := &dto.OperatorKYCDocument{
		ID:           uuid.New(),
		OperatorID:   int32(operatorID64),
		DocumentType: documentType,
		FileURL:      destPath,
		FileName:     filename,
		UploadDate:   now,
		Status:       dto.DocumentStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdDoc, err := h.kycStorage.CreateOperatorDocument(ctx.Request.Context(), doc)
	if err != nil {
		h.log.Error("Failed to create operator KYC document", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create operator KYC document"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    createdDoc,
		"message": "Operator KYC document uploaded successfully",
	})
}

func (h *KYCHandler) UpdateOperatorDocumentStatus(ctx *gin.Context) {
	var req struct {
		DocumentID      uuid.UUID `json:"document_id" binding:"required"`
		Status          string    `json:"status" binding:"required"`
		RejectionReason *string   `json:"rejection_reason,omitempty"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body"})
		return
	}
	adminUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	adminUUID, ok := adminUserID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid admin user ID"})
		return
	}
	if err := h.kycStorage.UpdateOperatorDocumentStatus(ctx.Request.Context(), req.DocumentID, req.Status, req.RejectionReason, adminUUID); err != nil {
		h.log.Error("Failed to update operator KYC document status", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update operator KYC document status", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Operator document status updated successfully"})
}

func (h *KYCHandler) GetOperatorKYCSubmissions(ctx *gin.Context) {
	operatorIDStr := ctx.Param("operator_id")
	operatorID64, err := strconv.ParseInt(operatorIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid operator ID"})
		return
	}
	limit := 20
	offset := 0
	if v := ctx.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := ctx.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	items, total, err := h.kycStorage.GetOperatorSubmissions(ctx.Request.Context(), int32(operatorID64), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get operator KYC submissions", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"message": "Operator KYC submissions retrieved successfully",
	})
}

// DownloadOperatorKYCDocument streams an operator KYC document file to the admin UI.
func (h *KYCHandler) DownloadOperatorKYCDocument(ctx *gin.Context) {
	operatorIDStr := ctx.Param("operator_id")
	documentIDStr := ctx.Param("document_id")

	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid document_id"})
		return
	}

	doc, err := h.kycStorage.GetOperatorDocumentByID(ctx.Request.Context(), documentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Document not found"})
			return
		}
		h.log.Error("Failed to fetch operator KYC document", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch document"})
		return
	}

	// Optional safety: ensure operator_id in path matches record
	if operatorIDStr != "" && operatorIDStr != strconv.Itoa(int(doc.OperatorID)) {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Operator mismatch for this document"})
		return
	}

	// FileURL currently stores full relative path (e.g. ./uploads/operator_kyc/100001/..pdf)
	if doc.FileURL == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "File path not set for this document"})
		return
	}

	// Let Gin handle content-type; attachment download
	ctx.FileAttachment(doc.FileURL, doc.FileName)
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

// GetKYCSettings retrieves all KYC settings
func (h *KYCHandler) GetKYCSettings(ctx *gin.Context) {
	settings, err := h.kycStorage.GetAllKYCSettings(ctx.Request.Context())
	if err != nil {
		h.log.Error("Failed to get KYC settings", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get KYC settings",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"total":   len(settings),
		"message": "KYC settings retrieved successfully",
	})
}

// UpdateKYCSettings updates a KYC setting
func (h *KYCHandler) UpdateKYCSettings(ctx *gin.Context) {
	var req struct {
		ID           string                 `json:"id" binding:"required"`
		SettingValue map[string]interface{} `json:"setting_value" binding:"required"`
		Description  *string                `json:"description,omitempty"`
		IsActive     bool                   `json:"is_active"`
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

	settingID, err := uuid.Parse(req.ID)
	if err != nil {
		h.log.Error("Invalid setting ID", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid setting ID",
			"error":   err.Error(),
		})
		return
	}

	err = h.kycStorage.UpdateKYCSettings(ctx.Request.Context(), settingID, req.SettingValue, req.Description, req.IsActive)
	if err != nil {
		h.log.Error("Failed to update KYC setting", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update KYC setting",
			"error":   err.Error(),
		})
		return
	}

	// Log admin activity
	h.logAdminActivity(ctx, "update", "kyc_settings", "Updated KYC setting", map[string]interface{}{
		"setting_id":  req.ID,
		"setting_key": req.SettingValue,
		"is_active":   req.IsActive,
	})

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "KYC setting updated successfully",
	})
}
