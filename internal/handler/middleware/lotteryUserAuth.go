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

func LotteryUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := viper.GetString("auth.jwt_secret")
		jwtKey := []byte(key)
		//check if authorization header is exist or not
		tokenString := c.GetHeader("x-user-token")
		if tokenString == "" {
			err := fmt.Errorf("authorization header is missing")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		// check if it bearer token or not
		if len(tokenString) <= 7 || strings.ToUpper(tokenString[:7]) != "BEARER " {
			err := fmt.Errorf("authorization format is Bearer <token> ")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// validate token
		xuserToken := strings.ReplaceAll(tokenString, "Bearer ", "")
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(xuserToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			err := fmt.Errorf("invalid or expired token ")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set("user-id", claims.UserID.String())
	}
}
