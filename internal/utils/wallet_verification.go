package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// CreatePersonalMessageHash builds the Ethereum personal_sign hash (EIP-191)
func CreatePersonalMessageHash(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := []byte(prefix + message)
	return crypto.Keccak256(prefixedMessage)
}

// VerifyWalletSignature verifies if the given signature was made by the walletAddress
func VerifyWalletSignature(walletAddress, message, signature string) (bool, error) {
	// Normalize wallet address (remove 0x prefix and convert to lowercase)
	normalizedAddr := strings.ToLower(strings.TrimPrefix(walletAddress, "0x"))

	// Decode signature (remove 0x prefix)
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	if len(sigBytes) != 65 {
		return false, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(sigBytes))
	}

	// Create message hash using Ethereum personal_sign format
	msgHash := CreatePersonalMessageHash(message)

	// Handle recovery ID adjustment for EIP-155
	// MetaMask and other wallets may use different recovery ID formats
	originalRecoveryID := sigBytes[64]

	fmt.Printf("DEBUG: Wallet Address: %s\n", walletAddress)
	fmt.Printf("DEBUG: Normalized Address: %s\n", normalizedAddr)
	fmt.Printf("DEBUG: Message: %s\n", message)
	fmt.Printf("DEBUG: Signature: %s\n", signature)
	fmt.Printf("DEBUG: Original Recovery ID: 0x%02x (%d)\n", originalRecoveryID, originalRecoveryID)

	// Try the original signature first
	pubKeyBytes, err := crypto.Ecrecover(msgHash, sigBytes)
	if err == nil {
		pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
		if err == nil {
			recoveredAddr := strings.ToLower(strings.TrimPrefix(crypto.PubkeyToAddress(*pubKey).Hex(), "0x"))
			fmt.Printf("DEBUG: Recovered address with original recovery ID: %s\n", recoveredAddr)
			if recoveredAddr == normalizedAddr {
				fmt.Printf("DEBUG: Signature verification successful with original recovery ID!\n")
				return true, nil
			}
		}
	} else {
		fmt.Printf("DEBUG: Ecrecover failed with original recovery ID: %v\n", err)
	}

	// Try all possible recovery IDs
	recoveryIDs := []byte{0, 1, 27, 28}

	// Handle special cases for high recovery IDs (EIP-155)
	if originalRecoveryID > 28 {
		fmt.Printf("DEBUG: High recovery ID detected, trying base recovery IDs\n")
		// Try base recovery IDs (0, 1) for EIP-155
		recoveryIDs = append([]byte{0, 1}, recoveryIDs...)
	}

	// Special case: if original recovery ID is 28, also try 1 (common MetaMask issue)
	if originalRecoveryID == 28 {
		fmt.Printf("DEBUG: Recovery ID 28 detected, adding recovery ID 1\n")
		recoveryIDs = append([]byte{1}, recoveryIDs...)
	}

	// Try each recovery ID
	for _, recoveryID := range recoveryIDs {
		if recoveryID == originalRecoveryID {
			continue // Skip the original one we already tried
		}

		sigBytes[64] = recoveryID
		pubKeyBytes, err := crypto.Ecrecover(msgHash, sigBytes)
		if err == nil {
			pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
			if err == nil {
				recoveredAddr := strings.ToLower(strings.TrimPrefix(crypto.PubkeyToAddress(*pubKey).Hex(), "0x"))
				fmt.Printf("DEBUG: Recovered address with recovery ID 0x%02x: %s\n", recoveryID, recoveredAddr)
				if recoveredAddr == normalizedAddr {
					fmt.Printf("DEBUG: Signature verification successful with recovery ID 0x%02x!\n", recoveryID)
					return true, nil
				}
			}
		}
	}

	fmt.Printf("DEBUG: All recovery IDs failed. Expected: %s\n", normalizedAddr)
	return false, fmt.Errorf("failed to verify signature: address mismatch")
}

// VerifyWalletSignatureBool returns only true/false for convenience
func VerifyWalletSignatureBool(walletAddress, message, signature string) bool {
	ok, _ := VerifyWalletSignature(walletAddress, message, signature)
	return ok
}
