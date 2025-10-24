package passkey

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type PasskeyCredential struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	CredentialID      string     `json:"credential_id"`
	RawID             []byte     `json:"raw_id"`
	PublicKey         []byte     `json:"public_key"`
	AttestationObject []byte     `json:"attestation_object"`
	ClientDataJSON    []byte     `json:"client_data_json"`
	Counter           int64      `json:"counter"`
	Name              string     `json:"name"`
	CreatedAt         time.Time  `json:"created_at"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	IsActive          bool       `json:"is_active"`
}

type PasskeyStorage interface {
	CreatePasskeyCredential(ctx context.Context, credential *PasskeyCredential) error
	GetPasskeyCredentialByID(ctx context.Context, credentialID string, userID uuid.UUID) (*PasskeyCredential, error)
	GetPasskeyCredentialsByUserID(ctx context.Context, userID uuid.UUID) ([]*PasskeyCredential, error)
	UpdatePasskeyCredentialCounter(ctx context.Context, credentialID string, userID uuid.UUID, counter int64) error
	DeletePasskeyCredential(ctx context.Context, credentialID string, userID uuid.UUID) error
	CheckPasskeyExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type passkeyStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func NewPasskeyStorage(database *persistencedb.PersistenceDB, log *zap.Logger) PasskeyStorage {
	return &passkeyStorage{
		db:  database,
		log: log,
	}
}

func (p *passkeyStorage) CreatePasskeyCredential(ctx context.Context, credential *PasskeyCredential) error {
	query := `
		INSERT INTO passkey_credentials (
			user_id, credential_id, raw_id, public_key, 
			attestation_object, client_data_json, counter, name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	var id uuid.UUID
	var createdAt time.Time

	err := p.db.GetPool().QueryRow(ctx, query,
		credential.UserID,
		credential.CredentialID,
		credential.RawID,
		credential.PublicKey,
		credential.AttestationObject,
		credential.ClientDataJSON,
		credential.Counter,
		credential.Name,
	).Scan(&id, &createdAt)

	if err != nil {
		p.log.Error("Failed to create passkey credential", zap.Error(err))
		return err
	}

	credential.ID = id
	credential.CreatedAt = createdAt

	p.log.Info("Passkey credential created successfully",
		zap.String("user_id", credential.UserID.String()),
		zap.String("credential_id", credential.CredentialID))

	return nil
}

func (p *passkeyStorage) GetPasskeyCredentialByID(ctx context.Context, credentialID string, userID uuid.UUID) (*PasskeyCredential, error) {
	query := `
		SELECT id, user_id, credential_id, raw_id, public_key, 
		       attestation_object, client_data_json, counter, name, 
		       created_at, last_used_at, is_active
		FROM passkey_credentials 
		WHERE credential_id = $1 AND user_id = $2 AND is_active = true`

	credential := &PasskeyCredential{}
	err := p.db.GetPool().QueryRow(ctx, query, credentialID, userID).Scan(
		&credential.ID,
		&credential.UserID,
		&credential.CredentialID,
		&credential.RawID,
		&credential.PublicKey,
		&credential.AttestationObject,
		&credential.ClientDataJSON,
		&credential.Counter,
		&credential.Name,
		&credential.CreatedAt,
		&credential.LastUsedAt,
		&credential.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		p.log.Error("Failed to get passkey credential", zap.Error(err))
		return nil, err
	}

	return credential, nil
}

func (p *passkeyStorage) GetPasskeyCredentialsByUserID(ctx context.Context, userID uuid.UUID) ([]*PasskeyCredential, error) {
	query := `
		SELECT id, user_id, credential_id, raw_id, public_key, 
		       attestation_object, client_data_json, counter, name, 
		       created_at, last_used_at, is_active
		FROM passkey_credentials 
		WHERE user_id = $1 AND is_active = true 
		ORDER BY created_at DESC`

	rows, err := p.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		p.log.Error("Failed to get passkey credentials", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var credentials []*PasskeyCredential
	for rows.Next() {
		credential := &PasskeyCredential{}
		err := rows.Scan(
			&credential.ID,
			&credential.UserID,
			&credential.CredentialID,
			&credential.RawID,
			&credential.PublicKey,
			&credential.AttestationObject,
			&credential.ClientDataJSON,
			&credential.Counter,
			&credential.Name,
			&credential.CreatedAt,
			&credential.LastUsedAt,
			&credential.IsActive,
		)
		if err != nil {
			p.log.Error("Failed to scan passkey credential", zap.Error(err))
			return nil, err
		}
		credentials = append(credentials, credential)
	}

	return credentials, nil
}

func (p *passkeyStorage) UpdatePasskeyCredentialCounter(ctx context.Context, credentialID string, userID uuid.UUID, counter int64) error {
	query := `
		UPDATE passkey_credentials 
		SET counter = $1, last_used_at = NOW() 
		WHERE credential_id = $2 AND user_id = $3 AND is_active = true`

	result, err := p.db.GetPool().Exec(ctx, query, counter, credentialID, userID)
	if err != nil {
		p.log.Error("Failed to update passkey credential counter", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		p.log.Warn("No passkey credential found to update",
			zap.String("credential_id", credentialID),
			zap.String("user_id", userID.String()))
	}

	return nil
}

func (p *passkeyStorage) DeletePasskeyCredential(ctx context.Context, credentialID string, userID uuid.UUID) error {
	query := `
		UPDATE passkey_credentials 
		SET is_active = false 
		WHERE credential_id = $1 AND user_id = $2`

	result, err := p.db.GetPool().Exec(ctx, query, credentialID, userID)
	if err != nil {
		p.log.Error("Failed to delete passkey credential", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		p.log.Warn("No passkey credential found to delete",
			zap.String("credential_id", credentialID),
			zap.String("user_id", userID.String()))
	}

	p.log.Info("Passkey credential deleted successfully",
		zap.String("credential_id", credentialID),
		zap.String("user_id", userID.String()))

	return nil
}

func (p *passkeyStorage) CheckPasskeyExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM passkey_credentials 
			WHERE user_id = $1 AND is_active = true
		)`

	var exists bool
	err := p.db.GetPool().QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		p.log.Error("Failed to check passkey existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}
