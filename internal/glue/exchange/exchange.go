package exchange

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	op handler.Exchange,
) {

	exhcnageRoute := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/conversion/rate",
			Handler: op.GetExcahnge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/currencies",
			Handler: op.GetAvailableCurrencies,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}
	routing.RegisterRoute(group, exhcnageRoute, log)
}
