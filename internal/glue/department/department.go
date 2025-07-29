package department

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
	dep handler.Departements,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {

	departmentRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/departments",
			Handler: dep.CreateDepartement,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "create departments", http.MethodPost),
				middleware.SystemLogs("Create Departments", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/departments",
			Handler: dep.GetDepartement,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get departments", http.MethodGet),
				middleware.SystemLogs("Get Department", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/departments",
			Handler: dep.UpdateDepartment,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update department", http.MethodPatch),
				middleware.SystemLogs("Update Department", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/departments/assign",
			Handler: dep.AssignUserToDepartment,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "assign userto depertment", http.MethodPost),
				middleware.SystemLogs("Assign User To Department", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, departmentRoutes, log)
}
