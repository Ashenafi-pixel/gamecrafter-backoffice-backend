package chain_config

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/chain_config"
	"github.com/tucanbit/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger) {
	chainConfigHandler := chain_config.NewChainConfigHandler(log)

	chainConfigRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/chain-configs",
			Handler: chainConfigHandler.GetChainConfigs,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	for _, route := range chainConfigRoutes {
		group.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}
