package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Banner struct {
	ID        uuid.UUID  `json:"id"`
	Page      string     `json:"page"`
	PageURL   string     `json:"page_url"`
	ImageURL  string     `json:"image_url"`
	Headline  string     `json:"headline"`
	Tagline   string     `json:"tagline"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type GetBannerReq struct {
	Page string `form:"page" validate:"required"`
}

type BannerDisplay struct {
	ImageURL string `json:"image_url"`
	Headline string `json:"headline"`
	Tagline  string `json:"tagline"`
}

type CreateBannerReq struct {
	Page     string `json:"page" validate:"required"`
	PageURL  string `json:"page_url" validate:"required,url"`
	ImageURL string `json:"image_url" validate:"required,url"`
	Headline string `json:"headline" validate:"required,max=255"`
	Tagline  string `json:"tagline" validate:"max=500"`
}

type UpdateBannerReq struct {
	ID       uuid.UUID `json:"id"`
	PageURL  string    `json:"page_url" validate:"omitempty,url"`
	Headline string    `json:"headline" validate:"omitempty,max=255"`
	Tagline  string    `json:"tagline" validate:"omitempty,max=500"`
	ImageURL string    `json:"image_url" validate:"omitempty,url"`
}

type GetBannersReq struct {
	Page    int `form:"page" validate:"min=1"`
	PerPage int `form:"per_page" validate:"min=1,max=100"`
}

type GetBannersRes struct {
	Banners     []Banner `json:"banners"`
	TotalPages  int      `json:"total_pages"`
	TotalCount  int      `json:"total_count"`
	CurrentPage int      `json:"current_page"`
}

type UploadBannerImageResp struct {
	Status string `json:"status"`
	Url    string `json:"url"`
}

// Validation functions
func ValidateCreateBanner(req CreateBannerReq) error {
	validate := validator.New()
	return validate.Struct(req)
}

func ValidateUpdateBanner(req UpdateBannerReq) error {
	validate := validator.New()
	return validate.Struct(req)
}
