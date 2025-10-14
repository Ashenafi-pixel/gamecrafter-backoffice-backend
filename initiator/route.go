package initiator

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/adds"
	"github.com/tucanbit/internal/glue/admin_notification"
	"github.com/tucanbit/internal/glue/agent"
	"github.com/tucanbit/internal/glue/analytics"
	"github.com/tucanbit/internal/glue/authz"
	"github.com/tucanbit/internal/glue/balance"
	"github.com/tucanbit/internal/glue/balancelogs"
	"github.com/tucanbit/internal/glue/banner"
	"github.com/tucanbit/internal/glue/bet"
	"github.com/tucanbit/internal/glue/campaign"
	"github.com/tucanbit/internal/glue/cashback"
	"github.com/tucanbit/internal/glue/company"
	"github.com/tucanbit/internal/glue/department"
	"github.com/tucanbit/internal/glue/exchange"
	"github.com/tucanbit/internal/glue/falcon_liquidity"
	"github.com/tucanbit/internal/glue/game"
	"github.com/tucanbit/internal/glue/groove"
	"github.com/tucanbit/internal/glue/logs"
	"github.com/tucanbit/internal/glue/lottery"
	"github.com/tucanbit/internal/glue/notification"
	"github.com/tucanbit/internal/glue/operationalgroup"
	"github.com/tucanbit/internal/glue/operationalgrouptype"
	"github.com/tucanbit/internal/glue/operationsdefinitions"
	"github.com/tucanbit/internal/glue/otp"
	"github.com/tucanbit/internal/glue/performance"
	"github.com/tucanbit/internal/glue/report"
	"github.com/tucanbit/internal/glue/risksettings"
	"github.com/tucanbit/internal/glue/sportsservice"
	"github.com/tucanbit/internal/glue/squads"
	"github.com/tucanbit/internal/glue/twofactor"
	"github.com/tucanbit/internal/glue/user"
	"github.com/tucanbit/internal/glue/ws"
	falconStorage "github.com/tucanbit/internal/storage/falcon_liquidity"
	"go.uber.org/zap"
)

func initRoute(grp *gin.RouterGroup, handler *Handler, module *Module, log *zap.Logger, enforcer *casbin.Enforcer, persistence *Persistence) {
	user.Init(grp, *log, handler.User, module.User, module.Authz, enforcer, module.SystemLogs)
	operationalgroup.Init(grp, *log, handler.OperationalGroup, module.Authz, enforcer, module.SystemLogs)
	operationalgrouptype.Init(grp, *log, handler.OperationalGroupType, module.Authz, enforcer, module.SystemLogs)
	operationsdefinitions.Init(grp, *log, handler.OperationsDefinitions, module.Authz, enforcer)
	balance.Init(grp, *log, handler.Balance, module.User, module.Authz, enforcer, module.SystemLogs)
	balancelogs.Init(grp, *log, handler.BalanceLogs, module.Authz, enforcer, module.SystemLogs)
	exchange.Init(grp, *log, handler.Exchange)
	ws.Init(grp, *log, handler.WS)
	bet.Init(grp, *log, handler.Bet, module.User, module.Authz, enforcer, module.SystemLogs)
	department.Init(grp, *log, handler.Departments, module.Authz, enforcer, module.SystemLogs)
	performance.Init(grp, *log, handler.Performance, module.Authz, enforcer, module.SystemLogs)
	authz.Init(grp, *log, handler.Authz, module.Authz, enforcer, module.SystemLogs)
	logs.Init(grp, *log, handler.SystemLogs, module.Authz, module.SystemLogs, enforcer)
	company.Init(grp, *log, handler.Company, module.Authz, enforcer, module.SystemLogs)
	report.Init(grp, *log, handler.Report, module.Authz, enforcer, module.SystemLogs)
	squads.Init(grp, *log, handler.Squads, module.Authz, enforcer, module.SystemLogs)
	notification.Init(grp, *log, handler.Notification)
	admin_notification.Init(grp, *log, handler.Notification)
	campaign.InitRoutes(grp, handler.Campaign, log, module.Authz, enforcer)
	adds.Init(grp, *log, module.Authz, enforcer, handler.Adds, module.SystemLogs)
	banner.Init(grp, *log, module.Authz, enforcer, handler.Banner, module.SystemLogs)
	lottery.Init(grp, *log, handler.Lottery, module.Authz, module.SystemLogs, enforcer)
	sportsservice.Init(grp, *log, handler.SportsService)
	risksettings.Init(grp, *log, module.Authz, enforcer, handler.RiskSettings, module.SystemLogs)
	agent.Init(grp, *log, handler.Agent, module.Authz, module.SystemLogs, enforcer)
	otp.Init(grp, *log, handler.OTP, module.OTP)
	cashback.Init(grp, *log, handler.Cashback, module.Authz, enforcer, module.SystemLogs)
	groove.Init(grp, log, handler.Groove, module.Groove, module.Authz, enforcer, module.SystemLogs)
	game.Init(grp, *log, handler.Game, handler.HouseEdge, module.Authz, module.SystemLogs, enforcer)
	analytics.Init(grp, log, handler.Analytics)
	twofactor.Init(grp, log, handler.TwoFactor)

	// Initialize Falcon Liquidity routes (no authentication required)
	falconStorageInstance := falconStorage.NewFalconMessageStorage(log, persistence.Database)
	falconRoutes := falcon_liquidity.GetFalconLiquidityRoutes(falconStorageInstance, log)
	for _, route := range falconRoutes {
		grp.Handle(route.Method, route.Path, route.Handler)
	}
}
