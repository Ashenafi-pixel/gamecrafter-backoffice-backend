package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// TwoFactorMiddleware handles 2FA verification for protected routes
type TwoFactorMiddleware struct {
	twoFactorService TwoFactorService
	logger           *zap.Logger
	config           TwoFactorMiddlewareConfig
}

// TwoFactorService interface for 2FA operations
type TwoFactorService interface {
	GetStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorAuthStatusResponse, error)
	VerifyToken(ctx context.Context, userID uuid.UUID, token, backupCode, ip, userAgent string) (*dto.TwoFactorVerifyResponse, error)
	IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error)
}

// TwoFactorMiddlewareConfig configuration for 2FA middleware
type TwoFactorMiddlewareConfig struct {
	// Whether 2FA is required for all authenticated routes
	Require2FAForAll bool
	// Routes that require 2FA (if Require2FAForAll is false)
	RequiredRoutes []string
	// Routes that are exempt from 2FA
	ExemptRoutes []string
	// Whether to check 2FA status from database
	Check2FAStatus bool
	// Rate limiting configuration
	RateLimitEnabled bool
	MaxAttempts      int
	WindowMinutes    int
}

// NewTwoFactorMiddleware creates a new 2FA middleware instance
func NewTwoFactorMiddleware(twoFactorService TwoFactorService, logger *zap.Logger, config TwoFactorMiddlewareConfig) *TwoFactorMiddleware {
	return &TwoFactorMiddleware{
		twoFactorService: twoFactorService,
		logger:           logger,
		config:           config,
	}
}

// Require2FA creates a middleware that requires 2FA verification
func (tfm *TwoFactorMiddleware) Require2FA() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Extract user information from context
		userIDStr, exists := c.Get("user-id")
		if !exists {
			tfm.logger.Error("User ID not found in context")
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "User authentication required",
			})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			tfm.logger.Error("Invalid user ID format", zap.Error(err))
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid user ID",
			})
			c.Abort()
			return
		}

		// Check if route requires 2FA
		if !tfm.shouldRequire2FA(c) {
			c.Next()
			return
		}

		// Check if user has 2FA enabled
		if tfm.config.Check2FAStatus {
			status, err := tfm.twoFactorService.GetStatus(c.Request.Context(), userID)
			if err != nil {
				tfm.logger.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Failed to check 2FA status",
				})
				c.Abort()
				return
			}

			// If user doesn't have 2FA enabled, allow access (or redirect to setup)
			if !status.IsEnabled {
				tfm.logger.Info("User does not have 2FA enabled", zap.String("user_id", userID.String()))
				c.JSON(http.StatusForbidden, dto.ErrorResponse{
					Code:    http.StatusForbidden,
					Message: "Two-factor authentication is required but not enabled. Please enable 2FA to access this resource.",
				})
				c.Abort()
				return
			}
		}

		// Check rate limiting
		if tfm.config.RateLimitEnabled {
			rateLimited, err := tfm.twoFactorService.IsRateLimited(c.Request.Context(), userID)
			if err != nil {
				tfm.logger.Error("Failed to check rate limit", zap.Error(err), zap.String("user_id", userID.String()))
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Failed to check rate limit",
				})
				c.Abort()
				return
			}
			if rateLimited {
				tfm.logger.Warn("User is rate limited for 2FA", zap.String("user_id", userID.String()))
				c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Too many 2FA attempts. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		// Check for 2FA token in request
		twoFactorToken := c.GetHeader("X-2FA-Token")
		backupCode := c.GetHeader("X-2FA-Backup-Code")

		if twoFactorToken == "" && backupCode == "" {
			tfm.logger.Warn("2FA token not provided", zap.String("user_id", userID.String()))
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Two-factor authentication token required",
			})
			c.Abort()
			return
		}

		// Verify 2FA token
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		verifyResp, err := tfm.twoFactorService.VerifyToken(c.Request.Context(), userID, twoFactorToken, backupCode, ip, userAgent)
		if err != nil {
			tfm.logger.Error("Failed to verify 2FA token", zap.Error(err), zap.String("user_id", userID.String()))
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to verify 2FA token",
			})
			c.Abort()
			return
		}

		if verifyResp == nil {
			tfm.logger.Warn("2FA token verification failed", zap.String("user_id", userID.String()))
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid 2FA token or backup code",
			})
			c.Abort()
			return
		}

		// Set 2FA verification context
		c.Set("2fa-verified", true)
		c.Set("2fa-method", tfm.get2FAMethod(twoFactorToken, backupCode))

		tfm.logger.Info("2FA verification successful", zap.String("user_id", userID.String()))
		c.Next()
	}
}

// Optional2FA creates a middleware that checks 2FA if enabled but doesn't require it
func (tfm *TwoFactorMiddleware) Optional2FA() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Extract user information from context
		userIDStr, exists := c.Get("user-id")
		if !exists {
			c.Next()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.Next()
			return
		}

		// Check if user has 2FA enabled
		if tfm.config.Check2FAStatus {
			status, err := tfm.twoFactorService.GetStatus(c.Request.Context(), userID)
			if err != nil {
				tfm.logger.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
				c.Next()
				return
			}

			// If user has 2FA enabled, require verification
			if status.IsEnabled {
				// Check for 2FA token
				twoFactorToken := c.GetHeader("X-2FA-Token")
				backupCode := c.GetHeader("X-2FA-Backup-Code")

				if twoFactorToken != "" || backupCode != "" {
					// Verify 2FA token
					ip := c.ClientIP()
					userAgent := c.GetHeader("User-Agent")

					verifyResp, err := tfm.twoFactorService.VerifyToken(c.Request.Context(), userID, twoFactorToken, backupCode, ip, userAgent)
					if err == nil && verifyResp != nil {
						c.Set("2fa-verified", true)
						c.Set("2fa-method", tfm.get2FAMethod(twoFactorToken, backupCode))
					}
				}
			}
		}

		c.Next()
	}
}

// shouldRequire2FA determines if the current route requires 2FA
func (tfm *TwoFactorMiddleware) shouldRequire2FA(c *gin.Context) bool {
	path := c.Request.URL.Path
	// Get the HTTP method for route matching
	_ = c.Request.Method

	// Check if route is exempt
	for _, exemptRoute := range tfm.config.ExemptRoutes {
		if strings.HasPrefix(path, exemptRoute) {
			return false
		}
	}

	// If 2FA is required for all routes
	if tfm.config.Require2FAForAll {
		return true
	}

	// Check if route is in required routes list
	for _, requiredRoute := range tfm.config.RequiredRoutes {
		if strings.HasPrefix(path, requiredRoute) {
			return true
		}
	}

	return false
}

// get2FAMethod determines which 2FA method was used
func (tfm *TwoFactorMiddleware) get2FAMethod(twoFactorToken, backupCode string) string {
	if twoFactorToken != "" {
		return "totp"
	}
	if backupCode != "" {
		return "backup_code"
	}
	return "unknown"
}

// TwoFactorSetupRequired creates a middleware that redirects users to 2FA setup if not enabled
func (tfm *TwoFactorMiddleware) TwoFactorSetupRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Extract user information from context
		userIDStr, exists := c.Get("user-id")
		if !exists {
			c.Next()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.Next()
			return
		}

		// Check if user has 2FA enabled
		if tfm.config.Check2FAStatus {
			status, err := tfm.twoFactorService.GetStatus(c.Request.Context(), userID)
			if err != nil {
				tfm.logger.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
				c.Next()
				return
			}

			// If user doesn't have 2FA enabled, redirect to setup
			if !status.IsEnabled {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{
					Code:    http.StatusForbidden,
					Message: "Two-factor authentication setup required",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
