package middleware

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := viper.GetString("app.jwt_secret")
		if key == "" {
			key = viper.GetString("auth.jwt_secret") // Fallback for backward compatibility
		}
		if key == "" {
			err := fmt.Errorf("JWT secret not configured")
			err = errors.ErrInvalidAccessToken.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}
		jwtKey := []byte(key)
		//check if authorization header is exist or not
		tokenString := c.GetHeader("Authorization")
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
		tokenString = tokenString[7:]
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
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
		c.Set("is-verified", claims.IsVerified)
		c.Set("email-verified", claims.EmailVerified)
		c.Set("phone-verified", claims.PhoneVerified)
	}
}

// RequireVerification middleware ensures the user is verified before accessing protected routes
func RequireVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user is verified
		isVerified, exists := c.Get("is-verified")
		if !exists {
			c.JSON(401, gin.H{"error": "Verification status not found"})
			c.Abort()
			return
		}

		if !isVerified.(bool) {
			c.JSON(403, gin.H{
				"error":   "Account verification required",
				"message": "Please verify your account before accessing this resource",
				"code":    "VERIFICATION_REQUIRED",
			})
			c.Abort()
			return
		}
	}
}

// RequireEmailVerification middleware ensures the user's email is verified
func RequireEmailVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user's email is verified
		emailVerified, exists := c.Get("email-verified")
		if !exists {
			c.JSON(401, gin.H{"error": "Email verification status not found"})
			c.Abort()
			return
		}

		if !emailVerified.(bool) {
			c.JSON(403, gin.H{
				"error":   "Email verification required",
				"message": "Please verify your email address before accessing this resource",
				"code":    "EMAIL_VERIFICATION_REQUIRED",
			})
			c.Abort()
			return
		}
	}
}

// RequirePhoneVerification middleware ensures the user's phone is verified
func RequirePhoneVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure user is authenticated
		Auth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user's phone is verified
		phoneVerified, exists := c.Get("phone-verified")
		if !exists {
			c.JSON(401, gin.H{"error": "Phone verification status not found"})
			c.Abort()
			return
		}

		if !phoneVerified.(bool) {
			c.JSON(403, gin.H{
				"error":   "Phone verification required",
				"message": "Please verify your phone number before accessing this resource",
				"code":    "PHONE_VERIFICATION_REQUIRED",
			})
			c.Abort()
			return
		}
	}
}
