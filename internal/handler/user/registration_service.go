package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/cashback"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/module/groove"
	"github.com/tucanbit/internal/module/otp"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

// RegistrationService handles enterprise-grade user registration with email verification
type RegistrationService struct {
	userModule     module.User
	otpModule      otp.OTPModule
	emailService   email.EmailService
	balanceStorage storage.Balance
	cashbackModule *cashback.CashbackService
	grooveModule   groove.GrooveService
	logger         *zap.Logger
	redisClient    RedisClient
}

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewRegistrationService creates a new instance of RegistrationService
func NewRegistrationService(
	userModule module.User,
	otpModule otp.OTPModule,
	emailService email.EmailService,
	balanceStorage storage.Balance,
	cashbackModule *cashback.CashbackService,
	grooveModule groove.GrooveService,
	redisClient RedisClient,
	logger *zap.Logger,
) *RegistrationService {
	return &RegistrationService{
		userModule:     userModule,
		otpModule:      otpModule,
		emailService:   emailService,
		balanceStorage: balanceStorage,
		cashbackModule: cashbackModule,
		grooveModule:   grooveModule,
		redisClient:    redisClient,
		logger:         logger,
	}
}

// InitiateUserRegistration handles the initial registration request and sends email verification
// @Summary Initiate user registration
// @Description Start user registration process and send email verification OTP
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.DetailedUserRegistration true "User registration request"
// @Success 200 {object} dto.RegistrationPendingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (rs *RegistrationService) InitiateUserRegistration(c *gin.Context) {
	// Try to bind to detailed registration first
	var detailedReq dto.DetailedUserRegistration
	rs.logger.Info("InitiateUserRegistration called in registration service")
	if err := c.ShouldBindJSON(&detailedReq); err == nil {
		// Handle detailed registration
		rs.logger.Info("Detailed registration request received", zap.String("username", detailedReq.Username), zap.String("email", detailedReq.Email))
		rs.handleDetailedRegistration(c, &detailedReq)
		return
	} else {
		rs.logger.Info("Detailed registration binding failed, trying simple registration", zap.Error(err))
	}

	// Fall back to simple registration
	var simpleReq dto.User
	if err := c.ShouldBindJSON(&simpleReq); err != nil {
		rs.logger.Error("Failed to bind registration request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Handle simple registration
	rs.logger.Info("Simple registration request received", zap.String("username", simpleReq.Username), zap.String("email", simpleReq.Email))
	rs.handleSimpleRegistration(c, &simpleReq)
}

// handleDetailedRegistration processes detailed registration requests
func (rs *RegistrationService) handleDetailedRegistration(c *gin.Context, req *dto.DetailedUserRegistration) {
	// Validate required fields
	rs.logger.Info("Validating detailed registration request",
		zap.String("email", req.Email),
		zap.String("username", req.Username))

	if err := rs.validateDetailedRegistrationRequest(req); err != nil {
		rs.logger.Warn("Detailed registration validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rs.logger.Info("Detailed registration validation passed")

	// Check unique constraints before proceeding
	rs.logger.Info("Checking unique constraints for detailed registration",
		zap.String("email", req.Email),
		zap.String("username", req.Username))

	if err := rs.validateUniqueConstraints(c.Request.Context(), req); err != nil {
		rs.logger.Warn("Unique constraint violation during registration",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("phone", req.PhoneNumber),
			zap.String("ip", c.ClientIP()))

		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Code:    http.StatusConflict,
			Message: err.Error(),
		})
		return
	}

	rs.logger.Info("Unique constraints validation passed for detailed registration",
		zap.String("email", req.Email),
		zap.String("username", req.Username))

	// Generate unique user ID for temporary data storage
	userID := uuid.New()

	// Store registration data temporarily in Redis (24 hour expiration)
	registrationData := dto.RegistrationData{
		ID:              userID.String(),
		Username:        req.Username,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Password:        req.Password,
		Type:            req.Type,
		ReferalType:     req.ReferalType,
		ReferedByCode:   req.ReferedByCode,
		ReferralCode:    req.ReferralCode,
		City:            req.City,
		Country:         req.Country,
		State:           req.State,
		StreetAddress:   req.StreetAddress,
		PostalCode:      req.PostalCode,
		DateOfBirth:     req.DateOfBirth,
		DefaultCurrency: req.DefaultCurrency,
		KYCStatus:       req.KYCStatus,
		ProfilePicture:  req.ProfilePicture,
		AgentRequestID:  req.AgentRequestID,
		Accounts:        []dto.Account{}, // Convert from detailed accounts if needed
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(24 * time.Hour),
	}

	// Store in Redis
	var err error

	// Convert registration data to JSON for Redis storage
	registrationDataJSON, err := json.Marshal(registrationData)
	if err != nil {
		rs.logger.Error("Failed to marshal registration data",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process registration request",
		})
		return
	}

	err = rs.redisClient.Set(c.Request.Context(),
		fmt.Sprintf("registration:%s", userID.String()),
		string(registrationDataJSON),
		24*time.Hour)
	if err != nil {
		rs.logger.Error("Failed to store registration data",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process registration request",
		})
		return
	}

	// Store email-to-user-id mapping for resend OTP functionality
	err = rs.redisClient.Set(c.Request.Context(),
		fmt.Sprintf("email_to_user_id:%s", req.Email),
		userID.String(),
		24*time.Hour)
	if err != nil {
		rs.logger.Error("Failed to store email-to-user-id mapping",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("user_id", userID.String()))
		// Don't fail the registration for this, just log the error
	}

	// Send email verification OTP
	otpResponse, err := rs.otpModule.CreateEmailVerification(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP(), userID.String())
	if err != nil {
		rs.logger.Error("Failed to create email verification OTP",
			zap.Error(err),
			zap.String("email", req.Email))

		// Clean up stored data
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", userID.String()))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send verification email",
		})
		return
	}

	rs.logger.Info("User registration initiated successfully",
		zap.String("email", req.Email),
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()))

	// Return response indicating email verification is required
	c.JSON(http.StatusOK, dto.RegistrationPendingResponse{
		Message:     "Please check your email to verify your account and complete registration",
		UserID:      userID,
		Email:       req.Email,
		OTPID:       otpResponse.OTPID,
		ExpiresAt:   otpResponse.ExpiresAt.Format(time.RFC3339),
		ResendAfter: otpResponse.ResendAfter.Format(time.RFC3339),
	})
}

// handleSimpleRegistration processes simple registration requests
func (rs *RegistrationService) handleSimpleRegistration(c *gin.Context, req *dto.User) {
	// Validate required fields
	if err := rs.validateRegistrationRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Check unique constraints before proceeding
	if err := rs.validateUniqueConstraints(c.Request.Context(), req); err != nil {
		rs.logger.Warn("Unique constraint violation during registration",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("phone", req.PhoneNumber),
			zap.String("ip", c.ClientIP()))

		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Code:    http.StatusConflict,
			Message: err.Error(),
		})
		return
	}

	// Generate unique user ID for temporary data storage
	userID := uuid.New()

	// Store registration data temporarily in Redis (24 hour expiration)
	registrationData := dto.RegistrationData{
		ID:              userID.String(),
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Password:        req.Password,
		Type:            string(req.Type),
		ReferalType:     string(req.ReferalType),
		ReferedByCode:   req.ReferedByCode,
		ReferralCode:    req.ReferralCode,
		City:            req.City,
		Country:         req.Country,
		State:           req.State,
		StreetAddress:   req.StreetAddress,
		PostalCode:      req.PostalCode,
		DateOfBirth:     req.DateOfBirth,
		DefaultCurrency: req.DefaultCurrency,
		KYCStatus:       req.KYCStatus,
		ProfilePicture:  req.ProfilePicture,
		AgentRequestID:  req.AgentRequestID,
		Accounts:        []dto.Account{}, // Convert from dto.Balance if needed
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(24 * time.Hour),
	}

	// Store in Redis
	var err error

	// Convert registration data to JSON for Redis storage
	registrationDataJSON, err := json.Marshal(registrationData)
	if err != nil {
		rs.logger.Error("Failed to marshal registration data",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process registration request",
		})
		return
	}

	err = rs.redisClient.Set(c.Request.Context(),
		fmt.Sprintf("registration:%s", userID.String()),
		string(registrationDataJSON),
		24*time.Hour)
	if err != nil {
		rs.logger.Error("Failed to store registration data",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process registration request",
		})
		return
	}

	// Store email-to-user-id mapping for resend OTP functionality
	err = rs.redisClient.Set(c.Request.Context(),
		fmt.Sprintf("email_to_user_id:%s", req.Email),
		userID.String(),
		24*time.Hour)
	if err != nil {
		rs.logger.Error("Failed to store email-to-user-id mapping",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("user_id", userID.String()))
		// Don't fail the registration for this, just log the error
	}

	// Send email verification OTP
	otpResponse, err := rs.otpModule.CreateEmailVerification(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP(), userID.String())
	if err != nil {
		rs.logger.Error("Failed to create email verification OTP",
			zap.Error(err),
			zap.String("email", req.Email))

		// Clean up stored data
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", userID.String()))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send verification email",
		})
		return
	}

	rs.logger.Info("User registration initiated successfully",
		zap.String("email", req.Email),
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()))

	// Return response indicating email verification is required
	c.JSON(http.StatusOK, dto.RegistrationPendingResponse{
		Message:     "Please check your email to verify your account and complete registration",
		UserID:      userID,
		Email:       req.Email,
		OTPID:       otpResponse.OTPID,
		ExpiresAt:   otpResponse.ExpiresAt.Format(time.RFC3339),
		ResendAfter: otpResponse.ResendAfter.Format(time.RFC3339),
	})
}

// CompleteUserRegistration verifies OTP and creates the user account
// @Summary Complete user registration
// @Description Verify OTP and complete user registration process
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.CompleteRegistrationRequest true "Complete registration request"
// @Success 200 {object} dto.RegistrationCompleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register/complete [post]
func (rs *RegistrationService) CompleteUserRegistration(c *gin.Context) {
	var req dto.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rs.logger.Error("Failed to bind complete registration request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Retrieve stored registration data first to get email
	registrationData, err := rs.retrieveRegistrationData(c.Request.Context(), req.UserID.String())
	if err != nil {
		rs.logger.Error("Failed to retrieve registration data",
			zap.Error(err),
			zap.String("user_id", req.UserID.String()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Registration session expired or invalid",
		})
		return
	}

	// Verify OTP
	_, err = rs.otpModule.VerifyOTP(c.Request.Context(), &dto.OTPVerificationRequest{
		Email:   registrationData.Email,
		OTPCode: req.OTPCode,
		OTPID:   req.OTPID,
	})
	if err != nil {
		rs.logger.Error("Failed to verify OTP during registration completion",
			zap.Error(err),
			zap.String("email", registrationData.Email),
			zap.String("otp_id", req.OTPID.String()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Check if user already exists (in case of race conditions)
	exists, err := rs.userModule.CheckUserExistsByEmail(c.Request.Context(), registrationData.Email)
	if err != nil {
		rs.logger.Error("Failed to check existing user",
			zap.Error(err),
			zap.String("email", registrationData.Email))
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Code:    http.StatusServiceUnavailable,
			Message: "Service temporarily unavailable. Please try again later.",
		})
		return
	}

	if exists {
		rs.logger.Warn("User already exists during registration completion",
			zap.String("email", registrationData.Email),
			zap.String("ip", c.ClientIP()))

		// Clean up stored registration data
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", req.UserID.String()))

		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Code:    http.StatusConflict,
			Message: "Registration failed: email '" + registrationData.Email + "' is already in use",
		})
		return
	}

	// Create user account using existing user module
	rs.logger.Debug("Registration data before creating user", zap.String("username", registrationData.Username), zap.String("email", registrationData.Email))
	userResponse, _, err := rs.userModule.RegisterUser(c.Request.Context(), dto.User{
		Username:        registrationData.Username,
		Email:           registrationData.Email,
		PhoneNumber:     registrationData.PhoneNumber,
		FirstName:       registrationData.FirstName,
		LastName:        registrationData.LastName,
		Password:        registrationData.Password,
		Type:            dto.Type(registrationData.Type),
		ReferalType:     dto.Type(registrationData.ReferalType),
		ReferedByCode:   registrationData.ReferedByCode,
		ReferralCode:    registrationData.ReferralCode,
		City:            registrationData.City,
		Country:         registrationData.Country,
		State:           registrationData.State,
		StreetAddress:   registrationData.StreetAddress,
		PostalCode:      registrationData.PostalCode,
		DateOfBirth:     registrationData.DateOfBirth,
		DefaultCurrency: registrationData.DefaultCurrency,
		KYCStatus:       registrationData.KYCStatus,
		ProfilePicture:  registrationData.ProfilePicture,
		AgentRequestID:  registrationData.AgentRequestID,
		Accounts:        []dto.Balance{}, // Convert from Account if needed
	})
	if err != nil {
		rs.logger.Error("Failed to create user account",
			zap.Error(err),
			zap.String("email", registrationData.Email))
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Code:    http.StatusServiceUnavailable,
			Message: "Service temporarily unavailable. Please try again later.",
		})
		return
	}

	// Clean up stored registration data
	_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", req.UserID.String()))

	// Create initial wallet/balance for the user
	rs.logger.Info("Creating initial wallet for user",
		zap.String("user_id", userResponse.UserID.String()),
		zap.String("email", registrationData.Email))

	// Create balance using the balance storage
	_, err = rs.balanceStorage.CreateBalance(c.Request.Context(), dto.Balance{
		UserId:       userResponse.UserID,
		CurrencyCode: "USD", // Default currency
		RealMoney:    decimal.Zero,
		BonusMoney:   decimal.Zero,
		Points:       0,
	})
	if err != nil {
		// Check if it's a duplicate key error (wallet already exists)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			rs.logger.Info("Wallet already exists for user - this is expected in some cases",
				zap.String("user_id", userResponse.UserID.String()),
				zap.String("email", registrationData.Email))
		} else {
			rs.logger.Error("Failed to create initial wallet for user",
				zap.Error(err),
				zap.String("user_id", userResponse.UserID.String()),
				zap.String("email", registrationData.Email))

			// Don't fail the registration, but log the error for monitoring
			rs.logger.Warn("User registration completed but wallet creation failed - this requires immediate attention",
				zap.String("user_id", userResponse.UserID.String()),
				zap.String("email", registrationData.Email))
		}
	} else {
		rs.logger.Info("Initial wallet created successfully for user",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	}

	// Initialize user level (Bronze tier by default) for cashback system
	rs.logger.Info("Initializing user level for cashback system",
		zap.String("user_id", userResponse.UserID.String()),
		zap.String("email", registrationData.Email))

	err = rs.cashbackModule.InitializeUserLevel(c.Request.Context(), userResponse.UserID)
	if err != nil {
		rs.logger.Error("Failed to initialize user level",
			zap.Error(err),
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))

		// Don't fail the registration, but log the error for monitoring
		rs.logger.Warn("User registration completed but user level initialization failed - this requires immediate attention",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	} else {
		rs.logger.Info("User level initialized successfully",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	}

	// Create GrooveTech account for gaming functionality
	rs.logger.Info("Creating GrooveTech account for gaming functionality",
		zap.String("user_id", userResponse.UserID.String()),
		zap.String("email", registrationData.Email))

	_, err = rs.grooveModule.CreateAccount(c.Request.Context(), userResponse.UserID)
	if err != nil {
		rs.logger.Error("Failed to create GrooveTech account",
			zap.Error(err),
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))

		// Don't fail the registration, but log the error for monitoring
		rs.logger.Warn("User registration completed but GrooveTech account creation failed - this requires immediate attention",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	} else {
		rs.logger.Info("GrooveTech account created successfully",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	}

	// Update user's email verification status to verified
	// This is critical for production - users must be marked as verified after successful OTP verification
	_, err = rs.userModule.UpdateUserVerificationStatus(c.Request.Context(), userResponse.UserID, true)
	if err != nil {
		rs.logger.Error("Failed to update user email verification status",
			zap.Error(err),
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))

		// Don't fail the registration, but log the error for monitoring
		// In production, this should trigger an alert as it's a critical failure
		rs.logger.Warn("User registration completed but email verification status update failed - this requires immediate attention",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	} else {
		rs.logger.Info("User email verification status updated successfully",
			zap.String("user_id", userResponse.UserID.String()),
			zap.String("email", registrationData.Email))
	}

	rs.logger.Info("User registration completed successfully",
		zap.String("email", registrationData.Email),
		zap.String("user_id", userResponse.UserID.String()),
		zap.String("ip", c.ClientIP()))

	// Return success response with user data and tokens
	c.JSON(http.StatusOK, dto.RegistrationCompleteResponse{
		Message:      "Registration completed successfully! Welcome to TucanBIT!",
		UserID:       userResponse.UserID,
		AccessToken:  userResponse.AccessToken,
		RefreshToken: userResponse.RefreshToken,
		IsNewUser:    true,
		UserProfile: &dto.UserProfile{
			Username:     registrationData.Username,
			UserID:       userResponse.UserID,
			Email:        registrationData.Email,
			PhoneNumber:  registrationData.PhoneNumber,
			FirstName:    registrationData.FirstName,
			LastName:     registrationData.LastName,
			Type:         dto.Type(registrationData.Type),
			ReferralCode: registrationData.ReferralCode,
		},
	})
}

// ResendVerificationEmail resends the verification email for pending registrations
// @Summary Resend verification email
// @Description Resend verification email for pending user registration
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.ResendRegistrationOTPRequest true "Resend verification request"
// @Success 200 {object} dto.ResendRegistrationOTPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register/resend-verification [post]
func (rs *RegistrationService) ResendVerificationEmail(c *gin.Context) {
	var req dto.ResendRegistrationOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rs.logger.Error("Failed to bind resend verification request",
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
		})
		return
	}

	// Validate email
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Email is required",
		})
		return
	}

	// Find registration data in Redis by email
	registrationData, err := rs.findRegistrationDataByEmail(c.Request.Context(), req.Email)
	if err != nil {
		rs.logger.Error("Failed to find registration data",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Registration not found or expired",
		})
		return
	}

	// Check if registration data is expired
	if time.Now().After(registrationData.ExpiresAt) {
		rs.logger.Warn("Registration data expired",
			zap.String("email", req.Email),
			zap.Time("expires_at", registrationData.ExpiresAt))

		// Clean up expired data
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", registrationData.ID))
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("email_to_user_id:%s", req.Email))

		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Registration expired. Please register again.",
		})
		return
	}

	// Generate new OTP for the registration
	otpResponse, err := rs.otpModule.CreateEmailVerification(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP(), registrationData.ID)
	if err != nil {
		rs.logger.Error("Failed to create new email verification OTP",
			zap.Error(err),
			zap.String("email", req.Email))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send verification email",
		})
		return
	}

	rs.logger.Info("Verification email resent successfully",
		zap.String("email", req.Email),
		zap.String("user_id", registrationData.ID),
		zap.String("otp_id", otpResponse.OTPID.String()),
		zap.String("ip", c.ClientIP()))

	// Return response with new OTP details
	response := dto.ResendRegistrationOTPResponse{
		Message:     "Verification email resent successfully",
		UserID:      uuid.MustParse(registrationData.ID),
		Email:       req.Email,
		OTPID:       otpResponse.OTPID,
		ExpiresAt:   otpResponse.ExpiresAt.Format(time.RFC3339),
		ResendAfter: otpResponse.ResendAfter.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// validateRegistrationRequest validates the registration request data
func (rs *RegistrationService) validateRegistrationRequest(req *dto.User) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if req.Type == "" {
		return fmt.Errorf("user type is required")
	}

	// Validate email format
	if err := rs.validateEmailFormat(req.Email); err != nil {
		return err
	}

	// Validate username format if provided
	if req.Username != "" {
		if err := rs.validateUsernameFormat(req.Username); err != nil {
			return err
		}
	}

	// Validate password strength
	if err := rs.validatePasswordStrength(req.Password); err != nil {
		return err
	}

	return nil
}

// validateDetailedRegistrationRequest validates the detailed registration request data
func (rs *RegistrationService) validateDetailedRegistrationRequest(req *dto.DetailedUserRegistration) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Validate email format
	if err := rs.validateEmailFormat(req.Email); err != nil {
		return err
	}

	// Validate username format
	if err := rs.validateUsernameFormat(req.Username); err != nil {
		return err
	}

	// Validate password strength
	if err := rs.validatePasswordStrength(req.Password); err != nil {
		return err
	}

	// Set default currency to USD if not provided
	if req.DefaultCurrency == "" {
		req.DefaultCurrency = "USD"
	}

	// Set default user type to PLAYER if not provided
	if req.Type == "" {
		req.Type = "PLAYER"
	}

	return nil
}

// validateUniqueConstraints checks for unique constraints (e.g., email, phone number, username)
func (rs *RegistrationService) validateUniqueConstraints(ctx context.Context, req interface{}) error {
	var email, phone, username string

	// Extract email, phone, and username based on the request type
	switch r := req.(type) {
	case *dto.DetailedUserRegistration:
		email = r.Email
		phone = r.PhoneNumber
		username = r.Username
	case *dto.User:
		email = r.Email
		phone = r.PhoneNumber
		username = r.Username
	default:
		return fmt.Errorf("unsupported request type for validation")
	}

	// Check if email is already in use
	exists, err := rs.userModule.CheckUserExistsByEmail(ctx, email)
	if err != nil {
		rs.logger.Error("Failed to check email uniqueness",
			zap.Error(err),
			zap.String("email", email))
		// Don't expose internal database errors to users
		return fmt.Errorf("service temporarily unavailable")
	}
	if exists {
		return fmt.Errorf("email '%s' is already in use", email)
	}

	// Check if username is already in use
	if username != "" {
		exists, err = rs.userModule.CheckUserExistsByUsername(ctx, username)
		if err != nil {
			rs.logger.Error("Failed to check username uniqueness",
				zap.Error(err),
				zap.String("username", username))
			// Don't expose internal database errors to users
			return fmt.Errorf("service temporarily unavailable")
		}
		if exists {
			return fmt.Errorf("username '%s' is already taken", username)
		}
	}

	// Check if phone number is already in use (only if phone number is provided)
	if phone != "" {
		exists, err = rs.userModule.CheckUserExistsByPhoneNumber(ctx, phone)
		if err != nil {
			rs.logger.Error("Failed to check phone number uniqueness",
				zap.Error(err),
				zap.String("phone_number", phone))
			// Don't expose internal database errors to users
			return fmt.Errorf("service temporarily unavailable")
		}
		if exists {
			return fmt.Errorf("phone number '%s' is already in use", phone)
		}
	}

	return nil
}

// generateSessionID generates a unique session ID for temporary data storage
func (rs *RegistrationService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// retrieveRegistrationData retrieves registration data from Redis
func (rs *RegistrationService) retrieveRegistrationData(ctx context.Context, sessionID string) (*dto.RegistrationData, error) {
	key := fmt.Sprintf("registration:%s", sessionID)

	// Check if session exists
	exists, err := rs.redisClient.Exists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("registration session not found")
	}

	// Retrieve data
	data, err := rs.redisClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve registration data: %w", err)
	}

	// Parse JSON data
	var registrationData dto.RegistrationData
	if err := json.Unmarshal([]byte(data), &registrationData); err != nil {
		return nil, fmt.Errorf("failed to parse registration data: %w", err)
	}

	return &registrationData, nil
}

// validateEmailFormat validates email format using regex
func (rs *RegistrationService) validateEmailFormat(email string) error {
	// Email regex pattern that matches common email formats
	// This pattern allows for various valid email formats including:
	// - abcd1234@example.com
	// - abc@company.com
	// - emailtest@example.com
	// - user.name@domain.co.uk
	// - user+tag@example.org
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: '%s'", email)
	}

	// Additional length validation
	if len(email) > 254 {
		return fmt.Errorf("email address is too long (maximum 254 characters)")
	}

	return nil
}

// validateUsernameFormat validates username format
func (rs *RegistrationService) validateUsernameFormat(username string) error {
	// Username should be 3-30 characters long
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 30 {
		return fmt.Errorf("username must be no more than 30 characters long")
	}

	// Username should only contain alphanumeric characters, underscores, and hyphens
	// Must start and end with alphanumeric character
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens, and must start and end with a letter or number")
	}

	return nil
}

// validatePasswordStrength validates password strength for professional security standards
func (rs *RegistrationService) validatePasswordStrength(password string) error {
	// Minimum length requirement
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Maximum length requirement
	if len(password) > 128 {
		return fmt.Errorf("password must be no more than 128 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character (!@#$%%^&*()_+-=[]{}|;':\",./<>?~`)")
	}

	// Check for common weak patterns
	weakPatterns := []string{
		"password", "123456", "qwerty", "abc123", "password123",
		"admin", "user", "test", "guest", "root", "login",
	}

	for _, pattern := range weakPatterns {
		if regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pattern)).MatchString(password) {
			return fmt.Errorf("password contains common weak patterns and is not secure")
		}
	}

	// Check for repeated characters (more than 3 consecutive)
	repeatedChars := regexp.MustCompile(`(.)\\1{3,}`).MatchString(password)
	if repeatedChars {
		return fmt.Errorf("password cannot contain more than 3 consecutive identical characters")
	}

	// Check for sequential characters (like 123, abc, etc.)
	sequentialPatterns := []string{
		"123", "234", "345", "456", "567", "678", "789", "890",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij", "ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr", "qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
		"qwe", "wer", "ert", "rty", "tyu", "yui", "uio", "iop", "asd", "sdf", "dfg", "fgh", "ghj", "hjk", "jkl", "zxc", "xcv", "cvb", "vbn", "bnm",
	}

	lowerPassword := regexp.MustCompile(`[^a-z]`).ReplaceAllString(password, "")
	for _, pattern := range sequentialPatterns {
		if regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pattern)).MatchString(lowerPassword) {
			return fmt.Errorf("password contains sequential characters and is not secure")
		}
	}

	return nil
}

// findRegistrationDataByEmail finds registration data in Redis by email
func (rs *RegistrationService) findRegistrationDataByEmail(ctx context.Context, email string) (*dto.RegistrationData, error) {
	// Use email-to-user-id mapping to find the registration data
	emailToUserIDKey := fmt.Sprintf("email_to_user_id:%s", email)
	userIDStr, err := rs.redisClient.Get(ctx, emailToUserIDKey)
	if err != nil {
		return nil, fmt.Errorf("registration not found for email: %s", email)
	}

	// Get the registration data using the user ID
	registrationKey := fmt.Sprintf("registration:%s", userIDStr)
	registrationDataJSON, err := rs.redisClient.Get(ctx, registrationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get registration data: %w", err)
	}

	var registrationData dto.RegistrationData
	if err := json.Unmarshal([]byte(registrationDataJSON), &registrationData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registration data: %w", err)
	}

	return &registrationData, nil
}
