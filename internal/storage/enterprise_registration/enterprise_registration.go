package enterprise_registration

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"go.uber.org/zap"
)

// Database interface for enterprise registration storage
type Database interface {
	CreateRegistration(ctx context.Context, req *dto.EnterpriseRegistrationRequest) (*dto.EnterpriseRegistration, error)
	GetRegistrationByID(ctx context.Context, id uuid.UUID) (*dto.EnterpriseRegistration, error)
	GetRegistrationByEmail(ctx context.Context, email string) (*dto.EnterpriseRegistration, error)
	GetRegistrationByUserID(ctx context.Context, userID uuid.UUID) (*dto.EnterpriseRegistration, error)
	UpdateRegistrationStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *time.Time) error
	UpdateOTP(ctx context.Context, id uuid.UUID, otp string, expiresAt time.Time) error
	IncrementVerificationAttempts(ctx context.Context, id uuid.UUID) error
	GetRegistrationStats(ctx context.Context) (*dto.EnterpriseRegistrationStats, error)
	DeleteExpiredRegistrations(ctx context.Context, before time.Time) (int64, error)
}

// EnterpriseRegistration represents the database model (using DTO struct)
type EnterpriseRegistration = dto.EnterpriseRegistration

// EnterpriseRegistrationDatabase implements the Database interface
type EnterpriseRegistrationDatabase struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewEnterpriseRegistrationDatabase creates a new enterprise registration database
func NewEnterpriseRegistrationDatabase(db *sql.DB, logger *zap.Logger) Database {
	return &EnterpriseRegistrationDatabase{
		db:     db,
		logger: logger,
	}
}

// CreateRegistration creates a new enterprise registration
func (d *EnterpriseRegistrationDatabase) CreateRegistration(ctx context.Context, req *dto.EnterpriseRegistrationRequest) (*dto.EnterpriseRegistration, error) {
	// Generate new user ID for enterprise registration
	userID := uuid.New()

	query := `
		INSERT INTO enterprise_registrations (
			user_id, email, first_name, last_name, user_type, phone_number, company_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`

	var registration EnterpriseRegistration
	err := d.db.QueryRowContext(ctx, query,
		userID,
		req.Email,
		req.FirstName,
		req.LastName,
		req.UserType,
		req.PhoneNumber,
		req.CompanyName,
	).Scan(
		&registration.ID,
		&registration.UserID,
		&registration.Email,
		&registration.FirstName,
		&registration.LastName,
		&registration.UserType,
		&registration.PhoneNumber,
		&registration.CompanyName,
		&registration.RegistrationStatus,
		&registration.VerificationOTP,
		&registration.OTPExpiresAt,
		&registration.VerificationAttempts,
		&registration.MaxVerificationAttempts,
		&registration.EmailVerifiedAt,
		&registration.PhoneVerifiedAt,
		&registration.CreatedAt,
		&registration.UpdatedAt,
		&registration.VerifiedAt,
		&registration.RejectedAt,
		&registration.RejectionReason,
		&registration.Metadata,
	)

	if err != nil {
		d.logger.Error("Failed to create enterprise registration",
			zap.Error(err),
			zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create enterprise registration: %w", err)
	}

	d.logger.Info("Created enterprise registration",
		zap.String("id", registration.ID.String()),
		zap.String("email", registration.Email))

	return &registration, nil
}

// GetRegistrationByID retrieves a registration by ID
func (d *EnterpriseRegistrationDatabase) GetRegistrationByID(ctx context.Context, id uuid.UUID) (*dto.EnterpriseRegistration, error) {
	query := `SELECT * FROM enterprise_registrations WHERE id = $1`

	var registration EnterpriseRegistration
	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&registration.ID,
		&registration.UserID,
		&registration.Email,
		&registration.FirstName,
		&registration.LastName,
		&registration.UserType,
		&registration.PhoneNumber,
		&registration.CompanyName,
		&registration.RegistrationStatus,
		&registration.VerificationOTP,
		&registration.OTPExpiresAt,
		&registration.VerificationAttempts,
		&registration.MaxVerificationAttempts,
		&registration.EmailVerifiedAt,
		&registration.PhoneVerifiedAt,
		&registration.CreatedAt,
		&registration.UpdatedAt,
		&registration.VerifiedAt,
		&registration.RejectedAt,
		&registration.RejectionReason,
		&registration.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("enterprise registration not found: %w", err)
		}
		d.logger.Error("Failed to get enterprise registration by ID",
			zap.Error(err),
			zap.String("id", id.String()))
		return nil, fmt.Errorf("failed to get enterprise registration: %w", err)
	}

	return &registration, nil
}

// GetRegistrationByEmail retrieves a registration by email
func (d *EnterpriseRegistrationDatabase) GetRegistrationByEmail(ctx context.Context, email string) (*dto.EnterpriseRegistration, error) {
	query := `SELECT * FROM enterprise_registrations WHERE email = $1`

	var registration EnterpriseRegistration
	err := d.db.QueryRowContext(ctx, query, email).Scan(
		&registration.ID,
		&registration.UserID,
		&registration.Email,
		&registration.FirstName,
		&registration.LastName,
		&registration.UserType,
		&registration.PhoneNumber,
		&registration.CompanyName,
		&registration.RegistrationStatus,
		&registration.VerificationOTP,
		&registration.OTPExpiresAt,
		&registration.VerificationAttempts,
		&registration.MaxVerificationAttempts,
		&registration.EmailVerifiedAt,
		&registration.PhoneVerifiedAt,
		&registration.CreatedAt,
		&registration.UpdatedAt,
		&registration.VerifiedAt,
		&registration.RejectedAt,
		&registration.RejectionReason,
		&registration.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("enterprise registration not found: %w", err)
		}
		d.logger.Error("Failed to get enterprise registration by email",
			zap.Error(err),
			zap.String("email", email))
		return nil, fmt.Errorf("failed to get enterprise registration: %w", err)
	}

	return &registration, nil
}

// GetRegistrationByUserID retrieves a registration by user ID
func (d *EnterpriseRegistrationDatabase) GetRegistrationByUserID(ctx context.Context, userID uuid.UUID) (*dto.EnterpriseRegistration, error) {
	query := `SELECT * FROM enterprise_registrations WHERE user_id = $1`

	var registration EnterpriseRegistration
	err := d.db.QueryRowContext(ctx, query, userID).Scan(
		&registration.ID,
		&registration.UserID,
		&registration.Email,
		&registration.FirstName,
		&registration.LastName,
		&registration.UserType,
		&registration.PhoneNumber,
		&registration.CompanyName,
		&registration.RegistrationStatus,
		&registration.VerificationOTP,
		&registration.OTPExpiresAt,
		&registration.VerificationAttempts,
		&registration.MaxVerificationAttempts,
		&registration.EmailVerifiedAt,
		&registration.PhoneVerifiedAt,
		&registration.CreatedAt,
		&registration.UpdatedAt,
		&registration.VerifiedAt,
		&registration.RejectedAt,
		&registration.RejectionReason,
		&registration.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("enterprise registration not found: %w", err)
		}
		d.logger.Error("Failed to get enterprise registration by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get enterprise registration: %w", err)
	}

	return &registration, nil
}

// UpdateRegistrationStatus updates the registration status
func (d *EnterpriseRegistrationDatabase) UpdateRegistrationStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *time.Time) error {
	query := `
		UPDATE enterprise_registrations 
		SET registration_status = $2, verified_at = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := d.db.ExecContext(ctx, query, id, status, verifiedAt)
	if err != nil {
		d.logger.Error("Failed to update enterprise registration status",
			zap.Error(err),
			zap.String("id", id.String()),
			zap.String("status", status))
		return fmt.Errorf("failed to update registration status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no enterprise registration found with ID: %s", id.String())
	}

	d.logger.Info("Updated enterprise registration status",
		zap.String("id", id.String()),
		zap.String("status", status))

	return nil
}

// UpdateOTP updates the verification OTP and expiration
func (d *EnterpriseRegistrationDatabase) UpdateOTP(ctx context.Context, id uuid.UUID, otp string, expiresAt time.Time) error {
	query := `
		UPDATE enterprise_registrations 
		SET verification_otp = $2, otp_expires_at = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := d.db.ExecContext(ctx, query, id, otp, expiresAt)
	if err != nil {
		d.logger.Error("Failed to update enterprise registration OTP",
			zap.Error(err),
			zap.String("id", id.String()))
		return fmt.Errorf("failed to update OTP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no enterprise registration found with ID: %s", id.String())
	}

	d.logger.Info("Updated enterprise registration OTP",
		zap.String("id", id.String()))

	return nil
}

// IncrementVerificationAttempts increments the verification attempts counter
func (d *EnterpriseRegistrationDatabase) IncrementVerificationAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE enterprise_registrations 
		SET verification_attempts = verification_attempts + 1, updated_at = NOW()
		WHERE id = $1
	`

	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		d.logger.Error("Failed to increment verification attempts",
			zap.Error(err),
			zap.String("id", id.String()))
		return fmt.Errorf("failed to increment verification attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no enterprise registration found with ID: %s", id.String())
	}

	d.logger.Info("Incremented verification attempts",
		zap.String("id", id.String()))

	return nil
}

// GetRegistrationStats retrieves registration statistics
func (d *EnterpriseRegistrationDatabase) GetRegistrationStats(ctx context.Context) (*dto.EnterpriseRegistrationStats, error) {
	query := `
		SELECT 
			registration_status,
			user_type,
			COUNT(*) as count
		FROM enterprise_registrations
		GROUP BY registration_status, user_type
	`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		d.logger.Error("Failed to get enterprise registration stats",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get registration stats: %w", err)
	}
	defer rows.Close()

	stats := &dto.EnterpriseRegistrationStats{
		StatusCounts: make(map[string]map[string]int),
		TotalCount:   0,
	}

	for rows.Next() {
		var status, userType string
		var count int

		if err := rows.Scan(&status, &userType, &count); err != nil {
			d.logger.Error("Failed to scan registration stats row",
				zap.Error(err))
			continue
		}

		if stats.StatusCounts[status] == nil {
			stats.StatusCounts[status] = make(map[string]int)
		}
		stats.StatusCounts[status][userType] = count
		stats.TotalCount += count
	}

	if err := rows.Err(); err != nil {
		d.logger.Error("Error iterating registration stats rows",
			zap.Error(err))
		return nil, fmt.Errorf("error iterating stats rows: %w", err)
	}

	return stats, nil
}

// DeleteExpiredRegistrations deletes expired registrations
func (d *EnterpriseRegistrationDatabase) DeleteExpiredRegistrations(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM enterprise_registrations 
		WHERE created_at < $1 AND registration_status = 'PENDING'
	`

	result, err := d.db.ExecContext(ctx, query, before)
	if err != nil {
		d.logger.Error("Failed to delete expired enterprise registrations",
			zap.Error(err),
			zap.Time("before", before))
		return 0, fmt.Errorf("failed to delete expired registrations: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	d.logger.Info("Deleted expired enterprise registrations",
		zap.Int64("count", rowsAffected),
		zap.Time("before", before))

	return rowsAffected, nil
}
