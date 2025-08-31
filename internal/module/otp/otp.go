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
	VerifyOTP(ctx context.Context, req *dto.OTPVerificationRequest) (*dto.OTPVerificationResponse, error)
	ResendOTP(ctx context.Context, email, userAgent, ipAddress string) (*dto.ResendOTPResponse, error)
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
	err = s.email.SendVerificationEmail(email, otpCode, otp.ID.String(), userID, expiresAt, userAgent, ipAddress)
	if err != nil {
		// Log error but don't fail the request
		s.logger.Error("Failed to send verification email",
			zap.String("email", email),
			zap.Error(err))
		// You might want to handle this differently in production
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

	// Get or create user
	user, exists, err := s.user.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var userID uuid.UUID
	var isNewUser bool

	if !exists {
		// Create new user
		newUser := dto.User{
			Email:        req.Email,
			Type:         "PLAYER",
			ReferralCode: s.generateReferralCode(),
		}

		createdUser, err := s.user.CreateUser(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		userID = createdUser.ID
		isNewUser = true

		// Send welcome email
		go func() {
			if err := s.email.SendWelcomeEmail(req.Email, ""); err != nil {
				s.logger.Error("Failed to send welcome email",
					zap.String("email", req.Email),
					zap.Error(err))
			}
		}()
	} else {
		// Use existing user
		userID = user.ID
		isNewUser = false

		// Update existing user's email verification status
		// This is critical for production - users must be marked as verified after successful OTP verification
		_, err = s.user.UpdateUserVerificationStatus(ctx, userID, true)
		if err != nil {
			s.logger.Error("Failed to update user email verification status",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("email", req.Email))

			// Don't fail the verification, but log the error for monitoring
			// In production, this should trigger an alert as it's a critical failure
			s.logger.Warn("OTP verification completed but email verification status update failed - this requires immediate attention",
				zap.String("user_id", userID.String()),
				zap.String("email", req.Email))
		} else {
			s.logger.Info("User email verification status updated successfully",
				zap.String("user_id", userID.String()),
				zap.String("email", req.Email))
		}

		s.logger.Info("Existing user verified email",
			zap.String("user_id", userID.String()),
			zap.String("email", req.Email))
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWTTokens(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Get current database time for consistency
	dbTime, err := s.storage.GetCurrentDBTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database time: %w", err)
	}

	s.logger.Info("OTP verified successfully",
		zap.String("email", req.Email),
		zap.String("user_id", userID.String()),
		zap.Bool("is_new_user", isNewUser))

	return &dto.OTPVerificationResponse{
		Message:      "Email verified successfully",
		IsVerified:   true,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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
