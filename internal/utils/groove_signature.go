package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// GrooveSignatureValidator handles HMAC-SHA256 signature validation for GrooveTech APIs
type GrooveSignatureValidator struct {
	secretKey string
}

// NewGrooveSignatureValidator creates a new signature validator
func NewGrooveSignatureValidator(secretKey string) *GrooveSignatureValidator {
	return &GrooveSignatureValidator{
		secretKey: secretKey,
	}
}

// ValidateSignature validates the HMAC-SHA256 signature from GrooveTech
func (v *GrooveSignatureValidator) ValidateSignature(signature, method, path string, params map[string]string) bool {
	expectedSignature := v.GenerateSignature(method, path, params)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateSignature generates HMAC-SHA256 signature for outgoing requests to GrooveTech
func (v *GrooveSignatureValidator) GenerateSignature(method, path string, params map[string]string) string {
	// Create query string from parameters (excluding 'request' parameter)
	queryParams := make([]string, 0, len(params))

	for key, value := range params {
		if key != "request" {
			// Treat nogsgameid as gameid for sorting purposes as per GrooveTech docs
			sortKey := key
			if key == "nogsgameid" {
				sortKey = "gameid"
			}
			queryParams = append(queryParams, fmt.Sprintf("%s:%s", sortKey, value))
		}
	}

	// Sort parameters alphabetically
	sort.Strings(queryParams)

	// Concatenate values (without keys)
	concatenatedValues := strings.Join(queryParams, "")
	for i, param := range queryParams {
		parts := strings.SplitN(param, ":", 2)
		if len(parts) == 2 {
			queryParams[i] = parts[1]
		}
	}
	concatenatedValues = strings.Join(queryParams, "")

	// Create HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(v.secretKey))
	mac.Write([]byte(concatenatedValues))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature
}

// GenerateSignatureFromQuery generates signature from URL query string
func (v *GrooveSignatureValidator) GenerateSignatureFromQuery(queryString string) string {
	// Parse query string
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return ""
	}

	// Convert to map
	params := make(map[string]string)
	for key, vals := range values {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}

	return v.GenerateSignature("GET", "", params)
}

// ValidateGrooveSignature validates signature from GrooveTech request
func (v *GrooveSignatureValidator) ValidateGrooveSignature(signature string, queryParams map[string]string) bool {
	// Remove 'request' parameter if present
	delete(queryParams, "request")

	// Generate expected signature
	expectedSignature := v.GenerateSignature("GET", "", queryParams)

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateGrooveSignature generates signature for requests to GrooveTech
func (v *GrooveSignatureValidator) GenerateGrooveSignature(params map[string]string) string {
	return v.GenerateSignature("GET", "", params)
}
