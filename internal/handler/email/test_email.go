package email

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/email"
	"go.uber.org/zap"
)

type TestEmailHandler struct {
	emailService email.EmailService
	log          *zap.Logger
}

func NewTestEmailHandler(emailService email.EmailService, log *zap.Logger) *TestEmailHandler {
	return &TestEmailHandler{
		emailService: emailService,
		log:          log,
	}
}

// SendTestEmail sends a test email to verify SMTP configuration
// @Summary Send test email
// @Description Send a test email to verify SMTP configuration from config.yaml
// @Tags email
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Test email request" example({"email":"test@example.com","first_name":"Test User"})
// @Success 200 {object} object "Success response with email details"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/admin/test-email [post]
func (h *TestEmailHandler) SendTestEmail(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		FirstName string `json:"first_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request format for test email", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	if h.emailService == nil {
		h.log.Error("Email service not available")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Email service not available",
		})
		return
	}

	firstName := req.FirstName
	if firstName == "" {
		firstName = "Test User"
	}

	h.log.Info("Sending test email", zap.String("email", req.Email), zap.String("first_name", firstName))

	err := h.emailService.SendWelcomeEmail(req.Email, firstName)
	if err != nil {
		h.log.Error("Failed to send test email", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send test email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test email sent successfully to " + req.Email,
		"data": map[string]interface{}{
			"email":      req.Email,
			"first_name": firstName,
			"email_type": "Welcome Email",
			"sent_at":    time.Now().Format(time.RFC3339),
		},
	})
}
