package dto

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AddsServiceClaim represents the JWT claims for adds service authentication
type AddsServiceClaim struct {
	ServiceID   uuid.UUID `json:"service_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceName string    `json:"service_name" example:"Adds Service"`
	jwt.StandardClaims
}

// AddsServiceConfig represents the configuration for adds service
type AddsServiceConfig struct {
	ID            uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceID     string    `json:"service_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceName   string    `json:"service_name" example:"Adds Service"`
	ServiceSecret string    `json:"service_secret,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status        string    `json:"status" example:"active"`
	CreatedAt     time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

// CreateAddsServiceReq represents the request to create a new adds service
type CreateAddsServiceReq struct {
	Name          string    `json:"name" validate:"required,min=3,max=255" example:"Adds Service"`
	Description   string    `json:"description" example:"Adds Service Description"`
	ServiceURL    string    `json:"service_url" validate:"required,url" example:"https://adds.service.com"`
	ServiceID     string    `json:"service_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceSecret string    `json:"service_secret" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedBy     uuid.UUID `json:"created_by" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status        string    `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

// AddsServiceResData represents the data structure for adds service
type AddsServiceResData struct {
	ID            uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name          string     `json:"name" example:"Adds Service"`
	Description   string     `json:"description" example:"Adds Service Description"`
	ServiceID     string     `json:"service_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceSecret string     `json:"service_secret,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status        string     `json:"status" example:"active"`
	ServiceURL    string     `json:"service_url" example:"https://adds.service.com"`
	CreatedBy     uuid.UUID  `json:"created_by" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt     time.Time  `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" example:"2021-01-01T00:00:00Z"`
}

// CreateAddsServiceRes represents the response for creating adds service
type CreateAddsServiceRes struct {
	Message string             `json:"message" example:"Adds service created successfully"`
	Data    AddsServiceResData `json:"data" `
}

type GetAddServicesRequest struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

// GetAddsServicesRes represents the response for getting adds services
type GetAddsServicesRes struct {
	Message string               `json:"message" example:"Adds services fetched successfully"`
	Data    []AddsServiceResData `json:"data"`
}

type SignInReq struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// ValidateSignIn validates the adds service sign-in request
func ValidateSignIn(req SignInReq) error {
	validate := validator.New()
	return validate.Struct(req)
}

// SignInReq represents the sign-in request for adds service
type AddSignInReq struct {
	ServiceID     string `json:"service_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceSecret string `json:"service_secret" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// SignInRes represents the sign-in response for adds service
type AddSignInRes struct {
	Token string `json:"token" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateBalanceReq represents the balance update request for adds service
type AddUpdateBalanceReq struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Currency  string    `json:"currency" validate:"required"`
	Component string    `json:"component" validate:"required"`
	Amount    string    `json:"amount" validate:"required"`
	Reason    string    `json:"reason"`
	Reference string    `json:"reference"`
}

// UpdateBalanceRes represents the balance update response for adds service
type AddUpdateBalanceRes struct {
	Message string `json:"message" example:"Balance updated successfully"`
	Status  string `json:"status" example:"success"`
}

// ValidateUpdateBalance validates the adds balance update request
func ValidateUpdateBalance(req UpdateBalanceReq) error {
	validate := validator.New()
	return validate.Struct(req)
}
