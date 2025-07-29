package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	ws handler.WS,
) {

	authRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ws",
			Handler: ws.HandleWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/single/player",
			Handler: ws.SinglePlayerStreamWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/level/player",
			Handler: ws.PlayerLevelWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/notify",
			Handler: ws.NotificationWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/session",
			Handler: ws.SessionWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/player/level/progress",
			Handler: ws.PlayerProgressBarWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/squads/progress",
			Handler: ws.InitiateTriggerSquadsProgressBar,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/ws/balance/player",
			Handler: ws.UserBalanceWS,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}
	routing.RegisterRoute(group, authRoutes, log)
}
