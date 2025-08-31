package authz

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/response"
)

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
	walletResponse, err := a.cryptoWallet.CreateWalletConnection(c.Request.Context(), &req, userID)
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
		err := errors.ErrInvalidUserInput.Wrap(err, "Wallet authentication failed")
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, walletResponse)
}
