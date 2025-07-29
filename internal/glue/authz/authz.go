package authz

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
	authzModule handler.Authz,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
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
				middleware.Authz(authModule, enforcer, "get permissions", http.MethodGet),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/roles",
			Handler: authzModule.CreateRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "create role", http.MethodPost),
				middleware.SystemLogs("create role", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/roles",
			Handler: authzModule.GetRoles,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get roles", http.MethodGet),
				middleware.SystemLogs("get roles", &log, systemLog),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/roles",
			Handler: authzModule.UpdateRolePermissions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update role permissions", http.MethodPatch),
				middleware.SystemLogs("update role permissions", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/roles",
			Handler: authzModule.RemoveRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "remove role", http.MethodDelete),
				middleware.SystemLogs("remove role", &log, systemLog),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/users/roles",
			Handler: authzModule.AssignRoleToUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "assign role", http.MethodPost),
				middleware.SystemLogs("assign role", &log, systemLog),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/users/roles",
			Handler: authzModule.RevokeUserRole,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "revoke role", http.MethodDelete),
				middleware.SystemLogs("revoke role", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/roles/:id/users",
			Handler: authzModule.GetRoleUsers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get role users", http.MethodGet),
				middleware.SystemLogs("get role users", &log, systemLog),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/users/:id/roles",
			Handler: authzModule.GetUserRoles,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get user roles", http.MethodGet),
				middleware.SystemLogs("get user roles", &log, systemLog),
			},
		},
	}
	routing.RegisterRoute(group, authzRoute, log)
}
