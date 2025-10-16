package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
	oauth "golang.org/x/oauth2"
	"google.golang.org/api/oauth2/v2"
)

func (u *User) GoogleLoginReq(ctx context.Context) string {
	return u.googleOauth.AuthCodeURL("randomstate", oauth.AccessTypeOffline)
}

func (u *User) GoogleLoginRes(ctx context.Context, code string, loginattempt dto.LoginAttempt) (dto.UserRegisterResponse, error) {
	var userName string
	var usr dto.User
	var exist bool
	// Exchange the code for a token
	token, err := u.googleOauth.Exchange(context.Background(), code)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	// Use the token to get user info
	client := u.googleOauth.Client(context.Background(), token)
	service, err := oauth2.New(client)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	// Get the user email from Google APIs
	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	//check if the user is already exist or not
	usr, exist, err = u.userStorage.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return dto.UserRegisterResponse{}, err
	}

	if !exist {
		//create user and return
		//generate random username

		for {
			userName = utils.GenerateRandomUsername(15)
			//check if the username is taken or not
			_, exist, err := u.userStorage.GetUserByUserName(ctx, userName)
			if err != nil {
				return dto.UserRegisterResponse{}, err
			}
			if exist {
				time.Sleep(time.Second)
				continue
			}
			break
		}

		//generate default password
		password := utils.GenerateRandomPassowrd(12)

		//register user
		usr = dto.User{
			FirstName: userName,
			Email:     userInfo.Email,
			Password:  password,
			Source:    constant.SOURCE_GMAIL,
		}
		response, err := u.OuathRegister(ctx, usr)
		if err != nil {
			return dto.UserRegisterResponse{}, err
		}
		return response, nil
	}
	//signin user

	response, err := u.OuathSignIn(ctx, usr, loginattempt)
	if err != nil {
		return dto.UserRegisterResponse{}, err
	}
	return response, nil
}

func (u *User) OuathRegister(ctx context.Context, userRequest dto.User) (dto.UserRegisterResponse, error) {
	var err error
	if userRequest.DefaultCurrency == "" {
		userRequest.DefaultCurrency = constant.DEFAULT_CURRENCY
	} else {
		if yes := dto.IsValidCurrency(userRequest.DefaultCurrency); !yes {
			userRequest.DefaultCurrency = constant.DEFAULT_CURRENCY
		}
	}

	hashPassword, err := utils.HashPassword(userRequest.Password)
	if err != nil {
		u.log.Error("unable to hash password", zap.Error(err), zap.Any("user", userRequest))
		err = errors.ErrInternalServerError.Wrap(err, "unable to hash password")
		return dto.UserRegisterResponse{}, err
	}

	//create user
	userRequest.Password = hashPassword
	usrRes, err := u.userStorage.CreateUser(ctx, userRequest)
	if err != nil {
		return dto.UserRegisterResponse{}, err
	}

	//create user balance
	u.balanceStorage.CreateBalance(ctx, dto.Balance{
		UserId:       usrRes.ID,
		CurrencyCode: constant.DEFAULT_CURRENCY,
		AmountUnits:    decimal.Zero,
		ReservedUnits:   decimal.Zero,
		ReservedCents:       0,
	})

	// generate jwt token  to the user
	token, err := utils.GenerateJWT(usrRes.ID)
	if err != nil {
		err = fmt.Errorf("unable to generate jwt token")
		u.log.Warn(err.Error(), zap.Any("registerReq", userRequest))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	// generate refresh token
	refreshToken, err := utils.GenerateRefreshJWT(usrRes.ID)
	if err != nil {
		u.log.Error("unable to generate refresh token", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}
	refreshTokenExpiry := time.Now().Add(30 * time.Minute)

	// save user session with refresh token
	userSession := dto.UserSessions{
		UserID:                usrRes.ID,
		Token:                 token,
		ExpiresAt:             time.Now().Add(time.Minute * 10),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiry,
		IpAddress:             "",
		UserAgent:             "",
	}
	u.logsStorage.CreateLoginSessions(ctx, userSession)

	return dto.UserRegisterResponse{
		Message:      constant.USER_REGISTRATION_SUCCESS,
		UserID:       usrRes.ID,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (u *User) OuathSignIn(ctx context.Context, usr dto.User, loginLogs dto.LoginAttempt) (dto.UserRegisterResponse, error) {
	// generate jwt token  to the user
	token, err := utils.GenerateJWT(usr.ID)
	if err != nil {
		err = fmt.Errorf("unable to generate jwt token")
		u.log.Warn(err.Error(), zap.Any("loginRequest", usr))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}
	u.logsStorage.CreateLoginAttempts(ctx, loginLogs)

	// generate refresh token
	refreshToken, err := utils.GenerateRefreshJWT(usr.ID)
	if err != nil {
		u.log.Error("unable to generate refresh token", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}
	refreshTokenExpiry := time.Now().Add(30 * time.Minute)

	// save  user session with refresh token
	userSession := dto.UserSessions{
		UserID:                usr.ID,
		Token:                 token,
		ExpiresAt:             time.Now().Add(time.Minute * 10),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiry,
		IpAddress:             loginLogs.IPAddress,
		UserAgent:             loginLogs.UserAgent,
	}
	u.logsStorage.CreateLoginSessions(ctx, userSession)

	return dto.UserRegisterResponse{
		Message:      constant.LOGIN_SUCCESS,
		AccessToken:  token,
		UserID:       usr.ID,
		RefreshToken: refreshToken,
	}, nil
}

func (u *User) FacebookLoginReq(ctx context.Context) string {
	return u.facebookOauth.AuthCodeURL("random", oauth.AccessTypeOffline)
}

func (u *User) HandleFacebookOauthRes(ctx context.Context, code string, loginattempt dto.LoginAttempt) (dto.UserRegisterResponse, error) {
	var usr dto.User
	var exist bool
	// Exchange the code for an access token
	token, err := u.facebookOauth.Exchange(ctx, code)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	// Fetch user information using the access token
	client := u.facebookOauth.Client(ctx, token)
	resp, err := client.Get(constant.FACEBOOK_OAUTH_RESPONSE_REQ)
	if err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	defer resp.Body.Close()
	var usrp dto.FacebookOauthRes
	if err := json.NewDecoder(resp.Body).Decode(&usrp); err != nil {
		u.log.Error(err.Error(), zap.Any("code", code))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, err
	}

	// register or signin user
	usr, exist, err = u.userStorage.GetUserByUserName(ctx, usrp.ID)
	if err != nil {
		return dto.UserRegisterResponse{}, err
	}

	//if not exist
	if !exist {
		//register user with facebookid as username
		password := utils.GenerateRandomPassowrd(12)

		//register user
		usr = dto.User{
			Email:    usrp.Email,
			Password: password,
			Source:   constant.SOURCE_FACEBOOK,
		}

		response, err := u.OuathRegister(ctx, usr)
		if err != nil {
			return dto.UserRegisterResponse{}, err
		}
		return response, nil
	}

	response, err := u.OuathSignIn(ctx, usr, loginattempt)
	if err != nil {
		return dto.UserRegisterResponse{}, err
	}
	return response, nil
}
