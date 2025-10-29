package operationsdefinitions

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
