package groove

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	googleUUID "github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module/groove"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type GrooveHandler struct {
	grooveService  groove.GrooveService
	userStorage    storage.User
	balanceStorage storage.Balance
	logger         *zap.Logger
}

func NewGrooveHandler(grooveService groove.GrooveService, userStorage storage.User, balanceStorage storage.Balance, logger *zap.Logger) *GrooveHandler {
	return &GrooveHandler{
		grooveService:  grooveService,
		userStorage:    userStorage,
		balanceStorage: balanceStorage,
		logger:         logger,
	}
}

// GetAccount - GET /groove/account
// This is the main endpoint Collins needs for game launch
func (h *GrooveHandler) GetAccount(c *gin.Context) {
	h.logger.Info("Getting GrooveTech account information")

	// Extract session ID from header (Collins will use access token as sessionId)
	sessionID := c.GetHeader("Authorization")
	if sessionID == "" {
		h.logger.Error("Missing Authorization header")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing Authorization header",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(sessionID) > 7 && sessionID[:7] == "Bearer " {
		sessionID = sessionID[7:]
	}

	// Get account information
	account, err := h.grooveService.GetAccount(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get account information",
		})
		return
	}

	h.logger.Info("Account retrieved successfully", zap.String("account_id", account.AccountID))
	c.JSON(http.StatusOK, account)
}

// GetBalance - GET /groove/balance
func (h *GrooveHandler) GetBalance(c *gin.Context) {
	h.logger.Info("Getting GrooveTech account balance")

	// Extract session ID from header
	sessionID := c.GetHeader("Authorization")
	if sessionID == "" {
		h.logger.Error("Missing Authorization header")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing Authorization header",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(sessionID) > 7 && sessionID[:7] == "Bearer " {
		sessionID = sessionID[7:]
	}

	// Get balance
	balance, err := h.grooveService.GetBalance(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get balance",
		})
		return
	}

	h.logger.Info("Balance retrieved successfully", zap.String("account_id", balance.AccountID))
	c.JSON(http.StatusOK, balance)
}

// DebitTransaction - POST /groove/debit
func (h *GrooveHandler) DebitTransaction(c *gin.Context) {
	h.logger.Info("Processing GrooveTech debit transaction")

	var req dto.GrooveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Extract user ID from JWT token to get account ID
	userID, err := h.extractUserIDFromToken(c)
	if err != nil {
		h.logger.Error("Failed to extract user ID from token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid or missing authentication token",
		})
		return
	}

	// Get account ID for this user (convert to google/uuid)
	googleUserID, err := googleUUID.Parse(userID.String())
	if err != nil {
		h.logger.Error("Failed to convert UUID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}
	account, err := h.grooveService.GetAccountByUserID(c.Request.Context(), googleUserID)
	if err != nil {
		h.logger.Error("Failed to get account for user", zap.Error(err))
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Account not found for user",
		})
		return
	}

	// Set account ID and transaction type
	req.AccountID = account.AccountID
	req.Type = "debit"

	// Process debit transaction
	transaction, err := h.grooveService.ProcessDebit(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process debit transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process debit transaction",
		})
		return
	}

	h.logger.Info("Debit transaction processed successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("balance", transaction.Balance.String()))

	c.JSON(http.StatusOK, dto.GrooveTransactionResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		Balance:       transaction.Balance,
		Status:        transaction.Status,
	})
}

// CreditTransaction - POST /groove/credit
func (h *GrooveHandler) CreditTransaction(c *gin.Context) {
	h.logger.Info("Processing GrooveTech credit transaction")

	var req dto.GrooveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Extract user ID from JWT token to get account ID
	userID, err := h.extractUserIDFromToken(c)
	if err != nil {
		h.logger.Error("Failed to extract user ID from token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid or missing authentication token",
		})
		return
	}

	// Get account ID for this user (convert to google/uuid)
	googleUserID, err := googleUUID.Parse(userID.String())
	if err != nil {
		h.logger.Error("Failed to convert UUID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}
	account, err := h.grooveService.GetAccountByUserID(c.Request.Context(), googleUserID)
	if err != nil {
		h.logger.Error("Failed to get account for user", zap.Error(err))
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Account not found for user",
		})
		return
	}

	// Set account ID and transaction type
	req.AccountID = account.AccountID
	req.Type = "credit"

	// Process credit transaction
	transaction, err := h.grooveService.ProcessCredit(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process credit transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process credit transaction",
		})
		return
	}

	h.logger.Info("Credit transaction processed successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("balance", transaction.Balance.String()))

	c.JSON(http.StatusOK, dto.GrooveTransactionResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		Balance:       transaction.Balance,
		Status:        transaction.Status,
	})
}

// BetTransaction - POST /groove/bet
func (h *GrooveHandler) BetTransaction(c *gin.Context) {
	h.logger.Info("Processing GrooveTech bet transaction")

	var req dto.GrooveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Set transaction type to bet
	req.Type = "bet"

	// Process bet transaction
	transaction, err := h.grooveService.ProcessBet(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process bet transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process bet transaction",
		})
		return
	}

	h.logger.Info("Bet transaction processed successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("balance", transaction.Balance.String()))

	c.JSON(http.StatusOK, dto.GrooveTransactionResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		Balance:       transaction.Balance,
		Status:        transaction.Status,
	})
}

// WinTransaction - POST /groove/win
func (h *GrooveHandler) WinTransaction(c *gin.Context) {
	h.logger.Info("Processing GrooveTech win transaction")

	var req dto.GrooveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Set transaction type to win
	req.Type = "win"

	// Process win transaction
	transaction, err := h.grooveService.ProcessWin(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process win transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to process win transaction",
		})
		return
	}

	h.logger.Info("Win transaction processed successfully",
		zap.String("transaction_id", transaction.TransactionID),
		zap.String("balance", transaction.Balance.String()))

	c.JSON(http.StatusOK, dto.GrooveTransactionResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		Balance:       transaction.Balance,
		Status:        transaction.Status,
	})
}

// GetTransactionHistory - GET /groove/transactions
func (h *GrooveHandler) GetTransactionHistory(c *gin.Context) {
	h.logger.Info("Getting GrooveTech transaction history")

	// Extract session ID from header
	sessionID := c.GetHeader("Authorization")
	if sessionID == "" {
		h.logger.Error("Missing Authorization header")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing Authorization header",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(sessionID) > 7 && sessionID[:7] == "Bearer " {
		sessionID = sessionID[7:]
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	transactionType := c.Query("type")

	var fromDate, toDate time.Time
	if fromDateStr := c.Query("fromDate"); fromDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			fromDate = parsed
		}
	}
	if toDateStr := c.Query("toDate"); toDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			toDate = parsed
		}
	}

	// Get transaction history
	history, err := h.grooveService.GetTransactionHistory(c.Request.Context(), dto.GrooveTransactionHistoryRequest{
		SessionID: sessionID,
		FromDate:  fromDate,
		ToDate:    toDate,
		Page:      page,
		PageSize:  pageSize,
		Type:      transactionType,
	})
	if err != nil {
		h.logger.Error("Failed to get transaction history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get transaction history",
		})
		return
	}

	h.logger.Info("Transaction history retrieved successfully",
		zap.Int("count", len(history.Transactions)))

	c.JSON(http.StatusOK, history)
}

// GetTransaction - GET /groove/transactions/:id
func (h *GrooveHandler) GetTransaction(c *gin.Context) {
	h.logger.Info("Getting GrooveTech transaction details")

	transactionID := c.Param("id")
	if transactionID == "" {
		h.logger.Error("Missing transaction ID")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing transaction ID",
		})
		return
	}

	// TODO: Implement GetTransaction method in service interface
	h.logger.Error("GetTransaction method not implemented")
	c.JSON(http.StatusNotImplemented, response.ErrorResponse{
		Code:    http.StatusNotImplemented,
		Message: "GetTransaction method not implemented",
	})
}

// HealthCheck - GET /groove/health
func (h *GrooveHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "groove-tech-api",
		"time":    time.Now().UTC(),
	})
}

// GetAccountOfficial - Official GrooveTech Get Account API
// Endpoint: {casino_endpoint}?request=getaccount&[parameters]
// Based on: https://groove-docs.pages.dev/transaction-api/get-account/
func (h *GrooveHandler) GetAccountOfficial(c *gin.Context) {
	h.logger.Info("GrooveTech Official Get Account request")

	// Extract parameters according to official API specification
	request := c.Query("request")
	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	device := c.Query("device")
	apiVersion := c.Query("apiversion")

	// Validate required parameters according to official spec
	if request != "getaccount" {
		h.logger.Error("Invalid request parameter", zap.String("request", request))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "Invalid request parameter. Must be 'getaccount'",
		})
		return
	}

	if accountID == "" || gameSessionID == "" || device == "" || apiVersion == "" {
		h.logger.Error("Missing required parameters",
			zap.String("accountid", accountID),
			zap.String("gamesessionid", gameSessionID),
			zap.String("device", device),
			zap.String("apiversion", apiVersion))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "Missing required parameters: accountid, gamesessionid, device, apiversion",
		})
		return
	}

	// Validate device parameter
	if device != "desktop" && device != "mobile" {
		h.logger.Error("Invalid device parameter", zap.String("device", device))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  "Invalid device parameter. Must be 'desktop' or 'mobile'",
		})
		return
	}

	// Get account information using gameSessionID as session identifier
	_, err := h.grooveService.GetAccount(c.Request.Context(), gameSessionID)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1000,
			"status": "Not logged on",
			"error":  "Player session is invalid or expired",
		})
		return
	}

	// Extract user ID from JWT token to get real user data
	claims := &dto.Claim{}
	key := viper.GetString("app.jwt_secret")
	if key == "" {
		key = viper.GetString("auth.jwt_secret")
	}
	if key == "" {
		h.logger.Error("JWT secret not configured")
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"status": "Technical error",
			"error":  "Internal server error",
		})
		return
	}

	jwtKey := []byte(key)
	token, err := jwt.ParseWithClaims(gameSessionID, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		h.logger.Error("Invalid JWT token", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1000,
			"status": "Not logged on",
			"error":  "Player session is invalid or expired",
		})
		return
	}

	// Get real user data from database
	user, exists, err := h.userStorage.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil || !exists {
		h.logger.Error("Failed to get user data", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"status": "Technical error",
			"error":  "Internal server error",
		})
		return
	}

	// Get user balance from balances table
	userBalances, err := h.balanceStorage.GetBalancesByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"status": "Technical error",
			"error":  "Internal server error",
		})
		return
	}

	// Calculate total real balance (USD)
	var realBalance float64 = 0.0
	var bonusBalance float64 = 0.0
	var userCurrency string = "USD"

	// Use user's default currency if available
	if user.DefaultCurrency != "" {
		userCurrency = user.DefaultCurrency
	}

	// Find USD balance or default currency balance
	for _, balance := range userBalances {
		if balance.CurrencyCode == userCurrency {
			// Convert from cents to dollars
			realBalance = float64(balance.AmountCents) / 100.0
			break
		}
	}

	h.logger.Info("Account retrieved successfully",
		zap.String("accountid", accountID),
		zap.String("gamesessionid", gameSessionID),
		zap.String("user_id", claims.UserID.String()),
		zap.String("username", user.Username),
		zap.Float64("real_balance", realBalance),
		zap.String("currency", userCurrency))

	// Get user location data with defaults
	city := user.City
	if city == "" {
		city = "Unknown"
	}

	country := user.Country
	if country == "" {
		country = "US" // Default to US
	}

	// Return response in official GrooveTech format with real data
	c.JSON(http.StatusOK, gin.H{
		"code":          200,
		"status":        "Success",
		"accountid":     accountID,
		"city":          city,
		"country":       country,
		"currency":      userCurrency,
		"gamesessionid": gameSessionID,
		"real_balance":  realBalance,
		"bonus_balance": bonusBalance,
		"game_mode":     1,            // 1 = Real money mode, 2 = Bonus mode
		"order":         "cash_money", // cash_money or bonus_money
		"apiversion":    apiVersion,
	})
}

// extractUserIDFromToken extracts user ID from JWT token in Authorization header
func (h *GrooveHandler) extractUserIDFromToken(c *gin.Context) (uuid.UUID, error) {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return uuid.Nil, fmt.Errorf("missing Authorization header")
	}

	// Remove "Bearer " prefix if present
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		authHeader = authHeader[7:]
	}

	// Parse JWT token
	claims := &dto.Claim{}
	key := viper.GetString("app.jwt_secret")
	if key == "" {
		key = viper.GetString("auth.jwt_secret")
	}
	if key == "" {
		return uuid.Nil, fmt.Errorf("JWT secret not configured")
	}

	jwtKey := []byte(key)
	token, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Convert from google/uuid to gofrs/uuid
	return uuid.FromStringOrNil(claims.UserID.String()), nil
}
