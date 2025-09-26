package handler

import "github.com/gin-gonic/gin"

// RegistrationServiceInterface defines the interface for registration service
type RegistrationServiceInterface interface {
	InitiateUserRegistration(c *gin.Context)
	CompleteUserRegistration(c *gin.Context)
}

type User interface {
	Login(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfilePicture(c *gin.Context)
	ChangePassword(c *gin.Context)
	ForgetPassword(c *gin.Context)
	VerifyResetPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ConfirmUpdateProfile(c *gin.Context)
	HandleGoogleOauthReq(c *gin.Context)
	HandleGoogleOauthRes(c *gin.Context)
	FacebookLoginReq(c *gin.Context)
	HandleFacebookOauthRes(c *gin.Context)
	BlockAccount(c *gin.Context)
	GetBlockedAccount(c *gin.Context)
	AddIpFilter(c *gin.Context)
	GetIpFilter(c *gin.Context)
	AdminUpdateProfile(c *gin.Context)
	AdminResetUsersPassword(c *gin.Context)
	GetUsers(c *gin.Context)
	RemoveIPFilter(c *gin.Context)
	GetMyReferalCodes(c *gin.Context)
	GetMyRefferedUsersAndPoints(c *gin.Context)
	GetCurrentReferralMultiplier(c *gin.Context)
	UpdateReferralMultiplier(c *gin.Context)
	UpdateUsersPointsForReferrances(c *gin.Context)
	GetAdminAssignedPoints(c *gin.Context)
	GetUserPoints(c *gin.Context)
	AdminRegisterPlayer(c *gin.Context)
	AdminLogin(c *gin.Context)
	UpdateSignupBonus(c *gin.Context)
	GetSignupBonus(c *gin.Context)
	UpdateReferralBonus(c *gin.Context)
	GetReferralBonus(c *gin.Context)
	RefreshToken(c *gin.Context)
	VerifyUser(c *gin.Context)
	ReSendVerificationOTP(c *gin.Context)
	GetOtp(c *gin.Context)
	GetAdmins(c *gin.Context)
}

type Analytics interface {
	GetUserTransactions(c *gin.Context)
	GetUserAnalytics(c *gin.Context)
	GetRealTimeStats(c *gin.Context)
	GetDailyReport(c *gin.Context)
	GetTopGames(c *gin.Context)
	GetTopPlayers(c *gin.Context)
	GetUserBalanceHistory(c *gin.Context)
}

type OTP interface {
	CreateEmailVerification(c *gin.Context)
	VerifyOTP(c *gin.Context)
	ResendOTP(c *gin.Context)
	GetOTPInfo(c *gin.Context)
	InvalidateOTP(c *gin.Context)
	CleanupExpiredOTPs(c *gin.Context)
	GetOTPStats(c *gin.Context)
}

type OpeartionalGroup interface {
	CreateOperationalGroup(c *gin.Context)
	GetOperationalGroups(c *gin.Context)
}

type OperationalGroupType interface {
	CreateOperationalGroupType(c *gin.Context)
	GetOperationalGroupTypesByGroupID(c *gin.Context)
	GetOperationalGroupTypes(c *gin.Context)
}

type OperationsDefinition interface {
	GetOperationsDefinitions(c *gin.Context)
}
type Balance interface {
	GetUserBalances(c *gin.Context)
	ExchangeBalance(c *gin.Context)
	GetManualFundLogs(c *gin.Context)
	ManualFunding(c *gin.Context)
	CreditWallet(c *gin.Context)
}

type BalanceLogs interface {
	GetBalanceLogs(c *gin.Context)
	GetBalanceLogByID(c *gin.Context)
	GetBalanceLogsForAdmin(c *gin.Context)
}

type Exchange interface {
	GetExcahnge(c *gin.Context)
	GetAvailableCurrencies(c *gin.Context)
}

type WS interface {
	HandleWS(c *gin.Context)
	SinglePlayerStreamWS(c *gin.Context)
	PlayerLevelWS(c *gin.Context)
	NotificationWS(c *gin.Context)
	SessionWS(c *gin.Context)
	PlayerProgressBarWS(c *gin.Context)
	InitiateTriggerSquadsProgressBar(c *gin.Context)
	UserBalanceWS(c *gin.Context)
}

type Bet interface {
	GetOpenRound(c *gin.Context)
	PlaceBet(c *gin.Context)
	CashOut(c *gin.Context)
	GetBetHistory(c *gin.Context)
	CancelBet(c *gin.Context)
	GetLeaders(c *gin.Context)
	GetMyBetHistory(c *gin.Context)
	GetAllFailedRounds(c *gin.Context)
	ManualRefundFailedRound(c *gin.Context)
	GetPlinkoGameConfig(c *gin.Context)
	PlacePlinkoBet(c *gin.Context)
	GetUserPlinkoBetHistory(c *gin.Context)
	GetPlinkoGameStats(c *gin.Context)
	CreateLeague(c *gin.Context)
	GetLeagues(c *gin.Context)
	CreateClub(c *gin.Context)
	GetClubs(c *gin.Context)
	UpdateFootballCardMultiplierValue(c *gin.Context)
	GetFootballCardMultiplier(c *gin.Context)
	CreateFootballMatchRound(c *gin.Context)
	GetFootballMatchRounds(c *gin.Context)
	CreateFootballMatch(c *gin.Context)
	GetFootballRoundMatchs(c *gin.Context)
	GetCurrentFootballRound(c *gin.Context)
	CloseFootballMatch(c *gin.Context)
	UpdateFootballRoundPrice(c *gin.Context)
	GetFootballRoundPrice(c *gin.Context)
	PleceBetOnFootballRound(c *gin.Context)
	GetUserFootballBets(c *gin.Context)
	CreateStreetKingsGame(c *gin.Context)
	CashOutStreetKingsBet(c *gin.Context)
	GetStreetkingHistory(c *gin.Context)
	SetCrytoKingsConfig(c *gin.Context)
	PlaceCryptoKingsBet(c *gin.Context)
	GetCryptoKingsBetHistory(c *gin.Context)
	GetCryptoKingsCurrentCryptoPrice(c *gin.Context)
	PlaceQuickHustleBet(c *gin.Context)
	UserSelectCard(c *gin.Context)
	GetQuickHustleBetHistory(c *gin.Context)
	CreateRollDaDice(c *gin.Context)
	GetRollDaDiceHistory(c *gin.Context)
	GetScratchGamePrice(c *gin.Context)
	PlaceScratchCardBet(c *gin.Context)
	GetUserScratchCardBetHistories(c *gin.Context)
	GetSpinningWheelPrice(c *gin.Context)
	PlaceSpinningWheelBet(c *gin.Context)
	GetSpinningWheelUserBetHistory(c *gin.Context)
	UpdateGame(c *gin.Context)
	GetGames(c *gin.Context)
	DisableAllGames(c *gin.Context)
	GetAvailableGames(c *gin.Context)
	DeleteGame(c *gin.Context)
	AddGame(c *gin.Context)
	UpdateGameStatus(c *gin.Context)
	CreateSpinningWheelMysteries(c *gin.Context)
	GetSpinningWheelMysteries(c *gin.Context)
	UpdateSpinningWheelMystery(c *gin.Context)
	DeleteSpinningWheelMystery(c *gin.Context)
	CreateSpinningWheelConfig(c *gin.Context)
	GetSpinningWheelConfigs(c *gin.Context)
	UpdateSpinningWheelConfig(c *gin.Context)
	DeleteSpinningWheelConfig(c *gin.Context)
	UpdateBetIcon(c *gin.Context)
	GetScratchCardsConfig(c *gin.Context)
	UpdateScratchGameConfig(c *gin.Context)
	CreateLevel(c *gin.Context)
	GetLevels(c *gin.Context)
	CreateLevelRequirements(c *gin.Context)
	UpdateLevelRequirement(c *gin.Context)
	GetUserLevel(c *gin.Context)
	AddFakeBalanceLog(c *gin.Context)
	CreateLootBox(c *gin.Context)
	UpdateLootBox(c *gin.Context)
	DeleteLootBox(c *gin.Context)
	GetLootBox(c *gin.Context)
	PlaceLootBoxBet(c *gin.Context)
	SelectLootBox(c *gin.Context)
}

type Departements interface {
	CreateDepartement(c *gin.Context)
	GetDepartement(c *gin.Context)
	UpdateDepartment(c *gin.Context)
	AssignUserToDepartment(c *gin.Context)
}

type Performance interface {
	GetFinancialMetrics(c *gin.Context)
	GameMatrics(c *gin.Context)
}

type Authz interface {
	GetPermissions(c *gin.Context)
	CreateRole(c *gin.Context)
	GetRoles(c *gin.Context)
	UpdateRolePermissions(c *gin.Context)
	RemoveRole(c *gin.Context)
	AssignRoleToUser(c *gin.Context)
	RevokeUserRole(c *gin.Context)
	GetRoleUsers(c *gin.Context)
	GetUserRoles(c *gin.Context)
	// Crypto Wallet Methods
	ConnectWallet(c *gin.Context)
	DisconnectWallet(c *gin.Context)
	GetUserWallets(c *gin.Context)
	CreateWalletChallenge(c *gin.Context)
	VerifyWalletChallenge(c *gin.Context)
	LoginWithWallet(c *gin.Context)
	TestWalletSignature(c *gin.Context)
}

type AirtimeProvider interface {
	RefereshAirtimeUtilities(c *gin.Context)
	GetAvailableAirtime(c *gin.Context)
	UpdateAirtimeStatus(c *gin.Context)
	UpdateAirtimeUtilityPrice(c *gin.Context)
	ClaimPoints(c *gin.Context)
	GetActiveAvailableAirtime(c *gin.Context)
	GetUserAirtimeTransactions(c *gin.Context)
	GetAllAirtimeUtilitiesTransactions(c *gin.Context)
	UpdateAirtimeAmount(c *gin.Context)
	GetAirtimeUtilitiesStats(c *gin.Context)
}

type SystemLogs interface {
	GetSystemLogs(c *gin.Context)
	GetAvailableLogModules(c *gin.Context)
}

type Company interface {
	CreateCompany(c *gin.Context)
	GetCompanyByID(c *gin.Context)
	GetCompanies(c *gin.Context)
	UpdateCompany(c *gin.Context)
	DeleteCompany(c *gin.Context)
	AddIP(c *gin.Context)
}

type Report interface {
	GetDailyReport(c *gin.Context)
}

type Squads interface {
	CreateSquad(c *gin.Context)
	GetMySquads(c *gin.Context)
	GetMyOwnSquads(c *gin.Context)
	GetSquadsByType(c *gin.Context)
	UpdateSquadHandle(c *gin.Context)
	DeleteSquad(c *gin.Context)
	CreateSquadMember(c *gin.Context)
	GetSquadMembersBySquadID(c *gin.Context)
	DeleteSquadMember(c *gin.Context)
	DeleteSquadMembersBySquadID(c *gin.Context)
	GetSquadEarns(c *gin.Context)
	GetMySquadEarns(c *gin.Context)
	GetSquadTotalEarns(c *gin.Context)
	GetSquadByName(c *gin.Context)
	GetTornamentStyleRanking(c *gin.Context)
	CreateTournaments(c *gin.Context)
	GetTornamentStyles(c *gin.Context)
	GetSquadMembersEarnings(c *gin.Context)
	LeaveSquad(c *gin.Context)
	JoinSquad(c *gin.Context)
	GetSquadWaitlist(c *gin.Context)
	RemoveWaitingSquadWaitingUser(c *gin.Context)
	ApproveWaitingSquadMember(c *gin.Context)
}

type Notification interface {
	CreateNotification(c *gin.Context)
	GetUserNotifications(c *gin.Context)
	MarkNotificationRead(c *gin.Context)
	MarkAllNotificationsRead(c *gin.Context)
	DeleteNotification(c *gin.Context)
}
type Adds interface {
	SignIn(c *gin.Context)
	UpdateBalance(c *gin.Context)
	SaveAddsService(c *gin.Context)
	GetAddsServices(c *gin.Context)
}

type Banner interface {
	GetAllBanners(c *gin.Context)
	GetBannerByPage(c *gin.Context)
	UpdateBanner(c *gin.Context)
	CreateBanner(c *gin.Context)
	DeleteBanner(c *gin.Context)
	UploadBannerImage(c *gin.Context)
}

type Lottery interface {
	CreateLotteryService(c *gin.Context)
	CreateLotteryRequest(c *gin.Context)
	CheckUserBalanceAndDeductBalance(c *gin.Context)
}

type SportsService interface {
	SignIn(c *gin.Context)
	PlaceBet(c *gin.Context)
	AwardWinnings(c *gin.Context)
}

type SportBets interface {
	PlaceSportBet(c *gin.Context)
	GetSportBetByID(c *gin.Context)
	GetUserSportBets(c *gin.Context)
	UpdateSportBetStatus(c *gin.Context)
	GetUserBettingStats(c *gin.Context)
	CreateSportMatch(c *gin.Context)
	UpdateSportMatchStatus(c *gin.Context)
}

type RiskSettings interface {
	GetRiskSettings(c *gin.Context)
	SetRiskSettings(c *gin.Context)
}

type Agent interface {
	CreateAgentReferralLink(c *gin.Context)
	GetAgentReferrals(c *gin.Context)
	GetReferralStats(c *gin.Context)
	CreateAgentProvider(c *gin.Context)
}
