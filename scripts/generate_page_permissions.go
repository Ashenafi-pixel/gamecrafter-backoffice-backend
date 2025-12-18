package main

import (
	"fmt"
)

// PagePermission represents a permission for a specific page
type PagePermission struct {
	Name          string
	Description   string
	PagePath      string
	RequiresValue bool
}

func main() {
	permissions := []PagePermission{
		// Dashboard
		{"view dashboard", "Access to view the main dashboard", "/dashboard", false},

		// Reports
		{"view analytics dashboard", "Access to view analytics dashboard", "/reports", false},
		{"view game report", "Access to view game reports", "/reports/game", false},
		{"view transaction report", "Access to view transaction/wallet reports", "/reports/transaction", false},
		{"view transaction details", "Access to view transaction details", "/transactions/details", false},
		{"view daily report", "Access to view daily reports", "/reports/daily", false},
		{"view big winners report", "Access to view big winners report", "/reports/big-winners", false},
		{"view player metrics report", "Access to view player metrics report", "/reports/player-metrics", false},
		{"view country report", "Access to view country report", "/reports/country", false},
		{"view game performance report", "Access to view game performance report", "/reports/game-performance", false},
		{"view provider performance report", "Access to view provider performance report", "/reports/provider-performance", false},
		{"export reports", "Permission to export reports to CSV/Excel", "/reports", false},

		// Player Management
		{"view players", "Access to view player list", "/players", false},
		{"view player details", "Access to view individual player details", "/players", false},
		{"edit player", "Permission to edit player information", "/players", false},
		{"suspend player", "Permission to suspend player accounts", "/players", false},
		{"unsuspend player", "Permission to unsuspend player accounts", "/players", false},
		{"block player", "Permission to block player accounts", "/players", false},
		{"unblock player", "Permission to unblock player accounts", "/players", false},
		{"reset player password", "Permission to reset player passwords", "/players", false},
		{"manual fund player", "Permission to manually fund player accounts", "/players", true},              // Requires value (amount limit)
		{"manual withdraw player", "Permission to manually withdraw from player accounts", "/players", true}, // Requires value (amount limit)

		// Welcome Bonus
		{"view welcome bonus", "Access to view welcome bonus settings", "/welcome-bonus", false},
		{"edit welcome bonus", "Permission to edit welcome bonus settings", "/welcome-bonus", false},
		{"view welcome bonus channels", "Access to view welcome bonus channel settings", "/welcome-bonus", false},
		{"create welcome bonus channel", "Permission to create welcome bonus channel rules", "/welcome-bonus", false},
		{"edit welcome bonus channel", "Permission to edit welcome bonus channel rules", "/welcome-bonus", false},
		{"delete welcome bonus channel", "Permission to delete welcome bonus channel rules", "/welcome-bonus", false},

		// Notifications
		{"view notifications", "Access to view player notifications", "/notifications", false},
		{"create notification", "Permission to create notifications", "/notifications", false},
		{"send notification", "Permission to send notifications to players", "/notifications", false},
		{"view campaigns", "Access to view notification campaigns", "/notifications", false},
		{"create campaign", "Permission to create notification campaigns", "/notifications", false},
		{"edit campaign", "Permission to edit notification campaigns", "/notifications", false},
		{"delete campaign", "Permission to delete notification campaigns", "/notifications", false},
		{"send campaign", "Permission to send notification campaigns", "/notifications", false},

		// KYC Management
		{"view kyc management", "Access to view KYC management page", "/kyc-management", false},
		{"approve kyc", "Permission to approve KYC requests", "/kyc-management", false},
		{"reject kyc", "Permission to reject KYC requests", "/kyc-management", false},
		{"view kyc risk", "Access to view KYC risk management", "/kyc-risk", false},
		{"update kyc risk settings", "Permission to update KYC risk settings", "/kyc-risk", false},

		// Rakeback/Cashback
		{"view cashback", "Access to view cashback/VIP levels", "/cashback", false},
		{"edit cashback", "Permission to edit cashback/VIP level settings", "/cashback", false},
		{"view rakéback override", "Access to view rakéback override (Happy Hour)", "/admin/rakeback-override", false},
		{"create rakéback override", "Permission to create rakéback override", "/admin/rakeback-override", false},
		{"edit rakéback override", "Permission to edit rakéback override", "/admin/rakeback-override", false},
		{"delete rakéback override", "Permission to delete rakéback override", "/admin/rakeback-override", false},
		{"view rakéback schedules", "Access to view scheduled rakéback overrides", "/admin/rakeback-schedules", false},
		{"create rakéback schedule", "Permission to create scheduled rakéback overrides", "/admin/rakeback-schedules", false},
		{"edit rakéback schedule", "Permission to edit scheduled rakéback overrides", "/admin/rakeback-schedules", false},
		{"delete rakéback schedule", "Permission to delete scheduled rakéback overrides", "/admin/rakeback-schedules", false},

		// Transactions
		{"view withdrawals", "Access to view withdrawal requests", "/transactions/withdrawals", false},
		{"approve withdrawal", "Permission to approve withdrawal requests", "/transactions/withdrawals", true}, // Requires value (amount limit)
		{"reject withdrawal", "Permission to reject withdrawal requests", "/transactions/withdrawals", false},
		{"view withdrawal dashboard", "Access to view withdrawal dashboard", "/transactions/withdrawals/dashboard", false},
		{"view withdrawal settings", "Access to view withdrawal settings", "/transactions/withdrawals/settings", false},
		{"edit withdrawal settings", "Permission to edit withdrawal settings", "/transactions/withdrawals/settings", false},
		{"view deposits", "Access to view deposit transactions", "/transactions/deposits", false},
		{"approve deposit", "Permission to approve deposit transactions", "/transactions/deposits", false},
		{"reject deposit", "Permission to reject deposit transactions", "/transactions/deposits", false},
		{"view manual funds", "Access to view manual fund management", "/transactions/manual-funds", false},
		{"create manual fund", "Permission to create manual fund transactions", "/transactions/manual-funds", true}, // Requires value (amount limit)
		{"view fund logs", "Access to view fund logs", "/transactions/manual-funds", false},

		// Wallet Management
		{"view wallet management", "Access to view wallet management", "/wallet/management", false},
		{"create wallet", "Permission to create wallets", "/wallet/management", false},
		{"edit wallet", "Permission to edit wallet settings", "/wallet/management", false},
		{"delete wallet", "Permission to delete wallets", "/wallet/management", false},

		// Game Management
		{"view game management", "Access to view game management", "/admin/game-management", false},
		{"add game", "Permission to add new games", "/admin/game-management", false},
		{"edit game", "Permission to edit game settings", "/admin/game-management", false},
		{"delete game", "Permission to delete games", "/admin/game-management", false},
		{"enable game", "Permission to enable games", "/admin/game-management", false},
		{"disable game", "Permission to disable games", "/admin/game-management", false},
		{"update game status", "Permission to update game status", "/admin/game-management", false},

		// Brand Management
		{"view brand management", "Access to view brand management", "/admin/brand-management", false},
		{"create brand", "Permission to create brands", "/admin/brand-management", false},
		{"edit brand", "Permission to edit brand settings", "/admin/brand-management", false},
		{"delete brand", "Permission to delete brands", "/admin/brand-management", false},

		// Falcon Liquidity
		{"view falcon liquidity", "Access to view Falcon Liquidity", "/admin/falcon-liquidity", false},
		{"edit falcon liquidity", "Permission to edit Falcon Liquidity settings", "/admin/falcon-liquidity", false},

		// Settings
		{"view settings", "Access to view site settings", "/settings", false},
		{"edit settings", "Permission to edit site settings", "/settings", false},
		{"view welcome bonus settings", "Access to view welcome bonus settings in settings page", "/settings", false},
		{"edit welcome bonus settings", "Permission to edit welcome bonus settings", "/settings", false},
		{"view ip filters", "Access to view IP filters in settings", "/settings", false},
		{"add ip filter", "Permission to add IP filters", "/settings", false},
		{"remove ip filter", "Permission to remove IP filters", "/settings", false},

		// Access Control
		{"view access control", "Access to view access control page", "/access-control", false},
		{"view roles", "Access to view roles", "/access-control", false},
		{"create role", "Permission to create roles", "/access-control", false},
		{"edit role", "Permission to edit roles", "/access-control", false},
		{"delete role", "Permission to delete roles", "/access-control", false},
		{"view permissions", "Access to view permissions list", "/access-control", false},
		{"assign role", "Permission to assign roles to users", "/access-control", false},
		{"revoke role", "Permission to revoke roles from users", "/access-control", false},
		{"view admin users", "Access to view admin users", "/access-control", false},
		{"create admin user", "Permission to create admin users", "/access-control", false},
		{"edit admin user", "Permission to edit admin users", "/access-control", false},
		{"delete admin user", "Permission to delete admin users", "/access-control", false},
		{"view kyc settings", "Access to view KYC settings in access control", "/access-control", false},
		{"edit kyc settings", "Permission to edit KYC settings", "/access-control", false},

		// Admin Activity Logs
		{"view activity logs", "Access to view admin activity logs", "/admin/activity-logs", false},
		{"export activity logs", "Permission to export activity logs", "/admin/activity-logs", false},

		// Alert Management
		{"view alerts", "Access to view alert management", "/admin/alerts", false},
		{"create alert", "Permission to create alerts", "/admin/alerts", false},
		{"edit alert", "Permission to edit alerts", "/admin/alerts", false},
		{"delete alert", "Permission to delete alerts", "/admin/alerts", false},
		{"view alert configurations", "Access to view alert configurations", "/admin/alerts", false},
		{"create alert configuration", "Permission to create alert configurations", "/admin/alerts", false},
		{"edit alert configuration", "Permission to edit alert configurations", "/admin/alerts", false},
		{"delete alert configuration", "Permission to delete alert configurations", "/admin/alerts", false},
		{"view email groups", "Access to view email groups", "/admin/alerts", false},
		{"create email group", "Permission to create email groups", "/admin/alerts", false},
		{"edit email group", "Permission to edit email groups", "/admin/alerts", false},
		{"delete email group", "Permission to delete email groups", "/admin/alerts", false},
	}

	// Generate SQL insert statements
	fmt.Println("-- Page Permissions for Back Office")
	fmt.Println("-- Generated automatically from page routes")
	fmt.Println()

	for _, perm := range permissions {
		requiresValue := "FALSE"
		if perm.RequiresValue {
			requiresValue = "TRUE"
		}

		fmt.Printf("INSERT INTO permissions (name, description, requires_value) VALUES ('%s', '%s', %s) ON CONFLICT (name) DO NOTHING;\n",
			perm.Name, perm.Description, requiresValue)
	}

	fmt.Println()
	fmt.Println("-- Total permissions:", len(permissions))
}
