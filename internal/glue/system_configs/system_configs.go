package system_configs

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/system_configs"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger) {
	systemConfigsHandler := system_configs.NewSystemConfigsHandler(log)

	systemConfigsRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/system-configs",
			Handler: systemConfigsHandler.GetSystemConfigs,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	for _, route := range systemConfigsRoutes {
		group.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}
