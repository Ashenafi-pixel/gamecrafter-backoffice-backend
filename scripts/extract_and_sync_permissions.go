package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
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

	// Extract permissions from codebase
	permissions, err := extractPermissionsFromCodebase()
	if err != nil {
		log.Fatalf("Failed to extract permissions: %v", err)
	}

	fmt.Printf("\nFound %d unique permissions in codebase\n\n", len(permissions))

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

	fmt.Printf("Found %d existing permissions in database\n\n", len(existingPerms))

	// Insert missing permissions
	inserted := 0
	skipped := 0

	for _, permName := range permissions {
		permNameLower := strings.ToLower(permName)
		if existingPerms[permNameLower] {
			fmt.Printf("✓ Permission '%s' already exists\n", permName)
			skipped++
			continue
		}

		// Generate description from permission name
		description := generateDescription(permName)

		// Check if permission already exists by name
		var existingID string
		err := db.QueryRowContext(ctx, "SELECT id FROM permissions WHERE LOWER(name) = $1", permNameLower).Scan(&existingID)
		if err == nil {
			// Permission already exists (case-insensitive match)
			fmt.Printf("✓ Permission '%s' already exists (case variant)\n", permName)
			skipped++
			continue
		} else if err != sql.ErrNoRows {
			log.Printf("Error checking permission '%s': %v", permName, err)
			continue
		}

		// Insert new permission
		_, err = db.ExecContext(ctx,
			`INSERT INTO permissions (id, name, description) 
			 VALUES ($1, $2, $3)`,
			uuid.New().String(),
			permName,
			description,
		)

		if err != nil {
			log.Printf("Error inserting permission '%s': %v", permName, err)
			continue
		}

		fmt.Printf("+ Added permission: '%s' - %s\n", permName, description)
		inserted++
		existingPerms[permNameLower] = true
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total permissions found in codebase: %d\n", len(permissions))
	fmt.Printf("New permissions added: %d\n", inserted)
	fmt.Printf("Existing permissions skipped: %d\n", skipped)
	fmt.Printf("\nDone!\n")
}

func extractPermissionsFromCodebase() ([]string, error) {
	// Get current directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	// Check if we're in tucanbit-back or need to change
	if !strings.Contains(wd, "tucanbit-back") {
		if _, err := os.Stat("tucanbit-back"); err == nil {
			os.Chdir("tucanbit-back")
		}
	}

	// Use grep to find all middleware.Authz calls
	cmd := exec.Command("grep", "-rh", "middleware.Authz", "internal/glue", "--include=*.go")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to grep: %w", err)
	}

	// Regex to match permission strings in middleware.Authz calls
	// Pattern: middleware.Authz(..., "permission name", ...)
	permRegex := regexp.MustCompile(`middleware\.Authz\([^,]+,\s*"([^"]+)"`)
	
	permissionsMap := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		matches := permRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				permName := strings.TrimSpace(match[1])
				if permName != "" && permName != "admin" {
					permissionsMap[permName] = true
				}
			}
		}
	}

	// Also check for the special "cashback" and "admin" pattern
	cashbackRegex := regexp.MustCompile(`middleware\.Authz\([^,]+,\s*"([^"]+)"\s*,\s*"([^"]+)"`)
	scanner2 := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner2.Scan() {
		line := scanner2.Text()
		matches := cashbackRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 2 {
				permName := strings.TrimSpace(match[1])
				action := strings.TrimSpace(match[2])
				if permName != "" && action == "admin" {
					permissionsMap[permName] = true
				}
			}
		}
	}

	// Convert map to sorted slice
	permissions := make([]string, 0, len(permissionsMap))
	for perm := range permissionsMap {
		permissions = append(permissions, perm)
	}

	// Sort
	for i := 0; i < len(permissions)-1; i++ {
		for j := i + 1; j < len(permissions); j++ {
			if permissions[i] > permissions[j] {
				permissions[i], permissions[j] = permissions[j], permissions[i]
			}
		}
	}

	return permissions, nil
}

func generateDescription(permName string) string {
	// Convert permission name to a readable description
	permName = strings.ToLower(permName)
	
	// Common patterns
	descriptions := map[string]string{
		"cashback": "Cashback management",
		"get permissions": "Get all permissions",
		"create role": "Create a new role",
		"get roles": "Get all roles",
		"update role permissions": "Update role permissions",
		"remove role": "Remove a role",
		"assign role": "Assign role to user",
		"revoke role": "Revoke role from user",
		"get role users": "Get users with a role",
		"get user roles": "Get roles for a user",
	}

	if desc, ok := descriptions[permName]; ok {
		return desc
	}

	// Generate description from permission name
	words := strings.Fields(permName)
	if len(words) > 0 {
		action := strings.Title(words[0])
		resource := strings.Join(words[1:], " ")
		if resource != "" {
			return fmt.Sprintf("%s %s", action, strings.Title(resource))
		}
		return strings.Title(permName)
	}

	return strings.Title(permName)
}

