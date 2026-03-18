package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// OperatorRateLimiter limits requests per operator (keyed by X-Operator-Id or operator_id from context).
// Uses in-memory map; for production use Redis for multi-instance consistency.
type OperatorRateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewOperatorRateLimiter creates a per-operator rate limiter. limit is max requests per window (e.g. 100 per second).
func NewOperatorRateLimiter(limit int, window time.Duration) *OperatorRateLimiter {
	return &OperatorRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *OperatorRateLimiter) allow(key string) bool {
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

const defaultOperatorRateLimit = 100
const defaultOperatorRateWindow = time.Second

// OperatorRateLimitMiddleware limits request rate per operator. Operator ID from X-Operator-Id header
// or gin "operator_id" (set e.g. by OperatorSignatureMiddleware).
// If no operator is identified, uses client IP as fallback key.
func OperatorRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = defaultOperatorRateLimit
	}
	if window <= 0 {
		window = defaultOperatorRateWindow
	}
	limiter := NewOperatorRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if id, exists := c.Get("operator_id"); exists {
			if operatorID, ok := id.(int32); ok {
				key = "operator:" + strconv.FormatInt(int64(operatorID), 10)
			}
		}
		if operatorIDStr := c.GetHeader(HeaderOperatorID); operatorIDStr != "" {
			key = "operator:" + operatorIDStr
		}
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Operator rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

