package risksettings

import (
	"net/http"

	"github.com/casbin/casbin/v2"
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
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	riskSettings handler.RiskSettings,
	systemLogs module.SystemLogs,
) {
	riskSettingsRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/risk/settings",
			Handler: riskSettings.GetRiskSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "risksettings get", http.MethodGet),
				middleware.SystemLogs("Get Risk Settings", &log, systemLogs),
			},
		},
		{

			Method:  http.MethodPut,
			Path:    "/api/admin/risk/settings",
			Handler: riskSettings.SetRiskSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "risksettings update", http.MethodPut),
				middleware.SystemLogs("Update Risk Settings", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, riskSettingsRoutes, log)
}
