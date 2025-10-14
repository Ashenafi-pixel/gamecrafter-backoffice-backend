package analytics

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler"
	"go.uber.org/zap"
)

func Init(grp *gin.RouterGroup, log *zap.Logger, analyticsHandler handler.Analytics) {
	analyticsGroup := grp.Group("/api/admin/analytics")
	{
		// User analytics endpoints
		analyticsGroup.GET("/users/:user_id/transactions", analyticsHandler.GetUserTransactions)
		analyticsGroup.GET("/users/:user_id/analytics", analyticsHandler.GetUserAnalytics)
		analyticsGroup.GET("/users/:user_id/balance-history", analyticsHandler.GetUserBalanceHistory)

		// Real-time analytics
		analyticsGroup.GET("/realtime/stats", analyticsHandler.GetRealTimeStats)

		// Reporting endpoints
		analyticsGroup.GET("/reports/daily", analyticsHandler.GetDailyReport)
		analyticsGroup.GET("/reports/daily-enhanced", analyticsHandler.GetEnhancedDailyReport)
		analyticsGroup.GET("/reports/top-games", analyticsHandler.GetTopGames)
		analyticsGroup.GET("/reports/top-players", analyticsHandler.GetTopPlayers)

		// Daily report email endpoints
		analyticsGroup.POST("/daily-report/send", analyticsHandler.SendDailyReportEmail)
		analyticsGroup.POST("/daily-report/send-configured", analyticsHandler.SendConfiguredDailyReportEmail)
		analyticsGroup.POST("/daily-report/yesterday", analyticsHandler.SendYesterdayReportEmail)
		analyticsGroup.POST("/daily-report/schedule", analyticsHandler.ScheduleDailyReportCronJob)
		analyticsGroup.POST("/daily-report/last-week", analyticsHandler.SendLastWeekReportEmail)
		analyticsGroup.POST("/daily-report/test", analyticsHandler.SendTestDailyReport)
		analyticsGroup.GET("/daily-report/cronjob-status", analyticsHandler.GetCronjobStatus)
	}
}
