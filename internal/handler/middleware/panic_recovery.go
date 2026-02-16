package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func PanicRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  500,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
