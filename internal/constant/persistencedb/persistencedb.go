package persistencedb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tucanbit/internal/constant/model/db"
	"go.uber.org/zap"
)

type PersistenceDB struct {
	*db.Queries
	pool *pgxpool.Pool
	log  *zap.Logger
}

type Sibling string

func New(pool *pgxpool.Pool, log *zap.Logger) PersistenceDB {
	return PersistenceDB{
		Queries: db.New(pool),
		pool:    pool,
		log:     log,
	}
}

// UpdateUserStatus updates the status of a user in the database
// This is a production-grade implementation for a world-class casino environment
func (p *PersistenceDB) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) (db.User, error) {
	query := `UPDATE users SET status = $1 WHERE id = $2 RETURNING id, username, phone_number, password, created_at, default_currency, profile, email, first_name, last_name, date_of_birth, source, user_type, referal_code, street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, status, refered_by_code, referal_type`

	row := p.pool.QueryRow(ctx, query, status, userID)

	var updatedUser db.User
	err := row.Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.PhoneNumber,
		&updatedUser.Password,
		&updatedUser.CreatedAt,
		&updatedUser.DefaultCurrency,
		&updatedUser.Profile,
		&updatedUser.Email,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.DateOfBirth,
		&updatedUser.Source,
		&updatedUser.UserType,
		&updatedUser.ReferalCode,
		&updatedUser.StreetAddress,
		&updatedUser.Country,
		&updatedUser.State,
		&updatedUser.City,
		&updatedUser.PostalCode,
		&updatedUser.KycStatus,
		&updatedUser.CreatedBy,
		&updatedUser.IsAdmin,
		&updatedUser.Status,
		&updatedUser.ReferedByCode,
		&updatedUser.ReferalType,
	)

	return updatedUser, err
}

// UpdateUserVerificationStatus updates the verification status of a user in the database
// This is a production-grade implementation for a world-class casino environment
func (p *PersistenceDB) UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, verified bool) (db.User, error) {
	query := `UPDATE users SET is_email_verified = $1 WHERE id = $2 RETURNING id, username, phone_number, password, created_at, default_currency, profile, email, first_name, last_name, date_of_birth, source, user_type, referal_code, street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, status, refered_by_code, referal_type`

	row := p.pool.QueryRow(ctx, query, verified, userID)

	var updatedUser db.User
	err := row.Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.PhoneNumber,
		&updatedUser.Password,
		&updatedUser.CreatedAt,
		&updatedUser.DefaultCurrency,
		&updatedUser.Profile,
		&updatedUser.Email,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.DateOfBirth,
		&updatedUser.Source,
		&updatedUser.UserType,
		&updatedUser.ReferalCode,
		&updatedUser.StreetAddress,
		&updatedUser.Country,
		&updatedUser.State,
		&updatedUser.City,
		&updatedUser.PostalCode,
		&updatedUser.KycStatus,
		&updatedUser.CreatedBy,
		&updatedUser.IsAdmin,
		&updatedUser.Status,
		&updatedUser.ReferedByCode,
		&updatedUser.ReferalType,
	)

	return updatedUser, err
}

// GetUserByEmailFull retrieves a complete user by email for login authentication
// This is a production-grade implementation for a world-class casino environment
func (p *PersistenceDB) GetUserByEmailFull(ctx context.Context, email string) (db.User, error) {
	query := `SELECT id, username, phone_number, password, created_at, default_currency, profile, email, 
	          first_name, last_name, date_of_birth, source, is_email_verified, referal_code, 
	          street_address, country, state, city, postal_code, kyc_status, created_by, is_admin, 
	          status, referal_type, refered_by_code, user_type, primary_wallet_address, 
	          wallet_verification_status FROM users WHERE email = $1`

	row := p.pool.QueryRow(ctx, query, email)

	var usr db.User
	err := row.Scan(
		&usr.ID,
		&usr.Username,
		&usr.PhoneNumber,
		&usr.Password,
		&usr.CreatedAt,
		&usr.DefaultCurrency,
		&usr.Profile,
		&usr.Email,
		&usr.FirstName,
		&usr.LastName,
		&usr.DateOfBirth,
		&usr.Source,
		&usr.IsEmailVerified,
		&usr.ReferalCode,
		&usr.StreetAddress,
		&usr.Country,
		&usr.State,
		&usr.City,
		&usr.PostalCode,
		&usr.KycStatus,
		&usr.CreatedBy,
		&usr.IsAdmin,
		&usr.Status,
		&usr.ReferalType,
		&usr.ReferedByCode,
		&usr.UserType,
		&usr.PrimaryWalletAddress,
		&usr.WalletVerificationStatus,
	)

	return usr, err
}
