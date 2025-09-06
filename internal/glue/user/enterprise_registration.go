package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"go.uber.org/zap"
)

// InitEnterpriseRegistration initializes enterprise registration routes
func InitEnterpriseRegistration(
	group *gin.RouterGroup,
	log zap.Logger,
	user UserHandler,
) {
	log.Info("🔧 Initializing enterprise registration routes...")
	enterpriseRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/enterprise/register",
			Handler: user.InitiateEnterpriseRegistration,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/enterprise/register/complete",
			Handler: user.CompleteEnterpriseRegistration,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/enterprise/register/status/:user_id",
			Handler: user.GetEnterpriseRegistrationStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/enterprise/register/resend",
			Handler: user.ResendEnterpriseVerificationEmail,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}

	log.Info("🔧 Registering enterprise registration routes...", zap.Int("route_count", len(enterpriseRoutes)))
	routing.RegisterRoute(group, enterpriseRoutes, log)
	log.Info("Enterprise registration routes initialized successfully")
}
