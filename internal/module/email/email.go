package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendVerificationEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error
	SendWelcomeEmail(email, firstName string) error
	SendPasswordResetEmail(email, resetToken string, expiresAt time.Time) error
	SendPasswordResetOTPEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error
	SendPasswordResetConfirmationEmail(email, firstName string, userAgent, ipAddress string) error
	SendAdminGeneratedPasswordEmail(email, firstName, newPassword string) error
	SendSecurityAlert(email, alertType, details string) error
	SendTwoFactorOTPEmail(email, firstName, otpCode string, expiresAt time.Time, userAgent, ipAddress string) error
	SendAlertEmail(ctx context.Context, to []string, alertConfig interface{}, trigger interface{}) error
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// EmailServiceImpl provides enterprise-grade email functionality
type EmailServiceImpl struct {
	config    SMTPConfig
	logger    *zap.Logger
	templates *template.Template
}

// NewEmailService creates a new instance of EmailService
func NewEmailService(config SMTPConfig, logger *zap.Logger) (EmailService, error) {
	// Load email templates
	templates, err := template.ParseGlob("templates/emails/*.html")
	if err != nil {
		// Fallback to embedded templates if file templates not found
		templates = template.New("emails")
		templates = template.Must(templates.New("verification").Parse(verificationTemplate))
		templates = template.Must(templates.New("welcome").Parse(welcomeTemplate))
		templates = template.Must(templates.New("password_reset").Parse(passwordResetTemplate))
		templates = template.Must(templates.New("password_reset_otp").Parse(passwordResetOTPTemplate))
		templates = template.Must(templates.New("password_reset_confirmation").Parse(passwordResetConfirmationTemplate))
		templates = template.Must(templates.New("admin_generated_password").Parse(adminGeneratedPasswordTemplate))
		templates = template.Must(templates.New("security_alert").Parse(securityAlertTemplate))
		templates = template.Must(templates.New("two_factor_otp").Parse(twoFactorOTPTemplate))
	}

	// Log the config being stored in the email service
	logger.Info("EmailService initialized with SMTP config",
		zap.String("config_source", "SMTPConfig struct passed to NewEmailService"),
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("username", config.Username),
		zap.String("username_raw", config.Username),
		zap.Int("username_length", len(config.Username)),
		zap.Bool("password_set", config.Password != ""),
		zap.Int("password_length", len(config.Password)),
		zap.String("password_raw", config.Password),
		zap.String("from", config.From),
		zap.Bool("use_tls", config.UseTLS))

	return &EmailServiceImpl{
		config:    config,
		logger:    logger,
		templates: templates,
	}, nil
}

// SendVerificationEmail sends a verification email with OTP
func (e *EmailServiceImpl) SendVerificationEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error {
	subject := "Welcome to TucanBIT - Verify your Account"

	// Create template data with device and location info
	templateData := EmailTemplateData{
		Email:            email,
		OTPCode:          otpCode,
		OTPId:            otpId,
		UserID:           userId,
		OTPExpiresAt:     expiresAt,
		VerificationLink: fmt.Sprintf("http://localhost:8094/verify?otp_code=%s&otp_id=%s&user_id=%s", otpCode, otpId, userId),
		SupportEmail:     "support@tucanbit.com",
		WebsiteURL:       "https://app.tucanbit.com",
		CurrentYear:      time.Now().Year(),
		LogoURL:          "cid:tucan.png",
		DeviceInfo:       GetDeviceInfo(userAgent),
		LocationInfo:     GetLocationInfo(ipAddress),
	}

	// Use the modern template
	tmpl := GetModernVerificationEmailTemplate()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("failed to render modern verification template: %w", err)
	}
	htmlBody := buf.String()

	// Debug: Check if UserID is in the rendered HTML
	if !strings.Contains(htmlBody, userId) {
		e.logger.Error("UserID not found in rendered email template",
			zap.String("user_id", userId),
			zap.String("email", email),
			zap.String("template_user_id", templateData.UserID),
			zap.String("verification_link", templateData.VerificationLink))
	} else {
		e.logger.Info("UserID found in rendered email template",
			zap.String("user_id", userId),
			zap.String("email", email),
			zap.String("verification_link", templateData.VerificationLink))
	}

	// Log SMTP configuration for debugging
	e.logger.Info("Attempting to send verification email",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("user_id", userId),
		zap.String("otp_id", otpId),
		zap.String("device", templateData.DeviceInfo),
		zap.String("location", templateData.LocationInfo),
		zap.String("smtp_host", e.config.Host),
		zap.Int("smtp_port", e.config.Port),
		zap.String("smtp_username", e.config.Username),
		zap.String("smtp_from", e.config.From),
		zap.Bool("use_tls", e.config.UseTLS))

	return e.sendEmail(email, subject, htmlBody)
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailServiceImpl) SendWelcomeEmail(email, firstName string) error {
	subject := "Welcome to TucanBIT!"

	// Use the new welcome email template
	tmpl := GetWelcomeEmailTemplate()

	// Prepare template data
	data := struct {
		FirstName    string
		Email        string
		BrandName    string
		LoginURL     string
		SupportEmail string
	}{
		FirstName:    firstName,
		Email:        email,
		BrandName:    "TucanBIT",
		LoginURL:     "http://localhost:8094/login",
		SupportEmail: "support@tucanbit.com",
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute welcome template: %w", err)
	}

	htmlBody := buf.String()
	return e.sendEmail(email, subject, htmlBody)
}

// SendPasswordResetEmail sends a password reset email
func (e *EmailServiceImpl) SendPasswordResetEmail(email, resetToken string, expiresAt time.Time) error {
	subject := "Reset Your Password - TucanBIT"

	data := map[string]interface{}{
		"ResetToken":   resetToken,
		"ExpiresAt":    expiresAt.Format("2006-01-02 15:04:05 UTC"),
		"Email":        email,
		"BrandName":    "TucanBIT",
		"ResetURL":     fmt.Sprintf("https://app.tucanbit.com/reset-password?token=%s", resetToken),
		"SupportEmail": "support@tucanbit.com",
	}

	htmlBody, err := e.renderTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}

	return e.sendEmail(email, subject, htmlBody)
}

// SendPasswordResetOTPEmail sends a password reset OTP email
func (e *EmailServiceImpl) SendPasswordResetOTPEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error {
	subject := "Reset Your Password - TucanBIT"

	// Get device and location info
	device := "Unknown Device"
	location := "Unknown Location"
	if userAgent != "" {
		device = GetDeviceInfo(userAgent)
		location = GetLocationInfo(ipAddress)
	}

	data := map[string]interface{}{
		"OTPCode":      otpCode,
		"OTPID":        otpId,
		"UserID":       userId,
		"ExpiresAt":    expiresAt.Format("2006-01-02 15:04:05 UTC"),
		"Email":        email,
		"BrandName":    "TucanBIT",
		"Device":       device,
		"Location":     location,
		"SupportEmail": "support@tucanbit.com",
		"ResetURL":     fmt.Sprintf("http://localhost:8094/reset-password?otp_code=%s&otp_id=%s&user_id=%s", otpCode, otpId, userId),
	}

	htmlBody, err := e.renderTemplate("password_reset_otp", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset OTP template: %w", err)
	}

	e.logger.Info("Attempting to send password reset OTP email",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("user_id", userId),
		zap.String("otp_id", otpId),
		zap.String("device", device),
		zap.String("location", location),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)),
		zap.String("smtp_username", e.config.Username),
		zap.String("smtp_from", e.config.From),
		zap.Bool("use_tls", e.config.UseTLS))

	err = e.sendEmail(email, subject, htmlBody)
	if err != nil {
		e.logger.Error("Failed to send password reset OTP email",
			zap.String("to", email),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	e.logger.Info("Password reset OTP email sent successfully",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)))

	return nil
}

// SendPasswordResetConfirmationEmail sends a password reset confirmation email
func (e *EmailServiceImpl) SendPasswordResetConfirmationEmail(email, firstName, userAgent, ipAddress string) error {
	subject := "Password Successfully Reset - TucanBIT"

	// Get device and location info
	device := "Unknown Device"
	location := "Unknown Location"
	if userAgent != "" {
		device = GetDeviceInfo(userAgent)
		location = GetLocationInfo(ipAddress)
	}

	// Get current time
	currentTime := time.Now()

	data := map[string]interface{}{
		"FirstName":    firstName,
		"Email":        email,
		"BrandName":    "TucanBIT",
		"ResetTime":    currentTime.Format("January 2, 2006 at 3:04 PM MST"),
		"Device":       device,
		"Location":     location,
		"IPAddress":    ipAddress,
		"LoginURL":     "https://tucanbit.tv/login",
		"SupportEmail": "support@tucanbit.com",
		"CurrentYear":  currentTime.Year(),
	}

	htmlBody, err := e.renderTemplate("password_reset_confirmation", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset confirmation template: %w", err)
	}

	e.logger.Info("Attempting to send password reset confirmation email",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("device", device),
		zap.String("location", location),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)),
		zap.String("smtp_username", e.config.Username),
		zap.String("smtp_from", e.config.From),
		zap.Bool("use_tls", e.config.UseTLS))

	err = e.sendEmail(email, subject, htmlBody)
	if err != nil {
		e.logger.Error("Failed to send password reset confirmation email",
			zap.String("to", email),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	e.logger.Info("Password reset confirmation email sent successfully",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)))

	return nil
}

// SendAdminGeneratedPasswordEmail sends an email with admin-generated password to the player
func (e *EmailServiceImpl) SendAdminGeneratedPasswordEmail(email, firstName, newPassword string) error {
	e.logger.Info("=== SendAdminGeneratedPasswordEmail CALLED ===",
		zap.String("to", email),
		zap.String("first_name", firstName),
		zap.Bool("password_provided", newPassword != ""))

	subject := "Your Password Has Been Reset - TucanBIT"

	data := map[string]interface{}{
		"FirstName":    firstName,
		"Email":        email,
		"BrandName":    "TucanBIT",
		"NewPassword":  newPassword,
		"LoginURL":     "https://tucanbit.tv/login",
		"SupportEmail": "support@tucanbit.com",
		"CurrentYear":  time.Now().Year(),
	}

	e.logger.Info("Rendering email template",
		zap.String("template", "admin_generated_password"))

	htmlBody, err := e.renderTemplate("admin_generated_password", data)
	if err != nil {
		e.logger.Error("Failed to render admin generated password template",
			zap.Error(err))
		return fmt.Errorf("failed to render admin generated password template: %w", err)
	}

	e.logger.Info("Template rendered successfully",
		zap.Int("html_body_length", len(htmlBody)))

	e.logger.Info("Attempting to send admin-generated password email",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)),
		zap.String("smtp_username", e.config.Username),
		zap.String("smtp_from", e.config.From),
		zap.Bool("use_tls", e.config.UseTLS),
		zap.Bool("password_configured", e.config.Password != ""))

	err = e.sendEmail(email, subject, htmlBody)
	if err != nil {
		e.logger.Error("=== FAILED to send admin-generated password email ===",
			zap.String("to", email),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	e.logger.Info("=== Admin-generated password email sent successfully ===",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)))

	return nil
}

// SendSecurityAlert sends a security alert email
func (e *EmailServiceImpl) SendSecurityAlert(email, alertType, details string) error {
	subject := fmt.Sprintf("Security Alert - %s", alertType)

	data := map[string]interface{}{
		"AlertType":    alertType,
		"Details":      details,
		"Email":        email,
		"BrandName":    "TucanBIT",
		"Timestamp":    time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		"SupportEmail": "support@tucanbit.com",
	}

	htmlBody, err := e.renderTemplate("security_alert", data)
	if err != nil {
		return fmt.Errorf("failed to render security alert template: %w", err)
	}

	return e.sendEmail(email, subject, htmlBody)
}

// SendTwoFactorOTPEmail sends a professional two-factor authentication OTP email
func (e *EmailServiceImpl) SendTwoFactorOTPEmail(email, firstName, otpCode string, expiresAt time.Time, userAgent, ipAddress string) error {
	subject := "Two-Factor Authentication Code - TucanBIT Security"

	// Prepare template data
	data := map[string]interface{}{
		"FirstName":        firstName,
		"Email":            email,
		"OTPCode":          otpCode,
		"OTPExpiresAt":     expiresAt,
		"OTPExpiryMinutes": int(time.Until(expiresAt).Minutes()),
		"UserAgent":        userAgent,
		"IPAddress":        ipAddress,
		"DeviceInfo":       GetDeviceInfo(userAgent),
		"LocationInfo":     GetLocationInfo(ipAddress),
		"CompanyName":      "TucanBIT",
		"SupportEmail":     "support@tucanbit.com",
		"SupportPhone":     "+1-800-TUCANBIT",
		"WebsiteURL":       "https://app.tucanbit.com",
		"CurrentYear":      time.Now().Year(),
		"BrandName":        "TucanBIT",
		"Timestamp":        time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}

	// Render the 2FA OTP template
	htmlBody, err := e.renderTemplate("two_factor_otp", data)
	if err != nil {
		e.logger.Error("Failed to render 2FA OTP template",
			zap.String("email", email),
			zap.Error(err))
		return fmt.Errorf("failed to render 2FA OTP template: %w", err)
	}

	// Send the email
	err = e.sendEmail(email, subject, htmlBody)
	if err != nil {
		e.logger.Error("Failed to send 2FA OTP email",
			zap.String("to", email),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	e.logger.Info("2FA OTP email sent successfully",
		zap.String("to", email),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.String("smtp_port", fmt.Sprintf("%d", e.config.Port)))

	return nil
}

// SendAlertEmail sends alert emails to multiple recipients
func (e *EmailServiceImpl) SendAlertEmail(ctx context.Context, to []string, alertConfig interface{}, trigger interface{}) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	// Type assert to get alert config and trigger details
	config, ok := alertConfig.(*dto.AlertConfiguration)
	if !ok {
		// Try to use reflection or handle differently if needed
		config = &dto.AlertConfiguration{Name: "Alert"}
	}

	triggerData, ok := trigger.(*dto.AlertTrigger)
	if !ok {
		triggerData = &dto.AlertTrigger{}
	}

	subject := fmt.Sprintf("🚨 Alert Triggered: %s", config.Name)

	// Build alert details
	alertDetails := fmt.Sprintf(`
Alert Name: %s
Alert Type: %s
Threshold: %.2f
Triggered Value: %.2f
Time Window: %d minutes
`, config.Name, config.AlertType, config.ThresholdAmount, triggerData.TriggerValue, config.TimeWindowMinutes)

	if triggerData.UserID != nil {
		alertDetails += fmt.Sprintf("User ID: %s\n", triggerData.UserID.String())
	}
	if triggerData.TransactionID != nil {
		alertDetails += fmt.Sprintf("Transaction ID: %s\n", *triggerData.TransactionID)
	}
	if triggerData.AmountUSD != nil {
		alertDetails += fmt.Sprintf("Amount USD: %.2f\n", *triggerData.AmountUSD)
	}

	// Create template data
	data := map[string]interface{}{
		"AlertType":    config.Name,
		"AlertDetails": alertDetails,
		"BrandName":    "TucanBIT",
		"Timestamp":    time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		"SupportEmail": "support@tucanbit.com",
	}

	htmlBody, err := e.renderTemplate("security_alert", data)
	if err != nil {
		// Fallback to simple HTML
		htmlBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
	<h2 style="color: #d32f2f;">Alert Triggered: %s</h2>
	<div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
		<pre style="white-space: pre-wrap;">%s</pre>
	</div>
	<p>This is an automated alert from TucanBIT monitoring system.</p>
	<p style="color: #666; font-size: 12px;">Timestamp: %s</p>
</body>
</html>`, config.Name, alertDetails, time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
	}

	// Send to all recipients
	for _, email := range to {
		if email != "" {
			if err := e.sendEmail(email, subject, htmlBody); err != nil {
				e.logger.Error("Failed to send alert email", zap.String("to", email), zap.Error(err))
				// Continue sending to other recipients
				continue
			}
			e.logger.Info("Alert email sent", zap.String("to", email), zap.String("alert", config.Name))
		}
	}

	return nil
}

// sendEmail sends an email using SMTP with logo attachment
func (e *EmailServiceImpl) sendEmail(to, subject, htmlBody string) error {
	e.logger.Info("=== sendEmail START ===",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.Int("smtp_port", e.config.Port),
		zap.String("smtp_username", e.config.Username),
		zap.String("smtp_username_raw", e.config.Username),
		zap.Int("smtp_username_length", len(e.config.Username)),
		zap.String("smtp_password_raw", e.config.Password),
		zap.Int("smtp_password_length", len(e.config.Password)),
		zap.String("smtp_from", e.config.From),
		zap.Bool("use_tls", e.config.UseTLS))

	// Create multipart message for HTML + logo attachment
	boundary := "boundary123"

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", e.config.FromName, e.config.From)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/related; boundary=\"%s\"", boundary)
	headers["X-Mailer"] = "TucanBIT-Email-Service/1.0"

	e.logger.Info("Email headers prepared",
		zap.String("from", headers["From"]),
		zap.String("to", headers["To"]))

	// Build multipart email message with logo attachment
	var message bytes.Buffer
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")

	// Add HTML part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n")

	// Add logo attachment
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: image/png; name=\"tucan.png\"\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("Content-ID: <tucan.png>\r\n")
	message.WriteString("Content-Disposition: inline; filename=\"tucan.png\"\r\n")
	message.WriteString("\r\n")

	// Read and encode the logo file
	logoPath := "internal/module/email/tucan.png"
	logoData, err := os.ReadFile(logoPath)
	if err != nil {
		e.logger.Warn("Failed to read logo file, sending email without logo", zap.Error(err))
	} else {
		// Encode logo as base64
		encodedLogo := make([]byte, base64.StdEncoding.EncodedLen(len(logoData)))
		base64.StdEncoding.Encode(encodedLogo, logoData)
		message.Write(encodedLogo)
	}

	// Add social media icons
	socialIcons := []struct {
		filename  string
		contentID string
	}{
		{"icons8-discord-48.png", "discord.png"},
		{"icons8-telegram-48.png", "telegram.png"},
		{"icons8-instagram-48.png", "instagram.png"},
		{"icons8-twitter-bird-48.png", "twitter.png"},
	}

	for _, icon := range socialIcons {
		iconPath := fmt.Sprintf("internal/module/email/%s", icon.filename)
		iconData, err := os.ReadFile(iconPath)
		if err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to read %s file", icon.filename), zap.Error(err))
			continue
		}

		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString(fmt.Sprintf("Content-Type: image/png; name=\"%s\"\r\n", icon.filename))
		message.WriteString("Content-Transfer-Encoding: base64\r\n")
		message.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", icon.contentID))
		message.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", icon.filename))
		message.WriteString("\r\n")

		encodedIcon := make([]byte, base64.StdEncoding.EncodedLen(len(iconData)))
		base64.StdEncoding.Encode(encodedIcon, iconData)
		message.Write(encodedIcon)
	}

	// Add currency icons
	currencyIcons := []struct {
		filename  string
		contentID string
	}{
		{"icons8-bitcoin-94.png", "bitcoin.png"},
		{"icons8-ethereum-24.png", "ethereum.png"},
		{"icons8-tether-48.png", "tether.png"},
		{"ton_symbol.png", "ton.png"},
		{"dollar-symbol.png", "dollar.png"},
	}

	for _, icon := range currencyIcons {
		iconPath := fmt.Sprintf("internal/module/email/%s", icon.filename)
		iconData, err := os.ReadFile(iconPath)
		if err != nil {
			e.logger.Warn(fmt.Sprintf("Failed to read %s file", icon.filename), zap.Error(err))
			continue
		}

		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString(fmt.Sprintf("Content-Type: image/png; name=\"%s\"\r\n", icon.filename))
		message.WriteString("Content-Transfer-Encoding: base64\r\n")
		message.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", icon.contentID))
		message.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", icon.filename))
		message.WriteString("\r\n")

		encodedIcon := make([]byte, base64.StdEncoding.EncodedLen(len(iconData)))
		base64.StdEncoding.Encode(encodedIcon, iconData)
		message.Write(encodedIcon)
	}

	message.WriteString("\r\n")
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// SMTP authentication
	passwordLength := len(e.config.Password)
	passwordPreview := ""
	if passwordLength > 0 {
		if passwordLength > 4 {
			passwordPreview = e.config.Password[:4] + "..."
		} else {
			passwordPreview = "***"
		}
	}
	// Log password as bytes to detect encoding issues
	passwordBytes := []byte(e.config.Password)
	e.logger.Info("Creating SMTP authentication",
		zap.String("username", e.config.Username),
		zap.String("username_raw", e.config.Username), // Log full username for debugging
		zap.Int("username_length", len(e.config.Username)),
		zap.String("host", e.config.Host),
		zap.Bool("password_set", e.config.Password != ""),
		zap.Int("password_length", passwordLength),
		zap.String("password_preview", passwordPreview),
		zap.String("password_raw", e.config.Password),                      // Log full password for debugging (remove after testing)
		zap.String("password_bytes_hex", fmt.Sprintf("%x", passwordBytes)), // Log password as hex to detect encoding issues
		zap.String("expected_password", "dqys bnjk hhny khbk"),             // Expected value for comparison
		zap.Bool("password_matches_expected", e.config.Password == "dqys bnjk hhny khbk"))
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	e.logger.Info("Connecting to SMTP server",
		zap.String("address", addr),
		zap.Int("port", e.config.Port))

	if e.config.Port == 465 {
		// Port 465 requires SSL (not TLS)
		e.logger.Info("Using SSL (port 465)")
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification for production servers that may not have CA certs
			ServerName:         e.config.Host,
		}
		e.logger.Info("TLS config set",
			zap.Bool("insecure_skip_verify", true),
			zap.String("server_name", e.config.Host),
			zap.String("note", "Certificate verification skipped - common for production servers without CA certs"))

		e.logger.Info("Dialing TLS connection", zap.String("address", addr))
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			e.logger.Error("Failed to establish SSL connection",
				zap.String("address", addr),
				zap.Error(err))
			return fmt.Errorf("failed to establish SSL connection to %s %w", addr, err)
		}
		defer conn.Close()
		e.logger.Info("TLS connection established successfully")

		e.logger.Info("Creating SMTP client", zap.String("host", e.config.Host))
		client, err := smtp.NewClient(conn, e.config.Host)
		if err != nil {
			e.logger.Error("Failed to create SMTP client", zap.Error(err))
			return fmt.Errorf("failed to create SMTP client %w", err)
		}
		defer client.Close()
		e.logger.Info("SMTP client created successfully")

		e.logger.Info("Authenticating with SMTP server")
		if err = client.Auth(auth); err != nil {
			e.logger.Error("Failed to authenticate with SMTP server",
				zap.String("username", e.config.Username),
				zap.Error(err))
			return fmt.Errorf("failed to authenticate with SMTP server %w", err)
		}
		e.logger.Info("SMTP authentication successful")

		e.logger.Info("Setting sender", zap.String("from", e.config.From))
		if err = client.Mail(e.config.From); err != nil {
			e.logger.Error("Failed to set sender", zap.Error(err))
			return fmt.Errorf("failed to set sender %w", err)
		}
		e.logger.Info("Sender set successfully")

		e.logger.Info("Setting recipient", zap.String("to", to))
		if err = client.Rcpt(to); err != nil {
			e.logger.Error("Failed to set recipient", zap.Error(err))
			return fmt.Errorf("failed to set recipient %w", err)
		}
		e.logger.Info("Recipient set successfully")

		e.logger.Info("Getting data writer")
		writer, err := client.Data()
		if err != nil {
			e.logger.Error("Failed to get data writer", zap.Error(err))
			return fmt.Errorf("failed to get data writer %w", err)
		}
		e.logger.Info("Data writer obtained, writing message",
			zap.Int("message_size", len(message.Bytes())))

		_, err = writer.Write(message.Bytes())
		if err != nil {
			e.logger.Error("Failed to write message", zap.Error(err))
			return fmt.Errorf("failed to write message %w", err)
		}
		e.logger.Info("Message written successfully")

		e.logger.Info("Closing writer")
		if err = writer.Close(); err != nil {
			e.logger.Error("Failed to close writer", zap.Error(err))
			return fmt.Errorf("failed to close writer %w", err)
		}
		e.logger.Info("Writer closed successfully")
	} else if e.config.UseTLS {
		// Use STARTTLS for other ports (like 587)
		conn, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server %w", err)
		}
		defer conn.Close()

		if err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true, ServerName: e.config.Host}); err != nil {
			return fmt.Errorf("failed to start TLS %w", err)
		}

		if err = conn.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate %w", err)
		}

		if err = conn.Mail(e.config.From); err != nil {
			return fmt.Errorf("failed to set sender %w", err)
		}

		if err = conn.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient %w", err)
		}

		writer, err := conn.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer%w", err)
		}

		_, err = writer.Write(message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write message %w", err)
		}

		if err = writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer %w", err)
		}
	} else {
		// Use regular SMTP without TLS
		err := smtp.SendMail(addr, auth, e.config.From, []string{to}, message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to send email %w", err)
		}
	}

	e.logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("smtp_host", e.config.Host),
		zap.Int("smtp_port", e.config.Port))

	return nil
}

// renderTemplate renders an email template with data
func (e *EmailServiceImpl) renderTemplate(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := e.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// GetDailyReportEmailTemplate returns the template for daily report emails
func GetDailyReportEmailTemplate() *template.Template {
	return template.Must(template.New("daily_report").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.BrandName}} - Daily Report</title>
    <style>
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0; 
            padding: 0; 
            background-color: #f5f6fa;
        }
        .email-container { 
            max-width: 800px; 
            margin: 0 auto; 
            background-color: white; 
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            border-radius: 10px;
            overflow: hidden;
        }
        .header { 
            background: #2c3e50;
            color: white; 
            padding: 30px; 
            text-align: center;
            position: relative;
        }
        .header h1 { 
            margin: 0; 
            font-size: 32px; 
            font-weight: 700;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }
        .header .subtitle {
            margin: 10px 0 0 0;
            font-size: 16px;
            opacity: 0.9;
        }
        .content { 
            padding: 40px 30px; 
            background: white;
        }
        .daily-report-content {
            margin: 0;
        }
        .daily-report-content h2 {
            color: #2c3e50;
            margin-bottom: 20px;
            font-size: 24px;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        .daily-report-content h3 {
            color: #2c3e50;
            margin: 30px 0 15px 0;
            font-size: 20px;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .metric-card {
            background: white;
            padding: 25px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border: 1px solid #e1e8ed;
            position: relative;
        }
        .metric-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: #2c3e50;
        }
        .metric-card h3 {
            margin: 0 0 10px 0;
            font-size: 36px;
            font-weight: bold;
            color: #2c3e50;
        }
        .metric-card p {
            margin: 0;
            font-size: 14px;
            color: #7f8c8d;
            font-weight: 500;
        }
        .financial-metrics table, .top-games table, .top-players table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
            border-radius: 10px;
            overflow: hidden;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .financial-metrics th, .top-games th, .top-players th {
            background: #2c3e50;
            color: white;
            padding: 18px 15px;
            text-align: left;
            font-weight: 600;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .financial-metrics td, .top-games td, .top-players td {
            padding: 15px;
            border: 1px solid #e1e8ed;
            font-size: 14px;
        }
        .financial-metrics tr:nth-child(even), .top-games tr:nth-child(even), .top-players tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        .financial-metrics tr:hover, .top-games tr:hover, .top-players tr:hover {
            background-color: #e3f2fd;
            transition: background-color 0.3s ease;
        }
        .highlight-revenue {
            background: #f8f9fa !important;
            font-weight: bold;
            color: #2c3e50 !important;
            border-left: 4px solid #2c3e50;
        }
        .footer { 
            background-color: #2c3e50;
            color: #ecf0f1;
            padding: 30px; 
            text-align: center;
            font-size: 14px;
        }
        .footer p { 
            margin: 5px 0; 
        }
        .footer a {
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }
        .footer a:hover {
            color: #5dade2;
            text-decoration: underline;
        }
        .brand-logo {
            width: 60px;
            height: 60px;
            margin-bottom: 20px;
            border-radius: 50%;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        .report-summary {
            background: #f8f9fa;
            color: #2c3e50;
            padding: 25px;
            border-radius: 8px;
            margin: 30px 0;
            text-align: center;
            border: 1px solid #e1e8ed;
        }
        .report-summary h3 {
            margin: 0 0 15px 0;
            font-size: 24px;
        }
        .report-summary p {
            margin: 0;
            font-size: 18px;
            opacity: 0.9;
        }
        .social-icons {
            margin: 20px 0;
            text-align: center;
        }
        .social-icons a {
            display: inline-block;
            margin: 0 10px;
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: #34495e;
            text-align: center;
            line-height: 40px;
            color: white;
            text-decoration: none;
            transition: all 0.3s ease;
        }
        .social-icons a:hover {
            background: #2c3e50;
        }
        @media (max-width: 600px) {
            .metrics-grid {
                grid-template-columns: 1fr;
            }
            .email-container {
                margin: 10px;
            }
            .content {
                padding: 20px 15px;
            }
            .header {
                padding: 20px;
            }
            .header h1 {
                font-size: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <img src="cid:tucan.png" alt="{{.BrandName}} Logo" class="brand-logo">
            <h1>{{.BrandName}}</h1>
            <p class="subtitle">Daily Analytics Report</p>
        </div>
        
        <div class="content">
            <div class="report-summary">
                <h3>Daily Performance Summary</h3>
                <p>Comprehensive overview of your platform's performance</p>
            </div>
            
            {{.ReportHTML}}
            
            <div style="text-align: center; margin: 40px 0 20px 0;">
                <p style="font-size: 16px; color: #7f8c8d; margin: 0;">
                    Keep track of your platform's growth and performance with TucanBIT Analytics.
                </p>
            </div>
        </div>
        
        <div class="footer">
            <div class="social-icons">
                <a href="https://discord.gg/tucanbit" title="Discord">D</a>
                <a href="https://t.me/tucanbit" title="Telegram">T</a>
                <a href="https://instagram.com/tucanbit" title="Instagram">I</a>
                <a href="https://twitter.com/tucanbit" title="Twitter">X</a>
            </div>
            <p><strong>{{.BrandName}} Analytics Platform</strong></p>
            <p>Your trusted partner in gaming analytics</p>
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; {{.CurrentYear}} {{.BrandName}}. All rights reserved.</p>
            <p style="font-size: 12px; margin-top: 15px; opacity: 0.7;">
                This email was automatically generated by the TucanBIT Analytics System.<br>
                For support, contact our technical team at {{.SupportEmail}}
            </p>
        </div>
    </div>
</body>
</html>`))
}

// LoadSMTPConfigFromEnv loads SMTP configuration from environment variables
func LoadSMTPConfigFromEnv() SMTPConfig {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	useTLS, _ := strconv.ParseBool(getEnv("SMTP_USE_TLS", "true"))

	return SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     port,
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "noreply@tucanbit.com"),
		FromName: getEnv("SMTP_FROM_NAME", "TucanBIT"),
		UseTLS:   useTLS,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Embedded email templates
const verificationTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email - {{.BrandName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .otp-box { background: #3498db; color: white; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .warning { background: #f39c12; color: white; padding: 10px; border-radius: 3px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.BrandName}}</h1>
        </div>
        <div class="content">
            <h2>Verify Your Email Address</h2>
            <p>Hello!</p>
            <p>Thank you for signing up with {{.BrandName}}. To complete your registration, please use the verification code below:</p>
            
            <div class="otp-box">
                {{.OTPCode}}
            </div>
            
            <div class="warning">
                ⚠️ This code will expire at {{.ExpiresAt}}
            </div>
            
            <p>If you didn't create an account with {{.BrandName}}, please ignore this email.</p>
            
            <p>Best regards,<br>The {{.BrandName}} Team</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const welcomeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to {{.BrandName}}!</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #27ae60; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .cta-button { display: inline-block; background: #27ae60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.BrandName}}!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FirstName}}!</h2>
            <p>Welcome to {{.BrandName}}! Your account has been successfully created and verified.</p>
            
            <p>You can now:</p>
            <ul>
                <li>Access all platform features</li>
                <li>Start playing games</li>
                <li>Manage your profile</li>
                <li>Connect with other players</li>
            </ul>
            
            <a href="{{.LoginURL}}" class="cta-button">Get Started</a>
            
            <p>If you have any questions or need assistance, our support team is here to help!</p>
            
            <p>Best regards,<br>The {{.BrandName}} Team</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const passwordResetTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password - {{.BrandName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #e74c3c; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .reset-button { display: inline-block; background: #e74c3c; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .warning { background: #f39c12; color: white; padding: 10px; border-radius: 3px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset Your Password</h1>
        </div>
        <div class="content">
            <h2>Password Reset Request</h2>
            <p>Hello!</p>
            <p>We received a request to reset your password for your {{.BrandName}} account.</p>
            
            <a href="{{.ResetURL}}" class="reset-button">Reset Password</a>
            
            <div class="warning">
                ⚠️ This link will expire at {{.ExpiresAt}}
            </div>
            
            <p>If you didn't request a password reset, please ignore this email. Your password will remain unchanged.</p>
            
            <p>For security reasons, this link can only be used once.</p>
            
            <p>Best regards,<br>The {{.BrandName}} Team</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const passwordResetOTPTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password - {{.BrandName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #e74c3c; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .otp-code { background: #2c3e50; color: white; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .warning { background: #f39c12; color: white; padding: 10px; border-radius: 3px; margin: 15px 0; }
        .security-info { background: #ecf0f1; padding: 15px; border-radius: 5px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset Your Password</h1>
        </div>
        <div class="content">
            <h2>Password Reset Request</h2>
            <p>Hello!</p>
            <p>We received a request to reset your password for your {{.BrandName}} account.</p>
            
            <p>To complete the password reset process, please use the following One-Time Password (OTP):</p>
            
            <div class="otp-code">{{.OTPCode}}</div>
            
            <div class="warning">
                ⚠️ This code will expire at {{.ExpiresAt}}
            </div>
            
            <div class="security-info">
                <strong>Security Information:</strong><br>
                • Device: {{.Device}}<br>
                • Location: {{.Location}}<br>
                • Requested at: {{.ExpiresAt}}
            </div>
            
            <p>Enter this code in the password reset page to create a new password.</p>
            
            <p>If you didn't request a password reset, please ignore this email and contact our support team immediately.</p>
            
            <p>For security reasons, this code can only be used once and will expire in 10 minutes.</p>
            
            <p>Best regards,<br>The {{.BrandName}} Team</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const securityAlertTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Security Alert - {{.BrandName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #e74c3c; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .alert-box { background: #e74c3c; color: white; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Security Alert</h1>
        </div>
        <div class="content">
            <h2>Security Alert: {{.AlertType}}</h2>
            <p>Hello!</p>
            
            <div class="alert-box">
                <strong>Alert Details:</strong><br>
                {{.Details}}
            </div>
            
            <p><strong>Time:</strong> {{.Timestamp}}</p>
            <p><strong>Account:</strong> {{.Email}}</p>
            
            <p>If this activity was not authorized by you, please:</p>
            <ol>
                <li>Change your password immediately</li>
                <li>Enable two-factor authentication</li>
                <li>Contact our support team</li>
            </ol>
            
            <p>For security reasons, we recommend reviewing your account activity and ensuring your account is secure.</p>
            
            <p>Best regards,<br>The {{.BrandName}} Security Team</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const adminGeneratedPasswordTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset - {{.BrandName}}</title>
    <style>
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0; 
            padding: 0; 
            background-color: #f5f6fa;
        }
        .email-container { 
            max-width: 600px; 
            margin: 0 auto; 
            background-color: white; 
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            border-radius: 10px;
            overflow: hidden;
        }
        .header { 
            background: linear-gradient(135deg, #e74c3c, #c0392b);
            color: white; 
            padding: 40px 30px; 
            text-align: center;
        }
        .header h1 { 
            margin: 0; 
            font-size: 28px; 
            font-weight: 700;
        }
        .content { 
            padding: 40px 30px; 
            background: white;
        }
        .greeting {
            font-size: 20px;
            color: #2c3e50;
            margin-bottom: 20px;
            font-weight: 600;
        }
        .message {
            font-size: 16px;
            line-height: 1.8;
            margin-bottom: 30px;
            color: #555;
        }
        .password-box {
            background: #f8f9fa;
            border: 2px dashed #e74c3c;
            border-radius: 8px;
            padding: 20px;
            margin: 30px 0;
            text-align: center;
        }
        .password-label {
            font-size: 14px;
            color: #7f8c8d;
            margin-bottom: 10px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .password-value {
            font-size: 24px;
            color: #2c3e50;
            font-family: 'Courier New', monospace;
            font-weight: 700;
            letter-spacing: 2px;
            word-break: break-all;
            background: white;
            padding: 15px;
            border-radius: 5px;
            margin-top: 10px;
        }
        .warning-box {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 20px;
            margin: 30px 0;
            border-radius: 0 8px 8px 0;
        }
        .warning-box h3 {
            margin: 0 0 10px 0;
            color: #856404;
            font-size: 18px;
        }
        .warning-box p {
            margin: 0;
            color: #856404;
            font-size: 14px;
        }
        .cta-section {
            text-align: center;
            margin: 40px 0;
        }
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #e74c3c, #c0392b);
            color: white;
            padding: 15px 30px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 16px;
            box-shadow: 0 4px 8px rgba(231, 76, 60, 0.3);
            transition: all 0.3s ease;
        }
        .cta-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(231, 76, 60, 0.4);
        }
        .footer { 
            background-color: #2c3e50;
            color: #ecf0f1;
            padding: 30px; 
            text-align: center;
            font-size: 14px;
        }
        .footer p { 
            margin: 5px 0; 
        }
        .footer a {
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }
        .footer a:hover {
            color: #5dade2;
            text-decoration: underline;
        }
        @media (max-width: 600px) {
            .email-container {
                margin: 10px;
            }
            .content {
                padding: 20px 15px;
            }
            .header {
                padding: 30px 20px;
            }
            .header h1 {
                font-size: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>Password Reset by Administrator</h1>
        </div>
        <div class="content">
            <div class="greeting">Hello {{.FirstName}},</div>
            <div class="message">
                <p>Your password has been reset by an administrator. Below is your new temporary password. Please log in and change it to a password of your choice for security purposes.</p>
            </div>
            
            <div class="password-box">
                <div class="password-label">Your New Password</div>
                <div class="password-value">{{.NewPassword}}</div>
            </div>
            
            <div class="warning-box">
                <h3>⚠️ Security Notice</h3>
                <p>For your security, please change this password immediately after logging in. Do not share this password with anyone.</p>
            </div>
            
            <div class="cta-section">
                <a href="{{.LoginURL}}" class="cta-button">Log In Now</a>
            </div>
            
            <div class="message">
                <p>If you did not request this password reset, please contact our support team immediately at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>
                <p>Best regards,<br>The {{.BrandName}} Team</p>
            </div>
        </div>
        <div class="footer">
            <p>This email was sent to {{.Email}} by an administrator.</p>
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; {{.CurrentYear}} {{.BrandName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const twoFactorOTPTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Two-Factor Authentication Code - TucanBIT</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f4f4;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 300;
        }
        .logo {
            width: 80px;
            height: 80px;
            margin-bottom: 20px;
            border-radius: 50%;
            background-color: rgba(255,255,255,0.2);
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 24px;
            font-weight: bold;
        }
        .content {
            padding: 40px 30px;
        }
        .welcome {
            font-size: 24px;
            color: #2c3e50;
            margin-bottom: 20px;
            text-align: center;
        }
        .message {
            font-size: 16px;
            color: #555;
            margin-bottom: 30px;
            text-align: center;
        }
        .otp-container {
            background-color: #f8f9fa;
            border: 2px solid #e9ecef;
            border-radius: 10px;
            padding: 30px;
            text-align: center;
            margin: 30px 0;
        }
        .otp-code {
            font-size: 32px;
            font-weight: bold;
            color: #2c3e50;
            letter-spacing: 5px;
            margin: 20px 0;
            font-family: 'Courier New', monospace;
        }
        .otp-expires {
            font-size: 14px;
            color: #6c757d;
            margin-top: 15px;
        }
        .security-box {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 20px;
            margin: 30px 0;
            border-radius: 5px;
        }
        .security-box h3 {
            margin: 0 0 10px 0;
            color: #856404;
            font-size: 18px;
        }
        .security-box p {
            margin: 0;
            color: #856404;
        }
        .info-box {
            background-color: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 20px;
            margin: 30px 0;
            border-radius: 5px;
        }
        .info-box h3 {
            margin: 0 0 10px 0;
            color: #1976d2;
            font-size: 18px;
        }
        .info-box p {
            margin: 0;
            color: #1565c0;
        }
        .footer {
            background-color: #2c3e50;
            color: white;
            padding: 30px;
            text-align: center;
        }
        .footer p {
            margin: 5px 0;
            font-size: 14px;
        }
        .social-links {
            margin: 20px 0;
        }
        .social-links a {
            color: white;
            text-decoration: none;
            margin: 0 10px;
            font-size: 16px;
        }
        .support-info {
            background-color: #f8f9fa;
            border-radius: 5px;
            padding: 20px;
            margin: 20px 0;
            text-align: center;
        }
        .support-info h4 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .support-info p {
            margin: 5px 0;
            color: #6c757d;
        }
        @media (max-width: 600px) {
            .container {
                margin: 10px;
            }
            .content {
                padding: 20px 15px;
            }
            .header {
                padding: 20px 15px;
            }
            .otp-code {
                font-size: 24px;
                letter-spacing: 3px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">TB</div>
            <h1>TucanBIT Security</h1>
        </div>
        
        <div class="content">
            <div class="welcome">Two-Factor Authentication Code</div>
            
            <div class="message">
                Hello {{.FirstName}}, you've requested a two-factor authentication code to secure your TucanBIT account.
            </div>
            
            <div class="otp-container">
                <h3>Your Security Code</h3>
                <div class="otp-code">{{.OTPCode}}</div>
                <p>Enter this code in the application to complete your login.</p>
                <div class="otp-expires">
                    This code expires at {{.OTPExpiresAt.Format "3:04 PM MST on January 2, 2006"}}
                </div>
            </div>
            
            <div class="security-box">
                <h3>🔒 Security Notice</h3>
                <p>Never share this code with anyone. TucanBIT staff will never ask for your two-factor authentication code. If you didn't request this code, please secure your account immediately.</p>
            </div>
            
            <div class="info-box">
                <h3>ℹ️ Important Information</h3>
                <p>This code is valid for {{.OTPExpiryMinutes}} minutes only. If you're having trouble logging in, please contact our support team.</p>
            </div>
            
            <div class="support-info">
                <h4>Need Help?</h4>
                <p>Email: {{.SupportEmail}}</p>
                <p>Phone: {{.SupportPhone}}</p>
                <p>Website: <a href="{{.WebsiteURL}}">{{.WebsiteURL}}</a></p>
            </div>
        </div>
        
        <div class="footer">
            <div class="social-links">
                <a href="#">Twitter</a> |
                <a href="#">LinkedIn</a> |
                <a href="#">Facebook</a>
            </div>
            <p>&copy; {{.CurrentYear}} TucanBIT. All rights reserved.</p>
            <p>This email was sent to {{.Email}} for security purposes.</p>
        </div>
    </div>
</body>
</html>`

const passwordResetConfirmationTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Successfully Reset - {{.BrandName}}</title>
    <style>
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0; 
            padding: 0; 
            background-color: #f5f6fa;
        }
        .email-container { 
            max-width: 600px; 
            margin: 0 auto; 
            background-color: white; 
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            border-radius: 10px;
            overflow: hidden;
        }
        .header { 
            background: linear-gradient(135deg, #27ae60, #2ecc71);
            color: white; 
            padding: 40px 30px; 
            text-align: center;
            position: relative;
        }
        .header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><pattern id="grain" width="100" height="100" patternUnits="userSpaceOnUse"><circle cx="25" cy="25" r="1" fill="white" opacity="0.1"/><circle cx="75" cy="75" r="1" fill="white" opacity="0.1"/><circle cx="50" cy="10" r="0.5" fill="white" opacity="0.1"/><circle cx="10" cy="60" r="0.5" fill="white" opacity="0.1"/><circle cx="90" cy="40" r="0.5" fill="white" opacity="0.1"/></pattern></defs><rect width="100" height="100" fill="url(%23grain)"/></svg>');
            opacity: 0.3;
        }
        .header h1 { 
            margin: 0; 
            font-size: 32px; 
            font-weight: 700;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
            position: relative;
            z-index: 1;
        }
        .header .subtitle {
            margin: 10px 0 0 0;
            font-size: 16px;
            opacity: 0.9;
            position: relative;
            z-index: 1;
        }
        .success-icon {
            width: 80px;
            height: 80px;
            background: white;
            border-radius: 50%;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 40px;
            color: #27ae60;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
            position: relative;
            z-index: 1;
        }
        .content { 
            padding: 40px 30px; 
            background: white;
        }
        .greeting {
            font-size: 24px;
            color: #2c3e50;
            margin-bottom: 20px;
            font-weight: 600;
        }
        .message {
            font-size: 16px;
            line-height: 1.8;
            margin-bottom: 30px;
            color: #555;
        }
        .security-info {
            background: #f8f9fa;
            border-left: 4px solid #27ae60;
            padding: 20px;
            margin: 30px 0;
            border-radius: 0 8px 8px 0;
        }
        .security-info h3 {
            margin: 0 0 15px 0;
            color: #2c3e50;
            font-size: 18px;
        }
        .security-info ul {
            margin: 0;
            padding-left: 20px;
        }
        .security-info li {
            margin-bottom: 8px;
            color: #555;
        }
        .device-info {
            background: #ecf0f1;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            font-size: 14px;
            color: #7f8c8d;
        }
        .cta-section {
            text-align: center;
            margin: 40px 0;
        }
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #27ae60, #2ecc71);
            color: white;
            padding: 15px 30px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 16px;
            box-shadow: 0 4px 8px rgba(39, 174, 96, 0.3);
            transition: all 0.3s ease;
        }
        .cta-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(39, 174, 96, 0.4);
        }
        .footer { 
            background-color: #2c3e50;
            color: #ecf0f1;
            padding: 30px; 
            text-align: center;
            font-size: 14px;
        }
        .footer p { 
            margin: 5px 0; 
        }
        .footer a {
            color: #3498db;
            text-decoration: none;
            font-weight: 500;
        }
        .footer a:hover {
            color: #5dade2;
            text-decoration: underline;
        }
        .brand-logo {
            width: 60px;
            height: 60px;
            margin-bottom: 20px;
            border-radius: 50%;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        .social-icons {
            margin: 20px 0;
            text-align: center;
        }
        .social-icons a {
            display: inline-block;
            margin: 0 10px;
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: #34495e;
            text-align: center;
            line-height: 40px;
            color: white;
            text-decoration: none;
            transition: all 0.3s ease;
        }
        .social-icons a:hover {
            background: #2c3e50;
        }
        @media (max-width: 600px) {
            .email-container {
                margin: 10px;
            }
            .content {
                padding: 20px 15px;
            }
            .header {
                padding: 30px 20px;
            }
            .header h1 {
                font-size: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <img src="cid:tucan.png" alt="{{.BrandName}} Logo" class="brand-logo">
            <div class="success-icon">✓</div>
            <h1>Password Successfully Reset</h1>
            <p class="subtitle">Your account security has been updated</p>
        </div>
        
        <div class="content">
            <div class="greeting">Hello {{.FirstName}}!</div>
            
            <div class="message">
                <p>We're writing to confirm that your password has been successfully reset for your {{.BrandName}} account.</p>
                
                <p><strong>This email serves as confirmation that:</strong></p>
                <ul>
                    <li>Your password reset request was completed successfully</li>
                    <li>Your new password is now active and secure</li>
                    <li>All previous login sessions have been terminated for security</li>
                </ul>
            </div>
            
            <div class="security-info">
                <h3>🔒 Security Information</h3>
                <p><strong>Reset Details:</strong></p>
                <ul>
                    <li><strong>Date & Time:</strong> {{.ResetTime}}</li>
                    <li><strong>Device:</strong> {{.Device}}</li>
                    <li><strong>Location:</strong> {{.Location}}</li>
                    <li><strong>IP Address:</strong> {{.IPAddress}}</li>
                </ul>
            </div>
            
            <div class="message">
                <p><strong>Important Security Reminders:</strong></p>
                <ul>
                    <li>Keep your password confidential and don't share it with anyone</li>
                    <li>Use a unique password that you don't use for other accounts</li>
                    <li>Consider enabling two-factor authentication for added security</li>
                    <li>If you didn't request this password reset, contact our support team immediately</li>
                </ul>
            </div>
            
            <div class="cta-section">
                <a href="{{.LoginURL}}" class="cta-button">Sign In to Your Account</a>
            </div>
            
            <div class="message">
                <p>If you have any questions or concerns about this password reset, please don't hesitate to contact our support team. We're here to help ensure your account remains secure.</p>
                
                <p>Thank you for choosing {{.BrandName}}!</p>
            </div>
        </div>
        
        <div class="footer">
            <div class="social-icons">
                <a href="https://discord.gg/tucanbit" title="Discord">D</a>
                <a href="https://t.me/tucanbit" title="Telegram">T</a>
                <a href="https://instagram.com/tucanbit" title="Instagram">I</a>
                <a href="https://twitter.com/tucanbit" title="Twitter">X</a>
            </div>
            <p><strong>{{.BrandName}} Security Team</strong></p>
            <p>Your account security is our priority</p>
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; {{.CurrentYear}} {{.BrandName}}. All rights reserved.</p>
            <p style="font-size: 12px; margin-top: 15px; opacity: 0.7;">
                This email was automatically generated by the TucanBIT Security System.<br>
                For security concerns, contact our support team at {{.SupportEmail}}
            </p>
        </div>
    </div>
</body>
</html>`
