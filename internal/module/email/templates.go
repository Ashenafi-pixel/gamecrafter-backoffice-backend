package email

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailTemplateData represents the data structure for email templates
type EmailTemplateData struct {
	FirstName        string
	LastName         string
	Email            string
	UserType         string
	OTPCode          string
	OTPId            string
	UserID           string
	OTPExpiresAt     time.Time
	VerificationLink string
	CompanyName      string
	SupportEmail     string
	SupportPhone     string
	WebsiteURL       string
	CurrentYear      int
	LogoURL          string
	DeviceInfo       string
	LocationInfo     string
}

// GetVerificationEmailTemplate returns the HTML template for verification emails
func GetVerificationEmailTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email - TucanBIT</title>
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
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-decoration: none;
            padding: 15px 30px;
            border-radius: 25px;
            font-weight: 600;
            margin: 20px 0;
            transition: transform 0.3s ease;
        }
        .cta-button:hover {
            transform: translateY(-2px);
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
            <h1>TucanBIT</h1>
        </div>
        
        <div class="content">
            <div class="welcome">Welcome to TucanBIT, {{.FirstName}}!</div>
            
            <div class="message">
                Thank you for choosing TucanBIT for your {{.UserType}} account. To complete your registration and start using our platform, please verify your email address.
            </div>
            
            <div class="otp-container">
                <h3>Your Verification Code</h3>
                <div class="otp-code">{{.OTPCode}}</div>
                <p>Enter this code in the verification page to complete your registration.</p>
                <div class="otp-expires">
                    This code expires at {{.OTPExpiresAt.Format "3:04 PM MST on January 2, 2006"}}
                </div>
            </div>
            
            <div class="info-box">
                <h3>🔒 Security Notice</h3>
                <p>Never share this verification code with anyone. TucanBIT staff will never ask for your verification code.</p>
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
            <p>This email was sent to {{.Email}} for account verification purposes.</p>
        </div>
    </div>
</body>
</html>`

	return template.Must(template.New("verification").Parse(tmpl))
}

// GetPasswordResetEmailTemplate returns the HTML template for password reset emails
func GetPasswordResetEmailTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password - TucanBIT</title>
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
            background: linear-gradient(135deg, #dc3545 0%, #fd7e14 100%);
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
        .warning-box {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 20px;
            margin: 30px 0;
            border-radius: 5px;
        }
        .warning-box h3 {
            margin: 0 0 10px 0;
            color: #856404;
            font-size: 18px;
        }
        .warning-box p {
            margin: 0;
            color: #856404;
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
            <h1>Password Reset Request</h1>
        </div>
        
        <div class="content">
            <div class="message">
                Hi {{.FirstName}}, we received a request to reset your TucanBIT account password.
            </div>
            
            <div class="otp-container">
                <h3>Your Reset Code</h3>
                <div class="otp-code">{{.OTPCode}}</div>
                <p>Enter this code in the password reset page to create a new password.</p>
                <div class="otp-expires">
                    This code expires at {{.OTPExpiresAt.Format "3:04 PM MST on January 2, 2006"}}
                </div>
            </div>
            
            <div class="warning-box">
                <h3>⚠️ Security Notice</h3>
                <p>If you didn't request this password reset, please ignore this email and contact our support team immediately.</p>
            </div>
            
            <div style="text-align: center; margin-top: 30px;">
                <p>Need help? Contact us at {{.SupportEmail}}</p>
            </div>
        </div>
        
        <div class="footer">
            <p>&copy; {{.CurrentYear}} TucanBIT. All rights reserved.</p>
            <p>This email was sent to {{.Email}} for password reset purposes.</p>
        </div>
    </div>
</body>
</html>`

	return template.Must(template.New("password_reset").Parse(tmpl))
}

// GetPlainTextVerificationEmail returns a plain text version of the verification email
func GetPlainTextVerificationEmail(data EmailTemplateData) string {
	return fmt.Sprintf(`Welcome to TucanBIT, %s!

Thank you for choosing TucanBIT for your %s account. To complete your registration and start using our platform, please verify your email address.

Your Verification Code: %s

This code expires at %s

Enter this code in the verification page to complete your registration.

Security Notice:
- Never share this verification code with anyone
- TucanBIT staff will never ask for your verification code

Need Help?
Email: %s
Phone: %s
Website: %s

© %d TucanBIT. All rights reserved.

This email was sent to %s for account verification purposes.`,
		data.FirstName,
		data.UserType,
		data.OTPCode,
		data.OTPExpiresAt.Format("3:04 PM MST on January 2, 2006"),
		data.SupportEmail,
		data.SupportPhone,
		data.WebsiteURL,
		data.CurrentYear,
		data.Email)
}

// GetPlainTextWelcomeEmail returns a plain text version of the welcome email
func GetPlainTextWelcomeEmail(data EmailTemplateData) string {
	return fmt.Sprintf(`Congratulations, %s!

Your %s account has been successfully verified and activated. You're now ready to explore all the amazing features TucanBIT has to offer!

What's Next?
Complete your profile setup
💰 Explore our gaming platforms
🎮 Start playing and earning
📱 Download our mobile app

Get Started Now: %s

If you have any questions, our support team is here to help!
Email: %s
Phone: %s

© %d TucanBIT. All rights reserved.

This email was sent to %s to confirm your account activation.`,
		data.FirstName,
		data.UserType,
		data.WebsiteURL,
		data.SupportEmail,
		data.SupportPhone,
		data.CurrentYear,
		data.Email)
}

// GetPlainTextPasswordResetEmail returns a plain text version of the password reset email
func GetPlainTextPasswordResetEmail(data EmailTemplateData) string {
	return fmt.Sprintf(`Password Reset Request

Hi %s, we received a request to reset your TucanBIT account password.

Your Reset Code: %s

This code expires at %s

Enter this code in the password reset page to create a new password.

Security Notice:
If you didn't request this password reset, please ignore this email and contact our support team immediately.

Need help? Contact us at %s

© %d TucanBIT. All rights reserved.

This email was sent to %s for password reset purposes.`,
		data.FirstName,
		data.OTPCode,
		data.OTPExpiresAt.Format("3:04 PM MST on January 2, 2006"),
		data.SupportEmail,
		data.CurrentYear,
		data.Email)
}

// FormatPhoneNumber formats a phone number for display
func FormatPhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}

	// Remove any non-digit characters except +
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Ensure it starts with +
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}

	return cleaned
}

// GetDeviceInfo returns a user-friendly device description
func GetDeviceInfo(userAgent string) string {
	if userAgent == "" {
		return "Unknown Device"
	}

	// Simple device detection based on user agent
	if strings.Contains(strings.ToLower(userAgent), "mobile") || strings.Contains(strings.ToLower(userAgent), "android") || strings.Contains(strings.ToLower(userAgent), "iphone") {
		return "Mobile Device"
	} else if strings.Contains(strings.ToLower(userAgent), "windows") {
		return "Windows PC"
	} else if strings.Contains(strings.ToLower(userAgent), "mac") || strings.Contains(strings.ToLower(userAgent), "darwin") {
		return "Mac Computer"
	} else if strings.Contains(strings.ToLower(userAgent), "linux") {
		return "Linux Computer"
	} else if strings.Contains(strings.ToLower(userAgent), "tablet") || strings.Contains(strings.ToLower(userAgent), "ipad") {
		return "Tablet Device"
	}

	return "Desktop Computer"
}

// GetLocationInfo returns location information (can be enhanced with IP geolocation)
func GetLocationInfo(ipAddress string) string {
	if ipAddress == "" {
		return "Unknown Location"
	}

	// Simple IP-based location detection
	if ipAddress == "127.0.0.1" || ipAddress == "::1" || ipAddress == "localhost" {
		return "Local Development"
	}

	// You can enhance this with actual IP geolocation service
	// For now, return a generic location
	return "TucanBIT Platform"
}

// GetModernVerificationEmailTemplate returns a modern, dark-themed verification email template
// Updated: 2025-08-31 - Removed verification details card, centered everything, fixed button sizes
func GetModernVerificationEmailTemplate() *template.Template {
	return template.Must(template.New("modernVerification").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your TucanBIT Account</title>
    <style>
        /* Email client compatibility */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        /* Force table layout for better email client support */
        table {
            border-collapse: collapse;
            mso-table-lspace: 0pt;
            mso-table-rspace: 0pt;
        }
        
        /* Force image display */
        img {
            border: 0;
            height: auto;
            line-height: 100%;
            outline: none;
            text-decoration: none;
            -ms-interpolation-mode: bicubic;
        }
        
        /* Email container optimization */
        .email-container {
            max-width: 95% !important;
            width: 95% !important;
            margin: 0 auto !important;
        }
        
        /* Force full width in email clients */
        body {
            width: 100% !important;
            margin: 0 !important;
            padding: 0 !important;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #ffffff;
            line-height: 1.6;
            min-height: 100vh;
        }
        
        .email-container {
            max-width: 95%;
            width: 95%;
            margin: 0 auto;
            padding: 30px;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            border-radius: 20px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
            margin-top: 20px;
            margin-bottom: 20px;
        }
        
        .header {
            text-align: center;
            padding: 30px 0;
            border-bottom: 2px solid rgba(255, 255, 255, 0.1);
            margin-bottom: 30px;
        }
        
        .logo {
            display: inline-block;
            margin-bottom: 20px;
            vertical-align: middle;
        }
        
        .logo img {
            height: 60px;
            width: auto;
            border-radius: 12px;
            display: inline-block;
            vertical-align: middle;
        }
        
        .brand-name {
            font-size: 28px;
            font-weight: 700;
            color: #ffffff;
            letter-spacing: -1px;
            margin-left: 15px;
            display: inline-block;
            vertical-align: middle;
        }
        
        .main-heading {
            font-size: 24px;
            font-weight: 700;
            text-align: center;
            margin-bottom: 25px;
            color: #ffffff;
            text-shadow: 0 3px 6px rgba(0, 0, 0, 0.4);
        }
        

        
        .verification-info {
            text-align: center;
            margin-bottom: 25px;
        }
        
        .info-item {
            font-size: 14px;
            color: #e5e5e5;
            margin-bottom: 8px;
            font-weight: 500;
        }
        
        .instruction-text {
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 20px;
            color: #e5e5e5;
            text-align: center;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
        }
        
        .verification-code-container {
            text-align: center;
            margin: 30px 0;
        }
        
        .verification-code {
            display: inline-block;
            background: #ffffff;
            color: #333333;
            font-size: 32px;
            font-weight: 700;
            padding: 12px 24px;
            border-radius: 8px;
            letter-spacing: 4px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            border: 2px solid #e0e0e0;
            margin-bottom: 20px;
            width: 200px;
            text-align: center;
        }
        
        .verification-button {
            display: inline-block;
            background: linear-gradient(135deg, #ffffff 0%, #ff8c42 100%);
            color: #333333;
            text-decoration: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-weight: 600;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
            box-shadow: 0 4px 12px rgba(255, 140, 66, 0.3);
            transition: all 0.3s ease;
            border: 1px solid #ff8c42;
            width: 200px;
            text-align: center;
        }
        
        .verification-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 16px rgba(255, 140, 66, 0.4);
            background: linear-gradient(135deg, #ff8c42 0%, #ff6b35 100%);
            color: #ffffff;
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 30px;
            border-top: 2px solid rgba(255, 255, 255, 0.1);
            text-align: center;
        }
        
        .footer-logo {
            margin-bottom: 25px;
        }
        
        .footer-logo img {
            height: 60px;
            width: auto;
            border-radius: 10px;
            display: block;
            margin: 0 auto;
        }
        
        .social-media {
            margin-bottom: 30px;
        }
        
        .social-icon {
            display: inline-block;
            margin: 0 10px;
            transition: transform 0.3s ease;
        }
        
        .social-icon:hover {
            transform: scale(1.15);
        }
        
        .social-icon img {
            width: 50px;
            height: 50px;
            border-radius: 50%;
            display: block;
        }
        
        .supported-currencies {
            margin-bottom: 30px;
        }
        
        .currencies-title {
            font-size: 15px;
            color: #b0b0b0;
            margin-bottom: 20px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            font-weight: 600;
        }
        
        .currency-icons {
            display: flex;
            justify-content: center;
            gap: 20px;
            flex-wrap: wrap;
        }
        
        .currency-icon {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.12);
            display: flex;
            align-items: center;
            justify-content: center;
            transition: transform 0.3s ease;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        
        .currency-icon:hover {
            transform: scale(1.15);
            background: rgba(255, 255, 255, 0.18);
        }
        
        .currency-icon img {
            width: 32px;
            height: 32px;
            display: block;
        }
        
        .compliance-badges {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-bottom: 25px;
            flex-wrap: wrap;
        }
        
        .compliance-badge {
            background: rgba(255, 255, 255, 0.12);
            border: 1px solid rgba(255, 255, 255, 0.25);
            border-radius: 10px;
            padding: 10px 16px;
            font-size: 13px;
            color: #b0b0b0;
            text-align: center;
            font-weight: 500;
        }
        
        .legal-text {
            font-size: 12px;
            color: #909090;
            line-height: 1.6;
            max-width: 600px;
            margin: 0 auto;
        }
        
        @media (max-width: 800px) {
            .email-container {
                margin: 10px;
                padding: 20px;
                max-width: 98%;
                width: 98%;
            }
            
            .verification-code {
                font-size: 28px;
                padding: 16px 32px;
                letter-spacing: 3px;
            }
            
            .main-heading {
                font-size: 20px;
            }
            
            .brand-name {
                font-size: 24px;
            }
            
            .detail-row {
                flex-direction: column;
                text-align: center;
                gap: 5px;
            }
            
            .verification-button {
                padding: 10px 20px;
                font-size: 12px;
            }
            
            .social-icon img {
                width: 35px;
                height: 35px;
            }
            
            .currency-icon img {
                width: 24px;
                height: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <div class="logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
                <span class="brand-name">TucanBIT</span>
            </div>
        </div>
        
        <h1 class="main-heading">Verify Your TucanBIT Account</h1>
        
        <div class="verification-info">
            <div class="info-item">When: {{.OTPExpiresAt.Format "Jan 2, 2006, 3:04 PM UTC"}}</div>
            <div class="info-item">Device: {{.DeviceInfo}}</div>
            <div class="info-item">Location: {{.LocationInfo}}</div>
        </div>
        
        <p class="instruction-text">
            We've received a request to verify your TucanBIT account. To complete the process and unlock your full access, please use the verification code below.
        </p>
        
        <p class="instruction-text">
            If you didn't initiate this verification, you can safely ignore this email.
        </p>
        
        <p class="instruction-text">
            To verify your account, enter the code below in the app or site:
        </p>
        
        		<div class="verification-code-container">
            <div class="verification-code">{{.OTPCode}}</div>
            <a href="http://localhost:8080/verify?otp_code={{.OTPCode}}&otp_id={{.OTPId}}&user_id={{.UserID}}" class="verification-button">Verify My Account</a>
        </div>
        
        <div style="text-align: center; margin: 20px 0; padding: 20px; background: rgba(255,255,255,0.05); border-radius: 10px;">
            <p style="color: #b0b0b0; font-size: 14px; margin: 0;">Or use this verification link:</p>
            <p style="color: #ffffff; font-size: 12px; margin: 5px 0; word-break: break-all;">http://localhost:8080/verify?otp_code={{.OTPCode}}&otp_id={{.OTPId}}&user_id={{.UserID}}</p>
        </div>
        
        <div class="footer">
            <div class="footer-logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
            </div>
            
            <div class="social-media">
                <a href="https://discord.gg/tucanbit" class="social-icon">
                    <img src="cid:discord.png" alt="Discord">
                </a>
                <a href="https://t.me/tucanbit" class="social-icon">
                    <img src="cid:telegram.png" alt="Telegram">
                </a>
                <a href="https://instagram.com/tucanbit" class="social-icon">
                    <img src="cid:instagram.png" alt="Instagram">
                </a>
                <a href="https://twitter.com/tucanbit" class="social-icon">
                    <img src="cid:twitter.png" alt="Twitter">
                </a>
            </div>
            
            <div class="supported-currencies">
                <div class="currencies-title">Supported Currencies</div>
                <div class="currency-icons">
                    <div class="currency-icon">
                        <img src="cid:bitcoin.png" alt="Bitcoin">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:ethereum.png" alt="Ethereum">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:tether.png" alt="Tether">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:ton.png" alt="TON">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:dollar.png" alt="USD">
                    </div>
                </div>
            </div>
            
            <div class="compliance-badges">
                <div class="compliance-badge">#1 Casino Platform</div>
                <div class="compliance-badge">BeGambleAware</div>
                <div class="compliance-badge">18+</div>
            </div>
            
            <p class="legal-text">
                TucanBIT is owned and operated by TucanBIT Entertainment Ltd. and holds all necessary licenses for operation. 
                Please gamble responsibly and ensure you are of legal age in your jurisdiction.
            </p>
        </div>
    </div>
</body>
</html>
`))
}

// GetWelcomeEmailTemplate returns a welcome email template with TucanBIT branding
func GetWelcomeEmailTemplate() *template.Template {
	return template.Must(template.New("welcomeEmail").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to TucanBIT!</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #ffffff;
            line-height: 1.6;
            min-height: 100vh;
        }
        
        .email-container {
            max-width: 95%;
            margin: 0 auto;
            padding: 20px;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
        }
        
        .header {
            text-align: center;
            padding: 30px 20px;
            background: linear-gradient(135deg, #ff6b35 0%, #f7931e 100%);
            border-radius: 15px 15px 0 0;
            margin: -20px -20px 30px -20px;
            position: relative;
            overflow: hidden;
        }
        
        .header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: url('cid:tucan.png') no-repeat center center;
            background-size: 60px 60px;
            opacity: 0.1;
        }
        
        .logo {
            display: inline-block;
            margin-bottom: 15px;
        }
        
        .logo img {
            width: 60px;
            height: 60px;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.1);
            padding: 10px;
        }
        
        .brand-name {
            font-size: 32px;
            font-weight: 700;
            color: #ffffff;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
            margin-bottom: 10px;
        }
        
        .welcome-title {
            font-size: 24px;
            font-weight: 600;
            color: #ffffff;
            margin-bottom: 5px;
        }
        
        .content {
            padding: 0 20px 30px 20px;
        }
        
        .greeting {
            font-size: 28px;
            font-weight: 600;
            color: #ff6b35;
            text-align: center;
            margin-bottom: 30px;
            text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.3);
        }
        
        .welcome-message {
            font-size: 18px;
            color: #e5e5e5;
            text-align: center;
            margin-bottom: 40px;
            line-height: 1.8;
        }
        
        .features-container {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 40px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .features-title {
            font-size: 22px;
            font-weight: 600;
            color: #ff6b35;
            text-align: center;
            margin-bottom: 25px;
            text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.3);
        }
        
        .feature-list {
            list-style: none;
            padding: 0;
        }
        
        .feature-item {
            display: flex;
            align-items: center;
            margin-bottom: 20px;
            padding: 15px;
            background: rgba(255, 255, 255, 0.03);
            border-radius: 10px;
            border-left: 4px solid #ff6b35;
            transition: all 0.3s ease;
        }
        
        .feature-item:hover {
            background: rgba(255, 255, 255, 0.08);
            transform: translateX(5px);
        }
        
        .feature-icon {
            font-size: 24px;
            color: #ff6b35;
            margin-right: 15px;
            min-width: 30px;
        }
        
        .feature-text {
            font-size: 16px;
            color: #e5e5e5;
            font-weight: 500;
        }
        
        .cta-section {
            text-align: center;
            margin-bottom: 40px;
        }
        
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #ff6b35 0%, #f7931e 100%);
            color: #ffffff;
            text-decoration: none;
            padding: 18px 40px;
            border-radius: 50px;
            font-size: 18px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 1px;
            box-shadow: 0 8px 25px rgba(255, 107, 53, 0.3);
            transition: all 0.3s ease;
            border: none;
            cursor: pointer;
        }
        
        .cta-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 12px 35px rgba(255, 107, 53, 0.4);
            background: linear-gradient(135deg, #f7931e 0%, #ff6b35 100%);
        }
        
        .footer {
            text-align: center;
            padding: 30px 20px;
            background: rgba(0, 0, 0, 0.2);
            border-radius: 0 0 15px 15px;
            margin: 0 -20px -20px -20px;
        }
        
        .footer-logo {
            margin-bottom: 20px;
        }
        
        .footer-logo img {
            width: 50px;
            height: 50px;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.1);
            padding: 8px;
        }
        
        .social-media {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-bottom: 25px;
        }
        
        .social-icon {
            display: inline-block;
            width: 45px;
            height: 45px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 50%;
            text-align: center;
            line-height: 45px;
            color: #ffffff;
            text-decoration: none;
            transition: all 0.3s ease;
        }
        
        .social-icon:hover {
            background: #ff6b35;
            transform: scale(1.1);
        }
        
        .social-icon img {
            width: 25px;
            height: 25px;
            vertical-align: middle;
        }
        
        .supported-currencies {
            margin-bottom: 25px;
        }
        
        .currencies-title {
            font-size: 16px;
            color: #b0b0b0;
            margin-bottom: 15px;
            font-weight: 500;
        }
        
        .currency-icons {
            display: flex;
            justify-content: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        
        .currency-icon {
            width: 40px;
            height: 40px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
        }
        
        .currency-icon:hover {
            background: rgba(255, 255, 255, 0.2);
            transform: scale(1.1);
        }
        
        .currency-icon img {
            width: 25px;
            height: 25px;
        }
        
        .compliance-badges {
            display: flex;
            justify-content: center;
            gap: 15px;
            margin-bottom: 25px;
            flex-wrap: wrap;
        }
        
        .compliance-badge {
            background: rgba(255, 107, 53, 0.2);
            color: #ff6b35;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            border: 1px solid rgba(255, 107, 53, 0.3);
        }
        
        .legal-text {
            font-size: 12px;
            color: #888;
            line-height: 1.6;
            max-width: 600px;
            margin: 0 auto;
        }
        
        .contact-info {
            margin-top: 20px;
            padding: 20px;
            background: rgba(255, 255, 255, 0.03);
            border-radius: 10px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .contact-title {
            font-size: 18px;
            color: #ff6b35;
            text-align: center;
            margin-bottom: 15px;
            font-weight: 600;
        }
        
        .contact-item {
            display: flex;
            align-items: center;
            justify-content: center;
            margin-bottom: 10px;
            color: #e5e5e5;
        }
        
        .contact-icon {
            margin-right: 10px;
            color: #ff6b35;
        }
        
        /* Responsive Design */
        @media (max-width: 768px) {
            .email-container {
                max-width: 100%;
                padding: 15px;
            }
            
            .header {
                padding: 20px 15px;
                margin: -15px -15px 20px -15px;
            }
            
            .brand-name {
                font-size: 28px;
            }
            
            .greeting {
                font-size: 24px;
            }
            
            .welcome-message {
                font-size: 16px;
            }
            
            .features-container {
                padding: 20px;
            }
            
            .cta-button {
                padding: 15px 30px;
                font-size: 16px;
            }
            
            .social-media {
                gap: 15px;
            }
            
            .currency-icons {
                gap: 10px;
            }
        }
        
        @media (max-width: 480px) {
            .brand-name {
                font-size: 24px;
            }
            
            .greeting {
                font-size: 20px;
            }
            
            .feature-item {
                padding: 12px;
            }
            
            .cta-button {
                padding: 12px 25px;
                font-size: 14px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <div class="logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
            </div>
            <div class="brand-name">TucanBIT</div>
            <div class="welcome-title">Welcome to the Future of Gaming!</div>
        </div>
        
        <div class="content">
            <div class="greeting">Hello {{if .FirstName}}{{.FirstName}}{{else}}there{{end}}!</div>
            
            <div class="welcome-message">
                Welcome to TucanBIT! Your account has been successfully created and verified. 
                You're now part of the most exciting online casino platform in the world.
            </div>
            
            <div class="features-container">
                <div class="features-title">🎮 What You Can Do Now</div>
                <ul class="feature-list">
                    <li class="feature-item">
                        <span class="feature-icon">🎯</span>
                        <span class="feature-text">Access all platform features and games</span>
                    </li>
                    <li class="feature-item">
                        <span class="feature-icon">🎲</span>
                        <span class="feature-text">Start playing your favorite casino games</span>
                    </li>
                    <li class="feature-item">
                        <span class="feature-icon">👤</span>
                        <span class="feature-text">Manage your profile and preferences</span>
                    </li>
                    <li class="feature-item">
                        <span class="feature-icon">🤝</span>
                        <span class="feature-text">Connect with other players worldwide</span>
                    </li>
                    <li class="feature-item">
                        <span class="feature-icon">💰</span>
                        <span class="feature-text">Enjoy secure transactions and bonuses</span>
                    </li>
                    <li class="feature-item">
                        <span class="feature-icon">🏆</span>
                        <span class="feature-text">Participate in tournaments and events</span>
                    </li>
                </ul>
            </div>
            
            <div class="cta-section">
                <a href="http://localhost:8080/login" class="cta-button">Start Playing Now</a>
            </div>
            
            <div class="contact-info">
                <div class="contact-title">Need Help?</div>
                <div class="contact-item">
                    <span class="contact-icon">📧</span>
                    <span>Email: support@tucanbit.com</span>
                </div>
                <div class="contact-item">
                    <span class="contact-icon">🌐</span>
                    <span>Website: https://app.tucanbit.com</span>
                </div>
            </div>
        </div>
        
        <div class="footer">
            <div class="footer-logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
            </div>
            
            <div class="social-media">
                <a href="https://discord.gg/tucanbit" class="social-icon">
                    <img src="cid:discord.png" alt="Discord">
                </a>
                <a href="https://t.me/tucanbit" class="social-icon">
                    <img src="cid:telegram.png" alt="Telegram">
                </a>
                <a href="https://instagram.com/tucanbit" class="social-icon">
                    <img src="cid:instagram.png" alt="Instagram">
                </a>
                <a href="https://twitter.com/tucanbit" class="social-icon">
                    <img src="cid:twitter.png" alt="Twitter">
                </a>
            </div>
            
            <div class="supported-currencies">
                <div class="currencies-title">Supported Currencies</div>
                <div class="currency-icons">
                    <div class="currency-icon">
                        <img src="cid:bitcoin.png" alt="Bitcoin">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:ethereum.png" alt="Ethereum">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:tether.png" alt="Tether">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:ton.png" alt="TON">
                    </div>
                    <div class="currency-icon">
                        <img src="cid:dollar.png" alt="USD">
                    </div>
                </div>
            </div>
            
            <div class="compliance-badges">
                <div class="compliance-badge">#1 Casino Platform</div>
                <div class="compliance-badge">BeGambleAware</div>
                <div class="compliance-badge">18+</div>
            </div>
            
            <p class="legal-text">
                TucanBIT is owned and operated by TucanBIT Entertainment Ltd. and holds all necessary licenses for operation. 
                Please gamble responsibly and ensure you are of legal age in your jurisdiction.
            </p>
        </div>
    </div>
</body>
</html>
`))
}

// GetVerificationSuccessTemplate returns a verification success page template
func GetVerificationSuccessTemplate() *template.Template {
	return template.Must(template.New("verificationSuccess").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verification Successful - TucanBIT</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #ffffff;
            line-height: 1.6;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .success-container {
            max-width: 600px;
            margin: 0 auto;
            padding: 40px;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            border-radius: 20px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
            text-align: center;
        }
        
        .logo {
            margin-bottom: 30px;
        }
        
        .logo img {
            height: 80px;
            width: auto;
            border-radius: 15px;
        }
        
        .brand-name {
            font-size: 36px;
            font-weight: 800;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            letter-spacing: -1px;
            margin-bottom: 20px;
        }
        
        .success-icon {
            font-size: 80px;
            margin: 30px 0;
            color: #4ade80;
        }
        
        .main-heading {
            font-size: 32px;
            font-weight: 700;
            margin-bottom: 20px;
            color: #ffffff;
        }
        
        .success-message {
            font-size: 18px;
            color: #e5e5e5;
            margin-bottom: 30px;
            line-height: 1.6;
        }
        
        .login-button {
            display: inline-block;
            background: linear-gradient(135deg, #4ade80 0%, #22c55e 100%);
            color: #ffffff;
            text-decoration: none;
            padding: 18px 40px;
            border-radius: 15px;
            font-weight: 700;
            font-size: 18px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            box-shadow: 0 10px 30px rgba(74, 222, 128, 0.4);
            transition: all 0.3s ease;
            margin: 20px 10px;
        }
        
        .login-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 40px rgba(74, 222, 128, 0.5);
        }
        
        .home-button {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #ffffff;
            text-decoration: none;
            padding: 18px 40px;
            border-radius: 15px;
            font-weight: 700;
            font-size: 18px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            box-shadow: 0 10px 30px rgba(102, 126, 234, 0.4);
            transition: all 0.3s ease;
            margin: 20px 10px;
        }
        
        .home-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 40px rgba(102, 126, 234, 0.5);
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 30px;
            border-top: 2px solid rgba(255, 255, 255, 0.1);
        }
        
        .footer-logo {
            margin-bottom: 20px;
        }
        
        .footer-logo img {
            height: 40px;
            width: auto;
            border-radius: 10px;
        }
        
        .social-media {
            margin-bottom: 20px;
        }
        
        .social-icon {
            display: inline-block;
            margin: 0 8px;
            transition: transform 0.3s ease;
        }
        
        .social-icon:hover {
            transform: scale(1.1);
        }
        
        .social-icon img {
            width: 35px;
            height: 35px;
            border-radius: 50%;
        }
        
        @media (max-width: 600px) {
            .success-container {
                margin: 20px;
                padding: 30px;
            }
            
            .main-heading {
                font-size: 28px;
            }
            
            .brand-name {
                font-size: 32px;
            }
            
            .success-icon {
                font-size: 60px;
            }
        }
    </style>
</head>
<body>
    <div class="success-container">
        <div class="logo">
            <img src="cid:tucan.png" alt="TucanBIT Logo">
        </div>
        
        <div class="brand-name">TucanBIT</div>
        
        <div class="success-icon"></div>
        
        <h1 class="main-heading">Verification Successful!</h1>
        
        <p class="success-message">
            Congratulations! Your TucanBIT account has been successfully verified. 
            You now have full access to all our premium casino features and services.
        </p>
        
        <div>
            <a href="http://localhost:8080/login" class="login-button">Login Now</a>
            <a href="http://localhost:8080" class="home-button">Go to Home</a>
        </div>
        
        <div class="footer">
            <div class="footer-logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
            </div>
            
            <div class="social-media">
                <a href="https://discord.gg/tucanbit" class="social-icon">
                    <img src="cid:discord.png" alt="Discord">
                </a>
                <a href="https://t.me/tucanbit" class="social-icon">
                    <img src="cid:telegram.png" alt="Telegram">
                </a>
                <a href="https://instagram.com/tucanbit" class="social-icon">
                    <img src="cid:instagram.png" alt="Instagram">
                </a>
                <a href="https://twitter.com/tucanbit" class="social-icon">
                    <img src="cid:twitter.png" alt="Twitter">
                </a>
            </div>
        </div>
    </div>
</body>
</html>
`))
}

// GetTwoFactorOTPEmailTemplate returns the HTML template for 2FA OTP emails
func GetTwoFactorOTPEmailTemplate() *template.Template {
	tmpl := `
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

	return template.Must(template.New("two_factor_otp").Parse(tmpl))
}

// GetVerificationPageTemplate returns a verification page template that handles API calls
func GetVerificationPageTemplate() *template.Template {
	return template.Must(template.New("verificationPage").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verifying Your Account - TucanBIT</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #ffffff;
            line-height: 1.6;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .verification-container {
            max-width: 600px;
            margin: 0 auto;
            padding: 40px;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            border-radius: 20px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
            text-align: center;
        }
        
        .logo {
            margin-bottom: 30px;
        }
        
        .logo img {
            height: 80px;
            width: auto;
            border-radius: 15px;
        }
        
        .brand-name {
            font-size: 36px;
            font-weight: 800;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            letter-spacing: -1px;
            margin-bottom: 20px;
        }
        
        .loading-spinner {
            width: 80px;
            height: 80px;
            border: 8px solid rgba(255, 255, 255, 0.1);
            border-top: 8px solid #4ade80;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 30px auto;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .main-heading {
            font-size: 32px;
            font-weight: 700;
            margin-bottom: 20px;
            color: #ffffff;
        }
        
        .status-message {
            font-size: 18px;
            color: #e5e5e5;
            margin-bottom: 30px;
            line-height: 1.6;
        }
        
        .success-icon {
            font-size: 80px;
            margin: 30px 0;
            color: #4ade80;
            display: none;
        }
        
        .error-icon {
            font-size: 80px;
            margin: 30px 0;
            color: #ef4444;
            display: none;
        }
        
        .login-button {
            display: inline-block;
            background: linear-gradient(135deg, #4ade80 0%, #22c55e 100%);
            color: #ffffff;
            text-decoration: none;
            padding: 18px 40px;
            border-radius: 15px;
            font-weight: 700;
            font-size: 18px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            box-shadow: 0 10px 30px rgba(74, 222, 128, 0.4);
            transition: all 0.3s ease;
            margin: 20px 10px;
            display: none;
        }
        
        .login-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 40px rgba(74, 222, 128, 0.5);
        }
        
        .home-button {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #ffffff;
            text-decoration: none;
            padding: 18px 40px;
            border-radius: 15px;
            font-weight: 700;
            font-size: 18px;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            box-shadow: 0 10px 30px rgba(102, 126, 234, 0.4);
            transition: all 0.3s ease;
            margin: 20px 10px;
            display: none;
        }
        
        .home-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 40px rgba(102, 126, 234, 0.5);
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 30px;
            border-top: 2px solid rgba(255, 255, 255, 0.1);
        }
        
        .footer-logo {
            margin-bottom: 20px;
        }
        
        .footer-logo img {
            height: 40px;
            width: auto;
            border-radius: 10px;
        }
        
        .social-media {
            margin-bottom: 20px;
        }
        
        .social-icon {
            display: inline-block;
            margin: 0 8px;
            transition: transform 0.3s ease;
        }
        
        .social-icon:hover {
            transform: scale(1.1);
        }
        
        .social-icon img {
            width: 35px;
            height: 35px;
            border-radius: 50%;
        }
        
        @media (max-width: 600px) {
            .verification-container {
                margin: 20px;
                padding: 30px;
            }
            
            .main-heading {
                font-size: 28px;
            }
            
            .brand-name {
                font-size: 32px;
            }
            
            .loading-spinner {
                width: 60px;
                height: 60px;
                border-width: 6px;
            }
        }
    </style>
</head>
<body>
    <div class="verification-container">
        <div class="logo">
            <img src="cid:tucan.png" alt="TucanBIT Logo">
        </div>
        
        <div class="brand-name">TucanBIT</div>
        
        <div id="loading" class="loading-spinner"></div>
        <div id="success" class="success-icon"></div>
        <div id="error" class="error-icon"></div>
        
        <h1 class="main-heading" id="heading">Verifying Your Account...</h1>
        
        <p class="status-message" id="message">
            Please wait while we verify your account. This may take a few moments.
        </p>
        
        <div id="buttons" style="display: none;">
            <a href="http://localhost:8080/login" class="login-button">Login Now</a>
            <a href="http://localhost:8080" class="home-button">Go to Home</a>
        </div>
        
        <div class="footer">
            <div class="footer-logo">
                <img src="cid:tucan.png" alt="TucanBIT Logo">
            </div>
            
            <div class="social-media">
                <a href="https://discord.gg/tucanbit" class="social-icon">
                    <img src="cid:discord.png" alt="Discord">
                </a>
                <a href="https://t.me/tucanbit" class="social-icon">
                    <img src="cid:telegram.png" alt="Telegram">
                </a>
                <a href="https://instagram.com/tucanbit" class="social-icon">
                    <img src="cid:instagram.png" alt="Instagram">
                </a>
                <a href="https://twitter.com/tucanbit" class="social-icon">
                    <img src="cid:twitter.png" alt="Twitter">
                </a>
            </div>
        </div>
    </div>

    <script>
        // Get URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const otpCode = urlParams.get('otp_code');
        const otpId = urlParams.get('otp_id');
        const userId = urlParams.get('user_id');
        
        // Verify account on page load
        window.addEventListener('DOMContentLoaded', async function() {
                    // Validate required parameters
        if (!otpCode || !otpId) {
            console.error('Missing required parameters:', { otpCode, otpId, userId });
            document.getElementById('loading').style.display = 'none';
            document.getElementById('error').style.display = 'block';
            document.getElementById('heading').textContent = 'Invalid Verification Link';
            document.getElementById('message').textContent = 'The verification link is missing required information. Please check your email for the correct verification link or contact support.';
            document.getElementById('buttons').style.display = 'block';
            return;
        }
        
        // If user_id is missing, we'll try to proceed without it and let the server handle it
        if (!userId) {
            console.warn('User ID is missing, proceeding with verification attempt');
        }
            
            try {
                console.log('Sending verification request:', {
                    otp_code: otpCode,
                    otp_id: otpId,
                    user_id: userId
                });
                
                const response = await fetch('/register/complete', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Accept': 'application/json'
                    },
                    body: JSON.stringify({
                        otp_code: otpCode,
                        otp_id: otpId,
                        user_id: userId
                    })
                });
                
                console.log('Response status:', response.status);
                
                if (response.ok) {
                    const result = await response.json();
                    console.log('Verification successful:', result);
                    
                    // Success
                    document.getElementById('loading').style.display = 'none';
                    document.getElementById('success').style.display = 'block';
                    document.getElementById('heading').textContent = 'Verification Successful!';
                    document.getElementById('message').textContent = 'Congratulations! Your TucanBIT account has been successfully verified. You now have full access to all our premium casino features and services.';
                    document.getElementById('buttons').style.display = 'block';
                } else {
                    const errorData = await response.json().catch(() => ({}));
                    console.error('Verification failed:', errorData);
                    throw new Error(errorData.message || 'Verification failed');
                }
            } catch (error) {
                console.error('Error during verification:', error);
                
                // Handle error
                document.getElementById('loading').style.display = 'none';
                document.getElementById('error').style.display = 'block';
                document.getElementById('heading').textContent = 'Verification Failed';
                document.getElementById('message').textContent = 'We encountered an issue while verifying your account. Please try again or contact support if the problem persists.';
                document.getElementById('buttons').style.display = 'block';
            }
        });
    </script>
</body>
</html>
`))
}
