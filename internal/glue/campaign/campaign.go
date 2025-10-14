package campaign

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func InitRoutes(router *gin.RouterGroup, campaignHandler handler.Campaign, log *zap.Logger, authModule module.Authz, enforcer *casbin.Enforcer) {
	campaigns := router.Group("/api/v1/campaigns")
	campaigns.Use(middleware.Auth())
	{
		campaigns.POST("", campaignHandler.CreateCampaign)
		campaigns.GET("", campaignHandler.GetCampaigns)
		campaigns.GET("/:id", campaignHandler.GetCampaignByID)
		campaigns.PUT("/:id", campaignHandler.UpdateCampaign)
		campaigns.DELETE("/:id", campaignHandler.DeleteCampaign)
		campaigns.POST("/:id/send", campaignHandler.SendCampaign)
		campaigns.GET("/:id/recipients", campaignHandler.GetCampaignRecipients)
		campaigns.GET("/:id/stats", campaignHandler.GetCampaignStats)
	}

	// Campaign notifications dashboard endpoint
	campaignNotificationsDashboard := router.Group("/api/v1/campaign-notifications-dashboard")
	campaignNotificationsDashboard.Use(middleware.Auth())
	{
		campaignNotificationsDashboard.GET("", campaignHandler.GetCampaignNotificationsDashboard)
	}
}
