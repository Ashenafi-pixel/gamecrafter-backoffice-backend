package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

func main() {
	// Database connection string from config
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@localhost:5433/tucanbit?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// All permissions used in the backend
	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		// Authz/RBAC permissions
		{"get permissions", "Get all permissions", "rbac", "read"},
		{"create role", "Create a new role", "rbac", "create"},
		{"get roles", "Get all roles", "rbac", "read"},
		{"update role permissions", "Update role permissions", "rbac", "update"},
		{"remove role", "Remove a role", "rbac", "delete"},
		{"assign role", "Assign role to user", "rbac", "create"},
		{"revoke role", "Revoke role from user", "rbac", "delete"},
		{"get role users", "Get users with a role", "rbac", "read"},
		{"get user roles", "Get roles for a user", "rbac", "read"},

		// Cashback permissions
		{"cashback", "Cashback management", "cashback", "admin"},
		{"process level progression", "Process single user level progression", "cashback", "update"},
		{"process bulk level progression", "Process bulk level progression", "cashback", "update"},
		{"get level progression info", "Get level progression information", "cashback", "read"},

		// Rakeback Override permissions
		{"get rakeback override", "Get rakeback override configuration", "rakeback", "read"},
		{"get active rakeback override", "Get active rakeback override", "rakeback", "read"},
		{"create or update rakeback override", "Create or update rakeback override", "rakeback", "write"},
		{"toggle rakeback override", "Toggle rakeback override status", "rakeback", "update"},

		// Brand permissions
		{"create brand", "Create a new brand", "brand", "create"},
		{"get brands", "Get all brands", "brand", "read"},
		{"get brand", "Get a specific brand", "brand", "read"},
		{"update brand", "Update a brand", "brand", "update"},
		{"delete brand", "Delete a brand", "brand", "delete"},

		// User/Player permissions
		{"block user account", "Block a user account", "user", "update"},
		{"get blocked account", "Get blocked account information", "user", "read"},
		{"add ip filter", "Add IP filter rule", "user", "create"},
		{"get ip filters", "Get IP filter rules", "user", "read"},
		{"update user profile", "Update user profile", "user", "update"},
		{"get users", "Get users list", "user", "read"},
		{"reset user account password", "Reset user account password", "user", "update"},
		{"get player details", "Get player details", "user", "read"},
		{"get player manual funds", "Get player manual funds", "user", "read"},
		{"remove ip filter", "Remove IP filter", "user", "delete"},
		{"get referral multiplier", "Get referral multiplier", "user", "read"},
		{"update referral multiplier", "Update referral multiplier", "user", "update"},
		{"add point to users", "Add points to users", "user", "update"},
		{"get point to users", "Get points for users", "user", "read"},
		{"register players", "Register new players", "user", "create"},
		{"get admins", "Get admin users", "user", "read"},
		{"get admin users", "Get admin users list", "user", "read"},
		{"create admin user", "Create admin user", "user", "create"},
		{"update admin user", "Update admin user", "user", "update"},
		{"delete admin user", "Delete admin user", "user", "delete"},
		{"update signup bonus", "Update signup bonus", "user", "update"},
		{"get signup bonus", "Get signup bonus", "user", "read"},
		{"update referral bonus", "Update referral bonus", "user", "update"},
		{"get referral bonus", "Get referral bonus", "user", "read"},

		// Admin Activity Logs permissions
		{"Create Admin Activity Log", "Create admin activity log", "logs", "create"},
		{"Get Admin Activity Logs", "Get admin activity logs", "logs", "read"},
		{"Get Admin Activity Log By ID", "Get admin activity log by ID", "logs", "read"},
		{"Get Admin Activity Stats", "Get admin activity statistics", "logs", "read"},
		{"Get Admin Activity Categories", "Get admin activity categories", "logs", "read"},
		{"Get Admin Activity Actions", "Get admin activity actions", "logs", "read"},
		{"Get Admin Activity Actions By Category", "Get admin activity actions by category", "logs", "read"},
		{"Delete Admin Activity Log", "Delete admin activity log", "logs", "delete"},
		{"Delete Admin Activity Logs By Admin", "Delete admin activity logs by admin", "logs", "delete"},
		{"Delete Old Admin Activity Logs", "Delete old admin activity logs", "logs", "delete"},

		// KYC permissions
		{"create kyc documents", "Create KYC documents", "kyc", "create"},
		{"get kyc documents", "Get KYC documents", "kyc", "read"},
		{"update kyc document status", "Update KYC document status", "kyc", "update"},
		{"update user kyc status", "Update user KYC status", "kyc", "update"},
		{"get user kyc status", "Get user KYC status", "kyc", "read"},
		{"block user withdrawals", "Block user withdrawals", "kyc", "update"},
		{"unblock user withdrawals", "Unblock user withdrawals", "kyc", "update"},
		{"get kyc submissions", "Get KYC submissions", "kyc", "read"},
		{"get kyc status changes", "Get KYC status changes", "kyc", "read"},
		{"get user withdrawal block status", "Get user withdrawal block status", "kyc", "read"},
		{"get kyc settings", "Get KYC settings", "kyc", "read"},
		{"update kyc settings", "Update KYC settings", "kyc", "update"},

		// Game permissions
		{"create game", "Create a new game", "game", "create"},
		{"get games", "Get all games", "game", "read"},
		{"get game stats", "Get game statistics", "game", "read"},
		{"get game by id", "Get game by ID", "game", "read"},
		{"update game", "Update a game", "game", "update"},
		{"delete game", "Delete a game", "game", "delete"},
		{"bulk update game status", "Bulk update game status", "game", "update"},
		{"create house edge", "Create house edge configuration", "game", "create"},
		{"get house edges", "Get house edges", "game", "read"},
		{"get house edge stats", "Get house edge statistics", "game", "read"},
		{"get house edge by id", "Get house edge by ID", "game", "read"},
		{"get house edges by game type", "Get house edges by game type", "game", "read"},
		{"get house edges by game variant", "Get house edges by game variant", "game", "read"},
		{"update house edge", "Update house edge", "game", "update"},
		{"delete house edge", "Delete house edge", "game", "delete"},
		{"bulk update house edge status", "Bulk update house edge status", "game", "update"},

		// Balance/Funding permissions
		{"manual funding", "Manual funding operations", "balance", "create"},
		{"get fund logs", "Get fund logs", "balance", "read"},
		{"get all manual funds", "Get all manual funds", "balance", "read"},

		// Bet permissions
		{"get game history", "Get game history", "bet", "read"},
		{"get failed rounds", "Get failed rounds", "bet", "read"},
		{"manual refund failed rounds", "Manual refund failed rounds", "bet", "update"},
		{"create league", "Create league", "bet", "create"},
		{"get leagues", "Get leagues", "bet", "read"},
		{"create clubs", "Create clubs", "bet", "create"},
		{"get clubs", "Get clubs", "bet", "read"},
		{"update football match multiplier", "Update football match multiplier", "bet", "update"},
		{"get football match multiplier", "Get football match multiplier", "bet", "read"},
		{"create football match round", "Create football match round", "bet", "create"},
		{"get football match round", "Get football match round", "bet", "read"},
		{"create football matches", "Create football matches", "bet", "create"},
		{"get football match round matches", "Get football match round matches", "bet", "read"},
		{"close football matche", "Close football match", "bet", "update"},
		{"udpate football round price", "Update football round price", "bet", "update"},
		{"update cryptokings config", "Update cryptokings config", "bet", "update"},
		{"get game summary", "Get game summary", "bet", "read"},
		{"get transaction summary", "Get transaction summary", "bet", "read"},
		{"update games", "Update games", "bet", "update"},
		{"disable games", "Disable games", "bet", "update"},
		{"get available games", "Get available games", "bet", "read"},
		{"delete games", "Delete games", "bet", "delete"},
		{"add games", "Add games", "bet", "create"},
		{"update game status", "Update game status", "bet", "update"},
		{"create mysteries", "Create mysteries", "bet", "create"},
		{"get mysteries", "Get mysteries", "bet", "read"},
		{"delete mysteries", "Delete mysteries", "bet", "delete"},
		{"update mysteries", "Update mysteries", "bet", "update"},
		{"create spinning wheel config", "Create spinning wheel config", "bet", "create"},
		{"get spinning wheel configs", "Get spinning wheel configs", "bet", "read"},
		{"update spinning wheel config", "Update spinning wheel config", "bet", "update"},
		{"delete spinning wheel config", "Delete spinning wheel config", "bet", "delete"},
		{"create spinning wheel rewards", "Create spinning wheel rewards", "bet", "create"},
		{"get spinning wheel rewards", "Get spinning wheel rewards", "bet", "read"},
		{"update spinning wheel rewards", "Update spinning wheel rewards", "bet", "update"},
		{"delete spinning wheel rewards", "Delete spinning wheel rewards", "bet", "delete"},
		{"create loot box", "Create loot box", "bet", "create"},
		{"get loot box", "Get loot box", "bet", "read"},
		{"update loot box", "Update loot box", "bet", "update"},
		{"delete loot box", "Delete loot box", "bet", "delete"},
		{"create tournament", "Create tournament", "bet", "create"},
		{"get tournaments", "Get tournaments", "bet", "read"},
		{"update tournament", "Update tournament", "bet", "update"},
		{"delete tournament", "Delete tournament", "bet", "delete"},
		{"create bet level", "Create bet level", "bet", "create"},
		{"get bet levels", "Get bet levels", "bet", "read"},
		{"update bet level", "Update bet level", "bet", "update"},
		{"delete bet level", "Delete bet level", "bet", "delete"},
		{"create level requirements", "Create level requirements", "bet", "create"},
		{"get level requirements", "Get level requirements", "bet", "read"},
		{"update level requirements", "Update level requirements", "bet", "update"},
		{"delete level requirements", "Delete level requirements", "bet", "delete"},

		// Agent permissions
		{"Get Agent Referrals", "Get agent referrals", "agent", "read"},
		{"Get Agent Referral Stats", "Get agent referral stats", "agent", "read"},
		{"Create Agent Provider", "Create agent provider", "agent", "create"},

		// Lottery permissions
		{"Create Lottery Service", "Create lottery service", "lottery", "create"},
		{"Create Lottery", "Create lottery", "lottery", "create"},

		// Department permissions
		{"create departments", "Create departments", "department", "create"},
		{"get departments", "Get departments", "department", "read"},
		{"update department", "Update department", "department", "update"},
		{"assign userto depertment", "Assign user to department", "department", "update"},

		// Company permissions
		{"create company", "Create company", "company", "create"},
		{"get companies", "Get companies", "company", "read"},
		{"get company", "Get company", "company", "read"},
		{"update company", "Update company", "company", "update"},
		{"delete company", "Delete company", "company", "delete"},
		{"add ip to company", "Add IP to company", "company", "update"},

		// Operational Group permissions
		{"add operational group", "Add operational group", "operational_group", "create"},
		{"get operational groups", "Get operational groups", "operational_group", "read"},
		{"update operational group", "Update operational group", "operational_group", "update"},
		{"delete operational group", "Delete operational group", "operational_group", "delete"},

		// Operations Definitions permissions
		{"add operations definitions", "Add operations definitions", "operations", "create"},
		{"get operations definitions", "Get operations definitions", "operations", "read"},
		{"update operations definitions", "Update operations definitions", "operations", "update"},
		{"delete operations definitions", "Delete operations definitions", "operations", "delete"},

		// Banner permissions
		{"banner create", "Create banner", "banner", "create"},
		{"banner read", "Read banner", "banner", "read"},
		{"banner update", "Update banner", "banner", "update"},
		{"banner delete", "Delete banner", "banner", "delete"},
		{"banner display", "Display banner", "banner", "read"},
		{"banner image upload", "Upload banner image", "banner", "update"},

		// Adds/Advertisements permissions
		{"create adds service", "Create adds service", "adds", "create"},
		{"get adds services", "Get adds services", "adds", "read"},
		{"update adds service", "Update adds service", "adds", "update"},
		{"delete adds service", "Delete adds service", "adds", "delete"},

		// Airtime permissions
		{"get airtime utilities", "Get airtime utilities", "airtime", "read"},
		{"get airtime utilities stats", "Get airtime utilities stats", "airtime", "read"},
		{"get airtime transactions", "Get airtime transactions", "airtime", "read"},

		// Balance Logs permissions
		{"get balance logs", "Get balance logs", "balance_logs", "read"},

		// Financial Metrics permissions
		{"get financial metrics", "Get financial metrics", "analytics", "read"},

		// Game Metrics permissions
		{"get game metrics", "Get game metrics", "analytics", "read"},

		// Daily Report permissions
		{"get daily report", "Get daily report", "analytics", "read"},

		// Audit Logs permissions
		{"Get Audit Logs", "Get audit logs", "logs", "read"},

		// Available Logs Module permissions
		{"Get Available Logs Module", "Get available logs module", "logs", "read"},
	}

	// Get existing permissions from database
	existingPerms := make(map[string]bool)
	rows, err := db.QueryContext(ctx, "SELECT name FROM permissions")
	if err != nil {
		log.Fatalf("Failed to query existing permissions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Printf("Error scanning permission: %v", err)
			continue
		}
		existingPerms[strings.ToLower(name)] = true
	}

	fmt.Printf("\nFound %d existing permissions in database\n", len(existingPerms))

	// Insert missing permissions
	inserted := 0
	skipped := 0

	for _, perm := range permissions {
		permNameLower := strings.ToLower(perm.name)
		if existingPerms[permNameLower] {
			fmt.Printf("✓ Permission '%s' already exists\n", perm.name)
			skipped++
			continue
		}

		// Check if permission already exists by name
		var existingID string
		err := db.QueryRowContext(ctx, "SELECT id FROM permissions WHERE name = $1", perm.name).Scan(&existingID)
		if err == nil {
			// Permission already exists
			fmt.Printf("✓ Permission '%s' already exists\n", perm.name)
			skipped++
			continue
		} else if err != sql.ErrNoRows {
			log.Printf("Error checking permission '%s': %v", perm.name, err)
			continue
		}

		// Insert new permission (schema: id, name, description only)
		_, err = db.ExecContext(ctx,
			`INSERT INTO permissions (id, name, description) 
			 VALUES ($1, $2, $3)`,
			uuid.New().String(),
			perm.name,
			perm.description,
		)

		if err != nil {
			log.Printf("Error inserting permission '%s': %v", perm.name, err)
			continue
		}

		fmt.Printf("+ Added permission: '%s' (%s/%s)\n", perm.name, perm.resource, perm.action)
		inserted++
		existingPerms[permNameLower] = true
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total permissions checked: %d\n", len(permissions))
	fmt.Printf("New permissions added: %d\n", inserted)
	fmt.Printf("Existing permissions skipped: %d\n", skipped)
	fmt.Printf("\nDone!\n")
}

