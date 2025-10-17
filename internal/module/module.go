package module

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
)

type User interface {
	RegisterUser(ctx context.Context, userRequest dto.User) (dto.UserRegisterResponse, string, error)
	Login(ctx context.Context, loginRequest dto.UserLoginReq, loginLogs dto.LoginAttempt, adminLogin bool) (dto.UserLoginRes, string, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (dto.UserProfile, error)
	UploadProfilePicture(ctx context.Context, img multipart.File, header *multipart.FileHeader, userID uuid.UUID) (string, error)
	ChangePassword(ctx context.Context, changePasswordReq dto.ChangePasswordReq) (dto.ChangePasswordRes, error)
	ForgetPassword(ctx context.Context, usernameOrPhoneOrEmail, userAgent, ipAddress string) (*dto.ForgetPasswordRes, error)
	VerifyResetPassword(ctx context.Context, resetPasswordReq dto.VerifyResetPasswordReq) (*dto.VerifyResetPasswordRes, error)
	ResetPassword(ctx context.Context, resetPasswordReq dto.ResetPasswordReq) (dto.ResetPasswordRes, error)
	UpdateProfile(ctx context.Context, profileupateReq dto.UpdateProfileReq) (dto.UpdateProfileRes, error)
	SaveUpdateProfile(ctx context.Context, profile dto.UpdateProfileReq, usr dto.User) (dto.User, error)
	ConfirmUpdateProfile(ctx context.Context, confirmOTP dto.ConfirmUpdateProfile) (dto.User, error)
	GoogleLoginReq(ctx context.Context) string
	GoogleLoginRes(ctx context.Context, code string, loginattempt dto.LoginAttempt) (dto.UserRegisterResponse, error)
	FacebookLoginReq(ctx context.Context) string
	HandleFacebookOauthRes(ctx context.Context, code string, loginattempt dto.LoginAttempt) (dto.UserRegisterResponse, error)
	BlockUser(ctx context.Context, blockAcc dto.AccountBlockReq) (dto.AccountBlockRes, error)
	GetBlockedAccount(ctx context.Context, blockAccountReq dto.GetBlockedAccountLogReq) ([]dto.GetBlockedAccountLogRep, error)
	AddIpFilter(ctx context.Context, ipFilter dto.IpFilterReq) (dto.IPFilterRes, error)
	GetIPFilters(ctx context.Context, getIPFilterReq dto.GetIPFilterReq) (dto.GetIPFilterRes, error)
	EnforceIPFilerRule(ctx context.Context, ip string) (bool, error)
	AdminUpdateProfile(ctx context.Context, userReq dto.EditProfileAdminReq) (dto.EditProfileAdminRes, error)
	AdminResetPassword(ctx context.Context, req dto.AdminResetPasswordReq) (dto.AdminResetPasswordRes, error)
	GetPlayers(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error)
	RemoveIPFilter(ctx context.Context, req dto.RemoveIPBlockReq) (dto.RemoveIPBlockRes, error)
	GetMyReferralCode(ctx context.Context, userID uuid.UUID) (string, error)
	GetUserReferalUsersByUserID(ctx context.Context, userID uuid.UUID) (dto.MyRefferedUsers, error)
	GetReferalMultiplier(ctx context.Context) (dto.ReferalUpdateResp, error)
	UpdateReferalMultiplier(ctx context.Context, mul dto.UpdateReferralPointReq) (dto.ReferalUpdateResp, error)
	UpdateUsersPointsForReferrances(ctx context.Context, adminID uuid.UUID, req []dto.MassReferralReq) (dto.MassReferralRes, error)
	GetAdminAssignedPoints(ctx context.Context, req dto.GetAdminAssignedPointsReq) (dto.GetAdminAssignedPointsRes, error)
	GetUserPoints(ctx context.Context, useID uuid.UUID) (dto.GetPointsResp, error)
	AdminCreatePlayer(ctx context.Context, userRequest dto.User) (dto.UserRegisterResponse, string, error)
	AdminLogin(ctx context.Context, loginRequest dto.UserLoginReq, loginLogs dto.LoginAttempt) (dto.UserLoginRes, string, error)
	GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
	UpdateSignupBonus(ctx context.Context, req dto.SignUpBonusReq) (dto.SignUpBonusRes, error)
	GetSignupBonusConfig(ctx context.Context) (dto.SignUpBonusRes, error)
	UpdateReferralBonus(ctx context.Context, req dto.ReferralBonusReq) (dto.ReferralBonusRes, error)
	GetReferralBonusConfig(ctx context.Context) (dto.ReferralBonusRes, error)
	RefreshTokenFlow(ctx context.Context, refreshToken string) (string, string, time.Time, error)
	// Session monitoring methods
	AddSessionSocketConnection(userID uuid.UUID, conn *websocket.Conn)
	RemoveSessionSocketConnection(userID uuid.UUID, conn *websocket.Conn)
	SendSessionEvent(userID uuid.UUID, event dto.SessionEventMessage) bool
	NotifySessionExpired(userID uuid.UUID) bool
	NotifySessionRefreshed(userID uuid.UUID) bool
	MonitorUserSessions(ctx context.Context)
	HandleSessionExpiry(ctx context.Context, userID uuid.UUID) error
	Stop()
	GetUserByID(ctx context.Context, userID uuid.UUID) (dto.User, bool, error)
	GetPlayerSuspensionHistory(ctx context.Context, userID uuid.UUID) ([]dto.SuspensionHistory, error)
	GetPlayerBalanceLogs(ctx context.Context, userID uuid.UUID) ([]dto.BalanceLog, error)
	GetPlayerGameActivity(ctx context.Context, userID uuid.UUID) ([]dto.GameActivity, error)
	GetPlayerBalances(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error)
	GetPlayerStatistics(ctx context.Context, userID uuid.UUID) (dto.PlayerStatistics, error)
	VerifyUser(ctx context.Context, req dto.VerifyPhoneNumberReq) (dto.UserRegisterResponse, string, error)
	GetOtp(ctx context.Context, phone string) string
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	CheckUserExistsByPhoneNumber(ctx context.Context, phone string) (bool, error)
	UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, verified bool) (dto.User, error)
	ReSendVerificationOTP(ctx context.Context, phoneNumber string) (*dto.ForgetPasswordRes, error)
	CheckUserExistsByUsername(ctx context.Context, username string) (bool, error)
}

type Balance interface {
	Update(ctx context.Context, updateBalanceReq dto.UpdateBalanceReq) (dto.UpdateBalanceRes, error)
	GetBalanceByUserID(ctx context.Context, userID uuid.UUID) ([]dto.Balance, error)
	Exchange(ctx context.Context, exchngeReq dto.ExchangeBalanceReq) (dto.ExchangeBalanceRes, error)
	AddManualFunds(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error)
	RemoveFundManualy(ctx context.Context, fund dto.ManualFundReq) (dto.ManualFundRes, error)
	GetManualFundLogs(ctx context.Context, filter dto.GetManualFundReq) (dto.GetManualFundRes, error)
	CreditWallet(ctx context.Context, req dto.CreditWalletReq) (dto.CreditWalletRes, error)
}

type OperationalGroup interface {
	CreateOperationalGroup(ctx context.Context, opReq dto.OperationalGroup) (dto.OperationalGroup, error)
	GetOperationalGroups(ctx context.Context) ([]dto.OperationalGroup, error)
}

type OperationalGroupType interface {
	CreateOperationalGroupType(ctx context.Context, optReq dto.OperationalGroupType) (dto.OperationalGroupType, error)
	GetOperationalGroupTypeByGroupID(ctx context.Context, groupID uuid.UUID) ([]dto.OperationalGroupType, error)
	GetOperationalGroupTypes(ctx context.Context) ([]dto.OperationalTypesRes, error)
}

type OperationsDefinitions interface {
	GetOperationalDefinition(ctx context.Context) (dto.OperationsDefinition, error)
}

type BalanceLogs interface {
	GetBalanceLogs(ctx context.Context, balanceLogsReq dto.GetBalanceLogReq) (dto.GetBalanceLogRes, error)
	GetBalanceLogByID(ctx context.Context, balanceLogID uuid.UUID) (dto.BalanceLogsRes, error)
	GetBalanceLogsForAdmin(ctx context.Context, req dto.AdminGetBalanceLogsReq) (dto.AdminGetBalanceLogsRes, error)
}

type Exchange interface {
	GetExchange(ctx context.Context, exchangeReq dto.ExchangeReq) (dto.ExchangeRes, error)
	GetAvailableCurrencies(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error)
	GetCurrency(ctx context.Context, req dto.GetRequest) (dto.GetCurrencyReq, error)
	CreateCurrency(ctx context.Context, req dto.Currency) (dto.Currency, error)
}

type Bet interface {
	AddConnection(ctx context.Context, connReq dto.BroadCastPayload)
	GetOpenRound(ctx context.Context) (dto.OpenRoundRes, error)
	PlaceBet(ctx context.Context, placeBetReq dto.PlaceBetReq) (dto.PlaceBetRes, error)
	CashOut(ctx context.Context, cashOutReq dto.CashOutReq) (dto.CashOutRes, error)
	GetBetHistory(ctx context.Context, betReq dto.GetBetHistoryReq) (dto.BetHistoryResp, error)
	CancelBet(ctx context.Context, cancelReq dto.CancelBetReq) (dto.CancelBetResp, error)
	GetLeaders(ctx context.Context) (dto.LeadersResp, error)
	AddToBroadcastConnection(ctx context.Context, conn *websocket.Conn)
	GetFailedRounds(ctx context.Context, req dto.GetFailedRoundsReq) (dto.GetFailedRoundsRes, error)
	ManualRefundFailedRounds(ctx context.Context, req dto.ManualRefundFailedRoundsReq) (dto.ManualRefundFailedRoundsRes, error)
	GetPlinkoGameConfig(ctx context.Context) (dto.PlinkoGameConfig, error)
	PlacePlinkoGame(ctx context.Context, req dto.PlacePlinkoGameReq) (dto.PlacePlinkoGameRes, error)
	GetMyPlinkoBetHistory(ctx context.Context, req dto.PlinkoBetHistoryReq) (dto.PlinkoBetHistoryRes, error)
	GetPlinkoGameStats(ctx context.Context, userID uuid.UUID) (dto.PlinkoGameStatRes, error)
	CreateLeague(ctx context.Context, req dto.League) (dto.League, error)
	GetLeagues(ctx context.Context, req dto.GetRequest) (dto.GetLeagueRes, error)
	CreateClub(ctx context.Context, req dto.Club) (dto.Club, error)
	GetClubs(ctx context.Context, req dto.GetRequest) (dto.GetClubRes, error)
	CreateFootballCardMultiplier(ctx context.Context, req dto.FootballCardMultiplier) (dto.Config, error)
	GetFootballCardMultiplier(ctx context.Context) (dto.Config, error)
	UpdateFootballCardMultiplierValue(ctx context.Context, req dto.FootballCardMultiplier) (dto.Config, error)
	CreateFootballMatchRound(ctx context.Context) (dto.FootballMatchRound, error)
	GetFootballMatchRounds(ctx context.Context, req dto.GetRequest) (dto.GetFootballMatchRoundRes, error)
	CreateFootballMatch(ctx context.Context, req []dto.FootballMatchReq) ([]dto.FootballMatch, error)
	GetFootballRoundMatchs(ctx context.Context, req dto.GetFootballRoundMatchesReq) (dto.GetFootballRoundMatchesRes, error)
	GetCurrentFootballRound(ctx context.Context) (dto.GetFootballRoundMatchesRes, error)
	CloseFootballMatch(ctx context.Context, req dto.CloseFootballMatchReq) (dto.FootballMatch, error)
	UpdateFootballRoundPrice(ctx context.Context, req dto.UpdateFootballBetPriceReq) (dto.UpdateFootballBetPriceRes, error)
	GetFootballMatchPrice(ctx context.Context) (dto.UpdateFootballBetPriceRes, error)
	PleceBetOnFootballRound(ctx context.Context, req dto.UserFootballMatchBetReq) (dto.UserFootballMatchBetRes, error)
	GetUserFootballBets(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetUserFootballBetRes, error)
	CreateStreetKingsGame(ctx context.Context, req dto.CreateCrashKingsReq, userID uuid.UUID) (dto.CreateStreetKingsResp, error)
	AddToSingleGameConnections(ctx context.Context, connReq dto.BroadCastPayload)
	CashOutStreetKings(ctx context.Context, req dto.CashOutReq) (dto.StreetKingsCrashResp, error)
	GetStreetkingHistory(ctx context.Context, req dto.GetStreetkingHistoryReq, userID uuid.UUID) (dto.GetStreetkingHistoryRes, error)
	SetCrytoKingsConfig(ctx context.Context, req dto.UpdateCryptoKingsConfigReq) (dto.UpdateCryptokingsConfigRes, error)
	PlaceCryptoKingsBet(ctx context.Context, req dto.PlaceCryptoKingsBetReq, userID uuid.UUID) (dto.PlaceCryptoKingsBetRes, error)
	GetCryptoKingsBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetCryptoKingsUserBetHistoryRes, error)
	GetCryptoKingsCurrentCryptoPrice(ctx context.Context, userID uuid.UUID) (dto.GetCryptoCurrencyPriceResp, error)
	PlaceQuickHustleBet(ctx context.Context, req dto.CreateQuickHustleBetReq) (dto.CreateQuickHustelBetRes, error)
	UserSelectCard(ctx context.Context, req dto.SelectQuickHustlePossibilityReq) (dto.CloseQuickHustleResp, error)
	GetQuickHustleBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetQuickHustleResp, error)
	CreateRollDaDice(ctx context.Context, req dto.CreateRollDaDiceReq) (dto.CreateRollDaDiceResp, error)
	GetRollDaDiceHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetRollDaDiceResp, error)
	GetScratchGamePrice(ctx context.Context) (dto.GetScratchCardRes, error)
	PlaceScratchCardBet(ctx context.Context, userID uuid.UUID) (dto.ScratchCard, error)
	GetUserScratchCardBetHistories(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetScratchBetHistoriesResp, error)
	GetSpinningWheelPrice(ctx context.Context) (dto.GetSpinningWheelPrice, error)
	PlaceSpinningWheelBet(ctx context.Context, userID uuid.UUID) (dto.PlaceSpinningWheelResp, error)
	GetSpinningWheelUserBetHistory(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetSpinningWheelHistoryResp, error)
	UpdateGame(ctx context.Context, game dto.Game) (dto.Game, error)
	GetGames(ctx context.Context, req dto.GetRequest) (dto.GetGamesResp, error)
	GetGameSummary(ctx context.Context) (dto.GetGameSummaryResp, error)
	GetTransactionSummary(ctx context.Context) (dto.GetTransactionSummaryResp, error)
	DisableAllGames(ctx context.Context) (dto.BlockGamesResp, error)
	ListInactiveGames(ctx context.Context) ([]dto.Game, error)
	DeleteGame(ctx context.Context, req dto.Game) (dto.DeleteResponse, error)
	AddGame(ctx context.Context, game dto.Game) (dto.Game, error)
	UpdateGameStatus(ctx context.Context, game dto.Game) (dto.Game, error)
	CreateSpinningWheelMystery(ctx context.Context, req dto.CreateSpinningWheelMysteryReq) (dto.CreateSpinningWheelMysteryRes, error)
	GetSpinningWheelMystery(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelMysteryRes, error)
	UpdateSpinningWheelMystery(ctx context.Context, req dto.UpdateSpinningWheelMysteryReq) (dto.UpdateSpinningWheelMysteryRes, error)
	DeleteSpinningWheelMystery(ctx context.Context, req dto.DeleteReq) error
	CreateSpinningWheelConfig(ctx context.Context, req dto.CreateSpinningWheelConfigReq) (dto.CreateSpinningWheelConfigRes, error)
	GetSpinningWheelConfig(ctx context.Context, req dto.GetRequest) (dto.GetSpinningWheelConfigRes, error)
	UpdateSpinningWheelConfig(ctx context.Context, req dto.UpdateSpinningWheelConfigReq) (dto.UpdateSpinningWheelConfigRes, error)
	DeleteSpinningWheelConfig(ctx context.Context, req dto.DeleteReq) error
	UploadBetIcons(ctx context.Context, img multipart.File, header *multipart.FileHeader) (dto.UploadIconsResp, error)
	GetScratchCardsConfig(ctx context.Context) (dto.GetScratchCardConfigs, error)
	UpdateScratchGameConfig(ctx context.Context, req dto.UpdateScratchGameConfigRequest) (dto.UpdateScratchGameConfigResponse, error)
	CreateLevel(ctx context.Context, level dto.Level) (dto.Level, error)
	GetLevels(ctx context.Context, req dto.GetRequest) (dto.GetLevelResp, error)
	CreateLevelReqirements(ctx context.Context, req dto.LevelRequirements) (dto.LevelRequirements, error)
	UpdateLevelRequirement(ctx context.Context, req dto.UpdateLevelRequirementReq) (dto.LevelRequirement, error)
	GetUserLevel(ctx context.Context, userID uuid.UUID) (dto.GetUserLevelResp2, error)
	AddPlayerLevelSocketConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn)
	TriggerLevelResponse(ctx context.Context, userID uuid.UUID)
	AddFakeBalanceLog(ctx context.Context, userID uuid.UUID, changeAmount decimal.Decimal, currency string) error
	TriggerPlayerProgressBar(ctx context.Context, userID uuid.UUID)
	AddPlayerProgressBarConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn)
	GetSquadLevel(ctx context.Context, userID uuid.UUID) (dto.GetUserLevelResp2, error)
	InitiateTriggerSquadsProgressBar(ctx context.Context, userID uuid.UUID)
	AddSquadsProgressBarConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn)
	CreateLootBox(ctx context.Context, req dto.CreateLootBoxReq) (dto.CreateLootBoxRes, error)
	UpdateLootBox(ctx context.Context, req dto.UpdateLootBoxReq) (dto.UpdateLootBoxRes, error)
	DeleteLootBox(ctx context.Context, id uuid.UUID) (dto.DeleteLootBoxRes, error)
	GetLootBox(ctx context.Context) ([]dto.LootBox, error)
	PlaceLootBoxBet(ctx context.Context, userID uuid.UUID) ([]dto.PlaceLootBoxResp, error)
	SelectLootBox(ctx context.Context, lootBox dto.PlaceLootBoxResp, userID uuid.UUID) (dto.LootBoxBetResp, error)
}

type Departements interface {
	CreateDepartement(ctx context.Context, deps dto.CreateDepartementReq) (dto.CreateDepartementRes, error)
	GetDepartments(ctx context.Context, depReq dto.GetDepartementsReq) (dto.GetDepartementsRes, error)
	UpdateDepartment(ctx context.Context, dep dto.UpdateDepartment) (dto.UpdateDepartment, error)
	AssignUserToDepartment(ctx context.Context, assignReq dto.AssignDepartmentToUserReq) (dto.AssignDepartmentToUserResp, error)
}

type Performance interface {
	GetFinancialMatrix(ctx context.Context) ([]dto.FinancialMatrix, error)
	GetGameMatrics(ctx context.Context) (dto.GameMatricsRes, error)
}

type Authz interface {
	GetPermissions(ctx context.Context, req dto.GetPermissionReq) ([]dto.Permissions, error)
	CreateRole(ctx context.Context, req dto.CreateRoleReq) (dto.Role, error)
	GetRoles(ctx context.Context, req dto.GetRoleReq) ([]dto.Role, error)
	UpdatePermissionsOfRole(ctx context.Context, req dto.UpdatePermissionToRoleReq) (dto.UpdatePermissionToRoleRes, error)
	RemoveRole(ctx context.Context, roleID uuid.UUID) error
	AssignRoleToUser(ctx context.Context, req dto.AssignRoleToUserReq) (dto.AssignRoleToUserRes, error)
	RevokeUserRole(ctx context.Context, req dto.UserRole) error
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]dto.User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) (dto.UserRolesRes, error)
	GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error)
}

type AirtimeProvider interface {
	Login()
	RefereshUtilies(ctx context.Context) ([]dto.AirtimeUtility, error)
	GetAvailableAirtimeUtilities(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error)
	UpdateAirtimeStatus(ctx context.Context, req dto.UpdateAirtimeStatusReq) (dto.UpdateAirtimeStatusResp, error)
	UpdateUtilityPrice(ctx context.Context, req dto.UpdateAirtimeUtilityPriceReq) (dto.UpdateAirtimeUtilityPriceRes, error)
	ClaimPoints(ctx context.Context, req dto.ClaimPointsReq) (dto.ClaimPointsResp, error)
	GetActiveAvailableAirtime(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeUtilitiesResp, error)
	GetUserAirtimeTransactions(ctx context.Context, req dto.GetRequest, userID uuid.UUID) (dto.GetAirtimeTransactionsResp, error)
	GetAllAirtimeUtilitiesTransactions(ctx context.Context, req dto.GetRequest) (dto.GetAirtimeTransactionsResp, error)
	UpdateAirtimeAmount(ctx context.Context, req dto.UpdateAirtimeAmountReq) (dto.AirtimeUtility, error)
	GetAirtimeUtilitiesStats(ctx context.Context) (dto.AirtimeUtilitiesStats, error)
}

type SystemLogs interface {
	CreateSystemLogs(ctx context.Context, systemLogReq dto.SystemLogs) (dto.SystemLogs, error)
	GetSystemLogs(ctx context.Context, req dto.GetSystemLogsReq) (dto.SystemLogsRes, error)
	GetAvailableLogsModule(ctx context.Context) ([]string, error)
}

type Company interface {
	CreateCompany(ctx context.Context, req dto.CreateCompanyReq) (dto.CreateCompanyRes, error)
	GetCompanyByID(ctx context.Context, id uuid.UUID) (dto.Company, error)
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
	GetMySquads(ctx context.Context, userID uuid.UUID) ([]dto.GetSquadsResp, error)
	GetSquadsByOwnerID(ctx context.Context, userID uuid.UUID) ([]dto.Squad, error)
	GetSquadsByType(ctx context.Context, squadType string) ([]dto.Squad, error)
	UpdateSquadHandle(ctx context.Context, sq dto.Squad, userID uuid.UUID) (dto.Squad, error)
	DeleteSquad(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CreateSquadMember(ctx context.Context, req dto.CreateSquadMemeberReq) (dto.SquadMember, error)
	GetSquadMembersBySquadID(ctx context.Context, req dto.GetSquadMemebersReq) (dto.GetSquadMemebersRes, error)
	DeleteSquadMember(ctx context.Context, id uuid.UUID, MemberID uuid.UUID) error
	DeleteSquadMembersBySquadID(ctx context.Context, id, userID uuid.UUID) error
	GetSquadEarns(ctx context.Context, req dto.GetSquadEarnsReq) (dto.GetSquadEarnsResp, error)
	GetMySquadEarns(ctx context.Context, req dto.GetSquadEarnsReq, UserID uuid.UUID) (dto.GetSquadEarnsResp, error)
	GetSquadTotalEarns(ctx context.Context, squadID uuid.UUID) (decimal.Decimal, error)
	GetSquadByName(ctx context.Context, name string) (*dto.Squad, error)
	GetTornamentStyleRanking(ctx context.Context, req dto.GetTornamentStyleRankingReq) (dto.GetTornamentStyleRankingResp, error)
	CreateTournaments(ctx context.Context, req dto.CreateTournamentReq) (dto.CreateTournamentResp, error)
	GetTornamentStyles(ctx context.Context) ([]dto.Tournament, error)
	GetSquadMembersEarnings(ctx context.Context, req dto.GetSquadMembersEarningsReq, ownerID uuid.UUID) (dto.GetSquadMembersEarningsResp, error)
	LeaveSquad(ctx context.Context, userID, squadID uuid.UUID) error
	JoinSquad(ctx context.Context, userID, squadID uuid.UUID) (dto.SquadMemberResp, error)
	GetWaitingSquadMembers(ctx context.Context, userID uuid.UUID) ([]dto.WaitingSquadMember, error)
	RemoveWaitingSquadWaitingUser(ctx context.Context, req dto.SquadMemberReq) error
	ApproveWaitingSquadMember(ctx context.Context, req dto.SquadMemberReq) error
}

type Notification interface {
	StoreNotification(ctx context.Context, req dto.NotificationPayload, delivered bool) (dto.NotificationResponse, error)
	GetUserNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error)
	GetAllNotifications(ctx context.Context, req dto.GetNotificationsRequest) (dto.GetNotificationsResponse, error)
	MarkNotificationRead(ctx context.Context, req dto.MarkNotificationReadRequest) (dto.MarkNotificationReadResponse, error)
	MarkAllNotificationsRead(ctx context.Context, req dto.MarkAllNotificationsReadRequest) (dto.MarkAllNotificationsReadResponse, error)
	DeleteNotification(ctx context.Context, req dto.DeleteNotificationRequest) (dto.DeleteNotificationResponse, error)
	GetUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int32, error)
	AddNotificationSocketConnection(userID uuid.UUID, conn *websocket.Conn)
	RemoveNotificationSocketConnection(userID uuid.UUID, conn *websocket.Conn)
	SendNotificationToUser(userID uuid.UUID, payload dto.NotificationPayload) bool
}

type Campaign interface {
	CreateCampaign(ctx context.Context, req dto.CreateCampaignRequest, createdBy uuid.UUID) (dto.CampaignResponse, error)
	GetCampaigns(ctx context.Context, req dto.GetCampaignsRequest) (dto.GetCampaignsResponse, error)
	GetCampaignByID(ctx context.Context, campaignID uuid.UUID) (dto.CampaignResponse, error)
	UpdateCampaign(ctx context.Context, campaignID uuid.UUID, req dto.UpdateCampaignRequest) (dto.CampaignResponse, error)
	DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error
	GetCampaignRecipients(ctx context.Context, req dto.GetCampaignRecipientsRequest) (dto.GetCampaignRecipientsResponse, error)
	GetCampaignStats(ctx context.Context, campaignID uuid.UUID) (dto.CampaignStatsResponse, error)
	SendCampaign(ctx context.Context, campaignID uuid.UUID) error
	GetScheduledCampaigns(ctx context.Context) ([]dto.CampaignResponse, error)
	GetCampaignNotificationsDashboard(ctx context.Context, req dto.GetCampaignNotificationsDashboardRequest) (dto.CampaignNotificationsDashboardResponse, error)
}

type Adds interface {
	SaveAddsService(ctx context.Context, req dto.CreateAddsServiceReq) (*dto.CreateAddsServiceRes, error)
	GetAddsServices(ctx context.Context, req dto.GetAddServicesRequest) (*dto.GetAddsServicesRes, error)
	UpdateBalance(ctx context.Context, req dto.AddUpdateBalanceReq) (*dto.AddUpdateBalanceRes, error)
	SignIn(ctx context.Context, req dto.AddSignInReq) (*dto.AddSignInRes, error)
}

type Banner interface {
	GetAllBanners(ctx context.Context, req dto.GetBannersReq) (dto.GetBannersRes, error)
	GetBannerByPage(ctx context.Context, req dto.GetBannerReq) (dto.Banner, error)
	UpdateBanner(ctx context.Context, req dto.UpdateBannerReq) (dto.Banner, error)
	CreateBanner(ctx context.Context, req dto.CreateBannerReq) (dto.Banner, error)
	DeleteBanner(ctx context.Context, id uuid.UUID) error
	UploadBannerImage(ctx context.Context, img multipart.File, header *multipart.FileHeader) (dto.UploadBannerImageResp, error)
}

type Lottery interface {
	CreateLotteryService(ctx context.Context, req dto.CreateLotteryServiceReq) (dto.CreateLotteryServiceRes, error)
	CreateLotteryRequest(ctx context.Context, req dto.LotteryRequestCreate) (dto.LotteryRequestCreate, error)
	CheckUserBalanceAndDeductBalance(ctx context.Context, req dto.LotteryVerifyAndDeductBalanceReq) (dto.LotteryVerifyAndDeductBalanceRes, error)
}

type SportsService interface {
	SignIn(ctx context.Context, req dto.SportsServiceSignInReq) (*dto.SportsServiceSignInRes, error)
	PlaceBet(ctx context.Context, req dto.PlaceBetRequest) (*dto.PlaceBetResponse, error)
	AwardWinnings(ctx context.Context, req dto.SportsServiceAwardWinningsReq) (*dto.SportsServiceAwardWinningsRes, error)
}

type RiskSettings interface {
	GetRiskSettings(ctx context.Context) (dto.RiskSettings, error)
	SetRiskSettings(ctx context.Context, req dto.RiskSettings) (dto.RiskSettings, error)
}

type Agent interface {
	CreateAgentReferralLink(ctx context.Context, req dto.CreateAgentReferralLinkReq) (dto.CreateAgentReferralLinkRes, error)
	UpdateAgentReferralWithConversion(ctx context.Context, req dto.UpdateAgentReferralWithConversionReq) (dto.UpdateAgentReferralWithConversionRes, error)
	GetAgentReferralByRequestID(ctx context.Context, req dto.GetAgentReferralReq) (dto.AgentReferral, bool, error)
	GetAgentReferralsByRequestID(ctx context.Context, req dto.GetAgentReferralsReq) (dto.GetAgentReferralsRes, error)
	GetReferralsByUserID(ctx context.Context, req dto.GetReferralsByUserReq) (dto.GetReferralsByUserRes, error)
	GetReferralStatsByRequestID(ctx context.Context, req dto.GetReferralStatsReq) (dto.GetReferralStatsRes, error)
	ProcessPendingCallbacks(ctx context.Context) error
	StartCallbackProcessor()
	CreateAgentProvider(ctx context.Context, req dto.CreateAgentProviderReq) (dto.CreateAgentProviderRes, error)
	ValidateAgentProviderCredentials(ctx context.Context, providerID string, secret string) (dto.AgentProviderRes, error)
}

type AdminActivityLogs interface {
	LogAdminActivity(ctx context.Context, req dto.CreateAdminActivityLogReq) error
	GetAdminActivityLogs(ctx context.Context, req dto.GetAdminActivityLogsReq) (dto.AdminActivityLogsRes, error)
	GetAdminActivityLogByID(ctx context.Context, id uuid.UUID) (dto.AdminActivityLog, error)
	GetAdminActivityStats(ctx context.Context, from, to *time.Time) (dto.AdminActivityStats, error)
	GetAdminActivityCategories(ctx context.Context) ([]dto.AdminActivityCategory, error)
	GetAdminActivityActions(ctx context.Context) ([]dto.AdminActivityAction, error)
	GetAdminActivityActionsByCategory(ctx context.Context, category string) ([]dto.AdminActivityAction, error)
	DeleteAdminActivityLog(ctx context.Context, id uuid.UUID) error
	DeleteAdminActivityLogsByAdmin(ctx context.Context, adminUserID uuid.UUID) error
	DeleteOldAdminActivityLogs(ctx context.Context, before time.Time) error
}
