package adds

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type adds struct {
	log *zap.Logger
	db  *persistencedb.PersistenceDB
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Adds {
	return &adds{
		log: log,
		db:  db,
	}
}

// SaveAddsService saves a new adds service configuration
func (a *adds) SaveAddsService(ctx context.Context, req dto.CreateAddsServiceReq) (dto.CreateAddsServiceRes, error) {
	var service dto.AddsServiceResData
	result, err := a.db.Queries.SaveAddsService(ctx, db.SaveAddsServiceParams{
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		ServiceID:     req.ServiceID,
		ServiceSecret: req.ServiceSecret,
		Status:        req.Status,
		CreatedBy:     req.CreatedBy,
		ServiceUrl:    req.ServiceURL,
	})

	if err != nil {
		a.log.Error("error saving adds service", zap.Error(err))
		err = errors.ErrUnableTocreate.Wrap(err, "error saving adds service")
		return dto.CreateAddsServiceRes{}, err
	}

	service = dto.AddsServiceResData{
		ID:          result.ID,
		Name:        result.Name,
		Description: result.Description.String,
		ServiceID:   result.ServiceID,
		Status:      result.Status,
		ServiceURL:  result.ServiceUrl,
		CreatedBy:   result.CreatedBy,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}
	return dto.CreateAddsServiceRes{
		Message: "Adds service saved successfully",
		Data:    service,
	}, nil
}

// GetAddsServiceByServiceID retrieves an adds service by service ID
func (a *adds) GetAddsServiceByServiceID(ctx context.Context, serviceID string) (dto.AddsServiceResData, bool, error) {
	service, err := a.db.Queries.GetAddsServiceByServiceID(ctx, serviceID)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error("error fetching adds service by service ID", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "error fetching adds service by service ID")
		return dto.AddsServiceResData{}, false, err
	}

	if err == sql.ErrNoRows {
		return dto.AddsServiceResData{}, false, nil
	}

	return dto.AddsServiceResData{
		ID:            service.ID,
		Name:          service.Name,
		Description:   service.Description.String,
		ServiceSecret: service.ServiceSecret,
		ServiceID:     service.ServiceID,
		Status:        service.Status,
		ServiceURL:    service.ServiceUrl,
		CreatedBy:     service.CreatedBy,
		CreatedAt:     service.CreatedAt,
		UpdatedAt:     service.UpdatedAt,
	}, true, nil
}

// GetAddsServiceByID retrieves an adds service by ID
func (a *adds) GetAddsServiceByID(ctx context.Context, id uuid.UUID) (dto.AddsServiceResData, bool, error) {
	service, err := a.db.Queries.GetAddsServiceByID(ctx, id)
	if err != nil && err.Error() != dto.ErrNoRows {
		a.log.Error("error fetching adds service by ID", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "error fetching adds service by ID")
		return dto.AddsServiceResData{}, false, err
	}

	if err == sql.ErrNoRows {
		return dto.AddsServiceResData{}, false, nil
	}

	return dto.AddsServiceResData{
		ID:          service.ID,
		Name:        service.Name,
		Description: service.Description.String,
		ServiceID:   service.ServiceID,
		Status:      service.Status,
		ServiceURL:  service.ServiceUrl,
		CreatedBy:   service.CreatedBy,
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
	}, true, nil
}

// GetAddsServices retrieves all adds services with pagination
func (a *adds) GetAddsServices(ctx context.Context, req dto.GetAddServicesRequest) (dto.GetAddsServicesRes, error) {
	var services []dto.AddsServiceResData
	results, err := a.db.Queries.GetAddsServices(ctx, db.GetAddsServicesParams{
		Limit:  int32(req.PerPage),
		Offset: int32(req.Page),
	})

	if err != nil && err != sql.ErrNoRows {
		a.log.Error("error fetching adds services", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "error fetching adds services")
		return dto.GetAddsServicesRes{}, err
	}

	if err == sql.ErrNoRows {
		return dto.GetAddsServicesRes{
			Message: "No adds services found",
			Data:    []dto.AddsServiceResData{},
		}, nil
	}

	for _, service := range results {
		services = append(services, dto.AddsServiceResData{
			ID:          service.ID,
			Name:        service.Name,
			Description: service.Description.String,
			ServiceID:   service.ServiceID,
			Status:      service.Status,
			CreatedBy:   service.CreatedBy,
			CreatedAt:   service.CreatedAt,
			UpdatedAt:   service.UpdatedAt,
			ServiceURL:  service.ServiceUrl,
		})
	}

	return dto.GetAddsServicesRes{
		Message: "Adds services fetched successfully",
		Data:    services,
	}, nil
}
