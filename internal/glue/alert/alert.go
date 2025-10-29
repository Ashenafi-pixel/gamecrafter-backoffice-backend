package alert

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/alert"
)

func Init(r *gin.RouterGroup, handler alert.AlertHandler) {
	alertGroup := r.Group("/alerts")
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
	}
}
