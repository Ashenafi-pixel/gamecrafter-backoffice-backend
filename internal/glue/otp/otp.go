package otp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module/otp"

	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	otpHandler handler.OTP,
	otpModule otp.OTPModule,
) {

	otpRoutes := []routing.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/otp/email-verification",
			Handler: otpHandler.CreateEmailVerification,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/otp/verify",
			Handler: otpHandler.VerifyOTP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/otp/resend",
			Handler: otpHandler.ResendOTP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/otp/:otp_id",
			Handler: otpHandler.GetOTPInfo,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/otp/:otp_id/invalidate",
			Handler: otpHandler.InvalidateOTP,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/otp/cleanup",
			Handler: otpHandler.CleanupExpiredOTPs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/otp/stats",
			Handler: otpHandler.GetOTPStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		},
	}

	routing.RegisterRoute(group, otpRoutes, log)
}
