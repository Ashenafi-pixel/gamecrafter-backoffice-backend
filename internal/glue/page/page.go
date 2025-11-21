package page

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/page"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(grp *gin.RouterGroup, log *zap.Logger, pageHandler page.PageHandler, pageModule module.Page) {
	pageRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/pages",
			Handler: pageHandler.GetAllPages,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/users/:id/pages",
			Handler: pageHandler.GetUserAllowedPages,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/users/:id/pages",
			Handler: pageHandler.UpdateUserAllowedPages,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(grp, pageRoutes, *log)
}

