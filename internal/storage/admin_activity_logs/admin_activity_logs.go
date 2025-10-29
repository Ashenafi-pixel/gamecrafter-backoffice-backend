package admin_activity_logs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type AdminActivityLogsStorage interface {
	CreateAdminActivityLog(ctx context.Context, req dto.CreateAdminActivityLogReq) (dto.AdminActivityLog, error)
	GetAdminActivityLogs(ctx context.Context, req dto.GetAdminActivityLogsReq) (dto.AdminActivityLogsRes, error)
	GetAdminActivityLogByID(ctx context.Context, id uuid.UUID) (dto.AdminActivityLog, error)
	GetAdminActivityStats(ctx context.Context, from, to *time.Time) (dto.AdminActivityStats, error)
	GetAdminActivityCategories(ctx context.Context) ([]dto.AdminActivityCategory, error)
	GetAdminActivityActions(ctx context.Context) ([]dto.AdminActivityAction, error)
	GetAdminActivityActionsByCategory(ctx context.Context, category string) ([]dto.AdminActivityAction, error)
	DeleteAdminActivityLog(ctx context.Context, id uuid.UUID) error
	DeleteAdminActivityLogsByAdmin(ctx context.Context, adminUserID uuid.UUID) error
	DeleteOldAdminActivityLogs(ctx context.Context, before time.Time) error
}

type adminActivityLogsStorage struct {
	db  persistencedb.PersistenceDB
	log *zap.Logger
}

func NewAdminActivityLogsStorage(db persistencedb.PersistenceDB, log *zap.Logger) AdminActivityLogsStorage {
	return &adminActivityLogsStorage{
		db:  db,
		log: log,
	}
}

func (a *adminActivityLogsStorage) CreateAdminActivityLog(ctx context.Context, req dto.CreateAdminActivityLogReq) (dto.AdminActivityLog, error) {
	var detailsJSON sql.NullString
	if req.Details != nil {
		detailsBytes, err := json.Marshal(req.Details)
		if err != nil {
			a.log.Error("Failed to marshal details", zap.Error(err))
			return dto.AdminActivityLog{}, err
		}
		detailsJSON = sql.NullString{String: string(detailsBytes), Valid: true}
	}

	var resourceID *uuid.UUID
	if req.ResourceID != nil {
		resourceID = req.ResourceID
	}

	var ipAddress sql.NullString
	if req.IPAddress != "" {
		ipAddress = sql.NullString{String: req.IPAddress, Valid: true}
	}

	var userAgent sql.NullString
	if req.UserAgent != "" {
		userAgent = sql.NullString{String: req.UserAgent, Valid: true}
	}

	var sessionID sql.NullString
	if req.SessionID != "" {
		sessionID = sql.NullString{String: req.SessionID, Valid: true}
	}

	// Set default severity if not provided
	severity := req.Severity
	if severity == "" {
		severity = "info"
	}

	query := `
		INSERT INTO admin_activity_logs (
			id, admin_user_id, action, resource_type, resource_id, 
			description, details, ip_address, user_agent, session_id, 
			severity, category, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		) RETURNING id, admin_user_id, action, resource_type, resource_id, 
			description, details, ip_address, user_agent, session_id, 
			severity, category, created_at, updated_at
	`

	var adminLog dto.AdminActivityLog
	var detailsStr sql.NullString
	var ipAddr sql.NullString
	var ua sql.NullString
	var sessID sql.NullString
	var resID sql.NullString

	err := a.db.GetPool().QueryRow(ctx, query,
		req.AdminUserID,
		req.Action,
		req.ResourceType,
		resourceID,
		req.Description,
		detailsJSON,
		ipAddress,
		userAgent,
		sessionID,
		severity,
		req.Category,
	).Scan(
		&adminLog.ID,
		&adminLog.AdminUserID,
		&adminLog.Action,
		&adminLog.ResourceType,
		&resID,
		&adminLog.Description,
		&detailsStr,
		&ipAddr,
		&ua,
		&sessID,
		&adminLog.Severity,
		&adminLog.Category,
		&adminLog.CreatedAt,
		&adminLog.UpdatedAt,
	)

	if err != nil {
		a.log.Error("Failed to create admin activity log", zap.Error(err))
		return dto.AdminActivityLog{}, err
	}

	// Parse resource ID
	if resID.Valid {
		if id, err := uuid.Parse(resID.String); err == nil {
			adminLog.ResourceID = &id
		}
	}

	// Parse details
	if detailsStr.Valid {
		var details interface{}
		if err := json.Unmarshal([]byte(detailsStr.String), &details); err == nil {
			adminLog.Details = details
		}
	}

	// Set nullable fields
	if ipAddr.Valid {
		adminLog.IPAddress = ipAddr.String
	}
	if ua.Valid {
		adminLog.UserAgent = ua.String
	}
	if sessID.Valid {
		adminLog.SessionID = sessID.String
	}

	return adminLog, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityLogs(ctx context.Context, req dto.GetAdminActivityLogsReq) (dto.AdminActivityLogsRes, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.AdminUserID != nil {
		whereClause += fmt.Sprintf(" AND admin_user_id = $%d", argIndex)
		args = append(args, *req.AdminUserID)
		argIndex++
	}

	if req.Action != "" {
		whereClause += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, req.Action)
		argIndex++
	}

	if req.ResourceType != "" {
		whereClause += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, req.ResourceType)
		argIndex++
	}

	if req.Category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, req.Category)
		argIndex++
	}

	if req.Severity != "" {
		whereClause += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, req.Severity)
		argIndex++
	}

	if req.From != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *req.From)
		argIndex++
	}

	if req.To != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *req.To)
		argIndex++
	}

	// Set default sort parameters
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}

	// Set default pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}

	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_activity_logs %s", whereClause)
	var total int64
	err := a.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		a.log.Error("Failed to count admin activity logs", zap.Error(err))
		return dto.AdminActivityLogsRes{}, err
	}

	// Get logs with admin user information
	query := fmt.Sprintf(`
		SELECT aal.id, aal.admin_user_id, aal.action, aal.resource_type, aal.resource_id, 
			aal.description, aal.details, aal.ip_address, aal.user_agent, aal.session_id, 
			aal.severity, aal.category, aal.created_at, aal.updated_at,
			COALESCE(u.username, 'Unknown') as admin_username,
			COALESCE(u.email, '') as admin_email
		FROM admin_activity_logs aal
		LEFT JOIN users u ON aal.admin_user_id = u.id
		%s 
		ORDER BY %s %s 
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, perPage, offset)

	rows, err := a.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		a.log.Error("Failed to get admin activity logs", zap.Error(err))
		return dto.AdminActivityLogsRes{}, err
	}
	defer rows.Close()

	var logs []dto.AdminActivityLog
	for rows.Next() {
		var log dto.AdminActivityLog
		var detailsStr sql.NullString
		var ipAddr sql.NullString
		var ua sql.NullString
		var sessID sql.NullString
		var resID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.AdminUserID,
			&log.Action,
			&log.ResourceType,
			&resID,
			&log.Description,
			&detailsStr,
			&ipAddr,
			&ua,
			&sessID,
			&log.Severity,
			&log.Category,
			&log.CreatedAt,
			&log.UpdatedAt,
			&log.AdminUsername,
			&log.AdminEmail,
		)
		if err != nil {
			a.log.Error("Failed to scan admin activity log", zap.Error(err))
			continue
		}

		// Parse resource ID
		if resID.Valid {
			if id, err := uuid.Parse(resID.String); err == nil {
				log.ResourceID = &id
			}
		}

		// Parse details
		if detailsStr.Valid {
			var details interface{}
			if err := json.Unmarshal([]byte(detailsStr.String), &details); err == nil {
				log.Details = details
			}
		}

		// Set nullable fields
		if ipAddr.Valid {
			log.IPAddress = ipAddr.String
		}
		if ua.Valid {
			log.UserAgent = ua.String
		}
		if sessID.Valid {
			log.SessionID = sessID.String
		}

		logs = append(logs, log)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return dto.AdminActivityLogsRes{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityLogByID(ctx context.Context, id uuid.UUID) (dto.AdminActivityLog, error) {
	query := `
		SELECT aal.id, aal.admin_user_id, aal.action, aal.resource_type, aal.resource_id, 
			aal.description, aal.details, aal.ip_address, aal.user_agent, aal.session_id, 
			aal.severity, aal.category, aal.created_at, aal.updated_at,
			COALESCE(u.username, 'Unknown') as admin_username,
			COALESCE(u.email, '') as admin_email
		FROM admin_activity_logs aal
		LEFT JOIN users u ON aal.admin_user_id = u.id
		WHERE aal.id = $1
	`

	var log dto.AdminActivityLog
	var detailsStr sql.NullString
	var ipAddr sql.NullString
	var ua sql.NullString
	var sessID sql.NullString
	var resID sql.NullString

	err := a.db.GetPool().QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.AdminUserID,
		&log.Action,
		&log.ResourceType,
		&resID,
		&log.Description,
		&detailsStr,
		&ipAddr,
		&ua,
		&sessID,
		&log.Severity,
		&log.Category,
		&log.CreatedAt,
		&log.UpdatedAt,
		&log.AdminUsername,
		&log.AdminEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return dto.AdminActivityLog{}, fmt.Errorf("admin activity log not found")
		}
		a.log.Error("Failed to get admin activity log by ID", zap.Error(err))
		return dto.AdminActivityLog{}, err
	}

	// Parse resource ID
	if resID.Valid {
		if id, err := uuid.Parse(resID.String); err == nil {
			log.ResourceID = &id
		}
	}

	// Parse details
	if detailsStr.Valid {
		var details interface{}
		if err := json.Unmarshal([]byte(detailsStr.String), &details); err == nil {
			log.Details = details
		}
	}

	// Set nullable fields
	if ipAddr.Valid {
		log.IPAddress = ipAddr.String
	}
	if ua.Valid {
		log.UserAgent = ua.String
	}
	if sessID.Valid {
		log.SessionID = sessID.String
	}

	return log, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityStats(ctx context.Context, from, to *time.Time) (dto.AdminActivityStats, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if from != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *from)
		argIndex++
	}

	if to != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *to)
		argIndex++
	}

	// Get total logs
	var totalLogs int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_activity_logs %s", whereClause)
	err := a.db.GetPool().QueryRow(ctx, countQuery, args...).Scan(&totalLogs)
	if err != nil {
		a.log.Error("Failed to count admin activity logs", zap.Error(err))
		return dto.AdminActivityStats{}, err
	}

	// Get logs by category
	categoryQuery := fmt.Sprintf(`
		SELECT category, COUNT(*) as count 
		FROM admin_activity_logs %s 
		GROUP BY category
	`, whereClause)

	categoryRows, err := a.db.GetPool().Query(ctx, categoryQuery, args...)
	if err != nil {
		a.log.Error("Failed to get logs by category", zap.Error(err))
		return dto.AdminActivityStats{}, err
	}
	defer categoryRows.Close()

	activitiesByCategory := make(map[string]int64)
	for categoryRows.Next() {
		var category string
		var count int64
		if err := categoryRows.Scan(&category, &count); err == nil {
			activitiesByCategory[category] = count
		}
	}

	// Get logs by action
	actionQuery := fmt.Sprintf(`
		SELECT action, COUNT(*) as count 
		FROM admin_activity_logs %s 
		GROUP BY action
	`, whereClause)

	actionRows, err := a.db.GetPool().Query(ctx, actionQuery, args...)
	if err != nil {
		a.log.Error("Failed to get logs by action", zap.Error(err))
		return dto.AdminActivityStats{}, err
	}
	defer actionRows.Close()

	activitiesByAction := make(map[string]int64)
	for actionRows.Next() {
		var action string
		var count int64
		if err := actionRows.Scan(&action, &count); err == nil {
			activitiesByAction[action] = count
		}
	}

	// Get recent activity (last 10 logs)
	recentQuery := fmt.Sprintf(`
		SELECT id, admin_user_id, action, resource_type, resource_id, 
			description, details, ip_address, user_agent, session_id, 
			severity, category, created_at, updated_at
		FROM admin_activity_logs %s 
		ORDER BY created_at DESC 
		LIMIT 10
	`, whereClause)

	recentRows, err := a.db.GetPool().Query(ctx, recentQuery, args...)
	if err != nil {
		a.log.Error("Failed to get recent activity", zap.Error(err))
		return dto.AdminActivityStats{}, err
	}
	defer recentRows.Close()

	var recentActivity []dto.AdminActivityLog
	for recentRows.Next() {
		var log dto.AdminActivityLog
		var detailsStr sql.NullString
		var ipAddr sql.NullString
		var ua sql.NullString
		var sessID sql.NullString
		var resID sql.NullString

		err := recentRows.Scan(
			&log.ID,
			&log.AdminUserID,
			&log.Action,
			&log.ResourceType,
			&resID,
			&log.Description,
			&detailsStr,
			&ipAddr,
			&ua,
			&sessID,
			&log.Severity,
			&log.Category,
			&log.CreatedAt,
			&log.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse resource ID
		if resID.Valid {
			if id, err := uuid.Parse(resID.String); err == nil {
				log.ResourceID = &id
			}
		}

		// Parse details
		if detailsStr.Valid {
			var details interface{}
			if err := json.Unmarshal([]byte(detailsStr.String), &details); err == nil {
				log.Details = details
			}
		}

		// Set nullable fields
		if ipAddr.Valid {
			log.IPAddress = ipAddr.String
		}
		if ua.Valid {
			log.UserAgent = ua.String
		}
		if sessID.Valid {
			log.SessionID = sessID.String
		}

		recentActivity = append(recentActivity, log)
	}

	return dto.AdminActivityStats{
		TotalActivities:      totalLogs,
		ActivitiesByCategory: activitiesByCategory,
		ActivitiesByAction:   activitiesByAction,
		RecentActivities:     recentActivity,
	}, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityCategories(ctx context.Context) ([]dto.AdminActivityCategory, error) {
	query := `
		SELECT id, name, description, color, icon, is_active, created_at
		FROM admin_activity_categories
		ORDER BY name
	`

	rows, err := a.db.GetPool().Query(ctx, query)
	if err != nil {
		a.log.Error("Failed to get admin activity categories", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var categories []dto.AdminActivityCategory
	for rows.Next() {
		var category dto.AdminActivityCategory
		var description sql.NullString
		var color sql.NullString
		var icon sql.NullString

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&description,
			&color,
			&icon,
			&category.IsActive,
			&category.CreatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			category.Description = description.String
		}
		if color.Valid {
			category.Color = color.String
		}
		if icon.Valid {
			category.Icon = icon.String
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityActions(ctx context.Context) ([]dto.AdminActivityAction, error) {
	query := `
		SELECT id, name, description, category_id, is_active, created_at
		FROM admin_activity_actions
		ORDER BY name
	`

	rows, err := a.db.GetPool().Query(ctx, query)
	if err != nil {
		a.log.Error("Failed to get admin activity actions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var actions []dto.AdminActivityAction
	for rows.Next() {
		var action dto.AdminActivityAction
		var description sql.NullString

		err := rows.Scan(
			&action.ID,
			&action.Name,
			&description,
			&action.CategoryID,
			&action.IsActive,
			&action.CreatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			action.Description = description.String
		}

		actions = append(actions, action)
	}

	return actions, nil
}

func (a *adminActivityLogsStorage) GetAdminActivityActionsByCategory(ctx context.Context, category string) ([]dto.AdminActivityAction, error) {
	query := `
		SELECT a.id, a.name, a.description, a.category_id, a.is_active, a.created_at
		FROM admin_activity_actions a
		JOIN admin_activity_categories c ON a.category_id = c.id
		WHERE c.name = $1
		ORDER BY a.name
	`

	rows, err := a.db.GetPool().Query(ctx, query, category)
	if err != nil {
		a.log.Error("Failed to get admin activity actions by category", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var actions []dto.AdminActivityAction
	for rows.Next() {
		var action dto.AdminActivityAction
		var description sql.NullString

		err := rows.Scan(
			&action.ID,
			&action.Name,
			&description,
			&action.CategoryID,
			&action.IsActive,
			&action.CreatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			action.Description = description.String
		}

		actions = append(actions, action)
	}

	return actions, nil
}

func (a *adminActivityLogsStorage) DeleteAdminActivityLog(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM admin_activity_logs WHERE id = $1"
	result, err := a.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		a.log.Error("Failed to delete admin activity log", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("admin activity log not found")
	}

	return nil
}

func (a *adminActivityLogsStorage) DeleteAdminActivityLogsByAdmin(ctx context.Context, adminUserID uuid.UUID) error {
	query := "DELETE FROM admin_activity_logs WHERE admin_user_id = $1"
	_, err := a.db.GetPool().Exec(ctx, query, adminUserID)
	if err != nil {
		a.log.Error("Failed to delete admin activity logs by admin", zap.Error(err))
		return err
	}

	return nil
}

func (a *adminActivityLogsStorage) DeleteOldAdminActivityLogs(ctx context.Context, before time.Time) error {
	query := "DELETE FROM admin_activity_logs WHERE created_at < $1"
	_, err := a.db.GetPool().Exec(ctx, query, before)
	if err != nil {
		a.log.Error("Failed to delete old admin activity logs", zap.Error(err))
		return err
	}

	return nil
}
