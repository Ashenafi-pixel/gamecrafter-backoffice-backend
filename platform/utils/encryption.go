package utils

import (
	"crypto/rand"
	"encoding/hex"
	rand2 "math/rand"
	"strconv"
	"time"

	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func ComparePasswords(hashedPassword, password string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func GenerateJWT(userID uuid.UUID) (string, error) {
	// Default to unverified for backward compatibility
	return GenerateJWTWithVerification(userID, false, false, false)
}

// GenerateJWTWithVerification generates a JWT token with verification status
func GenerateJWTWithVerification(userID uuid.UUID, isVerified, emailVerified, phoneVerified bool) (string, error) {
	key := viper.GetString("app.jwt_secret")
	if key == "" {
		key = viper.GetString("auth.jwt_secret") // Fallback for backward compatibility
	}
	if key == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Hour * 8) // 8 hours for backoffice admin session timeout

	claim := &dto.Claim{
		UserID:        userID,
		IsVerified:    isVerified,
		EmailVerified: emailVerified,
		PhoneVerified: phoneVerified,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "tucanbit",
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtKey)
}

// GenerateRefreshJWT generates a refresh token as a proper JWT
func GenerateRefreshJWT(userID uuid.UUID) (string, error) {
	key := viper.GetString("app.jwt_secret")
	if key == "" {
		key = viper.GetString("auth.jwt_secret") // Fallback for backward compatibility
	}
	if key == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Hour * 24 * 7) // 7 days for refresh token

	claim := &dto.Claim{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "tucanbit",
			Subject:   userID.String(),
		},
	}

	// Use a different signing method for refresh tokens to distinguish them
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	return token.SignedString(jwtKey)
}

func GenerateUniqueToken(size int) (string, error) {
	token := make([]byte, size)

	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), err
}

func GenerateOTP() string {
	rand2.Seed(time.Now().Unix())
	return strconv.Itoa(rand2.Intn(900000) + 100000)
}

func GenerateOTPJWT(userID uuid.UUID) (string, error) {
	key := viper.GetString("auth.otp_jwt_secret")
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Hour)
	claim := &dto.Claim{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtKey)
}
