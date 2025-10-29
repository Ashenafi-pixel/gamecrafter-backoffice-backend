package company

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
	company handler.Company,
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	systemLogs module.SystemLogs,
) {
	companyRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/companies",
			Handler: company.CreateCompany,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "create company", http.MethodPost),
				middleware.SystemLogs("Create Company", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/companies",
			Handler: company.GetCompanies,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get companies", http.MethodGet),
				middleware.SystemLogs("Get All Companies", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/companies/:id",
			Handler: company.GetCompanyByID,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "get company", http.MethodGet),
				middleware.SystemLogs("Get Company", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/companies/:id",
			Handler: company.UpdateCompany,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "update company", http.MethodPatch),
				middleware.SystemLogs("Update Company", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/companies/:id",
			Handler: company.DeleteCompany,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "delete company", http.MethodDelete),
				middleware.SystemLogs("Delete Company", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/companies/:id/add-ip",
			Handler: company.AddIP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "add ip to company", http.MethodPatch),
				middleware.SystemLogs("Add IP To Company", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, companyRoutes, log)
}
