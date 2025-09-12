package groove

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/handler/groove"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

// Init initializes GrooveTech API routes
// Based on official documentation: https://groove-docs.pages.dev/transaction-api/
func Init(grp *gin.RouterGroup, log zap.Logger, handler *groove.GrooveHandler, authz module.Authz, enforcer *casbin.Enforcer, systemLogs module.SystemLogs) {
	log.Info("Initializing GrooveTech Transaction API routes")

	// Official GrooveTech Transaction API routes
	// These match the exact specification from GrooveTech documentation
	grooveGroup := grp.Group("/groove")
	{
		// Authentication & Information Requests
		grooveGroup.GET("/account", handler.GetAccount) // Get Account (Legacy)
		grooveGroup.GET("/balance", handler.GetBalance) // Get Balance (Legacy)

		// Financial Transactions - TODO: Implement official GrooveTech API methods
		// grooveGroup.POST("/wager", handler.ProcessWager)           // Wager
		// grooveGroup.POST("/result", handler.ProcessResult)          // Result
		// grooveGroup.POST("/wager-and-result", handler.ProcessWagerAndResult) // Wager and Result
		// grooveGroup.POST("/rollback", handler.ProcessRollback)      // Rollback
		// grooveGroup.POST("/jackpot", handler.ProcessJackpot)       // Jackpot

		// Health check
		grooveGroup.GET("/health", handler.HealthCheck)
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

	// Official GrooveTech Transaction API endpoint
	// This matches the exact specification from GrooveTech documentation
	// Endpoint: {casino_endpoint}?request=getaccount&[parameters]
	grp.GET("/groove-official", handler.GetAccountOfficial) // Official Get Account API

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
