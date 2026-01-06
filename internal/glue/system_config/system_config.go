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
		// Alert Management Routes
		{
			Method:  "GET",
			Path:    "/api/admin/system-config/alerts/configurations",
			Handler: systemConfigHandler.GetAlertConfigurations,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/system-config/alerts/configurations",
			Handler: systemConfigHandler.CreateAlertConfiguration,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/system-config/alerts/configurations/:id",
			Handler: systemConfigHandler.UpdateAlertConfiguration,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "DELETE",
			Path:    "/api/admin/system-config/alerts/configurations/:id",
			Handler: systemConfigHandler.DeleteAlertConfiguration,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/system-config/alerts/triggers",
			Handler: systemConfigHandler.GetAlertTriggers,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/system-config/alerts/triggers/:id/acknowledge",
			Handler: systemConfigHandler.AcknowledgeAlert,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		// Settings Management Routes
		{
			Method:  "GET",
			Path:    "/api/admin/settings/general",
			Handler: systemConfigHandler.GetGeneralSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/general",
			Handler: systemConfigHandler.UpdateGeneralSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/payments",
			Handler: systemConfigHandler.GetPaymentSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/payments",
			Handler: systemConfigHandler.UpdatePaymentSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/tips",
			Handler: systemConfigHandler.GetTipSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/tips",
			Handler: systemConfigHandler.UpdateTipSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/welcome-bonus",
			Handler: systemConfigHandler.GetWelcomeBonusSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/welcome-bonus",
			Handler: systemConfigHandler.UpdateWelcomeBonusSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/welcome-bonus/channels",
			Handler: systemConfigHandler.GetWelcomeBonusChannels,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/settings/welcome-bonus/channels",
			Handler: systemConfigHandler.CreateWelcomeBonusChannel,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/welcome-bonus/channels/:id",
			Handler: systemConfigHandler.UpdateWelcomeBonusChannel,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "DELETE",
			Path:    "/api/admin/settings/welcome-bonus/channels/:id",
			Handler: systemConfigHandler.DeleteWelcomeBonusChannel,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/security",
			Handler: systemConfigHandler.GetSecuritySettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/security",
			Handler: systemConfigHandler.UpdateSecuritySettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/settings/geo-blocking",
			Handler: systemConfigHandler.GetGeoBlockingSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/geo-blocking",
			Handler: systemConfigHandler.UpdateGeoBlockingSettings,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		// Game Import Routes
		{
			Method:  "GET",
			Path:    "/api/admin/settings/game-import/config",
			Handler: systemConfigHandler.GetGameImportConfig,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/admin/settings/game-import/config",
			Handler: systemConfigHandler.UpdateGameImportConfig,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/settings/game-import/trigger",
			Handler: systemConfigHandler.TriggerGameImport,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(group, systemConfigRoutes, *log)
}
