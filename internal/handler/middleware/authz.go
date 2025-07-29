package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/module"
)

func Authz(authzModule module.Authz, enforcer *casbin.Enforcer, name, method string) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetString("user-id")
		userIDParsed, err := uuid.Parse(userID)
		if err != nil {
			err := fmt.Errorf("unable to convert user id to uuid")
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			c.Abort()
		}

		roles, err := authzModule.GetUserRoles(context.Background(), userIDParsed)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
		}

		hasPermission := false
		for _, role := range roles.Roles {
			if role.Name == "super" {
				hasPermission = true
				break
			}
			ok, err := enforcer.Enforce(role.ID.String(), name, method)
			if err != nil {
				_ = c.Error(err)
				c.Abort()
			}
			if ok {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
		}

		c.Next()
	}
}
