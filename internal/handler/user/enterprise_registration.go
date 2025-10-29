package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	userModule "github.com/tucanbit/internal/module/user"
	"go.uber.org/zap"
)

// EnterpriseRegistrationHandler handles enterprise registration requests
type EnterpriseRegistrationHandler struct {
	registrationService *userModule.EnterpriseRegistrationService
	logger              *zap.Logger
}

// NewEnterpriseRegistrationHandler creates a new enterprise registration handler
func NewEnterpriseRegistrationHandler(
	registrationService *userModule.EnterpriseRegistrationService,
	logger *zap.Logger,
) *EnterpriseRegistrationHandler {
	return &EnterpriseRegistrationHandler{
		registrationService: registrationService,
		logger:              logger,
	}
}

// InitiateRegistration starts the enterprise registration process
// @Summary Initiate enterprise registration
// @Description Start the enterprise registration process with email verification
// @Tags Enterprise Registration
// @Accept json
// @Produce json
// @Param request body dto.EnterpriseRegistrationRequest true "Enterprise registration request"
// @Success 200 {object} dto.EnterpriseRegistrationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/enterprise/register [post]
func (h *EnterpriseRegistrationHandler) InitiateRegistration(c *gin.Context) {
	var req dto.EnterpriseRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind enterprise registration request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.UserType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email, password, first name, last name, and user type are required",
		})
		return
	}

	h.logger.Info("Initiating enterprise registration",
		zap.String("email", req.Email),
		zap.String("user_type", req.UserType),
		zap.String("ip", c.ClientIP()))

	response, err := h.registrationService.InitiateRegistration(c.Request.Context(), &req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.logger.Error("Failed to initiate enterprise registration",
			zap.Error(err),
			zap.String("email", req.Email))

		// Handle specific error cases
		if err.Error() == "user with email "+req.Email+" already exists" {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Code:    http.StatusConflict,
				Message: "User with this email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to initiate registration",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteRegistration completes the enterprise registration process
// @Summary Complete enterprise registration
// @Description Complete the enterprise registration process by verifying OTP
// @Tags Enterprise Registration
// @Accept json
// @Produce json
// @Param request body dto.EnterpriseRegistrationCompleteRequest true "Registration completion request"
// @Success 200 {object} dto.EnterpriseRegistrationCompleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/enterprise/register/complete [post]
func (h *EnterpriseRegistrationHandler) CompleteRegistration(c *gin.Context) {
	var req dto.EnterpriseRegistrationCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind registration completion request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil || req.OTPID == uuid.Nil || req.OTPCode == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "User ID, OTP ID, and OTP code are required",
		})
		return
	}

	h.logger.Info("Completing enterprise registration",
		zap.String("user_id", req.UserID.String()),
		zap.String("otp_id", req.OTPID.String()),
		zap.String("ip", c.ClientIP()))

	response, err := h.registrationService.CompleteRegistration(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to complete enterprise registration",
			zap.Error(err),
			zap.String("user_id", req.UserID.String()))

		// Handle specific error cases
		if err.Error() == "OTP not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "OTP not found or invalid",
			})
			return
		}

		if err.Error() == "OTP has expired" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "OTP has expired. Please request a new one.",
			})
			return
		}

		if err.Error() == "invalid OTP code" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid OTP code. Please check and try again.",
			})
			return
		}

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to complete registration",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRegistrationStatus gets the current registration status
// @Summary Get registration status
// @Description Get the current status of an enterprise registration
// @Tags Enterprise Registration
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {object} dto.EnterpriseRegistrationStatus
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/enterprise/register/status/{user_id} [get]
func (h *EnterpriseRegistrationHandler) GetRegistrationStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID format",
		})
		return
	}

	h.logger.Info("Getting registration status",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()))

	status, err := h.registrationService.GetRegistrationStatus(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get registration status",
			zap.Error(err),
			zap.String("user_id", userID.String()))

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get registration status",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ResendVerificationEmail resends the verification email
// @Summary Resend verification email
// @Description Resend the verification email for enterprise registration
// @Tags Enterprise Registration
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationEmailRequest true "Resend verification request"
// @Success 200 {object} dto.EnterpriseRegistrationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/enterprise/register/resend [post]
func (h *EnterpriseRegistrationHandler) ResendVerificationEmail(c *gin.Context) {
	var req dto.ResendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind resend verification request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email is required",
		})
		return
	}

	h.logger.Info("Resending verification email",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	response, err := h.registrationService.ResendVerificationEmail(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		h.logger.Error("Failed to resend verification email",
			zap.Error(err),
			zap.String("email", req.Email))

		// Handle specific error cases
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
			})
			return
		}

		if err.Error() == "user is already verified" {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Code:    http.StatusConflict,
				Message: "User is already verified",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to resend verification email",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
