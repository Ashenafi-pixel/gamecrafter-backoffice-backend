package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"

	"go.uber.org/zap"
)

// UserHandler defines the interface for user HTTP handlers
type UserHandler interface {
	RegisterUser(c *gin.Context)
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
	AdminAutoResetUsersPassword(c *gin.Context)
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
	GetPlayerDetails(c *gin.Context)
	GetPlayerManualFunds(c *gin.Context)
	GetReferralBonus(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	VerifyUser(c *gin.Context)
	ReSendVerificationOTP(c *gin.Context)
	GetOtp(c *gin.Context)
	GetAdmins(c *gin.Context)
	GetAdminUsers(c *gin.Context)
	CreateAdminUser(c *gin.Context)
	UpdateAdminUser(c *gin.Context)
	DeleteAdminUser(c *gin.Context)
	SuspendAdminUser(c *gin.Context)
	UnsuspendAdminUser(c *gin.Context)
	// Enterprise Registration Methods
	InitiateEnterpriseRegistration(c *gin.Context)
	CompleteEnterpriseRegistration(c *gin.Context)
	GetEnterpriseRegistrationStatus(c *gin.Context)
	ResendEnterpriseVerificationEmail(c *gin.Context)
	// Regular Registration with Email Verification Methods
	InitiateUserRegistration(c *gin.Context)
	CompleteUserRegistration(c *gin.Context)
	ResendVerificationEmail(c *gin.Context)
	ServeVerificationPage(c *gin.Context)
	// Service Management
	SetRegistrationService(service interface{})
}

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	user UserHandler,
	userModule module.User,
	authModule module.Authz,
	systemLog module.SystemLogs,

) {

	authRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/register",
			Handler: user.InitiateUserRegistration, // Use new email verification registration
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/verify",
			Handler: user.ServeVerificationPage, // Add verification page endpoint
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/register/complete",
			Handler: user.CompleteUserRegistration, // Add completion endpoint
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/register/resend-verification",
			Handler: user.ResendVerificationEmail, // Add resend verification endpoint
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: user.Login,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/user/profile",
			Handler: user.GetProfile,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/profile/picture",
			Handler: user.UpdateProfilePicture,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/user/password",
			Handler: user.ChangePassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/password/forget",
			Handler: user.ForgetPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/password/forget/verify",
			Handler: user.VerifyResetPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/password/reset",
			Handler: user.ResetPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/profile",
			Handler: user.UpdateProfile,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/profile/confirm",
			Handler: user.ConfirmUpdateProfile,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/user/oauth/google",
			Handler: user.HandleGoogleOauthReq,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/user/oauth/google/res",
			Handler: user.HandleGoogleOauthRes,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/user/oauth/facebook",
			Handler: user.FacebookLoginReq,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/user/oauth/facebook/res",
			Handler: user.HandleFacebookOauthRes,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/block",
			Handler: user.BlockAccount,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "block player", http.MethodPost),
				middleware.SystemLogs("block user account", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/block/accounts",
			Handler: user.GetBlockedAccount,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view players", http.MethodPost),
				middleware.SystemLogs("get blocked account", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/ipfilters",
			Handler: user.AddIpFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "add ip filter", http.MethodPost),
				middleware.SystemLogs("add ip filter", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/ipfilters",
			Handler: user.GetIpFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view ip filters", http.MethodGet),
				middleware.SystemLogs("get ip filters", &log, systemLog),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/users",
			Handler: user.AdminUpdateProfile,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "edit player", http.MethodPatch),
				middleware.SystemLogs("update user profile", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users",
			Handler: user.GetUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view players", http.MethodPost),
				middleware.SystemLogs("get users", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/search",
			Handler: user.GetUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view players", http.MethodPost),
				middleware.SystemLogs("get users", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/password",
			Handler: user.AdminResetUsersPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "reset user account password", http.MethodPost),
				middleware.SystemLogs("reset user account password", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/password/auto-reset",
			Handler: user.AdminAutoResetUsersPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "reset user account password", http.MethodPost),
				middleware.SystemLogs("auto reset user account password", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/players/:user_id/details",
			Handler: user.GetPlayerDetails,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view player details", http.MethodGet),
				middleware.SystemLogs("get player details", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/players/:user_id/manual-funds",
			Handler: user.GetPlayerManualFunds,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list (manual funding is value-based permission)
				middleware.Authz(authModule, "manual fund player", http.MethodGet),
				middleware.SystemLogs("get player manual funds", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/ipfilters",
			Handler: user.RemoveIPFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "remove ip filter", http.MethodDelete),
				middleware.SystemLogs("remove ip filter", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/user/referral",
			Handler: user.GetMyReferalCodes,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/user/referral/users",
			Handler: user.GetMyRefferedUsersAndPoints,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/referrals",
			Handler: user.GetCurrentReferralMultiplier,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get referral multiplier", http.MethodGet),
				middleware.SystemLogs("get referral multiplier", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/referrals",
			Handler: user.UpdateReferralMultiplier,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update referral multiplier", http.MethodPost),
				middleware.SystemLogs("update referral multiplier", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/referrals/users",
			Handler: user.UpdateUsersPointsForReferrances,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "add point to users", http.MethodPost),
				middleware.SystemLogs("add point to users", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/referrals/users",
			Handler: user.GetAdminAssignedPoints,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get point to users", http.MethodGet),
				middleware.SystemLogs("get point to users", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/users/points",
			Handler: user.GetUserPoints,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/players/register",
			Handler: user.AdminRegisterPlayer,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "register players", http.MethodPost),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/login",
			Handler: user.AdminLogin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admins",
			Handler: user.GetAdmins,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get admins", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/users_admin",
			Handler: user.GetAdminUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view admin users", http.MethodGet),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users_admin",
			Handler: user.CreateAdminUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create admin user", http.MethodPost),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/users_admin/:id",
			Handler: user.UpdateAdminUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "edit admin user", http.MethodPut),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/users_admin/:id",
			Handler: user.DeleteAdminUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete admin user", http.MethodDelete),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users_admin/:id/suspend",
			Handler: user.SuspendAdminUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "suspend admin user", http.MethodPost),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users_admin/:id/unsuspend",
			Handler: user.UnsuspendAdminUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "unsuspend admin user", http.MethodPost),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/signup/bonus",
			Handler: user.UpdateSignupBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update signup bonus", http.MethodPut),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/signup/bonus",
			Handler: user.GetSignupBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get signup bonus", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/referral/bonus/config",
			Handler: user.UpdateSignupBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update referral bonus", http.MethodPut),
				middleware.SystemLogs("update referral bonus", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/referral/bonus/config",
			Handler: user.GetReferralBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get referral bonus", http.MethodGet),
				middleware.SystemLogs("get referral bonus", &log, systemLog),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/refresh",
			Handler: user.RefreshToken,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/refresh",
			Handler: user.RefreshToken,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/auth/logout",
			Handler: user.Logout,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/verify",
			Handler: user.VerifyUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/resend/verification/otp",
			Handler: user.ReSendVerificationOTP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/user/otp",
			Handler: user.GetOtp,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}

	routing.RegisterRoute(group, authRoutes, log)

	// Initialize enterprise registration routes
	InitEnterpriseRegistration(group, log, user)
}
