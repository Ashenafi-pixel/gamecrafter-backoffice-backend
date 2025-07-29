package airtime

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/glue/routing"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/handler/middleware"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	airtimeHandler handler.AirtimeProvider,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
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
				middleware.Authz(authModule, enforcer, "refresh airtime utilities", http.MethodGet),
				middleware.SystemLogs("Refresh Airtime Utilities", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/airtime/utilities",
			Handler: airtimeHandler.GetAvailableAirtime,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get airtime utilities", http.MethodGet),
				middleware.SystemLogs("Get Airtime Utilities", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime",
			Handler: airtimeHandler.UpdateAirtimeStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update airtime utility status", http.MethodPut),
				middleware.SystemLogs("Update Airtime Utility Status", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime/price",
			Handler: airtimeHandler.UpdateAirtimeUtilityPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update airtime price", http.MethodPut),
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
				middleware.Authz(authModule, enforcer, "get airtime transactions", http.MethodGet),
				middleware.SystemLogs("Get Airtime Transactions", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/airtime/amount",
			Handler: airtimeHandler.UpdateAirtimeAmount,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update airtime amount", http.MethodPut),
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
				middleware.Authz(authModule, enforcer, "get airtime utilities stats", http.MethodGet),
				middleware.SystemLogs("Get Airtime Utilities Stats", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, airtimeRoutes, log)
}
