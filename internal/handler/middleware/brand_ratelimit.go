package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// BrandRateLimiter limits requests per brand (keyed by X-Brand-Id or brand_id from context).
// Uses in-memory map; for production use Redis for multi-instance consistency.
type BrandRateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewBrandRateLimiter creates a per-brand rate limiter. limit is max requests per window (e.g. 100 per second).
func NewBrandRateLimiter(limit int, window time.Duration) *BrandRateLimiter {
	return &BrandRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *BrandRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	if times, ok := rl.requests[key]; ok {
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		rl.requests[key] = valid
	}
	if len(rl.requests[key]) >= rl.limit {
		return false
	}
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

const defaultBrandRateLimit = 100
const defaultBrandRateWindow = time.Second

// BrandRateLimitMiddleware limits request rate per brand. Brand ID from X-Brand-Id header or gin "brand_id" (set e.g. by BrandSignatureMiddleware).
// If no brand is identified, uses client IP as fallback key.
func BrandRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = defaultBrandRateLimit
	}
	if window <= 0 {
		window = defaultBrandRateWindow
	}
	limiter := NewBrandRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if id, exists := c.Get("brand_id"); exists {
			if bid, ok := id.(int32); ok {
				key = "brand:" + strconv.FormatInt(int64(bid), 10)
			}
		}
		if brandIDStr := c.GetHeader(HeaderBrandID); brandIDStr != "" {
			key = "brand:" + brandIDStr
		}
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Brand rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
