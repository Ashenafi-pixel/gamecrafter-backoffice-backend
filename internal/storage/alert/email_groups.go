package alert

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/persistencedb"
	"go.uber.org/zap"
)

type AlertEmailGroupStorage interface {
	CreateEmailGroup(ctx context.Context, req *dto.CreateAlertEmailGroupRequest, createdBy uuid.UUID) (*dto.AlertEmailGroup, error)
	GetEmailGroupByID(ctx context.Context, id uuid.UUID) (*dto.AlertEmailGroup, error)
	GetAllEmailGroups(ctx context.Context) ([]dto.AlertEmailGroup, error)
	UpdateEmailGroup(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertEmailGroupRequest, updatedBy uuid.UUID) (*dto.AlertEmailGroup, error)
	DeleteEmailGroup(ctx context.Context, id uuid.UUID) error
	GetEmailsByGroupIDs(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]string, error)
}

type alertEmailGroupStorage struct {
	db  persistencedb.PersistenceDB
	log *zap.Logger
}

func NewAlertEmailGroupStorage(db persistencedb.PersistenceDB, log *zap.Logger) AlertEmailGroupStorage {
	return &alertEmailGroupStorage{
		db:  db,
		log: log,
	}
}

// CreateEmailGroup creates a new email group with members
func (s *alertEmailGroupStorage) CreateEmailGroup(ctx context.Context, req *dto.CreateAlertEmailGroupRequest, createdBy uuid.UUID) (*dto.AlertEmailGroup, error) {
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Insert the group
	groupQuery := `
		INSERT INTO alert_email_groups (name, description, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_by, created_at, updated_at, updated_by
	`

	var group dto.AlertEmailGroup
	err = tx.QueryRow(ctx, groupQuery, req.Name, req.Description, createdBy).Scan(
		&group.ID, &group.Name, &group.Description, &group.CreatedBy,
		&group.CreatedAt, &group.UpdatedAt, &group.UpdatedBy,
	)
	if err != nil {
		s.log.Error("Failed to create email group", zap.Error(err))
		return nil, err
	}

	// Insert email members
	memberQuery := `
		INSERT INTO alert_email_group_members (group_id, email)
		VALUES ($1, $2)
		ON CONFLICT (group_id, email) DO NOTHING
	`

	for _, email := range req.Emails {
		_, err = tx.Exec(ctx, memberQuery, group.ID, email)
		if err != nil {
			s.log.Error("Failed to insert email member", zap.Error(err), zap.String("email", email))
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		s.log.Error("Failed to commit transaction", zap.Error(err))
		return nil, err
	}

	// Fetch group with emails
	return s.GetEmailGroupByID(ctx, group.ID)
}

// GetEmailGroupByID gets an email group by ID with its members
func (s *alertEmailGroupStorage) GetEmailGroupByID(ctx context.Context, id uuid.UUID) (*dto.AlertEmailGroup, error) {
	groupQuery := `
		SELECT id, name, description, created_by, created_at, updated_at, updated_by
		FROM alert_email_groups
		WHERE id = $1
	`

	var group dto.AlertEmailGroup
	err := s.db.GetPool().QueryRow(ctx, groupQuery, id).Scan(
		&group.ID, &group.Name, &group.Description, &group.CreatedBy,
		&group.CreatedAt, &group.UpdatedAt, &group.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.log.Error("Failed to get email group", zap.Error(err))
		return nil, err
	}

	// Get email members
	membersQuery := `
		SELECT email
		FROM alert_email_group_members
		WHERE group_id = $1
		ORDER BY created_at
	`

	rows, err := s.db.GetPool().Query(ctx, membersQuery, id)
	if err != nil {
		s.log.Error("Failed to get email members", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			s.log.Error("Failed to scan email", zap.Error(err))
			continue
		}
		emails = append(emails, email)
	}

	group.Emails = emails
	return &group, nil
}

// GetAllEmailGroups gets all email groups with their members
func (s *alertEmailGroupStorage) GetAllEmailGroups(ctx context.Context) ([]dto.AlertEmailGroup, error) {
	// Use a LEFT JOIN to get all groups and their emails in a single query
	query := `
		SELECT 
			g.id, 
			g.name, 
			g.description, 
			g.created_by, 
			g.created_at, 
			g.updated_at, 
			g.updated_by,
			m.email
		FROM alert_email_groups g
		LEFT JOIN alert_email_group_members m ON g.id = m.group_id
		ORDER BY g.created_at DESC, m.created_at ASC
	`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.log.Error("Failed to get email groups", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var groups []dto.AlertEmailGroup
	groupMap := make(map[uuid.UUID]*dto.AlertEmailGroup)

	for rows.Next() {
		var groupID uuid.UUID
		var name string
		var description *string
		var createdBy *uuid.UUID
		var createdAt time.Time
		var updatedAt time.Time
		var updatedBy *uuid.UUID
		var email sql.NullString

		err := rows.Scan(
			&groupID, &name, &description, &createdBy,
			&createdAt, &updatedAt, &updatedBy, &email,
		)
		if err != nil {
			s.log.Error("Failed to scan email group row", zap.Error(err))
			continue
		}

		// Check if we've seen this group before
		group, exists := groupMap[groupID]
		if !exists {
			// Create new group
			group = &dto.AlertEmailGroup{
				ID:          groupID,
				Name:        name,
				Description: description,
				CreatedBy:   createdBy,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				UpdatedBy:   updatedBy,
				Emails:      []string{},
			}
			groups = append(groups, *group)
			groupMap[groupID] = &groups[len(groups)-1]
		}

		// Add email if it exists (LEFT JOIN can return NULL)
		if email.Valid && email.String != "" {
			groupMap[groupID].Emails = append(groupMap[groupID].Emails, email.String)
		}
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		s.log.Error("Error occurred while iterating email groups", zap.Error(err))
		return nil, err
	}

	s.log.Debug("Successfully fetched email groups", zap.Int("total_groups", len(groups)))
	return groups, nil
}

// UpdateEmailGroup updates an email group
func (s *alertEmailGroupStorage) UpdateEmailGroup(ctx context.Context, id uuid.UUID, req *dto.UpdateAlertEmailGroupRequest, updatedBy uuid.UUID) (*dto.AlertEmailGroup, error) {
	tx, err := s.db.GetPool().Begin(ctx)
	if err != nil {
		s.log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *req.Name)
		argPos++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argPos))
		args = append(args, *req.Description)
		argPos++
	}

	// Always update updated_by and updated_at if there are any changes (name, description, or emails)
	hasChanges := len(updates) > 0 || req.Emails != nil
	if hasChanges {
		updates = append(updates, fmt.Sprintf("updated_by = $%d", argPos))
		args = append(args, updatedBy)
		argPos++

		query := fmt.Sprintf(`
			UPDATE alert_email_groups
			SET %s, updated_at = NOW()
			WHERE id = $%d
		`, strings.Join(updates, ", "), argPos)
		args = append(args, id)

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			s.log.Error("Failed to update email group", zap.Error(err))
			return nil, err
		}
	}

	// Update email members if provided
	if req.Emails != nil {
		// Delete existing members
		_, err = tx.Exec(ctx, "DELETE FROM alert_email_group_members WHERE group_id = $1", id)
		if err != nil {
			s.log.Error("Failed to delete email members", zap.Error(err))
			return nil, err
		}

		// Insert new members
		memberQuery := `
			INSERT INTO alert_email_group_members (group_id, email)
			VALUES ($1, $2)
		`

		for _, email := range req.Emails {
			_, err = tx.Exec(ctx, memberQuery, id, email)
			if err != nil {
				s.log.Error("Failed to insert email member", zap.Error(err), zap.String("email", email))
				return nil, err
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		s.log.Error("Failed to commit transaction", zap.Error(err))
		return nil, err
	}

	return s.GetEmailGroupByID(ctx, id)
}

// DeleteEmailGroup deletes an email group (cascade deletes members)
func (s *alertEmailGroupStorage) DeleteEmailGroup(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM alert_email_groups WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		s.log.Error("Failed to delete email group", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetEmailsByGroupIDs gets all emails for the given group IDs
func (s *alertEmailGroupStorage) GetEmailsByGroupIDs(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(groupIDs) == 0 {
		return make(map[uuid.UUID][]string), nil
	}

	query := `
		SELECT group_id, email
		FROM alert_email_group_members
		WHERE group_id = ANY($1)
		ORDER BY created_at
	`

	rows, err := s.db.GetPool().Query(ctx, query, pq.Array(groupIDs))
	if err != nil {
		s.log.Error("Failed to get emails by group IDs", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var groupID uuid.UUID
		var email string
		if err := rows.Scan(&groupID, &email); err != nil {
			s.log.Error("Failed to scan email", zap.Error(err))
			continue
		}
		result[groupID] = append(result[groupID], email)
	}

	return result, nil
}
