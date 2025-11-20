package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
)

func Authz(authzModule module.Authz, name, method string) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetString("user-id")
		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			err := fmt.Errorf("unable to convert user id to uuid")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Check if user has "super" role first
		roles, err := authzModule.GetUserRoles(context.Background(), userIDParsed)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Check for super admin role
		for _, role := range roles.Roles {
			if role.Name == "super" {
				c.Next()
				return
			}
		}

		// Check permission directly from database (user_roles -> role_permissions -> permissions)
		hasPermission, err := authzModule.CheckUserHasPermission(context.Background(), userIDParsed, name)
			if err != nil {
				_ = c.Error(err)
				c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": fmt.Sprintf("User does not have permission: %s", name)})
			c.Abort()
			return
		}

		c.Next()
	}
}
