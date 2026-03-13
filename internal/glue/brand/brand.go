package brand

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
	brand handler.Brand,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	brandRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands",
			Handler: brand.CreateBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create brand", http.MethodPost),
				middleware.SystemLogs("Create Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands",
			Handler: brand.GetBrands,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view brand management", http.MethodGet),
				middleware.SystemLogs("Get All Brands", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id",
			Handler: brand.GetBrandByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view brand management", http.MethodGet),
				middleware.SystemLogs("Get Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/brands/:id",
			Handler: brand.UpdateBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPatch),
				middleware.SystemLogs("Update Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id",
			Handler: brand.DeleteBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete brand", http.MethodDelete),
				middleware.SystemLogs("Delete Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/brands/:id/status",
			Handler: brand.ChangeBrandStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPatch),
				middleware.SystemLogs("Change Brand Status", &log, systemLogs),
			},
		},
		// Credentials
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/credentials",
			Handler: brand.CreateBrandCredential,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Create Brand Credential", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/credentials/:credentialId/rotate",
			Handler: brand.RotateBrandCredential,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Rotate Brand Credential", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id/credentials/:credentialId",
			Handler: brand.GetBrandCredentialByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view brand management", http.MethodGet),
				middleware.SystemLogs("Get Brand Credential", &log, systemLogs),
			},
		},
		// Allowed origins
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/allowed-origins",
			Handler: brand.AddBrandAllowedOrigin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Add Brand Allowed Origin", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id/allowed-origins/:originId",
			Handler: brand.RemoveBrandAllowedOrigin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodDelete),
				middleware.SystemLogs("Remove Brand Allowed Origin", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id/allowed-origins",
			Handler: brand.ListBrandAllowedOrigins,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view brand management", http.MethodGet),
				middleware.SystemLogs("List Brand Allowed Origins", &log, systemLogs),
			},
		},
		// Feature flags
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id/feature-flags",
			Handler: brand.GetBrandFeatureFlags,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "view brand management", http.MethodGet),
				middleware.SystemLogs("Get Brand Feature Flags", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/brands/:id/feature-flags",
			Handler: brand.UpdateBrandFeatureFlags,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPut),
				middleware.SystemLogs("Update Brand Feature Flags", &log, systemLogs),
			},
		},
		// Game assignments
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/games",
			Handler: brand.AssignGamesToBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Assign Games To Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id/games",
			Handler: brand.RevokeGamesFromBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodDelete),
				middleware.SystemLogs("Revoke Games From Brand", &log, systemLogs),
			},
		},
		// Provider assignments
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/providers",
			Handler: brand.AssignProviderToBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Assign Provider To Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id/providers/:providerId",
			Handler: brand.RevokeProviderFromBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodDelete),
				middleware.SystemLogs("Revoke Provider From Brand", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, brandRoutes, log)
}
