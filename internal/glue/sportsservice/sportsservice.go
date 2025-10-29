package sportsservice

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
	sportsService handler.SportsService,
) {
	sportsServiceRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/sports/signin",
			Handler: sportsService.SignIn,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/sports/placebet",
			Handler: sportsService.PlaceBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.SportsAuth(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/sports/awardwinnings",
			Handler: sportsService.AwardWinnings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.SportsAuth(),
			},
		},
	}

	routing.RegisterRoute(group, sportsServiceRoutes, log)
}
