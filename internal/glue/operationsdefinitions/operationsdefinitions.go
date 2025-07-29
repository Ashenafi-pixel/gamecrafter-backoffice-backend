package operationsdefinitions

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
	op handler.OperationsDefinition,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
) {

	operationdefintionsroutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/operations/definitions",
			Handler: op.GetOperationsDefinitions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, operationdefintionsroutes, log)
}
