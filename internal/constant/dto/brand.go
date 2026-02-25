package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Brand struct {
	ID         int32     `json:"id"`
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	Code       string    `json:"code" validate:"required,min=1,max=50"`
	Domain     *string   `json:"domain,omitempty" validate:"omitempty,max=255"`
	IsActive   bool      `json:"is_active"`
	WebhookURL *string   `json:"webhook_url,omitempty" validate:"omitempty,max=255"`
	APIURL     *string   `json:"api_url,omitempty" validate:"omitempty,max=255"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateBrandReq struct {
	Name            string `json:"name" validate:"required,min=1,max=255"`
	Code            string `json:"code" validate:"required,min=1,max=50"`
	Domain          string `json:"domain,omitempty" validate:"omitempty,max=255"`
	IsActive        bool   `json:"is_active,omitempty"`
	Description     string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Signature       string `json:"signature,omitempty" validate:"omitempty,max=255"`
	WebhookURL      string `json:"webhook_url,omitempty" validate:"omitempty,max=255"`
	IntegrationType string `json:"integration_type,omitempty" validate:"omitempty,max=255"`
	APIURL          string `json:"api_url,omitempty" validate:"omitempty,max=255"`
}

type CreateBrandRes struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Domain    *string   `json:"domain,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateBrandReq struct {
	ID         int32     `json:"id"`
	Name       *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code       *string   `json:"code,omitempty" validate:"omitempty,min=1,max=50"`
	Domain     *string   `json:"domain,omitempty" validate:"omitempty,max=255"`
	WebhookURL *string   `json:"webhook_url,omitempty" validate:"omitempty,max=255"`
	APIURL     *string   `json:"api_url,omitempty" validate:"omitempty,max=255"`
	IsActive   *bool     `json:"is_active,omitempty"`
}

type UpdateBrandRes struct {
	ID         int32     `json:"id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	Domain     *string   `json:"domain,omitempty"`
	WebhookURL *string   `json:"webhook_url,omitempty"`
	APIURL     *string   `json:"api_url,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type GetBrandsReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PerPage  int    `form:"per-page" binding:"required,min=1,max=100"`
	Search   string `form:"search,omitempty"`
	IsActive *bool  `form:"is_active,omitempty"`
}

type GetBrandsRes struct {
	Brands      []Brand `json:"brands"`
	TotalCount  int     `json:"total_count"`
	TotalPages  int     `json:"total_pages"`
	CurrentPage int     `json:"current_page"`
	PerPage     int     `json:"per_page"`
}

func ValidateCreateBrand(req CreateBrandReq) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidateUpdateBrand(req UpdateBrandReq) error {
	validate := validator.New()
	return validate.Struct(req)
}

// ChangeBrandStatusReq for dedicated status change endpoint
type ChangeBrandStatusReq struct {
	IsActive bool `json:"is_active"`
}

// --- Brand Credentials ---
type CreateBrandCredentialReq struct {
	Name     string  `json:"name" validate:"omitempty,max=100"`
	ClientID *string `json:"client_id,omitempty" validate:"omitempty,max=255"`
}

type BrandCredentialRes struct {
	ID           int32     `json:"id"`
	BrandID      int32     `json:"brand_id"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret,omitempty"` // only returned once on create/rotate
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	LastRotatedAt *time.Time `json:"last_rotated_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RotateBrandCredentialRes struct {
	ClientSecret string     `json:"client_secret"` // only returned once
	LastRotatedAt time.Time `json:"last_rotated_at"`
}

// --- Brand Allowed Origins ---
type AddBrandAllowedOriginReq struct {
	Origin string `json:"origin" validate:"required,max=500"`
}

type BrandAllowedOriginRes struct {
	ID        int32     `json:"id"`
	BrandID   int32     `json:"brand_id"`
	Origin    string    `json:"origin"`
	CreatedAt time.Time `json:"created_at"`
}

type ListBrandAllowedOriginsRes struct {
	Origins []BrandAllowedOriginRes `json:"origins"`
}

// --- Brand Feature Flags (stored in system_config or key-value) ---
type BrandFeatureFlagsRes struct {
	Flags map[string]bool `json:"flags"`
}

type UpdateBrandFeatureFlagsReq struct {
	Flags map[string]bool `json:"flags" validate:"required"`
}
