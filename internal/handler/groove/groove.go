package groove

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/model/response"
	"github.com/tucanbit/internal/module/groove"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/internal/utils"
	"go.uber.org/zap"
)

type GrooveHandler struct {
	grooveService      groove.GrooveService
	userStorage        storage.User
	balanceStorage     storage.Balance
	logger             *zap.Logger
	signatureValidator *utils.GrooveSignatureValidator
}

func NewGrooveHandler(grooveService groove.GrooveService, userStorage storage.User, balanceStorage storage.Balance, logger *zap.Logger) *GrooveHandler {
	// Initialize signature validator
	secretKey := viper.GetString("groove.signature_secret")
	if secretKey == "" {
		secretKey = "default_secret_key" // Fallback for development
	}

	return &GrooveHandler{
		grooveService:      grooveService,
		userStorage:        userStorage,
		balanceStorage:     balanceStorage,
		logger:             logger,
		signatureValidator: utils.NewGrooveSignatureValidator(secretKey),
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

	h.logger.Info("Balance retrieved successfully", zap.String("balance", balance.Balance.String()))
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
	googleUserID, err := uuid.Parse(userID.String())
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
	googleUserID, err := uuid.Parse(userID.String())
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

	// Return the UUID directly since we're using google/uuid everywhere
	return claims.UserID, nil
}

// LaunchGame - POST /api/groove/launch-game
// Secure endpoint for launching games with user authentication
func (h *GrooveHandler) LaunchGame(c *gin.Context) {
	h.logger.Info("Processing game launch request")

	var req dto.LaunchGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// Extract user ID from JWT token
	userID, err := h.extractUserIDFromToken(c)
	if err != nil {
		h.logger.Error("Failed to extract user ID from token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid or missing authentication token",
		})
		return
	}

	// Convert to google/uuid for service layer
	googleUserID, err := uuid.Parse(userID.String())
	if err != nil {
		h.logger.Error("Failed to convert UUID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	// Set CMA compliance defaults if not provided
	if req.Country == "" {
		req.Country = "ET" // Default to Ethiopia
	}
	if req.Currency == "" {
		req.Currency = "USD" // Default to USD
	}
	if req.Language == "" {
		req.Language = "en_US" // Default to English
	}
	if req.IsTestAccount == nil {
		testAccount := false // Default to real account
		req.IsTestAccount = &testAccount
	}
	if req.RealityCheckElapsed == 0 {
		req.RealityCheckElapsed = 0 // Default to 0 minutes elapsed
	}
	if req.RealityCheckInterval == 0 {
		req.RealityCheckInterval = 60 // Default to 60 minutes interval
	}

	// Launch the game
	launchResponse, err := h.grooveService.LaunchGame(c.Request.Context(), googleUserID, req)
	if err != nil {
		h.logger.Error("Failed to launch game", zap.Error(err))
		c.JSON(http.StatusInternalServerError, launchResponse)
		return
	}

	h.logger.Info("Game launched successfully",
		zap.String("user_id", userID.String()),
		zap.String("game_id", req.GameID),
		zap.String("session_id", launchResponse.SessionID))

	c.JSON(http.StatusOK, launchResponse)
}

// ValidateGameSession - GET /api/groove/validate-session/:sessionId
// Validates a game session and returns session details
func (h *GrooveHandler) ValidateGameSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		h.logger.Error("Missing session ID parameter")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing session ID parameter",
		})
		return
	}

	h.logger.Info("Validating game session", zap.String("session_id", sessionID))

	session, err := h.grooveService.ValidateGameSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to validate game session", zap.Error(err))
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Game session not found or expired",
		})
		return
	}

	h.logger.Info("Game session validated successfully", zap.String("session_id", sessionID))
	c.JSON(http.StatusOK, session)
}

// GetAccountOfficial - Official GrooveTech Get Account API
func (h *GrooveHandler) GetAccountOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.GetAccount(c)
}

// GetBalanceOfficial - Official GrooveTech Get Balance API
func (h *GrooveHandler) GetBalanceOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.GetBalance(c)
}

// ProcessWagerOfficial - Official GrooveTech Wager API
func (h *GrooveHandler) ProcessWagerOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessWager(c)
}

// ProcessResultOfficial - Official GrooveTech Result API
func (h *GrooveHandler) ProcessResultOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessResult(c)
}

// ProcessWagerAndResultOfficial - Official GrooveTech Wager and Result API
func (h *GrooveHandler) ProcessWagerAndResultOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessWagerAndResult(c)
}

// ProcessRollbackOfficial - Official GrooveTech Rollback API
func (h *GrooveHandler) ProcessRollbackOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessRollback(c)
}

// ProcessJackpotOfficial - Official GrooveTech Jackpot API
func (h *GrooveHandler) ProcessJackpotOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessJackpot(c)
}

// ProcessRollbackOnResultOfficial - Official GrooveTech Rollback on Result API
func (h *GrooveHandler) ProcessRollbackOnResultOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessRollbackOnResult(c)
}

// ProcessRollbackOnRollbackOfficial - Official GrooveTech Rollback on Rollback API
func (h *GrooveHandler) ProcessRollbackOnRollbackOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessRollbackOnRollback(c)
}

// ProcessWagerByBatchOfficial - Official GrooveTech Wager by Batch API
func (h *GrooveHandler) ProcessWagerByBatchOfficial(c *gin.Context) {
	// Delegate to GrooveOfficialHandler
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)
	officialHandler.ProcessWagerByBatch(c)
}

// ProcessGrooveOfficialRequest - Unified GrooveTech Official API Handler
// Handles all GrooveTech operations via 'request' query parameter
// GET /groove-official?request=getaccount&accountid=...&gamesessionid=...&device=desktop&apiversion=1.2
// GET /groove-official?request=getbalance&accountid=...&gamesessionid=...&device=desktop&nogsgameid=82695&apiversion=1.2
// GET /groove-official?request=wager&accountid=...&gamesessionid=...&device=desktop&gameid=82695&apiversion=1.2&betamount=10.0&roundid=...&transactionid=...
// GET /groove-official?request=result&accountid=...&gamesessionid=...&device=desktop&gameid=82695&apiversion=1.2&result=15.0&roundid=...&transactionid=...
// GET /groove-official?request=wagerAndResult&accountid=...&gamesessionid=...&device=desktop&gameid=82695&apiversion=1.2&betamount=10.0&result=15.0&roundid=...&transactionid=...
// GET /groove-official?request=rollback&accountid=...&gamesessionid=...&device=desktop&gameid=82695&apiversion=1.2&rollbackamount=10.0&roundid=...&transactionid=...
// GET /groove-official?request=jackpot&accountid=...&gamesessionid=...&gameid=82695&apiversion=1.2&amount=25.0&roundid=...&transactionid=...
// GET /groove-official?request=reversewin&accountid=...&gamesessionid=...&device=desktop&gameid=82695&apiversion=1.2&amount=10.0&roundid=...&transactionid=...
// GET /groove-official?request=rollbackrollback&accountid=...&gamesessionid=...&device=desktop&gameid=82695&rollbackAmount=5.0&roundid=...&transactionid=...
// POST /groove-official?request=wagerbybatch&accountid=...&gamesessionid=...&device=desktop&apiversion=1.2
func (h *GrooveHandler) ProcessGrooveOfficialRequest(c *gin.Context) {
	h.logger.Info("Processing unified GrooveTech Official request")

	// Get the request type from query parameters
	requestType := c.Query("request")
	if requestType == "" {
		h.logger.Error("Missing 'request' parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameter: request",
			"message": "The 'request' parameter is required to specify the operation type",
		})
		return
	}

	h.logger.Info("GrooveTech request type", zap.String("request", requestType))

	// Delegate to GrooveOfficialHandler based on request type
	officialHandler := NewGrooveOfficialHandler(h.grooveService, h.logger)

	switch requestType {
	case "getaccount":
		h.logger.Info("Routing to GetAccountOfficial")
		officialHandler.GetAccount(c)
	case "getbalance":
		h.logger.Info("Routing to GetBalanceOfficial")
		officialHandler.GetBalance(c)
	case "wager":
		h.logger.Info("Routing to ProcessWagerOfficial")
		officialHandler.ProcessWager(c)
	case "result":
		h.logger.Info("Routing to ProcessResultOfficial")
		officialHandler.ProcessResult(c)
	case "wagerAndResult":
		h.logger.Info("Routing to ProcessWagerAndResultOfficial")
		officialHandler.ProcessWagerAndResult(c)
	case "rollback":
		h.logger.Info("Routing to ProcessRollbackOfficial")
		officialHandler.ProcessRollback(c)
	case "jackpot":
		h.logger.Info("Routing to ProcessJackpotOfficial")
		officialHandler.ProcessJackpot(c)
	case "reversewin":
		h.logger.Info("Routing to ProcessRollbackOnResultOfficial")
		officialHandler.ProcessRollbackOnResult(c)
	case "rollbackrollback":
		h.logger.Info("Routing to ProcessRollbackOnRollbackOfficial")
		officialHandler.ProcessRollbackOnRollback(c)
	case "wagerbybatch":
		h.logger.Info("Routing to ProcessWagerByBatchOfficial")
		officialHandler.ProcessWagerByBatch(c)
	default:
		h.logger.Error("Unknown request type", zap.String("request", requestType))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request type",
			"message": "The 'request' parameter must be one of: getaccount, getbalance, wager, result, wagerAndResult, rollback, jackpot, reversewin, rollbackrollback, wagerbybatch",
		})
		return
	}
}
