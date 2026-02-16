package initiator

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/adds"
	"github.com/tucanbit/internal/module/admin_activity_logs"
	"github.com/tucanbit/internal/module/agent"
	alertModule "github.com/tucanbit/internal/module/alert"
	"github.com/tucanbit/internal/module/authz"
	"github.com/tucanbit/internal/module/balance"
	"github.com/tucanbit/internal/module/balancelogs"
	"github.com/tucanbit/internal/module/banner"
	"github.com/tucanbit/internal/module/bet"
	"github.com/tucanbit/internal/module/brand"
	"github.com/tucanbit/internal/module/campaign"
	"github.com/tucanbit/internal/module/cashback"
	"github.com/tucanbit/internal/module/company"
	"github.com/tucanbit/internal/module/crypto_wallet"
	"github.com/tucanbit/internal/module/falcon_liquidity"
	"github.com/tucanbit/internal/module/game"
	"github.com/tucanbit/internal/module/groove"
	"github.com/tucanbit/internal/module/provider"

	"github.com/tucanbit/internal/module/department"
	moduleExchange "github.com/tucanbit/internal/module/exchange"
	"github.com/tucanbit/internal/module/logs"
	"github.com/tucanbit/internal/module/lottery"
	"github.com/tucanbit/internal/module/notification"
	operationdefinition "github.com/tucanbit/internal/module/operationDefinition"
	"github.com/tucanbit/internal/module/operationalgroup"
	"github.com/tucanbit/internal/module/operationalgrouptype"
	otpModule "github.com/tucanbit/internal/module/otp"
	pageModule "github.com/tucanbit/internal/module/page"
	"github.com/tucanbit/internal/module/performance"
	"github.com/tucanbit/internal/module/rakeback_override"
	"github.com/tucanbit/internal/module/report"
	"github.com/tucanbit/internal/module/risksettings"
	"github.com/tucanbit/internal/module/sportsservice"
	"github.com/tucanbit/internal/module/squads"
	userModule "github.com/tucanbit/internal/module/user"
	"github.com/tucanbit/internal/service/twofactor"
	"github.com/tucanbit/platform"
	"github.com/tucanbit/platform/pisi"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"

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
	Brand                 module.Brand
	Provider              module.Provider
	CryptoWallet          *crypto_wallet.CasinoWalletService
	Report                module.Report
	Squads                module.Squads
	Notification          module.Notification
	Campaign              module.Campaign
	Adds                  module.Adds
	Banner                module.Banner
	Lottery               module.Lottery
	SportsService         module.SportsService
	RiskSettings          module.RiskSettings
	Agent                 module.Agent
	OTP                   otpModule.OTPModule
	Cashback              *cashback.CashbackService
	CashbackKafkaConsumer *cashback.CashbackKafkaConsumer
	Groove                groove.GrooveService
	Game                  *game.GameService
	HouseEdge             *game.HouseEdgeService
	Email                 email.EmailService
	RakebackOverride      module.RakebackOverride
	Redis                 *redis.RedisOTP
	TwoFactor             twofactor.TwoFactorService
	UserBalanceWS         utils.UserWS
	AdminActivityLogs     module.AdminActivityLogs
	Page                  module.Page
}

func initModule(persistence *Persistence, log *zap.Logger, locker map[uuid.UUID]*sync.Mutex, userBalanceWs utils.UserWS, kafka platform.Kafka, redis *redis.RedisOTP, pisiClient pisi.PisiClient, alertService alertModule.AlertService) *Module {

	spApiKey := viper.GetString("sportsservice.api_key")
	apiSecret := viper.GetString("sportsservice.api_secret")
	if spApiKey == "" || apiSecret == "" {
		log.Fatal("sports service API key and secret must be set in configuration")
	}

	// Initialize agent module first
	agentModule := agent.Init(persistence.Agent, log)

	// Initialize enterprise-grade email service
	// Using smtp configuration from config.yaml
	envSmtpPassword := os.Getenv("SMTP_PASSWORD")
	envSmtpUsername := os.Getenv("SMTP_USERNAME")
	// Read from smtp configuration
	smtpPasswordRaw := viper.GetString("smtp.password")
	// Remove any potential leading/trailing quotes from YAML parsing, but keep internal spaces
	smtpPassword := strings.Trim(smtpPasswordRaw, `"'`)
	// Use smtp.username as username (email address)
	smtpUsername := strings.TrimSpace(viper.GetString("smtp.username"))
	if smtpUsername == "" {
		// Fallback to smtp.from if username is not set
		smtpUsername = strings.TrimSpace(viper.GetString("smtp.from"))
	}
	smtpHost := viper.GetString("smtp.host")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := viper.GetInt("smtp.port")
	if smtpPort == 0 {
		smtpPort = 465
	}
	smtpFrom := strings.TrimSpace(viper.GetString("smtp.from"))
	if smtpFrom == "" {
		smtpFrom = smtpUsername
	}
	smtpFromName := viper.GetString("smtp.from_name")
	smtpUseTLS := viper.GetBool("smtp.use_tls")
	log.Info("Loading SMTP configuration from smtp",
		zap.String("config_source", "smtp.* from config.yaml (viper)"),
		zap.String("password_source", "viper.GetString('smtp.password')"),
		zap.String("username_source", "viper.GetString('smtp.username') or 'smtp.from'"),
		zap.String("env_SMTP_PASSWORD_set", fmt.Sprintf("%v", envSmtpPassword != "")),
		zap.String("env_SMTP_USERNAME_set", fmt.Sprintf("%v", envSmtpUsername != "")),
		zap.String("env_SMTP_PASSWORD_length", fmt.Sprintf("%d", len(envSmtpPassword))),
		zap.String("env_SMTP_USERNAME_length", fmt.Sprintf("%d", len(envSmtpUsername))),
		zap.String("host", smtpHost),
		zap.Int("port", smtpPort),
		zap.String("username", smtpUsername),
		zap.String("username_raw", smtpUsername),
		zap.Int("username_length", len(smtpUsername)),
		zap.Bool("password_set", smtpPassword != ""),
		zap.Int("password_length", len(smtpPassword)),
		zap.String("password_raw", smtpPassword),
		zap.String("from", smtpFrom),
		zap.String("from_name", smtpFromName),
		zap.Bool("use_tls", smtpUseTLS))
	emailService, err := email.NewEmailService(email.SMTPConfig{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUsername,
		Password: smtpPassword,
		From:     smtpFrom,
		FromName: smtpFromName,
		UseTLS:   smtpUseTLS,
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
			twofactor.NewTwoFactorService(persistence.TwoFactor, persistence.Passkey, log, twofactor.TwoFactorConfig{
				Issuer:           "TucanBIT",
				Algorithm:        otp.AlgorithmSHA1,
				Digits:           otp.DigitsSix,
				Period:           30,
				BackupCodesCount: 10,
				MaxAttempts:      5,
				LockoutDuration:  15 * time.Minute,
				EnabledMethods: []twofactor.TwoFactorMethod{
					twofactor.MethodTOTP,
					twofactor.MethodEmailOTP,
					twofactor.MethodSMSOTP,
					twofactor.MethodBiometric,
					twofactor.MethodBackupCodes,
					twofactor.MethodPasskey,
				},
				EmailOTPLength:   6,
				SMSOTPLength:     6,
				OTPExpiryMinutes: 5,
			}, emailService),
			persistence.SystemConfig,
			pageModule.Init(persistence.Page, log),
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
			persistence.Alert,
			alertService,
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
		Authz:         authz.Init(log, persistence.Authz, persistence.User),
		UserBalanceWS: userBalanceWs,
		SystemLogs:    logs.Init(log, persistence.Logs),
		Company:       company.Init(persistence.Company, log),
		Brand:         brand.Init(persistence.Brand, log),
		Provider:      provider.Init(persistence.Provider, log),
		CryptoWallet: crypto_wallet.NewCasinoWalletService(
			persistence.CryptoWallet,
			persistence.User,
			persistence.Balance,
			persistence.Groove,
			persistence.Cashback,
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
		Campaign:     campaign.Init(persistence.Campaign, persistence.Notification, log),
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
		OTP:                   otpModule.NewOTPService(persistence.OTP, otpModule.NewUserStorageAdapter(persistence.User), emailService, log),
		Cashback:              cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log, persistence.RakebackOverride),
		CashbackKafkaConsumer: cashback.NewCashbackKafkaConsumer(cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log, persistence.RakebackOverride), kafka, log),
		Groove: groove.NewGrooveService(persistence.Groove, persistence.GameSession, cashback.NewCashbackService(persistence.Cashback, persistence.Groove, userBalanceWs, log, persistence.RakebackOverride), persistence.User, userBalanceWs, falcon_liquidity.NewFalconLiquidityService(log, &dto.FalconLiquidityConfig{
			Enabled:        viper.GetBool("falcon_liquidity.enabled"),
			Host:           viper.GetString("falcon_liquidity.host"),
			Port:           viper.GetInt("falcon_liquidity.port"),
			VirtualHost:    viper.GetString("falcon_liquidity.virtual_host"),
			Username:       viper.GetString("falcon_liquidity.username"),
			Password:       viper.GetString("falcon_liquidity.password"),
			ExchangeName:   viper.GetString("falcon_liquidity.exchange_name"),
			QueueName:      viper.GetString("falcon_liquidity.queue_name"),
			RoutingKey:     viper.GetString("falcon_liquidity.routing_key"),
			ClientName:     viper.GetString("falcon_liquidity.client_name"),
			ManagementPort: viper.GetInt("falcon_liquidity.management_port"),
		}, persistence.FalconMessage), log),
		Game:             game.NewGameService(persistence.Game, log),
		HouseEdge:        game.NewHouseEdgeService(persistence.HouseEdge, log),
		Email:            emailService,
		RakebackOverride: rakeback_override.Init(persistence.RakebackOverride, log),
		TwoFactor: twofactor.NewTwoFactorService(persistence.TwoFactor, persistence.Passkey, log, twofactor.TwoFactorConfig{
			Issuer:           "TucanBIT",
			Algorithm:        otp.AlgorithmSHA1,
			Digits:           otp.DigitsSix,
			Period:           30,
			BackupCodesCount: 10,
			MaxAttempts:      5,
			LockoutDuration:  15 * time.Minute,
			EnabledMethods: []twofactor.TwoFactorMethod{
				twofactor.MethodTOTP,
				twofactor.MethodEmailOTP,
				twofactor.MethodSMSOTP,
				twofactor.MethodBiometric,
				twofactor.MethodBackupCodes,
				twofactor.MethodPasskey,
			},
			EmailOTPLength:   6,
			SMSOTPLength:     6,
			OTPExpiryMinutes: 5,
		}, emailService),
		Redis:             redis,
		AdminActivityLogs: admin_activity_logs.NewAdminActivityLogsModule(persistence.AdminActivityLogs, log),
		Page:              pageModule.Init(persistence.Page, log),
	}
}
