package initiator

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/adds"
	"github.com/joshjones612/egyptkingcrash/internal/glue/agent"
	"github.com/joshjones612/egyptkingcrash/internal/glue/airtime"
	"github.com/joshjones612/egyptkingcrash/internal/glue/authz"
	"github.com/joshjones612/egyptkingcrash/internal/glue/balance"
	"github.com/joshjones612/egyptkingcrash/internal/glue/balancelogs"
	"github.com/joshjones612/egyptkingcrash/internal/glue/banner"
	"github.com/joshjones612/egyptkingcrash/internal/glue/bet"
	"github.com/joshjones612/egyptkingcrash/internal/glue/company"
	"github.com/joshjones612/egyptkingcrash/internal/glue/department"
	"github.com/joshjones612/egyptkingcrash/internal/glue/exchange"
	"github.com/joshjones612/egyptkingcrash/internal/glue/logs"
	"github.com/joshjones612/egyptkingcrash/internal/glue/lottery"
	"github.com/joshjones612/egyptkingcrash/internal/glue/notification"
	"github.com/joshjones612/egyptkingcrash/internal/glue/operationalgroup"
	"github.com/joshjones612/egyptkingcrash/internal/glue/operationalgrouptype"
	"github.com/joshjones612/egyptkingcrash/internal/glue/operationsdefinitions"
	"github.com/joshjones612/egyptkingcrash/internal/glue/performance"
	"github.com/joshjones612/egyptkingcrash/internal/glue/report"
	"github.com/joshjones612/egyptkingcrash/internal/glue/risksettings"
	"github.com/joshjones612/egyptkingcrash/internal/glue/sportsservice"
	"github.com/joshjones612/egyptkingcrash/internal/glue/squads"
	"github.com/joshjones612/egyptkingcrash/internal/glue/user"
	"github.com/joshjones612/egyptkingcrash/internal/glue/ws"
	"go.uber.org/zap"
)

func initRoute(grp *gin.RouterGroup, handler *Handler, module *Module, log *zap.Logger, enforcer *casbin.Enforcer) {
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
	airtime.Init(grp, *log, handler.Airtime, module.Authz, enforcer, module.SystemLogs)
	logs.Init(grp, *log, handler.SystemLogs, module.Authz, module.SystemLogs, enforcer)
	company.Init(grp, *log, handler.Company, module.Authz, enforcer, module.SystemLogs)
	report.Init(grp, *log, handler.Report, module.Authz, enforcer, module.SystemLogs)
	squads.Init(grp, *log, handler.Squads, module.Authz, enforcer, module.SystemLogs)
	notification.Init(grp, *log, handler.Notification)
	adds.Init(grp, *log, module.Authz, enforcer, handler.Adds, module.SystemLogs)
	banner.Init(grp, *log, module.Authz, enforcer, handler.Banner, module.SystemLogs)
	lottery.Init(grp, *log, handler.Lottery, module.Authz, module.SystemLogs, enforcer)
	sportsservice.Init(grp, *log, handler.SportsService)
	risksettings.Init(grp, *log, module.Authz, enforcer, handler.RiskSettings, module.SystemLogs)
	agent.Init(grp, *log, handler.Agent, module.Authz, module.SystemLogs, enforcer)
}
