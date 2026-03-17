package dto

import "time"

type Operator struct {
	OperatorID           int32     `json:"operator_id"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	Domain               string    `json:"domain,omitempty"`
	LogoURL              string    `json:"logo_url,omitempty"`
	IsActive             bool      `json:"is_active"`
	AllowedEmbedDomains  []string  `json:"allowed_embed_domains,omitempty"`
	EmbedRefererRequired bool      `json:"embed_referer_required"`
	TransactionURL       string    `json:"transaction_url,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CreateOperatorReq struct {
	OperatorID           int32    `json:"operator_id" validate:"required,gte=100000,lte=999999"`
	Name                 string   `json:"name" validate:"required,min=1,max=255"`
	Code                 string   `json:"code" validate:"required,min=1,max=50"`
	Domain               string   `json:"domain,omitempty" validate:"omitempty,max=255"`
	LogoURL              string   `json:"logo_url,omitempty" validate:"omitempty,max=500"`
	IsActive             bool     `json:"is_active"`
	AllowedEmbedDomains  []string `json:"allowed_embed_domains,omitempty"`
	EmbedRefererRequired bool     `json:"embed_referer_required"`
	TransactionURL       string   `json:"transaction_url,omitempty" validate:"omitempty,max=500"`
}

type UpdateOperatorReq struct {
	OperatorID           int32    `json:"operator_id"`
	Name                 *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code                 *string  `json:"code,omitempty" validate:"omitempty,min=1,max=50"`
	Domain               *string  `json:"domain,omitempty" validate:"omitempty,max=255"`
	LogoURL              *string  `json:"logo_url,omitempty" validate:"omitempty,max=500"`
	IsActive             *bool    `json:"is_active,omitempty"`
	AllowedEmbedDomains  []string `json:"allowed_embed_domains,omitempty"`
	EmbedRefererRequired *bool    `json:"embed_referer_required,omitempty"`
	TransactionURL       *string  `json:"transaction_url,omitempty" validate:"omitempty,max=500"`
}

type ChangeOperatorStatusReq struct {
	IsActive bool `json:"is_active"`
}

type GetOperatorsReq struct {
	Page     int    `form:"page" json:"page"`
	PerPage  int    `form:"per_page" json:"per_page"`
	Search   string `form:"search" json:"search"`
	IsActive *bool  `form:"is_active" json:"is_active"`
}

type GetOperatorsRes struct {
	Operators []Operator `json:"operators"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
}

// Operator credential DTOs

type OperatorCredentialRes struct {
	ID          int32     `json:"id"`
	OperatorID  int32     `json:"operator_id"`
	APIKey      string    `json:"api_key"`
	SigningKey  string    `json:"signing_key"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RotateOperatorCredentialRes struct {
	APIKey     string    `json:"api_key"`
	SigningKey string    `json:"signing_key"`
	UpdatedAt  time.Time `json:"updated_at"`
}

