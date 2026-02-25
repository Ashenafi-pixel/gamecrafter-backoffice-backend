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
				// middleware.Authz(authModule, "create brand", http.MethodPost),
				// middleware.SystemLogs("Create Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands",
			Handler: brand.GetBrands,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				// middleware.Authz(authModule, "view brand management", http.MethodGet),
				// middleware.SystemLogs("Get All Brands", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id",
			Handler: brand.GetBrandByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// // Align with seeded permissions list
				// middleware.Authz(authModule, "view brand management", http.MethodGet),
				// middleware.SystemLogs("Get Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/brands/:id",
			Handler: brand.UpdateBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				// middleware.Authz(authModule, "edit brand", http.MethodPatch),
				// middleware.SystemLogs("Update Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id",
			Handler: brand.DeleteBrand,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "delete brand", http.MethodDelete),
				// middleware.SystemLogs("Delete Brand", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/brands/:id/status",
			Handler: brand.ChangeBrandStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
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
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/brands/:id/credentials/:credentialId/rotate",
			Handler: brand.RotateBrandCredential,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id/credentials/:credentialId",
			Handler: brand.GetBrandCredentialByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
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
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/brands/:id/allowed-origins/:originId",
			Handler: brand.RemoveBrandAllowedOrigin,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/brands/:id/allowed-origins",
			Handler: brand.ListBrandAllowedOrigins,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
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
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/brands/:id/feature-flags",
			Handler: brand.UpdateBrandFeatureFlags,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, brandRoutes, log)
}
