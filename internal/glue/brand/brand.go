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
				middleware.Authz(authModule, "get brands", http.MethodGet),
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
				middleware.Authz(authModule, "get brand", http.MethodGet),
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
				middleware.Authz(authModule, "update brand", http.MethodPatch),
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
	}
	routing.RegisterRoute(group, brandRoutes, log)
}

