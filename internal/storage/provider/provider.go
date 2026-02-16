// internal/storage/provider/provider.go

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type provider struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(persistenceDB *persistencedb.PersistenceDB, log *zap.Logger) storage.Provider {
	return &provider{
		db:  persistenceDB,
		log: log,
	}
}

func (p *provider) CreateProvider(ctx context.Context, req dto.CreateProviderRequest) (*dto.GameProvider, error) {
	// Set defaults
	integrationType := req.IntegrationType
	if integrationType == "" {
		integrationType = "API"
	}

	status := "ACTIVE"
	if !req.IsActive {
		status = "INACTIVE"
	}

	query := `INSERT INTO game_providers (
		name, code, description, api_url, webhook_url, 
		integration_type, is_active, status, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	RETURNING id, name, code, description, api_url, webhook_url, 
		integration_type, is_active, status, created_at, updated_at`

	var id uuid.UUID
	var name, code, integrationTypeDB, statusDB string
	var description, apiURL, webhookURL *string
	var isActive bool
	var createdAt, updatedAt time.Time

	err := p.db.GetPool().QueryRow(ctx, query,
		req.Name,
		req.Code,
		req.Description,
		req.APIURL,
		req.WebhookURL,
		integrationType,
		req.IsActive,
		status,
	).Scan(
		&id, &name, &code, &description, &apiURL, &webhookURL,
		&integrationTypeDB, &isActive, &statusDB, &createdAt, &updatedAt,
	)

	if err != nil {
		p.log.Error("unable to create provider", zap.Error(err),
			zap.String("name", req.Name), zap.String("code", req.Code))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create provider")
		return nil, err
	}

	return &dto.GameProvider{
		ID:              id,
		Name:            name,
		Code:            code,
		Description:     description,
		APIURL:          apiURL,
		WebhookURL:      webhookURL,
		IntegrationType: integrationTypeDB,
		IsActive:        isActive,
		Status:          statusDB,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
func (p *provider) GetProvider(ctx context.Context, id uuid.UUID) (*dto.GameProvider, error) {
	query := `SELECT id, name, code, description, api_url, webhook_url, 
		integration_type, is_active, status, created_at, updated_at
	FROM game_providers WHERE id = $1`

	var provider dto.GameProvider
	err := p.db.GetPool().QueryRow(ctx, query, id).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Code,
		&provider.Description,
		&provider.APIURL,
		&provider.WebhookURL,
		&provider.IntegrationType,
		&provider.IsActive,
		&provider.Status,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		p.log.Error("unable to get provider", zap.Error(err), zap.String("id", id.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get provider")
		return nil, err
	}

	return &provider, nil
}
func (p *provider) GetAllProviders(ctx context.Context) ([]dto.GameProvider, error) {
	query := `SELECT id, name, code, description, api_url, webhook_url, 
		integration_type, is_active, status, created_at, updated_at
	FROM game_providers`

	rows, err := p.db.GetPool().Query(ctx, query)
	if err != nil {
		p.log.Error("unable to get providers", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get providers")
		return nil, err
	}
	defer rows.Close()

	var providers []dto.GameProvider
	for rows.Next() {
		var provider dto.GameProvider
		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Code,
			&provider.Description,
			&provider.APIURL,
			&provider.WebhookURL,
			&provider.IntegrationType,
			&provider.IsActive,
			&provider.Status,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			p.log.Error("error scanning provider row", zap.Error(err))
			continue // Skip this row and continue with others
		}
		providers = append(providers, provider)
	}

	if err = rows.Err(); err != nil {
		p.log.Error("error iterating provider rows", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "error iterating provider rows")
		return nil, err
	}

	return providers, nil
}
func (p *provider) UpdateProvider(ctx context.Context, req dto.UpdateProviderRequest) (*dto.GameProvider, error) {
	query := `UPDATE game_providers SET
    name = COALESCE($1, name),
    code = COALESCE($2, code),
    description = COALESCE($3, description),
    api_url = COALESCE($4, api_url),
    webhook_url = COALESCE($5, webhook_url),
    integration_type = COALESCE($6, integration_type),
    is_active = COALESCE($7, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $8
RETURNING id, name, code, description, api_url, webhook_url,
    integration_type, status, is_active, created_at, updated_at`

	var (
		id                                      uuid.UUID
		name, code, integrationTypeDB, statusDB string
		description, apiURL, webhookURL         *string
		isActive                                bool
		createdAt, updatedAt                    time.Time
	)

	err := p.db.GetPool().QueryRow(ctx, query,
		req.Name,
		req.Code,
		req.Description,
		req.APIURL,
		req.WebhookURL,
		req.IntegrationType,
		req.IsActive,
		req.ID,
	).Scan(
		&id,
		&name,
		&code,
		&description,
		&apiURL,
		&webhookURL,
		&integrationTypeDB,
		&statusDB,
		&isActive,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		p.log.Error("unable to update provider", zap.Error(err), zap.String("id", req.ID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update provider")
		fmt.Println("Error updating provider:", err) // Debug log
		return nil, err
	}

	return &dto.GameProvider{
		ID:              id,
		Name:            name,
		Code:            code,
		Description:     description,
		APIURL:          apiURL,
		WebhookURL:      webhookURL,
		IntegrationType: integrationTypeDB,
		IsActive:        isActive,
		Status:          statusDB,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil

}
func (p *provider) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	query := `DELETE FROM game_providers WHERE id = $1`

	cmdTag, err := p.db.GetPool().Exec(ctx, query, providerID)
	if err != nil {
		p.log.Error("unable to delete provider", zap.Error(err), zap.String("id", providerID.String()))
		err = errors.ErrUnableToDelete.Wrap(err, "unable to delete provider")
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		err = errors.ErrNoRecordFound.New("provider not found")
		p.log.Warn("provider not found for deletion", zap.String("id", providerID.String()))
		return err
	}

	return nil
}
