package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetBrandSigningKeyFunc returns the signing key for the given brand ID (e.g. from brand_credentials).
// Used to validate X-Brand-Signature. If key is empty or error, validation fails.
type GetBrandSigningKeyFunc func(ctx context.Context, brandID int32) (string, error)

const (
	HeaderBrandID         = "X-Brand-Id"
	HeaderBrandSignature  = "X-Brand-Signature"
)

// BrandSignatureMiddleware validates HMAC-SHA256 signature of the request using the brand's signing key.
// Brand ID is read from X-Brand-Id header. Signature from X-Brand-Signature.
// Payload: for GET/DELETE use raw query string; for POST/PUT/PATCH use request body.
func BrandSignatureMiddleware(getKey GetBrandSigningKeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandIDStr := c.GetHeader(HeaderBrandID)
		if brandIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Missing X-Brand-Id header",
			})
			c.Abort()
			return
		}
		brandID, err := strconv.ParseInt(strings.TrimSpace(brandIDStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Invalid X-Brand-Id",
			})
			c.Abort()
			return
		}
		key, err := getKey(c.Request.Context(), int32(brandID))
		if err != nil || key == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "brand credentials not found or invalid",
			})
			c.Abort()
			return
		}
		signature := c.GetHeader(HeaderBrandSignature)
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Missing X-Brand-Signature header",
			})
			c.Abort()
			return
		}
		var payload []byte
		switch c.Request.Method {
		case http.MethodGet, http.MethodDelete:
			payload = []byte(c.Request.URL.RawQuery)
		default:
			payload, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    1001,
					"status":  "Invalid signature",
					"message": "Unable to read body",
				})
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
		}
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write(payload)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			c.Abort()
			return
		}
		c.Set("brand_id", int32(brandID))
		c.Next()
	}
}
