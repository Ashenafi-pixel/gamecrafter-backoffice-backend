package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/spf13/viper"
)

func GenerateProviderToken(providerID uuid.UUID) (string, error) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Minute * 10)
	claim := &dto.ProviderClaim{
		ProviderID: providerID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtKey)
}

func VerifyProviderToken(token string) (uuid.UUID, error) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	claim, err := jwt.ParseWithClaims(token, &dto.ProviderClaim{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !claim.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	providerClaim, ok := claim.Claims.(*dto.ProviderClaim)
	if !ok {
		return uuid.UUID{}, errors.New("invalid claim type")
	}

	return providerClaim.ProviderID, nil
}

func GenerateAddsServiceToken(serviceID uuid.UUID, serviceName string) (string, error) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Minute * 10)
	claim := &dto.AddsServiceClaim{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtKey)
}

func VerifyAddsServiceToken(token string) (uuid.UUID, error) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	claim, err := jwt.ParseWithClaims(token, &dto.AddsServiceClaim{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !claim.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	addsServiceClaim, ok := claim.Claims.(*dto.AddsServiceClaim)
	if !ok {
		return uuid.UUID{}, errors.New("invalid claim type")
	}

	return addsServiceClaim.ServiceID, nil
}

func GenerateSportsServiceToken(serviceID string, serviceName string) (string, error) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Minute * 10)
	claim := &dto.SportsServiceClaim{
		ServiceID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"), // Using a fixed UUID for sports service
		ServiceName: serviceName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtKey)
}
