package report

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
	report handler.Report,
	authModule module.Authz,
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
				middleware.Authz(authModule, "get daily report", http.MethodGet),
				middleware.SystemLogs("Get daily Report", &log, systemLogs),
			},
		},
	}

	routing.RegisterRoute(group, reportRoutes, log)
}
