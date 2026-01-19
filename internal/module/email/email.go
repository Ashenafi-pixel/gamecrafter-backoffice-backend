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

	"github.com/spf13/viper"
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

func NewEmailService(config SMTPConfig, logger *zap.Logger) (EmailService, error) {
	funcMap := template.FuncMap{
		"unsubscribe_url": func() string {
			return fmt.Sprintf("%s/unsubscribe", getFrontendURL())
		},
		"suscripciónURL": func() string {
			return fmt.Sprintf("%s/unsubscribe", getFrontendURL())
		},
		"odd": func(index int) bool {
			return index%2 != 0
		},
		"account": func() map[string]interface{} {
			// Company information for Tucanbit.io
			return map[string]interface{}{
				"company_name": "3-102-940901 SRL",
				"address":      "Avenues Eight and Ten, Street Thirty-Nine, LY Center",
				"city":         "San Pedro",
				"region":       "Montes De Oca",
				"postal_code":  "11501",
				"country":      "Costa Rica",
				"license_info": "Licensed and regulated by the Government of the Autonomous Island of Anjouan, Union of Comoros under License No. ALSI-202509055-FI2",
			}
		},
		"default": func(value interface{}, defaultValue interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	}

	// Load email templates from files with custom functions
	// ParseGlob creates templates with names based on the base filename (without extension)
	templates, err := template.New("").Funcs(funcMap).ParseGlob("templates/emails/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	// Debug: List all loaded templates
	var loadedTemplateNames []string
	for _, t := range templates.Templates() {
		if t.Name() != "" {
			loadedTemplateNames = append(loadedTemplateNames, t.Name())
		}
	}
	logger.Info("Loaded email templates", zap.Strings("template_names", loadedTemplateNames))

	// Verify required templates exist and load them manually if needed
	requiredTemplates := []string{"verification", "welcome", "password_reset", "password_reset_confirmation", "security_alert", "two_factor_otp"}
	for _, tmplName := range requiredTemplates {
		if templates.Lookup(tmplName) == nil {
			// Try to manually load the template if it wasn't found
			templatePath := fmt.Sprintf("templates/emails/%s.html", tmplName)
			templateContent, err := os.ReadFile(templatePath)
			if err != nil {
				return nil, fmt.Errorf("required template '%s' not found in templates/emails/ and file read failed: %w", tmplName, err)
			}
			// Parse the template with the correct name
			templates, err = templates.New(tmplName).Parse(string(templateContent))
			if err != nil {
				return nil, fmt.Errorf("failed to parse template '%s': %w", tmplName, err)
			}
			logger.Info("Manually loaded template", zap.String("template_name", tmplName))
		}
	}

	if templates.Lookup("password_reset_otp") == nil {
		passwordResetTmpl := templates.Lookup("password_reset")
		if passwordResetTmpl == nil {
			return nil, fmt.Errorf("required template 'password_reset' not found in templates/emails/")
		}
		passwordResetFile, err := os.ReadFile("templates/emails/password_reset.html")
		if err == nil {
			templates = template.Must(templates.New("password_reset_otp").Parse(string(passwordResetFile)))
		}
	}

	if templates.Lookup("admin_generated_password") == nil {
		adminPwdFile, err := os.ReadFile("templates/emails/admin_generated_password.html")
		if err == nil {
			templates = template.Must(templates.New("admin_generated_password").Parse(string(adminPwdFile)))
		} else {
			passwordResetFile, err := os.ReadFile("templates/emails/password_reset.html")
			if err == nil {
				templates = template.Must(templates.New("admin_generated_password").Parse(string(passwordResetFile)))
			}
		}
	}

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

func (e *EmailServiceImpl) SendVerificationEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error {
	subject := "Welcome to TucanBIT - Verify your Account"
	templateData := EmailTemplateData{
		Email:            email,
		OTPCode:          otpCode,
		OTPId:            otpId,
		UserID:           userId,
		OTPExpiresAt:     expiresAt,
		VerificationLink: fmt.Sprintf("%s/verify?otp_code=%s&otp_id=%s&user_id=%s", getFrontendURL(), otpCode, otpId, userId),
		SupportEmail:     getSupportEmail(),
		WebsiteURL:       getWebsiteURL(),
		CurrentYear:      time.Now().Year(),
		LogoURL:          "cid:tucan.png",
		DeviceInfo:       GetDeviceInfo(userAgent),
		LocationInfo:     GetLocationInfo(ipAddress),
	}

	tmpl := GetModernVerificationEmailTemplate()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("failed to render modern verification template: %w", err)
	}
	htmlBody := buf.String()
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

func (e *EmailServiceImpl) SendWelcomeEmail(email, firstName string) error {
	subject := "Welcome to TucanBIT!"
	tmpl := GetWelcomeEmailTemplate()
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
		LoginURL:     fmt.Sprintf("%s/login", getFrontendURL()),
		SupportEmail: getSupportEmail(),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute welcome template: %w", err)
	}

	htmlBody := buf.String()
	return e.sendEmail(email, subject, htmlBody)
}

func (e *EmailServiceImpl) SendPasswordResetEmail(email, resetToken string, expiresAt time.Time) error {
	subject := "Reset Your Password - TucanBIT"

	data := map[string]interface{}{
		"ResetToken":   resetToken,
		"ExpiresAt":    expiresAt.Format("2006-01-02 15:04:05 UTC"),
		"Email":        email,
		"BrandName":    "TucanBIT",
		"ResetURL":     fmt.Sprintf("%s/reset-password?token=%s", getFrontendURL(), resetToken),
		"SupportEmail": "support@tucanbit.com",
	}

	htmlBody, err := e.renderTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}

	return e.sendEmail(email, subject, htmlBody)
}

func (e *EmailServiceImpl) SendPasswordResetOTPEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error {
	subject := "Reset Your Password - TucanBIT"

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
		"SupportEmail": os.Getenv("SUPPORT_EMAIL"),
		"ResetURL":     fmt.Sprintf("https://%s/reset-password?otp_code=%s&otp_id=%s&user_id=%s", os.Getenv("BACKEND_HOST"), otpCode, otpId, userId),
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

	subject := fmt.Sprintf("Security Alert: %s", config.Name)

	// Build detailed alert information
	var username string
	var userEmail string
	if triggerData.Username != nil {
		username = *triggerData.Username
	}
	if triggerData.UserEmail != nil {
		userEmail = *triggerData.UserEmail
	}

	// Build alert details with all relevant information
	alertDetails := fmt.Sprintf(`
Alert Name: %s
Alert Type: %s
Threshold: $%.2f
Triggered Value: $%.2f
Time Window: %d minutes
`, config.Name, config.AlertType, config.ThresholdAmount, triggerData.TriggerValue, config.TimeWindowMinutes)

	if username != "" {
		alertDetails += fmt.Sprintf("Username: %s\n", username)
	}
	if userEmail != "" {
		alertDetails += fmt.Sprintf("User Email: %s\n", userEmail)
	}
	if triggerData.UserID != nil {
		alertDetails += fmt.Sprintf("User ID: %s\n", triggerData.UserID.String())
	}
	if triggerData.AmountUSD != nil {
		alertDetails += fmt.Sprintf("Amount: $%.2f\n", *triggerData.AmountUSD)
	}
	if triggerData.CurrencyCode != nil {
		alertDetails += fmt.Sprintf("Currency: %s\n", string(*triggerData.CurrencyCode))
	}
	if triggerData.TransactionID != nil {
		alertDetails += fmt.Sprintf("Transaction ID: %s\n", *triggerData.TransactionID)
	}

	// Create template data
	data := map[string]interface{}{
		"AlertType":    config.Name,
		"Details":      alertDetails,
		"Email":        userEmail, // User's email (not recipient email)
		"Username":     username,
		"BrandName":    "TucanBIT",
		"Timestamp":    triggerData.TriggeredAt.UTC().Format("2006-01-02 15:04:05 UTC"),
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

// getFrontendURL gets the frontend URL from configuration
// It tries multiple config keys in order of preference
func getFrontendURL() string {
	// Try dedicated frontend URL config first
	if url := viper.GetString("app.frontend_url"); url != "" {
		return url
	}
	// Try extracting from oauth frontend handler URL
	if oauthURL := viper.GetString("oauth.frontend_oauth_handler_url"); oauthURL != "" {
		// Extract base URL (remove /oauth path)
		if idx := strings.LastIndex(oauthURL, "/oauth"); idx > 0 {
			return oauthURL[:idx]
		}
		return oauthURL
	}
	// Fallback to environment variable
	if url := os.Getenv("FRONTEND_URL"); url != "" {
		return url
	}
	// Default fallback
	return "https://app.tucanbit.com"
}

// getSupportEmail gets the support email from configuration
func getSupportEmail() string {
	if email := viper.GetString("email.support_email"); email != "" {
		return email
	}
	if email := os.Getenv("SUPPORT_EMAIL"); email != "" {
		return email
	}
	return "support@tucanbit.com"
}

// getWebsiteURL gets the website URL from configuration
func getWebsiteURL() string {
	if url := viper.GetString("app.website_url"); url != "" {
		return url
	}
	if url := os.Getenv("WEBSITE_URL"); url != "" {
		return url
	}
	return "https://tucanbit.com"
}

// Email templates are loaded from templates/emails/*.html files
