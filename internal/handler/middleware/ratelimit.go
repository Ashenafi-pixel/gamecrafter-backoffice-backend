package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/axiaoxin-com/ratelimiter"
)

func RateLimiter() gin.HandlerFunc {
	limiter := ratelimiter.GinMemRatelimiter(ratelimiter.GinRatelimiterConfig{
		LimitKey: func(c *gin.Context) string {
			return c.ClientIP()
		},
		LimitedHandler: func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, "too many requests!!!")
			c.Abort()
			return
		},
		TokenBucketConfig: func(*gin.Context) (time.Duration, int) {
			return time.Second * 1, 5
		},
	})
	return limiter
}
