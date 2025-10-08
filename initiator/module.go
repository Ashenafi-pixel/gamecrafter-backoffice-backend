package initiator

import (
	"sync"

	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/adds"
	"github.com/tucanbit/internal/module/agent"
	"github.com/tucanbit/internal/module/authz"
	"github.com/tucanbit/internal/module/balance"
	"github.com/tucanbit/internal/module/balancelogs"
	"github.com/tucanbit/internal/module/banner"
	"github.com/tucanbit/internal/module/bet"
	"github.com/tucanbit/internal/module/cashback"
	"github.com/tucanbit/internal/module/company"
	"github.com/tucanbit/internal/module/crypto_wallet"
	"github.com/tucanbit/internal/module/groove"

	"github.com/tucanbit/internal/module/department"
	moduleExchange "github.com/tucanbit/internal/module/exchange"
	"github.com/tucanbit/internal/module/logs"
	"github.com/tucanbit/internal/module/lottery"
	"github.com/tucanbit/internal/module/notification"
	operationdefinition "github.com/tucanbit/internal/module/operationDefinition"
	"github.com/tucanbit/internal/module/operationalgroup"
	"github.com/tucanbit/internal/module/operationalgrouptype"
	"github.com/tucanbit/internal/module/otp"
	"github.com/tucanbit/internal/module/performance"
	"github.com/tucanbit/internal/module/report"
	"github.com/tucanbit/internal/module/risksettings"
	"github.com/tucanbit/internal/module/sportsservice"
	"github.com/tucanbit/internal/module/squads"
	userModule "github.com/tucanbit/internal/module/user"
	"github.com/tucanbit/platform"
	"github.com/tucanbit/platform/pisi"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/tucanbit/internal/module/email"
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
	SystemLogs            module.SystemLogs
	Company               module.Company
	CryptoWallet          *crypto_wallet.CasinoWalletService
	Report                module.Report
	Squads                module.Squads
	Notification          module.Notification
	Adds                  module.Adds
	Banner                module.Banner
	Lottery               module.Lottery
	SportsService         module.SportsService
	RiskSettings          module.RiskSettings
	Agent                 module.Agent
	OTP                   otp.OTPModule
	Cashback              *cashback.CashbackService
	CashbackKafkaConsumer *cashback.CashbackKafkaConsumer
	Groove                groove.GrooveService
	Email                 email.EmailService
	Redis                 *redis.RedisOTP
	UserBalanceWS         utils.UserWS
}

func initModule(persistence *Persistence, log *zap.Logger, locker map[uuid.UUID]*sync.Mutex, enforcer *casbin.Enforcer, userBalanceWs utils.UserWS, kafka platform.Kafka, redis *redis.RedisOTP, pisiClient pisi.PisiClient) *Module {

	spApiKey := viper.GetString("sportsservice.api_key")
	apiSecret := viper.GetString("sportsservice.api_secret")
	if spApiKey == "" || apiSecret == "" {
		log.Fatal("sports service API key and secret must be set in configuration")
	}

	// Initialize agent module first
	agentModule := agent.Init(persistence.Agent, log)

	// Initialize enterprise-grade email service
	emailService, err := email.NewEmailService(email.SMTPConfig{
		Host:     viper.GetString("smtp.host"),
		Port:     viper.GetInt("smtp.port"),
		Username: viper.GetString("smtp.username"),
		Password: viper.GetString("smtp.password"),
		From:     viper.GetString("smtp.from"),
		FromName: viper.GetString("smtp.from_name"),
		UseTLS:   viper.GetBool("smtp.use_tls"),
	}, log)
	if err != nil {
		log.Fatal("Failed to initialize email service", zap.Error(err))
	}

	return &Module{
		User: userModule.Init(persistence.User,
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
			persistence.OTP,
			emailService,
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

		Exchange: moduleExchange.Init(persistence.Exchange, log),
		Bet: bet.Init(
			persistence.Bet,
			persistence.Analytics,
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
		Departments:   department.Init(persistence.Departments, persistence.User, log),
		Performance:   performance.Init(persistence.Performance, log),
		Authz:         authz.Init(log, persistence.Authz, persistence.User, enforcer),
		UserBalanceWS: userBalanceWs,
		SystemLogs:    logs.Init(log, persistence.Logs),
		Company:       company.Init(persistence.Company, log),
		CryptoWallet: crypto_wallet.NewCasinoWalletService(
			persistence.CryptoWallet,
			persistence.User,
			persistence.Balance,
			viper.GetString("app.jwt_secret"),
		),
		// Crypto: crypto.Init(
		// 	persistence.Crypto,
		// 	persistence.Balance,
		// 	persistence.User,
		// 	fireblocksClient,
		// 	exchangeProvider,
		// ),
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
		SportsService:         sportsservice.Init(log, spApiKey, apiSecret, persistence.Sports, persistence.BalanageLogs, persistence.OperationalGroup, persistence.OperationalGroupType),
		RiskSettings:          risksettings.Init(persistence.RiskSettings, log),
		Agent:                 agentModule,
		OTP:                   otp.NewOTPService(persistence.OTP, otp.NewUserStorageAdapter(persistence.User), emailService, log),
		Cashback:              cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log),
		CashbackKafkaConsumer: cashback.NewCashbackKafkaConsumer(cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log), kafka, log),
		Groove:                groove.NewGrooveService(persistence.Groove, persistence.GameSession, cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log), userBalanceWs, log),
		Email:                 emailService,
		Redis:                 redis,
	}
}
