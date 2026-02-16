package report

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
	report handler.Report,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	reportRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/report/daily",
			Handler: report.GetDailyReport,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view daily report", http.MethodGet),
				middleware.SystemLogs("Get daily Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/duplicate-ip-accounts",
			Handler: report.GetDuplicateIPAccounts,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list (risk-focused report)
				middleware.Authz(authModule, "view kyc risk", http.MethodGet),
				middleware.SystemLogs("Get Duplicate IP Accounts Report", &log, systemLogs),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/report/duplicate-ip-accounts/suspend",
			Handler: report.SuspendAccountsByIP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "suspend player", http.MethodPost),
				middleware.SystemLogs("Suspend Accounts By IP", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/big-winners",
			Handler: report.GetBigWinners,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view big winners report", http.MethodGet),
				middleware.SystemLogs("Get Big Winners Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/big-winners/export",
			Handler: report.ExportBigWinners,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Big Winners Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/player-metrics",
			Handler: report.GetPlayerMetrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view player metrics report", http.MethodGet),
				middleware.SystemLogs("Get Player Metrics Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/player-metrics/:player_id/transactions",
			Handler: report.GetPlayerTransactions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view player metrics report", http.MethodGet),
				middleware.SystemLogs("Get Player Transactions", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/player-metrics/export",
			Handler: report.ExportPlayerMetrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Player Metrics Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/player-metrics/:player_id/transactions/export",
			Handler: report.ExportPlayerTransactions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Player Transactions", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/country",
			Handler: report.GetCountryMetrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view country report", http.MethodGet),
				middleware.SystemLogs("Get Country Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/country/:country/players",
			Handler: report.GetCountryPlayers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view country report", http.MethodGet),
				middleware.SystemLogs("Get Country Players", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/country/export",
			Handler: report.ExportCountryMetrics,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Country Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/country/:country/players/export",
			Handler: report.ExportCountryPlayers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Country Players", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/game-performance",
			Handler: report.GetGamePerformance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view game performance report", http.MethodGet),
				middleware.SystemLogs("Get Game Performance Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/game-performance/:game_id/players",
			Handler: report.GetGamePlayers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view game performance report", http.MethodGet),
				middleware.SystemLogs("Get Game Players", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/game-performance/export",
			Handler: report.ExportGamePerformance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Game Performance Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/game-performance/:game_id/players/export",
			Handler: report.ExportGamePlayers,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Game Players", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/provider-performance",
			Handler: report.GetProviderPerformance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view provider performance report", http.MethodGet),
				middleware.SystemLogs("Get Provider Performance Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/provider-performance/export",
			Handler: report.ExportProviderPerformance,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "export reports", http.MethodGet),
				middleware.SystemLogs("Export Provider Performance Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/affiliate",
			Handler: report.GetAffiliateReport,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view affiliate report", http.MethodGet),
				middleware.SystemLogs("Get Affiliate Report", &log, systemLogs),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/report/affiliate/players",
			Handler: report.GetAffiliatePlayersReport,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view affiliate report", http.MethodGet),
				middleware.SystemLogs("Get Affiliate Players Report", &log, systemLogs),
			},
		},
	}

	routing.RegisterRoute(group, reportRoutes, log)
}
