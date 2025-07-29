package initiator

import (
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/adds"
	"github.com/joshjones612/egyptkingcrash/internal/handler/agent"
	"github.com/joshjones612/egyptkingcrash/internal/handler/airtime"
	"github.com/joshjones612/egyptkingcrash/internal/handler/authz"
	"github.com/joshjones612/egyptkingcrash/internal/handler/balance"
	"github.com/joshjones612/egyptkingcrash/internal/handler/balancelogs"
	"github.com/joshjones612/egyptkingcrash/internal/handler/banner"
	"github.com/joshjones612/egyptkingcrash/internal/handler/bet"
	"github.com/joshjones612/egyptkingcrash/internal/handler/company"
	"github.com/joshjones612/egyptkingcrash/internal/handler/department"
	"github.com/joshjones612/egyptkingcrash/internal/handler/exchange"
	"github.com/joshjones612/egyptkingcrash/internal/handler/logs"
	"github.com/joshjones612/egyptkingcrash/internal/handler/lottery"
	"github.com/joshjones612/egyptkingcrash/internal/handler/notification"
	"github.com/joshjones612/egyptkingcrash/internal/handler/operationalgroup"
	"github.com/joshjones612/egyptkingcrash/internal/handler/operationalgrouptype"
	"github.com/joshjones612/egyptkingcrash/internal/handler/operationsdefinitions"
	"github.com/joshjones612/egyptkingcrash/internal/handler/performance"
	"github.com/joshjones612/egyptkingcrash/internal/handler/report"
	"github.com/joshjones612/egyptkingcrash/internal/handler/risksettings"
	"github.com/joshjones612/egyptkingcrash/internal/handler/sportsservice"
	"github.com/joshjones612/egyptkingcrash/internal/handler/squads"
	"github.com/joshjones612/egyptkingcrash/internal/handler/user"
	"github.com/joshjones612/egyptkingcrash/internal/handler/ws"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Handler struct {
	User                  handler.User
	OperationalGroup      handler.OpeartionalGroup
	OperationalGroupType  handler.OperationalGroupType
	OperationsDefinitions handler.OperationsDefinition
	Balance               handler.Balance
	BalanceLogs           handler.BalanceLogs
	Exchange              handler.Exchange
	WS                    handler.WS
	Bet                   handler.Bet
	Departments           handler.Departements
	Performance           handler.Performance
	Authz                 handler.Authz
	Airtime               handler.AirtimeProvider
	SystemLogs            handler.SystemLogs
	Company               handler.Company
	Report                handler.Report
	Squads                handler.Squads
	Notification          handler.Notification
	Adds                  handler.Adds
	Banner                handler.Banner
	Lottery               handler.Lottery
	SportsService         handler.SportsService
	RiskSettings          handler.RiskSettings
	Agent                 handler.Agent
}

func initHandler(module *Module, log *zap.Logger, userWS utils.UserWS) *Handler {
	return &Handler{
		User:                  user.Init(module.User, log, viper.GetString("oauth.frontend_oauth_handler_url")),
		OperationalGroup:      operationalgroup.Init(module.OperationalGroup, log),
		OperationalGroupType:  operationalgrouptype.Init(module.OperationalGroupType, log),
		OperationsDefinitions: operationsdefinitions.Init(module.OperationsDefinitions, log),
		Balance:               balance.Init(module.Balance, log),
		BalanceLogs:           balancelogs.Init(module.BalanceLogs, log),
		Exchange:              exchange.Init(module.Exchange, log),
		WS:                    ws.Init(log, module.Bet, module.Notification, userWS, module.User),
		Bet:                   bet.Init(module.Bet, log),
		Departments:           department.Init(module.Departments, log),
		Performance:           performance.Init(module.Performance, log),
		Authz:                 authz.Init(log, module.Authz),
		Airtime:               airtime.Init(log, module.AirtimeProvider),
		SystemLogs:            logs.Init(log, module.SystemLogs),
		Company:               company.Init(module.Company, log),
		Report:                report.Init(module.Report, log),
		Squads:                squads.Init(log, module.Squads),
		Notification:          notification.Init(log, module.Notification),
		Adds:                  adds.Init(module.Adds, log),
		Banner:                banner.Init(module.Banner, log),
		Lottery:               lottery.Init(module.Lottery, log),
		SportsService:         sportsservice.Init(module.SportsService, log),
		RiskSettings:          risksettings.Init(module.RiskSettings, log),
		Agent:                 agent.Init(module.Agent, log),
	}
}
