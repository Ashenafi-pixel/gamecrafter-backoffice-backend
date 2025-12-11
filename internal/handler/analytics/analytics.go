package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/handler"
	analyticsModule "github.com/tucanbit/internal/module/analytics"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type analytics struct {
	logger                    *zap.Logger
	analyticsStorage          storage.Analytics
	dailyReportService        analyticsModule.DailyReportService
	dailyReportCronjobService analyticsModule.DailyReportCronjobService
	pgPool                    *pgxpool.Pool
}

// checkAnalyticsStorage checks if analytics storage is available
func (a *analytics) checkAnalyticsStorage(c *gin.Context) bool {
	if a.analyticsStorage == nil {
		a.logger.Error("Analytics storage is not available - ClickHouse client not initialized")
		c.JSON(http.StatusServiceUnavailable, dto.AnalyticsResponse{
			Success: false,
			Error:   "Analytics service is not available - ClickHouse database is not running. Please contact the administrator to start the ClickHouse service.",
		})
		return false
	}
	return true
}

// getUserBrandID fetches the brand_id for a given user from PostgreSQL.
func (a *analytics) getUserBrandID(ctx context.Context, userID uuid.UUID) (*string, error) {
	if a.pgPool == nil {
		return nil, nil
	}
	row := a.pgPool.QueryRow(ctx, "SELECT brand_id FROM users WHERE id = $1", userID)
	var brandID uuid.UUID
	if err := row.Scan(&brandID); err != nil {
		return nil, err
	}
	brandStr := brandID.String()
	return &brandStr, nil
}

// getUsernameByID fetches the username for a given user ID from PostgreSQL.
func (a *analytics) getUsernameByID(ctx context.Context, userID uuid.UUID) *string {
	if a.pgPool == nil {
		return nil
	}
	row := a.pgPool.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", userID)
	var username string
	if err := row.Scan(&username); err != nil {
		return nil
	}
	return &username
}

// getUserIDsByTestAccount gets user IDs filtered by is_test_account from PostgreSQL
func (a *analytics) getUserIDsByTestAccount(ctx context.Context, isTestAccount *bool) ([]uuid.UUID, error) {
	if isTestAccount == nil {
		// Return nil to indicate no filtering (all users)
		return nil, nil
	}

	query := "SELECT id FROM users WHERE is_test_account = $1"
	rows, err := a.pgPool.Query(ctx, query, *isTestAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return userIDs, nil
}

func Init(log *zap.Logger, analyticsStorage storage.Analytics, dailyReportService analyticsModule.DailyReportService, dailyReportCronjobService analyticsModule.DailyReportCronjobService, pgPool *pgxpool.Pool) handler.Analytics {
	return &analytics{
		logger:                    log,
		analyticsStorage:          analyticsStorage,
		dailyReportService:        dailyReportService,
		dailyReportCronjobService: dailyReportCronjobService,
		pgPool:                    pgPool,
	}
}

// GetUserTransactions Get user transactions with filters
// @Summary Get user transactions
// @Description Retrieve user transactions with optional filters
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param date_from query string false "Start date (RFC3339)"
// @Param date_to query string false "End date (RFC3339)"
// @Param transaction_type query string false "Transaction type"
// @Param game_id query string false "Game ID"
// @Param status query string false "Transaction status"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/users/{user_id}/transactions [get]
func (a *analytics) GetUserTransactions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.TransactionFilters{}

	// Parse query parameters
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	// Default transaction_type to "groove_bet" if not provided or empty
	transactionType := c.Query("transaction_type")
	if transactionType == "" {
		transactionType = "groove_bet"
	}
	filters.TransactionType = &transactionType

	if gameID := c.Query("game_id"); gameID != "" {
		filters.GameID = &gameID
	}

	// Use the provided status if given, otherwise fetch all statuses
	if statusParam := c.Query("status"); statusParam != "" {
		filters.Status = &statusParam
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Brand isolation: get brand_id for this user from Postgres and pass into ClickHouse filter
	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	transactions, total, err := a.analyticsStorage.GetUserTransactions(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user transactions",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to retrieve transactions: %v", err),
		})
		return
	}

	// Get transaction totals (total_bet_amount and total_win_amount) for meta
	var totalBetAmountStr, totalWinAmountStr *string
	totals, totalsErr := a.analyticsStorage.GetUserTransactionsTotals(c.Request.Context(), userID, filters)
	if totalsErr == nil && totals != nil {
		totalBetStr := totals.TotalBetAmount.String()
		totalWinStr := totals.TotalWinAmount.String()
		totalBetAmountStr = &totalBetStr
		totalWinAmountStr = &totalWinStr
	}

	// Calculate pagination meta to match BACKOFFICE_PLAYER_ANALYTICS_ENDPOINTS.md
	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = total
	}
	page := 1
	if pageSize > 0 {
		page = (filters.Offset / pageSize) + 1
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    transactions,
		Meta: &dto.Meta{
			Total:          total,
			Page:           page,
			PageSize:       pageSize,
			Pages:          pages,
			TotalBetAmount: totalBetAmountStr,
			TotalWinAmount: totalWinAmountStr,
		},
	})
}

// GetUserRakebackTransactions implements GET /analytics/users/{user_id}/rakeback
func (a *analytics) GetUserRakebackTransactions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	// Default transaction_type is "earned" if not provided
	transactionType := c.DefaultQuery("transaction_type", "earned")
	filters := &dto.RakebackFilters{
		TransactionType: transactionType,
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 22 // default from doc for rakeback
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Brand isolation via Postgres
	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	rows, total, totals, err := a.analyticsStorage.GetUserRakebackTransactions(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user rakeback transactions",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve rakeback transactions",
		})
		return
	}

	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = total
	}
	page := 1
	if pageSize > 0 {
		page = (filters.Offset / pageSize) + 1
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	// meta.total_claimed_amount is part of the markdown spec; we expose totals alongside standard meta.
	totalClaimedAmountStr := totals.TotalClaimedAmount.String()
	meta := &dto.Meta{
		Total:              total,
		Page:               page,
		PageSize:           pageSize,
		Pages:              pages,
		TotalClaimedAmount: &totalClaimedAmountStr,
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    rows,
		Meta:    meta,
	})
}

// GetUserTips implements GET /analytics/users/{account_id}/tips
func (a *analytics) GetUserTips(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.TipFilters{}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Brand isolation via Postgres
	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	rows, total, err := a.analyticsStorage.GetUserTips(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user tips",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve tip transactions",
		})
		return
	}

	// Parse metadata and resolve usernames from PostgreSQL
	if a.pgPool != nil {
		for _, tip := range rows {
			if tip.Metadata != nil && *tip.Metadata != "" {
				var metadata map[string]interface{}
				if err := json.Unmarshal([]byte(*tip.Metadata), &metadata); err == nil {
					if tip.TransactionType == "tip_sent" {
						// For tip_sent: sender is current user, receiver is in metadata
						if recipientID, ok := metadata["recipient_id"].(string); ok && recipientID != "" {
							if receiverID, err := uuid.Parse(recipientID); err == nil {
								if username := a.getUsernameByID(c.Request.Context(), receiverID); username != nil {
									tip.ReceiverUsername = username
								}
							}
						} else if receiverID, ok := metadata["receiver_id"].(string); ok && receiverID != "" {
							if receiverUUID, err := uuid.Parse(receiverID); err == nil {
								if username := a.getUsernameByID(c.Request.Context(), receiverUUID); username != nil {
									tip.ReceiverUsername = username
								}
							}
						}
						// Set sender username to current user
						if senderUsername := a.getUsernameByID(c.Request.Context(), userID); senderUsername != nil {
							tip.SenderUsername = senderUsername
						}
					} else if tip.TransactionType == "tip_received" {
						// For tip_received: receiver is current user, sender is in metadata
						if senderID, ok := metadata["sender_id"].(string); ok && senderID != "" {
							if senderUUID, err := uuid.Parse(senderID); err == nil {
								if username := a.getUsernameByID(c.Request.Context(), senderUUID); username != nil {
									tip.SenderUsername = username
								}
							}
						}
						// Set receiver username to current user
						if receiverUsername := a.getUsernameByID(c.Request.Context(), userID); receiverUsername != nil {
							tip.ReceiverUsername = receiverUsername
						}
					}
				}
			}
		}
	}

	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = total
	}
	page := 1
	if pageSize > 0 {
		page = (filters.Offset / pageSize) + 1
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	meta := &dto.Meta{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    rows,
		Meta:    meta,
	})
}

// GetWelcomeBonusTransactions implements GET /api/admin/analytics/welcome_bonus - Admin only endpoint to get all welcome bonuses with filters
func (a *analytics) GetWelcomeBonusTransactions(c *gin.Context) {
	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.WelcomeBonusFilters{}

	// Optional user_id filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filters.UserID = &userID
		}
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	// Amount filters
	if minAmountStr := c.Query("min_amount"); minAmountStr != "" {
		if minAmount, err := decimal.NewFromString(minAmountStr); err == nil {
			filters.MinAmount = &minAmount
		}
	}
	if maxAmountStr := c.Query("max_amount"); maxAmountStr != "" {
		if maxAmount, err := decimal.NewFromString(maxAmountStr); err == nil {
			filters.MaxAmount = &maxAmount
		}
	}

	// Brand filter
	if brandIDStr := c.Query("brand_id"); brandIDStr != "" {
		filters.BrandID = &brandIDStr
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	rows, total, err := a.analyticsStorage.GetWelcomeBonusTransactions(c.Request.Context(), filters)
	if err != nil {
		a.logger.Error("Failed to get welcome bonus transactions",
			zap.Error(err),
			zap.Any("filters", filters))
		errorMsg := err.Error()
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   errorMsg,
		})
		return
	}

	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = total
	}
	page := 1
	if pageSize > 0 {
		page = (filters.Offset / pageSize) + 1
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	meta := &dto.Meta{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    rows,
		Meta:    meta,
	})
}

// GetUserWelcomeBonus implements GET /api/admin/analytics/users/{user_id}/welcome_bonus - Get welcome bonuses for a specific user
func (a *analytics) GetUserWelcomeBonus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.WelcomeBonusFilters{}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Brand isolation via Postgres
	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	rows, total, err := a.analyticsStorage.GetUserWelcomeBonus(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user welcome bonus transactions",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.Any("filters", filters))
		errorMsg := err.Error()
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   errorMsg,
		})
		return
	}

	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = total
	}
	page := 1
	if pageSize > 0 {
		page = (filters.Offset / pageSize) + 1
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}

	meta := &dto.Meta{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    rows,
		Meta:    meta,
	})
}

// GetUserTransactionsTotals implements /api/admin/analytics/users/{user_id}/transactions/totals
func (a *analytics) GetUserTransactionsTotals(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.TransactionFilters{}
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}
	if transactionType := c.Query("transaction_type"); transactionType != "" {
		filters.TransactionType = &transactionType
	}
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	totals, err := a.analyticsStorage.GetUserTransactionsTotals(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user transaction totals",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve transaction totals",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    totals,
	})
}

// GetUserRakebackTotals implements /api/admin/analytics/users/{user_id}/rakeback/totals
func (a *analytics) GetUserRakebackTotals(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	transactionType := c.DefaultQuery("transaction_type", "earned")
	filters := &dto.RakebackFilters{
		TransactionType: transactionType,
	}
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	totals, err := a.analyticsStorage.GetUserRakebackTotals(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user rakeback totals",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve rakeback totals",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    totals,
	})
}

// GetUserTipsTotals implements /api/admin/analytics/users/{user_id}/tips/totals
func (a *analytics) GetUserTipsTotals(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.TipFilters{}
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	totals, err := a.analyticsStorage.GetUserTipsTotals(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user tips totals",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve tips totals",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    totals,
	})
}

// GetUserWelcomeBonusTotals implements /api/admin/analytics/users/{user_id}/welcome_bonus/totals
func (a *analytics) GetUserWelcomeBonusTotals(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.WelcomeBonusFilters{}
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	if brandID, err := a.getUserBrandID(c.Request.Context(), userID); err == nil && brandID != nil {
		filters.BrandID = brandID
	}

	totals, err := a.analyticsStorage.GetUserWelcomeBonusTotals(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user welcome bonus totals",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve welcome bonus totals",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    totals,
	})
}

// GetUserAnalytics Get user analytics
// @Summary Get user analytics
// @Description Retrieve comprehensive analytics for a specific user
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param date_from query string false "Start date (RFC3339)"
// @Param date_to query string false "End date (RFC3339)"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/users/{user_id}/analytics [get]
func (a *analytics) GetUserAnalytics(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	dateRange := &dto.DateRange{}

	// Parse date range
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			dateRange.From = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			dateRange.To = &dateTo
		}
	}

	analytics, err := a.analyticsStorage.GetUserAnalytics(c.Request.Context(), userID, dateRange)
	if err != nil {
		a.logger.Error("Failed to get user analytics",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve analytics",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    analytics,
	})
}

// GetRealTimeStats Get real-time statistics
// @Summary Get real-time statistics
// @Description Retrieve real-time statistics for the last hour
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/realtime [get]
func (a *analytics) GetRealTimeStats(c *gin.Context) {
	a.logger.Info("GetRealTimeStats handler called")

	if a.analyticsStorage == nil {
		a.logger.Error("Analytics storage is nil - ClickHouse client not initialized")
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Analytics service not available - ClickHouse not initialized",
		})
		return
	}

	stats, err := a.analyticsStorage.GetRealTimeStats(c.Request.Context())
	if err != nil {
		a.logger.Error("Failed to get real-time stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve real-time statistics",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetDailyReport Get daily report
// @Summary Get daily report
// @Description Retrieve daily analytics report for a specific date
// @Tags analytics
// @Accept json
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/reports/daily [get]
func (a *analytics) GetDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Date parameter is required",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	report, err := a.analyticsStorage.GetDailyReport(c.Request.Context(), date)
	if err != nil {
		a.logger.Error("Failed to get daily report",
			zap.String("date", dateStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve daily report",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    report,
	})
}

// GetEnhancedDailyReport Get enhanced daily report with comparison metrics
// @Summary Get enhanced daily report
// @Description Retrieve enhanced daily analytics report with comparison metrics (previous day, MTD, SPLM)
// @Tags analytics
// @Accept json
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/reports/daily-enhanced [get]
func (a *analytics) GetEnhancedDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Date parameter is required",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	report, err := a.analyticsStorage.GetEnhancedDailyReport(c.Request.Context(), date)
	if err != nil {
		a.logger.Error("Failed to get enhanced daily report",
			zap.String("date", dateStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve enhanced daily report",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    report,
	})
}

// GetTopGames Get top games
// @Summary Get top games
// @Description Retrieve top performing games
// @Tags analytics
// @Accept json
// @Produce json
// @Param limit query int false "Number of games to return" default(10)
// @Param date_from query string false "Start date (YYYY-MM-DD format, e.g., 2025-11-16)"
// @Param date_to query string false "End date (YYYY-MM-DD format, e.g., 2025-11-23)"
// @Param is_test_account query bool false "Filter by test account (true/false), omit for all"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /api/admin/analytics/reports/top-games [get]
func (a *analytics) GetTopGames(c *gin.Context) {
	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	dateRange := &dto.DateRange{}

	// Parse date range (using "2006-01-02" format)
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateRange.From = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateRange.To = &dateTo
		}
	}

	// Parse is_test_account filter
	if isTestAccountStr := c.Query("is_test_account"); isTestAccountStr != "" {
		if isTestAccount, err := strconv.ParseBool(isTestAccountStr); err == nil {
			dateRange.IsTestAccount = &isTestAccount
			// Get filtered user IDs
			userIDs, err := a.getUserIDsByTestAccount(c.Request.Context(), &isTestAccount)
			if err != nil {
				a.logger.Error("Failed to get filtered user IDs", zap.Error(err))
				c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
					Success: false,
					Error:   "Failed to filter users",
				})
				return
			}
			dateRange.UserIDs = userIDs
		}
	}

	games, err := a.analyticsStorage.GetTopGames(c.Request.Context(), limit, dateRange)
	if err != nil {
		a.logger.Error("Failed to get top games", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve top games",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    games,
	})
}

// GetTopPlayers Get top players
// @Summary Get top players
// @Description Retrieve top players by various metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param limit query int false "Number of players to return" default(10)
// @Param date_from query string false "Start date (YYYY-MM-DD format, e.g., 2025-11-16)"
// @Param date_to query string false "End date (YYYY-MM-DD format, e.g., 2025-11-23)"
// @Param is_test_account query bool false "Filter by test account (true/false), omit for all"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /api/admin/analytics/reports/top-players [get]
func (a *analytics) GetTopPlayers(c *gin.Context) {
	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	dateRange := &dto.DateRange{}

	// Parse date range (using "2006-01-02" format)
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateRange.From = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateRange.To = &dateTo
		}
	}

	// Parse is_test_account filter
	if isTestAccountStr := c.Query("is_test_account"); isTestAccountStr != "" {
		if isTestAccount, err := strconv.ParseBool(isTestAccountStr); err == nil {
			dateRange.IsTestAccount = &isTestAccount
			// Get filtered user IDs
			userIDs, err := a.getUserIDsByTestAccount(c.Request.Context(), &isTestAccount)
			if err != nil {
				a.logger.Error("Failed to get filtered user IDs", zap.Error(err))
				c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
					Success: false,
					Error:   "Failed to filter users",
				})
				return
			}
			dateRange.UserIDs = userIDs
		}
	}

	players, err := a.analyticsStorage.GetTopPlayers(c.Request.Context(), limit, dateRange)
	if err != nil {
		a.logger.Error("Failed to get top players",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Any("dateRange", dateRange),
			zap.Int("userIDs_count", len(dateRange.UserIDs)))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to retrieve top players: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    players,
	})
}

// GetUserBalanceHistory Get user balance history
// @Summary Get user balance history
// @Description Retrieve user balance history for the last N hours
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param hours query int false "Number of hours to look back" default(24)
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/users/{user_id}/balance-history [get]
func (a *analytics) GetUserBalanceHistory(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.AnalyticsResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	hours := 24 // Default hours
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if parsedHours, err := strconv.Atoi(hoursStr); err == nil && parsedHours > 0 {
			hours = parsedHours
		}
	}

	history, err := a.analyticsStorage.GetUserBalanceHistory(c.Request.Context(), userID, hours)
	if err != nil {
		a.logger.Error("Failed to get user balance history",
			zap.String("user_id", userID.String()),
			zap.Int("hours", hours),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve balance history",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    history,
	})
}

// GetTransactionReport Get transaction report with filters
// @Summary Get transaction report
// @Description Retrieve transaction report with optional filters including player_id
// @Tags analytics
// @Accept json
// @Produce json
// @Param player_id query string false "Player ID to filter by"
// @Param date_from query string false "Start date (RFC3339)"
// @Param date_to query string false "End date (RFC3339)"
// @Param transaction_type query string false "Transaction type"
// @Param status query string false "Transaction status"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 400 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/reports/transactions [get]
func (a *analytics) GetTransactionReport(c *gin.Context) {
	// Check if analytics storage is available
	if !a.checkAnalyticsStorage(c) {
		return
	}

	filters := &dto.TransactionFilters{}

	// Parse query parameters
	if playerIDStr := c.Query("player_id"); playerIDStr != "" {
		if playerID, err := uuid.Parse(playerIDStr); err == nil {
			filters.UserID = &playerID
		}
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	if transactionType := c.Query("transaction_type"); transactionType != "" {
		filters.TransactionType = &transactionType
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	transactions, err := a.analyticsStorage.GetTransactionReport(c.Request.Context(), filters)
	if err != nil {
		a.logger.Error("Failed to get transaction report",
			zap.Any("filters", filters),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve transaction report",
		})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsResponse{
		Success: true,
		Data:    transactions,
		Meta: &dto.Meta{
			Total:    len(transactions),
			PageSize: filters.Limit,
		},
	})
}
