package balancelogs

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
	op handler.BalanceLogs,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {

	balanceRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/balance/logs",
			Handler: op.GetBalanceLogs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},{
			Method:  http.MethodGet,
			Path:   "/api/balance/logs/:id",
			Handler: op.GetBalanceLogByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/balance/logs",
			Handler: op.GetBalanceLogsForAdmin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get balance logs", http.MethodGet),
				middleware.SystemLogs("Get Balance Logs", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, balanceRoutes, log)
}
