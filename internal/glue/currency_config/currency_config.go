package currency_config

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/currency_config"
	"github.com/tucanbit/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger) {
	currencyConfigHandler := currency_config.NewCurrencyConfigHandler(log)

	currencyConfigRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/currency-configs",
			Handler: currencyConfigHandler.GetCurrencyConfigs,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/currency-configs",
			Handler: currencyConfigHandler.CreateCurrencyConfig,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/currency-configs/:id",
			Handler: currencyConfigHandler.UpdateCurrencyConfig,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "DELETE",
			Path:    "/api/admin/currency-configs/:id",
			Handler: currencyConfigHandler.DeleteCurrencyConfig,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	for _, route := range currencyConfigRoutes {
		group.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}
