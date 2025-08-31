package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

// RouteProtectionLevel defines the verification level required for a route
type RouteProtectionLevel int

const (
	// BasicAuth allows both verified and unverified users
	BasicAuth RouteProtectionLevel = iota

	// EmailVerification requires email verification
	EmailVerification

	// PhoneVerification requires phone verification
	PhoneVerification

	// PartialVerification requires at least one verification method
	PartialVerification

	// FullVerification requires complete account verification
	FullVerification

	// BettingAccess requires full verification for betting activities
	BettingAccess

	// FinancialAccess requires full verification for financial transactions
	FinancialAccess

	// KYCAccess requires KYC verification
	KYCAccess
)

// RouteProtectionConfig holds configuration for route protection
type RouteProtectionConfig struct {
	Level                RouteProtectionLevel
	CustomErrorMessage   string
	CustomErrorCode      string
	RateLimitEnabled     bool
	MaxAttempts          int
	WindowDuration       time.Duration
	AuditLogEnabled      bool
	CustomVerificationFn func(ctx context.Context, userID string) (bool, error)
}

// RouteProtector provides enterprise-grade route protection
type RouteProtector struct {
	userModule module.User
	logger     *zap.Logger
	config     *RouteProtectionConfig
}

// NewRouteProtector creates a new route protector instance
func NewRouteProtector(userModule module.User, logger *zap.Logger, config *RouteProtectionConfig) *RouteProtector {
	if config == nil {
		config = &RouteProtectionConfig{
			Level:            FullVerification,
			RateLimitEnabled: true,
			MaxAttempts:      5,
			WindowDuration:   15 * time.Minute,
			AuditLogEnabled:  true,
		}
	}

	return &RouteProtector{
		userModule: userModule,
		logger:     logger,
		config:     config,
	}
}

// Protect creates middleware based on the protection level
func (rp *RouteProtector) Protect(level RouteProtectionLevel) gin.HandlerFunc {
	config := &RouteProtectionConfig{
		Level:                level,
		RateLimitEnabled:     rp.config.RateLimitEnabled,
		MaxAttempts:          rp.config.MaxAttempts,
		WindowDuration:       rp.config.WindowDuration,
		AuditLogEnabled:      rp.config.AuditLogEnabled,
		CustomVerificationFn: rp.config.CustomVerificationFn,
	}

	return rp.createProtectionMiddleware(config)
}

// createProtectionMiddleware creates the actual protection middleware
func (rp *RouteProtector) createProtectionMiddleware(config *RouteProtectionConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Extract user information from context
		userID, exists := c.Get("user-id")
		if !exists {
			rp.logError(c, "User ID not found in context", nil)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "User authentication required",
			})
			c.Abort()
			return
		}

		// Rate limiting check
		if config.RateLimitEnabled {
			if !rp.checkRateLimit(c, userID.(string)) {
				rp.logError(c, "Rate limit exceeded for protected route access", nil)
				c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Too many access attempts. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		// Perform protection checks
		protectionResult := rp.performProtectionChecks(c, userID.(string), config)
		if !protectionResult.IsValid {
			rp.handleProtectionFailure(c, protectionResult, config)
			return
		}

		// Audit logging
		if config.AuditLogEnabled {
			rp.auditLog(c, userID.(string), "PROTECTION_SUCCESS", protectionResult)
		}

		// Set protection context for downstream handlers
		c.Set("protection-level", protectionResult.Level)
		c.Set("protection-details", protectionResult)
	}
}

// ProtectionResult holds the result of protection checks
type ProtectionResult struct {
	IsValid              bool
	Level                string
	EmailVerified        bool
	PhoneVerified        bool
	FullVerified         bool
	MissingVerifications []string
	CustomVerified       bool
	Error                error
}

// performProtectionChecks performs all required protection checks
func (rp *RouteProtector) performProtectionChecks(c *gin.Context, userID string, config *RouteProtectionConfig) *ProtectionResult {
	result := &ProtectionResult{
		IsValid: true,
		Level:   "BASIC",
	}

	// Get verification status from context (set by Auth middleware)
	isVerified, _ := c.Get("is-verified")
	emailVerified, _ := c.Get("email-verified")
	phoneVerified, _ := c.Get("phone-verified")

	result.EmailVerified = emailVerified.(bool)
	result.PhoneVerified = phoneVerified.(bool)
	result.FullVerified = isVerified.(bool)

	// Apply protection level requirements
	switch config.Level {
	case BasicAuth:
		// No additional verification required
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

	case FullVerification:
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

	case BettingAccess:
		if !result.FullVerified {
			result.IsValid = false
			if !result.EmailVerified {
				result.MissingVerifications = append(result.MissingVerifications, "email")
			}
			if !result.PhoneVerified {
				result.MissingVerifications = append(result.MissingVerifications, "phone")
			}
		}
		result.Level = "BETTING_ACCESS"

	case FinancialAccess:
		if !result.FullVerified {
			result.IsValid = false
			if !result.EmailVerified {
				result.MissingVerifications = append(result.MissingVerifications, "email")
			}
			if !result.PhoneVerified {
				result.MissingVerifications = append(result.MissingVerifications, "phone")
			}
		}
		result.Level = "FINANCIAL_ACCESS"

	case KYCAccess:
		if !result.FullVerified {
			result.IsValid = false
			if !result.EmailVerified {
				result.MissingVerifications = append(result.MissingVerifications, "email")
			}
			if !result.PhoneVerified {
				result.MissingVerifications = append(result.MissingVerifications, "phone")
			}
		}
		result.Level = "KYC_ACCESS"
	}

	// Custom verification logic
	if config.CustomVerificationFn != nil {
		customVerified, err := config.CustomVerificationFn(c.Request.Context(), userID)
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

// handleProtectionFailure handles protection failures with appropriate responses
func (rp *RouteProtector) handleProtectionFailure(c *gin.Context, result *ProtectionResult, config *RouteProtectionConfig) {
	// Use custom error message if provided
	if config.CustomErrorMessage != "" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Code:    http.StatusForbidden,
			Message: config.CustomErrorMessage,
		})
		c.Abort()
		return
	}

	// Generate appropriate error response based on protection level
	message := rp.generateProtectionErrorMessage(result, config)

	// Log the protection failure
	rp.logError(c, "Route protection failed", map[string]interface{}{
		"user_id":               c.GetString("user-id"),
		"missing_verifications": result.MissingVerifications,
		"protection_level":      result.Level,
		"required_level":        config.Level,
	})

	// Return error response
	c.JSON(http.StatusForbidden, dto.ErrorResponse{
		Code:    http.StatusForbidden,
		Message: message,
	})
	c.Abort()
}

// generateProtectionErrorMessage generates appropriate error messages
func (rp *RouteProtector) generateProtectionErrorMessage(result *ProtectionResult, config *RouteProtectionConfig) string {
	switch config.Level {
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
			return fmt.Sprintf("Verification required: %s", strings.Join(result.MissingVerifications, ", "))
		}
		return "Access denied - insufficient verification level"
	}
}

// checkRateLimit implements rate limiting for protected route access
func (rp *RouteProtector) checkRateLimit(c *gin.Context, userID string) bool {
	// This is a simplified rate limiting implementation
	// In production, you would use Redis or a similar distributed rate limiter
	_ = fmt.Sprintf("route_access_attempts:%s", userID)

	// For now, we'll implement a basic in-memory rate limiter
	// In production, replace this with Redis-based rate limiting
	return true // Placeholder - implement actual rate limiting
}

// auditLog logs protection attempts for audit purposes
func (rp *RouteProtector) auditLog(c *gin.Context, userID, action string, result *ProtectionResult) {
	rp.logger.Info("Route protection audit log",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("ip_address", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("protection_level", result.Level),
		zap.Bool("email_verified", result.EmailVerified),
		zap.Bool("phone_verified", result.PhoneVerified),
		zap.Bool("full_verified", result.FullVerified),
		zap.Time("timestamp", time.Now().UTC()),
		zap.String("request_id", c.GetHeader("X-Request-ID")),
	)
}

// logError logs protection errors with context
func (rp *RouteProtector) logError(c *gin.Context, message string, fields map[string]interface{}) {
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

	rp.logger.Error(message, zapFields...)
}
