package authz

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/utils"
)

// createPersonalMessageHash creates the Ethereum personal message hash
// This follows the same format as MetaMask's personal_sign: \x19Ethereum Signed Message:\n<length><message>
func createPersonalMessageHash(message string) []byte {
	return utils.CreatePersonalMessageHash(message)
}

// VerifyWalletSignature verifies an Ethereum signature
// Takes the claimed wallet address, the original challenge message, and the hex-encoded signature
func VerifyWalletSignature(walletAddress, message, signature string) bool {
	return utils.VerifyWalletSignatureBool(walletAddress, message, signature)
}

// logSignatureVerification logs detailed signature verification process for debugging
func logSignatureVerification(walletAddress, message, signature string, messageHash []byte, signatureBytes []byte) {
	fmt.Printf("=== SIGNATURE VERIFICATION DEBUG ===\n")
	fmt.Printf("Wallet Address: %s\n", walletAddress)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Signature: %s\n", signature)
	fmt.Printf("Message Hash (hex): %x\n", messageHash)
	fmt.Printf("Signature Length: %d bytes\n", len(signatureBytes))
	fmt.Printf("Signature Bytes (hex): %x\n", signatureBytes)
	fmt.Printf("=====================================\n")
}

// recoverPublicKeyFromSignature handles EIP-155 recovery IDs properly
func recoverPublicKeyFromSignature(messageHash []byte, signatureBytes []byte) (*ecdsa.PublicKey, error) {
	var publicKey *ecdsa.PublicKey
	originalRecoveryID := signatureBytes[0]

	// Try using crypto.Ecrecover first (handles recovery IDs better)
	pubKeyBytes, err := crypto.Ecrecover(messageHash[:], signatureBytes)
	if err == nil {
		publicKey, err = crypto.UnmarshalPubkey(pubKeyBytes)
		if err == nil {
			fmt.Printf("Successfully recovered public key using Ecrecover\n")
			return publicKey, nil
		}
	} else {
		fmt.Printf("Ecrecover failed: %v\n", err)
	}

	// Fallback: try different recovery IDs
	recoveryIDs := []byte{0, 1, 27, 28}

	// Handle EIP-155 encoded recovery IDs
	if originalRecoveryID > 28 {
		fmt.Printf("Original recovery ID: 0x%02x (%d) - attempting EIP-155 decoding\n", originalRecoveryID, originalRecoveryID)

		// For EIP-155: recovery_id = base_recovery_id + 27 + 2 * chain_id
		// Try to extract the base recovery ID
		baseRecoveryID := originalRecoveryID - 27
		if baseRecoveryID <= 1 {
			recoveryIDs = append(recoveryIDs, baseRecoveryID)
			fmt.Printf("Added EIP-155 base recovery ID: 0x%02x (%d)\n", baseRecoveryID, baseRecoveryID)
		}

		// Also try with different chain ID calculations
		// Some wallets use: recovery_id = base_recovery_id + 27 + 2 * (chain_id % 2)
		altBaseRecoveryID := originalRecoveryID - 27 - 2
		if altBaseRecoveryID <= 1 {
			recoveryIDs = append(recoveryIDs, altBaseRecoveryID)
			fmt.Printf("Added alternative EIP-155 base recovery ID: 0x%02x (%d)\n", altBaseRecoveryID, altBaseRecoveryID)
		}

		// Try with the original recovery ID as well (some implementations might handle it)
		recoveryIDs = append(recoveryIDs, originalRecoveryID)
		fmt.Printf("Added original recovery ID: 0x%02x (%d)\n", originalRecoveryID, originalRecoveryID)

		// For recovery ID 106 (0x6a), try different calculations
		// This might be: 106 = 0 + 27 + 2 * 39.5 (not valid)
		// Or: 106 = 1 + 27 + 2 * 39 (not valid)
		// Let's try: 106 - 27 = 79, then 79 % 2 = 1
		if originalRecoveryID == 106 {
			// Try recovery ID 1 (since 79 % 2 = 1)
			recoveryIDs = append(recoveryIDs, 1)
			fmt.Printf("Added special case recovery ID for 106: 0x%02x (%d)\n", 1, 1)

			// Try recovery ID 0 as well
			recoveryIDs = append(recoveryIDs, 0)
			fmt.Printf("Added special case recovery ID for 106: 0x%02x (%d)\n", 0, 0)
		}

		// For recovery ID 101 (0x65), try different calculations
		// This might be: 101 = 0 + 27 + 2 * 37 (valid: 0 + 27 + 74 = 101)
		// Or: 101 = 1 + 27 + 2 * 36.5 (not valid)
		if originalRecoveryID == 101 {
			// Try recovery ID 0 (since 101 - 27 = 74, then 74 % 2 = 0)
			recoveryIDs = append(recoveryIDs, 0)
			fmt.Printf("Added special case recovery ID for 101: 0x%02x (%d)\n", 0, 0)

			// Try recovery ID 1 as well
			recoveryIDs = append(recoveryIDs, 1)
			fmt.Printf("Added special case recovery ID for 101: 0x%02x (%d)\n", 1, 1)
		}
	}

	for _, recoveryID := range recoveryIDs {
		signatureBytes[0] = recoveryID
		publicKey, err = crypto.SigToPub(messageHash[:], signatureBytes)
		if err == nil {
			fmt.Printf("Successfully recovered public key with recovery ID: 0x%02x (%d)\n", recoveryID, recoveryID)
			return publicKey, nil
		}
		fmt.Printf("Failed with recovery ID 0x%02x (%d): %v\n", recoveryID, recoveryID, err)
	}

	return nil, fmt.Errorf("failed to recover public key with any recovery ID, original was 0x%02x", originalRecoveryID)
}

// ConnectWallet connects a crypto wallet to a user account
//
//	@Summary		Connect Wallet
//	@Description	Connect a crypto wallet to the user's account
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			request			body		dto.WalletConnectionRequest	true	"Wallet Connection Request"
//	@Success		200				{object}	object
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/wallet/connect [post]
func (a *authz) ConnectWallet(c *gin.Context) {
	// Get user ID from JWT token it's set by auth middleware
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		err := errors.ErrInvalidAccessToken.Wrap(fmt.Errorf("user not authenticated"), "User not authenticated")
		_ = c.Error(err)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid user ID")
		_ = c.Error(err)
		return
	}

	var req dto.WalletConnectionRequest
	if err := c.ShouldBind(&req); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid request body")
		_ = c.Error(err)
		return
	}

	// Create wallet connection using the real module
	walletResponse, err := a.cryptoWallet.CreateWalletConnectionWithRequest(c.Request.Context(), &req, userID)
	if err != nil {
		err := errors.ErrUnableTocreate.Wrap(err, "Failed to create wallet connection")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, walletResponse)
}

// DisconnectWallet disconnects a crypto wallet from a user account
//
//	@Summary		Disconnect Wallet
//	@Description	Disconnect a crypto wallet from the user's account
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			connection_id	path		string	true	"Wallet Connection ID"
//	@Success		200				{object}	object
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/wallet/disconnect/{connection_id} [delete]
func (a *authz) DisconnectWallet(c *gin.Context) {
	connectionIDStr := c.Param("connection_id")
	if connectionIDStr == "" {
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("connection ID required"), "Connection ID is required")
		_ = c.Error(err)
		return
	}

	connectionID, err := uuid.Parse(connectionIDStr)
	if err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid connection ID")
		_ = c.Error(err)
		return
	}

	// Disconnect wallet using the real module
	req := &dto.WalletDisconnectRequest{
		ConnectionID: connectionID,
	}

	disconnectResponse, err := a.cryptoWallet.DisconnectWallet(c.Request.Context(), req)
	if err != nil {
		err := errors.ErrUnableToUpdate.Wrap(err, "Failed to disconnect wallet")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, disconnectResponse)
}

// GetUserWallets retrieves all wallets connected to a user
//
//	@Summary		Get User Wallets
//	@Description	Get all crypto wallets connected to the user's account
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	array
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		401				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Router			/api/wallet/list [get]
func (a *authz) GetUserWallets(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		err := errors.ErrInvalidAccessToken.Wrap(fmt.Errorf("user not authenticated"), "User not authenticated")
		_ = c.Error(err)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid user ID")
		_ = c.Error(err)
		return
	}

	// Get user wallets using the real module
	wallets, err := a.cryptoWallet.GetUserWallets(c.Request.Context(), userID)
	if err != nil {
		err := errors.ErrUnableToGet.Wrap(err, "Failed to get user wallets")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, wallets)
}

// CreateWalletChallenge creates a verification challenge for wallet authentication
//
//	@Summary		Create Wallet Challenge
//	@Description	Create a verification challenge for wallet authentication
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.WalletChallengeRequest	true	"Wallet Challenge Request"
//	@Success		200		{object}	object
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/wallet/challenge [post]
func (a *authz) CreateWalletChallenge(c *gin.Context) {
	var req dto.WalletChallengeRequest
	if err := c.ShouldBind(&req); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid request body")
		_ = c.Error(err)
		return
	}

	// Create wallet challenge using the real module
	challenge, err := a.cryptoWallet.CreateWalletChallenge(c.Request.Context(), &req)
	if err != nil {
		err := errors.ErrUnableTocreate.Wrap(err, "Failed to create wallet challenge")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, challenge)
}

// VerifyWalletChallenge verifies a wallet challenge signature
//
//	@Summary		Verify Wallet Challenge
//	@Description	Verify a wallet challenge signature
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.WalletVerificationRequest	true	"Wallet Verification Request"
//	@Success		200		{object}	object
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/wallet/verify [post]
func (a *authz) VerifyWalletChallenge(c *gin.Context) {
	var req dto.WalletVerificationRequest
	if err := c.ShouldBind(&req); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid request body")
		_ = c.Error(err)
		return
	}

	// Verify wallet challenge using the real module
	verification, err := a.cryptoWallet.VerifyWalletSignature(c.Request.Context(), &req)
	if err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Failed to verify wallet challenge")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, verification)
}

// LoginWithWallet authenticates a user using their crypto wallet
//
//	@Summary		Login with Wallet
//	@Description	Authenticate a user using their crypto wallet
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.WalletLoginRequest	true	"Wallet Login Request"
//	@Success		200		{object}	object
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/wallet/login [post]
func (a *authz) LoginWithWallet(c *gin.Context) {
	var req dto.WalletLoginRequest
	if err := c.ShouldBind(&req); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid request body")
		_ = c.Error(err)
		return
	}

	// Use the real crypto wallet module for authentication
	walletResponse, err := a.cryptoWallet.LoginWithWallet(c.Request.Context(), &req)
	if err != nil {
		// Provide more specific error messages based on the error type
		var errorMessage string
		if strings.Contains(err.Error(), "invalid wallet signature") {
			errorMessage = "Invalid wallet signature. Please try signing the message again."
		} else if strings.Contains(err.Error(), "failed to create new user") {
			errorMessage = "Failed to create new user account. Please try again."
		} else if strings.Contains(err.Error(), "database error") {
			errorMessage = "Database error occurred. Please try again later."
		} else if strings.Contains(err.Error(), "challenge") {
			errorMessage = "Invalid or expired challenge. Please request a new challenge."
		} else {
			errorMessage = "Wallet authentication failed: " + err.Error()
		}

		err := errors.ErrInvalidUserInput.Wrap(err, errorMessage)
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, walletResponse)
}

// TestWalletSignature is a debug endpoint to test signature verification step by step
//
//	@Summary		Test Wallet Signature
//	@Description	Debug endpoint to test wallet signature verification
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.WalletVerificationRequest	true	"Wallet Verification Request"
//	@Success		200		{object}	object
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/api/wallet/test-signature [post]
func (a *authz) TestWalletSignature(c *gin.Context) {
	var req dto.WalletVerificationRequest
	if err := c.ShouldBind(&req); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "Invalid request body")
		_ = c.Error(err)
		return
	}

	// Use the CasinoWalletService for signature verification
	err := a.cryptoWallet.VerifyEthereumSignature(req.WalletAddress, req.Message, req.Signature)
	if err != nil {
		errResp := &response.ErrorResponse{
			Code:        http.StatusBadRequest,
			Message:     "Failed to recover public key",
			Description: fmt.Sprintf("failed to recover public key from signature: %v", err),
		}
		response.SendErrorResponse(c, errResp)
		return
	}

	// If we get here, signature verification was successful
	response.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":          "Signature verification successful",
		"wallet_address":   req.WalletAddress,
		"message_verified": req.Message,
	})
}
