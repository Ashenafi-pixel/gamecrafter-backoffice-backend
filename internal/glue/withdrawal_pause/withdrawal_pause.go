package withdrawal_pause

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
	withdrawalPauseHandler handler.WithdrawalPause,
) {
	withdrawalPauseRoutes := []routing.Route{
		// Pause settings management
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/withdrawal-pause/settings",
			Handler: withdrawalPauseHandler.GetWithdrawalPauseSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/v1/withdrawal-pause/settings",
			Handler: withdrawalPauseHandler.UpdateWithdrawalPauseSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},

		// Threshold management
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/withdrawal-pause/thresholds",
			Handler: withdrawalPauseHandler.GetWithdrawalThresholds,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/v1/withdrawal-pause/thresholds",
			Handler: withdrawalPauseHandler.CreateWithdrawalThreshold,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/v1/withdrawal-pause/thresholds/:id",
			Handler: withdrawalPauseHandler.UpdateWithdrawalThreshold,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/v1/withdrawal-pause/thresholds/:id",
			Handler: withdrawalPauseHandler.DeleteWithdrawalThreshold,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},

		// Paused withdrawals management
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/withdrawal-pause/paused-withdrawals",
			Handler: withdrawalPauseHandler.GetPausedWithdrawals,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/v1/withdrawal-pause/withdrawals/:id/action",
			Handler: withdrawalPauseHandler.ApproveWithdrawal,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},

		// Dashboard and statistics
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/withdrawal-pause/stats",
			Handler: withdrawalPauseHandler.GetWithdrawalPauseStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/withdrawal-pause/status",
			Handler: withdrawalPauseHandler.GetWithdrawalPauseStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, withdrawalPauseRoutes, log)
}





