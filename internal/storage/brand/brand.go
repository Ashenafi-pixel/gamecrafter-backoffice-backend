package brand

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type brand struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) storage.Brand {
	return &brand{
		db:  db,
		log: log,
	}
}

func (b *brand) CreateBrand(ctx context.Context, req dto.CreateBrandReq) (dto.CreateBrandRes, error) {
	var domain sql.NullString
	if req.Domain != nil && *req.Domain != "" {
		domain = sql.NullString{String: *req.Domain, Valid: true}
	}

	brand, err := b.db.Queries.CreateBrand(ctx, db.CreateBrandParams{
		Name:     req.Name,
		Code:     req.Code,
		Domain:   domain,
		IsActive: req.IsActive,
	})

	if err != nil {
		b.log.Error("unable to create brand", zap.Error(err), zap.Any("request", req))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create brand")
		return dto.CreateBrandRes{}, err
	}

	var domainPtr *string
	if brand.Domain.Valid {
		domainPtr = &brand.Domain.String
	}

	var createdAt, updatedAt time.Time
	if brand.CreatedAt.Valid {
		createdAt = brand.CreatedAt.Time
	}
	if brand.UpdatedAt.Valid {
		updatedAt = brand.UpdatedAt.Time
	}

	return dto.CreateBrandRes{
		ID:        brand.ID,
		Name:      brand.Name,
		Code:      brand.Code,
		Domain:    domainPtr,
		IsActive:  brand.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (b *brand) GetBrandByID(ctx context.Context, id uuid.UUID) (dto.Brand, bool, error) {
	brandRow, err := b.db.Queries.GetBrandByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Brand{}, false, nil
		}
		b.log.Error("unable to get brand by ID", zap.Error(err), zap.String("id", id.String()))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get brand by ID")
		return dto.Brand{}, false, err
	}

	var domainPtr *string
	if brandRow.Domain.Valid {
		domainPtr = &brandRow.Domain.String
	}

	var createdAt, updatedAt time.Time
	if brandRow.CreatedAt.Valid {
		createdAt = brandRow.CreatedAt.Time
	}
	if brandRow.UpdatedAt.Valid {
		updatedAt = brandRow.UpdatedAt.Time
	}

	return dto.Brand{
		ID:        brandRow.ID,
		Name:      brandRow.Name,
		Code:      brandRow.Code,
		Domain:    domainPtr,
		IsActive:  brandRow.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, true, nil
}

func (b *brand) GetBrands(ctx context.Context, req dto.GetBrandsReq) (dto.GetBrandsRes, error) {
	offset := (req.Page - 1) * req.PerPage

	var search sql.NullString
	if req.Search != "" {
		search = sql.NullString{String: req.Search, Valid: true}
	}

	var isActive sql.NullBool
	if req.IsActive != nil {
		isActive = sql.NullBool{Bool: *req.IsActive, Valid: true}
	}

	brands, err := b.db.Queries.GetBrands(ctx, db.GetBrandsParams{
		Column1: search,
		Column2: isActive,
		Limit:   int32(req.PerPage),
		Offset:  int32(offset),
	})

	if err != nil {
		b.log.Error("unable to get brands", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get brands")
		return dto.GetBrandsRes{}, err
	}

	var result dto.GetBrandsRes
	var total int

	if len(brands) > 0 {
		total = int(brands[0].Total)
	}

	result.TotalCount = total
	result.CurrentPage = req.Page
	result.TotalPages = int(math.Ceil(float64(total) / float64(req.PerPage)))
	result.PerPage = req.PerPage

		result.Brands = make([]dto.Brand, len(brands))
		for i, brand := range brands {
			var domainPtr *string
			if brand.Domain.Valid {
				domainPtr = &brand.Domain.String
			}

			var createdAt, updatedAt time.Time
			if brand.CreatedAt.Valid {
				createdAt = brand.CreatedAt.Time
			}
			if brand.UpdatedAt.Valid {
				updatedAt = brand.UpdatedAt.Time
			}

			result.Brands[i] = dto.Brand{
				ID:        brand.ID,
				Name:      brand.Name,
				Code:      brand.Code,
				Domain:    domainPtr,
				IsActive:  brand.IsActive,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
		}

	return result, nil
}

func (b *brand) UpdateBrand(ctx context.Context, req dto.UpdateBrandReq) (dto.UpdateBrandRes, error) {
	var name sql.NullString
	var code sql.NullString
	var domain sql.NullString
	var isActive sql.NullBool

	if req.Name != nil {
		name = sql.NullString{String: *req.Name, Valid: true}
	}
	if req.Code != nil {
		code = sql.NullString{String: *req.Code, Valid: true}
	}
	if req.Domain != nil {
		if *req.Domain != "" {
			domain = sql.NullString{String: *req.Domain, Valid: true}
		} else {
			domain = sql.NullString{Valid: false}
		}
	}
	if req.IsActive != nil {
		isActive = sql.NullBool{Bool: *req.IsActive, Valid: true}
	}

	updatedBrand, err := b.db.Queries.UpdateBrand(ctx, db.UpdateBrandParams{
		ID:       req.ID,
		Column2:  name,
		Column3:  code,
		Column4:  domain,
		Column5:  isActive,
	})

	if err != nil {
		b.log.Error("unable to update brand", zap.Error(err), zap.String("id", req.ID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update brand")
		return dto.UpdateBrandRes{}, err
	}

	var domainPtr *string
	if updatedBrand.Domain.Valid {
		domainPtr = &updatedBrand.Domain.String
	}

	var createdAt, updatedAt time.Time
	if updatedBrand.CreatedAt.Valid {
		createdAt = updatedBrand.CreatedAt.Time
	}
	if updatedBrand.UpdatedAt.Valid {
		updatedAt = updatedBrand.UpdatedAt.Time
	}

	return dto.UpdateBrandRes{
		ID:        updatedBrand.ID,
		Name:      updatedBrand.Name,
		Code:      updatedBrand.Code,
		Domain:    domainPtr,
		IsActive:  updatedBrand.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (b *brand) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	err := b.db.Queries.DeleteBrand(ctx, id)
	if err != nil {
		b.log.Error("unable to delete brand", zap.Error(err), zap.String("id", id.String()))
		err = errors.ErrDBDelError.Wrap(err, "unable to delete brand")
		return err
	}

	b.log.Info("brand deleted successfully", zap.String("id", id.String()))
	return nil
}

