package admin_notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	notification handler.Notification,
) {
	adminNotificationRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/campaignNotifications",
			Handler: notification.GetAdminNotifications,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, adminNotificationRoutes, log)
}
