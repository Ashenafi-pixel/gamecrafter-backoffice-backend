package twofactor

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

// TwoFactorService interface for 2FA operations
type TwoFactorService interface {
	GenerateSecret(ctx context.Context, userID uuid.UUID, email string) (*dto.TwoFactorAuthSetupResponse, error)
	VerifyAndEnable2FA(ctx context.Context, userID uuid.UUID, secret, token string) error
	Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSettings, error)
	Disable2FA(ctx context.Context, userID uuid.UUID, token, ip, userAgent string) error
	ValidateBackupCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	VerifyLoginToken(ctx context.Context, userID uuid.UUID, token, backupCode, ip, userAgent string) (bool, error)
	IsRateLimited(ctx context.Context, userID uuid.UUID) (bool, error)

	// Multi-method support
	GetAvailableMethods(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetEnabledMethods(ctx context.Context, userID uuid.UUID) ([]string, error)
	EnableMethod(ctx context.Context, userID uuid.UUID, method string, data map[string]interface{}) error
	DisableMethod(ctx context.Context, userID uuid.UUID, method, verificationData string) error
	VerifyLoginWithMethod(ctx context.Context, userID uuid.UUID, method, token, ip, userAgent string) (bool, error)
	GenerateEmailOTP(ctx context.Context, userID uuid.UUID, email string) error
	GenerateSMSOTP(ctx context.Context, userID uuid.UUID, phoneNumber string) error
	GenerateBackupCodes() []string
	SaveBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error
	GetSecret(ctx context.Context, userID uuid.UUID) (string, error)
	VerifyTOTPToken(secret, token string) bool

	// Passkey methods
	RegisterPasskey(ctx context.Context, userID uuid.UUID, credentialData map[string]interface{}) error
	GetPasskeyAssertionOptions(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	VerifyPasskey(ctx context.Context, userID uuid.UUID, credentialData map[string]interface{}) (bool, error)
	ListPasskeys(ctx context.Context, userID uuid.UUID) ([]map[string]interface{}, error)
	DeletePasskey(ctx context.Context, userID uuid.UUID, credentialID string) error
}

type twoFactorHandler struct {
	service TwoFactorService
	log     *zap.Logger
}

type TwoFactorHandler interface {
	// Setup endpoints
	GenerateSecret(c *gin.Context)
	VerifyAndEnable(c *gin.Context)

	// Verification endpoints
	VerifyToken(c *gin.Context)

	// Management endpoints
	GetStatus(c *gin.Context)
	Disable2FA(c *gin.Context)

	// Multi-method endpoints
	GetAvailableMethods(c *gin.Context)
	GetEnabledMethods(c *gin.Context)
	EnableMethod(c *gin.Context)
	DisableMethod(c *gin.Context)
	VerifyWithMethod(c *gin.Context)
	GenerateEmailOTP(c *gin.Context)
	GenerateSMSOTP(c *gin.Context)

	// Login-specific endpoints (no auth required)
	GenerateEmailOTPForLogin(c *gin.Context)
	GenerateSMSOTPForLogin(c *gin.Context)
	GetBackupCodes(c *gin.Context)

	// 2FA setup endpoints for login flow (no auth required)
	GenerateSecretForLogin(c *gin.Context)
	EnableTOTPForLogin(c *gin.Context)
	EnableEmailOTPForLogin(c *gin.Context)
	EnableSMSOTPForLogin(c *gin.Context)

	// Passkey endpoints
	RegisterPasskey(c *gin.Context)
	GetPasskeyAssertionOptions(c *gin.Context)
	VerifyPasskey(c *gin.Context)
	ListPasskeys(c *gin.Context)
	DeletePasskey(c *gin.Context)
}

func NewTwoFactorHandler(service TwoFactorService, log *zap.Logger) TwoFactorHandler {
	return &twoFactorHandler{
		service: service,
		log:     log,
	}
}

// GenerateSecret generates a new 2FA secret and QR code
// @Summary Generate 2FA Secret
// @Description Generate a new TOTP secret and QR code for 2FA setup
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 400 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/generate-secret [post]
func (h *twoFactorHandler) GenerateSecret(c *gin.Context) {
	userIDStr, exists := c.Get("user-id")
	if !exists {
		h.log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	// Get user email from context or request
	email := c.GetString("user-email")
	if email == "" {
		// If email is not in context, get it from request body
		var req struct {
			Email string `json:"email" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			h.log.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Error:   "Email is required",
			})
			return
		}
		email = req.Email
	}

	// Generate secret
	secret, err := h.service.GenerateSecret(c.Request.Context(), userID, email)
	if err != nil {
		h.log.Error("Failed to generate 2FA secret", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to generate 2FA secret",
		})
		return
	}

	h.log.Info("2FA secret generated successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "2FA secret generated successfully",
		Data:    secret,
	})
}

// VerifyAndEnable verifies a token and enables 2FA
// @Summary Enable 2FA
// @Description Verify a TOTP token and enable 2FA for the user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param request body dto.TwoFactorSetupRequest true "2FA setup request"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 400 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 429 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/enable [post]
func (h *twoFactorHandler) VerifyAndEnable(c *gin.Context) {
	userIDStr, exists := c.Get("user-id")
	if !exists {
		h.log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var secretReq struct {
		Secret string `json:"secret" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&secretReq); err != nil {
		h.log.Error("Failed to bind secret request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Secret and token are required",
		})
		return
	}

	// Verify and enable 2FA
	err = h.service.VerifyAndEnable2FA(c.Request.Context(), userID, secretReq.Secret, secretReq.Token)
	if err != nil {
		h.log.Error("Failed to enable 2FA", zap.Error(err), zap.String("user_id", userID.String()))

		// Check if it's a rate limiting error
		if err.Error() == "too many attempts, please try again later" {
			c.JSON(http.StatusTooManyRequests, dto.TwoFactorResponse{
				Success: false,
				Error:   "Too many attempts, please try again later",
			})
			return
		}

		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("2FA enabled successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "2FA enabled successfully",
	})
}

// VerifyToken verifies a 2FA token during login
// @Summary Verify 2FA Token
// @Description Verify a TOTP token or backup code during login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.TwoFactorVerifyRequest true "2FA verification request"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 400 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 429 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/verify [post]
func (h *twoFactorHandler) VerifyToken(c *gin.Context) {
	var req struct {
		dto.TwoFactorVerifyRequest
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Get user ID from request body or header
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		userIDStr = req.UserID
	}

	if userIDStr == "" {
		h.log.Error("User ID not provided")
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "User ID is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	// Get client IP and user agent
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Determine verification method
	var success bool

	if req.Method != "" {
		// Use method-specific verification
		success, err = h.service.VerifyLoginWithMethod(c.Request.Context(), userID, req.Method, req.Token, ip, userAgent)
	} else {
		// Fallback to legacy verification (for backward compatibility)
		success, err = h.service.VerifyLoginToken(c.Request.Context(), userID, req.Token, req.BackupCode, ip, userAgent)
	}
	if err != nil {
		h.log.Error("Failed to verify 2FA token", zap.Error(err), zap.String("user_id", userID.String()))

		// Check if it's a rate limiting error
		if err.Error() == "too many attempts, please try again later" {
			c.JSON(http.StatusTooManyRequests, dto.TwoFactorResponse{
				Success: false,
				Error:   "Too many attempts, please try again later",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to verify token",
		})
		return
	}

	if !success {
		h.log.Warn("2FA token verification failed", zap.String("user_id", userID.String()))
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid token or backup code",
		})
		return
	}

	h.log.Info("2FA token verified successfully", zap.String("user_id", userID.String()))

	// Generate JWT token for successful 2FA verification
	token, err := utils.GenerateJWTWithVerification(userID, true, true, true)
	if err != nil {
		h.log.Error("Failed to generate JWT token after 2FA", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to complete login",
		})
		return
	}

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "2FA verification successful",
		Data: map[string]interface{}{
			"access_token": token,
		},
	})
}

// GetStatus retrieves the current 2FA status
// @Summary Get 2FA Status
// @Description Get the current 2FA status and settings for the user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/status [get]
func (h *twoFactorHandler) GetStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user-id")
	if !exists {
		h.log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	status, err := h.service.Get2FAStatus(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to get 2FA status",
		})
		return
	}

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Data:    status,
	})
}

// Disable2FA disables 2FA for the user
// @Summary Disable 2FA
// @Description Disable 2FA for the user with token verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param request body dto.TwoFactorDisableRequest true "2FA disable request"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 400 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 429 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/disable [post]
func (h *twoFactorHandler) Disable2FA(c *gin.Context) {
	userIDStr, exists := c.Get("user-id")
	if !exists {
		h.log.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req dto.TwoFactorDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Disable 2FA
	err = h.service.Disable2FA(c.Request.Context(), userID, req.Token, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.log.Error("Failed to disable 2FA", zap.Error(err), zap.String("user_id", userID.String()))

		// Check if it's a rate limiting error
		if err.Error() == "too many attempts, please try again later" {
			c.JSON(http.StatusTooManyRequests, dto.TwoFactorResponse{
				Success: false,
				Error:   "Too many attempts, please try again later",
			})
			return
		}

		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("2FA disabled successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "2FA disabled successfully",
	})
}

// GetEnabledMethods returns enabled 2FA methods for a user
func (h *twoFactorHandler) GetEnabledMethods(c *gin.Context) {
	// Try to get user ID from context (authenticated request)
	userID, exists := c.Get("user_id")
	if !exists {
		// If not authenticated, try to get user ID from query parameter (for login flow)
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "User ID is required",
			})
			return
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid user ID format",
			})
			return
		}

		methods, err := h.service.GetEnabledMethods(c.Request.Context(), userUUID)
		if err != nil {
			h.log.Error("Failed to get enabled methods", zap.Error(err), zap.String("user_id", userUUID.String()))
			c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
				Success: false,
				Message: "Failed to get enabled methods",
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.TwoFactorResponse{
			Success: true,
			Message: "Enabled methods retrieved successfully",
			Data:    methods,
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	methods, err := h.service.GetEnabledMethods(c.Request.Context(), userUUID)
	if err != nil {
		h.log.Error("Failed to get enabled methods", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to get enabled methods",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Enabled methods retrieved successfully",
		Data:    methods,
	})
}

// GetAvailableMethods returns available 2FA methods for a user
func (h *twoFactorHandler) GetAvailableMethods(c *gin.Context) {
	// Try to get user ID from context (authenticated request)
	userID, exists := c.Get("user_id")
	if !exists {
		// If not authenticated, try to get user ID from query parameter (for login flow)
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "User ID is required",
			})
			return
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid user ID format",
			})
			return
		}

		methods, err := h.service.GetAvailableMethods(c.Request.Context(), userUUID)
		if err != nil {
			h.log.Error("Failed to get available methods", zap.Error(err), zap.String("user_id", userUUID.String()))
			c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
				Success: false,
				Message: "Failed to get available methods",
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, dto.TwoFactorResponse{
			Success: true,
			Message: "Available methods retrieved successfully",
			Data:    methods,
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	methods, err := h.service.GetAvailableMethods(c.Request.Context(), userUUID)
	if err != nil {
		h.log.Error("Failed to get available methods", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to get available methods",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Available methods retrieved successfully",
		Data:    methods,
	})
}

// EnableMethod enables a specific 2FA method
func (h *twoFactorHandler) EnableMethod(c *gin.Context) {
	// Try to get user ID from context (authenticated request)
	userID, exists := c.Get("user_id")
	if !exists {
		// If not authenticated, try to get user ID from request body or query parameter
		var req struct {
			Method string                 `json:"method" binding:"required"`
			Data   map[string]interface{} `json:"data"`
			UserID string                 `json:"user_id,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid request format",
				Error:   err.Error(),
			})
			return
		}

		if req.UserID == "" {
			// Try to get from query parameter
			req.UserID = c.Query("user_id")
		}

		if req.UserID == "" {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "User ID is required",
			})
			return
		}

		userUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid user ID format",
			})
			return
		}

		err = h.service.EnableMethod(c.Request.Context(), userUUID, req.Method, req.Data)
		if err != nil {
			h.log.Error("Failed to enable method", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
			c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
				Success: false,
				Message: "Failed to enable method",
				Error:   err.Error(),
			})
			return
		}

		h.log.Info("Method enabled successfully", zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
		c.JSON(http.StatusOK, dto.TwoFactorResponse{
			Success: true,
			Message: "Method enabled successfully",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req struct {
		Method string                 `json:"method" binding:"required"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	err := h.service.EnableMethod(c.Request.Context(), userUUID, req.Method, req.Data)
	if err != nil {
		h.log.Error("Failed to enable method", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to enable method",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Method enabled successfully", zap.String("user_id", userUUID.String()), zap.String("method", req.Method))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Method enabled successfully",
	})
}

// DisableMethod disables a specific 2FA method
func (h *twoFactorHandler) DisableMethod(c *gin.Context) {
	// Try to get user ID from context (authenticated request)
	userID, exists := c.Get("user_id")
	if !exists {
		// If not authenticated, try to get user ID from request body or query parameter
		var req struct {
			Method           string `json:"method" binding:"required"`
			VerificationData string `json:"verification_data"`
			UserID           string `json:"user_id,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid request format",
				Error:   err.Error(),
			})
			return
		}

		if req.UserID == "" {
			// Try to get from query parameter
			req.UserID = c.Query("user_id")
		}

		if req.UserID == "" {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "User ID is required",
			})
			return
		}

		userUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
				Success: false,
				Message: "Invalid user ID format",
			})
			return
		}

		err = h.service.DisableMethod(c.Request.Context(), userUUID, req.Method, req.VerificationData)
		if err != nil {
			h.log.Error("Failed to disable method", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
			c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
				Success: false,
				Message: "Failed to disable method",
				Error:   err.Error(),
			})
			return
		}

		h.log.Info("Method disabled successfully", zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
		c.JSON(http.StatusOK, dto.TwoFactorResponse{
			Success: true,
			Message: "Method disabled successfully",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req struct {
		Method           string `json:"method" binding:"required"`
		VerificationData string `json:"verification_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	err := h.service.DisableMethod(c.Request.Context(), userUUID, req.Method, req.VerificationData)
	if err != nil {
		h.log.Error("Failed to disable method", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("method", req.Method))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to disable method",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Method disabled successfully", zap.String("user_id", userUUID.String()), zap.String("method", req.Method))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Method disabled successfully",
	})
}

// VerifyWithMethod verifies login using a specific 2FA method
func (h *twoFactorHandler) VerifyWithMethod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req struct {
		Method string `json:"method" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	isValid, err := h.service.VerifyLoginWithMethod(c.Request.Context(), userUUID, req.Method, req.Token, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.log.Error("Failed to verify with method", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("method", req.Method))

		// Check if it's a rate limiting error
		if err.Error() == "too many failed attempts, please try again later" {
			c.JSON(http.StatusTooManyRequests, dto.TwoFactorResponse{
				Success: false,
				Message: "Too many failed attempts. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Verification failed",
			Error:   err.Error(),
		})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid verification code",
		})
		return
	}

	h.log.Info("Login verification successful", zap.String("user_id", userUUID.String()), zap.String("method", req.Method))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Verification successful",
	})
}

// GenerateEmailOTP generates and sends an email OTP
func (h *twoFactorHandler) GenerateEmailOTP(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	err := h.service.GenerateEmailOTP(c.Request.Context(), userUUID, req.Email)
	if err != nil {
		h.log.Error("Failed to generate email OTP", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to generate email OTP",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Email OTP generated successfully", zap.String("user_id", userUUID.String()), zap.String("email", req.Email))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Email OTP sent successfully",
	})
}

// GenerateSMSOTP generates and sends an SMS OTP
func (h *twoFactorHandler) GenerateSMSOTP(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	err := h.service.GenerateSMSOTP(c.Request.Context(), userUUID, req.PhoneNumber)
	if err != nil {
		h.log.Error("Failed to generate SMS OTP", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to generate SMS OTP",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("SMS OTP generated successfully", zap.String("user_id", userUUID.String()), zap.String("phone", req.PhoneNumber))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "SMS OTP sent successfully",
	})
}

// GenerateEmailOTPForLogin generates and sends an email OTP during login (no auth required)
func (h *twoFactorHandler) GenerateEmailOTPForLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Email  string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	err = h.service.GenerateEmailOTP(c.Request.Context(), userUUID, req.Email)
	if err != nil {
		h.log.Error("Failed to generate email OTP for login", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to generate email OTP",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Email OTP generated for login", zap.String("user_id", userUUID.String()), zap.String("email", req.Email))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Email OTP sent successfully",
	})
}

// GenerateSMSOTPForLogin generates and sends an SMS OTP during login (no auth required)
func (h *twoFactorHandler) GenerateSMSOTPForLogin(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	err = h.service.GenerateSMSOTP(c.Request.Context(), userUUID, req.PhoneNumber)
	if err != nil {
		h.log.Error("Failed to generate SMS OTP for login", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to generate SMS OTP",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("SMS OTP generated for login", zap.String("user_id", userUUID.String()), zap.String("phone", req.PhoneNumber))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "SMS OTP sent successfully",
	})
}

// GetBackupCodes retrieves backup codes for a user (for login flow)
func (h *twoFactorHandler) GetBackupCodes(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	codes, err := h.service.GetBackupCodes(c.Request.Context(), userUUID)
	if err != nil {
		h.log.Error("Failed to get backup codes", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to retrieve backup codes",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Backup codes retrieved", zap.String("user_id", userUUID.String()), zap.Int("count", len(codes)))

	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Backup codes retrieved successfully",
		Data: map[string]interface{}{
			"backup_codes": codes,
		},
	})
}

// GenerateSecretForLogin generates a 2FA secret for login flow (no auth required)
func (h *twoFactorHandler) GenerateSecretForLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "User ID is required",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Get user email for secret generation
	// We need to fetch the user's email from the database
	// For now, we'll use the user ID as a fallback if email is not available
	email := req.UserID + "@tucanbit.tv" // Use user ID as fallback email

	// Generate secret using the service
	response, err := h.service.GenerateSecret(c.Request.Context(), userID, email)
	if err != nil {
		h.log.Error("Failed to generate secret", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to generate 2FA secret",
		})
		return
	}

	h.log.Info("2FA secret generated successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "2FA secret generated successfully",
		Data: map[string]interface{}{
			"secret":       response.Secret,
			"qr_code_data": response.QRCodeData,
		},
	})
}

// EnableTOTPForLogin enables TOTP for login flow (no auth required)
func (h *twoFactorHandler) EnableTOTPForLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "User ID and token are required",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Get the secret that was generated earlier
	secret, err := h.service.GetSecret(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get TOTP secret", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "No TOTP secret found. Please generate a secret first.",
		})
		return
	}

	// Verify the token against the secret
	if !h.service.VerifyTOTPToken(secret, req.Token) {
		h.log.Warn("Invalid TOTP token provided", zap.String("user_id", userID.String()))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid TOTP token",
		})
		return
	}

	// Enable TOTP method directly (secret is already saved)
	err = h.service.EnableMethod(c.Request.Context(), userID, "totp", map[string]interface{}{})
	if err != nil {
		h.log.Error("Failed to enable TOTP", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to enable TOTP: " + err.Error(),
		})
		return
	}

	// The EnableMethod should handle enabling the overall 2FA status
	// Let's check if we need to add that functionality to the service

	// Generate proper backup codes using the service
	backupCodes := h.service.GenerateBackupCodes()

	// Save backup codes to the database
	err = h.service.SaveBackupCodes(c.Request.Context(), userID, backupCodes)
	if err != nil {
		h.log.Error("Failed to save backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to save backup codes",
		})
		return
	}

	// Generate JWT token for successful 2FA setup completion
	token, err := utils.GenerateJWTWithVerification(userID, true, true, true)
	if err != nil {
		h.log.Error("Failed to generate JWT token after 2FA setup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to complete 2FA setup",
		})
		return
	}

	h.log.Info("TOTP enabled successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "TOTP enabled successfully",
		Data: map[string]interface{}{
			"backup_codes": backupCodes,
			"access_token": token,
		},
	})
}

// EnableEmailOTPForLogin enables Email OTP for login flow (no auth required)
func (h *twoFactorHandler) EnableEmailOTPForLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "User ID and token are required",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Enable Email OTP using the service
	err = h.service.EnableMethod(c.Request.Context(), userID, "email_otp", map[string]interface{}{
		"verified": true, // Assume verified if they're setting it up
	})
	if err != nil {
		h.log.Error("Failed to enable Email OTP", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to enable Email OTP: " + err.Error(),
		})
		return
	}

	h.log.Info("Email OTP enabled successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Email OTP enabled successfully",
	})
}

// EnableSMSOTPForLogin enables SMS OTP for login flow (no auth required)
func (h *twoFactorHandler) EnableSMSOTPForLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "User ID and token are required",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	// Enable SMS OTP using the service
	err = h.service.EnableMethod(c.Request.Context(), userID, "sms_otp", map[string]interface{}{
		"verified": true, // Assume verified if they're setting it up
	})
	if err != nil {
		h.log.Error("Failed to enable SMS OTP", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to enable SMS OTP: " + err.Error(),
		})
		return
	}

	h.log.Info("SMS OTP enabled successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "SMS OTP enabled successfully",
	})
}

// RegisterPasskey registers a new passkey credential
func (h *twoFactorHandler) RegisterPasskey(c *gin.Context) {
	var req struct {
		Credential struct {
			ID       string `json:"id" binding:"required"`
			RawID    []int  `json:"rawId" binding:"required"`
			Response struct {
				AttestationObject []int `json:"attestationObject" binding:"required"`
				ClientDataJSON    []int `json:"clientDataJSON" binding:"required"`
			} `json:"response" binding:"required"`
			Type string `json:"type" binding:"required"`
		} `json:"credential" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	// Convert credential data to proper format
	credentialData := map[string]interface{}{
		"id":    req.Credential.ID,
		"rawId": req.Credential.RawID,
		"response": map[string]interface{}{
			"attestationObject": req.Credential.Response.AttestationObject,
			"clientDataJSON":    req.Credential.Response.ClientDataJSON,
		},
		"type": req.Credential.Type,
	}

	err = h.service.RegisterPasskey(c.Request.Context(), userUUID, credentialData)
	if err != nil {
		h.log.Error("Failed to register passkey", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to register passkey",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Passkey registered successfully", zap.String("user_id", userUUID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Passkey registered successfully",
	})
}

// GetPasskeyAssertionOptions gets assertion options for passkey verification
func (h *twoFactorHandler) GetPasskeyAssertionOptions(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	options, err := h.service.GetPasskeyAssertionOptions(c.Request.Context(), userUUID)
	if err != nil {
		h.log.Error("Failed to get passkey assertion options", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to get assertion options",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Passkey assertion options generated", zap.String("user_id", userUUID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Assertion options generated successfully",
		Data:    options,
	})
}

// VerifyPasskey verifies a passkey credential
func (h *twoFactorHandler) VerifyPasskey(c *gin.Context) {
	var req struct {
		Credential struct {
			ID       string `json:"id" binding:"required"`
			RawID    []int  `json:"rawId" binding:"required"`
			Response struct {
				AuthenticatorData []int  `json:"authenticatorData" binding:"required"`
				ClientDataJSON    []int  `json:"clientDataJSON" binding:"required"`
				Signature         []int  `json:"signature" binding:"required"`
				UserHandle        []int  `json:"userHandle,omitempty"`
			} `json:"response" binding:"required"`
			Type string `json:"type" binding:"required"`
		} `json:"credential" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	// Convert credential data to proper format
	credentialData := map[string]interface{}{
		"id":    req.Credential.ID,
		"rawId": req.Credential.RawID,
		"response": map[string]interface{}{
			"authenticatorData": req.Credential.Response.AuthenticatorData,
			"clientDataJSON":    req.Credential.Response.ClientDataJSON,
			"signature":         req.Credential.Response.Signature,
			"userHandle":        req.Credential.Response.UserHandle,
		},
		"type": req.Credential.Type,
	}

	success, err := h.service.VerifyPasskey(c.Request.Context(), userUUID, credentialData)
	if err != nil {
		h.log.Error("Failed to verify passkey", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to verify passkey",
			Error:   err.Error(),
		})
		return
	}

	if !success {
		h.log.Warn("Passkey verification failed", zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusUnauthorized, dto.TwoFactorResponse{
			Success: false,
			Message: "Passkey verification failed",
		})
		return
	}

	h.log.Info("Passkey verified successfully", zap.String("user_id", userUUID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Passkey verification successful",
	})
}

// ListPasskeys lists user's passkey credentials
func (h *twoFactorHandler) ListPasskeys(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	passkeys, err := h.service.ListPasskeys(c.Request.Context(), userUUID)
	if err != nil {
		h.log.Error("Failed to list passkeys", zap.Error(err), zap.String("user_id", userUUID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to list passkeys",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Passkeys listed successfully", zap.String("user_id", userUUID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Passkeys retrieved successfully",
		Data:    passkeys,
	})
}

// DeletePasskey deletes a passkey credential
func (h *twoFactorHandler) DeletePasskey(c *gin.Context) {
	var req struct {
		CredentialID string `json:"credential_id" binding:"required"`
		UserID       string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	err = h.service.DeletePasskey(c.Request.Context(), userUUID, req.CredentialID)
	if err != nil {
		h.log.Error("Failed to delete passkey", zap.Error(err), zap.String("user_id", userUUID.String()), zap.String("credential_id", req.CredentialID))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Message: "Failed to delete passkey",
			Error:   err.Error(),
		})
		return
	}

	h.log.Info("Passkey deleted successfully", zap.String("user_id", userUUID.String()), zap.String("credential_id", req.CredentialID))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Passkey deleted successfully",
	})
}
