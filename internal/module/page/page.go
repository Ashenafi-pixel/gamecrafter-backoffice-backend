package page

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"go.uber.org/zap"
)

type pageModule struct {
	pageStorage storage.Page
	log         *zap.Logger
}

func Init(pageStorage storage.Page, log *zap.Logger) module.Page {
	return &pageModule{
		pageStorage: pageStorage,
		log:         log,
	}
}

// SeedPages seeds all pages from the sidebar and routes configuration
// This will be called during initialization to populate the pages table
func (p *pageModule) SeedPages(ctx context.Context) error {
	// Define all pages based on Sidebar.tsx menuItems and App.tsx Routes
	// Structure: parent pages first, then children with parent references
	
	// Parent pages (sidebar items) - these are the main menu items
	parentPages := []struct {
		path  string
		label string
	}{
		{"/dashboard", "Dashboard"},
		{"/reports", "Reports"},
		{"/players", "Player Management"},
		{"/notifications", "Player Notifications"},
		{"/kyc-management", "KYC Management"},
		{"/cashback", "Rakeback"},
		{"/admin/game-management", "Game Management"},
		{"/admin/brand-management", "Brand Management"},
		{"/admin/falcon-liquidity", "Falcon Liquidity"},
		{"/wallet", "Wallet"},
		{"/settings", "Site Settings"},
		{"/access-control", "Back Office Settings"},
		{"/admin/activity-logs", "Admin Activity Logs"},
		{"/admin/alerts", "Alert Management"},
	}

	// Child pages with their parent paths
	childPages := []struct {
		path       string
		label      string
		parentPath string
	}{
		// Reports children
		{"/reports", "Analytics Dashboard", "/reports"},
		// {"/reports/transaction", "Wallet Report", "/reports"},
		{"/reports/daily", "Daily Report", "/reports"},
		// {"/reports/game", "Game Report", "/reports"},
		{"/reports/big-winners", "Big Winners", "/reports"},
		{"/reports/player-metrics", "Player Metrics", "/reports"},
		{"/reports/country", "Country Report", "/reports"},
		{"/reports/game-performance", "Game Performance", "/reports"},
		{"/reports/provider-performance", "Provider Performance", "/reports"},
		
		// Rakeback children
		{"/cashback", "VIP Levels", "/cashback"},
		{"/admin/rakeback-override", "Happy Hour", "/cashback"},
		
		// Wallet children
		{"/transactions/details", "Transaction Details", "/wallet"},
		{"/transactions/withdrawals/dashboard", "Withdrawal Dashboard", "/wallet"},
		{"/transactions/deposits", "Deposit Management", "/wallet"},
		{"/transactions/manual-funds", "Fund Management", "/wallet"},
		{"/wallet/management", "Wallet Management", "/wallet"},
		{"/transactions/withdrawals", "Withdrawal Management", "/wallet"},
		{"/transactions/withdrawals/settings", "Withdrawal Settings", "/wallet"},
		
		// Additional routes from App.tsx
		{"/kyc-risk", "KYC Risk Management", "/kyc-management"},
	}

	// First, get all existing pages to avoid duplicates
	existingPages, err := p.pageStorage.GetAllPages(ctx)
	if err != nil {
		p.log.Error("failed to get existing pages", zap.Error(err))
		return fmt.Errorf("failed to get existing pages: %w", err)
	}

	existingPaths := make(map[string]bool)
	pathToID := make(map[string]uuid.UUID)
	for _, ep := range existingPages {
		existingPaths[ep.Path] = true
		pathToID[ep.Path] = ep.ID
	}

	// Create parent pages first
	for _, parent := range parentPages {
		if existingPaths[parent.path] {
			p.log.Debug("parent page already exists, skipping", zap.String("path", parent.path))
			continue
		}

		page, err := p.pageStorage.CreatePage(ctx, dto.CreatePageReq{
			Path:     parent.path,
			Label:    parent.label,
			ParentID: nil,
		})
		if err != nil {
			p.log.Error("failed to create parent page", zap.Error(err), zap.String("path", parent.path))
			continue
		}
		pathToID[parent.path] = page.ID
		p.log.Info("created parent page", zap.String("path", parent.path), zap.String("label", parent.label))
	}

	// Create child pages with parent references
	for _, child := range childPages {
		if existingPaths[child.path] {
			p.log.Debug("child page already exists, skipping", zap.String("path", child.path))
			continue
		}

		var parentID *uuid.UUID
		if parentIDVal, ok := pathToID[child.parentPath]; ok {
			parentID = &parentIDVal
		}

		_, err := p.pageStorage.CreatePage(ctx, dto.CreatePageReq{
			Path:     child.path,
			Label:    child.label,
			ParentID: parentID,
		})
		if err != nil {
			p.log.Error("failed to create child page", zap.Error(err), zap.String("path", child.path))
			continue
		}
		p.log.Info("created child page", zap.String("path", child.path), zap.String("label", child.label), zap.String("parent", child.parentPath))
	}

	return nil
}

func (p *pageModule) GetUserAllowedPages(ctx context.Context, userID uuid.UUID) ([]dto.Page, error) {
	return p.pageStorage.GetUserAllowedPages(ctx, userID)
}

func (p *pageModule) AssignAllPagesToUser(ctx context.Context, userID uuid.UUID) error {
	allPages, err := p.pageStorage.GetAllPages(ctx)
	if err != nil {
		p.log.Error("failed to get all pages", zap.Error(err))
		return fmt.Errorf("failed to get all pages: %w", err)
	}

	if len(allPages) == 0 {
		p.log.Warn("no pages found to assign", zap.String("user_id", userID.String()))
		return nil
	}

	pageIDs := make([]uuid.UUID, 0, len(allPages))
	for _, page := range allPages {
		pageIDs = append(pageIDs, page.ID)
	}

	err = p.pageStorage.AssignPagesToUser(ctx, userID, pageIDs)
	if err != nil {
		p.log.Error("failed to assign pages to user", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to assign pages to user: %w", err)
	}

	p.log.Info("assigned all pages to user", zap.String("user_id", userID.String()), zap.Int("page_count", len(pageIDs)))
	return nil
}

func (p *pageModule) GetAllPages(ctx context.Context) ([]dto.Page, error) {
	return p.pageStorage.GetAllPages(ctx)
}

func (p *pageModule) AssignPagesToUser(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error {
	return p.pageStorage.AssignPagesToUser(ctx, userID, pageIDs)
}

func (p *pageModule) ReplaceUserPages(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error {
	return p.pageStorage.ReplaceUserPages(ctx, userID, pageIDs)
}

