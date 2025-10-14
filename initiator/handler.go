package initiator

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/adds"
	"github.com/tucanbit/internal/handler/agent"
	analyticsHandler "github.com/tucanbit/internal/handler/analytics"
	"github.com/tucanbit/internal/handler/authz"
	"github.com/tucanbit/internal/handler/balance"
	"github.com/tucanbit/internal/handler/balancelogs"
	"github.com/tucanbit/internal/handler/banner"
	"github.com/tucanbit/internal/handler/bet"
	"github.com/tucanbit/internal/handler/campaign"
	"github.com/tucanbit/internal/handler/cashback"
	"github.com/tucanbit/internal/handler/company"
	"github.com/tucanbit/internal/handler/department"
	"github.com/tucanbit/internal/handler/exchange"
	"github.com/tucanbit/internal/handler/game"
	"github.com/tucanbit/internal/handler/groove"
	"github.com/tucanbit/internal/handler/logs"
	"github.com/tucanbit/internal/handler/lottery"
	"github.com/tucanbit/internal/handler/notification"
	"github.com/tucanbit/internal/handler/operationalgroup"
	"github.com/tucanbit/internal/handler/operationalgrouptype"
	"github.com/tucanbit/internal/handler/operationsdefinitions"
	"github.com/tucanbit/internal/handler/otp"
	"github.com/tucanbit/internal/handler/performance"
	"github.com/tucanbit/internal/handler/report"
	"github.com/tucanbit/internal/handler/risksettings"
	"github.com/tucanbit/internal/handler/sportsservice"
	"github.com/tucanbit/internal/handler/squads"
	"github.com/tucanbit/internal/handler/twofactor"
	"github.com/tucanbit/internal/handler/user"
	"github.com/tucanbit/internal/handler/ws"
	analyticsModule "github.com/tucanbit/internal/module/analytics"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

type Handler struct {
	User                  user.UserHandler
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
	OTP                   handler.OTP
	Cashback              *cashback.CashbackHandler
	Groove                *groove.GrooveHandler
	Game                  *game.GameHandler
	HouseEdge             *game.HouseEdgeHandler
	RegistrationService   *user.RegistrationService
	Campaign              handler.Campaign
	TwoFactor             handler.TwoFactor
	Analytics             handler.Analytics
}

func initHandler(module *Module, persistence *Persistence, log *zap.Logger, userWS utils.UserWS, dailyReportService analyticsModule.DailyReportService, dailyReportCronjobService analyticsModule.DailyReportCronjobService) *Handler {
	// Create Redis adapter for RegistrationService
	redisAdapter := &redisAdapter{client: module.Redis}

	// Initialize RegistrationService for email verification
	registrationService := user.NewRegistrationService(
		module.User,
		module.OTP,
		module.Email,
		persistence.Balance,
		module.Cashback,
		module.Groove,
		redisAdapter,
		log,
	)

	// Initialize user handler
	userHandler := user.Init(module.User, log, viper.GetString("oauth.frontend_oauth_handler_url"))

	// Set the registration service in the user handler
	userHandler.SetRegistrationService(registrationService)

	return &Handler{
		User:                  userHandler,
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
		Authz:                 authz.Init(log, module.Authz, module.CryptoWallet),
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
		OTP:                   otp.NewOTPHandler(module.OTP, log),
		Cashback:              cashback.NewCashbackHandler(module.Cashback, log),
		Groove:                groove.NewGrooveHandler(module.Groove, persistence.User, persistence.Balance, persistence.Groove, persistence.Database, log),
		Game:                  game.NewGameHandler(module.Game),
		HouseEdge:             game.NewHouseEdgeHandler(module.HouseEdge),
		RegistrationService:   registrationService,
		Campaign:              campaign.Init(module.Campaign, log),
		TwoFactor:             twofactor.NewTwoFactorHandler(module.TwoFactor, log),
		Analytics:             analyticsHandler.Init(log, persistence.Analytics, dailyReportService, dailyReportCronjobService),
	}
}

// redisAdapter adapts the platform Redis client to the RegistrationService interface
type redisAdapter struct {
	client *redis.RedisOTP
}

func (r *redisAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration)
}

func (r *redisAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

func (r *redisAdapter) Delete(ctx context.Context, key string) error {
	return r.client.Delete(ctx, key)
}

func (r *redisAdapter) Exists(ctx context.Context, key string) (bool, error) {
	// Since the platform Redis doesn't have Exists, we'll try to get the key
	// If it returns an error, the key doesn't exist
	_, err := r.client.Get(ctx, key)
	if err != nil {
		return false, nil // Key doesn't exist
	}
	return true, nil // Key exists
}
