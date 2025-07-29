package adds

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
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	adds handler.Adds,
	systemLog module.SystemLogs,
) {
	addsRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/adds/services",
			Handler: adds.GetAddsServices,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get adds services", http.MethodGet),
				middleware.SystemLogs("Get Adds Services", &log, systemLog),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/adds/services",
			Handler: adds.SaveAddsService,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "create adds service", http.MethodPost),
				middleware.SystemLogs("Create Adds Service", &log, systemLog),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/adds/signin",
			Handler: adds.SignIn,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/adds/balance/update",
			Handler: adds.UpdateBalance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.AddsAuth(),
			},
		},
	}

	routing.RegisterRoute(group, addsRoutes, log)
}
