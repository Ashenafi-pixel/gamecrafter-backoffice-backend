package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendVerificationEmail(email, otpCode, otpId, userId string, expiresAt time.Time, userAgent, ipAddress string) error
	SendWelcomeEmail(email, firstName string) error
	SendPasswordResetEmail(email, resetToken string, expiresAt time.Time) error
	SendSecurityAlert(email, alertType, details string) error
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
		templates = template.Must(templates.New("security_alert").Parse(securityAlertTemplate))
	}

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
		VerificationLink: fmt.Sprintf("http://localhost:8080/verify?otp_code=%s&otp_id=%s&user_id=%s", otpCode, otpId, userId),
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
		LoginURL:     "http://localhost:8080/login",
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

// sendEmail sends an email using SMTP with logo attachment
func (e *EmailServiceImpl) sendEmail(to, subject, htmlBody string) error {
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
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	if e.config.Port == 465 {
		// Port 465 requires SSL (not TLS)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         e.config.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to establish SSL connection to %s: %w", addr, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, e.config.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate with SMTP server: %w", err)
		}

		if err = client.Mail(e.config.From); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = writer.Write(message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		if err = writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %w", err)
		}
	} else if e.config.UseTLS {
		// Use STARTTLS for other ports (like 587)
		conn, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		if err = conn.StartTLS(&tls.Config{ServerName: e.config.Host}); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}

		if err = conn.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}

		if err = conn.Mail(e.config.From); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = conn.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		writer, err := conn.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}

		_, err = writer.Write(message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		if err = writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %w", err)
		}
	} else {
		// Use regular SMTP without TLS
		err := smtp.SendMail(addr, auth, e.config.From, []string{to}, message.Bytes())
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
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
