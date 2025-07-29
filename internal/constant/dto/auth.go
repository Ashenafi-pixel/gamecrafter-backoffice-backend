package dto

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Claim struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.StandardClaims
}

type ProviderClaim struct {
	ProviderID uuid.UUID `json:"provider_id"`
	jwt.StandardClaims
}
