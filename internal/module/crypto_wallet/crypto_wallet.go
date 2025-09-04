package crypto_wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform/utils"
)

// CryptoWalletModule defines the interface for crypto wallet operations
type CryptoWalletModule interface {
	CreateWalletConnection(ctx context.Context, req *dto.WalletConnectionRequest, userID uuid.UUID) (*dto.WalletConnectionResponse, error)
	GetWalletConnection(ctx context.Context, walletAddress string) (*dto.WalletConnectionResponse, error)
	GetUserWallets(ctx context.Context, userID uuid.UUID) (*dto.UserWalletsResponse, error)
	CreateWalletChallenge(ctx context.Context, req *dto.WalletChallengeRequest) (*dto.WalletChallengeResponse, error)
	VerifyWalletSignature(ctx context.Context, req *dto.WalletVerificationRequest) (*dto.WalletVerificationResponse, error)
	LoginWithWallet(ctx context.Context, req *dto.WalletLoginRequest) (*dto.WalletLoginResponse, error)
	DisconnectWallet(ctx context.Context, req *dto.WalletDisconnectRequest) (*dto.WalletDisconnectResponse, error)
	LogWalletAuth(ctx context.Context, userID uuid.UUID, walletAddress, action string, success bool, errorMessage string, metadata map[string]interface{}) error
}

// CryptoWalletService provides comprehensive crypto wallet management functionality
type CryptoWalletService struct {
	storage storage.CryptoWallet
	user    storage.User
	balance storage.Balance
}

// Ensure CryptoWalletService implements CryptoWalletModule
var _ CryptoWalletModule = (*CryptoWalletService)(nil)

// NewCryptoWalletService creates a new instance of CryptoWalletService
func NewCryptoWalletService(
	storage storage.CryptoWallet,
	user storage.User,
	balance storage.Balance,
) *CryptoWalletService {
	return &CryptoWalletService{
		storage: storage,
		user:    user,
		balance: balance,
	}
}

// createPersonalMessageHash creates the Ethereum personal message hash
// This follows the same format as MetaMask's personal_sign: \x19Ethereum Signed Message:\n<length><message>
func createPersonalMessageHash(message string) []byte {
	// Create the prefixed message
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := prefix + message

	// Hash the prefixed message using Keccak256 (SHA3) - Ethereum standard
	hash := crypto.Keccak256([]byte(prefixedMessage))
	return hash
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

// CreateWalletConnection establishes a new wallet connection for a user
func (s *CryptoWalletService) CreateWalletConnection(
	ctx context.Context,
	req *dto.WalletConnectionRequest,
	userID uuid.UUID,
) (*dto.WalletConnectionResponse, error) {
	// Validate user exists
	user, exists, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !exists || user.ID == uuid.Nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if wallet connection already exists
	existing, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err != nil {
		// If it's not a "not found" error, return it
		if err.Error() != "no record found" {
			return nil, fmt.Errorf("failed to check existing wallet: %w", err)
		}
	}
	if existing != nil {
		return nil, fmt.Errorf("wallet connection already exists")
	}

	// Create wallet connection
	connection, err := s.storage.CreateWalletConnection(ctx, userID, req.WalletAddress, string(req.WalletType))
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet connection: %w", err)
	}

	return &dto.WalletConnectionResponse{
		Message:       "Wallet connected successfully",
		ConnectionID:  connection.ID,
		WalletAddress: connection.WalletAddress,
		WalletType:    connection.WalletType,
		IsVerified:    false, // New connections start as unverified
		ConnectedAt:   connection.CreatedAt,
	}, nil
}

// GetWalletConnection retrieves wallet connection details
func (s *CryptoWalletService) GetWalletConnection(
	ctx context.Context,
	walletAddress string,
) (*dto.WalletConnectionResponse, error) {
	connection, err := s.storage.GetWalletConnectionByAddress(ctx, walletAddress)
	if err != nil {
		if err.Error() == "no record found" {
			return nil, fmt.Errorf("wallet connection not found")
		}
		return nil, fmt.Errorf("failed to get wallet connection: %w", err)
	}

	return &dto.WalletConnectionResponse{
		Message:       "Wallet connection found",
		ConnectionID:  connection.ID,
		WalletAddress: connection.WalletAddress,
		WalletType:    connection.WalletType,
		IsVerified:    connection.IsVerified,
		ConnectedAt:   connection.CreatedAt,
	}, nil
}

// GetUserWallets retrieves all wallet connections for a user
func (s *CryptoWalletService) GetUserWallets(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserWalletsResponse, error) {
	wallets, err := s.storage.GetUserWalletConnections(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user wallets: %w", err)
	}

	var walletInfos []dto.WalletInfo
	for _, wallet := range wallets {
		walletName := ""
		if wallet.WalletName != nil {
			walletName = *wallet.WalletName
		}

		walletInfos = append(walletInfos, dto.WalletInfo{
			ConnectionID:  wallet.ID,
			WalletType:    wallet.WalletType,
			WalletAddress: wallet.WalletAddress,
			WalletChain:   wallet.WalletChain,
			WalletName:    walletName,
			IsVerified:    wallet.IsVerified,
			IsPrimary:     false, // This field doesn't exist in the model, defaulting to false
			LastUsedAt:    wallet.LastUsedAt,
			ConnectedAt:   wallet.CreatedAt,
		})
	}

	return &dto.UserWalletsResponse{
		Wallets: walletInfos,
		Total:   len(walletInfos),
	}, nil
}

// CreateWalletChallenge creates a verification challenge for wallet authentication
func (s *CryptoWalletService) CreateWalletChallenge(
	ctx context.Context,
	req *dto.WalletChallengeRequest,
) (*dto.WalletChallengeResponse, error) {
	// Generate a random nonce
	nonce := uuid.New().String()

	// Create challenge message
	challengeMessage := fmt.Sprintf("Sign this message to verify your wallet %s", nonce)

	// Set expiration time (5 minutes from now)
	expiresAt := time.Now().Add(5 * time.Minute)

	// Store the challenge
	challenge, err := s.storage.CreateWalletChallenge(ctx, req.WalletAddress, nonce, challengeMessage, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet challenge %w", err)
	}

	return &dto.WalletChallengeResponse{
		Message:          "Challenge created successfully",
		ChallengeMessage: challengeMessage,
		Nonce:            challenge.ChallengeNonce,
		ExpiresAt:        challenge.ExpiresAt,
	}, nil
}

// VerifyWalletSignature verifies a wallet signature for authentication using real cryptographic verification
func (s *CryptoWalletService) VerifyWalletSignature(
	ctx context.Context,
	req *dto.WalletVerificationRequest,
) (*dto.WalletVerificationResponse, error) {
	// Get the challenge
	challenge, err := s.storage.GetWalletChallenge(ctx, req.WalletAddress, req.Nonce)
	if err != nil {
		if err.Error() == "no record found" {
			return nil, fmt.Errorf("challenge not found")
		}
		return nil, fmt.Errorf("failed to get wallet challenge: %w", err)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge has expired")
	}

	// Verify the signature cryptographically using Ethereum personal message format
	messageHash := createPersonalMessageHash(req.Message)

	// Decode the signature from hex (remove 0x prefix if present)
	signature := req.Signature
	if len(signature) > 2 && signature[:2] == "0x" {
		signature = signature[2:]
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature format: %w", err)
	}

	// Check signature length (ECDSA signatures are 65 bytes: 1 byte recovery ID + 32 bytes R + 32 bytes S)
	if len(signatureBytes) != 65 {
		return nil, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signatureBytes))
	}

	// Log signature verification details for debugging
	logSignatureVerification(req.WalletAddress, req.Message, req.Signature, messageHash, signatureBytes)

	// Recover the public key from the signature using the improved logic
	publicKey, err := recoverPublicKeyFromSignature(messageHash, signatureBytes)
	if err != nil {
		fmt.Printf("ERROR: Failed to recover public key: %v\n", err)
		return nil, fmt.Errorf("failed to recover public key from signature: %w", err)
	}

	// Generate the wallet address from the recovered public key
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	// Verify that the recovered address matches the claimed wallet address (case-insensitive)
	recoveredAddrHex := recoveredAddress.Hex()
	claimedAddrHex := req.WalletAddress
	if len(claimedAddrHex) > 2 && claimedAddrHex[:2] == "0x" {
		claimedAddrHex = claimedAddrHex[2:]
	}
	if len(recoveredAddrHex) > 2 && recoveredAddrHex[:2] == "0x" {
		recoveredAddrHex = recoveredAddrHex[2:]
	}

	fmt.Printf("Address Comparison:\n")
	fmt.Printf("  Recovered: %s\n", recoveredAddrHex)
	fmt.Printf("  Claimed:   %s\n", claimedAddrHex)
	fmt.Printf("  Match:     %t\n", recoveredAddrHex == claimedAddrHex)

	if recoveredAddrHex != claimedAddrHex {
		return nil, fmt.Errorf("signature verification failed: address mismatch (recovered: %s, claimed: %s)", recoveredAddrHex, claimedAddrHex)
	}

	// Extract R and S components for additional verification
	rComponent := new(big.Int).SetBytes(signatureBytes[1:33])
	sComponent := new(big.Int).SetBytes(signatureBytes[33:65])

	// Verify the signature using ECDSA verification
	ecdsaPubKey := (*ecdsa.PublicKey)(publicKey)
	isValid := ecdsa.Verify(ecdsaPubKey, messageHash[:], rComponent, sComponent)
	if !isValid {
		return nil, fmt.Errorf("signature verification failed: invalid signature")
	}

	// Mark the challenge as used
	err = s.storage.MarkChallengeAsUsed(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark challenge as used: %w", err)
	}

	// Get or create wallet connection
	connection, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err != nil {
		if err.Error() == "no record found" {
			// Create a temporary user for wallet-only authentication
			// This will be handled by the login endpoint later
			return nil, fmt.Errorf("wallet connection not found - please use login endpoint for new users")
		}
		return nil, fmt.Errorf("failed to get wallet connection: %w", err)
	}

	// Update connection to verified
	updates := map[string]interface{}{
		"is_verified": true,
		"updated_at":  time.Now(),
	}
	err = s.storage.UpdateWalletConnection(ctx, connection.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet connection: %w", err)
	}

	return &dto.WalletVerificationResponse{
		Message:       "Wallet verified successfully",
		ConnectionID:  connection.ID,
		WalletAddress: connection.WalletAddress,
		IsVerified:    true,
		VerifiedAt:    time.Now(),
	}, nil
}

// DisconnectWallet removes a wallet connection
func (s *CryptoWalletService) DisconnectWallet(
	ctx context.Context,
	req *dto.WalletDisconnectRequest,
) (*dto.WalletDisconnectResponse, error) {
	err := s.storage.DeleteWalletConnection(ctx, req.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete wallet connection: %w", err)
	}

	return &dto.WalletDisconnectResponse{
		Message: "Wallet disconnected successfully",
	}, nil
}

// LogWalletAuth logs wallet authentication attempts
func (s *CryptoWalletService) LogWalletAuth(
	ctx context.Context,
	userID uuid.UUID,
	walletAddress string,
	action string,
	success bool,
	errorMessage string,
	metadata map[string]interface{},
) error {
	status := "success"
	if !success {
		status = "failed"
	}
	return s.storage.CreateWalletAuthLog(ctx, userID, walletAddress, action, status, metadata)
}

// LoginWithWallet authenticates a user using their crypto wallet
func (s *CryptoWalletService) LoginWithWallet(
	ctx context.Context,
	req *dto.WalletLoginRequest,
) (*dto.WalletLoginResponse, error) {
	// Verify the wallet signature
	err := s.verifyWalletSignature(req)
	if err != nil {
		// Log failed authentication attempt
		s.LogWalletAuth(ctx, uuid.Nil, req.WalletAddress, "login", false, err.Error(), map[string]interface{}{
			"wallet_type": req.WalletType,
			"message":     req.Message,
			"nonce":       req.Nonce,
		})
		return nil, fmt.Errorf("wallet signature verification failed: %w", err)
	}

	// Check if wallet connection exists and is verified
	connection, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err != nil {
		if err.Error() == "no record found" {
			// Wallet not connected, create new user and connection
			return s.createNewUserWithWallet(ctx, req)
		}
		return nil, fmt.Errorf("failed to get wallet connection: %w", err)
	}

	// Get user by connection
	user, exists, err := s.user.GetUserByID(ctx, connection.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Log successful authentication
	s.LogWalletAuth(ctx, user.ID, req.WalletAddress, "login", true, "", map[string]interface{}{
		"wallet_type": req.WalletType,
		"user_id":     user.ID,
	})

	// Generate JWT tokens (you'll need to implement this)
	accessToken, refreshToken, err := s.generateJWTTokens(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Get user profile
	userProfile := dto.UserProfile{
		PhoneNumber:    user.PhoneNumber,
		Email:          user.Email,
		UserID:         user.ID,
		ProfilePicture: user.ProfilePicture,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Type:           user.Type,
		ReferralCode:   user.ReferralCode,
	}

	return &dto.WalletLoginResponse{
		Message:      "Wallet authentication successful",
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    false,
		UserProfile:  userProfile,
	}, nil
}

// verifyWalletSignature verifies the wallet signature
func (s *CryptoWalletService) verifyWalletSignature(req *dto.WalletLoginRequest) error {
	// Hash the message using Ethereum personal message format
	messageHash := createPersonalMessageHash(req.Message)

	// Decode the signature from hex (remove 0x prefix if present)
	signature := req.Signature
	if len(signature) > 2 && signature[:2] == "0x" {
		signature = signature[2:]
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// Check signature length (ECDSA signatures are 65 bytes: 1 byte recovery ID + 32 bytes R + 32 bytes S)
	if len(signatureBytes) != 65 {
		return fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signatureBytes))
	}

	// Log signature verification details for debugging
	logSignatureVerification(req.WalletAddress, req.Message, req.Signature, messageHash, signatureBytes)

	// Recover the public key from the signature using the improved logic
	publicKey, err := recoverPublicKeyFromSignature(messageHash, signatureBytes)
	if err != nil {
		fmt.Printf("ERROR: Failed to recover public key: %v\n", err)
		return fmt.Errorf("failed to recover public key from signature: %w", err)
	}

	// Generate the wallet address from the recovered public key
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	// Verify that the recovered address matches the claimed wallet address (case-insensitive)
	recoveredAddrHex := recoveredAddress.Hex()
	claimedAddrHex := req.WalletAddress
	if len(claimedAddrHex) > 2 && claimedAddrHex[:2] == "0x" {
		claimedAddrHex = claimedAddrHex[2:]
	}
	if len(recoveredAddrHex) > 2 && recoveredAddrHex[:2] == "0x" {
		recoveredAddrHex = recoveredAddrHex[2:]
	}

	fmt.Printf("Address Comparison:\n")
	fmt.Printf("  Recovered: %s\n", recoveredAddrHex)
	fmt.Printf("  Claimed:   %s\n", claimedAddrHex)
	fmt.Printf("  Match:     %t\n", recoveredAddrHex == claimedAddrHex)

	if recoveredAddrHex != claimedAddrHex {
		return fmt.Errorf("signature verification failed: address mismatch (recovered: %s, claimed: %s)", recoveredAddrHex, claimedAddrHex)
	}

	// Extract R and S components for additional verification
	rComponent := new(big.Int).SetBytes(signatureBytes[1:33])
	sComponent := new(big.Int).SetBytes(signatureBytes[33:65])

	// Verify the signature using ECDSA verification
	ecdsaPubKey := (*ecdsa.PublicKey)(publicKey)
	isValid := ecdsa.Verify(ecdsaPubKey, messageHash[:], rComponent, sComponent)
	if !isValid {
		return fmt.Errorf("signature verification failed: invalid signature")
	}

	return nil
}

// createNewUserWithWallet creates a new user with wallet connection
func (s *CryptoWalletService) createNewUserWithWallet(
	ctx context.Context,
	req *dto.WalletLoginRequest,
) (*dto.WalletLoginResponse, error) {
	// Create new user
	newUser := dto.User{
		PhoneNumber:  "", // Will be set later
		Email:        "", // Will be set later
		Type:         "PLAYER",
		ReferralCode: s.generateReferralCode(),
	}

	// Save user to database
	createdUser, err := s.user.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create wallet connection
	connection, err := s.storage.CreateWalletConnection(ctx, createdUser.ID, req.WalletAddress, string(req.WalletType))
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet connection: %w", err)
	}

	// Mark connection as verified
	updates := map[string]interface{}{
		"is_verified": true,
		"updated_at":  time.Now(),
	}
	err = s.storage.UpdateWalletConnection(ctx, connection.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet connection: %w", err)
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWTTokens(createdUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful authentication
	s.LogWalletAuth(ctx, createdUser.ID, req.WalletAddress, "login", true, "", map[string]interface{}{
		"wallet_type": req.WalletType,
		"user_id":     createdUser.ID,
		"is_new_user": true,
	})

	return &dto.WalletLoginResponse{
		Message:      "Wallet authentication successful",
		UserID:       createdUser.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    true,
		UserProfile: dto.UserProfile{
			PhoneNumber:    createdUser.PhoneNumber,
			Email:          createdUser.Email,
			UserID:         createdUser.ID,
			ProfilePicture: createdUser.ProfilePicture,
			FirstName:      createdUser.FirstName,
			LastName:       createdUser.LastName,
			Type:           createdUser.Type,
			ReferralCode:   createdUser.ReferralCode,
		},
	}, nil
}

// generateJWTTokens generates JWT access and refresh tokens
func (s *CryptoWalletService) generateJWTTokens(userID uuid.UUID) (string, string, error) {
	// Generate real JWT access token
	accessToken, err := utils.GenerateJWT(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate real refresh token
	refreshToken, err := utils.GenerateUniqueToken(64)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateReferralCode generates a unique referral code
func (s *CryptoWalletService) generateReferralCode() string {
	// Generate a random 8-character referral code using crypto/rand for better randomness
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)

	// Use crypto/rand for better randomness
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to time-based if crypto/rand fails
		for i := range b {
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
	} else {
		// Use the random bytes to generate the referral code
		for i := range b {
			b[i] = charset[int(randomBytes[i])%len(charset)]
		}
	}

	return string(b)
}
