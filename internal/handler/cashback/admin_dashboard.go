package cashback

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/cashback"
	"go.uber.org/zap"
)

// AdminDashboardHandler handles admin dashboard operations for cashback system
type AdminDashboardHandler struct {
	cashbackService *cashback.CashbackService
	logger          *zap.Logger
}

func NewAdminDashboardHandler(cashbackService *cashback.CashbackService, logger *zap.Logger) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		cashbackService: cashbackService,
		logger:          logger,
	}
}

// DashboardStats represents comprehensive dashboard statistics
type DashboardStats struct {
	TotalUsers             int64                   `json:"total_users"`
	ActiveUsers            int64                   `json:"active_users"`
	TotalCashbackEarned    decimal.Decimal         `json:"total_cashback_earned"`
	TotalCashbackClaimed   decimal.Decimal         `json:"total_cashback_claimed"`
	PendingCashback        decimal.Decimal         `json:"pending_cashback"`
	AverageCashbackPerUser decimal.Decimal         `json:"average_cashback_per_user"`
	TopCashbackUsers       []UserCashbackSummary   `json:"top_cashback_users"`
	CashbackTiers          []dto.CashbackTier      `json:"cashback_tiers"`
	RecentClaims           []dto.CashbackClaim     `json:"recent_claims"`
	DailyCashbackTrend     []DailyCashbackData     `json:"daily_cashback_trend"`
	GameTypeStats          []GameTypeCashbackStats `json:"game_type_stats"`
}

type UserCashbackSummary struct {
	UserID          uuid.UUID       `json:"user_id"`
	Username        string          `json:"username"`
	Email           string          `json:"email"`
	CurrentLevel    int             `json:"current_level"`
	TotalCashback   decimal.Decimal `json:"total_cashback"`
	PendingCashback decimal.Decimal `json:"pending_cashback"`
	ClaimedCashback decimal.Decimal `json:"claimed_cashback"`
	LastActivity    *time.Time      `json:"last_activity"`
}

type DailyCashbackData struct {
	Date         string          `json:"date"`
	TotalEarned  decimal.Decimal `json:"total_earned"`
	TotalClaimed decimal.Decimal `json:"total_claimed"`
	ActiveUsers  int64           `json:"active_users"`
	NewUsers     int64           `json:"new_users"`
}

type GameTypeCashbackStats struct {
	GameType        string          `json:"game_type"`
	TotalBets       int64           `json:"total_bets"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	TotalCashback   decimal.Decimal `json:"total_cashback"`
	AverageCashback decimal.Decimal `json:"average_cashback"`
	HouseEdge       decimal.Decimal `json:"house_edge"`
}

// GetDashboardStats returns comprehensive dashboard statistics
func (h *AdminDashboardHandler) GetDashboardStats(c *gin.Context) {
	h.logger.Info("Fetching admin dashboard statistics")

	// Get date range from query parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.logger.Error("Invalid start date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.logger.Error("Invalid end date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end date format. Use YYYY-MM-DD",
		})
		return
	}

	// Get comprehensive statistics
	stats, err := h.cashbackService.GetComprehensiveStats(c.Request.Context(), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get dashboard statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch dashboard statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserCashbackDetails returns detailed cashback information for a specific user
func (h *AdminDashboardHandler) GetUserCashbackDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	h.logger.Info("Fetching user cashback details", zap.String("user_id", userID.String()))

	// Get user level
	userLevel, err := h.cashbackService.GetUserLevel(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user level",
		})
		return
	}

	// Get user earnings
	earnings, err := h.cashbackService.GetUserCashbackEarnings(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user earnings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user earnings",
		})
		return
	}

	// Get user claims
	claims, err := h.cashbackService.GetUserCashbackClaims(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user claims", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_level": userLevel,
			"earnings":   earnings,
			"claims":     claims,
		},
	})
}

// UpdateCashbackTier updates a cashback tier configuration
func (h *AdminDashboardHandler) UpdateCashbackTier(c *gin.Context) {
	tierIDStr := c.Param("tier_id")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		h.logger.Error("Invalid tier ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tier ID format",
		})
		return
	}

	var updateRequest struct {
		TierName             string                 `json:"tier_name"`
		MinGGRRequired       decimal.Decimal        `json:"min_ggr_required"`
		CashbackPercentage   decimal.Decimal        `json:"cashback_percentage"`
		BonusMultiplier      decimal.Decimal        `json:"bonus_multiplier"`
		DailyCashbackLimit   *decimal.Decimal       `json:"daily_cashback_limit"`
		WeeklyCashbackLimit  *decimal.Decimal       `json:"weekly_cashback_limit"`
		MonthlyCashbackLimit *decimal.Decimal       `json:"monthly_cashback_limit"`
		SpecialBenefits      map[string]interface{} `json:"special_benefits"`
		IsActive             bool                   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating cashback tier", zap.String("tier_id", tierID.String()))

	// Update the tier
	updatedTier, err := h.cashbackService.UpdateCashbackTier(c.Request.Context(), tierID, dto.CashbackTier{
		ID:                   tierID,
		TierName:             updateRequest.TierName,
		TierLevel:            0, // Will be set by service
		MinGGRRequired:       updateRequest.MinGGRRequired,
		CashbackPercentage:   updateRequest.CashbackPercentage,
		BonusMultiplier:      updateRequest.BonusMultiplier,
		DailyCashbackLimit:   updateRequest.DailyCashbackLimit,
		WeeklyCashbackLimit:  updateRequest.WeeklyCashbackLimit,
		MonthlyCashbackLimit: updateRequest.MonthlyCashbackLimit,
		SpecialBenefits:      updateRequest.SpecialBenefits,
		IsActive:             updateRequest.IsActive,
	})

	if err != nil {
		h.logger.Error("Failed to update cashback tier", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cashback tier",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedTier,
	})
}

// CreateCashbackPromotion creates a new cashback promotion
func (h *AdminDashboardHandler) CreateCashbackPromotion(c *gin.Context) {
	var promotionRequest dto.CashbackPromotion
	if err := c.ShouldBindJSON(&promotionRequest); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating cashback promotion", zap.String("name", promotionRequest.PromotionName))

	// Create the promotion
	promotion, err := h.cashbackService.CreateCashbackPromotion(c.Request.Context(), promotionRequest)
	if err != nil {
		h.logger.Error("Failed to create cashback promotion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create cashback promotion",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    promotion,
	})
}

// GetCashbackAnalytics returns detailed analytics for the cashback system
func (h *AdminDashboardHandler) GetCashbackAnalytics(c *gin.Context) {
	// Get analytics parameters
	period := c.DefaultQuery("period", "30d") // 7d, 30d, 90d, 1y
	gameType := c.Query("game_type")          // Optional filter by game type

	h.logger.Info("Fetching cashback analytics",
		zap.String("period", period),
		zap.String("game_type", gameType))

	// Calculate date range based on period
	var startDate time.Time
	switch period {
	case "7d":
		startDate = time.Now().AddDate(0, 0, -7)
	case "30d":
		startDate = time.Now().AddDate(0, 0, -30)
	case "90d":
		startDate = time.Now().AddDate(0, 0, -90)
	case "1y":
		startDate = time.Now().AddDate(-1, 0, 0)
	default:
		startDate = time.Now().AddDate(0, 0, -30)
	}

	// Get analytics data
	analytics, err := h.cashbackService.GetCashbackAnalytics(c.Request.Context(), startDate, time.Now(), gameType)
	if err != nil {
		h.logger.Error("Failed to get cashback analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch analytics data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// ProcessManualCashback manually processes cashback for a user
func (h *AdminDashboardHandler) ProcessManualCashback(c *gin.Context) {
	var request struct {
		UserID   uuid.UUID       `json:"user_id"`
		Amount   decimal.Decimal `json:"amount"`
		Reason   string          `json:"reason"`
		GameType string          `json:"game_type"`
		GameID   string          `json:"game_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Processing manual cashback",
		zap.String("user_id", request.UserID.String()),
		zap.String("amount", request.Amount.String()),
		zap.String("reason", request.Reason))

	// Create manual cashback earning
	earning, err := h.cashbackService.CreateManualCashbackEarning(c.Request.Context(), request.UserID, request.Amount, request.Reason, request.GameType, request.GameID)
	if err != nil {
		h.logger.Error("Failed to process manual cashback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process manual cashback",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    earning,
	})
}

// GetSystemHealth returns the health status of the cashback system
func (h *AdminDashboardHandler) GetSystemHealth(c *gin.Context) {
	h.logger.Info("Checking cashback system health")

	health, err := h.cashbackService.GetSystemHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get system health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check system health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}
