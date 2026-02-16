package groove

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/storage/groove"
	"go.uber.org/zap"
)

type PlayerTransactionHistoryHandler struct {
	storage groove.GrooveStorage
	logger  *zap.Logger
}

func NewPlayerTransactionHistoryHandler(storage groove.GrooveStorage, logger *zap.Logger) *PlayerTransactionHistoryHandler {
	return &PlayerTransactionHistoryHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetPlayerTransactionHistory godoc
// @Summary Get player transaction history
// @Description Get comprehensive transaction history for a player including GrooveTech gaming, sports betting, and general betting transactions
// @Tags Player Transactions
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param account_id query string false "GrooveTech Account ID (optional)"
// @Param type query string false "Transaction type (wager, result, rollback, bet, sport_bet)"
// @Param status query string false "Transaction status"
// @Param category query string false "Transaction category (gaming, sports, general)"
// @Param limit query int false "Limit results (default 50, max 100)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param include_summary query bool false "Include summary statistics (default false)"
// @Success 200 {object} dto.PlayerTransactionHistoryResponse
// @Success 200 {object} dto.PlayerTransactionHistoryWithSummaryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /player/transactions [get]
func (h *PlayerTransactionHistoryHandler) GetPlayerTransactionHistory(c *gin.Context) {
	// Parse and validate user ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "user_id is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "user_id must be a valid UUID",
		})
		return
	}

	// Parse optional parameters
	var accountID, transactionType, status, category *string
	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		accountID = &accountIDStr
	}
	if transactionTypeStr := c.Query("type"); transactionTypeStr != "" {
		transactionType = &transactionTypeStr
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}
	if categoryStr := c.Query("category"); categoryStr != "" {
		category = &categoryStr
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Parse date filters
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsedDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsedDate
		}
	}

	// Check if summary is requested
	includeSummary := c.Query("include_summary") == "true"

	h.logger.Info("Fetching player transaction history",
		zap.String("user_id", userID.String()),
		zap.Stringp("account_id", accountID),
		zap.Stringp("type", transactionType),
		zap.Stringp("status", status),
		zap.Stringp("category", category),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Bool("include_summary", includeSummary))

	// Fetch transactions
	transactions, err := h.storage.GetPlayerTransactionHistory(
		c.Request.Context(),
		userID,
		accountID,
		transactionType,
		status,
		category,
		startDate,
		endDate,
		limit,
		offset,
	)
	if err != nil {
		h.logger.Error("Failed to fetch player transaction history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch transaction history",
		})
		return
	}

	// Fetch total count
	total, err := h.storage.GetPlayerTransactionHistoryCount(
		c.Request.Context(),
		userID,
		accountID,
		transactionType,
		status,
		category,
		startDate,
		endDate,
	)
	if err != nil {
		h.logger.Error("Failed to fetch transaction count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch transaction count",
		})
		return
	}

	// Use transactions directly from storage (already converted to DTO)
	responseTransactions := transactions

	// Check if there are more results
	hasMore := offset+len(transactions) < total

	if includeSummary {
		// Fetch summary statistics
		summary, err := h.storage.GetPlayerTransactionHistorySummary(
			c.Request.Context(),
			userID,
			accountID,
			transactionType,
			status,
			category,
			startDate,
			endDate,
		)
		if err != nil {
			h.logger.Error("Failed to fetch transaction summary", zap.Error(err))
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to fetch transaction summary",
			})
			return
		}

		// Calculate win rate
		winRate := 0.0
		if summary.TransactionCount > 0 {
			winRate = float64(summary.WinCount) / float64(summary.TransactionCount) * 100
		}

		response := dto.PlayerTransactionHistoryWithSummaryResponse{
			Transactions: responseTransactions,
			Summary: dto.PlayerTransactionSummary{
				UserID:           summary.UserID,
				TotalWagers:      summary.TotalWagers,
				TotalWins:        summary.TotalWins,
				TotalLosses:      summary.TotalLosses,
				NetResult:        summary.NetResult,
				TransactionCount: summary.TransactionCount,
				WinCount:         summary.WinCount,
				LossCount:        summary.LossCount,
				WinRate:          winRate,
				AverageBet:       summary.AverageBet,
				MaxWin:           summary.MaxWin,
				MaxLoss:          summary.MaxLoss,
				FirstTransaction: summary.FirstTransaction,
				LastTransaction:  summary.LastTransaction,
			},
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		}

		c.JSON(http.StatusOK, response)
	} else {
		response := dto.PlayerTransactionHistoryResponse{
			Transactions: responseTransactions,
			Total:        total,
			Limit:        limit,
			Offset:       offset,
			HasMore:      hasMore,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetPlayerTransactionHistoryByAccountID godoc
// @Summary Get player transaction history by GrooveTech account ID
// @Description Get transaction history for a specific GrooveTech account
// @Tags Player Transactions
// @Accept json
// @Produce json
// @Param account_id path string true "GrooveTech Account ID"
// @Param type query string false "Transaction type (wager, result, rollback)"
// @Param status query string false "Transaction status"
// @Param limit query int false "Limit results (default 50, max 100)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.PlayerTransactionHistoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /player/transactions/account/{account_id} [get]
func (h *PlayerTransactionHistoryHandler) GetPlayerTransactionHistoryByAccountID(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "account_id is required",
		})
		return
	}

	// Parse optional parameters
	var transactionType, status *string
	if transactionTypeStr := c.Query("type"); transactionTypeStr != "" {
		transactionType = &transactionTypeStr
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Parse date filters
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsedDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsedDate
		}
	}

	h.logger.Info("Fetching transaction history by account ID",
		zap.String("account_id", accountID),
		zap.Stringp("type", transactionType),
		zap.Stringp("status", status),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Fetch transactions
	transactions, err := h.storage.GetPlayerTransactionHistoryByAccountID(
		c.Request.Context(),
		accountID,
		transactionType,
		status,
		startDate,
		endDate,
		limit,
		offset,
	)
	if err != nil {
		h.logger.Error("Failed to fetch transaction history by account ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch transaction history",
		})
		return
	}

	// Fetch total count
	total, err := h.storage.GetPlayerTransactionHistoryByAccountIDCount(
		c.Request.Context(),
		accountID,
		transactionType,
		status,
		startDate,
		endDate,
	)
	if err != nil {
		h.logger.Error("Failed to fetch transaction count by account ID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch transaction count",
		})
		return
	}

	// Use transactions directly from storage (already converted to DTO)
	responseTransactions := transactions

	// Check if there are more results
	hasMore := offset+len(transactions) < total

	response := dto.PlayerTransactionHistoryResponse{
		Transactions: responseTransactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		HasMore:      hasMore,
	}

	c.JSON(http.StatusOK, response)
}
