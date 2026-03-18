package operator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	grp *gin.RouterGroup,
	log zap.Logger,
	operatorHandler handler.Operator,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	routes := []struct {
		Method     string
		Path       string
		Handler    gin.HandlerFunc
		Middleware []gin.HandlerFunc
	}{
		// Core CRUD
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators",
			Handler: operatorHandler.CreateOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Create Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/operators",
			Handler: operatorHandler.GetOperators,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view operator management", http.MethodGet),
				middleware.SystemLogs("Get Operators", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/operators/:id",
			Handler: operatorHandler.GetOperatorByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view operator management", http.MethodGet),
				middleware.SystemLogs("Get Operator By ID", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/operators/:id",
			Handler: operatorHandler.UpdateOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPatch),
				middleware.SystemLogs("Update Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/operators/:id",
			Handler: operatorHandler.DeleteOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodDelete),
				middleware.SystemLogs("Delete Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/operators/:id/status",
			Handler: operatorHandler.ChangeOperatorStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPatch),
				middleware.SystemLogs("Change Operator Status", &log, systemLogs),
			},
		},
		// Credentials
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/credentials",
			Handler: operatorHandler.CreateOperatorCredential,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Create Operator Credential", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/credentials/:credentialId/rotate",
			Handler: operatorHandler.RotateOperatorCredential,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Rotate Operator Credential", &log, systemLogs),
			},
		},
		// Game assignments
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/games",
			Handler: operatorHandler.AssignGamesToOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Assign Games To Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/operators/:id/games",
			Handler: operatorHandler.GetOperatorGames,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view operator management", http.MethodGet),
				middleware.SystemLogs("Get Operator Games", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/games/all",
			Handler: operatorHandler.AssignAllGamesToOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Assign All Games To Operator", &log, systemLogs),
			},
		},
		// Provider assignments
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/providers",
			Handler: operatorHandler.AssignProviderToOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Assign Provider To Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/operators/:id/providers/:providerId",
			Handler: operatorHandler.RevokeProviderFromOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodDelete),
				middleware.SystemLogs("Revoke Provider From Operator", &log, systemLogs),
			},
		},
		// Allowed origins
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/allowed-origins",
			Handler: operatorHandler.AddOperatorAllowedOrigin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPost),
				middleware.SystemLogs("Add Operator Allowed Origin", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/operators/:id/allowed-origins",
			Handler: operatorHandler.ListOperatorAllowedOrigins,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view operator management", http.MethodGet),
				middleware.SystemLogs("List Operator Allowed Origins", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/operators/:id/allowed-origins/:originId",
			Handler: operatorHandler.RemoveOperatorAllowedOrigin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodDelete),
				middleware.SystemLogs("Remove Operator Allowed Origin", &log, systemLogs),
			},
		},
		// Feature flags
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/operators/:id/feature-flags",
			Handler: operatorHandler.GetOperatorFeatureFlags,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view operator management", http.MethodGet),
				middleware.SystemLogs("Get Operator Feature Flags", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/operators/:id/feature-flags",
			Handler: operatorHandler.UpdateOperatorFeatureFlags,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodPut),
				middleware.SystemLogs("Update Operator Feature Flags", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/operators/:id/games",
			Handler: operatorHandler.RevokeGamesFromOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit operator", http.MethodDelete),
				middleware.SystemLogs("Revoke Games From Operator", &log, systemLogs),
			},
		},
	}

	for _, route := range routes {
		grp.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}

