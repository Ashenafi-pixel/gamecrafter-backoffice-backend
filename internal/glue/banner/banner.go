package banner

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
	authModule module.Authz,
	enforcer *casbin.Enforcer,
	banner handler.Banner,
	systemLogs module.SystemLogs,
) {
	bannerRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/banners",
			Handler: banner.GetAllBanners,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner read", http.MethodGet),
				middleware.SystemLogs("Get All Banners", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/banners",
			Handler: banner.CreateBanner,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner create", http.MethodPost),
				middleware.SystemLogs("Create Banner", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/banners/display",
			Handler: banner.GetBannerByPage,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner display", http.MethodGet),
				middleware.SystemLogs("Get banner by page", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/banners/upload",
			Handler: banner.UploadBannerImage,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner image upload", http.MethodPost),
				middleware.SystemLogs("Upload Banner Image", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/api/admin/banners/:id",
			Handler: banner.UpdateBanner,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner update", http.MethodPatch),
				middleware.SystemLogs("Update Banner", &log, systemLogs),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/banners/:id",
			Handler: banner.DeleteBanner,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, enforcer, "banner delete", http.MethodDelete),
				middleware.SystemLogs("Delete Banner", &log, systemLogs),
			},
		},
	}
	routing.RegisterRoute(group, bannerRoutes, log)
}
