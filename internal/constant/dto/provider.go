// internal/constant/dto/provider.go

package dto

import (
	"time"

	"github.com/google/uuid"
)

// GameProvider represents a game content provider
type GameProvider struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	Description     *string   `json:"description,omitempty"`
	APIURL          *string   `json:"api_url,omitempty"`
	WebhookURL      *string   `json:"webhook_url,omitempty"`
	IntegrationType string    `json:"integration_type"` // API, WEBHOOK, BOTH
	IsActive        bool      `json:"is_active"`
	Status          string    `json:"status"` // ACTIVE, INACTIVE, MAINTENANCE (for backward compatibility)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateProviderRequest for creating a provider
type CreateProviderRequest struct {
	Name            string  `json:"name" validate:"required,min=1,max=255"`
	Code            string  `json:"code" validate:"required,min=1,max=100"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	APIURL          *string `json:"api_url,omitempty" validate:"omitempty,url,max=500"`
	WebhookURL      *string `json:"webhook_url,omitempty" validate:"omitempty,url,max=500"`
	IntegrationType string  `json:"integration_type" validate:"omitempty,oneof=API WEBHOOK BOTH"`
	IsActive        bool    `json:"is_active"`
}

// UpdateProviderRequest for updating a provider
type UpdateProviderRequest struct {
	Name            *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code            *string   `json:"code,omitempty" validate:"omitempty,min=1,max=100"`
	Description     *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	APIURL          *string   `json:"api_url,omitempty" validate:"omitempty,url,max=500"`
	WebhookURL      *string   `json:"webhook_url,omitempty" validate:"omitempty,url,max=500"`
	IntegrationType *string   `json:"integration_type,omitempty" validate:"omitempty,oneof=API WEBHOOK BOTH"`
	IsActive        *bool     `json:"is_active,omitempty"`
	ID              uuid.UUID `json:"id" validate:"required"`
}

// ChangeProviderStatusRequest for changing provider status only
type ChangeProviderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE MAINTENANCE"`
}

// ProviderListRequest for listing providers with filters
type ProviderListRequest struct {
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"per_page" form:"per_page"`
	Search  string `json:"search" form:"search"`
	Status  string `json:"status" form:"status"`
}

// ProviderListResponse paginated list
type ProviderListResponse struct {
	Providers  []GameProvider `json:"providers"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// AssignProviderToBrandRequest body for assigning provider to brand
type AssignProviderToBrandRequest struct {
	BrandID    uuid.UUID `json:"brand_id" validate:"required"`
	ProviderID uuid.UUID `json:"provider_id" validate:"required"`
}

// AssignGameToBrandRequest body for assigning game to brand
type AssignGameToBrandRequest struct {
	BrandID uuid.UUID `json:"brand_id" validate:"required"`
	GameID  uuid.UUID `json:"game_id" validate:"required"`
}

// BrandProvidersResponse list of providers assigned to a brand
type BrandProvidersResponse struct {
	BrandID   uuid.UUID      `json:"brand_id"`
	Providers []GameProvider `json:"providers"`
}

// BrandGamesResponse list of games assigned to a brand (minimal game info)
type BrandGamesResponse struct {
	BrandID uuid.UUID      `json:"brand_id"`
	Games   []GameResponse `json:"games"`
}
