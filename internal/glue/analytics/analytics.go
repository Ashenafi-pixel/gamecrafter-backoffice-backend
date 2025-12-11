package analytics

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler"
	"go.uber.org/zap"
)

func Init(grp *gin.RouterGroup, log *zap.Logger, analyticsHandler handler.Analytics) {
	// Admin analytics routes (requires authentication)
	adminAnalyticsGroup := grp.Group("/api/admin/analytics")
	{
		// User analytics endpoints
		adminAnalyticsGroup.GET("/users/:user_id/transactions", analyticsHandler.GetUserTransactions)
		adminAnalyticsGroup.GET("/users/:user_id/transactions/totals", analyticsHandler.GetUserTransactionsTotals)
		adminAnalyticsGroup.GET("/users/:user_id/analytics", analyticsHandler.GetUserAnalytics)
		adminAnalyticsGroup.GET("/users/:user_id/balance-history", analyticsHandler.GetUserBalanceHistory)
		adminAnalyticsGroup.GET("/users/:user_id/rakeback", analyticsHandler.GetUserRakebackTransactions)
		adminAnalyticsGroup.GET("/users/:user_id/rakeback/totals", analyticsHandler.GetUserRakebackTotals)
		adminAnalyticsGroup.GET("/users/:user_id/tips", analyticsHandler.GetUserTips)
		adminAnalyticsGroup.GET("/users/:user_id/tips/totals", analyticsHandler.GetUserTipsTotals)
		adminAnalyticsGroup.GET("/users/:user_id/welcome_bonus", analyticsHandler.GetUserWelcomeBonus)
		adminAnalyticsGroup.GET("/users/:user_id/welcome_bonus/totals", analyticsHandler.GetUserWelcomeBonusTotals)
		// Admin endpoint to get all welcome bonuses with filters
		adminAnalyticsGroup.GET("/welcome_bonus", analyticsHandler.GetWelcomeBonusTransactions)

		// Real-time analytics
		adminAnalyticsGroup.GET("/realtime/stats", analyticsHandler.GetRealTimeStats)

		// Reporting endpoints
		adminAnalyticsGroup.GET("/reports/daily", analyticsHandler.GetDailyReport)
		adminAnalyticsGroup.GET("/reports/daily-enhanced", analyticsHandler.GetEnhancedDailyReport)
		adminAnalyticsGroup.GET("/reports/transactions", analyticsHandler.GetTransactionReport)
		adminAnalyticsGroup.GET("/reports/top-games", analyticsHandler.GetTopGames)
		adminAnalyticsGroup.GET("/reports/top-players", analyticsHandler.GetTopPlayers)

		// Daily report email endpoints
		adminAnalyticsGroup.POST("/daily-report/send", analyticsHandler.SendDailyReportEmail)
		adminAnalyticsGroup.POST("/daily-report/send-configured", analyticsHandler.SendConfiguredDailyReportEmail)
		adminAnalyticsGroup.POST("/daily-report/yesterday", analyticsHandler.SendYesterdayReportEmail)
		adminAnalyticsGroup.POST("/daily-report/schedule", analyticsHandler.ScheduleDailyReportCronJob)
		adminAnalyticsGroup.POST("/daily-report/last-week", analyticsHandler.SendLastWeekReportEmail)
		adminAnalyticsGroup.POST("/daily-report/test", analyticsHandler.SendTestDailyReport)
		adminAnalyticsGroup.GET("/daily-report/cronjob-status", analyticsHandler.GetCronjobStatus)
	}

	// Public analytics routes (no authentication required)
	publicAnalyticsGroup := grp.Group("/analytics")
	{
		// User analytics endpoints
		publicAnalyticsGroup.GET("/users/:user_id/transactions", analyticsHandler.GetUserTransactions)
		publicAnalyticsGroup.GET("/users/:user_id/analytics", analyticsHandler.GetUserAnalytics)
		publicAnalyticsGroup.GET("/users/:user_id/balance-history", analyticsHandler.GetUserBalanceHistory)
		publicAnalyticsGroup.GET("/users/:user_id/rakeback", analyticsHandler.GetUserRakebackTransactions)
		publicAnalyticsGroup.GET("/users/:user_id/tips", analyticsHandler.GetUserTips)

		// Real-time analytics
		publicAnalyticsGroup.GET("/realtime/stats", analyticsHandler.GetRealTimeStats)
	}
}
