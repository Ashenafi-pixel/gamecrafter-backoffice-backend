package operator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	grp *gin.RouterGroup,
	log zap.Logger,
	operatorHandler handler.Operator,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	routes := []struct {
		Method     string
		Path       string
		Handler    gin.HandlerFunc
		Middleware []gin.HandlerFunc
	}{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/operators/:id/games",
			Handler: operatorHandler.AssignGamesToOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Reuse "edit brand" or create a dedicated "edit operator" permission as needed.
				middleware.Authz(authModule, "edit brand", http.MethodPost),
				middleware.SystemLogs("Assign Games To Operator", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/operators/:id/games",
			Handler: operatorHandler.RevokeGamesFromOperator,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "edit brand", http.MethodDelete),
				middleware.SystemLogs("Revoke Games From Operator", &log, systemLogs),
			},
		},
	}

	for _, route := range routes {
		grp.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}

