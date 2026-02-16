package balancelogs

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
	op handler.BalanceLogs,
	authModule module.Authz,
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
		}, {
			Method:  http.MethodGet,
			Path:    "/api/balance/logs/:id",
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
				middleware.Authz(authModule, "get balance logs", http.MethodGet),
				middleware.SystemLogs("Get Balance Logs", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, balanceRoutes, log)
}
