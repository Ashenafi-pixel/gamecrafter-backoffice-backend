package operationalgroup

import (
	"net/http"

	"github.com/casbin/casbin/v2"
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
	op handler.OpeartionalGroup,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {

	operationalGroupRouts := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operationalgroup",
			Handler: op.CreateOperationalGroup,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "add operational group", http.MethodPost),
				middleware.SystemLogs("Create Operational Group", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/operationalgroup",
			Handler: op.GetOperationalGroups,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get operational group", http.MethodGet),
				middleware.SystemLogs("Get Operational Groups", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, operationalGroupRouts, log)
}
