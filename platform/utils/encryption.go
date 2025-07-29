package utils

import (
	"crypto/rand"
	"encoding/hex"
	rand2 "math/rand"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/spf13/viper"
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
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Minute * 10)
	claim := &dto.Claim{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
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
