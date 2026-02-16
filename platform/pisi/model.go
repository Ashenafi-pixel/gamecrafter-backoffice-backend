package pisi

import (
	"time"

	"github.com/tucanbit/platform/logger"
)

// Login request DTO
// Matches: { "password": "pisi@12345", "vaspid": "39" }
type LoginRequest struct {
	Password string `json:"password"`
	Vaspid   string `json:"vaspid"`
}

// Login response DTO
type LoginResponse struct {
	Success                bool   `json:"Success"`
	StatusCode             string `json:"StatusCode"`
	Message                string `json:"Message"`
	Provider               string `json:"Provider"`
	PisiAuthorizationToken string `json:"Pisi-authorization-token"`
	Pisisid                int    `json:"Pisisid"`
	Expiration             string `json:"Expiration"`
}

// SendBulkSMS request DTO (form-data)
type SendBulkSMSRequest struct {
	Message    string
	Recipients string
	SenderId   string
}

// PisiClient holds config and token
type pisiClient struct {
	baseURL    string
	password   string
	vaspid     string
	token      string
	timeout    time.Duration
	retryCount int
	retryDelay time.Duration
	logger     logger.Logger
	SenderID   string
}
