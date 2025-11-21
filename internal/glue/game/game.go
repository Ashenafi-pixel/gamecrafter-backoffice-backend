package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/game"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	gameHandler *game.GameHandler,
	houseEdgeHandler *game.HouseEdgeHandler,
	authModule module.Authz,
	logsModule module.SystemLogs,
) {
	log.Info("Initializing game management routes")

	gameRoutes := []routing.Route{
		// Game Management Routes
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/game-management",
			Handler: gameHandler.CreateGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create game", http.MethodPost),
				middleware.SystemLogs("create game", &log, logsModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/game-management",
			Handler: gameHandler.GetGames,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get games", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/game-management/stats",
			Handler: gameHandler.GetGameStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get game stats", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/game-management/:id",
			Handler: gameHandler.GetGameByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get game by id", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/game-management/:id",
			Handler: gameHandler.UpdateGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update game", http.MethodPut),
				middleware.SystemLogs("update game", &log, logsModule),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/game-management/:id",
			Handler: gameHandler.DeleteGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete game", http.MethodDelete),
				middleware.SystemLogs("delete game", &log, logsModule),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/game-management/bulk-status",
			Handler: gameHandler.BulkUpdateGameStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "bulk update game status", http.MethodPut),
				middleware.SystemLogs("bulk update game status", &log, logsModule),
			},
		},

		// House Edge Management Routes
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/house-edge-management",
			Handler: houseEdgeHandler.CreateHouseEdge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create house edge", http.MethodPost),
				middleware.SystemLogs("create house edge", &log, logsModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/house-edge-management",
			Handler: houseEdgeHandler.GetHouseEdges,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get house edges", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/house-edge-management/stats",
			Handler: houseEdgeHandler.GetHouseEdgeStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get house edge stats", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/house-edge-management/:id",
			Handler: houseEdgeHandler.GetHouseEdgeByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get house edge by id", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/house-edge-management/by-game-type/:game_type",
			Handler: houseEdgeHandler.GetHouseEdgesByGameType,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get house edges by game type", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/house-edge-management/by-game-variant/:game_type/:game_variant",
			Handler: houseEdgeHandler.GetHouseEdgesByGameVariant,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get house edges by game variant", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/house-edge-management/:id",
			Handler: houseEdgeHandler.UpdateHouseEdge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update house edge", http.MethodPut),
				middleware.SystemLogs("update house edge", &log, logsModule),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/house-edge-management/:id",
			Handler: houseEdgeHandler.DeleteHouseEdge,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete house edge", http.MethodDelete),
				middleware.SystemLogs("delete house edge", &log, logsModule),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/house-edge-management/bulk-status",
			Handler: houseEdgeHandler.BulkUpdateHouseEdgeStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "bulk update house edge status", http.MethodPut),
				middleware.SystemLogs("bulk update house edge status", &log, logsModule),
			},
		},
	}

	routing.RegisterRoute(group, gameRoutes, log)
	log.Info("Game management routes initialized successfully")
}
