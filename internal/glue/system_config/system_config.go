package system_config

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/system_config"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger, systemConfigHandler *system_config.SystemConfigHandler) {
	systemConfigRoutes := []routing.Route{
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/global-status",
			Handler: systemConfigHandler.GetWithdrawalGlobalStatus,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
            Path:    "/api/admin/system-config/withdrawal/global-status",
			Handler: systemConfigHandler.UpdateWithdrawalGlobalStatus,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/thresholds",
			Handler: systemConfigHandler.GetWithdrawalThresholds,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
            Path:    "/api/admin/system-config/withdrawal/thresholds",
			Handler: systemConfigHandler.UpdateWithdrawalThresholds,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/manual-review",
			Handler: systemConfigHandler.GetWithdrawalManualReview,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
            Path:    "/api/admin/system-config/withdrawal/manual-review",
			Handler: systemConfigHandler.UpdateWithdrawalManualReview,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/check-allowed",
			Handler: systemConfigHandler.CheckWithdrawalAllowed,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/check-thresholds",
			Handler: systemConfigHandler.CheckWithdrawalThresholds,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
            Path:    "/api/admin/system-config/withdrawal/pause-reasons",
			Handler: systemConfigHandler.GetWithdrawalPauseReasons,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(group, systemConfigRoutes, *log)
}
