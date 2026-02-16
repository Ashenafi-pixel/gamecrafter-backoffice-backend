package alert

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/alert"
	"github.com/tucanbit/internal/handler/middleware"
)

func Init(r *gin.RouterGroup, handler alert.AlertHandler, emailGroupHandler alert.AlertEmailGroupHandler) {
	alertGroup := r.Group("/alerts")
	// Apply authentication middleware to all alert routes
	alertGroup.Use(middleware.Auth())
	{
		// Alert Configuration routes
		alertGroup.POST("/configurations", handler.CreateAlertConfiguration)
		alertGroup.GET("/configurations", handler.GetAlertConfigurations)
		alertGroup.GET("/configurations/:id", handler.GetAlertConfiguration)
		alertGroup.PUT("/configurations/:id", handler.UpdateAlertConfiguration)
		alertGroup.DELETE("/configurations/:id", handler.DeleteAlertConfiguration)

		// Alert Trigger routes
		alertGroup.GET("/triggers", handler.GetAlertTriggers)
		alertGroup.GET("/triggers/:id", handler.GetAlertTrigger)
		alertGroup.PUT("/triggers/:id/acknowledge", handler.AcknowledgeAlert)
		alertGroup.POST("/triggers/test", handler.TestTriggerAlert) // Test endpoint to manually trigger alerts

		// Email Group routes
		alertGroup.POST("/email-groups", emailGroupHandler.CreateEmailGroup)
		alertGroup.GET("/email-groups", emailGroupHandler.GetAllEmailGroups)
		alertGroup.GET("/email-groups/:id", emailGroupHandler.GetEmailGroup)
		alertGroup.PUT("/email-groups/:id", emailGroupHandler.UpdateEmailGroup)
		alertGroup.DELETE("/email-groups/:id", emailGroupHandler.DeleteEmailGroup)
	}
}
