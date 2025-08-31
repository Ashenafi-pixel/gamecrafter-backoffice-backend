package performance

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
	p handler.Performance,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLog module.SystemLogs,
) {

	gameMatricRouts := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/performance/financial",
			Handler: p.GetFinancialMetrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get financial metrics", http.MethodGet),
				middleware.SystemLogs("Get Financial Matrics", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/performance/game",
			Handler: p.GameMatrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get game metrics", http.MethodGet),
				middleware.SystemLogs("Get Game Matrics", &log, systemLog),
			},
		},
	}
	routing.RegisterRoute(group, gameMatricRouts, log)
}
