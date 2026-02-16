package dto

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Claim struct {
	UserID        uuid.UUID `json:"user_id"`
	IsVerified    bool      `json:"is_verified"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	jwt.StandardClaims
}

type ProviderClaim struct {
	ProviderID uuid.UUID `json:"provider_id"`
	jwt.StandardClaims
}
