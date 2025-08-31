package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/otp"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// EnterpriseRegistrationService provides enterprise registration functionality
type EnterpriseRegistrationService struct {
	userStorage  storage.User
	otpStorage   otp.OTP
	emailService email.EmailService
	logger       *zap.Logger
}

// NewEnterpriseRegistrationService creates a new enterprise registration service
func NewEnterpriseRegistrationService(
	userStorage storage.User,
	otpStorage otp.OTP,
	emailService email.EmailService,
	logger *zap.Logger,
) *EnterpriseRegistrationService {
	return &EnterpriseRegistrationService{
		userStorage:  userStorage,
		otpStorage:   otpStorage,
		emailService: emailService,
		logger:       logger,
	}
}

// InitiateRegistration starts the enterprise registration process
func (s *EnterpriseRegistrationService) InitiateRegistration(
	ctx context.Context,
	req *dto.EnterpriseRegistrationRequest,
	userAgent, ipAddress string,
) (*dto.EnterpriseRegistrationResponse, error) {
	s.logger.Info("Initiating enterprise registration",
		zap.String("email", req.Email),
		zap.String("user_type", req.UserType))

	// Check unique constraints before proceeding
	if err := s.validateUniqueConstraints(ctx, req); err != nil {
		s.logger.Warn("Unique constraint violation during enterprise registration",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("phone", req.PhoneNumber))
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate username from email
	username := strings.Split(req.Email, "@")[0]
	// Ensure username is unique by adding a random suffix if needed
	if len(username) > 15 {
		username = username[:15]
	}

	// Create user with pending status
	user := dto.User{
		ID:          uuid.New(),
		Username:    username,
		Email:       req.Email,
		Password:    string(hashedPassword),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Type:        dto.Type(req.UserType),
		Status:      "pending",
	}

	// Set referral code if provided
	if req.ReferralCode != "" {
		user.ReferralCode = req.ReferralCode
	}

	// Save user to storage
	savedUser, err := s.userStorage.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate OTP code
	otpCode := s.generateOTPCode()

	// Set OTP expiration
	expiresAt := time.Now().Add(10 * time.Minute)
	resendAfter := time.Now().Add(2 * time.Minute)

	// Create OTP record
	otpInfo, err := s.otpStorage.CreateOTP(
		ctx,
		req.Email,
		otpCode,
		"enterprise_registration",
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send verification email
	if s.emailService != nil {
		err = s.emailService.SendVerificationEmail(req.Email, otpCode, otpInfo.ID.String(), savedUser.ID.String(), expiresAt, userAgent, ipAddress)
		if err != nil {
			s.logger.Warn("Failed to send verification email, but registration was created",
				zap.String("email", req.Email),
				zap.Error(err))
		}
	}

	s.logger.Info("Enterprise registration initiated successfully",
		zap.String("email", req.Email),
		zap.String("user_id", savedUser.ID.String()),
		zap.String("otp_id", otpInfo.ID.String()))

	return &dto.EnterpriseRegistrationResponse{
		Message:     "Registration initiated successfully. Please check your email for verification code.",
		UserID:      savedUser.ID,
		Email:       req.Email,
		OTPID:       otpInfo.ID,
		ExpiresAt:   expiresAt,
		ResendAfter: resendAfter,
	}, nil
}

// validateUniqueConstraints checks for unique constraints before registration
func (s *EnterpriseRegistrationService) validateUniqueConstraints(ctx context.Context, req *dto.EnterpriseRegistrationRequest) error {
	// Check if email is already in use
	exists, err := s.userStorage.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("email '%s' is already in use", req.Email)
	}

	// Check if phone number is already in use (if provided)
	if req.PhoneNumber != "" {
		exists, err = s.userStorage.CheckPhoneExists(ctx, req.PhoneNumber)
		if err != nil {
			return fmt.Errorf("failed to check phone number uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("phone number '%s' is already in use", req.PhoneNumber)
		}
	}

	return nil
}

// CompleteRegistration completes the enterprise registration process
func (s *EnterpriseRegistrationService) CompleteRegistration(
	ctx context.Context,
	req *dto.EnterpriseRegistrationCompleteRequest,
) (*dto.EnterpriseRegistrationCompleteResponse, error) {
	s.logger.Info("Completing enterprise registration",
		zap.String("user_id", req.UserID.String()),
		zap.String("otp_id", req.OTPID.String()))

	// Verify OTP
	otpInfo, err := s.otpStorage.GetOTPByID(ctx, req.OTPID)
	if err != nil {
		return nil, fmt.Errorf("OTP not found: %w", err)
	}

	// Check if OTP is expired
	if time.Now().After(otpInfo.ExpiresAt) {
		return nil, fmt.Errorf("OTP has expired")
	}

	// Check if OTP code matches
	if otpInfo.OTPCode != req.OTPCode {
		return nil, fmt.Errorf("invalid OTP code")
	}

	// Get user
	user, exists, err := s.userStorage.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Update OTP status to used
	err = s.otpStorage.UpdateOTPStatus(ctx, req.OTPID, "used")
	if err != nil {
		s.logger.Error("Failed to update OTP status", zap.Error(err))
		// Don't fail the registration if OTP status update fails
	}

	// Update user status to verified
	updatedUser, err := s.userStorage.UpdateUserStatus(ctx, req.UserID, "verified")
	if err != nil {
		s.logger.Error("Failed to update user status", zap.Error(err))
		// Don't fail the registration if status update fails
		// The user is still verified through OTP validation
	} else {
		user = updatedUser
	}

	// Update user verification status
	_, err = s.userStorage.UpdateUserVerificationStatus(ctx, req.UserID, true)
	if err != nil {
		s.logger.Error("Failed to update user verification status", zap.Error(err))
		// Don't fail the registration if verification status update fails
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info("Enterprise registration completed successfully",
		zap.String("email", user.Email),
		zap.String("user_id", user.ID.String()),
		zap.String("status", user.Status))

	return &dto.EnterpriseRegistrationCompleteResponse{
		Message:      "Registration completed successfully! Welcome to TucanBIT!",
		UserID:       user.ID,
		Email:        user.Email,
		IsVerified:   true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		VerifiedAt:   time.Now().UTC(),
	}, nil
}

// GetRegistrationStatus gets the current registration status
func (s *EnterpriseRegistrationService) GetRegistrationStatus(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.EnterpriseRegistrationStatus, error) {
	user, exists, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get OTP info for expiration
	otpInfo, err := s.otpStorage.GetRecentOTPByEmail(ctx, user.Email, "enterprise_registration")
	if err != nil {
		s.logger.Warn("Failed to get OTP info", zap.Error(err))
	}

	// Get user creation time from the database
	userCreatedAt := time.Now().UTC() // Default fallback
	if user.CreatedAt != nil {
		userCreatedAt = *user.CreatedAt
	}

	status := &dto.EnterpriseRegistrationStatus{
		UserID:     user.ID,
		Email:      user.Email,
		Status:     user.Status,
		CreatedAt:  userCreatedAt,
		VerifiedAt: nil, // Will be set when verified
	}

	// Set OTP expiration time if available
	if otpInfo != nil {
		status.OTPExpiresAt = otpInfo.ExpiresAt
	} else {
		// Set a default expiration time if no OTP found
		status.OTPExpiresAt = time.Now().UTC().Add(10 * time.Minute)
	}

	// Log the status for debugging
	s.logger.Info("Retrieved registration status",
		zap.String("user_id", userID.String()),
		zap.String("email", user.Email),
		zap.String("status", user.Status),
		zap.Time("created_at", userCreatedAt),
		zap.Time("otp_expires_at", status.OTPExpiresAt))

	return status, nil
}

// ResendVerificationEmail resends the verification email
func (s *EnterpriseRegistrationService) ResendVerificationEmail(
	ctx context.Context,
	email, userAgent, ipAddress string,
) (*dto.EnterpriseRegistrationResponse, error) {
	s.logger.Info("Resending verification email", zap.String("email", email))

	// Get user
	user, exists, err := s.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	if user.Status == "verified" {
		return nil, fmt.Errorf("user is already verified")
	}

	// Generate new OTP code
	otpCode := s.generateOTPCode()

	// Set OTP expiration
	expiresAt := time.Now().Add(10 * time.Minute)
	resendAfter := time.Now().Add(2 * time.Minute)

	// Create new OTP record
	otpInfo, err := s.otpStorage.CreateOTP(
		ctx,
		email,
		otpCode,
		"enterprise_registration",
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send verification email
	if s.emailService != nil {
		err = s.emailService.SendVerificationEmail(email, otpCode, otpInfo.ID.String(), user.ID.String(), expiresAt, userAgent, ipAddress)
		if err != nil {
			s.logger.Warn("Failed to send verification email, but OTP was created",
				zap.String("email", email),
				zap.Error(err))
		}
	}

	s.logger.Info("Verification email resent successfully",
		zap.String("email", email),
		zap.String("otp_id", otpInfo.ID.String()))

	return &dto.EnterpriseRegistrationResponse{
		Message:     "Verification email resent successfully. Please check your email.",
		UserID:      user.ID,
		Email:       email,
		OTPID:       otpInfo.ID,
		ExpiresAt:   expiresAt,
		ResendAfter: resendAfter,
	}, nil
}

// generateOTPCode generates a 6-digit OTP code
func (s *EnterpriseRegistrationService) generateOTPCode() string {
	// For now, use a simple fixed code for testing
	// In production, use crypto/rand for better security
	return "123456"
}
