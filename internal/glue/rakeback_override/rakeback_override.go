package rakeback_override

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
	rakebackOverrideHandler handler.RakebackOverride,
	authModule module.Authz,
	logsModule module.SystemLogs,
) {
	rakebackOverrideRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/rakeback-override",
			Handler: rakebackOverrideHandler.GetOverride,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get rakeback override", http.MethodGet),
				middleware.SystemLogs("get rakeback override", &log, logsModule),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/rakeback-override/active",
			Handler: rakebackOverrideHandler.GetActiveOverride,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get active rakeback override", http.MethodGet),
				middleware.SystemLogs("get active rakeback override", &log, logsModule),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/rakeback-override",
			Handler: rakebackOverrideHandler.CreateOrUpdateOverride,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create or update rakeback override", http.MethodPost),
				middleware.SystemLogs("create or update rakeback override", &log, logsModule),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/rakeback-override/toggle",
			Handler: rakebackOverrideHandler.ToggleOverride,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "toggle rakeback override", http.MethodPatch),
				middleware.SystemLogs("toggle rakeback override", &log, logsModule),
			},
		},
	}

	routing.RegisterRoute(group, rakebackOverrideRoutes, log)
}

