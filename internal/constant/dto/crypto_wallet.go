package dto

import (
	"time"

	"github.com/google/uuid"
)

// CryptoWalletType represents supported crypto wallet types
type CryptoWalletType string

const (
	WalletMetaMask      CryptoWalletType = "metamask"
	WalletWalletConnect CryptoWalletType = "walletconnect"
	WalletCoinbase      CryptoWalletType = "coinbase"
	WalletPhantom       CryptoWalletType = "phantom"
	WalletTrust         CryptoWalletType = "trust"
	WalletLedger        CryptoWalletType = "ledger"
)

// WalletConnectionRequest represents wallet connection request
type WalletConnectionRequest struct {
	WalletType    CryptoWalletType `json:"wallet_type" validate:"required,oneof=metamask walletconnect coinbase phantom trust ledger"`
	WalletAddress string           `json:"wallet_address" validate:"required,min=42,max=255"`
	WalletChain   string           `json:"wallet_chain" validate:"required"`
	WalletName    string           `json:"wallet_name,omitempty"`
	WalletIconURL string           `json:"wallet_icon_url,omitempty"`
}

// WalletConnectionResponse represents wallet connection response
type WalletConnectionResponse struct {
	Message       string    `json:"message"`
	ConnectionID  uuid.UUID `json:"connection_id"`
	WalletAddress string    `json:"wallet_address"`
	WalletType    string    `json:"wallet_type"`
	ChainType     string    `json:"chain_type"`
	IsVerified    bool      `json:"is_verified"`
	ConnectedAt   time.Time `json:"connected_at"`
}

// WalletLoginRequest represents wallet-based login request
type WalletLoginRequest struct {
	WalletType    CryptoWalletType `json:"wallet_type" validate:"required,oneof=metamask walletconnect coinbase phantom trust ledger"`
	WalletAddress string           `json:"wallet_address" validate:"required,min=42,max=255"`
	ChainType     string           `json:"chain_type,omitempty"`
	Signature     string           `json:"signature" validate:"required"`
	Message       string           `json:"message" validate:"required"`
	Nonce         string           `json:"nonce" validate:"required"`
}

// WalletLoginResponse represents wallet-based login response
type WalletLoginResponse struct {
	Message       string      `json:"message"`
	UserID        string      `json:"user_id"`
	AccessToken   string      `json:"access_token"`
	RefreshToken  string      `json:"refresh_token"`
	WalletAddress string      `json:"wallet_address"`
	ChainType     string      `json:"chain_type"`
	IsNewUser     bool        `json:"is_new_user"`
	ExpiresAt     time.Time   `json:"expires_at"`
	UserProfile   UserProfile `json:"user_profile,omitempty"`
}

// WalletChallengeRequest represents wallet verification challenge request
type WalletChallengeRequest struct {
	WalletType    CryptoWalletType `json:"wallet_type" validate:"required,oneof=metamask walletconnect coinbase phantom trust ledger"`
	WalletAddress string           `json:"wallet_address" validate:"required,min=42,max=255"`
}

// WalletChallengeResponse represents wallet verification challenge response
type WalletChallengeResponse struct {
	Message          string    `json:"message"`
	ChallengeMessage string    `json:"challenge_message"`
	Nonce            string    `json:"nonce"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// WalletVerificationRequest represents wallet verification request
type WalletVerificationRequest struct {
	WalletType    CryptoWalletType `json:"wallet_type" validate:"required,oneof=metamask walletconnect coinbase phantom trust ledger"`
	WalletAddress string           `json:"wallet_address" validate:"required,min=42,max=255"`
	ChainType     string           `json:"chain_type,omitempty"` // <-- THE FINAL FIX IS HERE
	Signature     string           `json:"signature" validate:"required"`
	Message       string           `json:"message" validate:"required"`
	Nonce         string           `json:"nonce" validate:"required"`
}

// WalletVerificationResponse represents wallet verification response
type WalletVerificationResponse struct {
	Message       string    `json:"message"`
	ConnectionID  uuid.UUID `json:"connection_id"`
	WalletAddress string    `json:"wallet_address"`
	ChainType     string    `json:"chain_type"`
	IsVerified    bool      `json:"is_verified"`
	VerifiedAt    time.Time `json:"verified_at"`
}

// WalletDisconnectRequest represents wallet disconnection request
type WalletDisconnectRequest struct {
	ConnectionID uuid.UUID `json:"connection_id" validate:"required"`
}

// WalletDisconnectResponse represents wallet disconnection response
type WalletDisconnectResponse struct {
	Message string `json:"message"`
}

// UserWalletsResponse represents user's connected wallets response
type UserWalletsResponse struct {
	Wallets []WalletInfo `json:"wallets"`
	Total   int          `json:"total"`
}

// WalletInfo represents wallet information
type WalletInfo struct {
	ConnectionID  uuid.UUID `json:"connection_id"`
	WalletType    string    `json:"wallet_type"`
	WalletAddress string    `json:"wallet_address"`
	WalletChain   string    `json:"wallet_chain"`
	WalletName    string    `json:"wallet_name"`
	IsVerified    bool      `json:"is_verified"`
	IsPrimary     bool      `json:"is_primary"`
	LastUsedAt    time.Time `json:"last_used_at"`
	ConnectedAt   time.Time `json:"connected_at"`
}

// WalletAuthLog represents wallet authentication log entry
type WalletAuthLog struct {
	WalletAddress string                 `json:"wallet_address"`
	WalletType    CryptoWalletType       `json:"wallet_type"`
	Action        string                 `json:"action"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateUserRequest represents user creation request for wallet authentication
type CreateUserRequest struct {
	Username      string `json:"username"`
	WalletAddress string `json:"wallet_address"`
	ChainType     string `json:"chain_type"`
}
