package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	notification handler.Notification,
) {
	notificationRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/notifications",
			Handler: notification.CreateNotification,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/notifications",
			Handler: notification.GetUserNotifications,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/notifications/:id/mark-read",
			Handler: notification.MarkNotificationRead,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/notifications/mark-all-read",
			Handler: notification.MarkAllNotificationsRead,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/notifications/:id",
			Handler: notification.DeleteNotification,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, notificationRoutes, log)
}
