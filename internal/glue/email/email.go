package email

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	emailHandler "github.com/tucanbit/internal/handler/email"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module/email"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	emailService email.EmailService,
) {
	testEmailHandler := emailHandler.NewTestEmailHandler(emailService, &log)

	emailRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/test-email",
			Handler: testEmailHandler.SendTestEmail,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}

	routing.RegisterRoute(group, emailRoutes, log)
}
