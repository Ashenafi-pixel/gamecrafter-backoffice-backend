package cashback

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/cashback"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(r *gin.RouterGroup, log zap.Logger, handler *cashback.CashbackHandler, authz module.Authz, enforcer *casbin.Enforcer, systemLogs module.SystemLogs) {
	log.Info("Initializing cashback routes")

	// Public routes (no authentication required)
	public := r.Group("/cashback")
	{
		public.GET("/tiers", handler.GetCashbackTiers)
		public.GET("/house-edge", handler.GetGameHouseEdge)
	}

	// User routes (authentication required)
	user := r.Group("/user/cashback")
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
	balance := r.Group("/user/balance")
	balance.Use(middleware.Auth())
	{
		balance.GET("/validate-sync", handler.ValidateBalanceSync)
		balance.POST("/reconcile", handler.ReconcileBalances)
	}

	// Admin routes (admin authentication required)
	admin := r.Group("/api/admin/cashback")
	admin.Use(middleware.Auth())
	admin.Use(middleware.Authz(authz, enforcer, "cashback", "admin"))
	{
		admin.GET("/stats", handler.GetCashbackStats)
		admin.GET("/tiers", handler.GetCashbackTiers)
		admin.POST("/tiers", handler.CreateCashbackTier)
		admin.PUT("/tiers/:id", handler.UpdateCashbackTier)
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
		admin.POST("/bulk-level-progression", handler.ProcessBulkLevelProgression)
	}

	log.Info("Cashback routes initialized successfully")
}
