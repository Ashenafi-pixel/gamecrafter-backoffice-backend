package groove

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/groove"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	grooveModule "github.com/tucanbit/internal/module/groove"
	"go.uber.org/zap"
)

// Init initializes GrooveTech API routes
// Based on official documentation: https://groove-docs.pages.dev/transaction-api/
func Init(grp *gin.RouterGroup, log *zap.Logger, handler *groove.GrooveHandler, grooveService grooveModule.GrooveService, authz module.Authz, enforcer *casbin.Enforcer, systemLogs module.SystemLogs) {
	log.Info("Initializing GrooveTech Transaction API routes")

	// Official GrooveTech Transaction API routes
	// These match the exact specification from GrooveTech documentation
	grooveGroup := grp.Group("/groove")
	{
		// Authentication & Information Requests
		grooveGroup.GET("/account", handler.GetAccount) // Get Account (Legacy)
		grooveGroup.GET("/balance", handler.GetBalance) // Get Balance (Legacy)

		// Health check
		grooveGroup.GET("/health", handler.HealthCheck)
	}

	// Official GrooveTech Transaction API routes with signature validation
	// These match the exact specification from GrooveTech documentation
	officialGroup := grp.Group("/groove-official")
	{
		// Get security key for signature validation
		securityKey := "test_key" // This should come from config in production

		// Authentication & Information Requests
		officialGroup.GET("", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.GetAccountOfficial)         // Get Account
		officialGroup.GET("/balance", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.GetBalanceOfficial) // Get Balance

		// Financial Transactions
		officialGroup.GET("/wager", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessWagerOfficial)                         // Wager
		officialGroup.GET("/result", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessResultOfficial)                       // Result
		officialGroup.GET("/wager-and-result", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessWagerAndResultOfficial)     // Wager and Result
		officialGroup.GET("/rollback", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessRollbackOfficial)                   // Rollback
		officialGroup.GET("/jackpot", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessJackpotOfficial)                     // Jackpot
		officialGroup.GET("/reversewin", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessRollbackOnResultOfficial)         // Rollback on Result
		officialGroup.GET("/rollbackrollback", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessRollbackOnRollbackOfficial) // Rollback on Rollback

		// Sportsbook - Wager by Batch (POST method)
		officialGroup.POST("/wagerbybatch", middleware.GrooveSignatureMiddlewareOptional(securityKey), handler.ProcessWagerByBatchOfficial) // Wager by Batch
	}

	// Legacy GrooveTech API routes (for backward compatibility)
	legacyGroup := grp.Group("/groove/legacy")
	{
		// Legacy transaction operations
		legacyGroup.POST("/debit", handler.DebitTransaction)
		legacyGroup.POST("/credit", handler.CreditTransaction)
		legacyGroup.POST("/bet", handler.BetTransaction)
		legacyGroup.POST("/win", handler.WinTransaction)

		// Transaction history
		legacyGroup.GET("/transactions", handler.GetTransactionHistory)

		// Game session management - TODO: Implement these methods
		// legacyGroup.POST("/session", handler.CreateGameSession)
		// legacyGroup.DELETE("/session/:sessionId", handler.EndGameSession)
	}

	// Add the new Wager And Result endpoint
	grp.GET("/groove", middleware.GrooveSignatureMiddlewareOptional("test_key"), handler.ProcessWagerAndResultOfficial) // Official Wager And Result API

	// Game Launch API routes (secure endpoints for frontend)
	gameGroup := grp.Group("/api/groove")
	{
		gameGroup.POST("/launch-game", handler.LaunchGame)                         // Launch Game
		gameGroup.GET("/validate-session/:sessionId", handler.ValidateGameSession) // Validate Game Session
	}

	// Admin GrooveTech API routes - TODO: Implement these methods
	// adminGrooveGroup := grp.Group("/admin/groove")
	// adminGrooveGroup.Use(func(c *gin.Context) {
	// 	// Add admin authentication middleware here
	// 	c.Next()
	// })
	// {
	// 	// Admin account management
	// 	adminGrooveGroup.GET("/accounts", handler.GetAllAccounts)
	// 	adminGrooveGroup.GET("/accounts/:accountId", handler.GetAccountByID)
	// 	adminGrooveGroup.PUT("/accounts/:accountId/status", handler.UpdateAccountStatus)
	//
	// 	// Admin transaction management
	// 	adminGrooveGroup.GET("/transactions", handler.GetAllTransactions)
	// 	adminGrooveGroup.GET("/transactions/:transactionId", handler.GetTransactionByID)
	//
	// 	// Admin session management
	// 	adminGrooveGroup.GET("/sessions", handler.GetAllSessions)
	// 	adminGrooveGroup.DELETE("/sessions/:sessionId", handler.ForceEndSession)
	//
	// 	// Admin statistics
	// 	adminGrooveGroup.GET("/stats", handler.GetGrooveStats)
	// 	adminGrooveGroup.GET("/stats/accounts", handler.GetAccountStats)
	// 	adminGrooveGroup.GET("/stats/transactions", handler.GetTransactionStats)
	// }

	log.Info("GrooveTech Transaction API routes initialized successfully")
}
