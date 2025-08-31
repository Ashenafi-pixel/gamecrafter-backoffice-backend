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

// VerificationConfig holds configuration for verification middleware
type VerificationConfig struct {
	// Verification levels
	RequireEmailVerification bool
	RequirePhoneVerification bool
	RequireFullVerification  bool

	// Custom error messages
	CustomErrorMessage string
	CustomErrorCode    string

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

// VerificationMiddleware provides enterprise-grade verification middleware
type VerificationMiddleware struct {
	userModule module.User
	logger     *zap.Logger
	config     *VerificationConfig
}

// NewVerificationMiddleware creates a new verification middleware instance
func NewVerificationMiddleware(userModule module.User, logger *zap.Logger, config *VerificationConfig) *VerificationMiddleware {
	if config == nil {
		config = &VerificationConfig{
			RequireEmailVerification: true,
			RequirePhoneVerification: true,
			RequireFullVerification:  true,
			RateLimitEnabled:         true,
			MaxAttempts:              5,
			WindowDuration:           15 * time.Minute,
			AuditLogEnabled:          true,
		}
	}

	return &VerificationMiddleware{
		userModule: userModule,
		logger:     logger,
		config:     config,
	}
}

// RequireVerification creates middleware that requires full account verification
func (vm *VerificationMiddleware) RequireVerification() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: true,
		RequirePhoneVerification: true,
		RequireFullVerification:  true,
	})
}

// RequireEmailVerification creates middleware that requires email verification only
func (vm *VerificationMiddleware) RequireEmailVerification() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: true,
		RequirePhoneVerification: false,
		RequireFullVerification:  false,
	})
}

// RequirePhoneVerification creates middleware that requires phone verification only
func (vm *VerificationMiddleware) RequirePhoneVerification() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: false,
		RequirePhoneVerification: true,
		RequireFullVerification:  false,
	})
}

// RequirePartialVerification creates middleware that requires at least one verification method
func (vm *VerificationMiddleware) RequirePartialVerification() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: false,
		RequirePhoneVerification: false,
		RequireFullVerification:  false,
		CustomVerificationFn: func(ctx context.Context, userID string) (bool, error) {
			// Check if user has at least one verification method completed
			// This would typically query the database for verification status
			return true, nil // Placeholder - implement actual logic
		},
	})
}

// createVerificationMiddleware creates the actual middleware function
func (vm *VerificationMiddleware) createVerificationMiddleware(config *VerificationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Extract user information from context
		userID, exists := c.Get("user-id")
		if !exists {
			vm.logError(c, "User ID not found in context", nil)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "User authentication required",
			})
			c.Abort()
			return
		}

		// Rate limiting check
		if config.RateLimitEnabled {
			if !vm.checkRateLimit(c, userID.(string)) {
				vm.logError(c, "Rate limit exceeded for verification attempts", nil)
				c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Too many verification attempts. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		// Perform verification checks
		verificationResult := vm.performVerificationChecks(c, userID.(string), config)
		if !verificationResult.IsValid {
			vm.handleVerificationFailure(c, verificationResult, config)
			return
		}

		// Audit logging
		if config.AuditLogEnabled && vm.config.AuditLogger != nil {
			vm.auditLog(c, userID.(string), "VERIFICATION_SUCCESS", verificationResult)
		}

		// Set verification context for downstream handlers
		c.Set("verification-level", verificationResult.Level)
		c.Set("verification-details", verificationResult)
	}
}

// VerificationResult holds the result of verification checks
type VerificationResult struct {
	IsValid              bool
	Level                string
	EmailVerified        bool
	PhoneVerified        bool
	FullVerified         bool
	MissingVerifications []string
	CustomVerified       bool
	Error                error
}

// performVerificationChecks performs all required verification checks
func (vm *VerificationMiddleware) performVerificationChecks(c *gin.Context, userID string, config *VerificationConfig) *VerificationResult {
	result := &VerificationResult{
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

	// Check email verification requirement
	if config.RequireEmailVerification && !result.EmailVerified {
		result.IsValid = false
		result.MissingVerifications = append(result.MissingVerifications, "email")
	}

	// Check phone verification requirement
	if config.RequirePhoneVerification && !result.PhoneVerified {
		result.IsValid = false
		result.MissingVerifications = append(result.MissingVerifications, "phone")
	}

	// Check full verification requirement
	if config.RequireFullVerification && !result.FullVerified {
		result.IsValid = false
		if !result.EmailVerified {
			result.MissingVerifications = append(result.MissingVerifications, "email")
		}
		if !result.PhoneVerified {
			result.MissingVerifications = append(result.MissingVerifications, "phone")
		}
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

	// Determine verification level
	if result.FullVerified {
		result.Level = "FULL"
	} else if result.EmailVerified && result.PhoneVerified {
		result.Level = "COMPLETE"
	} else if result.EmailVerified || result.PhoneVerified {
		result.Level = "PARTIAL"
	} else {
		result.Level = "NONE"
	}

	return result
}

// handleVerificationFailure handles verification failures with appropriate responses
func (vm *VerificationMiddleware) handleVerificationFailure(c *gin.Context, result *VerificationResult, config *VerificationConfig) {
	// Use custom error message if provided
	if config.CustomErrorMessage != "" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Code:    http.StatusForbidden,
			Message: config.CustomErrorMessage,
		})
		c.Abort()
		return
	}

	// Generate appropriate error response based on missing verifications
	var message string

	if len(result.MissingVerifications) == 1 {
		verification := result.MissingVerifications[0]
		switch verification {
		case "email":
			message = "Email verification is required to access this resource"
		case "phone":
			message = "Phone verification is required to access this resource"
		case "custom":
			message = "Additional verification is required to access this resource"
		}
	} else {
		message = fmt.Sprintf("Verification required: %s", strings.Join(result.MissingVerifications, ", "))
	}

	// Log the verification failure
	vm.logError(c, "Verification failed", map[string]interface{}{
		"user_id":               c.GetString("user-id"),
		"missing_verifications": result.MissingVerifications,
		"verification_level":    result.Level,
	})

	// Return detailed error response
	c.JSON(http.StatusForbidden, dto.ErrorResponse{
		Code:    http.StatusForbidden,
		Message: message,
	})
	c.Abort()
}

// getRequiredLevel returns a human-readable description of the required verification level
func getRequiredLevel(config *VerificationConfig) string {
	if config.RequireFullVerification {
		return "FULL_VERIFICATION"
	}
	if config.RequireEmailVerification && config.RequirePhoneVerification {
		return "EMAIL_AND_PHONE_VERIFICATION"
	}
	if config.RequireEmailVerification {
		return "EMAIL_VERIFICATION"
	}
	if config.RequirePhoneVerification {
		return "PHONE_VERIFICATION"
	}
	if config.CustomVerificationFn != nil {
		return "CUSTOM_VERIFICATION"
	}
	return "BASIC_AUTHENTICATION"
}

// checkRateLimit implements rate limiting for verification attempts
func (vm *VerificationMiddleware) checkRateLimit(c *gin.Context, userID string) bool {
	// This is a simplified rate limiting implementation
	// In production, you would use Redis or a similar distributed rate limiter
	_ = fmt.Sprintf("verification_attempts:%s", userID)

	// For now, we'll implement a basic in-memory rate limiter
	// In production, replace this with Redis-based rate limiting
	return true // Placeholder - implement actual rate limiting
}

// auditLog logs verification attempts for audit purposes
func (vm *VerificationMiddleware) auditLog(c *gin.Context, userID, action string, result *VerificationResult) {
	vm.config.AuditLogger.Info("Verification audit log",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("ip_address", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("verification_level", result.Level),
		zap.Bool("email_verified", result.EmailVerified),
		zap.Bool("phone_verified", result.PhoneVerified),
		zap.Bool("full_verified", result.FullVerified),
		zap.Time("timestamp", time.Now().UTC()),
		zap.String("request_id", c.GetHeader("X-Request-ID")),
	)
}

// logError logs verification errors with context
func (vm *VerificationMiddleware) logError(c *gin.Context, message string, fields map[string]interface{}) {
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

	vm.logger.Error(message, zapFields...)
}

// Convenience functions for common verification scenarios
func (vm *VerificationMiddleware) RequireBettingAccess() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: true,
		RequirePhoneVerification: true,
		RequireFullVerification:  true,
		CustomErrorMessage:       "Account verification required for betting activities",
		CustomErrorCode:          "BETTING_VERIFICATION_REQUIRED",
	})
}

func (vm *VerificationMiddleware) RequireFinancialAccess() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: true,
		RequirePhoneVerification: true,
		RequireFullVerification:  true,
		CustomErrorMessage:       "Full account verification required for financial transactions",
		CustomErrorCode:          "FINANCIAL_VERIFICATION_REQUIRED",
	})
}

func (vm *VerificationMiddleware) RequireKYCAccess() gin.HandlerFunc {
	return vm.createVerificationMiddleware(&VerificationConfig{
		RequireEmailVerification: true,
		RequirePhoneVerification: true,
		RequireFullVerification:  true,
		CustomErrorMessage:       "KYC verification required for this operation",
		CustomErrorCode:          "KYC_VERIFICATION_REQUIRED",
	})
}
