package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AgentReferral struct {
	ID               uuid.UUID       `json:"id"`
	RequestID        string          `json:"request_id"`
	CallbackURL      string          `json:"callback_url"`
	UserID           uuid.UUID       `json:"user_id,omitempty"`
	ConversionType   string          `json:"conversion_type,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	MSISDN           string          `json:"msisdn,omitempty"`
	ConvertedAt      time.Time       `json:"converted_at"`
	CallbackSent     bool            `json:"callback_sent"`
	CallbackAttempts int             `json:"callback_attempts"`
}

type CreateAgentReferralLinkReq struct {
	RequestID   string    `json:"request_id" validate:"required"`
	ProviderID  uuid.UUID `json:"provider_id"`
	CallbackURL string    `json:"callback_url`
}

type CreateAgentReferralLinkRes struct {
	Message string `json:"message"`
	Link    string `json:"link"`
}

type UpdateAgentReferralWithConversionReq struct {
	RequestID      string          `json:"request_id" validate:"required"`
	MSISDN         string          `json:"msisdn,validate:"required"`
	UserID         uuid.UUID       `json:"user_id" validate:"required"`
	ConversionType string          `json:"conversion_type"`
	Amount         decimal.Decimal `json:"amount"`
}

type UpdateAgentReferralWithConversionRes struct {
	Message       string        `json:"message"`
	AgentReferral AgentReferral `json:"agent_referral"`
}

type GetAgentReferralsReq struct {
	RequestID string `json:"request_id" binding:"required"`
	Page      int    `json:"page" binding:"min=1"`
	Limit     int    `json:"limit" binding:"min=1,max=100"`
}

type GetAgentReferralsRes struct {
	AgentReferrals []AgentReferral `json:"agent_referrals"`
	Total          int             `json:"total"`
	Page           int             `json:"page"`
	Limit          int             `json:"limit"`
	TotalPages     int             `json:"total_pages"`
}

type GetReferralStatsReq struct {
	RequestID string `form:"request_id" binding:"required"`
}

type GetReferralStatsRes struct {
	Message string        `json:"message"`
	Stats   ReferralStats `json:"stats"`
}

type GetReferralsByUserReq struct {
	UserID  uuid.UUID `json:"user_id" binding:"required"`
	Page    int       `json:"page" binding:"min=1"`
	PerPage int       `json:"limit" binding:"min=1,max=100"`
}

type GetReferralsByUserRes struct {
	Message        string          `json:"message"`
	AgentReferrals []AgentReferral `json:"agent_referrals"`
}

type ReferralStats struct {
	TotalConversions int                            `json:"total_conversions"`
	TotalAmount      decimal.Decimal                `json:"total_amount"`
	UniqueUsers      int                            `json:"unique_users"`
	ConversionTypes  map[string]ConversionTypeStats `json:"conversion_types"`
}

type ConversionTypeStats struct {
	ConversionType   string          `json:"conversion_type"`
	TotalConversions int             `json:"total_conversions"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
}

type GetAgentReferralReq struct {
	RequestID string `form:"request_id" binding:"required"`
}

type PendingCallback struct {
	ID               uuid.UUID       `json:"id"`
	RequestID        string          `json:"request_id"`
	CallbackURL      string          `json:"callback_url"`
	UserID           uuid.UUID       `json:"user_id"`
	ConversionType   string          `json:"conversion_type"`
	Amount           decimal.Decimal `json:"amount"`
	MSISDN           string          `json:"msisdn"`
	ConvertedAt      time.Time       `json:"converted_at"`
	CallbackAttempts int             `json:"callback_attempts"`
}

type CreateAgentProviderReq struct {
	Name         string `json:"name" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	Description  string `json:"description"`
	CallbackURL  string `json:"callback_url"`
}

type AgentProviderRes struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ClientID    string    `json:"client_id"`
	Description string    `json:"description"`
	CallbackURL string    `json:"callback_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateAgentProviderRes struct {
	Message  string           `json:"message"`
	Provider AgentProviderRes `json:"provider"`
}
