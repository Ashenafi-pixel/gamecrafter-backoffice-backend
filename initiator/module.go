package initiator

import (
	"sync"

	"github.com/joshjones612/egyptkingcrash/platform/pisi"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"github.com/joshjones612/egyptkingcrash/internal/module/adds"
	"github.com/joshjones612/egyptkingcrash/internal/module/agent"
	"github.com/joshjones612/egyptkingcrash/internal/module/airtime"
	"github.com/joshjones612/egyptkingcrash/internal/module/authz"
	"github.com/joshjones612/egyptkingcrash/internal/module/balance"
	"github.com/joshjones612/egyptkingcrash/internal/module/balancelogs"
	"github.com/joshjones612/egyptkingcrash/internal/module/banner"
	"github.com/joshjones612/egyptkingcrash/internal/module/bet"
	"github.com/joshjones612/egyptkingcrash/internal/module/company"
	"github.com/joshjones612/egyptkingcrash/internal/module/department"
	"github.com/joshjones612/egyptkingcrash/internal/module/exchange"
	"github.com/joshjones612/egyptkingcrash/internal/module/logs"
	"github.com/joshjones612/egyptkingcrash/internal/module/lottery"
	"github.com/joshjones612/egyptkingcrash/internal/module/notification"
	operationdefinition "github.com/joshjones612/egyptkingcrash/internal/module/operationDefinition"
	"github.com/joshjones612/egyptkingcrash/internal/module/operationalgroup"
	"github.com/joshjones612/egyptkingcrash/internal/module/operationalgrouptype"
	"github.com/joshjones612/egyptkingcrash/internal/module/performance"
	"github.com/joshjones612/egyptkingcrash/internal/module/report"
	"github.com/joshjones612/egyptkingcrash/internal/module/risksettings"
	"github.com/joshjones612/egyptkingcrash/internal/module/sportsservice"
	"github.com/joshjones612/egyptkingcrash/internal/module/squads"
	"github.com/joshjones612/egyptkingcrash/internal/module/user"
	"github.com/joshjones612/egyptkingcrash/platform"
	"github.com/joshjones612/egyptkingcrash/platform/redis"
	"github.com/joshjones612/egyptkingcrash/platform/utils"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Module struct {
	User                  module.User
	OperationalGroup      module.OperationalGroup
	OperationalGroupType  module.OperationalGroupType
	OperationsDefinitions module.OperationsDefinitions
	Balance               module.Balance
	BalanceLogs           module.BalanceLogs
	Exchange              module.Exchange
	Bet                   module.Bet
	Departments           module.Departements
	Performance           module.Performance
	Authz                 module.Authz
	AirtimeProvider       module.AirtimeProvider
	SystemLogs            module.SystemLogs
	Company               module.Company
	Report                module.Report
	Squads                module.Squads
	Notification          module.Notification
	Adds                  module.Adds
	Banner                module.Banner
	Lottery               module.Lottery
	SportsService         module.SportsService
	RiskSettings          module.RiskSettings
	Agent                 module.Agent
}

func initModule(persistence *Persistence, log *zap.Logger, locker map[uuid.UUID]*sync.Mutex, enforcer *casbin.Enforcer, userBalanceWs utils.UserWS, kafka platform.Kafka, redis *redis.RedisOTP, pisiClient pisi.PisiClient) *Module {
	spApiKey := viper.GetString("sportsservice.api_key")
	apiSecret := viper.GetString("sportsservice.api_secret")
	if spApiKey == "" || apiSecret == "" {
		log.Fatal("sports service API key and secret must be set in configuration")
	}

	// Initialize agent module first
	agentModule := agent.Init(persistence.Agent, log)

	return &Module{
		User: user.Init(persistence.User,
			persistence.Logs,
			log, viper.GetString("aws.bucket.name"),
			persistence.Balance,
			viper.GetString("auth.otp_jwt_secret"),
			viper.GetString("google.oauth.client_id"),
			viper.GetString("google.oauth.client_secret"),
			viper.GetString("google.oauth.redirect_url"),
			viper.GetString("facebook.oauth.client_id"),
			viper.GetString("facebook.oauth.client_secret"),
			viper.GetString("facebook.oauth.redirect_url"),
			persistence.BalanageLogs,
			persistence.OperationalGroup,
			persistence.OperationalGroupType,
			persistence.Config,
			agentModule,
			redis,
			pisiClient,
		),
		OperationalGroup:      operationalgroup.Init(persistence.OperationalGroup, log),
		OperationalGroupType:  operationalgrouptype.Init(persistence.OperationalGroupType, log),
		OperationsDefinitions: operationdefinition.Init(persistence.OperationalGroup, persistence.OperationalGroupType, log),
		Balance: balance.Init(persistence.Balance,
			persistence.BalanageLogs,
			persistence.Exchange,
			persistence.User,
			persistence.OperationalGroup,
			persistence.OperationalGroupType,
			log,
			locker),
		BalanceLogs: balancelogs.Init(persistence.BalanageLogs, log),

		Exchange: exchange.Init(persistence.Exchange, log),
		Bet: bet.Init(
			persistence.Bet,
			persistence.Balance,
			log,
			decimal.NewFromInt(int64(viper.GetInt("bet.max"))),
			viper.GetDuration("bet.open_duration"),
			locker,
			persistence.OperationalGroup,
			persistence.OperationalGroupType,
			persistence.BalanageLogs,
			viper.GetString("aws.bucket.name"),
			persistence.User,
			persistence.Config,
			persistence.Squad,
			userBalanceWs,
		),
		Departments: department.Init(persistence.Departments, persistence.User, log),
		Performance: performance.Init(persistence.Performance, log),
		Authz:       authz.Init(log, persistence.Authz, enforcer),
		AirtimeProvider: airtime.Init(
			log,
			viper.GetString("airtime.base_url"),
			viper.GetString("airtime.password"),
			viper.GetDuration("airtime.timeout"),
			viper.GetInt("airtime.vaspid"),
			persistence.AirtimeProvider,
			persistence.Balance,
			persistence.BalanageLogs,
			persistence.User,
			persistence.OperationalGroup,
			persistence.OperationalGroupType,
		),
		SystemLogs:   logs.Init(log, persistence.Logs),
		Company:      company.Init(persistence.Company, log),
		Report:       report.Init(persistence.Report, log),
		Squads:       squads.Init(log, persistence.Squad, persistence.User),
		Notification: notification.Init(persistence.Notification, log),
		Adds:         adds.Init(persistence.Adds, persistence.Balance, persistence.BalanageLogs, log),
		Banner:       banner.Init(persistence.Banner, log, viper.GetString("aws.bucket.name")),
		Lottery: lottery.Init(
			persistence.Lottery,
			log,
			kafka,
			persistence.BalanageLogs,
			persistence.Balance,
			persistence.OperationalGroup,
			persistence.OperationalGroupType,
		),

		// Initialize the sports service module with the logger
		SportsService: sportsservice.Init(log, spApiKey, apiSecret, persistence.Sports, persistence.BalanageLogs, persistence.OperationalGroup, persistence.OperationalGroupType),
		RiskSettings:  risksettings.Init(persistence.RiskSettings, log),
		Agent:         agentModule,
	}
}
