package db

import (
	"time"

	"github.com/google/uuid"
)

// CryptoWalletConnection represents crypto wallet connection in database
type CryptoWalletConnection struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	UserID                uuid.UUID `json:"user_id" db:"user_id"`
	WalletType            string    `json:"wallet_type" db:"wallet_type"`
	WalletAddress         string    `json:"wallet_address" db:"wallet_address"`
	WalletChain           string    `json:"wallet_chain" db:"wallet_chain"`
	WalletName            *string   `json:"wallet_name" db:"wallet_name"`
	WalletIconURL         *string   `json:"wallet_icon_url" db:"wallet_icon_url"`
	IsVerified            bool      `json:"is_verified" db:"is_verified"`
	VerificationSignature *string   `json:"verification_signature" db:"verification_signature"`
	VerificationMessage   *string   `json:"verification_message" db:"verification_message"`
	VerificationTimestamp *time.Time `json:"verification_timestamp" db:"verification_timestamp"`
	LastUsedAt            time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// CryptoWalletChallenge represents wallet verification challenge in database
type CryptoWalletChallenge struct {
	ID                uuid.UUID `json:"id" db:"id"`
	WalletAddress     string    `json:"wallet_address" db:"wallet_address"`
	WalletType        string    `json:"wallet_type" db:"wallet_type"`
	ChallengeMessage  string    `json:"challenge_message" db:"challenge_message"`
	ChallengeNonce    string    `json:"challenge_nonce" db:"challenge_nonce"`
	ExpiresAt         time.Time `json:"expires_at" db:"expires_at"`
	IsUsed            bool      `json:"is_used" db:"is_used"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// CryptoWalletAuthLog represents wallet authentication log in database
type CryptoWalletAuthLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	WalletAddress string                `json:"wallet_address" db:"wallet_address"`
	WalletType    string                `json:"wallet_type" db:"wallet_type"`
	Action        string                `json:"action" db:"action"`
	IPAddress     *string               `json:"ip_address" db:"ip_address"`
	UserAgent     *string               `json:"user_agent" db:"user_agent"`
	Success       bool                  `json:"success" db:"success"`
	ErrorMessage  *string               `json:"error_message" db:"error_message"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time             `json:"created_at" db:"created_at"`
}

// UserWalletInfo represents user wallet information for queries
type UserWalletInfo struct {
	ConnectionID   uuid.UUID `json:"connection_id" db:"connection_id"`
	WalletType     string    `json:"wallet_type" db:"wallet_type"`
	WalletAddress  string    `json:"wallet_address" db:"wallet_address"`
	WalletChain    string    `json:"wallet_chain" db:"wallet_chain"`
	WalletName     *string   `json:"wallet_name" db:"wallet_name"`
	IsVerified     bool      `json:"is_verified" db:"is_verified"`
	IsPrimary      bool      `json:"is_primary" db:"is_primary"`
	LastUsedAt     time.Time `json:"last_used_at" db:"last_used_at"`
	ConnectedAt    time.Time `json:"connected_at" db:"connected_at"`
}

// WalletConnectionWithUser represents wallet connection with user info
type WalletConnectionWithUser struct {
	ConnectionID   uuid.UUID `json:"connection_id" db:"connection_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	WalletType     string    `json:"wallet_type" db:"wallet_type"`
	WalletAddress  string    `json:"wallet_address" db:"wallet_address"`
	WalletChain    string    `json:"wallet_chain" db:"wallet_chain"`
	WalletName     *string   `json:"wallet_name" db:"wallet_name"`
	IsVerified     bool      `json:"is_verified" db:"is_verified"`
	UserPhone      *string   `json:"user_phone" db:"user_phone"`
	UserEmail      *string   `json:"user_email" db:"user_email"`
	UserFirstName  *string   `json:"user_first_name" db:"user_first_name"`
	UserLastName   *string   `json:"user_last_name" db:"user_last_name"`
	LastUsedAt     time.Time `json:"last_used_at" db:"last_used_at"`
	ConnectedAt    time.Time `json:"connected_at" db:"connected_at"`
} 