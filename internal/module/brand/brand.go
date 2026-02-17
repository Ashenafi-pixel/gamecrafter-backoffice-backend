package brand

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type brand struct {
	log          *zap.Logger
	brandStorage storage.Brand
}

func Init(brandStorage storage.Brand, log *zap.Logger) module.Brand {
	return &brand{
		log:          log,
		brandStorage: brandStorage,
	}
}
func ExtractDomain(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("invalid URL: no hostname found")
	}

	return host, nil
}

func (b *brand) CreateBrand(ctx context.Context, req dto.CreateBrandReq) (dto.CreateBrandRes, error) {
	if err := dto.ValidateCreateBrand(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CreateBrandRes{}, err
	}

	//extract domain from webhook URL if domain is not provided but webhook URL is provided
	if req.Domain == "" && (req.WebhookURL != "" || req.APIURL != "") {
		var sourceURL string
		if req.WebhookURL != "" {
			sourceURL = req.WebhookURL
		} else {
			sourceURL = req.APIURL
		}

		domain, err := ExtractDomain(sourceURL)
		if err == nil && domain != "" {
			req.Domain = domain
		}
	}
	fmt.Println("Brand domain data is ", req.Domain)
	return b.brandStorage.CreateBrand(ctx, req)
}

func (b *brand) GetBrandByID(ctx context.Context, id uuid.UUID) (dto.Brand, error) {
	if id == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid brand ID")
		return dto.Brand{}, err
	}

	brand, exists, err := b.brandStorage.GetBrandByID(ctx, id)
	if err != nil {
		return dto.Brand{}, err
	}

	if !exists {
		err := fmt.Errorf("brand not found with ID: %s", id.String())
		b.log.Warn("brand not found", zap.String("id", id.String()))
		err = errors.ErrResourceNotFound.Wrap(err, err.Error())
		return dto.Brand{}, err
	}

	return brand, nil
}

func (b *brand) GetBrands(ctx context.Context, req dto.GetBrandsReq) (dto.GetBrandsRes, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	brandsRes, err := b.brandStorage.GetBrands(ctx, req)
	if err != nil {
		return dto.GetBrandsRes{}, err
	}
	return brandsRes, nil
}

func (b *brand) UpdateBrand(ctx context.Context, req dto.UpdateBrandReq) (dto.UpdateBrandRes, error) {
	brand, exists, err := b.brandStorage.GetBrandByID(ctx, req.ID)
	if err != nil {
		return dto.UpdateBrandRes{}, err
	}
	if !exists {
		err := fmt.Errorf("brand not found with ID: %s", req.ID.String())
		b.log.Error(err.Error(), zap.String("brandID", req.ID.String()))
		err = errors.ErrResourceNotFound.Wrap(err, err.Error())
		return dto.UpdateBrandRes{}, err
	}

	if err := dto.ValidateUpdateBrand(req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.UpdateBrandRes{}, err
	}

	// Fill in missing fields from existing brand
	if req.Name == nil {
		req.Name = &brand.Name
	}
	if req.Code == nil {
		req.Code = &brand.Code
	}
	if req.Domain == nil {
		req.Domain = brand.Domain
	}
	if req.IsActive == nil {
		req.IsActive = &brand.IsActive
	}

	return b.brandStorage.UpdateBrand(ctx, req)
}

func (b *brand) DeleteBrand(ctx context.Context, brandID uuid.UUID) error {
	if brandID == uuid.Nil {
		err := errors.ErrInvalidUserInput.New("invalid brand ID")
		return err
	}

	// Check if brand exists before deleting
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return err
	}
	if !exists {
		err := fmt.Errorf("brand not found with ID: %s", brandID.String())
		b.log.Warn("brand not found for deletion", zap.String("id", brandID.String()))
		err = errors.ErrResourceNotFound.Wrap(err, err.Error())
		return err
	}

	if err := b.brandStorage.DeleteBrand(ctx, brandID); err != nil {
		return err
	}

	return nil
}
