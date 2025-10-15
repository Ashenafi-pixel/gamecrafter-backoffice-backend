package twofactor

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/twofactor"
	"go.uber.org/zap"
)

// Init initializes the 2FA routes
func Init(grp *gin.RouterGroup, log *zap.Logger, handler twofactor.TwoFactorHandler) {
	// Setup 2FA routes directly
	setup2FARoutes(grp, handler, log)

	log.Info("2FA routes initialized successfully")
}

// setup2FARoutes configures all 2FA routes
func setup2FARoutes(router *gin.RouterGroup, handler twofactor.TwoFactorHandler, log *zap.Logger) {
	// 2FA setup and management routes (require authentication)
	authGroup := router.Group("/api/admin/auth/2fa")
	authGroup.Use(middleware.RateLimiter(), middleware.Auth())
	{
		// Generate secret and QR code
		authGroup.POST("/generate-secret", handler.GenerateSecret)

		// Enable 2FA with token verification
		authGroup.POST("/enable", handler.VerifyAndEnable)

		// Get current 2FA status
		authGroup.GET("/status", handler.GetStatus)

		// Disable 2FA
		authGroup.POST("/disable", handler.Disable2FA)
	}

	// 2FA verification routes (used during login) - no auth required
	verifyGroup := router.Group("/api/admin/auth/2fa")
	verifyGroup.Use(middleware.RateLimiter())
	{
		// Verify 2FA token during login
		verifyGroup.POST("/verify", handler.VerifyToken)

		// Multi-method verification
		verifyGroup.POST("/verify-method", handler.VerifyWithMethod)

		// Get available methods (for login flow - requires user_id query param)
		verifyGroup.GET("/available-methods", handler.GetAvailableMethods)

		// Get enabled methods (for login flow - requires user_id query param)
		verifyGroup.GET("/enabled-methods", handler.GetEnabledMethods)

		// Generate OTPs for login (no auth required, but needs user_id)
		verifyGroup.POST("/generate-email-otp", handler.GenerateEmailOTPForLogin)
		verifyGroup.POST("/generate-sms-otp", handler.GenerateSMSOTPForLogin)
		verifyGroup.POST("/get-backup-codes", handler.GetBackupCodes)

		// Enable/disable methods (for settings - requires user_id in body)
		verifyGroup.POST("/methods/enable", handler.EnableMethod)
		verifyGroup.POST("/methods/disable", handler.DisableMethod)
	}

	// 2FA setup routes for login flow (no auth required) - separate group to avoid conflicts
	setupGroup := router.Group("/api/admin/auth/2fa/setup")
	setupGroup.Use(middleware.RateLimiter())
	{
		// 2FA setup endpoints for login flow (no auth required, needs user_id)
		setupGroup.POST("/generate-secret", handler.GenerateSecretForLogin)
		setupGroup.POST("/enable-totp", handler.EnableTOTPForLogin)
		setupGroup.POST("/enable-email-otp", handler.EnableEmailOTPForLogin)
		setupGroup.POST("/enable-sms-otp", handler.EnableSMSOTPForLogin)
	}

	// Multi-method management routes (require authentication)
	multiMethodGroup := router.Group("/api/admin/auth/2fa/methods")
	multiMethodGroup.Use(middleware.RateLimiter(), middleware.Auth())
	{
		// Generate OTPs for different methods
		multiMethodGroup.POST("/email-otp", handler.GenerateEmailOTP)
		multiMethodGroup.POST("/sms-otp", handler.GenerateSMSOTP)

		// Get available methods (requires authentication)
		multiMethodGroup.GET("/available-methods", handler.GetAvailableMethods)
	}
}
