package operationalgrouptype

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
	op handler.OperationalGroupType,
	authModule module.Authz,
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
				middleware.Authz(authModule, "add operations definitions", http.MethodPost),
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
				middleware.Authz(authModule, "get operations definitions", http.MethodGet),
				middleware.SystemLogs("Get Operational Group Types", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, operationalGroupRouts, log)
}
