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

// GetOperatorSigningKeyFunc returns the signing key for the given operator ID (e.g. from operator_credentials).
// Used to validate X-Operator-Signature. If key is empty or error, validation fails.
type GetOperatorSigningKeyFunc func(ctx context.Context, operatorID int32) (string, error)

const (
	HeaderOperatorID        = "X-Operator-Id"
	HeaderOperatorSignature = "X-Operator-Signature"
)

// OperatorSignatureMiddleware validates HMAC-SHA256 signature of the request using the operator's signing key.
// Operator ID is read from X-Operator-Id header. Signature from X-Operator-Signature.
// Payload: for GET/DELETE use raw query string; for POST/PUT/PATCH use request body.
func OperatorSignatureMiddleware(getKey GetOperatorSigningKeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorIDStr := c.GetHeader(HeaderOperatorID)
		if operatorIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Missing X-Operator-Id header",
			})
			c.Abort()
			return
		}
		operatorID, err := strconv.ParseInt(strings.TrimSpace(operatorIDStr), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Invalid X-Operator-Id",
			})
			c.Abort()
			return
		}
		key, err := getKey(c.Request.Context(), int32(operatorID))
		if err != nil || key == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "operator credentials not found or invalid",
			})
			c.Abort()
			return
		}
		signature := c.GetHeader(HeaderOperatorSignature)
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Missing X-Operator-Signature header",
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

		c.Set("operator_id", int32(operatorID))
		c.Next()
	}
}

