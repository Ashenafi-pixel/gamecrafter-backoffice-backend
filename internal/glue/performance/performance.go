package performance

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"github.com/joshjones612/egyptkingcrash/internal/module"
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
