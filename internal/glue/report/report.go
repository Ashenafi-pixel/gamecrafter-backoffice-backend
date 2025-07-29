package report

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
	"net/http"
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
