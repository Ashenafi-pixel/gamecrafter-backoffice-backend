package withdrawal_management

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/withdrawal_management"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger, withdrawalManagementHandler *withdrawal_management.WithdrawalManagementHandler) {
	withdrawalManagementRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawal-management/paused",
			Handler: withdrawalManagementHandler.GetPausedWithdrawals,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/v1/withdrawal-management/pause/:id",
			Handler: withdrawalManagementHandler.PauseWithdrawal,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/v1/withdrawal-management/unpause/:id",
			Handler: withdrawalManagementHandler.UnpauseWithdrawal,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/v1/withdrawal-management/approve/:id",
			Handler: withdrawalManagementHandler.ApproveWithdrawal,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/v1/withdrawal-management/reject/:id",
			Handler: withdrawalManagementHandler.RejectWithdrawal,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawal-management/stats",
			Handler: withdrawalManagementHandler.GetWithdrawalPauseStats,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(group, withdrawalManagementRoutes, *log)
}
