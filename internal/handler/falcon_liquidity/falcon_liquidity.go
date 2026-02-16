package falcon_liquidity

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/storage/falcon_liquidity"
	"go.uber.org/zap"
)

type FalconLiquidityHandler struct {
	storage falcon_liquidity.FalconMessageStorage
	logger  *zap.Logger
}

func NewFalconLiquidityHandler(storage falcon_liquidity.FalconMessageStorage, logger *zap.Logger) *FalconLiquidityHandler {
	return &FalconLiquidityHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetAllFalconLiquidityData retrieves all Falcon Liquidity data without authentication
//
//	@Summary		Get All Falcon Liquidity Data
//	@Description	Retrieve all Falcon Liquidity messages and data without authentication
//	@Tags			Falcon Liquidity
//	@Produce		json
//	@Param			page			query		int		false	"Page number (default: 1)"
//	@Param			per_page		query		int		false	"Items per page (default: 50, max: 100)"
//	@Param			message_type	query		string	false	"Filter by message type (casino, sport)"
//	@Param			status			query		string	false	"Filter by status (pending, sent, failed, acknowledged)"
//	@Param			transaction_id	query		string	false	"Filter by transaction ID"
//	@Param			user_id			query		string	false	"Filter by user ID"
//	@Param			message_id		query		string	false	"Filter by message ID"
//	@Param			date_from		query		string	false	"Filter by start date (YYYY-MM-DD)"
//	@Param			date_to			query		string	false	"Filter by end date (YYYY-MM-DD)"
//	@Param			reconciliation_status	query	string	false	"Filter by reconciliation status"
//	@Success		200				{object}	dto.FalconLiquidityDataResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/falcon-liquidity/data [get]
func (h *FalconLiquidityHandler) GetAllFalconLiquidityData(c *gin.Context) {
	// Parse pagination parameters (page-based)
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 {
		perPage = 50
	}
	if perPage > 100 {
		perPage = 100
	}

	// Convert page/per_page to limit/offset
	limit := perPage
	offset := (page - 1) * perPage

	// Parse filter parameters
	messageType := c.Query("message_type")
	status := c.Query("status")
	transactionID := c.Query("transaction_id")
	userIDStr := c.Query("user_id")
	messageID := c.Query("message_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	reconciliationStatus := c.Query("reconciliation_status")

	// Build query
	query := dto.FalconMessageQuery{
		Limit:  limit,
		Offset: offset,
	}

	// Add optional filters only if they have values
	if messageType != "" {
		msgType := dto.FalconMessageType(messageType)
		query.MessageType = &msgType
	}
	if status != "" {
		statusVal := dto.FalconMessageStatus(status)
		query.Status = &statusVal
	}
	if transactionID != "" {
		query.TransactionID = &transactionID
	}
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			query.UserID = &userID
		}
	}
	if messageID != "" {
		query.MessageID = &messageID
	}
	if dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			query.StartDate = &dateFrom
		}
	}
	if dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			// Add one day to include the entire end date
			dateTo = dateTo.Add(24 * time.Hour).Add(-1 * time.Second)
			query.EndDate = &dateTo
		}
	}
	if reconciliationStatus != "" {
		reconStatus := dto.FalconReconciliationStatus(reconciliationStatus)
		query.ReconciliationStatus = &reconStatus
	}

	// Get messages
	messages, err := h.storage.QueryFalconMessages(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query Falcon Liquidity messages",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset))

		err := errors.ErrInternalServerError.Wrap(err, "Failed to retrieve Falcon Liquidity data")
		_ = c.Error(err)
		return
	}

	// Get summary statistics
	summary, err := h.storage.GetFalconMessageSummary(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get Falcon Liquidity summary",
			zap.Error(err))

		// Don't fail the request, just log the error
		h.logger.Warn("Continuing without summary statistics due to error")
		summary = &dto.FalconMessageSummary{}
	}

	// Get total count from summary
	totalCount := summary.TotalMessages

	// Prepare response
	responseData := dto.FalconLiquidityDataResponse{
		Messages: messages,
		Summary:  summary,
		Pagination: dto.FalconLiquidityPagination{
			Limit:  limit,
			Offset: offset,
			Total:  totalCount, // Use total from summary, not len(messages)
		},
	}

	h.logger.Info("Successfully retrieved Falcon Liquidity data",
		zap.Int("message_count", len(messages)),
		zap.Int("total_count", totalCount),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	response.SendSuccessResponse(c, http.StatusOK, responseData)
}

// GetFalconLiquidityByTransactionID retrieves Falcon Liquidity data by transaction ID
//
//	@Summary		Get Falcon Liquidity by Transaction ID
//	@Description	Retrieve Falcon Liquidity messages by specific transaction ID
//	@Tags			Falcon Liquidity
//	@Produce		json
//	@Param			transaction_id	path		string	true	"Transaction ID"
//	@Success		200				{object}	dto.FalconLiquidityDataResponse
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		404				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/admin/falcon-liquidity/transaction/{transaction_id} [get]
func (h *FalconLiquidityHandler) GetFalconLiquidityByTransactionID(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		err := errors.ErrInvalidUserInput.New("Transaction ID is required")
		_ = c.Error(err)
		return
	}

	// Get messages by transaction ID
	messages, err := h.storage.GetFalconMessagesByTransactionID(c.Request.Context(), transactionID)
	if err != nil {
		h.logger.Error("Failed to get Falcon Liquidity messages by transaction ID",
			zap.Error(err),
			zap.String("transaction_id", transactionID))

		err := errors.ErrInternalServerError.Wrap(err, "Failed to retrieve Falcon Liquidity data")
		_ = c.Error(err)
		return
	}

	if len(messages) == 0 {
		err := errors.ErrInvalidUserInput.New("No Falcon Liquidity data found for transaction ID")
		_ = c.Error(err)
		return
	}

	// Prepare response
	responseData := dto.FalconLiquidityDataResponse{
		Messages: messages,
		Summary: &dto.FalconMessageSummary{
			TotalMessages: len(messages),
		},
		Pagination: dto.FalconLiquidityPagination{
			Total: len(messages),
		},
	}

	h.logger.Info("Successfully retrieved Falcon Liquidity data by transaction ID",
		zap.String("transaction_id", transactionID),
		zap.Int("message_count", len(messages)))

	response.SendSuccessResponse(c, http.StatusOK, responseData)
}

// GetFalconLiquidityByUserID retrieves Falcon Liquidity data by user ID
//
//	@Summary		Get Falcon Liquidity by User ID
//	@Description	Retrieve Falcon Liquidity messages by specific user ID
//	@Tags			Falcon Liquidity
//	@Produce		json
//	@Param			user_id	path		string	true	"User ID"
//	@Param			limit	query		int		false	"Limit number of results (default: 100, max: 1000)"
//	@Param			offset	query		int		false	"Offset for pagination (default: 0)"
//	@Success		200		{object}	dto.FalconLiquidityDataResponse
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/admin/falcon-liquidity/user/{user_id} [get]
func (h *FalconLiquidityHandler) GetFalconLiquidityByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		err := errors.ErrInvalidUserInput.New("User ID is required")
		_ = c.Error(err)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid user ID format")
		_ = c.Error(err)
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get messages by user ID
	messages, err := h.storage.GetFalconMessagesByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get Falcon Liquidity messages by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()))

		err := errors.ErrInternalServerError.Wrap(err, "Failed to retrieve Falcon Liquidity data")
		_ = c.Error(err)
		return
	}

	if len(messages) == 0 {
		err := errors.ErrInvalidUserInput.New("No Falcon Liquidity data found for user ID")
		_ = c.Error(err)
		return
	}

	// Prepare response
	responseData := dto.FalconLiquidityDataResponse{
		Messages: messages,
		Summary: &dto.FalconMessageSummary{
			TotalMessages: len(messages),
		},
		Pagination: dto.FalconLiquidityPagination{
			Limit:  limit,
			Offset: offset,
			Total:  len(messages),
		},
	}

	h.logger.Info("Successfully retrieved Falcon Liquidity data by user ID",
		zap.String("user_id", userID.String()),
		zap.Int("message_count", len(messages)))

	response.SendSuccessResponse(c, http.StatusOK, responseData)
}

// GetFalconLiquiditySummary retrieves summary statistics for Falcon Liquidity data
//
//	@Summary		Get Falcon Liquidity Summary
//	@Description	Retrieve summary statistics for Falcon Liquidity data
//	@Tags			Falcon Liquidity
//	@Produce		json
//	@Param			message_type	query		string	false	"Filter by message type"
//	@Param			status		query		string	false	"Filter by status"
//	@Param			transaction_id	query	string	false	"Filter by transaction ID"
//	@Success		200			{object}	dto.FalconMessageSummary
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/api/admin/falcon-liquidity/summary [get]
func (h *FalconLiquidityHandler) GetFalconLiquiditySummary(c *gin.Context) {
	// Parse query parameters
	messageType := c.Query("message_type")
	status := c.Query("status")
	transactionID := c.Query("transaction_id")

	// Build query
	query := dto.FalconMessageQuery{}

	// Add optional filters only if they have values
	if messageType != "" {
		msgType := dto.FalconMessageType(messageType)
		query.MessageType = &msgType
	}
	if status != "" {
		statusVal := dto.FalconMessageStatus(status)
		query.Status = &statusVal
	}
	if transactionID != "" {
		query.TransactionID = &transactionID
	}

	// Get summary
	summary, err := h.storage.GetFalconMessageSummary(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get Falcon Liquidity summary",
			zap.Error(err))

		err := errors.ErrInternalServerError.Wrap(err, "Failed to retrieve Falcon Liquidity summary")
		_ = c.Error(err)
		return
	}

	h.logger.Info("Successfully retrieved Falcon Liquidity summary",
		zap.String("message_type", messageType),
		zap.String("status", status),
		zap.String("transaction_id", transactionID))

	response.SendSuccessResponse(c, http.StatusOK, summary)
}
