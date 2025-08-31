package agent

import (
	"net/http"

	"github.com/casbin/casbin/v2"
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
	agentHandler handler.Agent,
	authModule module.Authz,
	logsModule module.SystemLogs,
	enforcer *casbin.Enforcer,
) {
	agentRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/agent/links",
			Handler: agentHandler.CreateAgentReferralLink,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/agent/referrals",
			Handler: agentHandler.GetAgentReferrals,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "Get Agent Referrals", http.MethodGet),
				middleware.SystemLogs("Get Agent Referrals", &log, logsModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/agent/stats",
			Handler: agentHandler.GetReferralStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "Get Agent Referral Stats", http.MethodGet),
				middleware.SystemLogs("Get Agent Referral Stats", &log, logsModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/agent/providers",
			Handler: agentHandler.CreateAgentProvider,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "Create Agent Provider", http.MethodPost),
				middleware.SystemLogs("Create Agent Provider", &log, logsModule),
			},
		},
	}

	routing.RegisterRoute(group, agentRoutes, log)
}
