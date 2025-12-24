package page

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type PageStorage interface {
	CreatePage(ctx context.Context, page dto.CreatePageReq) (dto.Page, error)
	GetPageByPath(ctx context.Context, path string) (dto.Page, bool, error)
	GetAllPages(ctx context.Context) ([]dto.Page, error)
	GetPagesByParentID(ctx context.Context, parentID uuid.UUID) ([]dto.Page, error)
	GetUserAllowedPages(ctx context.Context, userID uuid.UUID) ([]dto.Page, error)
	AssignPagesToUser(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error
	RemovePagesFromUser(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error
	ReplaceUserPages(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error
	GetAllPagesForUser(ctx context.Context, userID uuid.UUID) ([]dto.Page, error)
	// Role-based page access methods
	GetRoleAllowedPages(ctx context.Context, roleID uuid.UUID) ([]dto.Page, error)
	AssignPagesToRole(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error
	RemovePagesFromRole(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error
	ReplaceRolePages(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error
}

type pageStorage struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

func Init(db *persistencedb.PersistenceDB, log *zap.Logger) PageStorage {
	return &pageStorage{
		db:  db,
		log: log,
	}
}

func (p *pageStorage) CreatePage(ctx context.Context, pageReq dto.CreatePageReq) (dto.Page, error) {
	var parentID sql.NullString
	if pageReq.ParentID != nil {
		parentID = sql.NullString{String: pageReq.ParentID.String(), Valid: true}
	}

	var icon sql.NullString
	if pageReq.Icon != "" {
		icon = sql.NullString{String: pageReq.Icon, Valid: true}
	}

	var pageID uuid.UUID
	query := `
		INSERT INTO pages (path, label, parent_id, icon)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := p.db.GetPool().QueryRow(ctx, query, pageReq.Path, pageReq.Label, parentID, icon).Scan(&pageID)
	if err != nil {
		p.log.Error("failed to create page", zap.Error(err), zap.String("path", pageReq.Path))
		return dto.Page{}, fmt.Errorf("failed to create page: %w", err)
	}

	page := dto.Page{
		ID:       pageID,
		Path:     pageReq.Path,
		Label:    pageReq.Label,
		ParentID: pageReq.ParentID,
		Icon:     pageReq.Icon,
	}

	return page, nil
}

func (p *pageStorage) GetPageByPath(ctx context.Context, path string) (dto.Page, bool, error) {
	var page dto.Page
	var parentID sql.NullString
	var icon sql.NullString

	query := `
		SELECT id, path, label, parent_id, icon
		FROM pages
		WHERE path = $1
	`

	err := p.db.GetPool().QueryRow(ctx, query, path).Scan(
		&page.ID,
		&page.Path,
		&page.Label,
		&parentID,
		&icon,
	)

	if err == sql.ErrNoRows {
		return dto.Page{}, false, nil
	}
	if err != nil {
		p.log.Error("failed to get page by path", zap.Error(err), zap.String("path", path))
		return dto.Page{}, false, fmt.Errorf("failed to get page by path: %w", err)
	}

	if parentID.Valid {
		parentUUID, err := uuid.Parse(parentID.String)
		if err == nil {
			page.ParentID = &parentUUID
		}
	}

	if icon.Valid {
		page.Icon = icon.String
	}

	return page, true, nil
}

func (p *pageStorage) GetAllPages(ctx context.Context) ([]dto.Page, error) {
	query := `
		SELECT id, path, label, parent_id, icon
		FROM pages
		ORDER BY label
	`

	rows, err := p.db.GetPool().Query(ctx, query)
	if err != nil {
		p.log.Error("failed to get all pages", zap.Error(err))
		return nil, fmt.Errorf("failed to get all pages: %w", err)
	}
	defer rows.Close()

	var pages []dto.Page
	for rows.Next() {
		var page dto.Page
		var parentID sql.NullString
		var icon sql.NullString

		err := rows.Scan(&page.ID, &page.Path, &page.Label, &parentID, &icon)
		if err != nil {
			p.log.Error("failed to scan page", zap.Error(err))
			continue
		}

		if parentID.Valid {
			parentUUID, err := uuid.Parse(parentID.String)
			if err == nil {
				page.ParentID = &parentUUID
			}
		}

		if icon.Valid {
			page.Icon = icon.String
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func (p *pageStorage) GetPagesByParentID(ctx context.Context, parentID uuid.UUID) ([]dto.Page, error) {
	query := `
		SELECT id, path, label, parent_id, icon
		FROM pages
		WHERE parent_id = $1
		ORDER BY label
	`

	rows, err := p.db.GetPool().Query(ctx, query, parentID)
	if err != nil {
		p.log.Error("failed to get pages by parent ID", zap.Error(err), zap.String("parent_id", parentID.String()))
		return nil, fmt.Errorf("failed to get pages by parent ID: %w", err)
	}
	defer rows.Close()

	var pages []dto.Page
	for rows.Next() {
		var page dto.Page
		var parentID sql.NullString
		var icon sql.NullString

		err := rows.Scan(&page.ID, &page.Path, &page.Label, &parentID, &icon)
		if err != nil {
			p.log.Error("failed to scan page", zap.Error(err))
			continue
		}

		if parentID.Valid {
			parentUUID, err := uuid.Parse(parentID.String)
			if err == nil {
				page.ParentID = &parentUUID
			}
		}

		if icon.Valid {
			page.Icon = icon.String
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func (p *pageStorage) GetUserAllowedPages(ctx context.Context, userID uuid.UUID) ([]dto.Page, error) {
	query := `
		SELECT p.id, p.path, p.label, p.parent_id, p.icon
		FROM pages p
		INNER JOIN user_allowed_pages uap ON p.id = uap.page_id
		WHERE uap.user_id = $1
		ORDER BY p.label
	`

	rows, err := p.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		p.log.Error("failed to get user allowed pages", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user allowed pages: %w", err)
	}
	defer rows.Close()

	var pages []dto.Page
	for rows.Next() {
		var page dto.Page
		var parentID sql.NullString
		var icon sql.NullString

		err := rows.Scan(&page.ID, &page.Path, &page.Label, &parentID, &icon)
		if err != nil {
			p.log.Error("failed to scan page", zap.Error(err))
			continue
		}

		if parentID.Valid {
			parentUUID, err := uuid.Parse(parentID.String)
			if err == nil {
				page.ParentID = &parentUUID
			}
		}

		if icon.Valid {
			page.Icon = icon.String
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func (p *pageStorage) AssignPagesToUser(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error {
	if len(pageIDs) == 0 {
		return nil
	}

	tx, err := p.db.GetPool().Begin(ctx)
	if err != nil {
		p.log.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO user_allowed_pages (user_id, page_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, page_id) DO NOTHING
	`

	for _, pageID := range pageIDs {
		_, err := tx.Exec(ctx, query, userID, pageID)
		if err != nil {
			p.log.Error("failed to assign page to user", zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("page_id", pageID.String()))
			return fmt.Errorf("failed to assign page to user: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *pageStorage) RemovePagesFromUser(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error {
	if len(pageIDs) == 0 {
		return nil
	}

	query := `
		DELETE FROM user_allowed_pages
		WHERE user_id = $1 AND page_id = ANY($2::uuid[])
	`

	_, err := p.db.GetPool().Exec(ctx, query, userID, pageIDs)
	if err != nil {
		p.log.Error("failed to remove pages from user", zap.Error(err),
			zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to remove pages from user: %w", err)
	}

	return nil
}

// ReplaceUserPages replaces all pages for a user (removes all existing and adds new ones)
func (p *pageStorage) ReplaceUserPages(ctx context.Context, userID uuid.UUID, pageIDs []uuid.UUID) error {
	tx, err := p.db.GetPool().Begin(ctx)
	if err != nil {
		p.log.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First, delete all existing pages for this user
	deleteQuery := `DELETE FROM user_allowed_pages WHERE user_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, userID)
	if err != nil {
		p.log.Error("failed to delete existing pages for user", zap.Error(err),
			zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete existing pages for user: %w", err)
	}

	// If no pages to add, just commit the deletion
	if len(pageIDs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			p.log.Error("failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Then, insert the new pages
	insertQuery := `
		INSERT INTO user_allowed_pages (user_id, page_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, page_id) DO NOTHING
	`

	for _, pageID := range pageIDs {
		_, err := tx.Exec(ctx, insertQuery, userID, pageID)
		if err != nil {
			p.log.Error("failed to assign page to user", zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("page_id", pageID.String()))
			return fmt.Errorf("failed to assign page to user: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *pageStorage) GetAllPagesForUser(ctx context.Context, userID uuid.UUID) ([]dto.Page, error) {
	// Get all pages (parent + children) that the user has access to
	// This includes parent pages if any child is allowed, and all allowed child pages
	query := `
		WITH user_pages AS (
			SELECT DISTINCT p.id, p.path, p.label, p.parent_id, p.icon
			FROM pages p
			INNER JOIN user_allowed_pages uap ON p.id = uap.page_id
			WHERE uap.user_id = $1
		),
		parent_pages AS (
			SELECT DISTINCT p.id, p.path, p.label, p.parent_id, p.icon
			FROM pages p
			INNER JOIN user_pages up ON p.id = up.parent_id
		)
		SELECT id, path, label, parent_id, icon FROM user_pages
		UNION
		SELECT id, path, label, parent_id, icon FROM parent_pages
		ORDER BY label
	`

	rows, err := p.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		p.log.Error("failed to get all pages for user", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get all pages for user: %w", err)
	}
	defer rows.Close()

	var pages []dto.Page
	for rows.Next() {
		var page dto.Page
		var parentID sql.NullString
		var icon sql.NullString

		err := rows.Scan(&page.ID, &page.Path, &page.Label, &parentID, &icon)
		if err != nil {
			p.log.Error("failed to scan page", zap.Error(err))
			continue
		}

		if parentID.Valid {
			parentUUID, err := uuid.Parse(parentID.String)
			if err == nil {
				page.ParentID = &parentUUID
			}
		}

		if icon.Valid {
			page.Icon = icon.String
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// GetRoleAllowedPages gets allowed pages for a specific role
func (p *pageStorage) GetRoleAllowedPages(ctx context.Context, roleID uuid.UUID) ([]dto.Page, error) {
	query := `
		SELECT p.id, p.path, p.label, p.parent_id, p.icon
		FROM pages p
		INNER JOIN role_allowed_pages rap ON p.id = rap.page_id
		WHERE rap.role_id = $1
		ORDER BY p.label
	`

	rows, err := p.db.GetPool().Query(ctx, query, roleID)
	if err != nil {
		p.log.Error("failed to get role allowed pages", zap.Error(err), zap.String("role_id", roleID.String()))
		return nil, fmt.Errorf("failed to get role allowed pages: %w", err)
	}
	defer rows.Close()

	var pages []dto.Page
	for rows.Next() {
		var page dto.Page
		var parentID sql.NullString
		var icon sql.NullString

		err := rows.Scan(&page.ID, &page.Path, &page.Label, &parentID, &icon)
		if err != nil {
			p.log.Error("failed to scan page", zap.Error(err))
			continue
		}

		if parentID.Valid {
			parentUUID, err := uuid.Parse(parentID.String)
			if err == nil {
				page.ParentID = &parentUUID
			}
		}

		if icon.Valid {
			page.Icon = icon.String
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// AssignPagesToRole assigns pages to a role
func (p *pageStorage) AssignPagesToRole(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error {
	if len(pageIDs) == 0 {
		return nil
	}

	tx, err := p.db.GetPool().Begin(ctx)
	if err != nil {
		p.log.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO role_allowed_pages (role_id, page_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, page_id) DO NOTHING
	`

	for _, pageID := range pageIDs {
		_, err := tx.Exec(ctx, query, roleID, pageID)
		if err != nil {
			p.log.Error("failed to assign page to role", zap.Error(err),
				zap.String("role_id", roleID.String()),
				zap.String("page_id", pageID.String()))
			return fmt.Errorf("failed to assign page to role: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemovePagesFromRole removes pages from a role
func (p *pageStorage) RemovePagesFromRole(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error {
	if len(pageIDs) == 0 {
		return nil
	}

	query := `
		DELETE FROM role_allowed_pages
		WHERE role_id = $1 AND page_id = ANY($2::uuid[])
	`

	_, err := p.db.GetPool().Exec(ctx, query, roleID, pageIDs)
	if err != nil {
		p.log.Error("failed to remove pages from role", zap.Error(err),
			zap.String("role_id", roleID.String()))
		return fmt.Errorf("failed to remove pages from role: %w", err)
	}

	return nil
}

// ReplaceRolePages replaces all pages for a role (removes all existing and adds new ones)
func (p *pageStorage) ReplaceRolePages(ctx context.Context, roleID uuid.UUID, pageIDs []uuid.UUID) error {
	tx, err := p.db.GetPool().Begin(ctx)
	if err != nil {
		p.log.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First, delete all existing pages for this role
	deleteQuery := `DELETE FROM role_allowed_pages WHERE role_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, roleID)
	if err != nil {
		p.log.Error("failed to delete existing pages for role", zap.Error(err),
			zap.String("role_id", roleID.String()))
		return fmt.Errorf("failed to delete existing pages for role: %w", err)
	}

	// If no pages to add, just commit the deletion
	if len(pageIDs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			p.log.Error("failed to commit transaction", zap.Error(err))
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Then, insert the new pages
	insertQuery := `
		INSERT INTO role_allowed_pages (role_id, page_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, page_id) DO NOTHING
	`

	for _, pageID := range pageIDs {
		_, err := tx.Exec(ctx, insertQuery, roleID, pageID)
		if err != nil {
			p.log.Error("failed to assign page to role", zap.Error(err),
				zap.String("role_id", roleID.String()),
				zap.String("page_id", pageID.String()))
			return fmt.Errorf("failed to assign page to role: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		p.log.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

