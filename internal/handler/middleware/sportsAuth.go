package middleware

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/spf13/viper"
)

func SportsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := viper.GetString("auth.jwt_secret")
		jwtKey := []byte(key)

		// Check if authorization header exists
		tokenString := c.GetHeader("Service-Token")
		if tokenString == "" {
			err := fmt.Errorf("service authorization header is missing")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Check if it's bearer token
		if len(tokenString) <= 7 || strings.ToUpper(tokenString[:7]) != "BEARER " {
			err := fmt.Errorf("service authorization format is Bearer <token>")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Validate token
		tokenString = tokenString[7:]
		claims := &dto.SportsServiceClaim{}
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

		c.Set("sports-service-id", claims.ServiceID.String())
		c.Set("sports-service-name", claims.ServiceName)
		c.Next()
	}
}
