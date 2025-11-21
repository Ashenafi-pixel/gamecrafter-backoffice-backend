package admin_activity_logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	adminActivityLogsHandler handler.AdminActivityLogs,
	authModule module.Authz,
	adminActivityLogsModule module.AdminActivityLogs,
) {

	routes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/activity-logs",
			Handler: adminActivityLogsHandler.CreateAdminActivityLog,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Create Admin Activity Log", http.MethodPost),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs",
			Handler: adminActivityLogsHandler.GetAdminActivityLogs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Logs", http.MethodGet),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs/:id",
			Handler: adminActivityLogsHandler.GetAdminActivityLogByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Log By ID", http.MethodGet),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs/stats",
			Handler: adminActivityLogsHandler.GetAdminActivityStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Stats", http.MethodGet),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs/categories",
			Handler: adminActivityLogsHandler.GetAdminActivityCategories,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Categories", http.MethodGet),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs/actions",
			Handler: adminActivityLogsHandler.GetAdminActivityActions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Actions", http.MethodGet),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/activity-logs/actions/category/:category",
			Handler: adminActivityLogsHandler.GetAdminActivityActionsByCategory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Admin Activity Actions By Category", http.MethodGet),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/activity-logs/:id",
			Handler: adminActivityLogsHandler.DeleteAdminActivityLog,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Delete Admin Activity Log", http.MethodDelete),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/activity-logs/admin/:admin_id",
			Handler: adminActivityLogsHandler.DeleteAdminActivityLogsByAdmin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Delete Admin Activity Logs By Admin", http.MethodDelete),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/activity-logs/old",
			Handler: adminActivityLogsHandler.DeleteOldAdminActivityLogs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Delete Old Admin Activity Logs", http.MethodDelete),
			},
		},
	}

	routing.RegisterRoute(group, routes, log)
}
