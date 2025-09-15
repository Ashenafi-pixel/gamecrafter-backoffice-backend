package groove

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module/groove"
	"github.com/tucanbit/internal/utils"
	"go.uber.org/zap"
)

// GrooveOfficialHandler implements the official GrooveTech Transaction API
// Based on documentation: https://groove-docs.pages.dev/transaction-api/
type GrooveOfficialHandler struct {
	grooveService      groove.GrooveService
	logger             *zap.Logger
	signatureValidator *utils.GrooveSignatureValidator
}

func NewGrooveOfficialHandler(grooveService groove.GrooveService, logger *zap.Logger) *GrooveOfficialHandler {
	// Initialize signature validator
	secretKey := viper.GetString("groove.signature_secret")
	if secretKey == "" {
		secretKey = "default_secret_key" // Fallback for development
	}

	return &GrooveOfficialHandler{
		grooveService:      grooveService,
		logger:             logger,
		signatureValidator: utils.NewGrooveSignatureValidator(secretKey),
	}
}

// GetAccount - Official GrooveTech Get Account API
// Endpoint: {casino_endpoint}?request=getaccount&[parameters]
// Based on: https://groove-docs.pages.dev/transaction-api/get-account/
func (h *GrooveOfficialHandler) GetAccount(c *gin.Context) {
	h.logger.Info("GrooveTech Official Get Account request")

	// Validate signature if signature validation is enabled
	signatureValidationEnabled := viper.GetBool("groove.signature_validation")
	h.logger.Info("Signature validation check", zap.Bool("enabled", signatureValidationEnabled))
	if signatureValidationEnabled {
		signature := c.GetHeader("X-Groove-Signature")
		if signature == "" {
			h.logger.Error("Missing X-Groove-Signature header")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			return
		}

		// Extract query parameters for signature validation
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Validate signature
		if !h.signatureValidator.ValidateGrooveSignature(signature, queryParams) {
			h.logger.Error("Invalid signature")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			return
		}
	}

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

	// Validate game session and get user information
	session, err := h.grooveService.ValidateGameSession(c.Request.Context(), gameSessionID)
	if err != nil {
		h.logger.Error("Failed to validate game session", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1000,
			"status": "Not logged on",
			"error":  "Player session is invalid or expired",
		})
		return
	}

	// Get account information for the user
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1000,
			"status": "Not logged on",
			"error":  "Player session is invalid or expired",
		})
		return
	}

	account, err := h.grooveService.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account by user ID", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":   1000,
			"status": "Not logged on",
			"error":  "Player session is invalid or expired",
		})
		return
	}

	// Get user's real balance from the balances table
	realBalance, err := h.grooveService.GetUserBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		realBalance = decimal.Zero // Default to zero if balance not found
	}

	// Get user profile information for city, country, etc.
	userProfile, err := h.grooveService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		// Use defaults if profile not found
		userProfile = &dto.GrooveUserProfile{
			City:     "Unknown",
			Country:  "US",
			Currency: "USD",
		}
	}

	h.logger.Info("Account retrieved successfully",
		zap.String("accountid", account.AccountID),
		zap.String("gamesessionid", gameSessionID),
		zap.String("real_balance", realBalance.String()),
		zap.String("city", userProfile.City),
		zap.String("country", userProfile.Country))

	// Return response in official GrooveTech format with real account data
	c.JSON(http.StatusOK, gin.H{
		"code":          200,
		"status":        "Success",
		"accountid":     account.AccountID,
		"city":          userProfile.City,     // Real city from user profile
		"country":       userProfile.Country,  // Real country from user profile
		"currency":      userProfile.Currency, // Real currency from user profile
		"gamesessionid": gameSessionID,
		"real_balance":  realBalance.InexactFloat64(), // Real balance from balances table
		"bonus_balance": 0.00,                         // Bonus balance (not implemented yet)
		"game_mode":     1,                            // 1 = Real money mode, 2 = Bonus mode
		"order":         "cash_money",                 // cash_money or bonus_money
		"apiversion":    apiVersion,
	})
}

// GetBalance - GET /balance
// Retrieves current player balance from the operator's wallet
func (h *GrooveOfficialHandler) GetBalance(c *gin.Context) {
	h.logger.Info("GrooveTech Get Balance request")

	// Validate signature if signature validation is enabled
	if viper.GetBool("groove.signature_validation") {
		signature := c.GetHeader("X-Groove-Signature")
		if signature == "" {
			h.logger.Error("Missing X-Groove-Signature header")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			return
		}

		// Extract query parameters for signature validation
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Validate signature
		if !h.signatureValidator.ValidateGrooveSignature(signature, queryParams) {
			h.logger.Error("Invalid signature")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"status":  "Invalid signature",
				"message": "invalid signature",
			})
			return
		}
	}

	// Extract parameters according to official GrooveTech API specification
	request := c.Query("request")
	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	device := c.Query("device")
	nogsGameID := c.Query("nogsgameid")
	apiVersion := c.Query("apiversion")

	// Validate required parameters according to official spec
	if request != "getbalance" {
		h.logger.Error("Invalid request parameter", zap.String("request", request))
		c.JSON(http.StatusBadRequest, dto.GrooveGetBalanceResponse{
			Code:       400,
			Status:     "Bad Request",
			Message:    "Invalid request parameter. Must be 'getbalance'",
			APIVersion: "1.2",
		})
		return
	}

	if accountID == "" || gameSessionID == "" || device == "" || nogsGameID == "" || apiVersion == "" {
		h.logger.Error("Missing required parameters",
			zap.String("accountid", accountID),
			zap.String("gamesessionid", gameSessionID),
			zap.String("device", device),
			zap.String("nogsgameid", nogsGameID),
			zap.String("apiversion", apiVersion))
		c.JSON(http.StatusBadRequest, dto.GrooveGetBalanceResponse{
			Code:       400,
			Status:     "Bad Request",
			Message:    "Missing required parameters: accountid, gamesessionid, device, nogsgameid, apiversion",
			APIVersion: "1.2",
		})
		return
	}

	// Validate game session and get user information
	session, err := h.grooveService.ValidateGameSession(c.Request.Context(), gameSessionID)
	if err != nil {
		h.logger.Error("Failed to validate game session", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveGetBalanceResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	// Get account information for the user
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveGetBalanceResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	_, err = h.grooveService.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account by user ID", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveGetBalanceResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	// Get user's real balance from the balances table
	realBalance, err := h.grooveService.GetUserBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		realBalance = decimal.Zero // Default to zero if balance not found
	}

	h.logger.Info("Balance retrieved successfully",
		zap.String("account_id", accountID),
		zap.String("gamesessionid", gameSessionID),
		zap.String("real_balance", realBalance.String()))

	// Return response in official GrooveTech format
	c.JSON(http.StatusOK, dto.GrooveGetBalanceResponse{
		Code:         200,
		Status:       "Success",
		Balance:      realBalance,
		BonusBalance: decimal.Zero, // Bonus balance (not implemented yet)
		RealBalance:  realBalance,
		GameMode:     1,            // 1 = Real money mode, 2 = Bonus mode
		Order:        "cash_money", // cash_money or bonus_money
		APIVersion:   apiVersion,
	})
}

// ProcessWager - GET /wager
// Deducts funds from player balance for placing a bet
func (h *GrooveOfficialHandler) ProcessWager(c *gin.Context) {
	h.logger.Info("GrooveTech Wager request")

	// Validate signature if signature validation is enabled
	if viper.GetBool("groove.signature_validation") {
		signature := c.GetHeader("X-Groove-Signature")
		if signature == "" {
			h.logger.Error("Missing X-Groove-Signature header")
			c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
				Code:       1001,
				Status:     "Invalid signature",
				Message:    "invalid signature",
				APIVersion: "1.2",
			})
			return
		}

		// Extract query parameters for signature validation
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		// Validate signature
		if !h.signatureValidator.ValidateGrooveSignature(signature, queryParams) {
			h.logger.Error("Invalid signature")
			c.JSON(http.StatusUnauthorized, dto.GrooveWagerResponse{
				Code:       1001,
				Status:     "Invalid signature",
				Message:    "invalid signature",
				APIVersion: "1.2",
			})
			return
		}
	}

	// Extract parameters according to official GrooveTech API specification
	request := c.Query("request")
	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	device := c.Query("device")
	gameID := c.Query("gameid")
	apiVersion := c.Query("apiversion")
	betAmountStr := c.Query("betamount")
	roundID := c.Query("roundid")
	transactionID := c.Query("transactionid")
	frbid := c.Query("frbid") // Optional Free Round Bonus ID

	// Validate required parameters according to official spec
	if request != "wager" {
		h.logger.Error("Invalid request parameter", zap.String("request", request))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
			Code:       400,
			Status:     "Bad Request",
			Message:    "Invalid request parameter. Must be 'wager'",
			APIVersion: "1.2",
		})
		return
	}

	if accountID == "" || gameSessionID == "" || device == "" || gameID == "" || apiVersion == "" || betAmountStr == "" || roundID == "" || transactionID == "" {
		h.logger.Error("Missing required parameters",
			zap.String("accountid", accountID),
			zap.String("gamesessionid", gameSessionID),
			zap.String("device", device),
			zap.String("gameid", gameID),
			zap.String("apiversion", apiVersion),
			zap.String("betamount", betAmountStr),
			zap.String("roundid", roundID),
			zap.String("transactionid", transactionID))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
			Code:       400,
			Status:     "Bad Request",
			Message:    "Missing required parameters: accountid, gamesessionid, device, gameid, apiversion, betamount, roundid, transactionid",
			APIVersion: "1.2",
		})
		return
	}

	// Parse bet amount
	betAmount, err := decimal.NewFromString(betAmountStr)
	if err != nil || betAmount.LessThanOrEqual(decimal.Zero) {
		h.logger.Error("Invalid bet amount", zap.String("betamount", betAmountStr))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
			Code:       110,
			Status:     "Operation not allowed",
			Message:    "Invalid bet amount",
			APIVersion: "1.2",
		})
		return
	}

	// Validate game session and get user information
	session, err := h.grooveService.ValidateGameSession(c.Request.Context(), gameSessionID)
	if err != nil {
		h.logger.Error("Failed to validate game session", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	// Get account information for the user
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		h.logger.Error("Failed to parse user ID", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	account, err := h.grooveService.GetAccountByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account by user ID", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1000,
			Status:     "Not logged on",
			Message:    "Player session is invalid or expired",
			APIVersion: apiVersion,
		})
		return
	}

	// Check if account ID matches
	if account.AccountID != accountID {
		h.logger.Error("Account ID mismatch",
			zap.String("provided_accountid", accountID),
			zap.String("session_accountid", account.AccountID))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       110,
			Status:     "Operation not allowed",
			Message:    "Account ID doesn't match session ID",
			APIVersion: apiVersion,
		})
		return
	}

	// Get user's current balance
	currentBalance, err := h.grooveService.GetUserBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1,
			Status:     "Technical error",
			Message:    "Failed to retrieve balance",
			APIVersion: apiVersion,
		})
		return
	}

	// Check if user has sufficient funds
	if currentBalance.LessThan(betAmount) {
		h.logger.Error("Insufficient funds",
			zap.String("current_balance", currentBalance.String()),
			zap.String("bet_amount", betAmount.String()))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1006,
			Status:     "Out of money",
			Message:    "Insufficient funds to place the wager",
			APIVersion: apiVersion,
		})
		return
	}

	// Check for duplicate transaction (idempotency)
	existingTransaction, err := h.grooveService.GetTransactionByID(c.Request.Context(), transactionID)
	if err == nil && existingTransaction != nil {
		// Duplicate transaction found - return original response
		h.logger.Info("Duplicate transaction found", zap.String("transaction_id", transactionID))

		// Get current balance for the response
		currentBalance, _ := h.grooveService.GetUserBalance(c.Request.Context(), userID)

		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:                 200,
			Status:               "Success - duplicate request",
			AccountTransactionID: existingTransaction.AccountTransactionID,
			Balance:              currentBalance,
			BonusMoneyBet:        decimal.Zero, // No bonus money in our system
			RealMoneyBet:         betAmount,
			BonusBalance:         decimal.Zero, // No bonus balance in our system
			RealBalance:          currentBalance,
			GameMode:             1, // Real money mode
			Order:                "cash_money",
			APIVersion:           apiVersion,
		})
		return
	}

	// Process the wager transaction
	response, err := h.grooveService.ProcessWagerTransaction(c.Request.Context(), dto.GrooveWagerRequest{
		AccountID:     accountID,
		GameSessionID: gameSessionID,
		TransactionID: transactionID,
		RoundID:       roundID,
		GameID:        gameID,
		BetAmount:     betAmount,
		Device:        device,
		FRBID:         frbid,
		UserID:        userID,
	})

	if err != nil {
		h.logger.Error("Failed to process wager transaction", zap.Error(err))
		c.JSON(http.StatusOK, dto.GrooveWagerResponse{
			Code:       1,
			Status:     "Technical error",
			Message:    "Failed to process wager",
			APIVersion: apiVersion,
		})
		return
	}

	h.logger.Info("Wager processed successfully",
		zap.String("transaction_id", transactionID),
		zap.String("account_id", accountID),
		zap.String("bet_amount", betAmount.String()),
		zap.String("new_balance", response.RealBalance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessResult - GET /groove-official-result
// Official GrooveTech Result API - Processes game round outcomes and adds winnings
func (h *GrooveOfficialHandler) ProcessResult(c *gin.Context) {
	h.logger.Info("GrooveTech Result request")

	// Extract query parameters
	request := c.Query("request")
	if request != "result" {
		h.logger.Error("Invalid request parameter", zap.String("request", request))
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Code:       400,
			Status:     "Bad Request",
			APIVersion: "1.2",
		})
		return
	}

	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	device := c.Query("device")
	gameID := c.Query("gameid")
	apiVersion := c.Query("apiversion")
	gameStatus := c.Query("gamestatus")
	resultStr := c.Query("result")
	roundID := c.Query("roundid")
	transactionID := c.Query("transactionid")
	frbid := c.Query("frbid") // Optional

	// Validate required parameters
	if accountID == "" || gameSessionID == "" || device == "" || gameID == "" ||
		apiVersion == "" || gameStatus == "" || resultStr == "" || roundID == "" || transactionID == "" {
		h.logger.Error("Missing required parameters")
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Code:       400,
			Status:     "Bad Request",
			APIVersion: apiVersion,
		})
		return
	}

	// Validate game status
	if gameStatus != "completed" && gameStatus != "pending" {
		h.logger.Error("Invalid game status", zap.String("gamestatus", gameStatus))
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: apiVersion,
		})
		return
	}

	// Parse result amount
	result, err := decimal.NewFromString(resultStr)
	if err != nil {
		h.logger.Error("Invalid result amount", zap.String("result", resultStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: apiVersion,
		})
		return
	}

	// Validate result amount (should not be negative)
	if result.LessThan(decimal.Zero) {
		h.logger.Error("Negative result amount", zap.String("result", result.String()))
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: apiVersion,
		})
		return
	}

	// Create request object
	req := dto.GrooveResultRequest{
		Request:       request,
		AccountID:     accountID,
		APIVersion:    apiVersion,
		Device:        device,
		GameID:        gameID,
		GameSessionID: gameSessionID,
		GameStatus:    gameStatus,
		Result:        result,
		RoundID:       roundID,
		TransactionID: transactionID,
		FRBID:         frbid,
	}

	// Process result transaction
	response, err := h.grooveService.ProcessResultTransaction(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process result transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: apiVersion,
		})
		return
	}

	h.logger.Info("Result processed successfully",
		zap.String("transaction_id", transactionID),
		zap.String("result", result.String()),
		zap.String("balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessWagerAndResult - GET /groove?request=wagerAndResult&[parameters]
// Official GrooveTech Wager And Result API
// Based on: https://groove-docs.pages.dev/transaction-api/wager-and-result/
func (h *GrooveOfficialHandler) ProcessWagerAndResult(c *gin.Context) {
	h.logger.Info("GrooveTech Official Wager And Result request")

	// Extract query parameters
	requestType := c.Query("request")
	if requestType != "wagerAndResult" {
		h.logger.Error("Invalid request type", zap.String("request", requestType))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1002,
			"status":  "Invalid request",
			"message": "request parameter must be 'wagerAndResult'",
		})
		return
	}

	// Extract required parameters
	accountID := c.Query("accountid")
	sessionID := c.Query("gamesessionid")
	device := c.Query("device")
	gameID := c.Query("gameid")
	apiVersion := c.Query("apiversion")
	betAmountStr := c.Query("betamount")
	resultStr := c.Query("result")
	roundID := c.Query("roundid")
	transactionID := c.Query("transactionid")
	gameStatus := c.Query("gamestatus")

	// Validate required parameters
	if accountID == "" || sessionID == "" || device == "" || gameID == "" ||
		apiVersion == "" || betAmountStr == "" || resultStr == "" ||
		roundID == "" || transactionID == "" {
		h.logger.Error("Missing required parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1003,
			"status":  "Missing parameters",
			"message": "Missing required parameters: accountid, gamesessionid, device, gameid, apiversion, betamount, result, roundid, transactionid",
		})
		return
	}

	// Parse decimal amounts
	betAmount, err := decimal.NewFromString(betAmountStr)
	if err != nil {
		h.logger.Error("Invalid bet amount", zap.String("betamount", betAmountStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1004,
			"status":  "Invalid amount",
			"message": "Invalid betamount format",
		})
		return
	}

	resultAmount, err := decimal.NewFromString(resultStr)
	if err != nil {
		h.logger.Error("Invalid result amount", zap.String("result", resultStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1004,
			"status":  "Invalid amount",
			"message": "Invalid result format",
		})
		return
	}

	// Create request DTO
	req := dto.GrooveWagerAndResultRequest{
		AccountID:     accountID,
		SessionID:     sessionID,
		Device:        device,
		GameID:        gameID,
		APIVersion:    apiVersion,
		BetAmount:     betAmount,
		WinAmount:     resultAmount,
		RoundID:       roundID,
		TransactionID: transactionID,
		GameStatus:    gameStatus,
	}

	// Process wager and result
	response, err := h.grooveService.ProcessWagerAndResult(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process wager and result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1005,
			"status":  "Processing failed",
			"message": "Failed to process wager and result",
		})
		return
	}

	h.logger.Info("Wager and result processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("account_id", response.AccountID),
		zap.String("new_balance", response.Balance.String()))

	// Return response using the DTO structure
	c.JSON(http.StatusOK, *response)
}

// ProcessRollback - GET /groove-official/rollback
// Official GrooveTech Rollback API
// Based on: https://groove-docs.pages.dev/transaction-api/rollback/
func (h *GrooveOfficialHandler) ProcessRollback(c *gin.Context) {
	h.logger.Info("GrooveTech Official Rollback request")

	// Extract query parameters
	requestType := c.Query("request")
	if requestType != "rollback" {
		h.logger.Error("Invalid request type", zap.String("request", requestType))
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	device := c.Query("device")
	gameID := c.Query("gameid")
	apiVersion := c.Query("apiversion")
	transactionID := c.Query("transactionid")
	rollbackAmountStr := c.Query("rollbackamount")
	roundID := c.Query("roundid")

	// Validate required parameters
	if accountID == "" || gameSessionID == "" || device == "" || gameID == "" || apiVersion == "" || transactionID == "" {
		h.logger.Error("Missing required parameters for rollback",
			zap.String("accountid", accountID),
			zap.String("gamesessionid", gameSessionID),
			zap.String("device", device),
			zap.String("gameid", gameID),
			zap.String("apiversion", apiVersion),
			zap.String("transactionid", transactionID))
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Parse rollback amount (optional)
	var rollbackAmount decimal.Decimal
	if rollbackAmountStr != "" {
		var err error
		rollbackAmount, err = decimal.NewFromString(rollbackAmountStr)
		if err != nil {
			h.logger.Error("Invalid rollback amount", zap.String("rollbackamount", rollbackAmountStr))
			c.JSON(http.StatusBadRequest, dto.GrooveRollbackResponseOfficial{
				Code:       110,
				Status:     "Operation not allowed",
				APIVersion: "1.2",
			})
			return
		}
	}

	// Create request DTO
	req := dto.GrooveRollbackRequestOfficial{
		AccountID:      accountID,
		GameSessionID:  gameSessionID,
		Device:         device,
		GameID:         gameID,
		APIVersion:     apiVersion,
		TransactionID:  transactionID,
		RollbackAmount: rollbackAmount,
		RoundID:        roundID,
		Request:        "rollback",
	}

	// Process rollback
	response, err := h.grooveService.ProcessRollback(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process rollback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveRollbackResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		})
		return
	}

	h.logger.Info("Rollback processed successfully",
		zap.String("transaction_id", response.AccountTransactionID),
		zap.String("new_balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessJackpot - GET /groove-official/jackpot
// Official GrooveTech Jackpot API
// Based on: https://groove-docs.pages.dev/transaction-api/jackpot/
func (h *GrooveOfficialHandler) ProcessJackpot(c *gin.Context) {
	h.logger.Info("GrooveTech Official Jackpot request")

	// Extract query parameters
	requestType := c.Query("request")
	if requestType != "jackpot" {
		h.logger.Error("Invalid request type", zap.String("request", requestType))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	accountID := c.Query("accountid")
	gameSessionID := c.Query("gamesessionid")
	gameID := c.Query("gameid")
	apiVersion := c.Query("apiversion")
	transactionID := c.Query("transactionid")
	amountStr := c.Query("amount")
	roundID := c.Query("roundid")
	gameStatus := c.Query("gamestatus")

	// Validate required parameters
	if accountID == "" || gameSessionID == "" || gameID == "" || apiVersion == "" || transactionID == "" || amountStr == "" || roundID == "" || gameStatus == "" {
		h.logger.Error("Missing required parameters for jackpot",
			zap.String("accountid", accountID),
			zap.String("gamesessionid", gameSessionID),
			zap.String("gameid", gameID),
			zap.String("apiversion", apiVersion),
			zap.String("transactionid", transactionID),
			zap.String("amount", amountStr),
			zap.String("roundid", roundID),
			zap.String("gamestatus", gameStatus))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Parse amount
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		h.logger.Error("Invalid jackpot amount", zap.String("amount", amountStr))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Validate amount is positive
	if amount.LessThanOrEqual(decimal.Zero) {
		h.logger.Error("Jackpot amount must be positive", zap.String("amount", amount.String()))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Validate game status
	if gameStatus != "completed" && gameStatus != "pending" {
		h.logger.Error("Invalid game status", zap.String("gamestatus", gameStatus))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponseOfficial{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Create request DTO
	req := dto.GrooveJackpotRequestOfficial{
		AccountID:     accountID,
		GameSessionID: gameSessionID,
		GameID:        gameID,
		APIVersion:    apiVersion,
		TransactionID: transactionID,
		Amount:        amount,
		RoundID:       roundID,
		GameStatus:    gameStatus,
		Request:       "jackpot",
	}

	// Process jackpot
	response, err := h.grooveService.ProcessJackpot(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process jackpot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveJackpotResponseOfficial{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		})
		return
	}

	h.logger.Info("Jackpot processed successfully",
		zap.String("transaction_id", response.WalletTx),
		zap.String("new_balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessRollbackOnResult - Official GrooveTech Rollback on Result API
// Endpoint: {casino_endpoint}?request=reversewin&[parameters]
func (h *GrooveOfficialHandler) ProcessRollbackOnResult(c *gin.Context) {
	h.logger.Info("GrooveTech Official Rollback on Result request")

	// Parse query parameters
	req := dto.GrooveRollbackOnResultRequest{
		AccountID:        c.Query("accountid"),
		Amount:           decimal.Zero,
		APIVersion:       c.Query("apiversion"),
		Device:           c.Query("device"),
		GameID:           c.Query("gameid"),
		GameSessionID:    c.Query("gamesessionid"),
		Request:          c.Query("request"),
		RoundID:          c.Query("roundid"),
		TransactionID:    c.Query("transactionid"),
		WinTransactionID: c.Query("wintransactionid"),
	}

	// Parse amount
	if amountStr := c.Query("amount"); amountStr != "" {
		if amount, err := decimal.NewFromString(amountStr); err == nil {
			req.Amount = amount
		}
	}

	// Validate required parameters
	if req.AccountID == "" || req.APIVersion == "" || req.Device == "" ||
		req.GameID == "" || req.GameSessionID == "" || req.RoundID == "" ||
		req.TransactionID == "" || req.Amount.IsZero() {
		h.logger.Error("Missing required parameters for rollback on result")
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackOnResultResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Process rollback on result
	response, err := h.grooveService.ProcessRollbackOnResult(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process rollback on result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveRollbackOnResultResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		})
		return
	}

	h.logger.Info("Rollback on result processed successfully",
		zap.String("transaction_id", response.AccountTransactionID),
		zap.String("amount", req.Amount.String()),
		zap.String("balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessRollbackOnRollback - Official GrooveTech Rollback on Rollback API
// Endpoint: {casino_endpoint}?request=rollbackrollback&[parameters]
func (h *GrooveOfficialHandler) ProcessRollbackOnRollback(c *gin.Context) {
	h.logger.Info("GrooveTech Official Rollback on Rollback request")

	// Parse query parameters
	req := dto.GrooveRollbackOnRollbackRequest{
		AccountID:      c.Query("accountid"),
		RollbackAmount: decimal.Zero,
		APIVersion:     c.Query("apiversion"),
		Device:         c.Query("device"),
		GameID:         c.Query("gameid"),
		GameSessionID:  c.Query("gamesessionid"),
		Request:        c.Query("request"),
		RoundID:        c.Query("roundid"),
		TransactionID:  c.Query("transactionid"),
	}

	// Parse rollback amount
	if amountStr := c.Query("rollbackAmount"); amountStr != "" {
		if amount, err := decimal.NewFromString(amountStr); err == nil {
			req.RollbackAmount = amount
		}
	}

	// Validate required parameters
	if req.AccountID == "" || req.APIVersion == "" || req.Device == "" ||
		req.GameID == "" || req.GameSessionID == "" || req.RoundID == "" ||
		req.TransactionID == "" || req.RollbackAmount.IsZero() {
		h.logger.Error("Missing required parameters for rollback on rollback")
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackOnRollbackResponse{
			Code:       110,
			Status:     "Operation not allowed",
			APIVersion: "1.2",
		})
		return
	}

	// Process rollback on rollback
	response, err := h.grooveService.ProcessRollbackOnRollback(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process rollback on rollback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveRollbackOnRollbackResponse{
			Code:       1,
			Status:     "Technical error",
			APIVersion: "1.2",
		})
		return
	}

	h.logger.Info("Rollback on rollback processed successfully",
		zap.String("transaction_id", response.AccountTransactionID),
		zap.String("amount", req.RollbackAmount.String()),
		zap.String("balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessWagerByBatch - Official GrooveTech Wager by Batch API (Sportsbook)
// Endpoint: POST {casino_endpoint}?request=wagerbybatch&[parameters]
func (h *GrooveOfficialHandler) ProcessWagerByBatch(c *gin.Context) {
	h.logger.Info("GrooveTech Official Wager by Batch request")

	// Parse request body
	var req dto.GrooveWagerByBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse wager by batch request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    110,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required parameters
	if req.AccountID == "" || req.GameID == "" || req.GameSessionID == "" ||
		req.Device == "" || len(req.Bets) == 0 {
		h.logger.Error("Missing required parameters for wager by batch")
		c.JSON(http.StatusBadRequest, dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    110,
			Message: "Missing required parameters",
		})
		return
	}

	// Process wager by batch
	response, err := h.grooveService.ProcessWagerByBatch(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process wager by batch", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveWagerByBatchResponse{
			Status:  "Failed",
			Code:    1,
			Message: "Technical error",
		})
		return
	}

	h.logger.Info("Wager by batch processed successfully",
		zap.Int("bet_count", len(response.Bets)),
		zap.String("balance", response.Balance.String()))

	c.JSON(http.StatusOK, *response)
}

// HealthCheck - GET /health
// Simple health check endpoint
func (h *GrooveOfficialHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "groove-tech-api",
		"timestamp": time.Now().Unix(),
	})
}
