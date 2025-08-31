package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/module/otp"
	"go.uber.org/zap"
)

// RegistrationService handles enterprise-grade user registration with email verification
type RegistrationService struct {
	userModule   module.User
	otpModule    otp.OTPModule
	emailService email.EmailService
	logger       *zap.Logger
	redisClient  RedisClient
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
	redisClient RedisClient,
	logger *zap.Logger,
) *RegistrationService {
	return &RegistrationService{
		userModule:   userModule,
		otpModule:    otpModule,
		emailService: emailService,
		redisClient:  redisClient,
		logger:       logger,
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
	if err := c.ShouldBindJSON(&detailedReq); err == nil {
		// Handle detailed registration
		rs.handleDetailedRegistration(c, &detailedReq)
		return
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
	rs.handleSimpleRegistration(c, &simpleReq)
}

// handleDetailedRegistration processes detailed registration requests
func (rs *RegistrationService) handleDetailedRegistration(c *gin.Context, req *dto.DetailedUserRegistration) {
	// Validate required fields
	if err := rs.validateDetailedRegistrationRequest(req); err != nil {
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
		Type:            req.Type,
		ReferalType:     req.ReferalType,
		ReferedByCode:   req.ReferedByCode,
		ReferralCode:    req.ReferralCode,
		Username:        req.Username,
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
		Username:        req.Username,
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

	// Double-check unique constraints before creating user (prevents race conditions)
	if err := rs.validateUniqueConstraints(c.Request.Context(), &dto.User{
		Email:       registrationData.Email,
		PhoneNumber: registrationData.PhoneNumber,
	}); err != nil {
		rs.logger.Warn("Unique constraint violation during registration completion",
			zap.Error(err),
			zap.String("email", registrationData.Email),
			zap.String("phone", registrationData.PhoneNumber),
			zap.String("ip", c.ClientIP()))

		// Clean up stored registration data
		_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", req.UserID.String()))

		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Code:    http.StatusConflict,
			Message: "Registration failed: " + err.Error(),
		})
		return
	}

	// Create user account using existing user module
	userResponse, _, err := rs.userModule.RegisterUser(c.Request.Context(), dto.User{
		Email:           registrationData.Email,
		PhoneNumber:     registrationData.PhoneNumber,
		FirstName:       registrationData.FirstName,
		LastName:        registrationData.LastName,
		Password:        registrationData.Password,
		Type:            dto.Type(registrationData.Type),
		ReferalType:     dto.Type(registrationData.ReferalType),
		ReferedByCode:   registrationData.ReferedByCode,
		ReferralCode:    registrationData.ReferralCode,
		Username:        registrationData.Username,
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
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user account",
		})
		return
	}

	// Clean up stored registration data
	_ = rs.redisClient.Delete(c.Request.Context(), fmt.Sprintf("registration:%s", req.UserID.String()))

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
// @Param request body dto.ResendOTPRequest true "Resend verification request"
// @Success 200 {object} dto.ResendOTPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register/resend-verification [post]
func (rs *RegistrationService) ResendVerificationEmail(c *gin.Context) {
	var req dto.ResendOTPRequest
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

	// Resend OTP
	response, err := rs.otpModule.ResendOTP(c.Request.Context(), req.Email, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		rs.logger.Error("Failed to resend verification email",
			zap.Error(err),
			zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rs.logger.Info("Verification email resent successfully",
		zap.String("email", req.Email),
		zap.String("ip", c.ClientIP()))

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
	return nil
}

// validateDetailedRegistrationRequest validates the detailed registration request data
func (rs *RegistrationService) validateDetailedRegistrationRequest(req *dto.DetailedUserRegistration) error {
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
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("email '%s' is already in use", email)
	}

	// Check if phone number is already in use
	exists, err = rs.userModule.CheckUserExistsByPhoneNumber(ctx, phone)
	if err != nil {
		rs.logger.Error("Failed to check phone number uniqueness",
			zap.Error(err),
			zap.String("phone_number", phone))
		return fmt.Errorf("failed to check phone number uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("phone number '%s' is already in use", phone)
	}

	// Check if username is already in use (if provided)
	if username != "" {
		exists, err = rs.userModule.CheckUserExistsByUsername(ctx, username)
		if err != nil {
			rs.logger.Error("Failed to check username uniqueness",
				zap.Error(err),
				zap.String("username", username))
			return fmt.Errorf("failed to check username uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("username '%s' is already in use", username)
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
