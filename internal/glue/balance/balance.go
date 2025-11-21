package balance

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
	op handler.Balance,
	userModule module.User,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {

	balanceRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/balance",
			Handler: op.GetUserBalances,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/balance/exchange",
			Handler: op.ExchangeBalance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/players/funding",
			Handler: op.ManualFunding,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "manual funding", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/balance/log/funds",
			Handler: op.GetManualFundLogs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get fund logs", http.MethodGet),
				middleware.SystemLogs("Get Manual Funds Logs", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/manual-funds",
			Handler: op.GetAllManualFunds,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get all manual funds", http.MethodGet),
				middleware.SystemLogs("Get All Manual Funds", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/wallet/credit",
			Handler: op.CreditWallet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}
	routing.RegisterRoute(group, balanceRoutes, log)
}
