package user

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/errors/sqlcerr"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/otp"
	"github.com/tucanbit/platform/pisi"
	"github.com/tucanbit/platform/redis"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"

	"encoding/json"

	oauth "golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type User struct {
	userStorage                 storage.User
	logsStorage                 storage.Logs
	balanceStorage              storage.Balance
	bucketName                  string
	log                         *zap.Logger
	tmpOTP                      map[uuid.UUID]dto.OTPHolder
	tmpUpdateProfileOTP         map[uuid.UUID]dto.UpdateProfileTmpHolder
	mutex                       sync.Mutex
	jwtKey                      string
	locker                      map[uuid.UUID]*sync.Mutex
	googleOauth                 *oauth.Config
	facebookOauth               *oauth.Config
	IpFilterMap                 map[string][]dto.IPFilter
	operationalGroupStorage     storage.OperationalGroup
	operationalGroupTypeStorage storage.OperationalGroupType
	balanceLogStorage           storage.BalanceLogs
	cron                        *cron.Cron
	ConfigStorage               storage.Config
	UserBalanceSocket           map[uuid.UUID]map[*websocket.Conn]bool
	UserBalanceSocketLocker     map[*websocket.Conn]*sync.Mutex
	userWS                      utils.UserWS
	agentModule                 module.Agent
	// Session monitoring fields
	sessionSockets                map[uuid.UUID]map[*websocket.Conn]bool
	sessionSocketMutex            sync.RWMutex
	stopChan                      chan struct{}
	redis                         *redis.RedisOTP
	PisiClient                    pisi.PisiClient
	EnterpriseRegistrationService *EnterpriseRegistrationService
}

func Init(userStorage storage.User,
	logsStorage storage.Logs,
	log *zap.Logger,
	bucketName string,
	balanceStorage storage.Balance,
	jwtkey,
	gclientID,
	gclientSecret,
	gRedirectURL,
	fClientID,
	fclinetSecret,
	fRedirectURL string,
	balanceLogStorage storage.BalanceLogs,
	operationalGroupStorage storage.OperationalGroup,
	operationalGroupTypeStorage storage.OperationalGroupType,
	configStorage storage.Config,
	agentModule module.Agent,
	redis *redis.RedisOTP,
	pisiClient pisi.PisiClient,
	otpStorage otp.OTP,
	emailService email.EmailService,
) module.User {
	gOauth := &oauth.Config{
		ClientID:     gclientID,
		ClientSecret: gclientSecret,
		Scopes:       constant.GOOGLE_OAUTH_SCOPES,
		RedirectURL:  gRedirectURL,
		Endpoint:     google.Endpoint,
	}
	fOauth := &oauth.Config{
		ClientID:     fClientID,
		ClientSecret: fclinetSecret,
		RedirectURL:  fRedirectURL,
		Scopes:       constant.FACEBOOK_OAUTH_SCOPES,
		Endpoint:     facebook.Endpoint,
	}
	usr := &User{
		userStorage:                 userStorage,
		log:                         log,
		logsStorage:                 logsStorage,
		bucketName:                  bucketName,
		balanceStorage:              balanceStorage,
		tmpOTP:                      make(map[uuid.UUID]dto.OTPHolder),
		mutex:                       sync.Mutex{},
		jwtKey:                      jwtkey,
		tmpUpdateProfileOTP:         make(map[uuid.UUID]dto.UpdateProfileTmpHolder),
		googleOauth:                 gOauth,
		facebookOauth:               fOauth,
		IpFilterMap:                 make(map[string][]dto.IPFilter),
		locker:                      make(map[uuid.UUID]*sync.Mutex),
		balanceLogStorage:           balanceLogStorage,
		operationalGroupStorage:     operationalGroupStorage,
		operationalGroupTypeStorage: operationalGroupTypeStorage,
		ConfigStorage:               configStorage,
		UserBalanceSocket:           map[uuid.UUID]map[*websocket.Conn]bool{},
		UserBalanceSocketLocker:     make(map[*websocket.Conn]*sync.Mutex),
		agentModule:                 agentModule,
		sessionSockets:              make(map[uuid.UUID]map[*websocket.Conn]bool),
		stopChan:                    make(chan struct{}),
		redis:                       redis,
		PisiClient:                  pisiClient,
	}
	// Initialize enterprise registration service
	usr.EnterpriseRegistrationService = NewEnterpriseRegistrationService(
		userStorage,
		otpStorage,
		emailService,
		log,
	)

	usr.initializeBetEngine()

	// Start session monitoring service
	ctx := context.Background()
	go usr.MonitorUserSessions(ctx)
	log.Info("Session monitoring service started")

	return usr
}

func (u *User) initializeBetEngine() error {
	//check if the multiplier is exist or not
	_, exist, err := u.userStorage.GetReferalMultiplier(context.Background())
	if err != nil {
		u.log.Error("Failed to get referral multiplier", zap.Error(err))
		u.log.Error("Continuing without referral multiplier - this is expected for new installations")
		return nil
	}
	u.cron = cron.New(cron.WithSeconds())
	_, err = u.cron.AddFunc("@every 30s", u.updateipfilterDatabase)
	if err != nil {
		return err
	}
	u.cron.Start()
	if !exist {
		u.userStorage.CreateReferalCodeMultiplier(context.Background(), dto.ReferalMultiplierReq{
			PointMultiplier: 1,
		})
	}

	// check for the users who don't have refral code
	urs, err := u.userStorage.GetUsersDoseNotHaveReferalCode(context.Background())
	if err != nil {
		u.log.Error("Failed to get users without referral codes", zap.Error(err))
		u.log.Error("Continuing without referral code generation - this is expected for new installations")
		return nil
	}

	if len(urs) > 0 {
		// generate and add referal code for the users
		for _, usr := range urs {
			// generate new referal code with length of 12
			for {
				newReferal := utils.GenerateRandomUsername(12)
				_, exist, err := u.userStorage.GetUserPointsByReferalPoint(context.Background(), newReferal)

				if err != nil {
					u.log.Error("Failed to get user points by referral point", zap.Error(err))
					break
				}

				if exist {
					continue
				}

				if err := u.userStorage.AddReferalCode(context.Background(), usr.ID, newReferal); err != nil {
					u.log.Error("Failed to add referral code", zap.Error(err))
					break
				}

				break

			}

		}

	}

	alloweList, exist, err := u.userStorage.GetIPFilterByType(context.Background(), "allow")
	if err != nil {
		return err
	}
	if exist {
		u.IpFilterMap["allow"] = alloweList
	}

	denyList, exist, err := u.userStorage.GetIPFilterByType(context.Background(), "deny")
	if err != nil {
		return err
	}

	if exist {
		u.IpFilterMap["deny"] = denyList
	}

	// Check for referral bonus config - create default if not exists
	_, exist, err = u.ConfigStorage.GetConfigByName(context.Background(), constant.REFERRAL_BONUS)
	if err != nil {
		u.log.Error("Failed to check referral bonus config", zap.Error(err))
	}

	if !exist {
		// Create default referral bonus config
		_, err = u.ConfigStorage.CreateConfig(context.Background(), dto.Config{
			Name:  constant.REFERRAL_BONUS,
			Value: "0", // Default to 0, admin can update via endpoint
		})
		if err != nil {
			u.log.Error("Failed to create default referral bonus config", zap.Error(err))
		} else {
			u.log.Info("Created default referral bonus config with value 0")
		}
	}

	return nil
}

func (u *User) RegisterUser(ctx context.Context, userRequest dto.User) (dto.UserRegisterResponse, string, error) {
	var exist bool
	var err error
	// referralCode := userRequest.ReferralCode

	// validate user
	if err := dto.ValidateUser(userRequest); err != nil {
		u.log.Error("invalid user request ", zap.Error(err), zap.Any("user", userRequest))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}

	if userRequest.Email == "" {
		err = fmt.Errorf("email is required")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}

	// check currency and validate currency
	// if the currency is empty add USD as default currency
	userRequest.DefaultCurrency = constant.DEFAULT_CURRENCY

	if userRequest.DateOfBirth != "" {
		dateLayout := "2006-01-02"
		birthDate, err := time.Parse(dateLayout, userRequest.DateOfBirth)
		if err != nil {
			u.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.UserRegisterResponse{}, "", err
		}

		currentDate := time.Now()

		age := currentDate.Year() - birthDate.Year()
		if currentDate.YearDay() < birthDate.YearDay() {
			age--
		}

		if age < 18 {
			err = fmt.Errorf("invalid age, age must be greater than 18")
			err = errors.ErrAcessError.Wrap(err, err.Error())
			return dto.UserRegisterResponse{}, "", err
		}
	}

	// check if the email already exist or not
	if userRequest.Email != "" {
		_, exist, err = u.userStorage.GetUserByEmail(ctx, userRequest.Email)
		if err != nil {
			return dto.UserRegisterResponse{}, "", err
		}
		// if the email is exist and the source is not phone then return error
		if exist {
			err = fmt.Errorf("user already exist with this email")
			u.log.Warn("user already exist ", zap.Any("email", userRequest.Email))
			err = errors.ErrDataAlredyExist.Wrap(err, err.Error())
			return dto.UserRegisterResponse{}, "", err
		}

	}

	//check if user is exist with  requested phone number (only if phone number is provided)
	if userRequest.PhoneNumber != "" {
		_, exist, err = u.userStorage.GetUserByPhoneNumber(ctx, userRequest.PhoneNumber)
		if err != nil {
			return dto.UserRegisterResponse{}, "", err
		}

		if exist {
			err = fmt.Errorf("user already exist with this phone")
			u.log.Warn("user already exist ", zap.Any("phone", userRequest.PhoneNumber))
			err = errors.ErrDataAlredyExist.Wrap(err, err.Error())
			return dto.UserRegisterResponse{}, "", err
		}
	}
	hashPassword, err := utils.HashPassword(userRequest.Password)
	if err != nil {
		u.log.Error("unable to hash password", zap.Error(err), zap.Any("user", userRequest))
		err = errors.ErrInternalServerError.Wrap(err, "unable to hash password")
		return dto.UserRegisterResponse{}, "", err
	}

	//create user
	userRequest.Password = hashPassword
	userRequest.Source = constant.SOURCE_PHONE

	// Determine prefix based on the user type
	prefix := utils.GetPrefix(userRequest.Type)

	for {
		userRequest.ReferralCode = prefix + utils.GenerateRandomUsername(12)
		// Check if referral code already exists by checking if a user has this referral code
		existingUser, err := u.userStorage.GetUserByReferalCode(ctx, userRequest.ReferralCode)
		if err != nil {
			u.log.Error("Failed to check referral code uniqueness", zap.Error(err), zap.String("referral_code", userRequest.ReferralCode))
			return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "failed to check referral code uniqueness")
		}

		// If no user found with this referral code, it's unique
		if existingUser == nil {
			break
		}

		// If referral code exists, generate a new one
		u.log.Debug("Referral code already exists, generating new one", zap.String("referral_code", userRequest.ReferralCode))
	}

	// Set the username to the same value as the referral code for phone-based registration
	// For email-based registration, username is provided in the request
	u.log.Debug("Username before processing", zap.String("username", userRequest.Username), zap.String("phone", userRequest.PhoneNumber), zap.String("email", userRequest.Email))
	if userRequest.PhoneNumber != "" {
		userRequest.Username = userRequest.ReferralCode
		u.log.Debug("Username set to referral code for phone registration", zap.String("username", userRequest.Username))
	}

	// check the referal user is exist
	_, err = u.userStorage.GetUserByReferalCode(ctx, userRequest.ReferedByCode)
	if err != nil && err != sqlcerr.ErrNoRows {
		return dto.UserRegisterResponse{}, "", err
	}

	usrRes, err := u.userStorage.CreateUser(ctx, userRequest)
	if err != nil {
		return dto.UserRegisterResponse{}, "", err
	}

	//create user balance
	if userRequest.IsAdmin {
		u.balanceStorage.CreateBalance(ctx, dto.Balance{
			UserId:       usrRes.ID,
			CurrencyCode: constant.DEFAULT_CURRENCY,
			RealMoney:    decimal.Zero,
			BonusMoney:   decimal.Zero,
			Points:       0,
		})
	} else {
		signUpdBonus, exist, err := u.ConfigStorage.GetConfigByName(ctx, constant.SIGNUP_BONUS)
		if err != nil {
			u.log.Error("unable to get signup bonus config", zap.Error(err), zap.Any("user", userRequest))
		}

		if !exist {
			u.balanceStorage.CreateBalance(ctx, dto.Balance{
				UserId:       usrRes.ID,
				CurrencyCode: constant.DEFAULT_CURRENCY,
				RealMoney:    decimal.Zero,
				BonusMoney:   decimal.Zero,
				Points:       0,
			})
		} else {
			// if the signup bonus is exist then add the bonus to the user
			signUpBonusAmount, err := decimal.NewFromString(signUpdBonus.Value)
			if err != nil {
				u.log.Error("unable to convert signup bonus amount", zap.Error(err), zap.Any("user", userRequest))
				return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "unable to convert signup bonus amount")
			}
			u.balanceStorage.CreateBalance(ctx, dto.Balance{
				UserId:       usrRes.ID,
				CurrencyCode: constant.DEFAULT_CURRENCY,
				RealMoney:    decimal.Zero,
				BonusMoney:   signUpBonusAmount,
				Points:       0,
			})
		}

	}

	// save to temp data

	data := dto.ReferralData{
		ReferedByCode:  userRequest.ReferedByCode,
		AgentRequestID: userRequest.AgentRequestID,
		ReferralCode:   userRequest.ReferralCode,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		u.log.Error("Failed to marshal referral data", zap.Error(err), zap.String("userID", usrRes.ID.String()))
		return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "failed to marshal referral data")
	}

	if err := u.userStorage.SaveToTemp(ctx, dto.UserReferals{
		UserID: usrRes.ID,
		Data:   dataBytes,
	}); err != nil {
		u.log.Error("Failed to save user referral data to temp storage", zap.Error(err), zap.String("userID", usrRes.ID.String()))
		return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "failed to save user referral data to temp storage")
	}

	// Skip SMS for email-based registration - OTP is handled by email service
	// SMS and phone-based OTP are only used for phone-based registration flows

	// Generate JWT tokens for automatic login after registration
	accessToken, err := utils.GenerateJWT(usrRes.ID)
	if err != nil {
		u.log.Error("Failed to generate access token", zap.Error(err), zap.String("userID", usrRes.ID.String()))
		return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshJWT(usrRes.ID)
	if err != nil {
		u.log.Error("Failed to generate refresh token", zap.Error(err), zap.String("userID", usrRes.ID.String()))
		return dto.UserRegisterResponse{}, "", errors.ErrInternalServerError.Wrap(err, "failed to generate refresh token")
	}

	return dto.UserRegisterResponse{
		Message:      constant.USER_REGISTRATION_SUCCESS,
		UserID:       usrRes.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "", nil
}

// UpdateReferralBonus updates the referral bonus configuration
func (u *User) UpdateReferralBonus(ctx context.Context, req dto.ReferralBonusReq) (dto.ReferralBonusRes, error) {
	// Update or create the referral bonus config
	_, err := u.ConfigStorage.UpdateConfigByName(ctx, dto.Config{
		Name:  constant.REFERRAL_BONUS,
		Value: req.Amount.String(),
	})
	if err != nil {
		// If update fails, try to create
		_, err = u.ConfigStorage.CreateConfig(ctx, dto.Config{
			Name:  constant.REFERRAL_BONUS,
			Value: req.Amount.String(),
		})
		if err != nil {
			u.log.Error("Failed to create referral bonus config", zap.Error(err))
			return dto.ReferralBonusRes{}, errors.ErrInternalServerError.Wrap(err, "failed to update referral bonus configuration")
		}
	}

	u.log.Info("Referral bonus configuration updated", zap.String("amount", req.Amount.String()))

	return dto.ReferralBonusRes{
		Message: constant.SUCCESS,
		Amount:  req.Amount,
	}, nil
}

// GetReferralBonusConfig retrieves the current referral bonus configuration
func (u *User) GetReferralBonusConfig(ctx context.Context) (dto.ReferralBonusRes, error) {
	config, exist, err := u.ConfigStorage.GetConfigByName(ctx, constant.REFERRAL_BONUS)
	if err != nil {
		u.log.Error("Failed to get referral bonus config", zap.Error(err))
		return dto.ReferralBonusRes{}, errors.ErrInternalServerError.Wrap(err, "failed to get referral bonus configuration")
	}

	if !exist {
		return dto.ReferralBonusRes{
			Message: "Referral bonus configuration not found",
			Amount:  decimal.Zero,
		}, nil
	}

	amount, err := decimal.NewFromString(config.Value)
	if err != nil {
		u.log.Error("Failed to parse referral bonus amount", zap.Error(err))
		return dto.ReferralBonusRes{}, errors.ErrInternalServerError.Wrap(err, "failed to parse referral bonus amount")
	}

	return dto.ReferralBonusRes{
		Message: constant.SUCCESS,
		Amount:  amount,
	}, nil
}

// processReferralBonus processes the referral bonus for the referrer
func (u *User) processReferralBonus(ctx context.Context, referralCode string, newUserID uuid.UUID) error {
	// Get the referrer user by referral code
	referrer, err := u.userStorage.GetUserByReferalCode(ctx, referralCode)
	if err != nil {
		return fmt.Errorf("failed to get referrer by referral code: %w", err)
	}

	// Get referral bonus config
	referralBonusConfig, exist, err := u.ConfigStorage.GetConfigByName(ctx, constant.REFERRAL_BONUS)
	if err != nil {
		return fmt.Errorf("failed to get referral bonus config: %w", err)
	}

	if !exist {
		return fmt.Errorf("referral bonus config not found")
	}

	// Parse the bonus amount
	realAmount, err := decimal.NewFromString(referralBonusConfig.Value)
	if err != nil {
		return fmt.Errorf("failed to parse referral bonus amount: %w", err)
	}
	u.log.Info("Processing referral bonus", zap.Any("referralCode", referralCode))

	// Get or create referrer's balance
	referrerBalance, _, err := u.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       referrer.UserID,
		CurrencyCode: "P",
	})
	if err != nil {
		return fmt.Errorf("failed to get referrer balance: %w", err)
	}
	u.log.Info("Referrer balance after update", zap.Any("referrerBalance", referrerBalance.RealMoney.Add(realAmount)))

	// check if tthe bala
	newMoney := referrerBalance.RealMoney.Add(realAmount)
	_, err = u.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    referrer.UserID,
		Currency:  "P",
		Amount:    newMoney,
		Component: constant.REAL_MONEY,
	})
	if err != nil {
		return fmt.Errorf("failed to update referrer balance: %w", err)
	}

	// Create operational group and type for referral bonus
	operationalGroupAndType, err := u.CreateOrGetOperationalGroupAndType(ctx, constant.DEPOSIT, constant.REFERRAL_BONUS)
	if err != nil {
		return fmt.Errorf("failed to create operational group and type: %w", err)
	}

	// Save balance log
	transactionID := utils.GenerateTransactionId()
	_, err = u.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             referrer.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           constant.POINTS,
		Description:        fmt.Sprintf("Referral bonus for new user %s", newUserID.String()),
		ChangeAmount:       realAmount,
		OperationalGroupID: operationalGroupAndType.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndType.OperationalTypeID,
		TransactionID:      &transactionID,
		BalanceAfterUpdate: &referrerBalance.BonusMoney,
		Status:             constant.COMPLTE,
	})
	if err != nil {
		return fmt.Errorf("failed to save balance log: %w", err)
	}

	u.log.Info("Referral bonus processed successfully",
		zap.String("referrerID", referrer.UserID.String()),
		zap.String("newUserID", newUserID.String()),
		zap.String("bonusAmount", realAmount.String()))

	return nil
}

func (u *User) processAgentReferralConversion(ctx context.Context, agentRequestID string, userID uuid.UUID, msisdn string) error {
	u.log.Info("Processing agent referral conversion",
		zap.String("agentRequestID", agentRequestID),
		zap.String("userID", userID.String()),
		zap.String("msisdn", msisdn))

	// Get signup bonus amount for conversion
	signUpBonusAmount := decimal.Zero
	signUpdBonus, exist, err := u.ConfigStorage.GetConfigByName(ctx, constant.SIGNUP_BONUS)
	if err != nil {
		u.log.Error("unable to get signup bonus config for agent conversion", zap.Error(err))
	} else if exist {
		signUpBonusAmount, err = decimal.NewFromString(signUpdBonus.Value)
		if err != nil {
			u.log.Error("unable to convert signup bonus amount for agent conversion", zap.Error(err))
			signUpBonusAmount = decimal.Zero
		}
	}

	conversionReq := dto.UpdateAgentReferralWithConversionReq{
		RequestID:      agentRequestID,
		UserID:         userID,
		ConversionType: "registeration",
		Amount:         signUpBonusAmount,
		MSISDN:         msisdn,
	}

	_, err = u.agentModule.UpdateAgentReferralWithConversion(ctx, conversionReq)
	if err != nil {
		u.log.Error("Failed to record agent referral conversion",
			zap.Error(err),
			zap.String("agentRequestID", agentRequestID),
			zap.String("userID", userID.String()))
		return err
	}

	u.log.Info("Agent referral conversion recorded successfully",
		zap.String("agentRequestID", agentRequestID),
		zap.String("userID", userID.String()),
		zap.String("msisdn", msisdn),
		zap.String("conversionType", "registration"),
		zap.String("amount", signUpBonusAmount.String()))

	return nil
}

func (u *User) Login(ctx context.Context, loginRequest dto.UserLoginReq, loginLogs dto.LoginAttempt, adminLogin bool) (dto.UserLoginRes, string, error) {
	var usr dto.User
	var exist bool
	var err error
	if err := dto.ValidateLoginRequest(loginRequest); err != nil {
		u.log.Error("invalid login request ", zap.Error(err), zap.Any("loginrequest", loginRequest))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserLoginRes{}, "", err
	}
	// get user by phone or username
	usr, exist, err = u.userStorage.GetUserByUserName(ctx, loginRequest.LoginID)
	if err != nil {
		return dto.UserLoginRes{}, "", err
	}
	if !exist {
		// try to get user using phone number
		usr, exist, err = u.userStorage.GetUserByPhoneNumber(ctx, loginRequest.LoginID)
		if err != nil {
			return dto.UserLoginRes{}, "", err
		}
		if !exist {
			// try to get user using email
			usr, exist, err = u.userStorage.GetUserByEmail(ctx, loginRequest.LoginID)
			if err != nil {
				return dto.UserLoginRes{}, "", err
			}
			if !exist {
				err = fmt.Errorf("login id not correct")
				u.log.Warn(err.Error(), zap.Any("loginRequest", loginRequest))
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				return dto.UserLoginRes{}, "", err
			}
		}

	}
	//
	//check for user login blocked  or not
	if !adminLogin {
		acBlocks, exist, err := u.userStorage.GetBlockedAccountByUserID(ctx, usr.ID)
		if err != nil {
			return dto.UserLoginRes{}, "", err
		}

		if exist {
			// check if account blocked for login
			for _, acBlock := range acBlocks {
				if acBlock.Type == constant.BLOCK_TYPE_LOGIN || acBlock.Type == constant.BLOCK_TYPE_COMPLETE {
					// check for  duration and type
					if acBlock.Duration == constant.BLOCK_DURATION_PERMANENT {
						err = fmt.Errorf("user can not login, account permanently blocked ")
						err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
						return dto.UserLoginRes{}, "", err
					} else {
						//check if the user is still blocked
						if acBlock.BlockedTo != nil {
							blockedUntil := *acBlock.BlockedTo
							if blockedUntil.After(time.Now().In(time.Now().Location()).UTC()) {
								err = fmt.Errorf("user account temporary blocked  for %2f hours", blockedUntil.Sub(time.Now().In(time.Now().Location()).UTC()).Hours())
								err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
								return dto.UserLoginRes{}, "", err
							}
						}
					}
				}
			}
		}
	}

	loginLogs.UserID = usr.ID
	loginLogs.Success = true
	loginLogs.AttemptTime = time.Now()
	if adminLogin {
		if !usr.IsAdmin {
			err = fmt.Errorf("unauthorized")
			u.log.Error(err.Error())
			return dto.UserLoginRes{}, "", errors.ErrInvalidUserInput.Wrap(err, err.Error())
		}
	}
	//check password
	if ok := utils.ComparePasswords(usr.Password, loginRequest.Password); !ok {
		err = fmt.Errorf("login id or password not correct")
		u.log.Warn(err.Error(), zap.Any("loginRequest", loginRequest))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		loginLogs.Success = false
		u.logsStorage.CreateLoginAttempts(ctx, loginLogs)
		return dto.UserLoginRes{}, "", err
	}
	// Check user verification status
	isVerified := false
	emailVerified := false
	phoneVerified := false

	// For now, we'll consider users verified if they have completed registration
	// In a real implementation, you'd check actual verification flags
	if usr.Status == "verified" || usr.Status == "active" {
		isVerified = true
		emailVerified = true
		phoneVerified = true
	}

	// generate jwt token with verification status
	token, err := utils.GenerateJWTWithVerification(usr.ID, isVerified, emailVerified, phoneVerified)
	if err != nil {
		err = fmt.Errorf("unable to generate jwt token")
		u.log.Warn(err.Error(), zap.Any("loginRequest", loginRequest))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserLoginRes{}, "", err
	}
	u.logsStorage.CreateLoginAttempts(ctx, loginLogs)

	// generate refresh token
	refreshToken, err := utils.GenerateRefreshJWT(usr.ID)
	if err != nil {
		u.log.Error("unable to generate refresh token", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserLoginRes{}, "", err
	}
	refreshTokenExpiry := time.Now().Add(30 * time.Minute) // Restored to 30 minutes for production

	// save user session with refresh token
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

	// Get user profile for response
	u.log.Info("User data from database", zap.String("username", usr.Username), zap.String("email", usr.Email), zap.String("userID", usr.ID.String()))
	userProfile := dto.UserProfile{
		Username:     usr.Username,
		UserID:       usr.ID,
		Email:        usr.Email,
		PhoneNumber:  usr.PhoneNumber,
		FirstName:    usr.FirstName,
		LastName:     usr.LastName,
		Type:         usr.Type,
		ReferralCode: usr.ReferralCode,
	}

	return dto.UserLoginRes{
		Message:     constant.LOGIN_SUCCESS,
		AccessToken: token,
		UserProfile: &userProfile,
	}, refreshToken, nil
}

func (u *User) GetProfile(ctx context.Context, userID uuid.UUID) (dto.UserProfile, error) {
	u.log.Info("GetProfile called", zap.String("userID", userID.String()))
	profile, exist, err := u.userStorage.GetUserByID(ctx, userID)
	var profilePicture string
	if err != nil {
		return dto.UserProfile{}, err
	}
	if !exist {
		err := fmt.Errorf("user not found")
		u.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserProfile{}, err
	}
	if profile.ProfilePicture != "" {
		profilePicture = utils.GetFromS3Bucket(u.bucketName, profile.ProfilePicture)
	}

	u.log.Info("Profile data from database", zap.String("username", profile.Username), zap.String("email", profile.Email), zap.String("userID", profile.ID.String()))
	return dto.UserProfile{
		Username:       profile.Username,
		PhoneNumber:    profile.PhoneNumber,
		Email:          profile.Email,
		ProfilePicture: profilePicture,
		UserID:         profile.ID,
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
		ReferralCode:   profile.ReferralCode,
		ReferedByCode:  profile.ReferedByCode,
		ReferalType:    profile.ReferalType,
		Type:           profile.Type,
	}, err
}

func (u *User) UploadProfilePicture(ctx context.Context, img multipart.File, header *multipart.FileHeader, userID uuid.UUID) (string, error) {
	// Extract the original file name and get the extension
	fileExtension := filepath.Ext(header.Filename)
	if fileExtension == "" {
		err := fmt.Errorf("invalid file extension")
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return "", err
	}

	profilePictureName := uuid.New().String() + fileExtension

	// Create S3 instance
	s3Instance := utils.NewS3Instance(u.log, constant.VALID_IMGS)
	if s3Instance == nil {
		err := fmt.Errorf("unable to create s3 session")
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return "", err
	}

	// Update profile picture name in user storage
	profilePicture, err := u.userStorage.UpdateProfilePicuter(ctx, userID, profilePictureName)
	if err != nil {
		return "", err
	}

	// Upload file to S3
	return s3Instance.UploadToS3Bucket(u.bucketName, img, profilePicture, header.Header.Get("Content-Type"))
}

func (u *User) ChangePassword(ctx context.Context, changePasswordReq dto.ChangePasswordReq) (dto.ChangePasswordRes, error) {
	if err := dto.ValidateChangePassword(changePasswordReq); err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ChangePasswordRes{}, err
	}
	if changePasswordReq.NewPassword != changePasswordReq.ConfirmPassword {
		err := fmt.Errorf("new password and confirm password is not equal, %s not equal to %s", changePasswordReq.NewPassword, changePasswordReq.ConfirmPassword)
		u.log.Warn(err.Error(), zap.Any("userID", changePasswordReq.UserID.String()))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ChangePasswordRes{}, err
	}

	// check old password
	usr, exist, err := u.userStorage.GetUserByID(ctx, changePasswordReq.UserID)
	if err != nil {
		return dto.ChangePasswordRes{}, err
	}
	if !exist {
		err = fmt.Errorf("user not found with user id %s", changePasswordReq.UserID.String())
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ChangePasswordRes{}, err
	}
	ok := utils.ComparePasswords(usr.Password, changePasswordReq.OldPassword)
	if !ok {
		err = fmt.Errorf("invalid old password is given userID %s", changePasswordReq.UserID.String())
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ChangePasswordRes{}, err
	}
	hashedPassword, err := utils.HashPassword(changePasswordReq.NewPassword)
	if err != nil {
		u.log.Error("unable to hash password", zap.Error(err), zap.Any("user", changePasswordReq.UserID.String()))
		err = errors.ErrInternalServerError.Wrap(err, "unable to hash password")
		return dto.ChangePasswordRes{}, err
	}

	// updateUserPassword
	_, err = u.userStorage.UpdatePassword(ctx, changePasswordReq.UserID, hashedPassword)
	if err != nil {
		return dto.ChangePasswordRes{}, err
	}

	return dto.ChangePasswordRes{
		Message: constant.UPDATE_PASSWORD_SUCCESS,
		UserID:  changePasswordReq.UserID,
	}, nil
}

func (u *User) ForgetPassword(ctx context.Context, usernameOrPhoneOrEmail string) (*dto.ForgetPasswordRes, error) {
	//check if the username or phone or email is valid
	user, exist, err := u.userStorage.GetUserByEmail(ctx, usernameOrPhoneOrEmail)
	if err != nil {
		err := errors.ErrAcessError.New("invalid login information provided")
		u.log.Error(err.Error())
		return nil, err
	}
	if !exist {
		err = errors.ErrAcessError.New("invalid login information provided")
		u.log.Error(err.Error())
		return nil, err
	}

	// save OTP to redis
	otp := utils.GenerateOTP()

	u.PisiClient.SendBulkSMS(ctx, pisi.SendBulkSMSRequest{
		Message:    fmt.Sprintf("Your code is {%s}. Please do not share this code.", otp),
		Recipients: usernameOrPhoneOrEmail,
	})

	//validate the username or phone or email
	err = u.redis.SaveOTP(ctx, user.PhoneNumber, otp)
	if err != nil {
		return nil, err
	}
	return &dto.ForgetPasswordRes{
		Message: constant.FORGOT_PASSWORD_RES,
	}, nil
}

func (u *User) VerifyResetPassword(ctx context.Context, resetPasswordReq dto.VerifyResetPasswordReq) (*dto.VerifyResetPasswordRes, error) {
	valid, err := u.redis.VerifyAndRemoveOTP(ctx, resetPasswordReq.EmailOrPhoneOrUserame, resetPasswordReq.OTP)
	if err != nil {
		if err.Error() == "Too many invalid attempts. Please request a new OTP." {
			return nil, errors.ErrAcessError.New("Too many invalid attempts. Please request a new OTP.")
		}
		return nil, err
	}
	if !valid {
		err := errors.ErrAcessError.New("invalid otp")
		return nil, err
	}
	usr, exist, err := u.userStorage.GetUserByEmail(ctx, resetPasswordReq.EmailOrPhoneOrUserame)
	if err != nil {
		return nil, err
	}
	if !exist {
		err := errors.ErrAcessError.New("invalid user")
		return nil, err
	}
	// generate jwt token for the user to reset password
	token, err := utils.GenerateOTPJWT(usr.ID)
	if err != nil {
		return nil, err
	}
	return &dto.VerifyResetPasswordRes{
		UserID: usr.ID,
		Token:  token,
	}, nil
}

func (u *User) GenerateOTPToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := utils.GenerateOTPJWT(userID)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return "", err
	}
	return token, nil
}

func (u *User) ValidateOTPExp(ctx context.Context, createdAt time.Time, userID uuid.UUID) error {
	// Ensure both times are in UTC for consistent comparison
	validTimes := time.Now().In(time.Now().Location()).UTC().Add(-10 * time.Minute)
	if createdAt.Before(validTimes) {
		// OTP has expired
		//delete from database
		if err := u.userStorage.DeleteOTP(ctx, userID); err != nil {
			return err
		}
		err := fmt.Errorf("OTP has expired")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	return nil
}

func (u *User) ResetPassword(ctx context.Context, resetPasswordReq dto.ResetPasswordReq) (dto.ResetPasswordRes, error) {

	//verify token
	claims := &dto.Claim{}
	jwtkey := []byte(u.jwtKey)
	token, err := jwt.ParseWithClaims(resetPasswordReq.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtkey, nil
	})
	if err != nil || !token.Valid {
		err := fmt.Errorf("invalid or expired token ")
		err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())

		return dto.ResetPasswordRes{}, err
	}

	//verify password requirement
	if err := dto.ValidateResetPassword(resetPasswordReq); err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ResetPasswordRes{}, err
	}

	//check password and confirm password are equal
	if resetPasswordReq.NewPassword != resetPasswordReq.ConfirmPassword {
		err = fmt.Errorf("new_password and cofirm_password are not equal")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.ResetPasswordRes{}, err
	}

	//update passwrod
	hashPassword, err := utils.HashPassword(resetPasswordReq.NewPassword)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ResetPasswordRes{}, err
	}

	//save updated password to the user
	_, err = u.userStorage.UpdatePassword(ctx, claims.UserID, hashPassword)
	if err != nil {
		return dto.ResetPasswordRes{}, err
	}
	//delete otp from otp list
	delete(u.tmpOTP, claims.UserID)
	u.userStorage.DeleteOTP(ctx, claims.UserID)
	tokenStr, err := utils.GenerateJWT(claims.UserID)
	if err != nil {
		err = fmt.Errorf("unable to generate jwt token")
		u.log.Warn(err.Error(), zap.Any("loginRequest", resetPasswordReq))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.ResetPasswordRes{}, err
	}
	return dto.ResetPasswordRes{
		Token:  tokenStr,
		UserID: claims.UserID,
	}, nil
}

func (u *User) UpdateProfile(ctx context.Context, profileupateReq dto.UpdateProfileReq) (dto.UpdateProfileRes, error) {
	// save it to map of user reset request
	// delete if already reques exist
	u.mutex.Lock()
	defer u.mutex.Unlock()
	delete(u.tmpUpdateProfileOTP, profileupateReq.UserID)
	//generate otp
	otp := utils.GenerateOTP()

	//send otp to the user if they have email if they don't since we don't have sms gateway update the user profile
	//get user by UserID
	usr, exist, err := u.userStorage.GetUserByID(ctx, profileupateReq.UserID)
	if err != nil {
		return dto.UpdateProfileRes{}, err
	}

	if !exist {
		err = fmt.Errorf("unable to get user")
		u.log.Error(err.Error(), zap.Any("user-id", profileupateReq.UserID))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateProfileRes{}, err
	}

	// generate token for update profile
	token, err := utils.GenerateOTPJWT(profileupateReq.UserID)
	if err != nil {
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UpdateProfileRes{}, err
	}

	if usr.Email != "" {
		// send otp to the user

		//send otp to the email
		body := constant.GetUpdateProfileOTPTemplate(otp)
		utils.SendEmail(ctx, dto.EmailReq{
			Subject: "Upate Profile",
			To:      []string{usr.Email},
			Body:    []byte(body),
		})
		u.tmpUpdateProfileOTP[profileupateReq.UserID] = dto.UpdateProfileTmpHolder{
			TmpOTP:           otp,
			CreatedAT:        time.Now().In(time.Now().Location()).UTC(),
			Attempts:         0,
			UpdateProfileReq: profileupateReq,
			Any:              false,
		}
		return dto.UpdateProfileRes{
			Token:  token,
			UserID: profileupateReq.UserID,
			Any:    false,
		}, nil
	} else {
		// send with any true value
		u.tmpUpdateProfileOTP[profileupateReq.UserID] = dto.UpdateProfileTmpHolder{
			TmpOTP:           otp,
			CreatedAT:        time.Now().In(time.Now().Location()).UTC(),
			Attempts:         0,
			UpdateProfileReq: profileupateReq,
			Any:              true,
		}
		return dto.UpdateProfileRes{
			Token:  token,
			UserID: profileupateReq.UserID,
			Any:    true,
		}, nil
	}
}

// UpdateUserVerificationStatus updates the email verification status of a user
// This is a production-grade implementation that delegates to the storage layer
func (u *User) UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, verified bool) (dto.User, error) {
	if userID == uuid.Nil {
		u.log.Error("invalid user ID provided for verification status update", zap.String("user_id", "nil"))
		return dto.User{}, errors.ErrInvalidUserInput.New("invalid user ID")
	}

	u.log.Info("updating user verification status",
		zap.String("user_id", userID.String()),
		zap.Bool("verified", verified))

	// Delegate to the storage layer for the actual database operation
	updatedUser, err := u.userStorage.UpdateUserVerificationStatus(ctx, userID, verified)
	if err != nil {
		u.log.Error("failed to update user verification status",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Bool("verified", verified))
		return dto.User{}, err
	}

	u.log.Info("user verification status updated successfully",
		zap.String("user_id", userID.String()),
		zap.Bool("verified", verified))

	return updatedUser, nil
}

func (u *User) ConfirmUpdateProfile(ctx context.Context, confirmOTP dto.ConfirmUpdateProfile) (dto.User, error) {
	claims := &dto.Claim{}
	jwtkey := []byte(u.jwtKey)
	token, err := jwt.ParseWithClaims(confirmOTP.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtkey, nil
	})

	if err != nil || !token.Valid {
		err := fmt.Errorf("invalid or expired token ")
		err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())

		return dto.User{}, err
	}

	//check if the user has otp validatio end or not
	updateReq, ok := u.tmpUpdateProfileOTP[claims.UserID]
	if !ok {
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid otp"), "invalid otp")
		return dto.User{}, err
	}

	usr, exist, err := u.userStorage.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return dto.User{}, err
	}

	if !exist {
		err = errors.ErrInvalidUserInput.Wrap(fmt.Errorf("user not found"), "user not found")
		return dto.User{}, err
	}

	if updateReq.Any {
		// update without validating otp
		return u.SaveUpdateProfile(ctx, updateReq.UpdateProfileReq, usr)
	}
	// check otp
	if updateReq.Attempts < 3 {
		if updateReq.TmpOTP == confirmOTP.OTP {
			delete(u.tmpUpdateProfileOTP, claims.UserID)
			return u.SaveUpdateProfile(ctx, updateReq.UpdateProfileReq, usr)
		}
		updateReq.Attempts = updateReq.Attempts + 1
		u.tmpUpdateProfileOTP[claims.UserID] = updateReq
		err = errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid otp"), "invalid otp")
		return dto.User{}, err
	}
	delete(u.tmpUpdateProfileOTP, claims.UserID)
	err = errors.ErrInvalidUserInput.Wrap(fmt.Errorf("invalid otp"), "invalid otp")
	return dto.User{}, err
}
func (u *User) SaveUpdateProfile(ctx context.Context, profile dto.UpdateProfileReq, usr dto.User) (dto.User, error) {
	updatedUsr := usr
	//validate email and phone number is taken or not
	if err := u.ValidateUpdateReq(ctx, profile); err != nil {
		return dto.User{}, err
	}
	if profile.Email != "" && profile.Source != constant.SOURCE_GMAIL {
		updatedUsr.Email = profile.Email
	}

	if profile.FirstName != "" {
		updatedUsr.FirstName = profile.FirstName
	}
	if profile.LastName != "" {
		updatedUsr.LastName = profile.LastName
	}

	if profile.DateOfBirth != "" {
		dateLayout := "2006-01-02"
		birthDate, err := time.Parse(dateLayout, profile.DateOfBirth)
		if err != nil {
			u.log.Error(err.Error())
			err = errors.ErrInternalServerError.Wrap(err, err.Error())
			return dto.User{}, err
		}

		currentDate := time.Now()

		age := currentDate.Year() - birthDate.Year()
		if currentDate.YearDay() < birthDate.YearDay() {
			age--
		}

		if age < 18 {
			err = fmt.Errorf("invalid age, age must be greater than 18")
			err = errors.ErrAcessError.Wrap(err, err.Error())
			return dto.User{}, err
		}
		updatedUsr.DateOfBirth = profile.DateOfBirth
	}

	if profile.Phone != "" && profile.Source != constant.SOURCE_PHONE {
		updatedUsr.PhoneNumber = profile.Phone
	}

	return u.userStorage.UpdateUser(ctx, dto.UpdateProfileReq{
		UserID:      updatedUsr.ID,
		FirstName:   updatedUsr.FirstName,
		LastName:    updatedUsr.LastName,
		Email:       updatedUsr.Email,
		DateOfBirth: updatedUsr.DateOfBirth,
		Phone:       updatedUsr.PhoneNumber,
	})
}

func (u *User) ValidateUpdateReq(ctx context.Context, profile dto.UpdateProfileReq) error {
	// validate if the email  is taken or not
	if profile.Email != "" {
		//check if user already exist with this email
		usr, exist, err := u.userStorage.GetUserByEmail(ctx, profile.Email)
		if err != nil {
			return err
		}
		if exist && usr.ID != profile.UserID {
			err := fmt.Errorf("email already taken")
			u.log.Error(err.Error(), zap.Any("profile_update_req", profile))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return err
		}
	}
	//check if the phone is taken or not
	if profile.Phone != "" {
		//check if user already exist with this email
		usr, exist, err := u.userStorage.GetUserByPhoneNumber(ctx, profile.Phone)
		if err != nil {
			return err
		}
		if exist && usr.ID != profile.UserID {
			err := fmt.Errorf("phone number already taken")
			u.log.Error(err.Error(), zap.Any("profile_update_req", profile))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return err
		}
	}

	return nil
}

func (u *User) AdminUpdateProfile(ctx context.Context, userReq dto.EditProfileAdminReq) (dto.EditProfileAdminRes, error) {
	var usr dto.User
	var exist bool
	var err error
	var updatedUser dto.User
	usr, exist, err = u.userStorage.GetUserByID(ctx, userReq.UserID)
	if err != nil {
		return dto.EditProfileAdminRes{}, err
	}
	updatedUser = usr
	if !exist {
		err := fmt.Errorf("unable to find user")
		u.log.Error(err.Error(), zap.Any("updateReq", userReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.EditProfileAdminRes{}, err
	}
	if userReq.Email != "" {
		usr, exist, err := u.userStorage.GetUserByEmail(ctx, userReq.Email)
		if err != nil {
			return dto.EditProfileAdminRes{}, err
		}
		if exist && usr.ID != userReq.UserID {
			err := fmt.Errorf("email already taken")
			u.log.Error(err.Error(), zap.Any("profile_update_req", userReq))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.EditProfileAdminRes{}, err
		}
		updatedUser.Email = userReq.Email
	}

	if userReq.PhoneNumber != "" {
		usr, exist, err := u.userStorage.GetUserByPhoneNumber(ctx, userReq.PhoneNumber)
		if err != nil {
			return dto.EditProfileAdminRes{}, err
		}
		if exist && usr.ID != userReq.UserID {
			err := fmt.Errorf("phone already taken")
			u.log.Error(err.Error(), zap.Any("profile_update_req", userReq))
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.EditProfileAdminRes{}, err
		}
		updatedUser.PhoneNumber = userReq.PhoneNumber
	}

	if userReq.FirstName != "" {
		updatedUser.FirstName = userReq.FirstName
	}

	if userReq.LastName != "" {
		updatedUser.LastName = userReq.LastName
	}

	if userReq.StreetAddress != "" {
		updatedUser.StreetAddress = userReq.StreetAddress
	}
	if userReq.City != "" {
		updatedUser.City = userReq.City
	}
	if userReq.PostalCode != "" {
		updatedUser.PostalCode = userReq.PostalCode
	}
	if userReq.State != "" {
		updatedUser.State = userReq.State
	}

	if userReq.Country != "" {
		updatedUser.Country = userReq.Country
	}

	if userReq.KYCStatus != "" {
		if userReq.KYCStatus != constant.ACTIVE && userReq.KYCStatus != constant.PENDING && userReq.KYCStatus != constant.INACTIVE {
			err = fmt.Errorf("only ACTIVE, INACTIVE or PENDING KYC status allowed")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.EditProfileAdminRes{}, err
		}
		updatedUser.KYCStatus = userReq.KYCStatus
	}

	// Handle new fields
	if userReq.UserName != "" {
		updatedUser.Username = userReq.UserName
	}

	if userReq.DateOfBirth != "" {
		dateLayout := "2006-01-02"
		birthDate, err := time.Parse(dateLayout, userReq.DateOfBirth)
		if err != nil {
			u.log.Error(err.Error())
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			return dto.EditProfileAdminRes{}, err
		}
		// Validate age (must be 18+)
		age := time.Now().Year() - birthDate.Year()
		if time.Now().YearDay() < birthDate.YearDay() {
			age--
		}
		if age < 18 {
			err = fmt.Errorf("invalid age, age must be greater than 18")
			err = errors.ErrAcessError.Wrap(err, err.Error())
			return dto.EditProfileAdminRes{}, err
		}
		updatedUser.DateOfBirth = userReq.DateOfBirth
	}

	if userReq.Status != "" {
		updatedUser.Status = userReq.Status
	}

	if userReq.DefaultCurrency != "" {
		updatedUser.DefaultCurrency = userReq.DefaultCurrency
	}

	if userReq.WalletVerificationStatus != "" {
		updatedUser.WalletVerificationStatus = userReq.WalletVerificationStatus
	}

	// Handle boolean field
	updatedUser.IsEmailVerified = userReq.IsEmailVerified

	res, err := u.userStorage.UpdateUser(ctx, dto.UpdateProfileReq{
		UserID:                   updatedUser.ID,
		FirstName:                updatedUser.FirstName,
		LastName:                 updatedUser.LastName,
		Email:                    updatedUser.Email,
		DateOfBirth:              updatedUser.DateOfBirth,
		Phone:                    updatedUser.PhoneNumber,
		Username:                 updatedUser.Username,
		StreetAddress:            updatedUser.StreetAddress,
		City:                     updatedUser.City,
		PostalCode:               updatedUser.PostalCode,
		State:                    updatedUser.State,
		Country:                  updatedUser.Country,
		KYCStatus:                updatedUser.KYCStatus,
		Status:                   updatedUser.Status,
		IsEmailVerified:          updatedUser.IsEmailVerified,
		DefaultCurrency:          updatedUser.DefaultCurrency,
		WalletVerificationStatus: updatedUser.WalletVerificationStatus,
	})
	if err != nil {
		return dto.EditProfileAdminRes{}, err
	}
	return dto.EditProfileAdminRes{
		Message: constant.SUCCESS,
		User: dto.User{
			ID:              res.ID,
			PhoneNumber:     res.PhoneNumber,
			FirstName:       res.FirstName,
			LastName:        res.LastName,
			Email:           res.Email,
			DefaultCurrency: res.DefaultCurrency,
			ProfilePicture:  res.ProfilePicture,
			DateOfBirth:     res.DateOfBirth,
			ReferralCode:    res.ReferralCode,
			StreetAddress:   res.StreetAddress,
			Country:         res.Country,
			State:           res.State,
			City:            res.City,
			CreatedBy:       res.CreatedBy,
			PostalCode:      res.PostalCode,
			KYCStatus:       res.KYCStatus,
		},
	}, nil
}

func (u *User) AdminResetPassword(ctx context.Context, req dto.AdminResetPasswordReq) (dto.AdminResetPasswordRes, error) {
	//  validate Password Strength
	if err := dto.ValidateAdminResetPassword(req); err != nil {
		u.log.Error(err.Error(), zap.Any("reset_password_req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, "please provide strong password")
		return dto.AdminResetPasswordRes{}, err
	}
	if req.NewPassword != req.ConfirmPassword {
		err := fmt.Errorf("new password and confirm password dose not match")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AdminResetPasswordRes{}, err
	}

	//confirm user exist
	_, exist, err := u.userStorage.GetUserByID(ctx, req.UserID)
	if err != nil {
		return dto.AdminResetPasswordRes{}, err
	}

	if !exist {
		err := fmt.Errorf("unable to find user")
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.AdminResetPasswordRes{}, err
	}

	newPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		u.log.Error("unable to hash password", zap.Any("reset_password_req", req))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.AdminResetPasswordRes{}, err
	}

	resp, err := u.userStorage.UpdatePassword(ctx, req.UserID, newPassword)
	if err != nil {
		return dto.AdminResetPasswordRes{}, err
	}

	return dto.AdminResetPasswordRes{
		Message: constant.SUCCESS,
		User: dto.User{
			ID:              resp.ID,
			PhoneNumber:     resp.PhoneNumber,
			FirstName:       resp.FirstName,
			LastName:        resp.LastName,
			Email:           resp.Email,
			DefaultCurrency: resp.DefaultCurrency,
			ProfilePicture:  resp.ProfilePicture,
			DateOfBirth:     resp.DateOfBirth,
		},
	}, nil
}

func (u *User) GetPlayers(ctx context.Context, req dto.GetPlayersReq) (dto.GetPlayersRes, error) {
	plys := make([]dto.User, 0)
	// else check for page and per_page
	if req.Page == 0 || req.PerPage == 0 {
		req.Page = 1
		req.PerPage = 10
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	if req.Filter.Email != "" || req.Filter.Phone != "" || req.Filter.Username != "" {
		return u.userStorage.GetUsersByEmailAndPhone(ctx, req)
	}

	// get all users
	pls, err := u.userStorage.GetAllUsers(ctx, req)
	if err != nil {
		return dto.GetPlayersRes{}, err
	}

	u.log.Info("GetPlayers debug", zap.Int("users_from_storage", len(pls.Users)), zap.Int64("total_count", pls.TotalCount))

	for _, player := range pls.Users {
		u.log.Info("Processing player", zap.String("player_id", player.ID.String()), zap.String("username", player.Username))
		balance, err := u.balanceStorage.GetBalancesByUserID(ctx, player.ID)
		if err != nil {
			u.log.Error("Failed to get balance for player", zap.Error(err), zap.String("player_id", player.ID.String()))
			// Don't skip the player, just set empty accounts
			player.Accounts = []dto.Balance{}
		} else {
			player.Accounts = balance
		}
		plys = append(plys, player)
	}

	u.log.Info("GetPlayers result", zap.Int("final_users_count", len(plys)))
	return dto.GetPlayersRes{
		TotalPages: pls.TotalPages,
		Message:    constant.SUCCESS,
		Users:      plys,
		TotalCount: pls.TotalCount,
	}, nil
}

func (u *User) GetUserPoints(ctx context.Context, useID uuid.UUID) (dto.GetPointsResp, error) {
	userBalance, exist, err := u.userStorage.GetUserPoints(ctx, useID)
	if err != nil {
		return dto.GetPointsResp{}, err
	}
	if !exist {
		return dto.GetPointsResp{}, nil
	}

	return dto.GetPointsResp{
		Message: constant.SUCCESS,
		Data: dto.UserPoint{
			UserID: useID,
			Point:  int(userBalance.IntPart()),
		},
	}, nil
}

func (u *User) AdminCreatePlayer(ctx context.Context, userRequest dto.User) (dto.UserRegisterResponse, string, error) {
	return u.RegisterUser(ctx, userRequest)
}

func (u *User) AdminLogin(ctx context.Context, loginRequest dto.UserLoginReq, loginLogs dto.LoginAttempt) (dto.UserLoginRes, string, error) {
	loginRes, refreshToken, err := u.Login(ctx, loginRequest, loginLogs, true)
	return loginRes, refreshToken, err
}

func (u *User) GetAdmins(ctx context.Context, req dto.GetAdminsReq) ([]dto.Admin, error) {
	if req.PerPage <= 0 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PerPage
	req.Page = offset

	if req.RoleID == uuid.Nil && req.Status == "" {
		return u.userStorage.GetAdmins(ctx, req)
	} else if req.RoleID != uuid.Nil && req.Status == "" {
		return u.userStorage.GetAdminsByRole(ctx, req)
	} else if req.RoleID == uuid.Nil && req.Status != "" {
		return u.userStorage.GetAdminsByStatus(ctx, req)
	} else {
		return u.userStorage.GetAdminsByRoleAndStatus(ctx, req)
	}
}

func (u *User) UpdateSignupBonus(ctx context.Context, req dto.SignUpBonusReq) (dto.SignUpBonusRes, error) {

	if req.Amount.LessThan(decimal.Zero) {
		err := fmt.Errorf("bonus must be greater than or equal to zero")
		u.log.Error(err.Error(), zap.Any("signup_bonus_req", req))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.SignUpBonusRes{}, err
	}

	_, err := u.ConfigStorage.UpdateConfigByName(ctx, dto.Config{
		Name:  constant.SIGNUP_BONUS,
		Value: req.Amount.String(),
	})

	if err != nil {
		return dto.SignUpBonusRes{}, err
	}

	return dto.SignUpBonusRes{
		Message: constant.SUCCESS,
		Amount:  req.Amount,
	}, nil

}

func (u *User) GetSignupBonusConfig(ctx context.Context) (dto.SignUpBonusRes, error) {
	config, exist, err := u.ConfigStorage.GetConfigByName(ctx, constant.SIGNUP_BONUS)
	if err != nil {
		return dto.SignUpBonusRes{}, err
	}

	if !exist {
		err = fmt.Errorf("signup bonus config not found")
		u.log.Error(err.Error())
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.SignUpBonusRes{}, err
	}

	bonusAmount, err := decimal.NewFromString(config.Value)
	if err != nil {
		return dto.SignUpBonusRes{}, err
	}

	return dto.SignUpBonusRes{
		Message: constant.SUCCESS,
		Amount:  bonusAmount,
	}, nil
}

func (u *User) RefreshTokenFlow(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	session, err := u.logsStorage.GetUserSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		u.log.Error("Invalid or expired refresh token", zap.Error(err))
		return "", "", time.Time{}, errors.ErrInvalidAccessToken.New("Invalid or expired refresh token")
	}

	err = u.logsStorage.InvalidateOldUserSessions(ctx, session.UserID, session.ID)
	if err != nil {
		u.log.Error("Failed to invalidate old sessions", zap.Error(err))
	}

	// Generate new access token
	accessToken, err := utils.GenerateJWT(session.UserID)
	if err != nil {
		u.log.Error("Failed to generate new access token", zap.Error(err))
		return "", "", time.Time{}, errors.ErrInternalServerError.Wrap(err, "Failed to generate new access token")
	}

	// Rotate refresh token
	newRefreshToken, err := utils.GenerateRefreshJWT(session.UserID)
	if err != nil {
		u.log.Error("Failed to generate new refresh token", zap.Error(err))
		return "", "", time.Time{}, errors.ErrInternalServerError.Wrap(err, "Failed to generate new refresh token")
	}
	newExpiry := time.Now().Add(30 * time.Minute)
	err = u.logsStorage.UpdateUserSessionRefreshToken(ctx, session.ID, newRefreshToken, newExpiry)
	if err != nil {
		u.log.Error("Failed to update refresh token in DB", zap.Error(err))
		return "", "", time.Time{}, errors.ErrInternalServerError.Wrap(err, "Failed to update refresh token in DB")
	}

	u.NotifySessionRefreshed(session.UserID)

	return accessToken, newRefreshToken, newExpiry, nil
}

func (u *User) AddSessionSocketConnection(userID uuid.UUID, conn *websocket.Conn) {
	u.sessionSocketMutex.Lock()
	defer u.sessionSocketMutex.Unlock()

	if _, exists := u.sessionSockets[userID]; !exists {
		u.sessionSockets[userID] = make(map[*websocket.Conn]bool)
	}
	u.sessionSockets[userID][conn] = true

	// Send connection confirmation
	conn.WriteJSON(dto.SessionEventMessage{
		Type:    "connected",
		Message: "Connected to session monitoring",
		Action:  "dismiss",
		Timeout: 3,
	})
}

func (u *User) RemoveSessionSocketConnection(userID uuid.UUID, conn *websocket.Conn) {
	u.sessionSocketMutex.Lock()
	defer u.sessionSocketMutex.Unlock()

	if conns, exists := u.sessionSockets[userID]; exists {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(u.sessionSockets, userID)
		}
	}
}

func (u *User) SendSessionEvent(userID uuid.UUID, event dto.SessionEventMessage) bool {
	u.sessionSocketMutex.RLock()
	conns, exists := u.sessionSockets[userID]
	u.sessionSocketMutex.RUnlock()

	if !exists || len(conns) == 0 {
		u.log.Info("No active session sockets for user", zap.String("userID", userID.String()))
		return false
	}

	msg, err := json.Marshal(event)
	if err != nil {
		u.log.Error("Failed to marshal session event", zap.Error(err))
		return false
	}

	delivered := false
	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			u.log.Warn("Failed to send session event to user", zap.Error(err), zap.String("userID", userID.String()))
			u.RemoveSessionSocketConnection(userID, conn)
			continue
		}
		delivered = true
	}
	return delivered
}

func (u *User) NotifySessionExpired(userID uuid.UUID) bool {
	u.log.Info("Attempting to send session expired notification",
		zap.String("userID", userID.String()))

	event := dto.SessionEventMessage{
		Type:    dto.WS_SESSION_EXPIRED,
		Message: "Your session has expired due to inactivity. Please login again to continue.",
		Action:  "login",
		Timeout: 5,
		Data: map[string]interface{}{
			"redirect_url": "/login?reason=session_expired",
		},
	}
	result := u.SendSessionEvent(userID, event)
	u.log.Info("Session expired notification sent",
		zap.String("userID", userID.String()),
		zap.Bool("delivered", result))
	return result
}

func (u *User) NotifySessionRefreshed(userID uuid.UUID) bool {
	event := dto.SessionEventMessage{
		Type:    dto.WS_SESSION_REFRESHED,
		Message: "Your session has been refreshed successfully.",
		Action:  "dismiss",
		Timeout: 3,
	}
	return u.SendSessionEvent(userID, event)
}

func (u *User) MonitorUserSessions(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	u.log.Info("Starting session monitoring service")

	for {
		select {
		case <-ticker.C:
			u.checkExpiringSessions(ctx)
		case <-ctx.Done():
			u.log.Info("Stopping session monitoring service")
			return
		case <-u.stopChan:
			u.log.Info("Stopping session monitoring service")
			return
		}
	}
}

func (u *User) checkExpiringSessions(ctx context.Context) {
	// Get sessions expiring within the next 5 minutes
	expiringSessions, err := u.logsStorage.GetSessionsExpiringSoon(ctx, 5*time.Minute)
	if err != nil {
		u.log.Error("Failed to get expiring sessions", zap.Error(err))
		return
	}

	u.log.Info("Found expiring sessions", zap.Int("count", len(expiringSessions)))

	for _, session := range expiringSessions {
		// Calculate minutes left until expiry
		timeLeft := time.Until(session.RefreshTokenExpiry)
		minutesLeft := int(timeLeft.Minutes())

		u.log.Info("Processing session",
			zap.String("userID", session.UserID.String()),
			zap.Duration("timeLeft", timeLeft),
			zap.Int("minutesLeft", minutesLeft))

		if minutesLeft <= 0 {
			// Session has expired, send expiry notification
			u.log.Info("Sending expiry notification", zap.String("userID", session.UserID.String()))
			u.NotifySessionExpired(session.UserID)
		}
	}
}

func (u *User) HandleSessionExpiry(ctx context.Context, userID uuid.UUID) error {
	err := u.logsStorage.InvalidateAllUserSessions(ctx, userID)
	if err != nil {
		u.log.Error("Failed to invalidate user sessions", zap.Error(err), zap.String("userID", userID.String()))
		return err
	}

	u.NotifySessionExpired(userID)

	u.log.Info("Session expired for user", zap.String("userID", userID.String()))
	return nil
}

func (u *User) Stop() {
	close(u.stopChan)
	u.log.Info("Session monitoring service stopped")
}

func (u *User) VerifyUser(ctx context.Context, req dto.VerifyPhoneNumberReq) (dto.UserRegisterResponse, string, error) {
	valid, err := u.redis.VerifyAndRemoveOTP(ctx, req.PhoneNumber, req.OTP)
	if err != nil {
		if err.Error() == "Too many invalid attempts. Please request a new OTP." {
			return dto.UserRegisterResponse{}, "", errors.ErrAcessError.New("Too many invalid attempts. Please request a new OTP.")
		}
		return dto.UserRegisterResponse{}, "", err
	}
	if !valid {
		err := errors.ErrAcessError.New("invalid otp")
		return dto.UserRegisterResponse{}, "", err
	}

	// get user by phone number
	usr, exist, err := u.userStorage.GetUserByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return dto.UserRegisterResponse{}, "", err
	}

	if !exist {
		err := fmt.Errorf("user not found")
		u.log.Error(err.Error(), zap.String("phoneNumber", req.PhoneNumber))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}

	// Check if user is already verified
	if usr.Status == constant.ACTIVE {
		err := fmt.Errorf("user already verified")
		u.log.Error(err.Error(), zap.String("phoneNumber", req.PhoneNumber))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}

	// get temp by user id from storage

	resp, err := u.userStorage.GetTempData(ctx, usr.ID)
	if err != nil {
		return dto.UserRegisterResponse{}, "", err
	}

	// Check if temp data exists

	if resp.ReferralData.ReferedByCode != "" {
		// // Handle referral bonus if user was referred

		err = u.processReferralBonus(ctx, resp.ReferralData.ReferedByCode, usr.ID)
		if err != nil {
			u.log.Error("Failed to process referral bonus", zap.Error(err), zap.String("referralCode", resp.ReferralData.ReferedByCode))
		}

	}

	if resp.ReferralData.ReferralCode != "" {
		u.UpdateUserPointByReferingUser(ctx, resp.ReferralData.ReferedByCode, usr.ID.String())
	}

	// generate jwt token  to the user
	token, err := utils.GenerateJWT(usr.ID)
	if err != nil {
		err = fmt.Errorf("unable to generate jwt token")
		u.log.Warn(err.Error(), zap.Any("registerReq", usr))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}

	refreshToken, err := utils.GenerateRefreshJWT(usr.ID)
	if err != nil {
		u.log.Error("unable to generate refresh token", zap.Error(err))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		return dto.UserRegisterResponse{}, "", err
	}
	refreshTokenExpiry := time.Now().Add(30 * time.Minute)
	userSession := dto.UserSessions{
		UserID:                usr.ID,
		Token:                 token,
		ExpiresAt:             time.Now().Add(time.Minute * 10),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiry,
		IpAddress:             "",
		UserAgent:             "",
	}

	// If gin.Context is present, set IP and UserAgent, and set cookie
	if c, ok := ctx.Value("gin-context").(*gin.Context); ok {
		userSession.IpAddress = c.ClientIP()
		userSession.UserAgent = c.GetHeader("User-Agent")
		u.logsStorage.CreateLoginSessions(ctx, userSession)
	} else {
		u.logsStorage.CreateLoginSessions(ctx, userSession)
	}

	if resp.ReferralData.AgentRequestID != "" {
		err = u.processAgentReferralConversion(ctx, resp.ReferralData.AgentRequestID, usr.ID, req.PhoneNumber)
		if err != nil {
			u.log.Error("Failed to process agent referral conversion", zap.Error(err), zap.String("agentRequestID", resp.ReferralData.AgentRequestID))
		}
	}

	return dto.UserRegisterResponse{
		Message:      constant.USER_REGISTRATION_SUCCESS,
		UserID:       resp.ID,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, refreshToken, nil

}

func (u *User) ReSendVerificationOTP(ctx context.Context, phoneNumber string) (*dto.ForgetPasswordRes, error) {
	// check if the user exist by phone number or email
	usr, exist, err := u.userStorage.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		u.log.Error(err.Error())
		return nil, err
	}

	if !exist {
		err = fmt.Errorf("user not found")
		u.log.Error(err.Error(), zap.String("phone", phoneNumber))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return nil, err
	}
	// generate otp
	otp := utils.GenerateOTP()
	// send otp to the user

	if usr.Status == constant.ACTIVE {
		err = fmt.Errorf("user already verified")
		u.log.Error(err.Error(), zap.String("phone", phoneNumber))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return nil, err
	}

	// send to risi
	u.PisiClient.SendBulkSMS(ctx, pisi.SendBulkSMSRequest{
		Message:    fmt.Sprintf("Your verification code is {%s}. Please do not share this code.", otp),
		Recipients: usr.PhoneNumber,
	})
	// save otp to the redis
	err = u.redis.SaveOTP(ctx, usr.PhoneNumber, otp)
	if err != nil {
		u.log.Error("Failed to save OTP", zap.Error(err), zap.String("phoneNumber", usr.PhoneNumber))
		return nil, errors.ErrInternalServerError.Wrap(err, "Failed to save OTP")
	}
	msg := &dto.ForgetPasswordRes{
		Message: constant.SUCCESS,
	}

	return msg, nil
}

func (u *User) GetOtp(ctx context.Context, phone string) string {
	otp, err := u.redis.GetOTP(context.Background(), phone)
	if err != nil {
		u.log.Error("Failed to get OTP", zap.Error(err), zap.String("phone", phone))
		return ""
	}
	if otp == "" {
		u.log.Warn("No OTP found for phone", zap.String("phone", phone))
		return ""
	}
	u.log.Info("Retrieved OTP for phone", zap.String("phone", phone), zap.String("otp", otp))
	return otp
}

// CheckUserExistsByEmail checks if a user with the given email already exists
func (u *User) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}

	_, exists, err := u.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		u.log.Error("failed to check email existence", zap.Error(err), zap.String("email", email))
		return false, err
	}
	return exists, nil
}

// CheckUserExistsByPhoneNumber checks if a user with the given phone number already exists
func (u *User) CheckUserExistsByPhoneNumber(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}

	_, exists, err := u.userStorage.GetUserByPhoneNumber(ctx, phone)
	if err != nil {
		u.log.Error("failed to check phone existence", zap.Error(err), zap.String("phone", phone))
		return false, err
	}
	return exists, nil
}

// CheckUserExistsByUsername checks if a user with the given username already exists
func (u *User) CheckUserExistsByUsername(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, nil
	}

	_, exists, err := u.userStorage.GetUserByUserName(ctx, username)
	if err != nil {
		u.log.Error("failed to check username existence", zap.Error(err), zap.String("username", username))
		return false, err
	}
	return exists, nil
}
