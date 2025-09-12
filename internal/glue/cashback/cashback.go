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
	}

	// User routes (authentication required)
	user := r.Group("/user/cashback")
	user.Use(middleware.Auth())
	{
		user.GET("", handler.GetUserCashbackSummary)
		user.POST("/claim", handler.ClaimCashback)
		user.GET("/earnings", handler.GetUserCashbackEarnings)
		user.GET("/claims", handler.GetUserCashbackClaims)
	}

	// Admin routes (admin authentication required)
	admin := r.Group("/admin/cashback")
	admin.Use(middleware.Auth())
	admin.Use(middleware.Authz(authz, enforcer, "cashback", "admin"))
	{
		admin.GET("/stats", handler.GetCashbackStats)
		admin.POST("/tiers", handler.CreateCashbackTier)
		admin.PUT("/tiers/:id", handler.UpdateCashbackTier)
		admin.POST("/promotions", handler.CreateCashbackPromotion)
	}

	log.Info("Cashback routes initialized successfully")
}
