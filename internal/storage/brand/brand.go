package brand

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/persistencedb"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	var ID int32
	// Ensure the sequence exists before inserting
	_, _ = b.db.GetPool().Exec(ctx, `CREATE SEQUENCE IF NOT EXISTS brand_id_seq START WITH 100000 INCREMENT BY 1 MINVALUE 100000 MAXVALUE 999999`)
	
	query := `INSERT INTO brands (id, name, code, domain, is_active, description, signature, webhook_url, integration_type, api_url)
VALUES (nextval('brand_id_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9)
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

func (b *brand) GetBrandByID(ctx context.Context, id int32) (dto.Brand, bool, error) {
	query := `SELECT id, name, code, domain, is_active, webhook_url, api_url, created_at, updated_at
		FROM brands
		WHERE id = $1`
	
	var brandRow struct {
		ID         int32
		Name       string
		Code       string
		Domain     sql.NullString
		IsActive   bool
		WebhookURL sql.NullString
		APIURL     sql.NullString
		CreatedAt  sql.NullTime
		UpdatedAt  sql.NullTime
	}
	
	err := b.db.GetPool().QueryRow(ctx, query, id).Scan(
		&brandRow.ID,
		&brandRow.Name,
		&brandRow.Code,
		&brandRow.Domain,
		&brandRow.IsActive,
		&brandRow.WebhookURL,
		&brandRow.APIURL,
		&brandRow.CreatedAt,
		&brandRow.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.Brand{}, false, nil
		}
		b.log.Error("unable to get brand by ID", zap.Error(err), zap.Int32("id", id))
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

	var webhookURL *string
	if brandRow.WebhookURL.Valid {
		webhookURL = &brandRow.WebhookURL.String
	}

	var apiURL *string
	if brandRow.APIURL.Valid {
		apiURL = &brandRow.APIURL.String
	}

	return dto.Brand{
		ID:         brandRow.ID,
		Name:       brandRow.Name,
		Code:       brandRow.Code,
		Domain:     domainPtr,
		IsActive:   brandRow.IsActive,
		WebhookURL: webhookURL,
		APIURL:     apiURL,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
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
		ID         int32
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
		fmt.Printf("Brand ID: %d, Domain: %v, Domain Valid: %t\n", brand.ID, domainPtr, brand.Domain.Valid)
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
		b.log.Error("unable to update brand", zap.Error(err), zap.Int32("id", req.ID))
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

func (b *brand) DeleteBrand(ctx context.Context, id int32) error {
	query := `DELETE FROM brands WHERE id = $1`
	_, err := b.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		b.log.Error("unable to delete brand", zap.Error(err), zap.Int32("id", id))
		err = errors.ErrDBDelError.Wrap(err, "unable to delete brand")
		return err
	}

	b.log.Info("brand deleted successfully", zap.Int32("id", id))
	return nil
}

func (b *brand) UpdateBrandStatus(ctx context.Context, brandID int32, isActive bool) error {
	query := `UPDATE brands SET is_active = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := b.db.GetPool().Exec(ctx, query, brandID, isActive)
	if err != nil {
		b.log.Error("unable to update brand status", zap.Error(err), zap.Int32("id", brandID))
		return errors.ErrUnableToUpdate.Wrap(err, "unable to update brand status")
	}
	return nil
}

func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func (b *brand) CreateBrandCredential(ctx context.Context, brandID int32, req dto.CreateBrandCredentialReq) (dto.BrandCredentialRes, string, error) {
	name := "Default"
	if req.Name != "" {
		name = req.Name
	}
	clientID := fmt.Sprintf("br_%s", uuid.New().String()[:8])
	if req.ClientID != nil && *req.ClientID != "" {
		clientID = *req.ClientID
	}
	plainSecret, err := generateSecureSecret(32)
	if err != nil {
		return dto.BrandCredentialRes{}, "", errors.ErrUnableTocreate.Wrap(err, "unable to generate secret")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.DefaultCost)
	if err != nil {
		return dto.BrandCredentialRes{}, "", errors.ErrUnableTocreate.Wrap(err, "unable to hash secret")
	}
	query := `INSERT INTO brand_credentials (brand_id, client_id, client_secret_hash, signing_key_encrypted, name, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
	RETURNING id, brand_id, client_id, name, is_active, last_rotated_at, created_at, updated_at`
	var id int32
	var bid int32
	var isActive bool
	var lastRotatedAt sql.NullTime
	var createdAt, updatedAt time.Time
	err = b.db.GetPool().QueryRow(ctx, query, brandID, clientID, string(hash), plainSecret, name).Scan(
		&id, &bid, &clientID, &name, &isActive, &lastRotatedAt, &createdAt, &updatedAt)
	if err != nil {
		b.log.Error("unable to create brand credential", zap.Error(err), zap.Int32("brandID", brandID))
		return dto.BrandCredentialRes{}, "", errors.ErrUnableTocreate.Wrap(err, "unable to create brand credential")
	}
	var lra *time.Time
	if lastRotatedAt.Valid {
		lra = &lastRotatedAt.Time
	}
	return dto.BrandCredentialRes{
		ID: id, BrandID: bid, ClientID: clientID, ClientSecret: plainSecret, Name: name, IsActive: isActive, LastRotatedAt: lra, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, plainSecret, nil
}

func (b *brand) RotateBrandCredential(ctx context.Context, brandID int32, credentialID int32) (newSecret string, err error) {
	plainSecret, err := generateSecureSecret(32)
	if err != nil {
		return "", errors.ErrUnableToUpdate.Wrap(err, "unable to generate secret")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.ErrUnableToUpdate.Wrap(err, "unable to hash secret")
	}
	query := `UPDATE brand_credentials SET client_secret_hash = $3, signing_key_encrypted = $4, last_rotated_at = NOW(), updated_at = NOW()
	WHERE id = $2 AND brand_id = $1 RETURNING id`
	var id int32
	err = b.db.GetPool().QueryRow(ctx, query, brandID, credentialID, string(hash), plainSecret).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.ErrResourceNotFound.New("credential not found")
		}
		b.log.Error("unable to rotate brand credential", zap.Error(err))
		return "", errors.ErrUnableToUpdate.Wrap(err, "unable to rotate brand credential")
	}
	return plainSecret, nil
}

func (b *brand) GetBrandCredentialByID(ctx context.Context, brandID int32, credentialID int32) (dto.BrandCredentialRes, bool, error) {
	query := `SELECT id, brand_id, client_id, name, is_active, last_rotated_at, created_at, updated_at FROM brand_credentials WHERE id = $1 AND brand_id = $2`
	var id int32
	var bid int32
	var clientID string
	var name string
	var isActive bool
	var lastRotatedAt sql.NullTime
	var createdAt, updatedAt time.Time
	err := b.db.GetPool().QueryRow(ctx, query, credentialID, brandID).Scan(
		&id, &bid, &clientID, &name, &isActive, &lastRotatedAt, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.BrandCredentialRes{}, false, nil
		}
		return dto.BrandCredentialRes{}, false, err
	}
	var lra *time.Time
	if lastRotatedAt.Valid {
		lra = &lastRotatedAt.Time
	}
	return dto.BrandCredentialRes{ID: id, BrandID: bid, ClientID: clientID, Name: name, IsActive: isActive, LastRotatedAt: lra, CreatedAt: createdAt, UpdatedAt: updatedAt}, true, nil
}

func (b *brand) GetActiveSigningKeyByBrandID(ctx context.Context, brandID int32) (string, error) {
	query := `SELECT signing_key_encrypted FROM brand_credentials WHERE brand_id = $1 AND is_active = true ORDER BY created_at DESC LIMIT 1`
	var key string
	err := b.db.GetPool().QueryRow(ctx, query, brandID).Scan(&key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return key, nil
}

func (b *brand) AddBrandAllowedOrigin(ctx context.Context, brandID int32, origin string) (dto.BrandAllowedOriginRes, error) {
	query := `INSERT INTO brand_allowed_origins (brand_id, origin) VALUES ($1, $2) RETURNING id, brand_id, origin, created_at`
	var id int32
	var createdAt time.Time
	err := b.db.GetPool().QueryRow(ctx, query, brandID, origin).Scan(&id, &brandID, &origin, &createdAt)
	if err != nil {
		b.log.Error("unable to add allowed origin", zap.Error(err), zap.Int32("brandID", brandID))
		return dto.BrandAllowedOriginRes{}, errors.ErrUnableTocreate.Wrap(err, "unable to add allowed origin")
	}
	return dto.BrandAllowedOriginRes{ID: id, BrandID: brandID, Origin: origin, CreatedAt: createdAt}, nil
}

func (b *brand) RemoveBrandAllowedOrigin(ctx context.Context, brandID int32, originID int32) error {
	_, err := b.db.GetPool().Exec(ctx, `DELETE FROM brand_allowed_origins WHERE id = $1 AND brand_id = $2`, originID, brandID)
	if err != nil {
		return errors.ErrDBDelError.Wrap(err, "unable to remove allowed origin")
	}
	return nil
}

func (b *brand) ListBrandAllowedOrigins(ctx context.Context, brandID int32) ([]dto.BrandAllowedOriginRes, error) {
	rows, err := b.db.GetPool().Query(ctx, `SELECT id, brand_id, origin, created_at FROM brand_allowed_origins WHERE brand_id = $1 ORDER BY created_at`, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []dto.BrandAllowedOriginRes
	for rows.Next() {
		var r dto.BrandAllowedOriginRes
		if err := rows.Scan(&r.ID, &r.BrandID, &r.Origin, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (b *brand) GetBrandFeatureFlags(ctx context.Context, brandID int32) (map[string]bool, error) {
	rows, err := b.db.GetPool().Query(ctx, `SELECT flag_key, enabled FROM brand_feature_flags WHERE brand_id = $1`, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var k string
		var v bool
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (b *brand) UpdateBrandFeatureFlags(ctx context.Context, brandID int32, flags map[string]bool) error {
	tx, err := b.db.GetPool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `DELETE FROM brand_feature_flags WHERE brand_id = $1`, brandID)
	if err != nil {
		return err
	}
	for k, v := range flags {
		if k == "" {
			continue
		}
		_, err = tx.Exec(ctx, `INSERT INTO brand_feature_flags (brand_id, flag_key, enabled, updated_at) VALUES ($1, $2, $3, NOW())`, brandID, k, v)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
