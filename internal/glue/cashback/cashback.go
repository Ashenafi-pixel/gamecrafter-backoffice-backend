package cashback

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/handler/cashback"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(r *gin.RouterGroup, log zap.Logger, handler *cashback.CashbackHandler, authz module.Authz, systemLogs module.SystemLogs) {
	log.Info("Initializing cashback routes")

	// Public routes (no authentication required)
	public := r.Group("/api/cashback")
	{
		public.GET("/tiers", handler.GetCashbackTiers)
		public.GET("/house-edge", handler.GetGameHouseEdge)
		public.GET("/schedules/:id", handler.GetRakebackSchedule)
	}

	// User routes (authentication required)
	user := r.Group("/api/user/cashback")

	user.Use(middleware.Auth())
	{
		user.GET("", handler.GetUserCashbackSummary)
		user.POST("/claim", handler.ClaimCashback)
		user.GET("/earnings", handler.GetUserCashbackEarnings)
		user.GET("/claims", handler.GetUserCashbackClaims)

		// Retry operations routes
		user.GET("/retry-operations", handler.GetRetryableOperations)
		user.POST("/retry-operations/:operation_id", handler.ManualRetryOperation)

		// Level progression routes
		user.GET("/level-progression", handler.GetLevelProgressionInfo)
	}

	// Balance synchronization routes (authentication required)
	balance := r.Group("/api/user/balance")
	balance.Use(middleware.Auth())
	{
		balance.GET("/validate-sync", handler.ValidateBalanceSync)
		balance.POST("/reconcile", handler.ReconcileBalances)
	}

	// Admin routes (admin authentication required)
	admin := r.Group("/api/admin/cashback")
	admin.Use(middleware.Auth())
	// Check cashback permission for all admin routes
	admin.Use(func(c *gin.Context) {
		userID := c.GetString("user-id")
		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Check if user has "super" role first
		roles, err := authz.GetUserRoles(context.Background(), userIDParsed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
			c.Abort()
			return
		}

		// Check for super admin role
		for _, role := range roles.Roles {
			if role.Name == "super" {
				c.Next()
				return
			}
		}

		// Check permission directly from database
		hasPermission, err := authz.CheckUserHasPermission(context.Background(), userIDParsed, "cashback")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "User does not have permission: cashback"})
			c.Abort()
			return
		}

		c.Next()
	})
	{
		admin.GET("/stats", handler.GetCashbackStats)
		admin.GET("/tiers", handler.GetCashbackTiers)
		admin.POST("/tiers", handler.CreateCashbackTier)
		admin.PUT("/tiers/:id", handler.UpdateCashbackTier)
		admin.DELETE("/tiers/:id", handler.DeleteCashbackTier)
		admin.POST("/tiers/reorder", handler.ReorderCashbackTiers)
		admin.POST("/promotions", handler.CreateCashbackPromotion)
		admin.POST("/house-edge", handler.CreateGameHouseEdge)

		// Admin Dashboard routes
		admin.GET("/dashboard", handler.GetDashboardStats)
		admin.GET("/dashboard/analytics", handler.GetCashbackAnalytics)
		admin.GET("/dashboard/health", handler.GetSystemHealth)
		admin.GET("/dashboard/users/:user_id", handler.GetUserCashbackDetails)
		admin.POST("/dashboard/manual-cashback", handler.ProcessManualCashback)

		// Admin Retry Operations routes
		admin.POST("/retry-failed-operations", handler.RetryFailedOperations)

		// Admin Level Progression routes
		admin.GET("/level-progression-info", handler.GetLevelProgressionInfoForUser)
		admin.POST("/level-progression", handler.ProcessSingleLevelProgression)
		admin.POST("/bulk-level-progression", handler.ProcessBulkLevelProgression)

		// Global Rakeback Override routes (Happy Hour Mode)
		admin.GET("/global-override", handler.GetGlobalRakebackOverride)
		admin.PUT("/global-override", handler.UpdateGlobalRakebackOverride)

		// Rakeback Schedule routes (Scheduled Happy Hours)
		admin.POST("/schedules", handler.CreateRakebackSchedule)
		admin.GET("/schedules", handler.ListRakebackSchedules)
		admin.GET("/schedules/:id", handler.GetRakebackSchedule)
		admin.PUT("/schedules/:id", handler.UpdateRakebackSchedule)
		admin.DELETE("/schedules/:id", handler.DeleteRakebackSchedule)
	}

	log.Info("Cashback routes initialized successfully")
}
