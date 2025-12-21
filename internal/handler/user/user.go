package user

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	customErrors "github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/email"
	userModule "github.com/tucanbit/internal/module/user"
	"go.uber.org/zap"
)

// UserHandler defines the interface for user HTTP handlers
type UserHandler interface {
	RegisterUser(c *gin.Context)
	Login(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfilePicture(c *gin.Context)
	ChangePassword(c *gin.Context)
	ForgetPassword(c *gin.Context)
	VerifyResetPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ConfirmUpdateProfile(c *gin.Context)
	HandleGoogleOauthReq(c *gin.Context)
	HandleGoogleOauthRes(c *gin.Context)
	FacebookLoginReq(c *gin.Context)
	HandleFacebookOauthRes(c *gin.Context)
	BlockAccount(c *gin.Context)
	GetBlockedAccount(c *gin.Context)
	AddIpFilter(c *gin.Context)
	GetIpFilter(c *gin.Context)
	AdminUpdateProfile(c *gin.Context)
	AdminResetUsersPassword(c *gin.Context)
	AdminAutoResetUsersPassword(c *gin.Context)
	GetUsers(c *gin.Context)
	RemoveIPFilter(c *gin.Context)
	GetMyReferalCodes(c *gin.Context)
	GetMyRefferedUsersAndPoints(c *gin.Context)
	GetCurrentReferralMultiplier(c *gin.Context)
	UpdateReferralMultiplier(c *gin.Context)
	UpdateUsersPointsForReferrances(c *gin.Context)
	GetPlayerDetails(c *gin.Context)
	GetPlayerManualFunds(c *gin.Context)
	GetAdminAssignedPoints(c *gin.Context)
	GetUserPoints(c *gin.Context)
	AdminRegisterPlayer(c *gin.Context)
	AdminLogin(c *gin.Context)
	UpdateSignupBonus(c *gin.Context)
	GetSignupBonus(c *gin.Context)
	UpdateReferralBonus(c *gin.Context)
	GetReferralBonus(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	VerifyUser(c *gin.Context)
	ReSendVerificationOTP(c *gin.Context)
	GetOtp(c *gin.Context)
	GetAdmins(c *gin.Context)
	GetAdminUsers(c *gin.Context)
	CreateAdminUser(c *gin.Context)
	UpdateAdminUser(c *gin.Context)
	DeleteAdminUser(c *gin.Context)
	SuspendAdminUser(c *gin.Context)
	UnsuspendAdminUser(c *gin.Context)
	// Enterprise Registration Methods
	InitiateEnterpriseRegistration(c *gin.Context)
	CompleteEnterpriseRegistration(c *gin.Context)
	GetEnterpriseRegistrationStatus(c *gin.Context)
	ResendEnterpriseVerificationEmail(c *gin.Context)
	// Regular Registration with Email Verification Methods
	InitiateUserRegistration(c *gin.Context)
	CompleteUserRegistration(c *gin.Context)
	ResendVerificationEmail(c *gin.Context)
	ServeVerificationPage(c *gin.Context)
	// Service Management
	SetRegistrationService(service interface{})
}

type user struct {
	userModule              module.User
	log                     *zap.Logger
	frontendOAuthHandlerURL string
	registrationService     interface{}
}

func Init(userModule module.User, log *zap.Logger, frontendOauthHandlerURL string) UserHandler {
	return &user{
		userModule:              userModule,
		log:                     log,
		frontendOAuthHandlerURL: frontendOauthHandlerURL,
	}
}

// SetRegistrationService sets the registration service for email verification
func (u *user) SetRegistrationService(service interface{}) {
	u.registrationService = service
}

func (u *user) setSecureRefreshTokenCookie(c *gin.Context, refreshToken string, maxAge int) {
	// In development, use secure=false for HTTP localhost
	secure := c.Request.TLS != nil // Only secure if using HTTPS
	u.log.Info("Setting refresh token cookie",
		zap.String("domain", c.Request.Host),
		zap.Bool("secure", secure),
		zap.Int("maxAge", maxAge),
		zap.String("tokenPreview", func() string {
			if len(refreshToken) > 20 {
				return refreshToken[:20] + "..."
			}
			return refreshToken
		}()))
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", "", secure, true)
}

// Login handles user login requests.
//
//	@Summary		Login
//	@Description	Authenticate a user with username_or_phone and password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		dto.UserLoginReq	true	"Login Request"
//	@Success		200				{object}	dto.UserLoginRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/login [post]
func (u *user) Login(c *gin.Context) {
	var loginRequest dto.UserLoginReq
	userAgent := c.GetString("user-agent")
	ip := c.GetString("ip")

	requestInfo := dto.LoginAttempt{
		UserAgent: userAgent,
		IPAddress: ip,
	}
	if err := c.ShouldBind(&loginRequest); err != nil {
		u.log.Error(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	loginRes, refreshToken, err := u.userModule.Login(c, loginRequest, requestInfo, false)
	if err != nil {
		_ = c.Error(err)
		return
	}
	u.setSecureRefreshTokenCookie(c, refreshToken, 1800)
	response.SendSuccessResponse(c, http.StatusOK, loginRes)
}

// ServeVerificationPage serves the verification page HTML
func (u *user) ServeVerificationPage(c *gin.Context) {
	// Get the verification page template
	tmpl := email.GetVerificationPageTemplate()

	// Execute the template with empty data since the page will get data from URL params
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		u.log.Error("Failed to render verification page template", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load verification page"})
		return
	}

	// Set content type and serve the HTML
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

// GetProfile Get User profile information.
//
//	@Summary		GetProfile
//	@Description	Retrieve user information.
//	@Tags			User
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.UserProfile
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/user/profile [get]
func (u *user) GetProfile(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	profile, err := u.userModule.GetProfile(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, profile)
}

// UpdateProfilePicture Update user's profile picture.
//
//	@Summary		UpdateProfilePicture
//	@Description	Allows a user to upload and update their profile picture.
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			picture			formData	file	true	"Profile picture file (max size 8MB)"
//	@Success		200				{object}	string	"Profile picture URL"
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		413				{object}	response.ErrorResponse
//	@Router			/api/user/profile/picture [POST]
func (u *user) UpdateProfilePicture(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	file, header, err := c.Request.FormFile("picture")
	if err != nil {
		u.log.Error("Failed to retrieve file", zap.Error(err))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	defer file.Close()

	const maxFileSize = 8 * 1024 * 1024
	if header.Size > maxFileSize {
		err := customErrors.ErrInvalidUserInput.New("File size exceeds the 8 MB limit")
		u.log.Warn("File too large", zap.Int64("fileSize", header.Size))
		_ = c.Error(err)
		return
	}

	profile, err := u.userModule.UploadProfilePicture(c, file, header, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, profile)
}

// ChangePassword Update user's password.
//
//	@Summary		ChangePassword
//	@Description	Allows a user to update password.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string					true	"Bearer <token> "
//	@Param			changePAsswordReq	body		dto.ChangePasswordReq	true	"update password Request"
//	@Success		200					{object}	dto.ChangePasswordRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/user/password [PATCH]
func (u *user) ChangePassword(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var changePAsswordReq dto.ChangePasswordReq
	if err := c.ShouldBind(&changePAsswordReq); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	changePAsswordReq.UserID = userIDParsed
	successRes, err := u.userModule.ChangePassword(c, changePAsswordReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, successRes)
}

// ForgetPassword  user's password reset request.
//
//	@Summary		ForgetPassword
//	@Description	Allows a user to request reset  password.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			changePAsswordReq	body		dto.ForgetPasswordReq	true	"reset password Request"
//	@Success		200					{object}	dto.ForgetPasswordRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/user/password/forget [POST]
func (u *user) ForgetPassword(c *gin.Context) {
	var changePAsswordReq dto.ForgetPasswordReq
	if err := c.ShouldBind(&changePAsswordReq); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Extract user agent and IP address for email security features
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	res, err := u.userModule.ForgetPassword(c, changePAsswordReq.EmailOrPhoneOrUserame, userAgent, ipAddress)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// VerifyResetPassword  user's verify otp to reset password.
//
//	@Summary		VerifyResetPassword
//	@Description	Allows a user to verify otp to reset password.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			verifyResetPasswordReq	body		dto.VerifyResetPasswordReq	true	"verify otp Request"
//	@Success		200						{object}	dto.VerifyResetPasswordRes
//	@Failure		400						{object}	response.ErrorResponse
//	@Failure		401						{object}	response.ErrorResponse
//	@Router			/api/user/password/forget/verify [POST]
func (u *user) VerifyResetPassword(c *gin.Context) {
	var verifyResetPasswordReq dto.VerifyResetPasswordReq
	if err := c.ShouldBind(&verifyResetPasswordReq); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Extract user agent and IP address for email context
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	// Add to context for email service
	ctx := context.WithValue(c.Request.Context(), "user_agent", userAgent)
	ctx = context.WithValue(ctx, "ip_address", ipAddress)

	res, err := u.userModule.VerifyResetPassword(ctx, verifyResetPasswordReq)
	if err != nil {
		// Handle specific error types for better user experience
		errStr := err.Error()
		u.log.Error("Password reset OTP verification failed",
			zap.Error(err),
			zap.String("email", verifyResetPasswordReq.EmailOrPhoneOrUserame),
			zap.String("otp_id", verifyResetPasswordReq.OTPID.String()))

		// Check for OTP-related errors
		if strings.Contains(errStr, "invalid or expired OTP") ||
			strings.Contains(errStr, "invalid otp") ||
			strings.Contains(errStr, "OTP has expired") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Invalid or expired OTP. Please request a new password reset.",
			})
			return
		}

		// Check for user not found errors
		if strings.Contains(errStr, "invalid user") ||
			strings.Contains(errStr, "user not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "User not found. Please check your email address.",
			})
			return
		}

		// Check for access errors (invalid login information)
		if strings.Contains(errStr, "invalid login information") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Invalid email address. Please check your email and try again.",
			})
			return
		}

		// For any other errors, return a generic error without exposing internal details
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "An error occurred while verifying your OTP. Please try again later.",
		})
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// ResetPassword  user reset password.
//
//	@Summary		VerifyResetPassword
//	@Description	Allows a user to  reset password.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			resetPassword	body		dto.ResetPasswordReq	true	"reset password Request"
//	@Success		200				{object}	dto.ResetPasswordRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/user/password/reset [POST]
func (u *user) ResetPassword(c *gin.Context) {
	var resetPassword dto.ResetPasswordReq
	if err := c.ShouldBind(&resetPassword); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	// Extract user agent and IP address for email context
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	// Add to context for email service
	ctx := context.WithValue(c.Request.Context(), "user_agent", userAgent)
	ctx = context.WithValue(ctx, "ip_address", ipAddress)

	res, err := u.userModule.ResetPassword(ctx, resetPassword)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// UpdateProfile  user update profile.
//
//	@Summary		UpdateProfile
//	@Description	Allows a user to  update profile.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string					true	"Bearer <token> "
//	@Param			updateProfileReq	body		dto.UpdateProfileReq	true	"update profile Request"
//	@Success		200					{object}	dto.UpdateProfileReq
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/user/profile   [POST]
func (u *user) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var updateProfileReq dto.UpdateProfileReq
	if err := c.ShouldBind(&updateProfileReq); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	updateProfileReq.UserID = userIDParsed
	res, err := u.userModule.UpdateProfile(c, updateProfileReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// ConfirmUpdateProfile  user confirm update profile.
//
//	@Summary		ConfirmUpdateProfile
//	@Description	Allows a user to  update profile.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization			header		string						true	"Bearer <token> "
//	@Param			cofirmUpdateProfileReq	body		dto.ConfirmUpdateProfile	true	"confirm update profile Request"
//	@Success		200						{object}	dto.User
//	@Failure		400						{object}	response.ErrorResponse
//	@Failure		401						{object}	response.ErrorResponse
//	@Router			/api/user/profile/confirm   [POST]
func (u *user) ConfirmUpdateProfile(c *gin.Context) {
	var cofirmUpdateProfileReq dto.ConfirmUpdateProfile
	if err := c.ShouldBind(&cofirmUpdateProfileReq); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.ConfirmUpdateProfile(c, cofirmUpdateProfileReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// HandleGoogleOauthReq google oauth.
//
//	@Summary		HandleGoogleOauthReq
//	@Description	allow user to signin using google's oauth
//	@Tags			User
//	@Success		302	"Redirects to Google OAuth endpoint"
//	@Router			/user/oauth/google [get]
func (u *user) HandleGoogleOauthReq(c *gin.Context) {
	url := u.userModule.GoogleLoginReq(c)
	c.Redirect(http.StatusFound, url)
}

func (u *user) HandleGoogleOauthRes(c *gin.Context) {
	userAgent := c.GetString("user-agent")
	ip := c.GetString("ip")

	requestInfo := dto.LoginAttempt{
		UserAgent: userAgent,
		IPAddress: ip,
	}
	state := c.Query("state")
	if state != "randomstate" {
		err := fmt.Errorf("invalid state")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	code := c.Query("code")
	if code == "" {
		err := fmt.Errorf("code not found")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.GoogleLoginRes(c, code, requestInfo)
	if err != nil {
		_ = c.Error(err)
		return
	}

	//handle redirect to the frontend url
	url := fmt.Sprintf("%s?token=%s&user_id=%s", u.frontendOAuthHandlerURL, res.AccessToken, res.UserID.String())
	c.Redirect(http.StatusFound, url)
}

// FacebookLoginReq facebook oauth.
//
//	@Summary		FacebookLoginReq
//	@Description	allow user to signin using facebook's oauth
//	@Tags			User
//	@Success		302	"Redirects to Facebook OAuth endpoint"
//	@Router			/user/oauth/facebook [get]
func (u *user) FacebookLoginReq(c *gin.Context) {
	url := u.userModule.FacebookLoginReq(c)
	c.Redirect(http.StatusFound, url)
}

func (u *user) HandleFacebookOauthRes(c *gin.Context) {
	userAgent := c.GetString("user-agent")
	ip := c.GetString("ip")

	requestInfo := dto.LoginAttempt{
		UserAgent: userAgent,
		IPAddress: ip,
	}

	code := c.Query("code")
	if code == "" {
		err := fmt.Errorf("code not found")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.HandleFacebookOauthRes(c, code, requestInfo)
	if err != nil {
		_ = c.Error(err)
		return
	}

	//handle redirect to the frontend url
	url := fmt.Sprintf("%s?token=%s&user_id=%s", u.frontendOAuthHandlerURL, res.AccessToken, res.UserID.String())
	c.Redirect(http.StatusFound, url)
}

// BlockAccount  admin block account.
//
//	@Summary		BlockAccount
//	@Description	Allows a admins to  block account.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token> "
//	@Param			blockAccountReq	body		dto.AccountBlockReq	true	"block account request"
//	@Success		200				{object}	dto.AccountBlockRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/user/block   [POST]
func (u *user) BlockAccount(c *gin.Context) {
	var blockAccountReq dto.AccountBlockReq
	if err := c.ShouldBind(&blockAccountReq); err != nil {
		u.log.Warn(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	blockAccountReq.BlockedBy = userIDParsed
	res, err := u.userModule.BlockUser(c, blockAccountReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetBlockedAccount  admin block account.
//
//	@Summary		GetBlockedAccount
//	@Description	Allows a admins to  get blocked accounts.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization			header		string						true	"Bearer <token> "
//	@Param			getBlockedAccountLogReq	body		dto.GetBlockedAccountLogReq	true	"get blocked accounts request"
//	@Success		200						{object}	[]dto.GetBlockedAccountLogRep
//	@Failure		400						{object}	response.ErrorResponse
//	@Failure		401						{object}	response.ErrorResponse
//	@Router			/api/admin/user/block/accounts   [POST]
func (u *user) GetBlockedAccount(c *gin.Context) {
	var getBlockedAccountLogReq dto.GetBlockedAccountLogReq
	if err := c.ShouldBind(&getBlockedAccountLogReq); err != nil {
		u.log.Warn(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	getBlockedAccountLogReq.AdminID = userIDParsed

	res, err := u.userModule.GetBlockedAccount(c, getBlockedAccountLogReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AddIpFilter  admin add ip filter.
//
//	@Summary		AddIpFilter
//	@Description	Allows a admins to  add ip filter.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Bearer <token> "
//	@Param			addIpFilterReq	body		dto.IpFilterReq	true	"add ip filter"
//	@Success		200				{object}	dto.IPFilterRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/ipfilters   [POST]
func (u *user) AddIpFilter(c *gin.Context) {
	var addIpFilterReq dto.IpFilterReq
	if err := c.ShouldBind(&addIpFilterReq); err != nil {
		u.log.Warn(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	addIpFilterReq.CreatedBy = userIDParsed

	res, err := u.userModule.AddIpFilter(c, addIpFilterReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetIpFilter  admin get ip filter.
//
//	@Summary		GetIpFilter
//	@Description	Allows a admins to  get ip filter.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Param			page			query		string	true	"page type (required)"
//	@Param			per-page		query		string	true	"per-page type (required)"
//	@Param			type			query		string	false	"type (allow or deny) (optional)"
//	@Success		200				{object}	dto.GetIPFilterRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/ipfilters   [GET]
func (u *user) GetIpFilter(c *gin.Context) {
	page := c.Query("page")
	perpage := c.Query("per-page")
	filterType := c.Query("type")
	if perpage == "" || page == "" {
		err := fmt.Errorf("page and per_page query required")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var getIpFilterReq dto.GetIPFilterReq
	if err := c.ShouldBind(&getIpFilterReq); err != nil {
		u.log.Warn(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	pageParsed, err := strconv.Atoi(page)
	if err != nil {
		err := fmt.Errorf("unable to convert page to number")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	perPageParsed, err := strconv.Atoi(perpage)
	if err != nil {
		err := fmt.Errorf("unable to convert per_page to number")
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	getIpFilterReq.Page = pageParsed
	getIpFilterReq.PerPage = perPageParsed
	getIpFilterReq.Type = filterType

	res, err := u.userModule.GetIPFilters(c, getIpFilterReq)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AdminUpdateProfile  edit player information.
//
//	@Summary		AdminUpdateProfile
//	@Description	Allows a admins to edit player information.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			req				body		dto.EditProfileAdminReq	true	"edit player's information request"
//	@Param			Authorization	header		string					true	"Bearer <token> "
//	@Success		200				{object}	dto.EditProfileAdminRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users   [PATCH]
func (u *user) AdminUpdateProfile(c *gin.Context) {
	var req dto.EditProfileAdminReq
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.AdminID = userIDParsed
	res, err := u.userModule.AdminUpdateProfile(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AdminResetUsersPassword  reset player's password.
//
//	@Summary		AdminResetUsersPassword
//	@Description	Allows a admins to reset player information.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			req				body		dto.AdminResetPasswordReq	true	"reset  player's password request"
//	@Param			Authorization	header		string						true	"Bearer <token> "
//	@Success		200				{object}	dto.AdminResetPasswordRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users/password   [POST]
func (u *user) AdminResetUsersPassword(c *gin.Context) {
	var req dto.AdminResetPasswordReq
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.AdminID = userIDParsed
	res, err := u.userModule.AdminResetPassword(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// AdminAutoResetUsersPassword resets player's password with auto-generated password.
//
//	@Summary		AdminAutoResetUsersPassword
//	@Description	Allows admins to reset player's password with auto-generated password and send it via email.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			req				body		dto.AdminAutoResetPasswordReq	true	"auto reset player's password request"
//	@Param			Authorization	header		string								true	"Bearer <token> "
//	@Success		200				{object}	dto.AdminAutoResetPasswordRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users/password/auto-reset   [POST]
func (u *user) AdminAutoResetUsersPassword(c *gin.Context) {
	var req dto.AdminAutoResetPasswordReq
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.AdminID = userIDParsed
	res, err := u.userModule.AdminAutoResetPassword(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, res)
}

// GetUsers  allow get players.
//
//	@Summary		GetUsers
//	@Description	Allows a admins to get players.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			req				body		dto.GetPlayersReq	true	"get players req"
//	@Param			Authorization	header		string				true	"Bearer <token> "
//	@Success		200				{object}	dto.GetPlayersRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/users   [POST]
func (u *user) GetUsers(c *gin.Context) {
	var req dto.GetPlayersReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := u.userModule.GetPlayers(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RemoveIPFilter  admin remove ip filter.
//
//	@Summary		RemoveIPFilter
//	@Description	Allows a admins to  remove ip filter.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Bearer <token> "
//	@Param			req				body		dto.RemoveIPBlockReq	true	"remove ip filter request"
//	@Success		200				{object}	dto.RemoveIPBlockRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/ipfilters   [DELETE]
func (u *user) RemoveIPFilter(c *gin.Context) {
	var req dto.RemoveIPBlockReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.RemoveIPFilter(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetMyReferalCodes Get User referral code.
//
//	@Summary		GetMyReferalCodes
//	@Description	Retrieve user referral code.
//	@Tags			User
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	string
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/user/referral [get]
func (u *user) GetMyReferalCodes(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetMyReferralCode(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetMyRefferedUsersAndPoints Get User referred users.
//
//	@Summary		GetMyRefferedUsersAndPoints
//	@Description	Retrieve user reffered users and total points.
//	@Tags			User
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.MyRefferedUsers
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/user/referral/users [get]
func (u *user) GetMyRefferedUsersAndPoints(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetUserReferalUsersByUserID(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetCurrentReferralMultiplier allow admin to get the current multiplier for the referral.
//
//	@Summary		GetCurrentReferralMultiplier
//	@Description	Retrieve  referral multiplier
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.ReferalUpdateResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/referrals [get]
func (u *user) GetCurrentReferralMultiplier(c *gin.Context) {
	resp, err := u.userModule.GetReferalMultiplier(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateReferralMultiplier allow admin to update the current multiplier for the referral.
//
//	@Summary		UpdateReferralMultiplier
//	@Description	UpdateReferralMultiplier  referral multiplier
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Param			req	body		dto.UpdateReferralPointReq	true	"update multiplier request"
//	@Success		200	{object}	dto.ReferalUpdateResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/referrals [POST]
func (u *user) UpdateReferralMultiplier(c *gin.Context) {
	var req dto.UpdateReferralPointReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.UpdateReferalMultiplier(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateUsersPointsForReferrances allow admin to add points to  users .
//
//	@Summary		UpdateUsersPointsForReferrances
//	@Description	UpdateUsersPointsForReferrances  add point to list of users
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Param			req	body		dto.UpdateReferralPointReq	true	"update multiplier request"
//	@Success		200	{object}	dto.ReferalUpdateResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/referrals/users [POST]
func (u *user) UpdateUsersPointsForReferrances(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req []dto.MassReferralReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := u.userModule.UpdateUsersPointsForReferrances(c, userIDParsed, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetCurrentReferralMultiplier allow admin to get the current multiplier for the referral.
//
//	@Summary		GetCurrentReferralMultiplier
//	@Description	Retrieve  referral multiplier
//	@Tags			Admin
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per-page		query	string	true	"per-page type (required)"
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.ReferalUpdateResp
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/referrals [get]
func (u *user) GetAdminAssignedPoints(c *gin.Context) {
	var req dto.GetAdminAssignedPointsReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetAdminAssignedPoints(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetUserPoints  allow to get user points.
//
//	@Summary		GetUserPoints
//	@Description	GetUserPoints allow user  to get available points
//	@Tags			User
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetPointsResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/users/points [get]
func (u *user) GetUserPoints(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetUserPoints(c, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// AdminRegisterPlayer handles admin register players request .
//
//	@Summary		AdminRegisterPlayer
//	@Description	AdminRegisterPlayer  admin register player
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			userRequest	body		dto.User	true	"Register Request"
//	@Success		200			{object}	dto.UserRegisterResponse
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		409			{object}	response.ErrorResponse
//	@Router			/api/admin/players/register [post]
func (u *user) AdminRegisterPlayer(c *gin.Context) {
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var userRequest dto.User
	if err := c.ShouldBind(&userRequest); err != nil {
		u.log.Error(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userRequest.CreatedBy = userIDParsed
	userRequest.IsAdmin = false
	usr, _, err := u.userModule.AdminCreatePlayer(c, userRequest)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, usr)
}

// AdminLogin handles admin's login requests.
//
//	@Summary		AdminLogin
//	@Description	Authenticate a user with username_or_phone and password
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		dto.UserLoginReq	true	"admin login  Request"
//	@Success		200				{object}	dto.UserLoginRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/login [post]
func (u *user) AdminLogin(c *gin.Context) {
	var loginRequest dto.UserLoginReq
	userAgent := c.GetString("user-agent")
	ip := c.GetString("ip")

	requestInfo := dto.LoginAttempt{
		UserAgent: userAgent,
		IPAddress: ip,
	}
	if err := c.ShouldBind(&loginRequest); err != nil {
		u.log.Error(err.Error())
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	loginRes, refreshToken, err := u.userModule.AdminLogin(c, loginRequest, requestInfo)
	if err != nil {
		_ = c.Error(err)
		return
	}
	u.setSecureRefreshTokenCookie(c, refreshToken, 1800)
	response.SendSuccessResponse(c, http.StatusOK, loginRes)
}

// GetAdmins handles admin's login requests.
//
//	@Summary		GetAdmins
//	@Description	GetAdmins
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		dto.GetAdminsReq	true	"admin login  Request"
//	@Success		200				{object}	[]dto.Admin
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admins [get]
func (u *user) GetAdmins(c *gin.Context) {
	var req dto.GetAdminsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetAdmins(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetAdminUsers gets all admin users with is_admin=true and user_type='ADMIN'
//
//	@Summary		GetAdminUsers
//	@Description	Get all admin users with is_admin=true and user_type='ADMIN'
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"
//	@Param			per_page	query		int		false	"Items per page"
//	@Success		200			{object}	[]dto.Admin
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/admin/users [get]
func (u *user) GetAdminUsers(c *gin.Context) {
	var req dto.GetAdminsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.GetAllAdminUsers(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateAdminUser creates a new admin user
//
//	@Summary		CreateAdminUser
//	@Description	Create a new admin user with is_admin=true and user_type='ADMIN'
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreateAdminUserReq	true	"Create admin user request"
//	@Success		200		{object}	dto.Admin
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/users [post]
func (u *user) CreateAdminUser(c *gin.Context) {
	var req dto.CreateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.CreateAdminUser(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// UpdateAdminUser updates an existing admin user
//
//	@Summary		UpdateAdminUser
//	@Description	Update an existing admin user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			request	body		dto.UpdateAdminUserReq	true	"Update admin user request"
//	@Success		200		{object}	dto.Admin
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/users/{id} [put]
func (u *user) UpdateAdminUser(c *gin.Context) {
	userID := c.Param("id")
	var req dto.UpdateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.UpdateAdminUser(c, userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteAdminUser deletes an admin user
//
//	@Summary		DeleteAdminUser
//	@Description	Delete an admin user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/users/{id} [delete]
func (u *user) DeleteAdminUser(c *gin.Context) {
	userID := c.Param("id")
	err := u.userModule.DeleteAdminUser(c, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, map[string]string{"message": "Admin user deleted successfully"})
}

// SuspendAdminUser suspends an admin user
//
//	@Summary		SuspendAdminUser
//	@Description	Suspend an admin user to prevent login
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			request	body		dto.SuspendAdminUserReq	true	"Suspend admin user request"
//	@Success		200		{object}	dto.SuspendAdminUserRes
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admin/users_admin/{id}/suspend [post]
func (u *user) SuspendAdminUser(c *gin.Context) {
	userID := c.Param("id")
	var req dto.SuspendAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.SuspendAdminUser(c, userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UnsuspendAdminUser unsuspends an admin user
//
//	@Summary		UnsuspendAdminUser
//	@Description	Unsuspend an admin user to allow login
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	dto.SuspendAdminUserRes
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/users_admin/{id}/unsuspend [post]
func (u *user) UnsuspendAdminUser(c *gin.Context) {
	userID := c.Param("id")
	resp, err := u.userModule.UnsuspendAdminUser(c, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateSignupBonus updates the signup bonus for users.
//
//	@Summary		UpdateSignupBonus
//	@Description	Allows an admin to update the signup bonus for users.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token> "
//	@Param			signUpBonusReq	body		dto.SignUpBonusRes	true	"Update signup bonus request"
//	@Success		200				{object}	dto.SignUpBonusRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/signup/bonus [put]
func (u *user) UpdateSignupBonus(c *gin.Context) {
	var req dto.SignUpBonusReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.UpdateSignupBonus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSignupBonus retrieves the signup bonus configuration.
//
//	@Summary		GetSignupBonus
//	@Description	Retrieves the signup bonus configuration.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.SignUpBonusRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/signup/bonus [get]
func (u *user) GetSignupBonus(c *gin.Context) {
	resp, err := u.userModule.GetSignupBonusConfig(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateReferralBonus updates the referral bonus configuration.
//
//	@Summary		UpdateReferralBonus
//	@Description	Allows an admin to update the referral bonus configuration.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization		header		string					true	"Bearer <token> "
//	@Param			referralBonusReq	body		dto.ReferralBonusReq	true	"Update referral bonus request"
//	@Success		200					{object}	dto.ReferralBonusRes
//	@Failure		400					{object}	response.ErrorResponse
//	@Failure		401					{object}	response.ErrorResponse
//	@Router			/api/admin/referral/bonus/config [put]
func (u *user) UpdateReferralBonus(c *gin.Context) {
	var req dto.ReferralBonusReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := u.userModule.UpdateReferralBonus(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetReferralBonus retrieves the referral bonus configuration.
//
//	@Summary		GetReferralBonus
//	@Description	Retrieves the referral bonus configuration.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token> "
//	@Success		200				{object}	dto.ReferralBonusRes
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Router			/api/admin/referral/bonus/config [get]
func (u *user) GetReferralBonus(c *gin.Context) {
	resp, err := u.userModule.GetReferralBonusConfig(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RefreshToken issues a new access token and rotates the refresh token.
//
//	@Summary		Refresh Access Token
//	@Description	Issues a new access token and rotates the refresh token using the refresh token cookie.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"{ access_token: string }"
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/refresh [post]
func (u *user) RefreshToken(c *gin.Context) {
	// Debug: Log all cookies received
	u.log.Info("Refresh token request received",
		zap.String("allCookies", c.Request.Header.Get("Cookie")),
		zap.String("userAgent", c.Request.Header.Get("User-Agent")))

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		u.log.Warn("Refresh token missing",
			zap.Error(err),
			zap.String("allCookies", c.Request.Header.Get("Cookie")))
		err = customErrors.ErrInvalidAccessToken.New("Refresh token missing or invalid")
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing or invalid"})
		return
	}

	u.log.Info("Refresh token found", zap.String("tokenPreview", refreshToken[:20]+"..."))
	accessToken, newRefreshToken, newExpiry, err := u.userModule.RefreshTokenFlow(c, refreshToken)
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	// Set new refresh token cookie
	u.setSecureRefreshTokenCookie(c, newRefreshToken, int(newExpiry.Sub(time.Now()).Seconds()))
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// Logout handles user logout requests.
//
//	@Summary		Logout
//	@Description	Logout a user and invalidate their session
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/auth/logout [post]
func (u *user) Logout(c *gin.Context) {
	// Get user ID from the authenticated context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Invalidate all user sessions
	err := u.userModule.InvalidateAllUserSessions(c, userID)
	if err != nil {
		u.log.Error("Failed to invalidate user sessions", zap.Error(err), zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Clear refresh token cookie
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	u.log.Info("User logged out successfully", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// VerifyUser verifies a user's phone number.
//
//	@Summary		Verify User Phone Number
//	@Description	Verifies a user's phone number using a verification code.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.VerifyPhoneNumberReq	true	"Verification request"
//	@Success		200	{object}	dto.UserRegisterResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/user/verify [post]
func (u *user) VerifyUser(c *gin.Context) {
	var req dto.VerifyPhoneNumberReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, accessToken, err := u.userModule.VerifyUser(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	u.setSecureRefreshTokenCookie(c, accessToken, 1800)
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// ReSendVerificationOTP allows a user to resend the verification OTP.
//
//	@Summary		ReSendVerificationOTP
//	@Description	Allows a user to resend the verification OTP to their phone number.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.ResendVerificationOTPReq	true	"Re-send verification OTP request"
//	@Success		200	{object}	dto.ForgetPasswordRes
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/user/resend/verification/otp [post]
func (u *user) ReSendVerificationOTP(c *gin.Context) {
	var req dto.ResendVerificationOTPReq
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := u.userModule.ReSendVerificationOTP(c, req.PhoneNumber)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

func (u *user) GetOtp(c *gin.Context) {
	var req dto.TestOtp
	if err := c.ShouldBind(&req); err != nil {
		err = customErrors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp := u.userModule.GetOtp(c, req.PhoneNumber)

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// InitiateEnterpriseRegistration starts the enterprise registration process
func (u *user) InitiateEnterpriseRegistration(c *gin.Context) {
	var req dto.EnterpriseRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	u.log.Info("Initiating enterprise registration",
		zap.String("email", req.Email),
		zap.String("user_type", req.UserType),
		zap.String("ip", c.ClientIP()))

	response, err := u.userModule.(*userModule.User).EnterpriseRegistrationService.InitiateRegistration(c.Request.Context(), &req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		u.log.Error("Failed to initiate enterprise registration",
			zap.Error(err),
			zap.String("email", req.Email))

		// Debug: Log the exact error message to see what we're dealing with
		errStr := err.Error()
		u.log.Info("Error message for debugging", zap.String("error_message", errStr))

		// Handle specific business logic errors
		if strings.Contains(errStr, "user with email") && strings.Contains(errStr, "already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Code:    http.StatusConflict,
				Message: "An account with this email address already exists. Please use a different email or try logging in.",
			})
			return
		}

		// Handle specific database constraint violations
		// Check for wrapped database constraint errors
		if strings.Contains(errStr, "failed to create user") {
			// Extract the underlying database error
			if strings.Contains(errStr, "users_email_key") || strings.Contains(errStr, "duplicate key value violates unique constraint") && strings.Contains(errStr, "email") {
				c.JSON(http.StatusConflict, dto.ErrorResponse{
					Code:    http.StatusConflict,
					Message: "An account with this email address already exists. Please use a different email or try logging in.",
				})
				return
			}
			if strings.Contains(errStr, "users_phone_number_key") || strings.Contains(errStr, "duplicate key value violates unique constraint") && strings.Contains(errStr, "phone_number") {
				c.JSON(http.StatusConflict, dto.ErrorResponse{
					Code:    http.StatusConflict,
					Message: "An account with this phone number already exists. Please use a different phone number or try logging in.",
				})
				return
			}
			if strings.Contains(errStr, "users_username_key") || strings.Contains(errStr, "duplicate key value violates unique constraint") && strings.Contains(errStr, "username") {
				c.JSON(http.StatusConflict, dto.ErrorResponse{
					Code:    http.StatusConflict,
					Message: "This username is already taken. Please choose a different username.",
				})
				return
			}
		}

		// Handle validation errors
		if strings.Contains(errStr, "validation failed") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Please check your input data and try again.",
			})
			return
		}

		// Handle database connection errors
		if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "timeout") {
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
				Code:    http.StatusServiceUnavailable,
				Message: "Service temporarily unavailable. Please try again later.",
			})
			return
		}

		// Default error for unexpected issues
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration failed due to a system error. Please try again later or contact support.",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteEnterpriseRegistration completes the enterprise registration process
func (u *user) CompleteEnterpriseRegistration(c *gin.Context) {
	var req dto.EnterpriseRegistrationCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil || req.OTPCode == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "User ID and OTP code are required",
		})
		return
	}

	u.log.Info("Completing enterprise registration",
		zap.String("user_id", req.UserID.String()),
		zap.String("ip", c.ClientIP()))

	response, err := u.userModule.(*userModule.User).EnterpriseRegistrationService.CompleteRegistration(c.Request.Context(), &req)
	if err != nil {
		u.log.Error("Failed to complete enterprise registration",
			zap.Error(err),
			zap.String("user_id", req.UserID.String()))

		// Handle specific error cases
		if err.Error() == "invalid OTP code" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid OTP code",
			})
			return
		}
		if err.Error() == "OTP expired" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "OTP code has expired",
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

// GetEnterpriseRegistrationStatus gets the current registration status
func (u *user) GetEnterpriseRegistrationStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID format",
		})
		return
	}

	u.log.Info("Getting enterprise registration status",
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()))

	response, err := u.userModule.(*userModule.User).EnterpriseRegistrationService.GetRegistrationStatus(c.Request.Context(), userID)
	if err != nil {
		u.log.Error("Failed to get enterprise registration status",
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

	c.JSON(http.StatusOK, response)
}

// ResendEnterpriseVerificationEmail resends the verification email
func (u *user) ResendEnterpriseVerificationEmail(c *gin.Context) {
	var req dto.ResendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	u.log.Info("Resending enterprise verification email",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

	response, err := u.userModule.(*userModule.User).EnterpriseRegistrationService.ResendVerificationEmail(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		u.log.Error("Failed to resend enterprise verification email",
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
		if err.Error() == "user already verified" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "User is already verified",
			})
			return
		}
		if err.Error() == "too many resend attempts" {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Code:    http.StatusTooManyRequests,
				Message: "Too many resend attempts. Please try again later.",
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

// InitiateUserRegistration delegates to the registration service
func (u *user) InitiateUserRegistration(c *gin.Context) {
	u.log.Info("InitiateUserRegistration called")
	if u.registrationService == nil {
		u.log.Error("Registration service not initialized")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
		return
	}

	if service, ok := u.registrationService.(interface {
		InitiateUserRegistration(c *gin.Context)
	}); ok {
		service.InitiateUserRegistration(c)
	} else {
		u.log.Error("Registration service does not implement InitiateUserRegistration")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
	}
}

// CompleteUserRegistration delegates to the registration service
func (u *user) CompleteUserRegistration(c *gin.Context) {
	if u.registrationService == nil {
		u.log.Error("Registration service not initialized")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
		return
	}

	if service, ok := u.registrationService.(interface {
		CompleteUserRegistration(c *gin.Context)
	}); ok {
		service.CompleteUserRegistration(c)
	} else {
		u.log.Error("Registration service does not implement CompleteUserRegistration")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
	}
}

// ResendVerificationEmail delegates to the registration service
func (u *user) ResendVerificationEmail(c *gin.Context) {
	if u.registrationService == nil {
		u.log.Error("Registration service not initialized")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
		return
	}

	if service, ok := u.registrationService.(interface {
		ResendVerificationEmail(c *gin.Context)
	}); ok {
		service.ResendVerificationEmail(c)
	} else {
		u.log.Error("Registration service does not implement ResendVerificationEmail")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
	}
}

// RegisterUser handles user register requests.
func (u *user) RegisterUser(c *gin.Context) {
	// Delegate to the registration service for email verification flow
	if u.registrationService == nil {
		u.log.Error("Registration service not initialized")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
		return
	}

	if service, ok := u.registrationService.(interface {
		InitiateUserRegistration(c *gin.Context)
	}); ok {
		service.InitiateUserRegistration(c)
	} else {
		u.log.Error("Registration service does not implement InitiateUserRegistration")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Registration service not available",
		})
	}
}

// GetPlayerManualFunds - GET /api/admin/players/:user_id/manual-funds
// Get manual fund transactions for a specific player
func (u *user) GetPlayerManualFunds(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		u.log.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
		})
		return
	}

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Get manual fund transactions for the player
	manualFunds, totalCount, err := u.userModule.GetPlayerManualFundsPaginated(c.Request.Context(), userID, page, perPage)
	if err != nil {
		u.log.Error("Failed to get manual funds", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get manual funds",
		})
		return
	}

	totalPages := (totalCount + int64(perPage) - 1) / int64(perPage)

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Manual funds retrieved successfully",
		"data": gin.H{
			"manual_funds": manualFunds,
			"pagination": gin.H{
				"current_page": page,
				"per_page":     perPage,
				"total_pages":  totalPages,
				"total_count":  totalCount,
			},
		},
	})
}
