package initiator

import (
	"github.com/tucanbit/internal/constant/persistencedb"
	analyticsModule "github.com/tucanbit/internal/module/analytics"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/adds"
	"github.com/tucanbit/internal/storage/admin_activity_logs"
	"github.com/tucanbit/internal/storage/agent"
	"github.com/tucanbit/internal/storage/airtime"
	"github.com/tucanbit/internal/storage/alert"
	analyticsStorage "github.com/tucanbit/internal/storage/analytics"
	"github.com/tucanbit/internal/storage/authz"
	"github.com/tucanbit/internal/storage/balance"
	"github.com/tucanbit/internal/storage/balancelogs"
	"github.com/tucanbit/internal/storage/banner"
	"github.com/tucanbit/internal/storage/bet"
	"github.com/tucanbit/internal/storage/campaign"
	"github.com/tucanbit/internal/storage/cashback"
	"github.com/tucanbit/internal/storage/company"
	"github.com/tucanbit/internal/storage/config"
	"github.com/tucanbit/internal/storage/departements"
	"github.com/tucanbit/internal/storage/exchange"
	"github.com/tucanbit/internal/storage/falcon_liquidity"
	"github.com/tucanbit/internal/storage/game"
	"github.com/tucanbit/internal/storage/groove"
	"github.com/tucanbit/internal/storage/kyc"
	"github.com/tucanbit/internal/storage/logs"
	"github.com/tucanbit/internal/storage/lottery"
	"github.com/tucanbit/internal/storage/notification"
	"github.com/tucanbit/internal/storage/operationalgroup"
	"github.com/tucanbit/internal/storage/operationalgrouptype"
	"github.com/tucanbit/internal/storage/otp"
	"github.com/tucanbit/internal/storage/passkey"
	"github.com/tucanbit/internal/storage/performance"
	"github.com/tucanbit/internal/storage/report"
	"github.com/tucanbit/internal/storage/risksettings"
	"github.com/tucanbit/internal/storage/sports"
	"github.com/tucanbit/internal/storage/squads"
	"github.com/tucanbit/internal/storage/system_config"
	"github.com/tucanbit/internal/storage/twofactor"
	"github.com/tucanbit/internal/storage/user"
	"github.com/tucanbit/platform/clickhouse"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Persistence struct {
	User                 storage.User
	OperationalGroup     storage.OperationalGroup
	OperationalGroupType storage.OperationalGroupType
	Balance              storage.Balance
	Logs                 storage.Logs
	BalanageLogs         storage.BalanceLogs
	Exchange             storage.Exchage
	Bet                  storage.Bet
	Departments          storage.Departements
	Performance          storage.Performance
	Authz                storage.Authz
	Config               storage.Config
	AirtimeProvider      storage.Airtime
	Company              storage.Company
	CryptoWallet         storage.CryptoWallet
	Report               storage.Report
	Squad                storage.Squads
	Notification         storage.Notification
	Campaign             storage.Campaign
	Adds                 storage.Adds
	Banner               storage.Banner
	Lottery              storage.Lottery
	Sports               storage.Sports
	RiskSettings         storage.RiskSettings
	Agent                storage.Agent
	OTP                  otp.OTP
	Cashback             cashback.CashbackStorage
	Groove               groove.GrooveStorage
	GameSession          groove.GameSessionStorage
	Game                 game.GameStorage
	HouseEdge            game.HouseEdgeStorage
	TwoFactor            twofactor.TwoFactorStorage
	Passkey              passkey.PasskeyStorage
	FalconMessage        falcon_liquidity.FalconMessageStorage
	Analytics            storage.Analytics
	AdminActivityLogs    admin_activity_logs.AdminActivityLogsStorage
	Alert                alert.AlertStorage
	KYC                  kyc.KYCStorage
	SystemConfig         *system_config.SystemConfig
	Database             *persistencedb.PersistenceDB
}

func initPersistence(persistencdb *persistencedb.PersistenceDB, log *zap.Logger, gormDB *gorm.DB, redis *redis.RedisOTP, userWS utils.UserWS, clickhouseClient *clickhouse.ClickHouseClient) *Persistence {
	// Create analytics storage from clickhouse client
	var analyticsStorageInstance storage.Analytics
	if clickhouseClient != nil {
		analyticsStorageInstance = analyticsStorage.NewAnalyticsStorage(clickhouseClient, log)
	} else {
		analyticsStorageInstance = nil
	}

	return &Persistence{
		User:                 user.Init(persistencdb, log),
		OperationalGroup:     operationalgroup.Init(persistencdb, log),
		OperationalGroupType: operationalgrouptype.Init(persistencdb, log),
		Logs:                 logs.Init(persistencdb, log),
		Balance:              balance.Init(persistencdb, log),
		BalanageLogs:         balancelogs.Init(persistencdb, log),
		Exchange:             exchange.Init(persistencdb, log),
		Bet:                  bet.Init(persistencdb, analyticsStorageInstance, log),
		Departments:          departements.Init(persistencdb, log),
		Performance:          performance.Init(persistencdb, log),
		Authz:                authz.Init(gormDB, log, persistencdb),
		Config:               config.Init(persistencdb, log),
		AirtimeProvider:      airtime.Init(log, persistencdb),
		Company:              company.Init(persistencdb, log),
		CryptoWallet:         storage.Init(persistencdb, log),
		Report:               report.Init(persistencdb, log),
		Squad:                squads.Init(persistencdb, log),
		Notification:         notification.Init(persistencdb, log),
		Campaign:             campaign.Init(persistencdb, log),
		Adds:                 adds.Init(persistencdb, log),
		Banner:               banner.Init(persistencdb, log),
		Lottery:              lottery.Init(persistencdb, log),
		Sports:               sports.Init(persistencdb, log),
		RiskSettings:         risksettings.Init(persistencdb, log),
		Agent:                agent.Init(persistencdb, log),
		OTP:                  otp.NewOTP(otp.NewOTPDatabase(redis, log)),
		Cashback:             cashback.NewCashbackStorage(persistencdb, log, analyticsStorage.NewAnalyticsIntegration(analyticsModule.NewRealtimeSyncService(analyticsModule.NewSyncService(analyticsStorage.NewAnalyticsStorage(clickhouseClient, log), groove.NewGrooveStorage(persistencdb, userWS, nil, log), log), analyticsStorage.NewAnalyticsStorage(clickhouseClient, log), log), log)),
		Groove:               groove.NewGrooveStorage(persistencdb, userWS, analyticsStorage.NewAnalyticsIntegration(analyticsModule.NewRealtimeSyncService(analyticsModule.NewSyncService(analyticsStorage.NewAnalyticsStorage(clickhouseClient, log), groove.NewGrooveStorage(persistencdb, userWS, nil, log), log), analyticsStorage.NewAnalyticsStorage(clickhouseClient, log), log), log), log),
		GameSession:          groove.NewGameSessionStorage(persistencdb),
		Game:                 game.NewGameStorage(persistencdb, log),
		HouseEdge:            game.NewHouseEdgeStorage(persistencdb, log),
		TwoFactor:            twofactor.Init(persistencdb, log),
		Passkey:              passkey.NewPasskeyStorage(persistencdb, log),
		FalconMessage:        falcon_liquidity.NewFalconMessageStorage(log, persistencdb),
		Analytics:            analyticsStorage.NewAnalyticsStorage(clickhouseClient, log),
		AdminActivityLogs:    admin_activity_logs.NewAdminActivityLogsStorage(*persistencdb, log),
		Alert:                alert.NewAlertStorage(*persistencdb, log),
		KYC:                  kyc.NewKYCStorage(persistencdb, log),
		SystemConfig:         system_config.NewSystemConfig(persistencdb, log),
		Database:             persistencdb,
	}
}
