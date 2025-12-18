package kyc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	kyc handler.KYC,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {
	kycRoutes := []routing.Route{
		// Create KYC document (for testing - Postman only)
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/kyc/document/create",
			Handler: kyc.CreateKYCDocument,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create kyc documents", http.MethodPost),
				middleware.SystemLogs("Create KYC Document", &log, systemLogs),
			},
		},
		// Get user's KYC documents
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/documents/:user_id",
			Handler: kyc.GetKYCDocuments,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get kyc documents", http.MethodGet),
				middleware.SystemLogs("Get KYC Documents", &log, systemLogs),
			},
		},
		// Update document status
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/kyc/document/status",
			Handler: kyc.UpdateDocumentStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update kyc document status", http.MethodPut),
				middleware.SystemLogs("Update KYC Document Status", &log, systemLogs),
			},
		},
		// Update user KYC status
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/kyc/user/status",
			Handler: kyc.UpdateUserKYCStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update user kyc status", http.MethodPut),
				middleware.SystemLogs("Update User KYC Status", &log, systemLogs),
			},
		},
		// Get user KYC status
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/user/:user_id/status",
			Handler: kyc.GetUserKYCStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get user kyc status", http.MethodGet),
				middleware.SystemLogs("Get User KYC Status", &log, systemLogs),
			},
		},
		// Block user withdrawals
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/kyc/user/block-withdrawals",
			Handler: kyc.BlockUserWithdrawals,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "block user withdrawals", http.MethodPost),
				middleware.SystemLogs("Block User Withdrawals", &log, systemLogs),
			},
		},
		// Unblock user withdrawals
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/kyc/user/:user_id/unblock-withdrawals",
			Handler: kyc.UnblockUserWithdrawals,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "unblock user withdrawals", http.MethodPost),
				middleware.SystemLogs("Unblock User Withdrawals", &log, systemLogs),
			},
		},
		// Get KYC submissions
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/submissions/:user_id",
			Handler: kyc.GetKYCSubmissions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get kyc submissions", http.MethodGet),
				middleware.SystemLogs("Get KYC Submissions", &log, systemLogs),
			},
		},
		// Get status changes
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/status-changes/:user_id",
			Handler: kyc.GetStatusChanges,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get kyc status changes", http.MethodGet),
				middleware.SystemLogs("Get KYC Status Changes", &log, systemLogs),
			},
		},
		// Get withdrawal block status
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/user/:user_id/withdrawal-block",
			Handler: kyc.GetWithdrawalBlockStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get user withdrawal block status", http.MethodGet),
				middleware.SystemLogs("Get Withdrawal Block Status", &log, systemLogs),
			},
		},
		// List all submissions (queue)
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/submissions",
			Handler: kyc.GetAllSubmissions,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get kyc submissions", http.MethodGet),
				middleware.SystemLogs("List KYC Submissions", &log, systemLogs),
			},
		},
		// Get KYC settings
		{
			Method:  http.MethodGet,
			Path:    "/api/admin/kyc/settings",
			Handler: kyc.GetKYCSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "view kyc settings", http.MethodGet),
				middleware.SystemLogs("Get KYC Settings", &log, systemLogs),
			},
		},
		// Update KYC settings
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/kyc/settings",
			Handler: kyc.UpdateKYCSettings,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				// Align with seeded permissions list
				middleware.Authz(authModule, "edit kyc settings", http.MethodPut),
				middleware.SystemLogs("Update KYC Settings", &log, systemLogs),
			},
		},
	}

	routing.RegisterRoute(group, kycRoutes, log)
}
