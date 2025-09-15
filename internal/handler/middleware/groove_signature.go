package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/utils"
)

// GrooveSignatureMiddleware validates HMAC-SHA256 signatures for GrooveTech API requests
func GrooveSignatureMiddleware(secretKey string) gin.HandlerFunc {
	validator := utils.NewGrooveSignatureValidator(secretKey)

	return func(c *gin.Context) {
		// Get the signature from the X-Groove-Signature header
		signature := c.GetHeader("X-Groove-Signature")
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "Missing X-Groove-Signature header",
			})
			c.Abort()
			return
		}

		// Extract query parameters
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Validate the signature
		if !validator.ValidateSignature(signature, queryParams) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			c.Abort()
			return
		}

		// Signature is valid, continue to the next handler
		c.Next()
	}
}

// GrooveSignatureMiddlewareOptional validates signatures only if the header is present
// This is useful for environments where signature validation is optional
func GrooveSignatureMiddlewareOptional(secretKey string) gin.HandlerFunc {
	validator := utils.NewGrooveSignatureValidator(secretKey)

	return func(c *gin.Context) {
		// Get the signature from the X-Groove-Signature header
		signature := c.GetHeader("X-Groove-Signature")

		// If no signature header is present, skip validation
		if signature == "" {
			c.Next()
			return
		}

		// Extract query parameters
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Validate the signature
		if !validator.ValidateSignature(signature, queryParams) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			c.Abort()
			return
		}

		// Signature is valid, continue to the next handler
		c.Next()
	}
}

// GenerateGrooveSignature generates a signature for outgoing requests to GrooveTech
func GenerateGrooveSignature(secretKey string, params map[string]string) string {
	validator := utils.NewGrooveSignatureValidator(secretKey)
	return validator.GenerateSignature(params)
}

// GenerateGrooveSignatureFromURL generates a signature from a URL
func GenerateGrooveSignatureFromURL(secretKey, requestURL string) string {
	validator := utils.NewGrooveSignatureValidator(secretKey)
	return validator.GenerateSignatureFromURL(requestURL)
}

// ValidateGrooveSignature validates a signature against query parameters
func ValidateGrooveSignature(secretKey, signature string, params map[string]string) bool {
	validator := utils.NewGrooveSignatureValidator(secretKey)
	return validator.ValidateSignature(signature, params)
}

// ExtractQueryParamsFromURL extracts query parameters from a URL string
func ExtractQueryParamsFromURL(requestURL string) map[string]string {
	u, err := url.Parse(requestURL)
	if err != nil {
		return make(map[string]string)
	}

	params := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params
}

// BuildGrooveURL builds a GrooveTech URL with query parameters
func BuildGrooveURL(baseURL, endpoint string, params map[string]string) string {
	u, err := url.Parse(baseURL + endpoint)
	if err != nil {
		return ""
	}

	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}

	u.RawQuery = query.Encode()
	return u.String()
}

// GetGrooveSignatureHeader returns the signature header for a request
func GetGrooveSignatureHeader(secretKey string, params map[string]string) string {
	validator := utils.NewGrooveSignatureValidator(secretKey)
	return validator.GetSignatureHeader(params)
}
