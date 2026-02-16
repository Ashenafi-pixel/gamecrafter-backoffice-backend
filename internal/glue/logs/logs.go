package logs

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
	logsHandler handler.SystemLogs,
	authModule module.Authz,
	logsModule module.SystemLogs,
) {

	logs := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/logs",
			Handler: logsHandler.GetSystemLogs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Audit Logs", http.MethodPost),
				middleware.SystemLogs("Get Audit Logs", &log, logsModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/logs/modules",
			Handler: logsHandler.GetAvailableLogModules,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Available Logs Module", http.MethodGet),
				middleware.SystemLogs("Get Available Logs Module", &log, logsModule),
			},
		},
	}
	routing.RegisterRoute(group, logs, log)
}
