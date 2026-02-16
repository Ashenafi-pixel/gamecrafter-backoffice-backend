// internal/module/provider/provider.go

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type provider struct {
	log             *zap.Logger
	providerStorage storage.Provider
}

func Init(providerStorage storage.Provider, log *zap.Logger) module.Provider {
	return &provider{
		log:             log,
		providerStorage: providerStorage,
	}
}

func (p *provider) CreateProvider(ctx context.Context, req dto.CreateProviderRequest) (*dto.GameProvider, error) {
	// Validate request
	if err := ValidateCreateProviderRequest(req); err != nil {
		return nil, errors.ErrInvalidUserInput.Wrap(err, err.Error())
	}

	provider, err := p.providerStorage.CreateProvider(ctx, req)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// internal/constant/dto/provider.go - Add this function

func ValidateCreateProviderRequest(req dto.CreateProviderRequest) error {
	// Basic validation is handled by struct tags
	// Add any custom validation here if needed
	return nil
}
func (p *provider) GetAllProviders(ctx context.Context) ([]dto.GameProvider, error) {
	providers, err := p.providerStorage.GetAllProviders(ctx)
	if err != nil {
		p.log.Error("unable to get providers", zap.Error(err))
		return nil, errors.ErrInternalServerError.New("unable to get providers")
	}
	return providers, nil
}
func (p *provider) UpdateProvider(ctx context.Context, req dto.UpdateProviderRequest) (*dto.GameProvider, error) {
	// Validate request
	// if err := ValidateUpdateProviderRequest(req); err != nil {
	// 	return nil, errors.ErrInvalidUserInput.Wrap(err, err.Error())
	// }
	provider, err := p.providerStorage.UpdateProvider(ctx, req)
	if err != nil {
		p.log.Error("unable to update provider", zap.Error(err), zap.String("provider_id", req.ID.String()))
		return nil, errors.ErrInternalServerError.New("unable to update provider")
	}
	return provider, nil
}
func (p *provider) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	err := p.providerStorage.DeleteProvider(ctx, providerID)
	if err != nil {
		p.log.Error("unable to delete provider", zap.Error(err), zap.String("provider_id", providerID.String()))
		return errors.ErrInternalServerError.New("unable to delete provider")
	}
	return nil
}
