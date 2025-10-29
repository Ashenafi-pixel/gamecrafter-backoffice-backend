package report

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
	report handler.Report,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {
	reportRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/report/daily",
			Handler: report.GetDailyReport,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get daily report", http.MethodGet),
				middleware.SystemLogs("Get daily Report", &log, systemLogs),
			},
		},
	}

	routing.RegisterRoute(group, reportRoutes, log)
}
