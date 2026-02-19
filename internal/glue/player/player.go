package player

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

// PlayerHandler defines the interface for player HTTP handlers
type PlayerHandler interface {
	CreatePlayer(c *gin.Context)
	GetPlayerByID(c *gin.Context)
	GetPlayers(c *gin.Context)
	UpdatePlayer(c *gin.Context)
	DeletePlayer(c *gin.Context)
}

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	player PlayerHandler,
	authModule module.Authz,
	systemLog module.SystemLogs,
) {
	playerRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/player-management",
			Handler: player.CreatePlayer,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "create player", http.MethodPost),
				// middleware.SystemLogs("create player", &log, systemLog),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/player-management/:id",
			Handler: player.GetPlayerByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "view players", http.MethodGet),
				// middleware.SystemLogs("get player by ID", &log, systemLog),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/player-management",
			Handler: player.GetPlayers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "view players", http.MethodGet),
				// middleware.SystemLogs("get players", &log, systemLog),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/player-management/:id",
			Handler: player.UpdatePlayer,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "edit player", http.MethodPatch),
				// middleware.SystemLogs("update player", &log, systemLog),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/player-management/:id",
			Handler: player.DeletePlayer,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// middleware.Authz(authModule, "delete player", http.MethodDelete),
				// middleware.SystemLogs("delete player", &log, systemLog),
			},
		},
	}

	routing.RegisterRoute(group, playerRoutes, log)
}
