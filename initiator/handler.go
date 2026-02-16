package initiator

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/adds"
	"github.com/tucanbit/internal/handler/admin_activity_logs"
	"github.com/tucanbit/internal/handler/agent"
	"github.com/tucanbit/internal/handler/airtime"
	"github.com/tucanbit/internal/handler/alert"
	analyticsHandler "github.com/tucanbit/internal/handler/analytics"
	"github.com/tucanbit/internal/handler/authz"
	"github.com/tucanbit/internal/handler/balance"
	"github.com/tucanbit/internal/handler/balancelogs"
	"github.com/tucanbit/internal/handler/banner"
	"github.com/tucanbit/internal/handler/bet"
	"github.com/tucanbit/internal/handler/brand"
	"github.com/tucanbit/internal/handler/campaign"
	"github.com/tucanbit/internal/handler/cashback"
	"github.com/tucanbit/internal/handler/company"
	"github.com/tucanbit/internal/handler/department"
	"github.com/tucanbit/internal/handler/exchange"
	"github.com/tucanbit/internal/handler/game"
	"github.com/tucanbit/internal/handler/groove"
	"github.com/tucanbit/internal/handler/kyc"
	"github.com/tucanbit/internal/handler/logs"
	"github.com/tucanbit/internal/handler/lottery"
	"github.com/tucanbit/internal/handler/notification"
	"github.com/tucanbit/internal/handler/operationalgroup"
	"github.com/tucanbit/internal/handler/operationalgrouptype"
	"github.com/tucanbit/internal/handler/operationsdefinitions"
	"github.com/tucanbit/internal/handler/otp"
	"github.com/tucanbit/internal/handler/page"
	"github.com/tucanbit/internal/handler/performance"
	"github.com/tucanbit/internal/handler/provider"
	"github.com/tucanbit/internal/handler/rakeback_override"
	"github.com/tucanbit/internal/handler/report"
	"github.com/tucanbit/internal/handler/risksettings"
	"github.com/tucanbit/internal/handler/sportsservice"
	"github.com/tucanbit/internal/handler/squads"
	"github.com/tucanbit/internal/handler/system_config"
	"github.com/tucanbit/internal/handler/twofactor"
	"github.com/tucanbit/internal/handler/user"
	"github.com/tucanbit/internal/handler/withdrawal_management"
	"github.com/tucanbit/internal/handler/withdrawals"
	"github.com/tucanbit/internal/handler/ws"
	gameImportModule "github.com/tucanbit/internal/module/game_import"
	"github.com/tucanbit/internal/storage"
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
	Brand                 handler.Brand
	Provider              handler.Provider
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
	RakebackOverride      handler.RakebackOverride
	RegistrationService   *user.RegistrationService
	Campaign              handler.Campaign
	TwoFactor             handler.TwoFactor
	Analytics             handler.Analytics
	SystemConfig          *system_config.SystemConfigHandler
	Alert                 alert.AlertHandler
	AlertEmailGroup       alert.AlertEmailGroupHandler
	WithdrawalManagement  *withdrawal_management.WithdrawalManagementHandler
	Withdrawals           *withdrawals.WithdrawalsHandler
	AdminActivityLogs     handler.AdminActivityLogs
	AirtimeProvider       handler.AirtimeProvider
	KYC                   handler.KYC
	Page                  page.PageHandler
}

func initHandler(module *Module, persistence *Persistence, log *zap.Logger, userWS utils.UserWS) *Handler {
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
		Brand:                 brand.Init(module.Brand, log),
		Provider:              provider.Init(module.Provider, log),
		Report:                report.Init(module.Report, module.User, log),
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
		RakebackOverride:      rakeback_override.NewRakebackOverrideHandler(module.RakebackOverride, log),
		RegistrationService:   registrationService,
		Campaign:              campaign.Init(module.Campaign, log),
		TwoFactor:             twofactor.NewTwoFactorHandler(module.TwoFactor, log),
		Analytics:             analyticsHandler.Init(log, persistence.Analytics, persistence.Database.GetPool()),
		SystemConfig: func() *system_config.SystemConfigHandler {
			directusURL := viper.GetString("directus.api_url")
			if directusURL == "" {
				directusURL = "https://tucanbit-prod.directus.app/graphql"
			}
			gameImportService := gameImportModule.NewGameImportService(
				persistence.Database,
				persistence.Game,
				persistence.HouseEdge,
				persistence.SystemConfig,
				directusURL,
				log,
			)
			return system_config.NewSystemConfigHandler(persistence.Database, persistence.AdminActivityLogs, persistence.Alert, gameImportService, log)
		}(),
		Alert:                alert.NewAlertHandler(persistence.Alert, persistence.AlertEmailGroups, module.Email, log),
		AlertEmailGroup:      alert.NewAlertEmailGroupHandler(persistence.AlertEmailGroups, log),
		WithdrawalManagement: withdrawal_management.NewWithdrawalManagementHandler(persistence.Database, log),
		Withdrawals:          withdrawals.NewWithdrawalsHandler(persistence.Database, log),
		AdminActivityLogs:    admin_activity_logs.NewAdminActivityLogsHandler(module.AdminActivityLogs, log),
		AirtimeProvider:      airtime.Init(log, persistence.AirtimeProvider),
		KYC:                  kyc.NewKYCHandler(persistence.KYC, persistence.AdminActivityLogs, log),
		Page:                 page.Init(module.Page, log),
	}
}

type airtimeAdapter struct {
	storage storage.Airtime
}

func (a *airtimeAdapter) RefereshUtilies(c *gin.Context) (interface{}, error) {
	return map[string]interface{}{"message": "Airtime utilities refreshed"}, nil
}

func (a *airtimeAdapter) GetAvailableAirtimeUtilities(c *gin.Context, req dto.GetRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}, "total": 0}, nil
}

func (a *airtimeAdapter) UpdateAirtimeStatus(c *gin.Context, req dto.UpdateAirtimeStatusReq) (interface{}, error) {
	return map[string]interface{}{"message": "Airtime status updated"}, nil
}

func (a *airtimeAdapter) UpdateUtilityPrice(c *gin.Context, req dto.UpdateAirtimeUtilityPriceReq) (interface{}, error) {
	return map[string]interface{}{"message": "Airtime utility price updated"}, nil
}

func (a *airtimeAdapter) ClaimPoints(c *gin.Context, req dto.ClaimPointsReq) (interface{}, error) {

	return map[string]interface{}{"message": "Points claimed successfully"}, nil
}

func (a *airtimeAdapter) GetActiveAvailableAirtime(c *gin.Context, req dto.GetRequest) (interface{}, error) {

	return map[string]interface{}{"data": []interface{}{}, "total": 0}, nil
}

func (a *airtimeAdapter) GetUserAirtimeTransactions(c *gin.Context, req dto.GetRequest, userID uuid.UUID) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}, "total": 0}, nil
}

func (a *airtimeAdapter) GetAirtimeUtilitiesStats(c *gin.Context) (interface{}, error) {

	return map[string]interface{}{"total_utilities": 0, "active_utilities": 0}, nil
}

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

	_, err := r.client.Get(ctx, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}
