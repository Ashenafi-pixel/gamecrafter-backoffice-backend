package user

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"github.com/joshjones612/egyptkingcrash/internal/module"

	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	user handler.User,
	userModule module.User,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLog module.SystemLogs,

) {

	authRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/register",
			Handler: user.RegisterUser,
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
				middleware.Authz(authModule, enforcer, "block user account", http.MethodPost),
				middleware.SystemLogs("block user account", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/block/accounts",
			Handler: user.GetBlockedAccount,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get blocked account", http.MethodPost),
				middleware.SystemLogs("get blocked account", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/ipfilters",
			Handler: user.AddIpFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "add ip filter", http.MethodPost),
				middleware.SystemLogs("add ip filter", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/ipfilters",
			Handler: user.GetIpFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get ip filters", http.MethodGet),
				middleware.SystemLogs("get ip filters", &log, systemLog),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/users",
			Handler: user.AdminUpdateProfile,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update user profile", http.MethodPatch),
				middleware.SystemLogs("update user profile", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/password",
			Handler: user.AdminResetUsersPassword,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "reset user account password", http.MethodPost),
				middleware.SystemLogs("reset user account password", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users",
			Handler: user.GetUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get players", http.MethodPost),
				middleware.SystemLogs("get players", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/ipfilters",
			Handler: user.RemoveIPFilter,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "remove ip filter", http.MethodDelete),
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
				middleware.Authz(authModule, enforcer, "get referral multiplier", http.MethodGet),
				middleware.SystemLogs("get referral multiplier", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/referrals",
			Handler: user.UpdateReferralMultiplier,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update referral multiplier", http.MethodPost),
				middleware.SystemLogs("update referral multiplier", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/referrals/users",
			Handler: user.UpdateUsersPointsForReferrances,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "add point to users", http.MethodPost),
				middleware.SystemLogs("add point to users", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/referrals/users",
			Handler: user.GetAdminAssignedPoints,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get point to users", http.MethodGet),
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
				middleware.Authz(authModule, enforcer, "register players", http.MethodPost),
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
				middleware.Authz(authModule, enforcer, "get admins", http.MethodGet),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/signup/bonus",
			Handler: user.UpdateSignupBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update signup bonus", http.MethodPut),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/signup/bonus",
			Handler: user.GetSignupBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get signup bonus", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/referral/bonus/config",
			Handler: user.UpdateReferralBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update referral bonus", http.MethodPut),
				middleware.SystemLogs("update referral bonus", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/referral/bonus/config",
			Handler: user.GetReferralBonus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get referral bonus", http.MethodGet),
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
}
