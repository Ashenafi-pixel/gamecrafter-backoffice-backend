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

// ProcessWager - POST /wager
// Deducts funds from player balance for placing a bet
func (h *GrooveOfficialHandler) ProcessWager(c *gin.Context) {
	h.logger.Info("GrooveTech Wager request")

	var req dto.GrooveWagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid wager request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.TransactionID == "" || req.AccountID == "" || req.SessionID == "" || req.Amount.LessThanOrEqual(decimal.Zero) {
		h.logger.Error("Missing required fields in wager request")
		c.JSON(http.StatusBadRequest, dto.GrooveWagerResponse{
			Success:      false,
			ErrorCode:    "MISSING_FIELDS",
			ErrorMessage: "transactionId, accountId, sessionId, and amount are required",
		})
		return
	}

	// Process wager
	response, err := h.grooveService.ProcessWager(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process wager", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveWagerResponse{
			Success:      false,
			ErrorCode:    "WAGER_FAILED",
			ErrorMessage: "Failed to process wager",
		})
		return
	}

	h.logger.Info("Wager processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.NewBalance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessResult - POST /result
// Adds winnings to player balance after game round completion
func (h *GrooveOfficialHandler) ProcessResult(c *gin.Context) {
	h.logger.Info("GrooveTech Result request")

	var req dto.GrooveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid result request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.TransactionID == "" || req.AccountID == "" || req.SessionID == "" || req.Amount.LessThan(decimal.Zero) {
		h.logger.Error("Missing required fields in result request")
		c.JSON(http.StatusBadRequest, dto.GrooveResultResponse{
			Success:      false,
			ErrorCode:    "MISSING_FIELDS",
			ErrorMessage: "transactionId, accountId, sessionId, and amount are required",
		})
		return
	}

	// Process result
	response, err := h.grooveService.ProcessResult(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveResultResponse{
			Success:      false,
			ErrorCode:    "RESULT_FAILED",
			ErrorMessage: "Failed to process result",
		})
		return
	}

	h.logger.Info("Result processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.NewBalance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessWagerAndResult - POST /wager-and-result
// Combined operation for immediate bet and result processing
func (h *GrooveOfficialHandler) ProcessWagerAndResult(c *gin.Context) {
	h.logger.Info("GrooveTech Wager and Result request")

	var req dto.GrooveWagerAndResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid wager and result request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveWagerAndResultResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.TransactionID == "" || req.AccountID == "" || req.SessionID == "" ||
		req.WagerAmount.LessThanOrEqual(decimal.Zero) || req.WinAmount.LessThan(decimal.Zero) {
		h.logger.Error("Missing required fields in wager and result request")
		c.JSON(http.StatusBadRequest, dto.GrooveWagerAndResultResponse{
			Success:      false,
			ErrorCode:    "MISSING_FIELDS",
			ErrorMessage: "transactionId, accountId, sessionId, wagerAmount, and winAmount are required",
		})
		return
	}

	// Process wager and result
	response, err := h.grooveService.ProcessWagerAndResult(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process wager and result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveWagerAndResultResponse{
			Success:      false,
			ErrorCode:    "WAGER_RESULT_FAILED",
			ErrorMessage: "Failed to process wager and result",
		})
		return
	}

	h.logger.Info("Wager and result processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.NewBalance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessRollback - POST /rollback
// Reverses a previous wager transaction
func (h *GrooveOfficialHandler) ProcessRollback(c *gin.Context) {
	h.logger.Info("GrooveTech Rollback request")

	var req dto.GrooveRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid rollback request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.TransactionID == "" || req.AccountID == "" || req.SessionID == "" ||
		req.Amount.LessThanOrEqual(decimal.Zero) || req.OriginalTransactionID == "" {
		h.logger.Error("Missing required fields in rollback request")
		c.JSON(http.StatusBadRequest, dto.GrooveRollbackResponse{
			Success:      false,
			ErrorCode:    "MISSING_FIELDS",
			ErrorMessage: "transactionId, accountId, sessionId, amount, and originalTransactionId are required",
		})
		return
	}

	// Process rollback
	response, err := h.grooveService.ProcessRollback(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process rollback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveRollbackResponse{
			Success:      false,
			ErrorCode:    "ROLLBACK_FAILED",
			ErrorMessage: "Failed to process rollback",
		})
		return
	}

	h.logger.Info("Rollback processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.NewBalance.String()))

	c.JSON(http.StatusOK, *response)
}

// ProcessJackpot - POST /jackpot
// Processes special jackpot wins
func (h *GrooveOfficialHandler) ProcessJackpot(c *gin.Context) {
	h.logger.Info("GrooveTech Jackpot request")

	var req dto.GrooveJackpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid jackpot request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.TransactionID == "" || req.AccountID == "" || req.SessionID == "" || req.Amount.LessThanOrEqual(decimal.Zero) {
		h.logger.Error("Missing required fields in jackpot request")
		c.JSON(http.StatusBadRequest, dto.GrooveJackpotResponse{
			Success:      false,
			ErrorCode:    "MISSING_FIELDS",
			ErrorMessage: "transactionId, accountId, sessionId, and amount are required",
		})
		return
	}

	// Process jackpot
	response, err := h.grooveService.ProcessJackpot(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to process jackpot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.GrooveJackpotResponse{
			Success:      false,
			ErrorCode:    "JACKPOT_FAILED",
			ErrorMessage: "Failed to process jackpot",
		})
		return
	}

	h.logger.Info("Jackpot processed successfully",
		zap.String("transaction_id", response.TransactionID),
		zap.String("new_balance", response.NewBalance.String()))

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
