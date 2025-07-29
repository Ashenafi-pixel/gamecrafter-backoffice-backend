package user

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/joshjones612/egyptkingcrash/docs"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	_ "github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type user struct {
	userModule              module.User
	log                     *zap.Logger
	frontendOAuthHandlerURL string
}

func Init(userModule module.User, log *zap.Logger, frontendOauthHandlerURL string) handler.User {
	return &user{
		userModule:              userModule,
		log:                     log,
		frontendOAuthHandlerURL: frontendOauthHandlerURL,
	}
}

func (u *user) setSecureRefreshTokenCookie(c *gin.Context, refreshToken string, maxAge int) {
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", "", true, true)
}

// RegisterUser handles user register requests.
//
//	@Summary		Register
//	@Description	Register user using phone, username, and password,
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			userRequest	body		dto.User	true	"Register Request"
//	@Success		200			{object}	dto.UserRegisterResponse
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		409			{object}	response.ErrorResponse
//	@Router			/register [post]
func (u *user) RegisterUser(c *gin.Context) {
	var userRequest dto.User
	if err := c.ShouldBind(&userRequest); err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userRequest.CreatedBy = uuid.Nil
	userRequest.IsAdmin = false
	usr, refreshToken, err := u.userModule.RegisterUser(c, userRequest)
	if err != nil {
		_ = c.Error(err)
		return
	}
	u.setSecureRefreshTokenCookie(c, refreshToken, 1800)
	response.SendSuccessResponse(c, http.StatusCreated, usr)
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	file, header, err := c.Request.FormFile("picture")
	if err != nil {
		u.log.Error("Failed to retrieve file", zap.Error(err))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	defer file.Close()

	const maxFileSize = 8 * 1024 * 1024
	if header.Size > maxFileSize {
		err := errors.ErrInvalidUserInput.New("File size exceeds the 8 MB limit")
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var changePAsswordReq dto.ChangePasswordReq
	if err := c.ShouldBind(&changePAsswordReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.ForgetPassword(c, changePAsswordReq.EmailOrPhoneOrUserame)
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.VerifyResetPassword(c, verifyResetPasswordReq)
	if err != nil {
		_ = c.Error(err)
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	res, err := u.userModule.ResetPassword(c, resetPassword)
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var updateProfileReq dto.UpdateProfileReq
	if err := c.ShouldBind(&updateProfileReq); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	code := c.Query("code")
	if code == "" {
		err := fmt.Errorf("code not found")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	//get admin user id
	userID := c.GetString("user-id")
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		u.log.Warn(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var getIpFilterReq dto.GetIPFilterReq
	if err := c.ShouldBind(&getIpFilterReq); err != nil {
		u.log.Warn(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	pageParsed, err := strconv.Atoi(page)
	if err != nil {
		err := fmt.Errorf("unable to convert page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	perPageParsed, err := strconv.Atoi(perpage)
	if err != nil {
		err := fmt.Errorf("unable to convert per_page to number")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req []dto.MassReferralReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var userRequest dto.User
	if err := c.ShouldBind(&userRequest); err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		err = errors.ErrInvalidAccessToken.New("Refresh token missing or invalid")
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing or invalid"})
		return
	}
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
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
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp := u.userModule.GetOtp(c, req.PhoneNumber)

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
