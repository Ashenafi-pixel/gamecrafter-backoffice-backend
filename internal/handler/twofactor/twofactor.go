package twofactor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/response"
	"go.uber.org/zap"
)

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
	
	// Backup codes endpoints
	GetBackupCodes(c *gin.Context)
	RegenerateBackupCodes(c *gin.Context)
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
	secret, err := h.service.GenerateSecret(userID, email)
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

	var req dto.TwoFactorSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Get the secret from the request (it should be stored temporarily)
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
	var req dto.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.TwoFactorResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Get user ID from request (this endpoint is used during login)
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		h.log.Error("User ID not provided in header")
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

	// Verify the token
	success, err := h.service.VerifyLoginToken(c.Request.Context(), userID, req.Token, req.BackupCode, ip, userAgent)
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
		
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
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
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Token verified successfully",
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
	err = h.service.Disable2FA(c.Request.Context(), userID, req.Token)
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

// GetBackupCodes retrieves backup codes for the user
// @Summary Get Backup Codes
// @Description Get backup codes for the user (only if 2FA is enabled)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 403 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/backup-codes [get]
func (h *twoFactorHandler) GetBackupCodes(c *gin.Context) {
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

	// Check if 2FA is enabled
	status, err := h.service.Get2FAStatus(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to get 2FA status",
		})
		return
	}

	if !status.IsEnabled {
		h.log.Warn("User tried to get backup codes without 2FA enabled", zap.String("user_id", userID.String()))
		c.JSON(http.StatusForbidden, dto.TwoFactorResponse{
			Success: false,
			Error:   "2FA is not enabled",
		})
		return
	}

	// Return backup codes (without showing them in logs for security)
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Data: dto.BackupCodesResponse{
			BackupCodes: status.BackupCodes,
			Warning:     "Store these codes securely. Each code can only be used once.",
		},
	})
}

// RegenerateBackupCodes generates new backup codes
// @Summary Regenerate Backup Codes
// @Description Generate new backup codes for the user (invalidates old ones)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} dto.TwoFactorResponse
// @Failure 401 {object} dto.TwoFactorResponse
// @Failure 403 {object} dto.TwoFactorResponse
// @Failure 500 {object} dto.TwoFactorResponse
// @Router /api/auth/2fa/regenerate-codes [post]
func (h *twoFactorHandler) RegenerateBackupCodes(c *gin.Context) {
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

	// Check if 2FA is enabled
	status, err := h.service.Get2FAStatus(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get 2FA status", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to get 2FA status",
		})
		return
	}

	if !status.IsEnabled {
		h.log.Warn("User tried to regenerate backup codes without 2FA enabled", zap.String("user_id", userID.String()))
		c.JSON(http.StatusForbidden, dto.TwoFactorResponse{
			Success: false,
			Error:   "2FA is not enabled",
		})
		return
	}

	// Regenerate backup codes
	newCodes, err := h.service.RegenerateBackupCodes(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Failed to regenerate backup codes", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.TwoFactorResponse{
			Success: false,
			Error:   "Failed to regenerate backup codes",
		})
		return
	}

	h.log.Info("Backup codes regenerated successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, dto.TwoFactorResponse{
		Success: true,
		Message: "Backup codes regenerated successfully",
		Data: dto.BackupCodesResponse{
			BackupCodes: newCodes,
			Warning:     "Store these codes securely. Each code can only be used once. Old codes are now invalid.",
		},
	})
}
