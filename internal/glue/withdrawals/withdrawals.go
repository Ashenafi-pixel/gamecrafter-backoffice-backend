package withdrawals

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/withdrawals"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger, withdrawalsHandler *withdrawals.WithdrawalsHandler) {
	withdrawalsRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals",
			Handler: withdrawalsHandler.GetAllWithdrawals,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals/stats",
			Handler: withdrawalsHandler.GetWithdrawalStats,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals/id/:id",
			Handler: withdrawalsHandler.GetWithdrawalByID,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals/withdrawal-id/:withdrawal_id",
			Handler: withdrawalsHandler.GetWithdrawalByWithdrawalID,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals/user/:user_id",
			Handler: withdrawalsHandler.GetWithdrawalsByUserID,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/v1/withdrawals/date-range",
			Handler: withdrawalsHandler.GetWithdrawalsByDateRange,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(group, withdrawalsRoutes, *log)
}
