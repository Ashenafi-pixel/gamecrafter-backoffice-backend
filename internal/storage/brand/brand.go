package brand

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
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
	// var domain sql.NullString
	// if req.Domain != nil && *req.Domain != "" {
	// 	domain = sql.NullString{String: *req.Domain, Valid: true}
	// }
	//		type CreateBrandReq struct {
	//	    Name            string  `json:"name" validate:"required,min=1,max=255"`
	//	    Code            string  `json:"code" validate:"required,min=1,max=50"`
	//	    Domain          *string `json:"domain,omitempty" validate:"omitempty,max=255"`
	//	    IsActive        bool    `json:"is_active,omitempty"`
	//	    Description     string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	//	    Signature       string  `json:"signature,omitempty" validate:"omitempty,max=255"`
	//	    WebhookURL      string  `json:"webhook_url,omitempty" validate:"omitempty,max=255"`
	//	    IntegrationType string  `json:"integration_type,omitempty" validate:"omitempty,max=255"`
	//	    APIURL          string  `json:"api_url,omitempty" validate:"omitempty,max=255"`
	//	}
	var ID uuid.UUID
	query := `INSERT INTO brands (name, code, domain, is_active, description, signature, webhook_url, integration_type, api_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING  id,name, code, domain, is_active,created_at, updated_at`
	// brand, err := b.db.Queries.CreateBrand(ctx, db.CreateBrandParams{
	// 	Name:     req.Name,
	// 	Code:     req.Code,
	// 	Domain:   domain,
	// 	IsActive: req.IsActive,
	// })
	var brand dto.CreateBrandReq
	var createdAt, updatedAt time.Time
	err := b.db.GetPool().QueryRow(
		ctx,
		query,
		req.Name,
		req.Code,
		req.Domain,
		req.IsActive,
		req.Description,
		req.Signature,
		req.WebhookURL,
		req.IntegrationType,
		req.APIURL,
	).Scan(
		&ID,
		&brand.Name,
		&brand.Code,
		&brand.Domain,
		&brand.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		b.log.Error("unable to create brand", zap.Error(err), zap.Any("request", req))
		err = errors.ErrUnableTocreate.Wrap(err, "unable to create brand")
		return dto.CreateBrandRes{}, err
	}

	// var domainPtr *string
	// if req.Domain.Valid {
	// 	domainPtr = &req.Domain.String
	// }

	// var createdAt, updatedAt time.Time
	// if req.CreatedAt.Valid {
	// 	createdAt = req.CreatedAt.Time
	// }
	// if req.UpdatedAt.Valid {
	// 	updatedAt = req.UpdatedAt.Time
	// }

	return dto.CreateBrandRes{
		ID:        ID,
		Name:      brand.Name,
		Code:      brand.Code,
		Domain:    &req.Domain,
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

	// brands, err := b.db.Queries.GetBrands(ctx, db.GetBrandsParams{
	// 	Column1: search,
	// 	Column2: isActive,
	// 	Limit:   int32(req.PerPage),
	// 	Offset:  int32(offset),
	// })
	query := `
	SELECT 
    id, name, code, domain, is_active, webhook_url, api_url, created_at, updated_at,
    COUNT(*) OVER() AS total
	FROM brands
	WHERE 
    ($1::text IS NULL OR name ILIKE '%' || $1 || '%' OR code ILIKE '%' || $1 || '%') AND
    ($2::bool IS NULL OR is_active = $2)
	ORDER BY created_at DESC
	LIMIT $3 OFFSET $4`
	rows, err := b.db.GetPool().Query(ctx, query, search, isActive, req.PerPage, offset)
	if err != nil {
		b.log.Error("unable to get brands", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get brands")
		return dto.GetBrandsRes{}, err
	}
	defer rows.Close()

	type brandRow struct {
		ID         uuid.UUID
		Name       string
		Code       string
		Domain     sql.NullString
		IsActive   bool
		WebhookURL sql.NullString
		APIURL     sql.NullString
		CreatedAt  sql.NullTime
		UpdatedAt  sql.NullTime

		Total int64
	}

	var brands []brandRow

	for rows.Next() {
		var brand brandRow
		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Code,
			&brand.Domain,
			&brand.IsActive,
			&brand.WebhookURL,
			&brand.APIURL,
			&brand.CreatedAt,
			&brand.UpdatedAt,
			&brand.Total,
		)
		if err != nil {
			b.log.Error("unable to scan brand row", zap.Error(err))
			err = errors.ErrUnableToGet.Wrap(err, "unable to get brands")
			return dto.GetBrandsRes{}, err
		}
		brands = append(brands, brand)
	}

	if err := rows.Err(); err != nil {
		b.log.Error("error iterating brand rows", zap.Error(err))
		err = errors.ErrUnableToGet.Wrap(err, "unable to get brands")
		return dto.GetBrandsRes{}, err
	}

	if len(brands) == 0 {
		return dto.GetBrandsRes{
			Brands:      []dto.Brand{},
			TotalCount:  0,
			TotalPages:  0,
			CurrentPage: req.Page,
			PerPage:     req.PerPage,
		}, nil
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
			// If domain is valid (not NULL), return it even if it's an empty string
			domainPtr = &brand.Domain.String
		}
		fmt.Printf("Brand ID: %s, Domain: %v, Domain Valid: %t\n", brand.ID.String(), domainPtr, brand.Domain.Valid)
		// If brand.Domain.Valid is false, domainPtr remains nil (NULL in database)

		var createdAt, updatedAt time.Time
		if brand.CreatedAt.Valid {
			createdAt = brand.CreatedAt.Time
		}
		if brand.UpdatedAt.Valid {
			updatedAt = brand.UpdatedAt.Time
		}

		result.Brands[i] = dto.Brand{
			ID:         brand.ID,
			Name:       brand.Name,
			Code:       brand.Code,
			Domain:     domainPtr,
			WebhookURL: &brand.WebhookURL.String,
			APIURL:     &brand.APIURL.String,
			IsActive:   brand.IsActive,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
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

	// updatedBrand, err := b.db.Queries.UpdateBrand(ctx, db.UpdateBrandParams{
	// 	ID:      req.ID,
	// 	Column2: name,
	// 	Column3: code,
	// 	Column4: domain,
	// 	Column5: isActive,
	// })
	query := `UPDATE brands
			SET 
				name = COALESCE($2, name),
				code = COALESCE($3, code),
				domain = COALESCE($4, domain),
				is_active = COALESCE($5, is_active),
				webhook_url = COALESCE($6, webhook_url),
				api_url = COALESCE($7, api_url),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			RETURNING id, name, code, domain, is_active, webhook_url, api_url, created_at, updated_at`
	var createdAt, updatedAt time.Time
	err := b.db.GetPool().QueryRow(
		ctx,
		query,
		req.ID,
		name,
		code,
		domain,
		isActive,
		req.WebhookURL,
		req.APIURL,
	).Scan(
		&req.ID,
		&name,
		&code,
		&domain,
		&isActive,
		&req.WebhookURL,
		&req.APIURL,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		b.log.Error("unable to update brand", zap.Error(err), zap.String("id", req.ID.String()))
		err = errors.ErrUnableToUpdate.Wrap(err, "unable to update brand")
		return dto.UpdateBrandRes{}, err
	}

	return dto.UpdateBrandRes{
		ID:         req.ID,
		Name:       name.String,
		Code:       code.String,
		Domain:     &domain.String,
		IsActive:   isActive.Bool,
		WebhookURL: req.WebhookURL,
		APIURL:     req.APIURL,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
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
