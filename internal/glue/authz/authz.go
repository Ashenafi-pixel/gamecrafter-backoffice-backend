package authz

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	authzModule handler.Authz,
	authModule module.Authz,
	systemLog module.SystemLogs,
) {

	authzRoute := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/permissions",
			Handler: authzModule.GetPermissions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view permissions", http.MethodGet),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/roles",
			Handler: authzModule.CreateRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create role", http.MethodPost),
				middleware.SystemLogs("create role", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/roles",
			Handler: authzModule.GetRoles,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view roles", http.MethodGet),
				middleware.SystemLogs("get roles", &log, systemLog),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/roles",
			Handler: authzModule.UpdateRolePermissions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "edit role", http.MethodPatch),
				middleware.SystemLogs("update role permissions", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/roles",
			Handler: authzModule.RemoveRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "delete role", http.MethodDelete),
				middleware.SystemLogs("remove role", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/roles",
			Handler: authzModule.AssignRoleToUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "assign role", http.MethodPost),
				middleware.SystemLogs("assign role", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/users/roles",
			Handler: authzModule.RevokeUserRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "revoke role", http.MethodDelete),
				middleware.SystemLogs("revoke role", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/roles/:id/users",
			Handler: authzModule.GetRoleUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view roles", http.MethodGet),
				middleware.SystemLogs("get role users", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/users/:id/roles",
			Handler: authzModule.GetUserRoles,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view roles", http.MethodGet),
				middleware.SystemLogs("get user roles", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/rbac/users",
			Handler: authzModule.SearchAdminUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view admin users", http.MethodGet),
				middleware.SystemLogs("search admin users", &log, systemLog),
			},
		},
		// Crypto Wallet Routes
		{
			Method:  http.MethodPost,
			Path:    "/api/wallet/connect",
			Handler: authzModule.ConnectWallet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.SystemLogs("connect wallet", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/wallet/disconnect/:connection_id",
			Handler: authzModule.DisconnectWallet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.SystemLogs("disconnect wallet", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/wallet/list",
			Handler: authzModule.GetUserWallets,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.SystemLogs("get user wallets", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/wallet/challenge",
			Handler: authzModule.CreateWalletChallenge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.SystemLogs("create wallet challenge", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/wallet/verify",
			Handler: authzModule.VerifyWalletChallenge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.SystemLogs("verify wallet challenge", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/wallet/login",
			Handler: authzModule.LoginWithWallet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.SystemLogs("wallet login", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/wallet/test-signature",
			Handler: authzModule.TestWalletSignature,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.SystemLogs("test wallet signature", &log, systemLog),
			},
		},
	}
	routing.RegisterRoute(group, authzRoute, log)
}
