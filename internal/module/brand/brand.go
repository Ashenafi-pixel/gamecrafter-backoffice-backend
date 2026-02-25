package brand

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

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

func (b *brand) GetBrandByID(ctx context.Context, id int32) (dto.Brand, error) {
	if id <= 0 {
		err := errors.ErrInvalidUserInput.New("invalid brand ID")
		return dto.Brand{}, err
	}

	brand, exists, err := b.brandStorage.GetBrandByID(ctx, id)
	if err != nil {
		return dto.Brand{}, err
	}

	if !exists {
		err := fmt.Errorf("brand not found with ID: %d", id)
		b.log.Warn("brand not found", zap.Int32("id", id))
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
		err := fmt.Errorf("brand not found with ID: %d", req.ID)
		b.log.Error(err.Error(), zap.Int32("brandID", req.ID))
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

func (b *brand) DeleteBrand(ctx context.Context, brandID int32) error {
	if brandID <= 0 {
		err := errors.ErrInvalidUserInput.New("invalid brand ID")
		return err
	}

	// Check if brand exists before deleting
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return err
	}
	if !exists {
		err := fmt.Errorf("brand not found with ID: %d", brandID)
		b.log.Warn("brand not found for deletion", zap.Int32("id", brandID))
		err = errors.ErrResourceNotFound.Wrap(err, err.Error())
		return err
	}

	if err := b.brandStorage.DeleteBrand(ctx, brandID); err != nil {
		return err
	}

	return nil
}

func (b *brand) ChangeBrandStatus(ctx context.Context, brandID int32, req dto.ChangeBrandStatusReq) error {
	if brandID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("brand not found")
	}
	return b.brandStorage.UpdateBrandStatus(ctx, brandID, req.IsActive)
}

func (b *brand) CreateBrandCredential(ctx context.Context, brandID int32, req dto.CreateBrandCredentialReq) (dto.BrandCredentialRes, string, error) {
	if brandID <= 0 {
		return dto.BrandCredentialRes{}, "", errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return dto.BrandCredentialRes{}, "", err
	}
	if !exists {
		return dto.BrandCredentialRes{}, "", errors.ErrResourceNotFound.New("brand not found")
	}
	return b.brandStorage.CreateBrandCredential(ctx, brandID, req)
}

func (b *brand) RotateBrandCredential(ctx context.Context, brandID int32, credentialID int32) (dto.RotateBrandCredentialRes, error) {
	if brandID <= 0 || credentialID <= 0 {
		return dto.RotateBrandCredentialRes{}, errors.ErrInvalidUserInput.New("invalid brand or credential ID")
	}
	newSecret, err := b.brandStorage.RotateBrandCredential(ctx, brandID, credentialID)
	if err != nil {
		return dto.RotateBrandCredentialRes{}, err
	}
	return dto.RotateBrandCredentialRes{ClientSecret: newSecret, LastRotatedAt: time.Now()}, nil
}

func (b *brand) GetBrandCredentialByID(ctx context.Context, brandID int32, credentialID int32) (dto.BrandCredentialRes, bool, error) {
	return b.brandStorage.GetBrandCredentialByID(ctx, brandID, credentialID)
}

func (b *brand) AddBrandAllowedOrigin(ctx context.Context, brandID int32, req dto.AddBrandAllowedOriginReq) (dto.BrandAllowedOriginRes, error) {
	if brandID <= 0 {
		return dto.BrandAllowedOriginRes{}, errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	if req.Origin == "" {
		return dto.BrandAllowedOriginRes{}, errors.ErrInvalidUserInput.New("origin is required")
	}
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return dto.BrandAllowedOriginRes{}, err
	}
	if !exists {
		return dto.BrandAllowedOriginRes{}, errors.ErrResourceNotFound.New("brand not found")
	}
	return b.brandStorage.AddBrandAllowedOrigin(ctx, brandID, req.Origin)
}

func (b *brand) RemoveBrandAllowedOrigin(ctx context.Context, brandID int32, originID int32) error {
	if brandID <= 0 || originID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid brand or origin ID")
	}
	return b.brandStorage.RemoveBrandAllowedOrigin(ctx, brandID, originID)
}

func (b *brand) ListBrandAllowedOrigins(ctx context.Context, brandID int32) (dto.ListBrandAllowedOriginsRes, error) {
	if brandID <= 0 {
		return dto.ListBrandAllowedOriginsRes{}, errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	list, err := b.brandStorage.ListBrandAllowedOrigins(ctx, brandID)
	if err != nil {
		return dto.ListBrandAllowedOriginsRes{}, err
	}
	return dto.ListBrandAllowedOriginsRes{Origins: list}, nil
}

func (b *brand) GetBrandFeatureFlags(ctx context.Context, brandID int32) (dto.BrandFeatureFlagsRes, error) {
	if brandID <= 0 {
		return dto.BrandFeatureFlagsRes{}, errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	flags, err := b.brandStorage.GetBrandFeatureFlags(ctx, brandID)
	if err != nil {
		return dto.BrandFeatureFlagsRes{}, err
	}
	if flags == nil {
		flags = make(map[string]bool)
	}
	return dto.BrandFeatureFlagsRes{Flags: flags}, nil
}

func (b *brand) UpdateBrandFeatureFlags(ctx context.Context, brandID int32, req dto.UpdateBrandFeatureFlagsReq) error {
	if brandID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid brand ID")
	}
	_, exists, err := b.brandStorage.GetBrandByID(ctx, brandID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("brand not found")
	}
	return b.brandStorage.UpdateBrandFeatureFlags(ctx, brandID, req.Flags)
}
