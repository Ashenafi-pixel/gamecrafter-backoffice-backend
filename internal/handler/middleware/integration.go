package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

// MiddlewareManager provides centralized management of all middleware components
type MiddlewareManager struct {
	verificationMiddleware *VerificationMiddleware
	routeProtector         *RouteProtector
	userModule             module.User
	logger                 *zap.Logger
	config                 *MiddlewareConfig
}

// MiddlewareConfig holds configuration for the middleware manager
type MiddlewareConfig struct {
	// Verification settings
	VerificationEnabled      bool
	RequireEmailVerification bool
	RequirePhoneVerification bool
	RequireFullVerification  bool

	// Route protection settings
	RouteProtectionEnabled bool
	DefaultProtectionLevel RouteProtectionLevel

	// Rate limiting
	RateLimitEnabled bool
	MaxAttempts      int
	WindowDuration   time.Duration

	// Audit logging
	AuditLogEnabled bool
	AuditLogger     *zap.Logger

	// Custom verification logic
	CustomVerificationFn func(ctx context.Context, userID string) (bool, error)
}

// NewMiddlewareManager creates a new middleware manager instance
func NewMiddlewareManager(userModule module.User, logger *zap.Logger, config *MiddlewareConfig) *MiddlewareManager {
	if config == nil {
		config = &MiddlewareConfig{
			VerificationEnabled:      true,
			RequireEmailVerification: true,
			RequirePhoneVerification: true,
			RequireFullVerification:  true,
			RouteProtectionEnabled:   true,
			DefaultProtectionLevel:   FullVerification,
			RateLimitEnabled:         true,
			MaxAttempts:              5,
			WindowDuration:           15 * time.Minute,
			AuditLogEnabled:          true,
		}
	}

	// Initialize verification middleware
	verificationConfig := &VerificationConfig{
		RequireEmailVerification: config.RequireEmailVerification,
		RequirePhoneVerification: config.RequirePhoneVerification,
		RequireFullVerification:  config.RequireFullVerification,
		RateLimitEnabled:         config.RateLimitEnabled,
		MaxAttempts:              config.MaxAttempts,
		WindowDuration:           config.WindowDuration,
		AuditLogEnabled:          config.AuditLogEnabled,
		AuditLogger:              config.AuditLogger,
		CustomVerificationFn:     config.CustomVerificationFn,
	}

	verificationMiddleware := NewVerificationMiddleware(userModule, logger, verificationConfig)

	// Initialize route protector
	routeProtectionConfig := &RouteProtectionConfig{
		Level:                config.DefaultProtectionLevel,
		RateLimitEnabled:     config.RateLimitEnabled,
		MaxAttempts:          config.MaxAttempts,
		WindowDuration:       config.WindowDuration,
		AuditLogEnabled:      config.AuditLogEnabled,
		CustomVerificationFn: config.CustomVerificationFn,
	}

	routeProtector := NewRouteProtector(userModule, logger, routeProtectionConfig)

	return &MiddlewareManager{
		verificationMiddleware: verificationMiddleware,
		routeProtector:         routeProtector,
		userModule:             userModule,
		logger:                 logger,
		config:                 config,
	}
}

// GetVerificationMiddleware returns the verification middleware instance
func (mm *MiddlewareManager) GetVerificationMiddleware() *VerificationMiddleware {
	return mm.verificationMiddleware
}

// GetRouteProtector returns the route protector instance
func (mm *MiddlewareManager) GetRouteProtector() *RouteProtector {
	return mm.routeProtector
}

// CreateComprehensiveMiddleware creates middleware that combines authentication, verification, and route protection
func (mm *MiddlewareManager) CreateComprehensiveMiddleware(protectionLevel RouteProtectionLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Basic authentication
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Step 2: Rate limiting check
		if mm.config.RateLimitEnabled {
			userID, _ := c.Get("user-id")
			if !mm.checkComprehensiveRateLimit(c, userID.(string)) {
				mm.logError(c, "Rate limit exceeded for comprehensive middleware", nil)
				c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Too many access attempts. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		// Step 3: Verification checks
		if mm.config.VerificationEnabled {
			verificationResult := mm.performComprehensiveVerification(c, protectionLevel)
			if !verificationResult.IsValid {
				mm.handleComprehensiveVerificationFailure(c, verificationResult, protectionLevel)
				return
			}
		}

		// Step 4: Route protection checks
		if mm.config.RouteProtectionEnabled {
			protectionResult := mm.performComprehensiveProtection(c, protectionLevel)
			if !protectionResult.IsValid {
				mm.handleComprehensiveProtectionFailure(c, protectionResult, protectionLevel)
				return
			}
		}

		// Step 5: Audit logging
		if mm.config.AuditLogEnabled && mm.config.AuditLogger != nil {
			mm.auditComprehensiveAccess(c, protectionLevel)
		}

		// Step 6: Set comprehensive context
		mm.setComprehensiveContext(c, protectionLevel)
	}
}

// ComprehensiveVerificationResult holds the result of comprehensive verification checks
type ComprehensiveVerificationResult struct {
	IsValid              bool
	Level                string
	EmailVerified        bool
	PhoneVerified        bool
	FullVerified         bool
	MissingVerifications []string
	CustomVerified       bool
	Error                error
}

// performComprehensiveVerification performs all verification checks
func (mm *MiddlewareManager) performComprehensiveVerification(c *gin.Context, protectionLevel RouteProtectionLevel) *ComprehensiveVerificationResult {
	result := &ComprehensiveVerificationResult{
		IsValid: true,
		Level:   "BASIC",
	}

	// Get verification status from context
	isVerified, _ := c.Get("is-verified")
	emailVerified, _ := c.Get("email-verified")
	phoneVerified, _ := c.Get("phone-verified")

	result.EmailVerified = emailVerified.(bool)
	result.PhoneVerified = phoneVerified.(bool)
	result.FullVerified = isVerified.(bool)

	// Apply verification requirements based on protection level
	switch protectionLevel {
	case BasicAuth:
		result.Level = "BASIC"

	case EmailVerification:
		if !result.EmailVerified {
			result.IsValid = false
			result.MissingVerifications = append(result.MissingVerifications, "email")
		}
		result.Level = "EMAIL_VERIFIED"

	case PhoneVerification:
		if !result.PhoneVerified {
			result.IsValid = false
			result.MissingVerifications = append(result.MissingVerifications, "phone")
		}
		result.Level = "PHONE_VERIFIED"

	case PartialVerification:
		if !result.EmailVerified && !result.PhoneVerified {
			result.IsValid = false
			result.MissingVerifications = append(result.MissingVerifications, "email", "phone")
		}
		result.Level = "PARTIAL_VERIFIED"

	case FullVerification, BettingAccess, FinancialAccess, KYCAccess:
		if !result.FullVerified {
			result.IsValid = false
			if !result.EmailVerified {
				result.MissingVerifications = append(result.MissingVerifications, "email")
			}
			if !result.PhoneVerified {
				result.MissingVerifications = append(result.MissingVerifications, "phone")
			}
		}
		result.Level = "FULL_VERIFIED"
	}

	// Custom verification logic
	if mm.config.CustomVerificationFn != nil {
		userID, _ := c.Get("user-id")
		customVerified, err := mm.config.CustomVerificationFn(c.Request.Context(), userID.(string))
		if err != nil {
			result.Error = err
			result.IsValid = false
			return result
		}
		result.CustomVerified = customVerified
		if !customVerified {
			result.IsValid = false
			result.MissingVerifications = append(result.MissingVerifications, "custom")
		}
	}

	return result
}

// handleComprehensiveVerificationFailure handles verification failures
func (mm *MiddlewareManager) handleComprehensiveVerificationFailure(c *gin.Context, result *ComprehensiveVerificationResult, protectionLevel RouteProtectionLevel) {
	message := mm.generateComprehensiveVerificationErrorMessage(result, protectionLevel)

	mm.logError(c, "Comprehensive verification failed", map[string]interface{}{
		"user_id":               c.GetString("user-id"),
		"missing_verifications": result.MissingVerifications,
		"verification_level":    result.Level,
		"required_level":        protectionLevel,
	})

	c.JSON(http.StatusForbidden, dto.ErrorResponse{
		Code:    http.StatusForbidden,
		Message: message,
	})
	c.Abort()
}

// generateComprehensiveVerificationErrorMessage generates verification error messages
func (mm *MiddlewareManager) generateComprehensiveVerificationErrorMessage(result *ComprehensiveVerificationResult, protectionLevel RouteProtectionLevel) string {
	switch protectionLevel {
	case EmailVerification:
		return "Email verification is required to access this resource"
	case PhoneVerification:
		return "Phone verification is required to access this resource"
	case PartialVerification:
		return "At least one verification method is required to access this resource"
	case FullVerification:
		return "Full account verification is required to access this resource"
	case BettingAccess:
		return "Account verification required for betting activities"
	case FinancialAccess:
		return "Full account verification required for financial transactions"
	case KYCAccess:
		return "KYC verification required for this operation"
	default:
		if len(result.MissingVerifications) > 0 {
			return fmt.Sprintf("Verification required: %s", result.MissingVerifications[0])
		}
		return "Access denied - insufficient verification level"
	}
}

// performComprehensiveProtection performs route protection checks
func (mm *MiddlewareManager) performComprehensiveProtection(c *gin.Context, protectionLevel RouteProtectionLevel) *ProtectionResult {
	// This delegates to the route protector's logic
	return mm.routeProtector.performProtectionChecks(c, c.GetString("user-id"), &RouteProtectionConfig{
		Level: protectionLevel,
	})
}

// handleComprehensiveProtectionFailure handles protection failures
func (mm *MiddlewareManager) handleComprehensiveProtectionFailure(c *gin.Context, result *ProtectionResult, protectionLevel RouteProtectionLevel) {
	mm.routeProtector.handleProtectionFailure(c, result, &RouteProtectionConfig{
		Level: protectionLevel,
	})
}

// checkComprehensiveRateLimit implements rate limiting for comprehensive middleware
func (mm *MiddlewareManager) checkComprehensiveRateLimit(c *gin.Context, userID string) bool {
	// This is a simplified rate limiting implementation
	// In production, you would use Redis or a similar distributed rate limiter
	_ = fmt.Sprintf("comprehensive_access_attempts:%s", userID)

	// For now, we'll implement a basic in-memory rate limiter
	// In production, replace this with Redis-based rate limiting
	return true // Placeholder - implement actual rate limiting
}

// auditComprehensiveAccess logs comprehensive access attempts
func (mm *MiddlewareManager) auditComprehensiveAccess(c *gin.Context, protectionLevel RouteProtectionLevel) {
	userID, _ := c.Get("user-id")
	isVerified, _ := c.Get("is-verified")
	emailVerified, _ := c.Get("email-verified")
	phoneVerified, _ := c.Get("phone-verified")

	mm.config.AuditLogger.Info("Comprehensive middleware audit log",
		zap.String("user_id", userID.(string)),
		zap.String("action", "COMPREHENSIVE_ACCESS"),
		zap.String("ip_address", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.Int("protection_level", int(protectionLevel)),
		zap.Bool("email_verified", emailVerified.(bool)),
		zap.Bool("phone_verified", phoneVerified.(bool)),
		zap.Bool("full_verified", isVerified.(bool)),
		zap.Time("timestamp", time.Now().UTC()),
		zap.String("request_id", c.GetHeader("X-Request-ID")),
	)
}

// setComprehensiveContext sets comprehensive context for downstream handlers
func (mm *MiddlewareManager) setComprehensiveContext(c *gin.Context, protectionLevel RouteProtectionLevel) {
	c.Set("comprehensive-protection-level", protectionLevel)
	c.Set("comprehensive-timestamp", time.Now().UTC())
	c.Set("comprehensive-request-id", c.GetHeader("X-Request-ID"))
}

// logError logs comprehensive middleware errors
func (mm *MiddlewareManager) logError(c *gin.Context, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["ip_address"] = c.ClientIP()
	fields["user_agent"] = c.GetHeader("User-Agent")
	fields["request_id"] = c.GetHeader("X-Request-ID")
	fields["timestamp"] = time.Now().UTC()

	// Convert fields to zap fields
	var zapFields []zap.Field
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			zapFields = append(zapFields, zap.String(k, val))
		case int:
			zapFields = append(zapFields, zap.Int(k, val))
		case bool:
			zapFields = append(zapFields, zap.Bool(k, val))
		case time.Time:
			zapFields = append(zapFields, zap.Time(k, val))
		default:
			zapFields = append(zapFields, zap.Any(k, val))
		}
	}

	mm.logger.Error(message, zapFields...)
}
