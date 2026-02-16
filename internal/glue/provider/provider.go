package provider

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
	providerHandler handler.Provider,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	providerRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/providers",
			Handler: providerHandler.CreateProvider,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "create provider", http.MethodPost),
				// middleware.SystemLogs("Create Provider", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/providers",
			Handler: providerHandler.GetAllProviders,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "get all providers", http.MethodGet),
				// middleware.SystemLogs("Get All Providers", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/providers/:id",
			Handler: providerHandler.UpdateProvider,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "update provider", http.MethodPatch),
				// middleware.SystemLogs("Update Provider", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/providers/:id",
			Handler: providerHandler.DeleteProvider,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "delete provider", http.MethodDelete),
				// middleware.SystemLogs("Delete Provider", &log, systemLogs),
			},
		},
	}

	routing.RegisterRoute(group, providerRoutes, log)
}
