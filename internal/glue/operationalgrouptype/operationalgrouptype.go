package operationalgrouptype

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
	op handler.OperationalGroupType,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {

	operationalGroupRouts := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operationalgrouptype",
			Handler: op.CreateOperationalGroupType,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "add operations definitions", http.MethodPost),
				middleware.SystemLogs("Create Operational Group Type", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/operationaltypes",
			Handler: op.GetOperationalGroupTypes,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/operationalgrouptype/:groupID",
			Handler: op.GetOperationalGroupTypesByGroupID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get operations definitions", http.MethodGet),
				middleware.SystemLogs("Get Operational Group Types", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, operationalGroupRouts, log)
}
