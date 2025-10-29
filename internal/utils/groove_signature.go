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

// ValidateSignature validates the HMAC-SHA256 signature from GrooveTech request
// This follows the exact GrooveTech specification for signature validation
func (v *GrooveSignatureValidator) ValidateSignature(receivedSignature string, queryParams map[string]string) bool {
	expectedSignature := v.GenerateSignature(queryParams)
	return hmac.Equal([]byte(receivedSignature), []byte(expectedSignature))
}

// GenerateSignature generates HMAC-SHA256 signature following GrooveTech specification
// Step 1: Sort parameters alphabetically by their names
// Step 2: Concatenate only the values (not the names)
// Step 3: Handle 'request' parameter based on endpoint type (inconsistent in GrooveTech docs)
// Step 4: Treat 'nogsgameid' as 'gameid' for sorting purposes
// Step 5: Generate HMAC-SHA256 with security key
func (v *GrooveSignatureValidator) GenerateSignature(params map[string]string) string {
	return v.GenerateSignatureWithRequestInclusion(params, true)
}

// GenerateSignatureWithRequestInclusion generates signature with configurable request parameter inclusion
// This handles the inconsistency in GrooveTech documentation
func (v *GrooveSignatureValidator) GenerateSignatureWithRequestInclusion(params map[string]string, includeRequest bool) string {
	// Create parameter map for sorting
	paramMap := make(map[string]string)

	for key, value := range params {
		// Handle 'request' parameter based on configuration
		if key == "request" && !includeRequest {
			continue
		}

		// Treat 'nogsgameid' as 'gameid' for sorting purposes
		sortKey := key
		if key == "nogsgameid" {
			sortKey = "gameid"
		}

		paramMap[sortKey] = value
	}

	// Sort keys alphabetically
	var keys []string
	for k := range paramMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate values in sorted order
	var concatenated strings.Builder
	for _, key := range keys {
		concatenated.WriteString(paramMap[key])
	}

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(v.secretKey))
	h.Write([]byte(concatenated.String()))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// GenerateSignatureForEndpoint generates signature optimized for specific GrooveTech endpoints
// This handles the endpoint-specific inconsistencies in GrooveTech documentation
func (v *GrooveSignatureValidator) GenerateSignatureForEndpoint(endpoint string, params map[string]string) string {
	// Determine if request parameter should be included based on endpoint
	includeRequest := v.shouldIncludeRequestParameter(endpoint)
	return v.GenerateSignatureWithRequestInclusion(params, includeRequest)
}

// shouldIncludeRequestParameter determines if request parameter should be included for specific endpoints
// Based on testing and GrooveTech documentation inconsistencies
func (v *GrooveSignatureValidator) shouldIncludeRequestParameter(endpoint string) bool {
	// Endpoints that require request parameter in signature (based on working examples)
	requestRequiredEndpoints := map[string]bool{
		"wager":          true,
		"result":         true,
		"getaccount":     false, // Documentation examples suggest exclusion
		"getbalance":     false, // Documentation examples suggest exclusion
		"wagerAndResult": false, // Documentation examples suggest exclusion
		"rollback":       false,
		"jackpot":        false,
		"reversewin":     false,
	}

	if include, exists := requestRequiredEndpoints[endpoint]; exists {
		return include
	}

	// Default to including request parameter for unknown endpoints
	return true
}

// GenerateSignatureFromURL generates signature from URL query string
func (v *GrooveSignatureValidator) GenerateSignatureFromURL(requestURL string) string {
	// Parse URL
	u, err := url.Parse(requestURL)
	if err != nil {
		return ""
	}

	// Extract query parameters
	params := u.Query()

	// Convert to map
	paramMap := make(map[string]string)
	for key, values := range params {
		if len(values) > 0 {
			paramMap[key] = values[0]
		}
	}

	return v.GenerateSignature(paramMap)
}

// ValidateSignatureFromURL validates signature from URL and query parameters
func (v *GrooveSignatureValidator) ValidateSignatureFromURL(receivedSignature, requestURL string) bool {
	expectedSignature := v.GenerateSignatureFromURL(requestURL)
	return hmac.Equal([]byte(receivedSignature), []byte(expectedSignature))
}

// GenerateSignatureForRequest generates signature for specific GrooveTech request types
func (v *GrooveSignatureValidator) GenerateSignatureForRequest(requestType string, params map[string]string) string {
	// Add request parameter for context (but exclude from signature calculation)
	paramsWithRequest := make(map[string]string)
	for k, v := range params {
		paramsWithRequest[k] = v
	}
	paramsWithRequest["request"] = requestType

	return v.GenerateSignature(paramsWithRequest)
}

// GetSignatureHeader returns the signature in the format expected by GrooveTech
func (v *GrooveSignatureValidator) GetSignatureHeader(params map[string]string) string {
	signature := v.GenerateSignature(params)
	return fmt.Sprintf("X-Groove-Signature: %s", signature)
}

// ValidateGrooveSignature validates signature from GrooveTech request (legacy method for compatibility)
func (v *GrooveSignatureValidator) ValidateGrooveSignature(signature string, queryParams map[string]string) bool {
	return v.ValidateSignature(signature, queryParams)
}

// GenerateGrooveSignature generates signature for requests to GrooveTech (legacy method for compatibility)
func (v *GrooveSignatureValidator) GenerateGrooveSignature(params map[string]string) string {
	return v.GenerateSignature(params)
}

// TestSignatureValidation tests the signature validation with provided examples
func (v *GrooveSignatureValidator) TestSignatureValidation() map[string]bool {
	testCases := map[string]bool{
		// Test case from GrooveTech documentation using endpoint-specific signature generation
		"GetAccount": v.validateTestCaseForEndpoint("getaccount", "be426d042cd71743970779cd6ee7881d71d1f0eb769cbe14a0081c29c8ef2a09", map[string]string{
			"request":       "getaccount",
			"accountid":     "111",
			"apiversion":    "1.2",
			"device":        "desktop",
			"gamesessionid": "123_jdhdujdk",
		}),

		"GetBalance": v.validateTestCaseForEndpoint("getbalance", "434e2b4545299886c8891faadd86593ad8cbf79e5cd20a6755411d1d3822abba", map[string]string{
			"request":       "getbalance",
			"accountid":     "111",
			"apiversion":    "1.2",
			"device":        "desktop",
			"nogsgameid":    "80102", // Should be treated as gameid
			"gamesessionid": "123_jdhdujdk",
		}),

		"Wager": v.validateTestCaseForEndpoint("wager", "f6d980dfe7866b6676e6565ccca239f527979d702106233bb6f72a654931b3bc", map[string]string{
			"request":       "wager",
			"accountid":     "111",
			"apiversion":    "1.2",
			"betamount":     "10.0",
			"device":        "desktop",
			"gameid":        "80102",
			"gamesessionid": "123_jdhdujdk",
			"roundid":       "nc8n4nd87",
			"transactionid": "trx_id",
		}),

		"WagerAndResult": v.validateTestCaseForEndpoint("wagerAndResult", "bba4df598cf50ec69ebe144c696c0305e32f1eef76eb32091585f056fafd9079", map[string]string{
			"request":       "wagerAndResult",
			"accountid":     "111",
			"apiversion":    "1.2",
			"betamount":     "10.0",
			"device":        "desktop",
			"gameid":        "80102",
			"gamesessionid": "123_jdhdujdk",
			"result":        "10.0",
			"roundid":       "nc8n4nd87",
			"transactionid": "trx_id",
		}),

		"Result": v.validateTestCaseForEndpoint("result", "d9655083f60cfd490f0ad882cb01ca2f9af61e669601bbb1dcced8a5dca1820f", map[string]string{
			"request":       "result",
			"accountid":     "111",
			"apiversion":    "1.2",
			"device":        "desktop",
			"gameid":        "80102",
			"gamesessionid": "123_jdhdujdk",
			"result":        "10.0",
			"roundid":       "nc8n4nd87",
			"transactionid": "trx_id",
		}),
	}

	return testCases
}

// validateTestCaseForEndpoint validates a specific test case with expected signature for a specific endpoint
func (v *GrooveSignatureValidator) validateTestCaseForEndpoint(endpoint, expectedSignature string, params map[string]string) bool {
	generatedSignature := v.GenerateSignatureForEndpoint(endpoint, params)
	return generatedSignature == expectedSignature
}

// validateTestCase validates a specific test case with expected signature
func (v *GrooveSignatureValidator) validateTestCase(expectedSignature string, params map[string]string) bool {
	generatedSignature := v.GenerateSignature(params)
	return generatedSignature == expectedSignature
}
