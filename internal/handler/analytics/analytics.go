package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func Init(log *zap.Logger, analyticsStorage storage.Analytics, dailyReportService analyticsModule.DailyReportService, dailyReportCronjobService analyticsModule.DailyReportCronjobService) handler.Analytics {
	return &analytics{
		logger:                    log,
		analyticsStorage:          analyticsStorage,
		dailyReportService:        dailyReportService,
		dailyReportCronjobService: dailyReportCronjobService,
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

	if transactionType := c.Query("transaction_type"); transactionType != "" {
		filters.TransactionType = &transactionType
	}

	if gameID := c.Query("game_id"); gameID != "" {
		filters.GameID = &gameID
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

	transactions, err := a.analyticsStorage.GetUserTransactions(c.Request.Context(), userID, filters)
	if err != nil {
		a.logger.Error("Failed to get user transactions",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve transactions",
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
// @Param date_from query string false "Start date (RFC3339)"
// @Param date_to query string false "End date (RFC3339)"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/games/top [get]
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
// @Param date_from query string false "Start date (RFC3339)"
// @Param date_to query string false "End date (RFC3339)"
// @Success 200 {object} dto.AnalyticsResponse
// @Failure 500 {object} dto.AnalyticsResponse
// @Router /analytics/players/top [get]
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

	players, err := a.analyticsStorage.GetTopPlayers(c.Request.Context(), limit, dateRange)
	if err != nil {
		a.logger.Error("Failed to get top players", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.AnalyticsResponse{
			Success: false,
			Error:   "Failed to retrieve top players",
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
