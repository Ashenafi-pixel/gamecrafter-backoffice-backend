package lottery

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
	lotteryHanlder handler.Lottery,
	authModule module.Authz,
	logsModule module.SystemLogs,
	enforcer *casbin.Enforcer,
) {

	logs := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/admin/lottery/service",
			Handler: lotteryHanlder.CreateLotteryService,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "Create Lottery Service", http.MethodPost),
				middleware.SystemLogs("Create Lottery Service", &log, logsModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/admin/lottery/request",
			Handler: lotteryHanlder.CreateLotteryRequest,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "Create Lottery", http.MethodPost),
				middleware.SystemLogs("Create Lottery", &log, logsModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/lottery/verify/deduct/balance",
			Handler: lotteryHanlder.CheckUserBalanceAndDeductBalance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.LotteryUserAuth(),
			},
		},
	}
	routing.RegisterRoute(group, logs, log)
}
