package crypto_wallet

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/storage/groove"
	"github.com/tucanbit/internal/utils"
)

type WalletType string

const (
	WalletTypeMetaMask      WalletType = "metamask"
	WalletTypeTrustWallet   WalletType = "trust"
	WalletTypeCoinbase      WalletType = "coinbase"
	WalletTypePhantom       WalletType = "phantom"
	WalletTypeSolflare      WalletType = "solflare"
	WalletTypeLedger        WalletType = "ledger"
	WalletTypeWalletConnect WalletType = "walletconnect"
)

type ChainType string

const (
	ChainEthereum  ChainType = "ethereum"
	ChainPolygon   ChainType = "polygon"
	ChainBSC       ChainType = "bsc"
	ChainAvalanche ChainType = "avalanche"
	ChainArbitrum  ChainType = "arbitrum"
	ChainOptimism  ChainType = "optimism"
	ChainSolana    ChainType = "solana"
	ChainBitcoin   ChainType = "bitcoin"
)

type CasinoWalletService struct {
	storage       storage.CryptoWallet
	user          storage.User
	balance       storage.Balance
	grooveStorage groove.GrooveStorage
	jwtSecret     string
}

func NewCasinoWalletService(
	storage storage.CryptoWallet,
	user storage.User,
	balance storage.Balance,
	grooveStorage groove.GrooveStorage,
	jwtSecret string,
) *CasinoWalletService {
	return &CasinoWalletService{
		storage:       storage,
		user:          user,
		balance:       balance,
		grooveStorage: grooveStorage,
		jwtSecret:     jwtSecret,
	}
}

func (s *CasinoWalletService) DetectChainFromAddress(address string) ChainType {
	address = strings.TrimSpace(strings.ToLower(address))

	if strings.HasPrefix(address, "0x") && len(address) == 42 {
		return ChainEthereum
	}

	if len(address) >= 32 && len(address) <= 44 && !strings.HasPrefix(address, "0x") {
		if _, err := base58.Decode(address); err == nil {
			return ChainSolana
		}
	}

	if strings.HasPrefix(address, "1") || strings.HasPrefix(address, "3") || strings.HasPrefix(address, "bc1") {
		return ChainBitcoin
	}

	return ChainEthereum
}

func (s *CasinoWalletService) VerifyEthereumSignature(walletAddress, message, signature string) error {
	ok, err := utils.VerifyWalletSignature(walletAddress, message, signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}
	if !ok {
		return fmt.Errorf("signature verification failed")
	}

	fmt.Println("Signature verification successful!")
	return nil
}

func (s *CasinoWalletService) VerifySolanaSignature(walletAddress, message, signature string) error {
	pubKeyBytes, err := base58.Decode(walletAddress)
	if err != nil {
		return fmt.Errorf("invalid solana address: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid solana public key size")
	}

	signature = strings.TrimPrefix(signature, "0x")

	var sigBytes []byte
	if len(signature) == 128 {
		sigBytes, err = hex.DecodeString(signature)
		if err != nil {
			return fmt.Errorf("invalid hex signature: %w", err)
		}
	} else {
		sigBytes, err = base58.Decode(signature)
		if err != nil {
			return fmt.Errorf("invalid base58 signature: %w", err)
		}
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("invalid solana signature size")
	}

	messageBytes := []byte(message)
	if !ed25519.Verify(pubKeyBytes, messageBytes, sigBytes) {
		return fmt.Errorf("solana signature verification failed")
	}

	return nil
}

func (s *CasinoWalletService) CreateWalletChallenge(ctx context.Context, req *dto.WalletChallengeRequest) (*dto.WalletChallengeResponse, error) {
	nonce := uuid.New().String()
	challengeMessage := fmt.Sprintf("Welcome to TucanBit Casino! Sign this message to verify your wallet ownership and access your account. Nonce: %s", nonce)
	expiresAt := time.Now().Add(10 * time.Minute)

	challenge, err := s.storage.CreateWalletChallenge(ctx, req.WalletAddress, nonce, challengeMessage, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet challenge: %w", err)
	}

	return &dto.WalletChallengeResponse{
		Message:          "Challenge created successfully",
		ChallengeMessage: challengeMessage,
		Nonce:            challenge.ChallengeNonce,
		ExpiresAt:        challenge.ExpiresAt,
	}, nil
}

// LoginWithWallet handles user authentication via wallet signature.
func (s *CasinoWalletService) LoginWithWallet(ctx context.Context, req *dto.WalletLoginRequest) (*dto.WalletLoginResponse, error) {
	fmt.Printf("DEBUG: LoginWithWallet called with wallet: %s, nonce: %s\n", req.WalletAddress, req.Nonce)

	challenge, err := s.storage.GetWalletChallenge(ctx, req.WalletAddress, req.Nonce)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get challenge: %v\n", err)
		return nil, fmt.Errorf("invalid or expired challenge")
	}

	fmt.Printf("DEBUG: Found challenge: %+v\n", challenge)

	if time.Now().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge has expired")
	}

	if challenge.IsUsed {
		return nil, fmt.Errorf("challenge already used")
	}

	// --- UPDATED LOGIC ---
	// Prioritize the chain type from the request, falling back to detection for older clients.
	var chainType ChainType
	if req.ChainType != "" {
		chainType = ChainType(req.ChainType)
		fmt.Printf("DEBUG: Using provided chain type: %s\n", chainType)
	} else {
		chainType = s.DetectChainFromAddress(req.WalletAddress)
		fmt.Printf("DEBUG: No chain type provided, detected: %s\n", chainType)
	}
	// --- END OF UPDATE ---

	switch chainType {
	case ChainEthereum, ChainPolygon, ChainBSC, ChainAvalanche, ChainArbitrum, ChainOptimism:
		fmt.Printf("DEBUG: Verifying Ethereum signature for chain: %s\n", chainType)
		err = s.VerifyEthereumSignature(req.WalletAddress, req.Message, req.Signature)
	case ChainSolana:
		fmt.Printf("DEBUG: Verifying Solana signature for chain: %s\n", chainType)
		err = s.VerifySolanaSignature(req.WalletAddress, req.Message, req.Signature)
	default:
		return nil, fmt.Errorf("unsupported blockchain network: %s", chainType)
	}

	if err != nil {
		detailedError := fmt.Errorf("wallet signature verification failed: %w", err)
		fmt.Printf("DEBUG: %v\n", detailedError)
		s.LogWalletAuth(ctx, uuid.Nil, req.WalletAddress, "login", false, detailedError.Error(), map[string]interface{}{
			"wallet_type": req.WalletType,
			"chain_type":  chainType,
		})
		return nil, fmt.Errorf("invalid wallet signature. Please try signing the message again")
	}

	err = s.storage.MarkChallengeAsUsed(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark challenge as used: %w", err)
	}

	fmt.Printf("DEBUG: About to search for user with wallet address: %s\n", req.WalletAddress)
	dbUser, err := s.storage.GetUserByWalletAddress(ctx, req.WalletAddress, string(req.WalletType))
	isNewUser := false

	fmt.Printf("DEBUG: GetUserByWalletAddress result - Error: %v, User: %+v\n", err, dbUser)

	if err != nil {
		fmt.Printf("DEBUG: Error details - Type: %T, Message: %s\n", err, err.Error())

		// Check if it's a "no rows" error (user not found)
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "user not found") {
			fmt.Printf("DEBUG: User not found, creating new user for wallet: %s\n", req.WalletAddress)
			dbUser, err = s.createNewUserWithWallet(ctx, req, chainType)
			if err != nil {
				fmt.Printf("DEBUG:  Failed to create new user: %v\n", err)
				return nil, fmt.Errorf("failed to create new user account: %w", err)
			}
			isNewUser = true
			fmt.Printf("DEBUG: Successfully created new user with ID: %s\n", dbUser.ID.String())
		} else {
			fmt.Printf("DEBUG:  Database error while getting user: %v\n", err)
			return nil, fmt.Errorf("database error while retrieving user account: %w", err)
		}
	} else {
		fmt.Printf("DEBUG: User found with ID: %s\n", dbUser.ID.String())
	}

	connection, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err == nil {
		updates := map[string]interface{}{
			"is_verified":  true,
			"last_used_at": time.Now(),
		}
		s.storage.UpdateWalletConnection(ctx, connection.ID, updates)
	}

	tokens, err := s.generateTokens(dbUser.ID.String(), req.WalletAddress, chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.LogWalletAuth(ctx, dbUser.ID, req.WalletAddress, "login", true, "", map[string]interface{}{
		"wallet_type": req.WalletType,
		"chain_type":  chainType,
		"is_new_user": isNewUser,
	})

	return &dto.WalletLoginResponse{
		AccessToken:   tokens.AccessToken,
		RefreshToken:  tokens.RefreshToken,
		UserID:        dbUser.ID.String(),
		WalletAddress: req.WalletAddress,
		ChainType:     string(chainType),
		IsNewUser:     isNewUser,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		UserProfile: dto.UserProfile{
			PhoneNumber:    dbUser.PhoneNumber.String,
			Email:          dbUser.Email.String,
			UserID:         dbUser.ID,
			ProfilePicture: dbUser.Profile.String,
			FirstName:      dbUser.FirstName.String,
			LastName:       dbUser.LastName.String,
			Type:           dto.Type(dbUser.UserType.String),
			ReferralCode:   dbUser.ReferalCode.String,
		},
	}, nil
}

// VerifyWalletSignature handles the /verify endpoint logic.
func (s *CasinoWalletService) VerifyWalletSignature(ctx context.Context, req *dto.WalletVerificationRequest) (*dto.WalletVerificationResponse, error) {
	challenge, err := s.storage.GetWalletChallenge(ctx, req.WalletAddress, req.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired challenge")
	}

	if time.Now().After(challenge.ExpiresAt) {
		return nil, fmt.Errorf("challenge has expired")
	}

	if challenge.IsUsed {
		return nil, fmt.Errorf("challenge already used")
	}

	// --- UPDATED LOGIC ---
	// Prioritize the chain type from the request, falling back to detection for older clients.
	var chainType ChainType
	if req.ChainType != "" {
		chainType = ChainType(req.ChainType)
		fmt.Printf("DEBUG (Verify): Using provided chain type: %s\n", chainType)
	} else {
		chainType = s.DetectChainFromAddress(req.WalletAddress)
		fmt.Printf("DEBUG (Verify): No chain type provided, detected: %s\n", chainType)
	}
	// --- END OF UPDATE ---

	switch chainType {
	case ChainEthereum, ChainPolygon, ChainBSC, ChainAvalanche, ChainArbitrum, ChainOptimism:
		err = s.VerifyEthereumSignature(req.WalletAddress, req.Message, req.Signature)
	case ChainSolana:
		err = s.VerifySolanaSignature(req.WalletAddress, req.Message, req.Signature)
	default:
		return nil, fmt.Errorf("unsupported blockchain network: %s", chainType)
	}

	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	err = s.storage.MarkChallengeAsUsed(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark challenge as used: %w", err)
	}

	connection, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("wallet connection not found")
	}

	updates := map[string]interface{}{
		"is_verified": true,
	}
	err = s.storage.UpdateWalletConnection(ctx, connection.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet connection: %w", err)
	}

	return &dto.WalletVerificationResponse{
		Message:       "Wallet verified successfully",
		ConnectionID:  connection.ID,
		WalletAddress: connection.WalletAddress,
		ChainType:     string(chainType),
		IsVerified:    true,
		VerifiedAt:    time.Now(),
	}, nil
}

func (s *CasinoWalletService) ConnectAdditionalWallet(ctx context.Context, req *dto.WalletConnectionRequest, userID uuid.UUID) (*dto.WalletConnectionResponse, error) {
	_, exists, err := s.user.GetUserByID(ctx, userID)
	if err != nil || !exists {
		return nil, fmt.Errorf("user not found")
	}

	existing, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("wallet already connected to another account")
	}

	chainType := s.DetectChainFromAddress(req.WalletAddress)
	connection, err := s.storage.CreateWalletConnection(ctx, userID, req.WalletAddress, string(req.WalletType))
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet connection: %w", err)
	}

	userWalletCount, err := s.storage.CountUserWallets(ctx, userID)
	if err == nil && userWalletCount == 1 {
		s.storage.SetPrimaryWallet(ctx, userID, req.WalletAddress)
	}

	return &dto.WalletConnectionResponse{
		Message:       "Wallet connected successfully",
		ConnectionID:  connection.ID,
		WalletAddress: connection.WalletAddress,
		WalletType:    connection.WalletType,
		ChainType:     string(chainType),
		IsVerified:    connection.IsVerified,
		ConnectedAt:   connection.CreatedAt,
	}, nil
}

func (s *CasinoWalletService) GetUserWallets(ctx context.Context, userID uuid.UUID) (*dto.UserWalletsResponse, error) {
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
			IsPrimary:     false, // TODO: Implement primary wallet logic
			LastUsedAt:    wallet.LastUsedAt,
			ConnectedAt:   wallet.CreatedAt,
		})
	}

	return &dto.UserWalletsResponse{
		Wallets: walletInfos,
		Total:   len(walletInfos),
	}, nil
}

func (s *CasinoWalletService) DisconnectWallet(ctx context.Context, req *dto.WalletDisconnectRequest) (*dto.WalletDisconnectResponse, error) {
	err := s.storage.DeleteWalletConnection(ctx, req.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to disconnect wallet: %w", err)
	}

	return &dto.WalletDisconnectResponse{
		Message: "Wallet disconnected successfully",
	}, nil
}

// generateUniqueUsername generates a unique username for wallet users
func (s *CasinoWalletService) generateUniqueUsername(ctx context.Context) string {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// Generate a random username with timestamp and random suffix
		timestamp := time.Now().Unix() % 10000 // Last 4 digits of timestamp
		randomSuffix := uuid.New().String()[:4]
		username := fmt.Sprintf("player_%d%s", timestamp, randomSuffix)

		// Check if username exists
		exists, err := s.user.CheckUsernameExists(ctx, username)
		if err != nil {
			return fmt.Sprintf("player_%s", uuid.New().String()[:8])
		}

		if !exists {
			return username
		}
	}

	// Fallback to UUID-based username if all attempts fail
	return fmt.Sprintf("player_%s", uuid.New().String()[:8])
}

func (s *CasinoWalletService) createNewUserWithWallet(ctx context.Context, req *dto.WalletLoginRequest, chainType ChainType) (*db.User, error) {
	username := s.generateUniqueUsername(ctx)

	newUser := dto.User{
		Username:        username,
		PhoneNumber:     "",
		Password:        "",
		DefaultCurrency: "USD",
		Email:           "",
		Source:          "wallet",
		ReferralCode:    "",
		DateOfBirth:     "",
		CreatedBy:       uuid.Nil,
		IsAdmin:         false,
		FirstName:       "",
		LastName:        "",
		ReferalType:     dto.PLAYER,
		ReferedByCode:   "",
		Type:            dto.PLAYER,
		Status:          "active",
	}

	user, err := s.user.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Check if wallet connection already exists
	existingConnection, err := s.storage.GetWalletConnectionByAddress(ctx, req.WalletAddress)
	if err != nil {
		// If no connection exists, create a new one
		_, err = s.storage.CreateWalletConnection(ctx, user.ID, req.WalletAddress, string(req.WalletType))
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet connection: %w", err)
		}
	} else {
		// If connection exists, update it to link to the new user
		updates := map[string]interface{}{
			"user_id":      user.ID,
			"is_verified":  true,
			"last_used_at": time.Now(),
		}
		err = s.storage.UpdateWalletConnection(ctx, existingConnection.ID, updates)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing wallet connection: %w", err)
		}
	}

	err = s.storage.SetPrimaryWallet(ctx, user.ID, req.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to set primary wallet: %w", err)
	}

	// Create initial balance
	balance := dto.Balance{
		UserId:       user.ID,
		CurrencyCode: "USD",
		RealMoney:    decimal.NewFromInt(0),
		BonusMoney:   decimal.NewFromInt(0),
		Points:       0,
	}
	_, err = s.balance.CreateBalance(ctx, balance)
	if err != nil {
		return nil, fmt.Errorf("failed to create user balance: %w", err)
	}

	// Create GrooveTech account
	grooveAccount := dto.GrooveAccount{
		AccountID:    fmt.Sprintf("USD_%d_%s", 1000+int(time.Now().Unix()%9000), user.ID.String()),
		SessionID:    "",
		Balance:      decimal.NewFromInt(0),
		Currency:     "USD",
		Status:       "active",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	_, err = s.grooveStorage.CreateAccount(ctx, grooveAccount, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create GrooveTech account: %w", err)
	}

	dbUser := &db.User{
		ID:                       user.ID,
		Username:                 sql.NullString{String: user.Username, Valid: true},
		PhoneNumber:              sql.NullString{String: user.PhoneNumber, Valid: true},
		Password:                 user.Password,
		CreatedAt:                time.Now(),
		DefaultCurrency:          sql.NullString{String: user.DefaultCurrency, Valid: true},
		Profile:                  sql.NullString{String: "", Valid: false},
		Email:                    sql.NullString{String: user.Email, Valid: true},
		FirstName:                sql.NullString{String: user.FirstName, Valid: true},
		LastName:                 sql.NullString{String: user.LastName, Valid: true},
		DateOfBirth:              sql.NullString{String: user.DateOfBirth, Valid: false},
		Source:                   sql.NullString{String: user.Source, Valid: true},
		IsEmailVerified:          sql.NullBool{Bool: false, Valid: true},
		ReferalCode:              sql.NullString{String: user.ReferralCode, Valid: false},
		StreetAddress:            sql.NullString{String: user.StreetAddress, Valid: false},
		Country:                  sql.NullString{String: user.Country, Valid: false},
		State:                    sql.NullString{String: user.State, Valid: false},
		City:                     sql.NullString{String: user.City, Valid: false},
		PostalCode:               sql.NullString{String: user.PostalCode, Valid: false},
		KycStatus:                sql.NullString{String: user.KYCStatus, Valid: false},
		CreatedBy:                uuid.NullUUID{UUID: user.CreatedBy, Valid: true},
		IsAdmin:                  sql.NullBool{Bool: user.IsAdmin, Valid: true},
		Status:                   sql.NullString{String: user.Status, Valid: true},
		ReferalType:              sql.NullString{String: string(user.ReferalType), Valid: false},
		ReferedByCode:            sql.NullString{String: user.ReferedByCode, Valid: false},
		UserType:                 sql.NullString{String: string(user.Type), Valid: true},
		PrimaryWalletAddress:     sql.NullString{String: req.WalletAddress, Valid: true},
		WalletVerificationStatus: sql.NullString{String: "verified", Valid: true},
	}

	return dbUser, nil
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *CasinoWalletService) generateTokens(userID, walletAddress string, chainType ChainType) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id":        userID,
		"wallet_address": walletAddress,
		"chain_type":     chainType,
		"exp":            time.Now().Add(24 * time.Hour).Unix(),
		"iat":            time.Now().Unix(),
		"type":           "access",
		"iss":            "tucanbit-casino",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
		"iss":     "tucanbit-casino",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *CasinoWalletService) LogWalletAuth(ctx context.Context, userID uuid.UUID, walletAddress, action string, success bool, errorMessage string, metadata map[string]interface{}) error {
	status := "success"
	if !success {
		status = "failed"
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["timestamp"] = time.Now().Unix()
	metadata["success"] = success

	if errorMessage != "" {
		metadata["error"] = errorMessage
	}

	return s.storage.CreateWalletAuthLog(ctx, userID, walletAddress, action, status, metadata)
}

func (s *CasinoWalletService) CreateWalletConnection(ctx context.Context, userID uuid.UUID, walletAddress, walletType string) (*db.CryptoWalletConnection, error) {
	return s.storage.CreateWalletConnection(ctx, userID, walletAddress, walletType)
}

func (s *CasinoWalletService) CreateWalletConnectionWithRequest(ctx context.Context, req *dto.WalletConnectionRequest, userID uuid.UUID) (*dto.WalletConnectionResponse, error) {
	return s.ConnectAdditionalWallet(ctx, req, userID)
}
