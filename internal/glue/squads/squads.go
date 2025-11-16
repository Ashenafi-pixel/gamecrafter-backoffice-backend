package squads

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
	squadHandler handler.Squads,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {

	squads := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/squads",
			Handler: squadHandler.CreateSquad,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads",
			Handler: squadHandler.GetMySquads,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/own",
			Handler: squadHandler.GetMyOwnSquads,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/type",
			Handler: squadHandler.GetSquadsByType,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/name",
			Handler: squadHandler.GetSquadByName,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/squads/own",
			Handler: squadHandler.UpdateSquadHandle,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/squads/delete",
			Handler: squadHandler.DeleteSquad,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/squads/members",
			Handler: squadHandler.CreateSquadMember,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/members/bysquadid",
			Handler: squadHandler.GetSquadMembersBySquadID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/squads/members/:member_id",
			Handler: squadHandler.DeleteSquadMember,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/squads/members/all",
			Handler: squadHandler.DeleteSquadMembersBySquadID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/earns",
			Handler: squadHandler.GetSquadEarns,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/my/squads/earns",
			Handler: squadHandler.GetMySquadEarns,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/total/earns",
			Handler: squadHandler.GetSquadTotalEarns,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/members/earnings",
			Handler: squadHandler.GetSquadMembersEarnings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/tournament/ranking",
			Handler: squadHandler.GetTornamentStyleRanking,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admins/tournaments",
			Handler: squadHandler.CreateTournaments,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Create tournament", http.MethodPost),
				middleware.SystemLogs("Create Tournament", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/tournaments",
			Handler: squadHandler.GetTornamentStyles,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/squads/:squad_id/members/leave",
			Handler: squadHandler.LeaveSquad,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/squads/:squad_id/members/join",
			Handler: squadHandler.JoinSquad,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/squads/waitlist ",
			Handler: squadHandler.GetSquadWaitlist,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/squads/waitlist/remove",
			Handler: squadHandler.RemoveWaitingSquadWaitingUser,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/squads/waitlist/approve",
			Handler: squadHandler.ApproveWaitingSquadMember,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, squads, log)
}
