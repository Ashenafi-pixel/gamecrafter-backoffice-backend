package airtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	airtimeHandler handler.AirtimeProvider,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	airtimeRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/airtime/refresh",
			Handler: airtimeHandler.RefereshAirtimeUtilities,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "refresh airtime utilities", http.MethodGet),
				middleware.SystemLogs("Refresh Airtime Utilities", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/airtime/utilities",
			Handler: airtimeHandler.GetAvailableAirtime,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get airtime utilities", http.MethodGet),
				middleware.SystemLogs("Get Airtime Utilities", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime",
			Handler: airtimeHandler.UpdateAirtimeStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update airtime utility status", http.MethodPut),
				middleware.SystemLogs("Update Airtime Utility Status", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime/price",
			Handler: airtimeHandler.UpdateAirtimeUtilityPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update airtime price", http.MethodPut),
				middleware.SystemLogs("Update Airtime Utility Price", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/airtime/claim",
			Handler: airtimeHandler.ClaimPoints,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/airtime/active/utilities",
			Handler: airtimeHandler.GetActiveAvailableAirtime,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/airtime/transactions",
			Handler: airtimeHandler.GetUserAirtimeTransactions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/airtime/transactions",
			Handler: airtimeHandler.GetAllAirtimeUtilitiesTransactions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get airtime transactions", http.MethodGet),
				middleware.SystemLogs("Get Airtime Transactions", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime/amount",
			Handler: airtimeHandler.UpdateAirtimeAmount,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update airtime amount", http.MethodPut),
				middleware.SystemLogs("Update Airtime Amount", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/airtime/stats",
			Handler: airtimeHandler.GetAirtimeUtilitiesStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get airtime utilities stats", http.MethodGet),
				middleware.SystemLogs("Get Airtime Utilities Stats", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, airtimeRoutes, log)
}
