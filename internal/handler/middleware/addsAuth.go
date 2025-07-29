package middleware

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/spf13/viper"
)

func AddsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := viper.GetString("auth.jwt_secret")
		jwtKey := []byte(key)

		// Check if authorization header exists
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			err := fmt.Errorf("authorization header is missing")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Check if it's bearer token
		if len(tokenString) <= 7 || strings.ToUpper(tokenString[:7]) != "BEARER " {
			err := fmt.Errorf("authorization format is Bearer <token>")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Validate token
		tokenString = tokenString[7:]
		claims := &dto.AddsServiceClaim{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			err := fmt.Errorf("invalid or expired token")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set("adds-service-id", claims.ServiceID.String())
		c.Set("adds-service-name", claims.ServiceName)
		c.Next()
	}
}
