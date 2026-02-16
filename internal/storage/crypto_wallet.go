package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/constant/model/db"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

// CryptoWallet defines the interface for crypto wallet storage operations
type CryptoWallet interface {
	// Wallet connection operations
	CreateWalletConnection(ctx context.Context, userID uuid.UUID, walletAddress, walletType string) (*db.CryptoWalletConnection, error)
	CreateWalletConnectionWithChain(ctx context.Context, userID uuid.UUID, walletAddress, walletType, walletChain string) (*db.CryptoWalletConnection, error)
	GetWalletConnectionByAddress(ctx context.Context, walletAddress string) (*db.CryptoWalletConnection, error)
	GetUserWalletConnections(ctx context.Context, userID uuid.UUID) ([]*db.CryptoWalletConnection, error)
	UpdateWalletConnection(ctx context.Context, connectionID uuid.UUID, updates map[string]interface{}) error
	DeleteWalletConnection(ctx context.Context, connectionID uuid.UUID) error
	CheckWalletExists(ctx context.Context, walletAddress string) (bool, error)
	SetPrimaryWallet(ctx context.Context, userID uuid.UUID, walletAddress string) error
	GetUserByWalletAddress(ctx context.Context, walletAddress, walletType string) (*db.User, error)

	// Challenge operations
	CreateWalletChallenge(ctx context.Context, walletAddress, nonce, challengeMessage string, expiresAt time.Time) (*db.CryptoWalletChallenge, error)
	GetWalletChallenge(ctx context.Context, walletAddress, nonce string) (*db.CryptoWalletChallenge, error)
	MarkChallengeAsUsed(ctx context.Context, challengeID uuid.UUID) error
	CleanExpiredChallenges(ctx context.Context) error

	// Auth logging
	CreateWalletAuthLog(ctx context.Context, userID uuid.UUID, walletAddress, action, status string, metadata map[string]interface{}) error
	GetWalletAuthLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*db.CryptoWalletAuthLog, error)

	// User wallet info
	GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*db.UserWalletInfo, error)
	CountUserWallets(ctx context.Context, userID uuid.UUID) (int, error)

	// Balance operations (to avoid sqlc issues)
	CreateBalanceRaw(ctx context.Context, userID uuid.UUID, currencyCode string, amount decimal.Decimal) error
}

// cryptoWalletStorage implements the CryptoWallet interface with production database
type cryptoWalletStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

// Init creates a new crypto wallet storage instance
func Init(db *persistencedb.PersistenceDB, log *zap.Logger) CryptoWallet {
	return &cryptoWalletStorage{
		db:  db,
		log: log,
	}
}

// CreateWalletConnection creates a new wallet connection
func (c *cryptoWalletStorage) CreateWalletConnection(ctx context.Context, userID uuid.UUID, walletAddress, walletType string) (*db.CryptoWalletConnection, error) {
	// Detect wallet chain based on address format
	walletChain := c.detectWalletChain(walletAddress)

	query := `
		INSERT INTO crypto_wallet_connections (
			user_id, wallet_type, wallet_address, wallet_chain, wallet_name, wallet_icon_url
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING *;
	`

	var connection db.CryptoWalletConnection
	err := c.db.GetPool().QueryRow(ctx, query,
		userID, walletType, walletAddress, walletChain, nil, nil,
	).Scan(
		&connection.ID, &connection.UserID, &connection.WalletType, &connection.WalletAddress,
		&connection.WalletChain, &connection.WalletName, &connection.WalletIconURL,
		&connection.IsVerified, &connection.VerificationSignature, &connection.VerificationMessage,
		&connection.VerificationTimestamp, &connection.LastUsedAt, &connection.CreatedAt, &connection.UpdatedAt,
	)

	if err != nil {
		c.log.Error("failed to create wallet connection",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("walletAddress", walletAddress),
			zap.String("walletType", walletType),
			zap.String("walletChain", walletChain))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to create wallet connection")
	}

	return &connection, nil
}

// CreateWalletConnectionWithChain creates a new wallet connection with specified chain type
func (c *cryptoWalletStorage) CreateWalletConnectionWithChain(ctx context.Context, userID uuid.UUID, walletAddress, walletType, walletChain string) (*db.CryptoWalletConnection, error) {
	query := `
		INSERT INTO crypto_wallet_connections (
			user_id, wallet_type, wallet_address, wallet_chain, wallet_name, wallet_icon_url
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING *;
	`

	var connection db.CryptoWalletConnection
	err := c.db.GetPool().QueryRow(ctx, query,
		userID, walletType, walletAddress, walletChain, nil, nil,
	).Scan(
		&connection.ID, &connection.UserID, &connection.WalletType, &connection.WalletAddress,
		&connection.WalletChain, &connection.WalletName, &connection.WalletIconURL,
		&connection.IsVerified, &connection.VerificationSignature, &connection.VerificationMessage,
		&connection.VerificationTimestamp, &connection.LastUsedAt, &connection.CreatedAt, &connection.UpdatedAt,
	)

	if err != nil {
		c.log.Error("failed to create wallet connection with chain",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("walletAddress", walletAddress),
			zap.String("walletType", walletType),
			zap.String("walletChain", walletChain))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to create wallet connection with chain")
	}

	return &connection, nil
}

// detectWalletChain detects the blockchain chain based on wallet address format
func (c *cryptoWalletStorage) detectWalletChain(address string) string {
	address = strings.TrimSpace(strings.ToLower(address))

	// Ethereum-style addresses (0x prefix, 42 chars)
	if strings.HasPrefix(address, "0x") && len(address) == 42 {
		return "ethereum"
	}

	// Solana-style addresses (base58, 32-44 chars, no 0x prefix)
	if len(address) >= 32 && len(address) <= 44 && !strings.HasPrefix(address, "0x") {
		return "solana"
	}

	// Bitcoin-style addresses
	if strings.HasPrefix(address, "1") || strings.HasPrefix(address, "3") || strings.HasPrefix(address, "bc1") {
		return "bitcoin"
	}

	// Default to ethereum for unknown formats
	return "ethereum"
}

// GetWalletConnectionByAddress retrieves wallet connection by address
func (c *cryptoWalletStorage) GetWalletConnectionByAddress(ctx context.Context, walletAddress string) (*db.CryptoWalletConnection, error) {
	query := `
		SELECT * FROM crypto_wallet_connections 
		WHERE wallet_address = $1 AND wallet_type = $2;
	`

	var connection db.CryptoWalletConnection
	err := c.db.GetPool().QueryRow(ctx, query, walletAddress, "ethereum").Scan(
		&connection.ID, &connection.UserID, &connection.WalletType, &connection.WalletAddress,
		&connection.WalletChain, &connection.WalletName, &connection.WalletIconURL,
		&connection.IsVerified, &connection.VerificationSignature, &connection.VerificationMessage,
		&connection.VerificationTimestamp, &connection.LastUsedAt, &connection.CreatedAt, &connection.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrResourceNotFound.Wrap(err, "wallet connection not found")
		}
		c.log.Error("failed to get wallet connection by address",
			zap.Error(err),
			zap.String("walletAddress", walletAddress))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get wallet connection")
	}

	return &connection, nil
}

// GetUserWalletConnections retrieves all wallet connections for a user
func (c *cryptoWalletStorage) GetUserWalletConnections(ctx context.Context, userID uuid.UUID) ([]*db.CryptoWalletConnection, error) {
	query := `
		SELECT * FROM crypto_wallet_connections 
		WHERE user_id = $1 
		ORDER BY last_used_at DESC;
	`

	rows, err := c.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		c.log.Error("failed to get user wallet connections",
			zap.Error(err),
			zap.String("userID", userID.String()))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get user wallet connections")
	}
	defer rows.Close()

	var connections []*db.CryptoWalletConnection
	for rows.Next() {
		var connection db.CryptoWalletConnection
		err := rows.Scan(
			&connection.ID, &connection.UserID, &connection.WalletType, &connection.WalletAddress,
			&connection.WalletChain, &connection.WalletName, &connection.WalletIconURL,
			&connection.IsVerified, &connection.VerificationSignature, &connection.VerificationMessage,
			&connection.VerificationTimestamp, &connection.LastUsedAt, &connection.CreatedAt, &connection.UpdatedAt,
		)
		if err != nil {
			c.log.Error("failed to scan wallet connection", zap.Error(err))
			continue
		}
		connections = append(connections, &connection)
	}

	return connections, nil
}

// UpdateWalletConnection updates wallet connection
func (c *cryptoWalletStorage) UpdateWalletConnection(ctx context.Context, connectionID uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic update query
	query := `
		UPDATE crypto_wallet_connections 
		SET 
			wallet_name = COALESCE($2, wallet_name),
			wallet_icon_url = COALESCE($3, wallet_icon_url),
			is_verified = COALESCE($4, is_verified),
			verification_signature = COALESCE($5, verification_signature),
			verification_message = COALESCE($6, verification_message),
			verification_timestamp = COALESCE($7, verification_timestamp),
			last_used_at = NOW(),
			updated_at = NOW()
		WHERE id = $1;
	`

	// Extract values from updates map
	var walletName, walletIconURL, verificationSignature, verificationMessage sql.NullString
	var isVerified sql.NullBool
	var verificationTimestamp sql.NullTime

	if name, ok := updates["wallet_name"].(string); ok {
		walletName = sql.NullString{String: name, Valid: true}
	}
	if iconURL, ok := updates["wallet_icon_url"].(string); ok {
		walletIconURL = sql.NullString{String: iconURL, Valid: true}
	}
	if verified, ok := updates["is_verified"].(bool); ok {
		isVerified = sql.NullBool{Bool: verified, Valid: true}
	}
	if signature, ok := updates["verification_signature"].(string); ok {
		verificationSignature = sql.NullString{String: signature, Valid: true}
	}
	if message, ok := updates["verification_message"].(string); ok {
		verificationMessage = sql.NullString{String: message, Valid: true}
	}
	if timestamp, ok := updates["verification_timestamp"].(time.Time); ok {
		verificationTimestamp = sql.NullTime{Time: timestamp, Valid: true}
	}

	result, err := c.db.GetPool().Exec(ctx, query, connectionID, walletName, walletIconURL, isVerified, verificationSignature, verificationMessage, verificationTimestamp)
	if err != nil {
		c.log.Error("failed to update wallet connection",
			zap.Error(err),
			zap.String("connectionID", connectionID.String()))
		return errors.ErrUnableTocreate.Wrap(err, "failed to update wallet connection")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrResourceNotFound.Wrap(fmt.Errorf("wallet connection not found"), "wallet connection not found")
	}

	return nil
}

// DeleteWalletConnection deletes wallet connection
func (c *cryptoWalletStorage) DeleteWalletConnection(ctx context.Context, connectionID uuid.UUID) error {
	query := `DELETE FROM crypto_wallet_connections WHERE id = $1;`

	result, err := c.db.GetPool().Exec(ctx, query, connectionID)
	if err != nil {
		c.log.Error("failed to delete wallet connection",
			zap.Error(err),
			zap.String("connectionID", connectionID.String()))
		return errors.ErrUnableTocreate.Wrap(err, "failed to delete wallet connection")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrResourceNotFound.Wrap(fmt.Errorf("wallet connection not found"), "wallet connection not found")
	}

	return nil
}

// CheckWalletExists checks if wallet exists
func (c *cryptoWalletStorage) CheckWalletExists(ctx context.Context, walletAddress string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM crypto_wallet_connections 
			WHERE wallet_address = $1 AND wallet_type = $2
		) as exists;
	`

	var exists bool
	err := c.db.GetPool().QueryRow(ctx, query, walletAddress, "ethereum").Scan(&exists)
	if err != nil {
		c.log.Error("failed to check wallet exists",
			zap.Error(err),
			zap.String("walletAddress", walletAddress))
		return false, errors.ErrUnableTocreate.Wrap(err, "failed to check wallet exists")
	}

	return exists, nil
}

// SetPrimaryWallet sets primary wallet
func (c *cryptoWalletStorage) SetPrimaryWallet(ctx context.Context, userID uuid.UUID, walletAddress string) error {
	query := `
		UPDATE users 
		SET primary_wallet_address = $2, wallet_verification_status = 'verified'
		WHERE id = $1;
	`

	result, err := c.db.GetPool().Exec(ctx, query, userID, walletAddress)
	if err != nil {
		c.log.Error("failed to set primary wallet",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("walletAddress", walletAddress))
		return errors.ErrUnableTocreate.Wrap(err, "failed to set primary wallet")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrResourceNotFound.Wrap(fmt.Errorf("user not found"), "user not found")
	}

	return nil
}

// GetUserByWalletAddress gets user by wallet address
func (c *cryptoWalletStorage) GetUserByWalletAddress(ctx context.Context, walletAddress, walletType string) (*db.User, error) {
	query := `
		SELECT u.id, u.username, u.phone_number, u.password, u.created_at, u.default_currency, u.profile, u.email, u.first_name, u.last_name, u.date_of_birth, u.source, u.is_email_verified, u.referal_code, u.street_address, u.country, u.state, u.city, u.postal_code, u.kyc_status, u.created_by, u.is_admin, u.status, u.referal_type, u.refered_by_code, u.user_type, u.primary_wallet_address, u.wallet_verification_status, u.is_test_account, u.two_factor_enabled, u.two_factor_setup_at
		FROM users u
		JOIN crypto_wallet_connections cwc ON u.id = cwc.user_id
		WHERE cwc.wallet_address = $1 AND cwc.wallet_type = $2;
	`

	var user db.User
	err := c.db.GetPool().QueryRow(ctx, query, walletAddress, walletType).Scan(
		&user.ID, &user.Username, &user.PhoneNumber, &user.Password, &user.CreatedAt,
		&user.DefaultCurrency, &user.Profile, &user.Email, &user.FirstName, &user.LastName,
		&user.DateOfBirth, &user.Source, &user.IsEmailVerified, &user.ReferalCode, &user.StreetAddress,
		&user.Country, &user.State, &user.City, &user.PostalCode, &user.KycStatus,
		&user.CreatedBy, &user.IsAdmin, &user.Status, &user.ReferalType, &user.ReferedByCode,
		&user.UserType, &user.PrimaryWalletAddress, &user.WalletVerificationStatus, &user.IsTestAccount,
		&user.TwoFactorEnabled, &user.TwoFactorSetupAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrResourceNotFound.Wrap(err, "user not found for wallet address")
		}
		c.log.Error("failed to get user by wallet address",
			zap.Error(err),
			zap.String("walletAddress", walletAddress))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get user by wallet address")
	}

	return &user, nil
}

// CreateWalletChallenge creates wallet challenge
func (c *cryptoWalletStorage) CreateWalletChallenge(ctx context.Context, walletAddress, nonce, challengeMessage string, expiresAt time.Time) (*db.CryptoWalletChallenge, error) {
	// Automatically detect wallet type based on address format
	walletType := "ethereum" // default
	if len(walletAddress) == 44 && walletAddress[:2] == "0x" {
		// Ethereum-style address (0x + 40 hex chars)
		walletType = "ethereum"
	} else if len(walletAddress) == 44 && walletAddress[:2] != "0x" {
		// Solana-style address (base58, 44 chars)
		walletType = "solana"
	} else if len(walletAddress) == 42 && walletAddress[:2] == "0x" {
		// Some other EVM chain
		walletType = "ethereum"
	}

	query := `
		INSERT INTO crypto_wallet_challenges (
			wallet_address, wallet_type, challenge_message, challenge_nonce, expires_at
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING *;
	`

	var challenge db.CryptoWalletChallenge
	err := c.db.GetPool().QueryRow(ctx, query,
		walletAddress, walletType, challengeMessage, nonce, expiresAt,
	).Scan(
		&challenge.ID, &challenge.WalletAddress, &challenge.WalletType, &challenge.ChallengeMessage,
		&challenge.ChallengeNonce, &challenge.ExpiresAt, &challenge.IsUsed, &challenge.CreatedAt,
	)

	if err != nil {
		c.log.Error("failed to create wallet challenge",
			zap.Error(err),
			zap.String("walletAddress", walletAddress),
			zap.String("nonce", nonce))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to create wallet challenge")
	}

	return &challenge, nil
}

// GetWalletChallenge gets wallet challenge
func (c *cryptoWalletStorage) GetWalletChallenge(ctx context.Context, walletAddress, nonce string) (*db.CryptoWalletChallenge, error) {
	query := `
		SELECT * FROM crypto_wallet_challenges 
		WHERE wallet_address = $1 AND challenge_nonce = $2 AND expires_at > NOW() AND is_used = false
		ORDER BY created_at DESC LIMIT 1;
	`

	var challenge db.CryptoWalletChallenge
	err := c.db.GetPool().QueryRow(ctx, query, walletAddress, nonce).Scan(
		&challenge.ID, &challenge.WalletAddress, &challenge.WalletType, &challenge.ChallengeMessage,
		&challenge.ChallengeNonce, &challenge.ExpiresAt, &challenge.IsUsed, &challenge.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrResourceNotFound.Wrap(err, "challenge not found or expired")
		}
		c.log.Error("failed to get wallet challenge",
			zap.Error(err),
			zap.String("walletAddress", walletAddress),
			zap.String("nonce", nonce))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get wallet challenge")
	}

	return &challenge, nil
}

// MarkChallengeAsUsed marks challenge as used
func (c *cryptoWalletStorage) MarkChallengeAsUsed(ctx context.Context, challengeID uuid.UUID) error {
	query := `
		UPDATE crypto_wallet_challenges 
		SET is_used = true 
		WHERE id = $1;
	`

	result, err := c.db.GetPool().Exec(ctx, query, challengeID)
	if err != nil {
		c.log.Error("failed to mark challenge as used",
			zap.Error(err),
			zap.String("challengeID", challengeID.String()))
		return errors.ErrUnableTocreate.Wrap(err, "failed to mark challenge as used")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.ErrResourceNotFound.Wrap(fmt.Errorf("challenge not found"), "challenge not found")
	}

	return nil
}

// CleanExpiredChallenges cleans expired challenges
func (c *cryptoWalletStorage) CleanExpiredChallenges(ctx context.Context) error {
	query := `DELETE FROM crypto_wallet_challenges WHERE expires_at < NOW();`

	_, err := c.db.GetPool().Exec(ctx, query)
	if err != nil {
		c.log.Error("failed to clean expired challenges", zap.Error(err))
		return errors.ErrUnableTocreate.Wrap(err, "failed to clean expired challenges")
	}

	return nil
}

// CreateWalletAuthLog creates wallet auth log
func (c *cryptoWalletStorage) CreateWalletAuthLog(ctx context.Context, userID uuid.UUID, walletAddress, action, status string, metadata map[string]interface{}) error {
	// Convert metadata to JSON string
	metadataJSON := "{}"
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			c.log.Error("failed to marshal metadata", zap.Error(err))
			metadataJSON = "{}"
		} else {
			metadataJSON = string(metadataBytes)
		}
	}

	query := `
		INSERT INTO crypto_wallet_auth_logs (
			wallet_address, wallet_type, action, ip_address, user_agent, success, error_message, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		);
	`

	_, err := c.db.GetPool().Exec(ctx, query,
		walletAddress, "ethereum", action, nil, nil, status == "success", nil, metadataJSON,
	)
	if err != nil {
		c.log.Error("failed to create wallet auth log",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("walletAddress", walletAddress),
			zap.String("action", action))
		return errors.ErrUnableTocreate.Wrap(err, "failed to create wallet auth log")
	}

	return nil
}

// GetWalletAuthLogs gets wallet auth logs
func (c *cryptoWalletStorage) GetWalletAuthLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*db.CryptoWalletAuthLog, error) {
	// First get user's wallet addresses
	connections, err := c.GetUserWalletConnections(ctx, userID)
	if err != nil {
		return nil, err
	}

	var allLogs []*db.CryptoWalletAuthLog
	for _, conn := range connections {
		query := `
			SELECT * FROM crypto_wallet_auth_logs 
			WHERE wallet_address = $1 
			ORDER BY created_at DESC 
			LIMIT $2 OFFSET $3;
		`

		rows, err := c.db.GetPool().Query(ctx, query, conn.WalletAddress, limit, offset)
		if err != nil {
			c.log.Error("failed to get wallet auth logs",
				zap.Error(err),
				zap.String("userID", userID.String()),
				zap.String("walletAddress", conn.WalletAddress))
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var log db.CryptoWalletAuthLog
			err := rows.Scan(
				&log.ID, &log.WalletAddress, &log.WalletType, &log.Action, &log.IPAddress,
				&log.UserAgent, &log.Success, &log.ErrorMessage, &log.Metadata, &log.CreatedAt,
			)
			if err != nil {
				c.log.Error("failed to scan wallet auth log", zap.Error(err))
				continue
			}
			allLogs = append(allLogs, &log)
		}
	}

	return allLogs, nil
}

// GetUserWallets gets user wallets
func (c *cryptoWalletStorage) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*db.UserWalletInfo, error) {
	query := `
		SELECT 
			cwc.id as connection_id,
			cwc.wallet_type,
			cwc.wallet_address,
			cwc.wallet_chain,
			cwc.wallet_name,
			cwc.is_verified,
			CASE WHEN u.primary_wallet_address = cwc.wallet_address THEN true ELSE false END as is_primary,
			cwc.last_used_at,
			cwc.created_at as connected_at
		FROM crypto_wallet_connections cwc
		WHERE cwc.user_id = $1
		ORDER BY cwc.last_used_at DESC;
	`

	rows, err := c.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		c.log.Error("failed to get user wallets",
			zap.Error(err),
			zap.String("userID", userID.String()))
		return nil, errors.ErrUnableTocreate.Wrap(err, "failed to get user wallets")
	}
	defer rows.Close()

	var wallets []*db.UserWalletInfo
	for rows.Next() {
		var wallet db.UserWalletInfo
		err := rows.Scan(
			&wallet.ConnectionID, &wallet.WalletType, &wallet.WalletAddress, &wallet.WalletChain,
			&wallet.WalletName, &wallet.IsVerified, &wallet.IsPrimary, &wallet.LastUsedAt, &wallet.ConnectedAt,
		)
		if err != nil {
			c.log.Error("failed to scan user wallet", zap.Error(err))
			continue
		}
		wallets = append(wallets, &wallet)
	}

	return wallets, nil
}

// CountUserWallets counts user wallets
func (c *cryptoWalletStorage) CountUserWallets(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM crypto_wallet_connections WHERE user_id = $1;`

	var count int
	err := c.db.GetPool().QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		c.log.Error("failed to count user wallets",
			zap.Error(err),
			zap.String("userID", userID.String()))
		return 0, errors.ErrUnableTocreate.Wrap(err, "failed to count user wallets")
	}

	return count, nil
}

// CreateBalanceRaw creates a balance using raw SQL to avoid sqlc issues
func (c *cryptoWalletStorage) CreateBalanceRaw(ctx context.Context, userID uuid.UUID, currencyCode string, amount decimal.Decimal) error {
	// Convert decimal amount to cents (multiply by 100)
	amountCents := amount.Mul(decimal.NewFromInt(100)).IntPart()

	// Use raw SQL query to match current database structure
	query := `
		INSERT INTO balances(user_id, currency_code, amount_cents, amount_units, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	var balanceID uuid.UUID
	err := c.db.GetPool().QueryRow(ctx, query,
		userID,
		currencyCode,
		amountCents,
		amount,
		time.Now(),
	).Scan(&balanceID)

	if err != nil {
		c.log.Error("failed to create balance with raw SQL",
			zap.Error(err),
			zap.String("userID", userID.String()),
			zap.String("currencyCode", currencyCode),
			zap.String("amount", amount.String()))
		return errors.ErrUnableTocreate.Wrap(err, "unable to create balance")
	}

	c.log.Info("Balance created successfully with raw SQL",
		zap.String("balanceID", balanceID.String()),
		zap.String("userID", userID.String()),
		zap.String("currencyCode", currencyCode),
		zap.String("amount", amount.String()))

	return nil
}
