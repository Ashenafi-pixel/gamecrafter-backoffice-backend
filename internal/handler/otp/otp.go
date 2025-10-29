package otp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/otp"
	"go.uber.org/zap"
)

// OTPHandler handles OTP-related HTTP requests
type OTPHandler struct {
	otpModule otp.OTPModule
	logger    *zap.Logger
}

// NewOTPHandler creates a new instance of OTPHandler
func NewOTPHandler(otpModule otp.OTPModule, logger *zap.Logger) *OTPHandler {
	return &OTPHandler{
		otpModule: otpModule,
		logger:    logger,
	}
}

// CreateEmailVerification creates a new email verification OTP
// @Summary Create email verification OTP
// @Description Send a verification OTP to the specified email address
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body dto.EmailVerificationRequest true "Email verification request"
// @Success 200 {object} dto.EmailVerificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/email-verification [post]
func (h *OTPHandler) CreateEmailVerification(c *gin.Context) {
	var req dto.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind email verification request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate email format
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email is required",
		})
		return
	}

	h.logger.Info("Creating email verification OTP",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	response, err := h.otpModule.CreateEmailVerification(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP(), "")
	if err != nil {
		h.logger.Error("Failed to create email verification OTP",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create verification OTP",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyOTP verifies an OTP code
// @Summary Verify OTP code
// @Description Verify the OTP code sent to the user's email
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body dto.OTPVerificationRequest true "OTP verification request"
// @Success 200 {object} dto.OTPVerificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/verify [post]
func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req dto.OTPVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind OTP verification request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate request
	if req.Email == "" || req.OTPCode == "" || req.OTPID == uuid.Nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email, OTP code, and OTP ID are required",
		})
		return
	}

	h.logger.Info("Verifying OTP",
		zap.String("email", req.Email),
		zap.String("otp_id", req.OTPID.String()),
		zap.String("ip", c.ClientIP()))

	response, err := h.otpModule.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to verify OTP",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("otp_id", req.OTPID.String()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResendOTP resends an OTP to the specified email
// @Summary Resend OTP
// @Description Resend an OTP to the specified email address
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body dto.ResendOTPRequest true "Resend OTP request"
// @Success 200 {object} dto.ResendOTPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/resend [post]
func (h *OTPHandler) ResendOTP(c *gin.Context) {
	var req dto.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind resend OTP request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate email
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email is required",
		})
		return
	}

	h.logger.Info("Resending OTP",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	response, err := h.otpModule.ResendOTP(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.logger.Error("Failed to resend OTP",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResendPasswordResetOTP resends a password reset OTP to the specified email
// @Summary Resend Password Reset OTP
// @Description Resend a password reset OTP to the specified email address
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body dto.ResendOTPRequest true "Resend password reset OTP request"
// @Success 200 {object} dto.ResendOTPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/resend-password-reset [post]
func (h *OTPHandler) ResendPasswordResetOTP(c *gin.Context) {
	var req dto.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind resend password reset OTP request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate email
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email is required",
		})
		return
	}

	h.logger.Info("Resending password reset OTP",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	response, err := h.otpModule.ResendPasswordResetOTP(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.logger.Error("Failed to resend password reset OTP",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// InvalidateOTP invalidates an OTP
// @Summary Invalidate OTP
// @Description Invalidate an OTP by ID
// @Tags OTP
// @Accept json
// @Produce json
// @Param otp_id path string true "OTP ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/{otp_id}/invalidate [delete]
func (h *OTPHandler) InvalidateOTP(c *gin.Context) {
	otpIDStr := c.Param("otp_id")
	if otpIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "OTP ID is required",
		})
		return
	}

	otpID, err := uuid.Parse(otpIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid OTP ID format",
		})
		return
	}

	h.logger.Info("Invalidating OTP",
		zap.String("otp_id", otpID.String()),
		zap.String("ip", c.ClientIP()))

	err = h.otpModule.InvalidateOTP(c.Request.Context(), otpID)
	if err != nil {
		h.logger.Error("Failed to invalidate OTP",
			zap.Error(err),
			zap.String("otp_id", otpID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to invalidate OTP",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "OTP invalidated successfully",
	})
}

// GetOTPInfo retrieves OTP information
// @Summary Get OTP information
// @Description Get information about a specific OTP
// @Tags OTP
// @Accept json
// @Produce json
// @Param otp_id path string true "OTP ID"
// @Success 200 {object} dto.OTPInfo
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/otp/{otp_id} [get]
func (h *OTPHandler) GetOTPInfo(c *gin.Context) {
	otpIDStr := c.Param("otp_id")
	if otpIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "OTP ID is required",
		})
		return
	}

	otpID, err := uuid.Parse(otpIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid OTP ID format",
		})
		return
	}

	h.logger.Info("Getting OTP info",
		zap.String("otp_id", otpID.String()),
		zap.String("ip", c.ClientIP()))

	otpInfo, err := h.otpModule.GetOTPByID(c.Request.Context(), otpID)
	if err != nil {
		h.logger.Error("Failed to get OTP info",
			zap.Error(err),
			zap.String("otp_id", otpID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve OTP information",
		})
		return
	}

	if otpInfo == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "OTP not found",
		})
		return
	}

	c.JSON(http.StatusOK, otpInfo)
}

// CleanupExpiredOTPs cleans up expired OTPs (admin endpoint)
// @Summary Cleanup expired OTPs
// @Description Remove expired OTPs from the database
// @Tags OTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/admin/otp/cleanup [post]
func (h *OTPHandler) CleanupExpiredOTPs(c *gin.Context) {
	// This endpoint should be protected by admin middleware
	// For now, we'll just log the request
	h.logger.Info("Cleaning up expired OTPs",
		zap.String("ip", c.ClientIP()))

	err := h.otpModule.CleanupExpiredOTPs(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to cleanup expired OTPs",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to cleanup expired OTPs",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Expired OTPs cleaned up successfully",
	})
}

// GetOTPStats retrieves OTP statistics (admin endpoint)
// @Summary Get OTP statistics
// @Description Get statistics about OTP usage and performance
// @Tags OTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param days query int false "Number of days to look back (default: 30)"
// @Success 200 {object} dto.OTPStatsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/admin/otp/stats [get]
func (h *OTPHandler) GetOTPStats(c *gin.Context) {
	// Parse query parameters
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	h.logger.Info("Getting OTP statistics",
		zap.Int("days", days),
		zap.String("ip", c.ClientIP()))

	// This would typically call a method on the OTP module
	// For now, return placeholder data
	stats := dto.OTPStatsResponse{
		TotalOTPs:           0,
		VerifiedOTPs:        0,
		ExpiredOTPs:         0,
		FailedAttempts:      0,
		SuccessRate:         0.0,
		AverageResponseTime: 0.0,
		Period:              days,
	}

	c.JSON(http.StatusOK, stats)
}
