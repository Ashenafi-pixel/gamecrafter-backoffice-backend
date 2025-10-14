package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/email"
	"github.com/tucanbit/internal/storage/otp"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

// OTPModule defines the interface for OTP operations
type OTPModule interface {
	CreateEmailVerification(ctx context.Context, email, userAgent, ipAddress string, userID string) (*dto.EmailVerificationResponse, error)
	CreatePasswordResetOTP(ctx context.Context, email, userAgent, ipAddress string, userID string) (*dto.EmailVerificationResponse, error)
	VerifyOTP(ctx context.Context, req *dto.OTPVerificationRequest) (*dto.OTPVerificationResponse, error)
	ResendOTP(ctx context.Context, email, userAgent, ipAddress string) (*dto.ResendOTPResponse, error)
	ResendPasswordResetOTP(ctx context.Context, email, userAgent, ipAddress string) (*dto.ResendOTPResponse, error)
	InvalidateOTP(ctx context.Context, otpID uuid.UUID) error
	CleanupExpiredOTPs(ctx context.Context) error
	GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error)
}

// OTPService provides comprehensive OTP management functionality
type OTPService struct {
	storage otp.OTP
	user    UserStorage
	logger  *zap.Logger
	email   email.EmailService
}

// UserStorage defines the interface for user storage operations
type UserStorage interface {
	GetUserByEmail(ctx context.Context, email string) (dto.User, bool, error)
	CreateUser(ctx context.Context, user dto.User) (dto.User, error)
	UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, isVerified bool) (dto.User, error)
}

// Ensure OTPService implements OTPModule
var _ OTPModule = (*OTPService)(nil)

// NewOTPService creates a new instance of OTPService
func NewOTPService(storage otp.OTP, user UserStorage, email email.EmailService, logger *zap.Logger) *OTPService {
	return &OTPService{
		storage: storage,
		user:    user,
		logger:  logger,
		email:   email,
	}
}

// CreateEmailVerification creates a new email verification OTP
func (s *OTPService) CreateEmailVerification(ctx context.Context, email, userAgent, ipAddress string, userID string) (*dto.EmailVerificationResponse, error) {
	// Check if user already exists
	existingUser, exists, err := s.user.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// If user exists, check if we should allow verification (for now, allow all)
	// In production, you might want to check if email is already verified
	if exists {
		s.logger.Info("User already exists, allowing email verification",
			zap.String("email", email),
			zap.String("user_id", existingUser.ID.String()))
	}

	// Generate OTP code
	otpCode := s.generateOTPCode()

	// Get current database time for consistency
	dbTime, err := s.storage.GetCurrentDBTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database time: %w", err)
	}

	expiresAt := dbTime.Add(10 * time.Minute)
	resendAfter := dbTime.Add(2 * time.Minute)
	// Create OTP record
	otp, err := s.storage.CreateOTP(ctx, email, otpCode, string(dto.OTPTypeEmailVerification), expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send verification email
	if s.email != nil {
		err = s.email.SendVerificationEmail(email, otpCode, otp.ID.String(), userID, expiresAt, userAgent, ipAddress)
		if err != nil {
			// Log error but don't fail the request
			s.logger.Error("Failed to send verification email",
				zap.String("email", email),
				zap.Error(err))
			// You might want to handle this differently in production
		}
	} else {
		s.logger.Warn("Email service is not available, skipping email sending",
			zap.String("email", email),
			zap.String("otp_code", otpCode))
	}

	s.logger.Info("Email verification OTP created",
		zap.String("email", email),
		zap.String("otp_id", otp.ID.String()))

	return &dto.EmailVerificationResponse{
		Message:     "Verification email sent successfully",
		OTPID:       otp.ID,
		Email:       email,
		ExpiresAt:   expiresAt,
		ResendAfter: resendAfter,
	}, nil
}

// CreatePasswordResetOTP creates a new password reset OTP
func (s *OTPService) CreatePasswordResetOTP(ctx context.Context, email, userAgent, ipAddress string, userID string) (*dto.EmailVerificationResponse, error) {
	// Check if user exists
	existingUser, exists, err := s.user.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("user with email %s does not exist", email)
	}

	// Generate OTP code
	otpCode := s.generateOTPCode()

	// Get current database time for consistency
	dbTime, err := s.storage.GetCurrentDBTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database time: %w", err)
	}

	expiresAt := dbTime.Add(10 * time.Minute)
	resendAfter := dbTime.Add(2 * time.Minute)

	// Create OTP record for password reset
	otp, err := s.storage.CreateOTP(ctx, email, otpCode, string(dto.OTPTypePasswordReset), expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send password reset email
	if s.email != nil {
		err = s.email.SendPasswordResetOTPEmail(email, otpCode, otp.ID.String(), existingUser.ID.String(), expiresAt, userAgent, ipAddress)
		if err != nil {
			// Log error but don't fail the request
			s.logger.Error("Failed to send password reset email",
				zap.String("email", email),
				zap.Error(err))
			// You might want to handle this differently in production
		}
	} else {
		s.logger.Warn("Email service is not available, skipping password reset email sending",
			zap.String("email", email),
			zap.String("otp_code", otpCode))
	}

	s.logger.Info("Password reset OTP created",
		zap.String("email", email),
		zap.String("otp_id", otp.ID.String()),
		zap.String("user_id", existingUser.ID.String()))

	return &dto.EmailVerificationResponse{
		Message:     "Password reset email sent successfully",
		OTPID:       otp.ID,
		Email:       email,
		ExpiresAt:   expiresAt,
		ResendAfter: resendAfter,
	}, nil
}

// VerifyOTP verifies an OTP code
func (s *OTPService) VerifyOTP(ctx context.Context, req *dto.OTPVerificationRequest) (*dto.OTPVerificationResponse, error) {
	// Get OTP record
	otp, err := s.storage.GetOTPByID(ctx, req.OTPID)
	if err != nil {
		return nil, fmt.Errorf("OTP not found: %w", err)
	}

	// Verify email matches
	if otp.Email != req.Email {
		return nil, fmt.Errorf("email mismatch")
	}

	// Check if OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		// Mark as expired
		_ = s.storage.UpdateOTPStatus(ctx, req.OTPID, string(dto.OTPStatusExpired))
		return nil, fmt.Errorf("OTP has expired")
	}

	// Check if OTP is already used
	if string(otp.Status) == string(dto.OTPStatusUsed) || string(otp.Status) == string(dto.OTPStatusVerified) {
		return nil, fmt.Errorf("OTP already used")
	}

	// Verify OTP code
	if otp.OTPCode != req.OTPCode {
		// Increment failed attempts (you might want to implement rate limiting)
		s.logger.Warn("Invalid OTP attempt",
			zap.String("email", req.Email),
			zap.String("otp_id", req.OTPID.String()))
		return nil, fmt.Errorf("invalid OTP code")
	}

	// Mark OTP as verified
	err = s.storage.UpdateOTPStatus(ctx, req.OTPID, string(dto.OTPStatusVerified))
	if err != nil {
		return nil, fmt.Errorf("failed to update OTP status: %w", err)
	}

	// For registration flow, we don't create users here
	// Users are created only after successful OTP verification in registration completion
	// This prevents duplicate user creation issues

	// Check if this is a registration OTP verification
	// If so, we just verify the OTP without creating/updating user
	s.logger.Info("OTP verified successfully for registration flow",
		zap.String("email", req.Email),
		zap.String("otp_id", req.OTPID.String()))

	// For registration flow, we don't generate JWT tokens here
	// Tokens are generated after user creation in registration completion
	s.logger.Info("OTP verification completed, ready for user creation in registration completion")

	// Get current database time for consistency
	dbTime, err := s.storage.GetCurrentDBTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database time: %w", err)
	}

	s.logger.Info("OTP verified successfully for registration flow",
		zap.String("email", req.Email),
		zap.String("otp_id", req.OTPID.String()))

	return &dto.OTPVerificationResponse{
		Message:      "Email verified successfully",
		IsVerified:   true,
		UserID:       uuid.Nil, // No user created yet
		AccessToken:  "",       // No token generated yet
		RefreshToken: "",       // No token generated yet
		VerifiedAt:   dbTime,
	}, nil
}

// ResendOTP resends an OTP to the specified email
func (s *OTPService) ResendOTP(ctx context.Context, email, userAgent, ipAddress string) (*dto.ResendOTPResponse, error) {
	// Check if there's a recent OTP that hasn't expired
	recentOTP, err := s.storage.GetRecentOTPByEmail(ctx, email, string(dto.OTPTypeEmailVerification))
	if err == nil && recentOTP != nil {
		// Get current database time for consistency
		dbTime, err := s.storage.GetCurrentDBTime(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get database time: %w", err)
		}

		// Check if we can resend (2 minutes cooldown)
		if dbTime.Before(recentOTP.CreatedAt.Add(2 * time.Minute)) {
			timeUntilResend := recentOTP.CreatedAt.Add(2 * time.Minute).Sub(dbTime)
			return nil, fmt.Errorf("please wait %v before requesting another OTP", timeUntilResend.Round(time.Second))
		}
	}

	// Create new OTP
	response, err := s.CreateEmailVerification(ctx, email, userAgent, ipAddress, "")
	if err != nil {
		return nil, err
	}

	return &dto.ResendOTPResponse{
		Message:     response.Message,
		OTPID:       response.OTPID,
		Email:       response.Email,
		ExpiresAt:   response.ExpiresAt,
		ResendAfter: response.ResendAfter,
	}, nil
}

// ResendPasswordResetOTP resends a password reset OTP to the specified email
func (s *OTPService) ResendPasswordResetOTP(ctx context.Context, email, userAgent, ipAddress string) (*dto.ResendOTPResponse, error) {
	// Check if there's a recent password reset OTP that hasn't expired
	recentOTP, err := s.storage.GetRecentOTPByEmail(ctx, email, string(dto.OTPTypePasswordReset))
	if err == nil && recentOTP != nil {
		// Get current database time for consistency
		dbTime, err := s.storage.GetCurrentDBTime(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get database time: %w", err)
		}

		// Check if we can resend (2 minutes cooldown)
		if dbTime.Before(recentOTP.CreatedAt.Add(2 * time.Minute)) {
			timeUntilResend := recentOTP.CreatedAt.Add(2 * time.Minute).Sub(dbTime)
			return nil, fmt.Errorf("please wait %v before requesting another password reset OTP", timeUntilResend.Round(time.Second))
		}
	}

	// Check if user exists
	existingUser, exists, err := s.user.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("user with email %s does not exist", email)
	}

	// Create new password reset OTP
	response, err := s.CreatePasswordResetOTP(ctx, email, userAgent, ipAddress, existingUser.ID.String())
	if err != nil {
		return nil, err
	}

	return &dto.ResendOTPResponse{
		Message:     response.Message,
		OTPID:       response.OTPID,
		Email:       response.Email,
		ExpiresAt:   response.ExpiresAt,
		ResendAfter: response.ResendAfter,
	}, nil
}

// InvalidateOTP marks an OTP as invalid
func (s *OTPService) InvalidateOTP(ctx context.Context, otpID uuid.UUID) error {
	return s.storage.UpdateOTPStatus(ctx, otpID, string(dto.OTPStatusUsed))
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (s *OTPService) CleanupExpiredOTPs(ctx context.Context) error {
	err := s.storage.DeleteExpiredOTPs(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}

	s.logger.Info("Expired OTPs cleaned up successfully")
	return nil
}

// GetOTPByID retrieves an OTP by ID
func (s *OTPService) GetOTPByID(ctx context.Context, otpID uuid.UUID) (*dto.OTPInfo, error) {
	otp, err := s.storage.GetOTPByID(ctx, otpID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP by ID: %w", err)
	}

	return otp, nil
}

// generateOTPCode generates a 6-digit OTP code
func (s *OTPService) generateOTPCode() string {
	// Generate a random 6-digit number
	const min = 100000
	const max = 999999

	// Use crypto/rand for better randomness
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to time-based if crypto/rand fails
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}

	// Convert random bytes to number
	randomNum := new(big.Int).SetBytes(randomBytes)
	otpNum := min + (randomNum.Uint64() % uint64(max-min+1))

	return fmt.Sprintf("%06d", otpNum)
}

// generateReferralCode generates a unique referral code
func (s *OTPService) generateReferralCode() string {
	// Generate a random 8-character referral code using crypto/rand for better randomness
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)

	// Use crypto/rand for better randomness
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to time-based if crypto/rand fails
		for i := range b {
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
	} else {
		// Use the random bytes to generate the referral code
		for i := range b {
			b[i] = charset[int(randomBytes[i])%len(charset)]
		}
	}

	return string(b)
}

// generateTemporaryPassword generates a temporary password for new users
func (s *OTPService) generateTemporaryPassword() string {
	// Generate a random temporary password
	// In production, you might want to use a more secure approach
	return "TempPass123!"
}

// generateJWTTokens generates JWT access and refresh tokens
func (s *OTPService) generateJWTTokens(userID uuid.UUID) (string, string, error) {
	// Generate real JWT access token
	accessToken, err := utils.GenerateJWT(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate real refresh token
	refreshToken, err := utils.GenerateRefreshJWT(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
