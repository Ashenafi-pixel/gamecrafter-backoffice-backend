package initiator

import (
	"github.com/joshjones612/egyptkingcrash/internal/constant/persistencedb"
	"github.com/joshjones612/egyptkingcrash/internal/storage"
	"github.com/joshjones612/egyptkingcrash/internal/storage/adds"
	"github.com/joshjones612/egyptkingcrash/internal/storage/agent"
	"github.com/joshjones612/egyptkingcrash/internal/storage/airtime"
	"github.com/joshjones612/egyptkingcrash/internal/storage/authz"
	"github.com/joshjones612/egyptkingcrash/internal/storage/balance"
	"github.com/joshjones612/egyptkingcrash/internal/storage/balancelogs"
	"github.com/joshjones612/egyptkingcrash/internal/storage/banner"
	"github.com/joshjones612/egyptkingcrash/internal/storage/bet"
	"github.com/joshjones612/egyptkingcrash/internal/storage/company"
	"github.com/joshjones612/egyptkingcrash/internal/storage/config"
	"github.com/joshjones612/egyptkingcrash/internal/storage/departements"
	"github.com/joshjones612/egyptkingcrash/internal/storage/exchange"
	"github.com/joshjones612/egyptkingcrash/internal/storage/logs"
	"github.com/joshjones612/egyptkingcrash/internal/storage/lottery"
	"github.com/joshjones612/egyptkingcrash/internal/storage/notification"
	"github.com/joshjones612/egyptkingcrash/internal/storage/operationalgroup"
	"github.com/joshjones612/egyptkingcrash/internal/storage/operationalgrouptype"
	"github.com/joshjones612/egyptkingcrash/internal/storage/performance"
	"github.com/joshjones612/egyptkingcrash/internal/storage/report"
	"github.com/joshjones612/egyptkingcrash/internal/storage/risksettings"
	"github.com/joshjones612/egyptkingcrash/internal/storage/sports"
	"github.com/joshjones612/egyptkingcrash/internal/storage/squads"
	"github.com/joshjones612/egyptkingcrash/internal/storage/user"
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
	Report               storage.Report
	Squad                storage.Squads
	Notification         storage.Notification
	Adds                 storage.Adds
	Banner               storage.Banner
	Lottery              storage.Lottery
	Sports               storage.Sports
	RiskSettings         storage.RiskSettings
	Agent                storage.Agent
}

func initPersistence(persistencdb *persistencedb.PersistenceDB, log *zap.Logger, gormDB *gorm.DB) *Persistence {
	return &Persistence{
		User:                 user.Init(persistencdb, log),
		OperationalGroup:     operationalgroup.Init(persistencdb, log),
		OperationalGroupType: operationalgrouptype.Init(persistencdb, log),
		Logs:                 logs.Init(persistencdb, log),
		Balance:              balance.Init(persistencdb, log),
		BalanageLogs:         balancelogs.Init(persistencdb, log),
		Exchange:             exchange.Init(persistencdb, log),
		Bet:                  bet.Init(persistencdb, log),
		Departments:          departements.Init(persistencdb, log),
		Performance:          performance.Init(persistencdb, log),
		Authz:                authz.Init(gormDB, log, persistencdb),
		Config:               config.Init(persistencdb, log),
		AirtimeProvider:      airtime.Init(log, persistencdb),
		Company:              company.Init(persistencdb, log),
		Report:               report.Init(persistencdb, log),
		Squad:                squads.Init(persistencdb, log),
		Notification:         notification.Init(persistencdb, log),
		Adds:                 adds.Init(persistencdb, log),
		Banner:               banner.Init(persistencdb, log),
		Lottery:              lottery.Init(persistencdb, log),
		Sports:               sports.Init(persistencdb, log),
		RiskSettings:         risksettings.Init(persistencdb, log),
		Agent:                agent.Init(persistencdb, log),
	}
}
