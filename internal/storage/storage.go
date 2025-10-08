package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/db"
)

type User interface {
	CreateUser(ctx context.Context, userRequest dto.User) (dto.User, error)
	GetUserByUserName(ctx context.Context, username string) (dto.User, bool, error)
	GetUserByPhoneNumber(ctx context.Context, phone string) (dto.User, bool, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (dto.User, bool, error)
	UpdateProfilePicuter(ctx context.Context, userID uuid.UUID, filename string) (string, error)
	UpdatePassword(ctx context.Context, UserID uuid.UUID, newPassword string) (dto.User, error)
	GetUserByEmail(ctx context.Context, email string) (dto.User, bool, error)
	SaveUserOTP(ctx context.Context, otpReq dto.ForgetPasswordOTPReq) error
	GetUserOTP(ctx context.Context, userID uuid.UUID) (dto.OTPHolder, bool, error)
	DeleteOTP(ctx context.Context, userID uuid.UUID) error
	UpdateUser(ctx context.Context, updateProfile dto.UpdateProfileReq) (dto.User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) (dto.User, error)
	UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, verified bool) (dto.User, error)
	BlockAccount(ctx context.Context, account dto.AccountBlockReq) (dto.AccountBlockReq, error)
	GetBlockedAccountByType(ctx context.Context, userID uuid.UUID, tp, duration string) (dto.AccountBlockReq, bool, error)
	GetBlockedAccountByUserID(ctx context.Context, userID uuid.UUID) ([]dto.AccountBlockReq, bool, error)
	GetBlockedAccountByUserIDWithPagination(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.SuspensionHistory, error)
	GetBalanceLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.BalanceLog, error)
	AaccountUnlock(ctx context.Context, ID uuid.UUID) (dto.AccountBlockReq, error)
	GetBlockedAllAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByDurationAndTypeAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByDurationAndTypeAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByDurationAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByTypeAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByTypeAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByDurationAndUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetBlockedByUserIDAccount(ctx context.Context, getBlockedAcReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, bool, error)
	GetUsersByDepartmentNotificationTypes(ctx context.Context, notificationTypes []string) ([]dto.GetUsersForNotificationRes, error)
	AddIpFilter(ctx context.Context, ipFilter dto.IpFilterReq) (dto.IPFilterRes, error)
	GetIPFilterByType(ctx context.Context, tp string) ([]dto.IPFilter, bool, error)
	GetIpFilterByTypeWithLimitAndOffset(ctx context.Context, ipFilter dto.GetIPFilterReq) (dto.GetIPFilterRes, bool, error)
	GetAllIpFilterWithLimitAndOffset(ctx context.Context, ipFilter dto.GetIPFilterReq) (dto.GetIPFilterRes, bool, error)
	GetAllUsers(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error)
	RemoveIPFilters(ctx context.Context, id uuid.UUID) (dto.RemoveIPBlockRes, error)
	GetIPFilterByID(ctx context.Context, id uuid.UUID) (dto.IPFilter, bool, error)
	CreateReferalCodeMultiplier(ctx context.Context, req dto.ReferalMultiplierReq) (dto.ReferalData, error)
	GetReferalMultiplier(ctx context.Context) (dto.ReferalData, bool, error)
	UpdateReferalMultiplier(ctx context.Context, mul decimal.Decimal) (dto.ReferalData, error)
	GetUserPointsByReferalPoint(ctx context.Context, referal string) (dto.UserPoint, bool, error)
	UpdateUserPointByUserID(ctx context.Context, userID uuid.UUID, points decimal.Decimal) error
	GetUsersDoseNotHaveReferalCode(ctx context.Context) ([]dto.User, error)
	AddReferalCode(ctx context.Context, userID uuid.UUID, referalCode string) error
	GetUserReferalUsersByUserID(ctx context.Context, userID uuid.UUID) (dto.MyRefferedUsers, bool, error)
	GetCurrentReferralMultiplier(ctx context.Context) (int, error)
	UpdateReferralMultiplier(ctx context.Context, newValue int) (dto.ReferalData, error)
	GetAdminAssignedPoints(ctx context.Context, limit, offset int) (dto.GetAdminAssignedResp, bool, error)
	GetUserPoints(ctx context.Context, userID uuid.UUID) (decimal.Decimal, bool, error)
	UpdateUserPoints(ctx context.Context, userID uuid.UUID, points decimal.Decimal) (decimal.Decimal, error)
	CreateUserPoint(ctx context.Context, userID uuid.UUID, points decimal.Decimal) (dto.UserPoint, error)
	UpdateIpFilter(ctx context.Context, req dto.IPFilter) (dto.IPFilter, error)
	GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
	GetAdminsByRole(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
	GetAdminsByStatus(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
	GetAdminsByRoleAndStatus(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
	GetUserByReferalCode(ctx context.Context, code string) (*dto.UserProfile, error)
	GetUsersByEmailAndPhone(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error)
	DeleteTempData(ctx context.Context, ID uuid.UUID) error
	GetTempData(ctx context.Context, userID uuid.UUID) (dto.GetUserReferals, error)
	SaveToTemp(ctx context.Context, req dto.UserReferals) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	ValidateUniqueConstraints(ctx context.Context, userRequest dto.User) error
}

type Analytics interface {
	// Transaction methods
	InsertTransaction(ctx context.Context, transaction *dto.AnalyticsTransaction) error
	InsertTransactions(ctx context.Context, transactions []*dto.AnalyticsTransaction) error
	GetUserTransactions(ctx context.Context, userID uuid.UUID, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)
	GetGameTransactions(ctx context.Context, gameID string, filters *dto.TransactionFilters) ([]*dto.AnalyticsTransaction, error)

	// Analytics methods
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, dateRange *dto.DateRange) (*dto.UserAnalytics, error)
	GetGameAnalytics(ctx context.Context, gameID string, dateRange *dto.DateRange) (*dto.GameAnalytics, error)
	GetSessionAnalytics(ctx context.Context, sessionID string) (*dto.SessionAnalytics, error)

	// Reporting methods
	GetDailyReport(ctx context.Context, date time.Time) (*dto.DailyReport, error)
	GetEnhancedDailyReport(ctx context.Context, date time.Time) (*dto.EnhancedDailyReport, error)
	GetMonthlyReport(ctx context.Context, year int, month int) (*dto.MonthlyReport, error)
	GetTopGames(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.GameStats, error)
	GetTopPlayers(ctx context.Context, limit int, dateRange *dto.DateRange) ([]*dto.PlayerStats, error)

	// Real-time methods
	GetRealTimeStats(ctx context.Context) (*dto.RealTimeStats, error)
	GetUserBalanceHistory(ctx context.Context, userID uuid.UUID, hours int) ([]*dto.BalanceSnapshot, error)
	InsertBalanceSnapshot(ctx context.Context, snapshot *dto.BalanceSnapshot) error
}

type Balance interface {
	CreateBalance(ctx context.Context, createBalanceReq dto.Balance) (dto.Balance, error)
	GetUserBalanaceByUserID(ctx context.Context, getBalanceReq dto.Balance) (dto.Balance, bool, error)
	UpdateBalance(ctx context.Context, updatedBalance dto.Balance) (dto.Balance, error)
	GetBalancesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error)
	UpdateMoney(ctx context.Context, updateReq dto.UpdateBalanceReq) (dto.Balance, error)
	SaveManualFunds(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error)
	GetManualFundLogs(ctx context.Context, filter dto.GetManualFundReq) (dto.GetManualFundRes, error)
}

type OperationalGroup interface {
	CreateOperationalGroup(ctx context.Context, op dto.OperationalGroup) (dto.OperationalGroup, error)
	GetOperationalGroupByName(ctx context.Context, name string) (dto.OperationalGroup, bool, error)
	GetOperationalGroups(ctx context.Context) ([]dto.OperationalGroup, bool, error)
	GetOperationalGroupByID(ctx context.Context, groupID uuid.UUID) (dto.OperationalGroup, bool, error)
}

type OperationalGroupType interface {
	CreateOperationalType(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, error)
	GetOperationalGroupByGroupIDandName(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, bool, error)
	GetOperationalGroupTypeByGroupID(ctx context.Context, groupID uuid.UUID) ([]dto.OperationalGroupType, bool, error)
	GetOperationalGroupTypes(ctx context.Context) ([]dto.OperationalGroupType, bool, error)
	GetOperationalGroupTypeByID(ctx context.Context, groupTypeID uuid.UUID) (dto.OperationalGroupType, bool, error)
}
type Logs interface {
	CreateLoginAttempts(ctx context.Context, loginAttemptReq dto.LoginAttempt) (dto.LoginAttempt, error)
	CreateLoginSessions(ctx context.Context, userSessionReq dto.UserSessions) (dto.UserSessions, error)
	GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (dto.UserSessions, error)
	UpdateUserSessionRefreshToken(ctx context.Context, sessionID uuid.UUID, newToken string, newExpiry time.Time) error
	InvalidateOldUserSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	CreateSystemLogs(ctx context.Context, systemLogReq dto.SystemLogs) (dto.SystemLogs, error)
	GetSystemLogs(ctx context.Context, req dto.GetRequest) (dto.SystemLogsRes, error)
	GetSystemLogsByStartAndEndDate(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetSystemLogsByEndDate(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetSystemLogsByStartData(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetSystemLogsByUser(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetSystemLogsByModule(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetAvailableModules(ctx context.Context) ([]string, error)
	GetSessionsExpiringSoon(ctx context.Context, duration time.Duration) ([]dto.UserSession, error)
	InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

type BalanceLogs interface {
	SaveBalanceLogs(ctx context.Context, blanceLogReq dto.BalanceLogs) (dto.BalanceLogs, error)
	GetBalanceLog(ctx context.Context, blangLogsReq dto.GetBalanceLogReq) (dto.GetBalanceLogRes, error)
	GetBalanceLogByID(ctx context.Context, balanceLogID uuid.UUID) (dto.BalanceLogsRes, error)
	DeleteBalanceLog(ctx context.Context, balanceLogID uuid.UUID) error
	GetBalanceLogsForAdmin(ctx context.Context, req dto.AdminGetBalanceLogsReq) (dto.AdminGetBalanceLogsRes, error)
	GetBalanceLogByTransactionID(ctx context.Context, transactionID string) (dto.BalanceLogsRes, error)
}

type Exchage interface {
	GetExchange(ctx context.Context, exchangeReq dto.ExchangeReq) (dto.ExchangeRes, bool, error)
	GetCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error)
	UpdateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error)
	GetAvailableCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error)
	CreateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error)
}

type Bet interface {
	SaveRounds(ctx context.Context, betroundReq dto.BetRound) (dto.BetRound, error)
	GetBetRoundsByStatus(ctx context.Context, status string) ([]dto.BetRound, bool, error)
	UpdateRoundStatusByID(ctx context.Context, roundID uuid.UUID, status string) (dto.BetRound, error)
	CloseRound(ctx context.Context, roundID uuid.UUID) (dto.BetRound, error)
	GetRoundByID(ctx context.Context, roundID uuid.UUID) (dto.BetRound, bool, error)
	SaveUserBet(ctx context.Context, betReq dto.Bet) (dto.Bet, error)
	GetUserBetByUserIDAndRoundID(ctx context.Context, betReq dto.Bet) ([]dto.Bet, bool, error)
	CashOut(ctx context.Context, cashOut dto.SaveCashoutReq) (dto.Bet, error)
	ReverseCashOut(ctx context.Context, betID uuid.UUID) error
	GetBetHistory(ctx context.Context, getBetHistoryReq dto.GetBetHistoryReq) (dto.BetHistoryResp, bool, error)
	UpdateBetStatus(ctx context.Context, betID uuid.UUID, status string) (dto.Bet, error)
	GetLeaders(ctx context.Context) (dto.LeadersResp, error)
	GetBetHistoryByUserID(ctx context.Context, getBetHistoryReq dto.GetBetHistoryReq) (dto.BetHistoryResp, bool, error)
	GetFailedRounds(ctx context.Context) ([]dto.BetRound, error)
	SaveFailedBetsLogaAuto(ctx context.Context, req dto.SaveFailedBetsLog) error
	GetAllFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error)
	GetNotRefundedFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error)
	GetUserActiveBetWithRound(ctx context.Context, userID, roundID uuid.UUID) (dto.BetRound, bool, error)
	SavePlinkoBet(ctx context.Context, req dto.PlacePlinkoGame) (dto.PlacePlinkoGame, error)
	GetUserPlinkoBetHistoryByID(ctx context.Context, req dto.PlinkoBetHistoryReq) (dto.PlinkoBetHistoryRes, bool, error)
	GetPlinkoGameStats(ctx context.Context, userID uuid.UUID) (dto.PlinkoGameStatRes, bool, error)
	CreateLeague(ctx context.Context, league dto.League) (dto.League, error)
	GetLeagues(ctx context.Context, req dto.GetRequest) (dto.GetLeagueRes, error)
	GetLeagueByID(ctx context.Context, ID uuid.UUID) (dto.League, bool, error)
	CreateClub(ctx context.Context, club dto.Club) (dto.Club, error)
	GetClubs(ctx context.Context, req dto.GetRequest) (dto.GetClubRes, error)
	CreateFootballCardMultiplier(ctx context.Context, m decimal.Decimal) (dto.Config, error)
	UpdateFootballCardMultiplierValue(ctx context.Context, m decimal.Decimal) (dto.Config, error)
	GetFootballCardMultiplier(ctx context.Context) (dto.Config, bool, error)
	GetFootballRoundByID(ctx context.Context, id uuid.UUID) (dto.FootballMatchRound, bool, error)
	CreateFootBallMatchRound(ctx context.Context, req dto.FootballMatchRound) (dto.FootballMatchRound, error)
	GetFootballMatchRound(ctx context.Context, req dto.GetRequest) (dto.GetFootballMatchRoundRes, bool, error)
	GetFootballMatchRoundByStatus(ctx context.Context, req dto.GetFootballMatchRoundsByStatusReq) ([]dto.FootballMatchRound, bool, error)
	CreateFootballMatch(ctx context.Context, req dto.FootballMatch) (dto.FootballMatch, error)
	GetClubByID(ctx context.Context, ID uuid.UUID) (dto.Club, bool, error)
	GetFootballRoundMatchs(ctx context.Context, req dto.GetFootballRoundMatchesReq) (dto.GetFootballRoundMatchesRes, bool, error)
	UpdateFootballMatchRoundStatus(ctx context.Context, req dto.FootballMatchRoundUpdateReq) (dto.FootballMatchRound, error)
	CloseFootballMatch(ctx context.Context, req dto.CloseFootballMatchReq) (dto.FootballMatch, error)
	GetFootballMatchByID(ctx context.Context, ID uuid.UUID) (dto.FootballMatch, bool, error)
	SetFootballMatchPrice(ctx context.Context, config dto.Config) (dto.Config, error)
	GetFootballMatchPrice(ctx context.Context) (dto.Config, error)
	CreateFootballBet(ctx context.Context, req dto.UserFootballMatcheRound) (dto.UserFootballMatcheRound, error)
	CreateFootballBetUserSelection(ctx context.Context, req dto.UserFootballMatchSelection) (dto.UserFootballMatchSelection, error)
	GetFootballMatchesByRoundID(ctx context.Context, roundID uuid.UUID) ([]dto.FootballMatch, bool, error)
	UpdateUserFootballMatchStatusByMatchID(ctx context.Context, status string, ID uuid.UUID) (dto.UserFootballMatchSelection, error)
	GetUserFootballMatchSelectionsByMatchID(ctx context.Context, matchID uuid.UUID) ([]dto.UserFootballMatchSelection, bool, error)
	GetFootballMatchesByStatus(ctx context.Context, status string, roundID uuid.UUID) ([]dto.FootballMatch, bool, error)
	GetAllFootBallMatchByRoundByRoundID(ctx context.Context, roundID uuid.UUID) ([]dto.UserFootballMatcheRound, bool, error)
	GetAllUserFootballBetByStatusAndRoundID(ctx context.Context, roundID uuid.UUID, status string) ([]dto.UserFootballMatchSelection, bool, error)
	UpdateUserFootballMatcheRoundsByID(ctx context.Context, roundID uuid.UUID, status, wonStatus string) error
	UpdateFootballmatchByRoundID(ctx context.Context, roundID uuid.UUID, status string) error
	GetUserFootballBets(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetUserFootballBetRes, error)
	CreateStreetKingsGame(ctx context.Context, req dto.CreateStreetKingsReqData) (dto.CreateStreetKingsRespData, error)
	CloseStreetKingsCrash(ctx context.Context, req dto.StreetKingsCashoutReq)
	GetStreetKingsCrashByID(ctx context.Context, ID uuid.UUID) (dto.StreetKingsCrashResp, error)
	GetStreetKingsGamesByUserIDAndVersion(ctx context.Context, req dto.GetStreetkingHistoryReq, userID uuid.UUID) (dto.GetStreetkingHistoryRes, bool, error)
	CreateCryptoKings(ctx context.Context, req dto.CreateCryptoKingData) (dto.CreateCryptoKingData, error)
	GetCrytoKingsBetHistoryByUserID(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetCryptoKingsUserBetHistoryRes, bool, error)
	CreateQuickHustleBet(ctx context.Context, req dto.CreateQuickHustleBetReq) (dto.CreateQuickHustelBetResData, error)
	CloseQuickHustelBet(ctx context.Context, req dto.CloseQuickHustleBetData) (dto.QuickHustelBetResData, error)
	GetQuickHustleByID(ctx context.Context, ID uuid.UUID) (dto.QuickHustelBetResData, bool, error)
	GetQuickHustleBetHistoryByUserID(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetQuickHustleResp, bool, error)
	CreateRollDaDice(ctx context.Context, req dto.CreateRollDaDiceReq) (dto.RollDaDiceData, error)
	GetUserRollDicBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetRollDaDiceRespData, bool, error)
	CreateScratchCardsBet(ctx context.Context, req dto.ScratchCardsBetData) (dto.ScratchCardsBetData, error)
	GetUserScratchCardBetHistories(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetScratchBetHistoriesResp, bool, error)
	CreateSpinningWheel(ctx context.Context, req dto.SpinningWheelData) (dto.SpinningWheelData, error)
	GetSpinningWheelUserBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetSpinningWheelHistoryResp, bool, error)
	CreateGame(ctx context.Context, req dto.Game) (dto.Game, error)
	GetGames(ctx context.Context, req dto.GetRequest) (dto.GetGamesResp, error)
	GetAllGames(ctx context.Context) (dto.GetGamesResp, error)
	GetGameByID(ctx context.Context, ID uuid.UUID) (dto.Game, error)
	UpdageGame(ctx context.Context, game dto.Game) (dto.Game, error)
	ListInactiveGames(ctx context.Context) ([]dto.Game, error)
	DeleteGame(ctx context.Context, ID uuid.UUID) error
	AddGame(ctx context.Context, ID uuid.UUID) (dto.Game, error)
	UpdateEnableStatus(ctx context.Context, game dto.Game) (dto.Game, error)
	CreateSpinningWheelMystery(ctx context.Context, req dto.CreateSpinningWheelMysteryReq) (dto.CreateSpinningWheelMysteryRes, error)
	GetSpinningWheelMystery(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelMysteryRes, error)
	UpdateSpinningWheelMystery(ctx context.Context, req dto.UpdateSpinningWheelMysteryReq) (dto.UpdateSpinningWheelMysteryRes, error)
	DeleteSpinningWheelMystery(ctx context.Context, id uuid.UUID) error
	CreateSpinningWheelConfig(ctx context.Context, req dto.CreateSpinningWheelConfigReq) (dto.CreateSpinningWheelConfigRes, error)
	GetSpinningWheelConfig(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelConfigRes, error)
	GetAllSpinningWheelConfigs(ctx context.Context) ([]dto.SpinningWheelConfigData, error)
	GetAllSpinningWheelMysteries(ctx context.Context) ([]dto.SpinningWheelMysteryResData, error)
	UpdateSpinningWheelConfig(ctx context.Context, req dto.UpdateSpinningWheelConfigReq) (dto.UpdateSpinningWheelConfigRes, error)
	DeleteSpinningWheelConfig(ctx context.Context, id uuid.UUID) error
	CreateLevel(ctx context.Context, level dto.Level) (dto.Level, error)
	GetLevels(ctx context.Context, req dto.GetRequest) (dto.GetLevelResp, error)
	GetLevelByID(ctx context.Context, id uuid.UUID) (dto.Level, error)
	CreateLevelRequirements(ctx context.Context, req dto.LevelRequirements) (dto.LevelRequirements, error)
	GetLevelRequirementsByLevelID(ctx context.Context, levelID uuid.UUID, req dto.GetRequest) (dto.LevelRequirements, error)
	UpdateLevelRequirement(ctx context.Context, req dto.LevelRequirement) (dto.LevelRequirement, error)
	GetAllLevels(ctx context.Context, tp string) ([]dto.Level, error)
	CalculateTotalBetAmount(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	GetAllLevelRequirementsByLevelID(ctx context.Context, levelID uuid.UUID) (dto.LevelRequirements, error)
	AddFakeBalanceLog(ctx context.Context, userID uuid.UUID, changeAmount decimal.Decimal, currency string) error
	CalculateSquadBets(ctx context.Context, squadID uuid.UUID) (decimal.Decimal, error)
	GetUserSquads(ctx context.Context, userID uuid.UUID) ([]dto.Squad, error)
	GetAllSquadMembersBySquadId(ctx context.Context, squadID uuid.UUID) ([]dto.SquadMember, error)
	CreateLootbox(ctx context.Context, req dto.CreateLootBoxReq) (dto.CreateLootBoxRes, error)
	DeleteLootbox(ctx context.Context, id uuid.UUID) (dto.DeleteLootBoxRes, error)
	UpdateLootbox(ctx context.Context, req dto.UpdateLootBoxReq) (dto.UpdateLootBoxRes, error)
	GetLootboxByID(ctx context.Context, id uuid.UUID) (dto.LootBox, error)
	GetAllLootboxes(ctx context.Context) ([]dto.LootBox, error)
	PlaceLootBoxBet(ctx context.Context, req dto.PlaceLootBoxBetReq) (dto.PlaceLootBoxBetRes, error)
	UpdateLootBoxBet(ctx context.Context, req dto.PlaceLootBoxBetReq) (dto.PlaceLootBoxBetRes, error)
	GetLootBoxBetByID(ctx context.Context, id uuid.UUID) (dto.PlaceLootBoxBetRes, error)
}

type Departements interface {
	CreateDepartement(ctx context.Context, dep dto.CreateDepartementReq) (dto.CreateDepartementRes, error)
	GetDepartmentsByID(ctx context.Context, id uuid.UUID) (dto.CreateDepartementRes, bool, error)
	GetDepartmentsByName(ctx context.Context, name string) (dto.CreateDepartementRes, bool, error)
	UpdateDepartment(ctx context.Context, dep dto.UpdateDepartment) (dto.UpdateDepartment, error)
	AssignUserToDepartment(ctx context.Context, assignDep dto.AssignDepartmentToUserReq) (dto.AssignDepartmentToUserResp, error)
	GetDepartments(ctx context.Context, getDepReq dto.GetDepartementsReq) (dto.GetDepartementsRes, bool, error)
}

type Performance interface {
	GetFinancialMatrix(ctx context.Context) ([]dto.FinancialMatrix, error)
	GetGameMatrics(ctx context.Context) (dto.GameMatricsRes, error)
}

type Authz interface {
	CreateRole(ctx context.Context, req dto.CreateRoleReq) (dto.Role, error)
	GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (dto.Permissions, bool, error)
	AssignPermissionToRole(ctx context.Context, permissionID, roleID uuid.UUID) (dto.AssignPermissionToRoleRes, error)
	GetRoleByName(ctx context.Context, name string) (dto.Role, bool, error)
	RemoveRoleByID(ctx context.Context, roleID uuid.UUID) error
	RemoveRolePermissions(ctx context.Context, id uuid.UUID) error
	RemoveRolePermissionsByRoleID(ctx context.Context, id uuid.UUID) error
	GetPermissions(ctx context.Context, req dto.GetPermissionReq) ([]dto.Permissions, error)
	GetRoles(ctx context.Context, getRoleReq dto.GetRoleReq) ([]dto.Role, bool, error)
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (dto.Role, bool, error)
	RemoveRolePermissionFromCasbinRule(ctx context.Context, roleID uuid.UUID) error
	GetRolePermissionsByRoleID(ctx context.Context, roleID uuid.UUID) (dto.RolePermissions, bool, error)
	AddRoleToUser(ctx context.Context, roleID, userID uuid.UUID) (dto.AssignRoleToUserReq, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, bool, error)
	GetUserRoleUsingUserIDandRole(ctx context.Context, userID, roleID uuid.UUID) (dto.UserRole, bool, error)
	RemoveRoleFromUserRoles(ctx context.Context, roleID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]dto.User, error)
}

type Config interface {
	UpdateConfigByName(ctx context.Context, cn dto.Config) (dto.Config, error)
	UpdateScratchGameConfig(ctx context.Context, cn dto.Config) (dto.Config, error)
	UpdateConfigByID(ctx context.Context, cn dto.Config) (dto.Config, error)
	CreateConfig(ctx context.Context, cn dto.Config) (dto.Config, error)
	GetConfigByName(ctx context.Context, name string) (dto.Config, bool, error)
	GetScratchCardConfigs(ctx context.Context) (dto.GetScratchCardConfigs, error)
}

type Airtime interface {
	GetAllAirtimeUtilities(ctx context.Context) ([]dto.AirtimeUtility, bool, error)
	CreateUtility(ctx context.Context, req dto.AirtimeUtility) (dto.AirtimeUtility, error)
	GetAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, bool, error)
	UpdateAirtimeStatus(ctx context.Context, ID uuid.UUID, status string) (dto.UpdateAirtimeStatusResp, error)
	GetAirtimeUtilityByLocalID(ctx context.Context, localID uuid.UUID) (dto.AirtimeUtility, bool, error)
	UpdateAirtimeUtilityPrice(ctx context.Context, req dto.UpdateAirtimeUtilityPriceReq) (dto.AirtimeUtility, error)
	GetActiveAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error)
	SaveAirtimeTransactions(ctx context.Context, req dto.AirtimeTransactions) (dto.AirtimeTransactions, error)
	GetUserAirtimeTransactions(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetAirtimeTransactionsResp, error)
	GetAllAirtimeUtilitiesTransactions(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeTransactionsResp, error)
	UpdateAirtimeAmount(ctx context.Context, req dto.UpdateAirtimeAmountReq) (dto.AirtimeUtility, error)
	GetAirtimeUtilitiesStats(ctx context.Context) (dto.AirtimeUtilitiesStats, error)
}

type Company interface {
	CreateCompany(ctx context.Context, req dto.CreateCompanyReq) (dto.CreateCompanyRes, error)
	GetCompanyByID(ctx context.Context, id uuid.UUID) (dto.Company, bool, error)
	GetCompanies(ctx context.Context, req dto.GetCompaniesReq) (dto.GetCompaniesRes, error)
	UpdateCompany(ctx context.Context, req dto.UpdateCompanyReq) (dto.UpdateCompanyRes, error)
	DeleteCompany(ctx context.Context, id uuid.UUID) error
	AddIP(ctx context.Context, companyID uuid.UUID, ip string) (dto.UpdateCompanyRes, error)
}

type Report interface {
	DailyReport(ctx context.Context, req dto.DailyReportReq) (dto.DailyReportRes, error)
}

type Squads interface {
	CreateSquads(ctx context.Context, req dto.CreateSquadsReq) (dto.CreateSquadsRes, error)
	GetSquadByHandle(ctx context.Context, handle string) (dto.Squad, bool, error)
	GetSquadByID(ctx context.Context, id uuid.UUID) (dto.Squad, bool, error)
	GetUserSquads(ctx context.Context, id uuid.UUID) ([]dto.GetSquadsResp, error)
	GetSquadByOwnerID(ctx context.Context, id uuid.UUID) ([]dto.Squad, bool, error)
	GetSquadsByType(ctx context.Context, squadType string) ([]dto.Squad, error)
	UpdateSquadHundle(ctx context.Context, sd dto.Squad) (dto.Squad, error)
	DeleteSquad(ctx context.Context, id uuid.UUID) error
	CreateSquadMember(ctx context.Context, req dto.CreateSquadReq) (dto.SquadMember, error)
	GetSquadMembersBySquadID(ctx context.Context, req dto.GetSquadMemebersReq) (dto.GetSquadMemebersRes, error)
	DeleteSquadMemberByID(ctx context.Context, id uuid.UUID) error
	DeleteSquadMembersBySquadID(ctx context.Context, id uuid.UUID) error
	AddSquadEarn(ctx context.Context, req dto.SquadEarns) (dto.SquadEarns, error)
	GetSquadEarns(ctx context.Context, req dto.GetSquadEarnsReq) (dto.GetSquadEarnsResp, error)
	GetSquadEarnsByUserID(ctx context.Context, req dto.GetSquadEarnsReq, UserID uuid.UUID) (dto.GetSquadEarnsResp, error)
	GetSquadTotalEarns(ctx context.Context, id uuid.UUID) (decimal.Decimal, error)
	GetUserEarnsForSquad(ctx context.Context, squadID, userID uuid.UUID) (decimal.Decimal, error)
	GetTornamentStyleRanking(ctx context.Context, req dto.GetTornamentStyleRankingReq) (dto.GetTornamentStyleRankingResp, error)
	CreateTournaments(ctx context.Context, req dto.CreateTournamentReq) (dto.CreateTournamentResp, error)
	GetTornamentStyles(ctx context.Context) ([]dto.Tournament, error)
	CreateTournamentClaim(ctx context.Context, tournamentID, squadID uuid.UUID) (dto.TournamentClaim, error)
	GetTournamentClaimBySquadID(ctx context.Context, tournamentID, squadID uuid.UUID) (dto.TournamentClaim, error)
	GetSquadMemberByID(ctx context.Context, id uuid.UUID) (*dto.GetSquadMemberByIDresp, error)
	GetSquadMembersEarnings(ctx context.Context, req dto.GetSquadMembersEarningsReq, ownerID uuid.UUID) (dto.GetSquadMembersEarningsResp, error)
	LeaveSquad(ctx context.Context, userID, squadID uuid.UUID) error
	GetSquadMembersByUserID(ctx context.Context, userID uuid.UUID) ([]dto.GetSquadMemberByIDresp, error)
	AddToWaitingSquadMembers(ctx context.Context, req dto.CreateSquadReq) (dto.CreateSquadReq, error)
	GetWaitingSquadMembers(ctx context.Context, squadID uuid.UUID) ([]dto.WaitingSquadMember, error)
	DeleteWaitingSquadMember(ctx context.Context, id uuid.UUID) error
	GetWaitingSquadsOwner(ctx context.Context, squadID uuid.UUID) (dto.WaitingsquadOwner, error)
	GetSquadsByOwner(ctx context.Context, ownerID uuid.UUID) ([]dto.GetSquadsResp, error)
	AddWaitingSquadMember(ctx context.Context, userID, squadID uuid.UUID) (dto.SquadMember, error)
	ApproveWaitingSquadMember(ctx context.Context, ID uuid.UUID) error
}

type Notification interface {
	StoreNotification(ctx context.Context, req dto.NotificationPayload, delivered bool) (dto.NotificationResponse, error)
	GetUserNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error)
	MarkNotificationRead(ctx context.Context, req dto.MarkNotificationReadRequest) (dto.MarkNotificationReadResponse, error)
	MarkAllNotificationsRead(ctx context.Context, req dto.MarkAllNotificationsReadRequest) (dto.MarkAllNotificationsReadResponse, error)
	DeleteNotification(ctx context.Context, req dto.DeleteNotificationRequest) (dto.DeleteNotificationResponse, error)
	GetUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int32, error)
}

type Adds interface {
	SaveAddsService(ctx context.Context, req dto.CreateAddsServiceReq) (dto.CreateAddsServiceRes, error)
	GetAddsServices(ctx context.Context, req dto.GetAddServicesRequest) (dto.GetAddsServicesRes, error)
	GetAddsServiceByServiceID(ctx context.Context, serviceID string) (dto.AddsServiceResData, bool, error)
	GetAddsServiceByID(ctx context.Context, id uuid.UUID) (dto.AddsServiceResData, bool, error)
}

type Banner interface {
	GetAllBanners(ctx context.Context, req dto.GetBannersReq) (dto.GetBannersRes, error)
	GetBannerByPage(ctx context.Context, req dto.GetBannerReq) (dto.Banner, bool, error)
	UpdateBanner(ctx context.Context, req dto.UpdateBannerReq) (dto.Banner, error)
	GetBannerByID(ctx context.Context, id uuid.UUID) (dto.Banner, error)
	CreateBanner(ctx context.Context, req dto.CreateBannerReq) (dto.Banner, error)
	DeleteBanner(ctx context.Context, id uuid.UUID) error
}

type Lottery interface {
	CreateLotteryService(ctx context.Context, req dto.CreateLotteryServiceReq) (dto.CreateLotteryServiceRes, error)
	GetLotteryServiceByID(ctx context.Context, serviceID uuid.UUID) (dto.CreateLotteryServiceReq, error)
	CreateLotteryWinnersLogs(ctx context.Context, req dto.LotteryLog) (dto.LotteryLog, error)
	CreateLotteryLog(ctx context.Context, req dto.LotteryKafkaLog) (dto.LotteryKafkaLog, error)
	GetAvailableLotteryService(ctx context.Context) (dto.CreateLotteryServiceReq, error)
	GetLotteryLogsByUniqIdentifier(ctx context.Context, uniqIdentifier uuid.UUID) ([]dto.LotteryKafkaLog, error)
}

type Sports interface {
	PlaceBet(ctx context.Context, req dto.PlaceBetRequest) (*dto.PlaceBetResponse, error)
	AwardWinnings(ctx context.Context, req dto.SportsServiceAwardWinningsReq) (*dto.SportsServiceAwardWinningsRes, error)
}

type RiskSettings interface {
	GetRiskSettings(ctx context.Context) (dto.RiskSettings, error)
	SetRiskSettings(ctx context.Context, req dto.RiskSettings) (dto.RiskSettings, error)
}

type Agent interface {
	CreateAgentReferralLink(ctx context.Context, req dto.CreateAgentReferralLinkReq) (dto.AgentReferral, error)
	UpdateAgentReferralWithConversion(ctx context.Context, req dto.UpdateAgentReferralWithConversionReq) (dto.AgentReferral, error)
	GetAgentReferralByRequestID(ctx context.Context, requestID string) (dto.AgentReferral, bool, error)
	GetAgentReferralsByRequestID(ctx context.Context, requestID string, limit, offset int) ([]dto.AgentReferral, error)
	CountAgentReferralsByRequestID(ctx context.Context, requestID string) (int, error)
	GetReferralsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.AgentReferral, error)
	CountReferralsByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	GetPendingCallbacks(ctx context.Context, limit, offset int) ([]dto.PendingCallback, error)
	MarkCallbackSent(ctx context.Context, referralID uuid.UUID) error
	IncrementCallbackAttempts(ctx context.Context, referralID uuid.UUID) error
	GetReferralStatsByRequestID(ctx context.Context, requestID string) (dto.ReferralStats, error)
	GetReferralStatsByConversionType(ctx context.Context, requestID string) ([]dto.ConversionTypeStats, error)
	CreateAgentProvider(ctx context.Context, req dto.CreateAgentProviderReq) (dto.AgentProviderRes, error)
	GetAgentProviderByID(ctx context.Context, id uuid.UUID) (db.AgentProvider, error)
	GetAgentProviderByClientID(ctx context.Context, clientID string) (db.AgentProvider, error)
}


// GameInfoProvider provides game information for analytics
type GameInfoProvider interface {
	GetGameInfo(ctx context.Context, gameID string) (*dto.GameInfo, error)
}
