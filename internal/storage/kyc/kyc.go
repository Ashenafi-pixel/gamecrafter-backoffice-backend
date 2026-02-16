package kyc

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

// KYCStorage interface for KYC operations
type KYCStorage interface {
	// Document operations
	CreateDocument(ctx context.Context, doc *dto.KYCDocument) (*dto.KYCDocument, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*dto.KYCDocument, error)
	GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCDocument, error)
	UpdateDocumentStatus(ctx context.Context, documentID uuid.UUID, status string, rejectionReason *string, reviewedBy uuid.UUID) error

	// Submission operations
	CreateSubmission(ctx context.Context, submission *dto.KYCSubmission) (*dto.KYCSubmission, error)
	GetSubmissionByID(ctx context.Context, id uuid.UUID) (*dto.KYCSubmission, error)
	GetSubmissionsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCSubmission, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID uuid.UUID, status string, notes *string, reviewedBy uuid.UUID) error

	// Status change operations
	CreateStatusChange(ctx context.Context, change *dto.KYCStatusChange) (*dto.KYCStatusChange, error)
	GetStatusChangesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCStatusChange, error)

	// List submissions
	GetAllSubmissions(ctx context.Context, status *string, limit, offset int) ([]dto.KYCSubmission, int64, error)

	// User KYC status operations
	UpdateUserKYCStatus(ctx context.Context, userID uuid.UUID, newStatus string, reason *string, changedBy uuid.UUID) error
	GetUserKYCStatus(ctx context.Context, userID uuid.UUID) (string, error)

	// Withdrawal block operations
	BlockWithdrawals(ctx context.Context, userID uuid.UUID, reason string, blockedBy uuid.UUID) error
	UnblockWithdrawals(ctx context.Context, userID uuid.UUID) error
	IsWithdrawalBlocked(ctx context.Context, userID uuid.UUID) (bool, error)

	// KYC Settings operations
	GetAllKYCSettings(ctx context.Context) ([]dto.KYCSettings, error)
	UpdateKYCSettings(ctx context.Context, id uuid.UUID, settingValue map[string]interface{}, description *string, isActive bool) error
}

// kycImpl implements KYCStorage interface
type kycImpl struct {
	db  *persistencedb.PersistenceDB
	log *zap.Logger
}

// NewKYCStorage creates a new KYC storage instance
func NewKYCStorage(db *persistencedb.PersistenceDB, log *zap.Logger) KYCStorage {
	return &kycImpl{
		db:  db,
		log: log,
	}
}

// CreateDocument creates a new KYC document
func (s *kycImpl) CreateDocument(ctx context.Context, doc *dto.KYCDocument) (*dto.KYCDocument, error) {
	s.log.Info("Creating KYC document", zap.String("userID", doc.UserID.String()))

	query := `
		INSERT INTO kyc_documents (id, user_id, document_type, file_url, file_name, upload_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, document_type, file_url, file_name, upload_date, status, rejection_reason, reviewed_by, review_date, created_at, updated_at
	`

	var createdDoc dto.KYCDocument
	err := s.db.GetPool().QueryRow(ctx, query,
		doc.ID, doc.UserID, doc.DocumentType, doc.FileUrl, doc.FileName,
		doc.UploadDate, doc.Status, doc.CreatedAt, doc.UpdatedAt,
	).Scan(
		&createdDoc.ID, &createdDoc.UserID, &createdDoc.DocumentType, &createdDoc.FileUrl, &createdDoc.FileName,
		&createdDoc.UploadDate, &createdDoc.Status, &createdDoc.RejectionReason, &createdDoc.ReviewedBy,
		&createdDoc.ReviewDate, &createdDoc.CreatedAt, &createdDoc.UpdatedAt,
	)

	if err != nil {
		s.log.Error("Failed to create KYC document", zap.Error(err))
		return nil, err
	}

	// Check if a submission already exists for this user
	existingSubmission := false
	checkQuery := `SELECT COUNT(*) FROM kyc_submissions WHERE user_id = $1 AND status = 'PENDING'`
	var count int
	err = s.db.GetPool().QueryRow(ctx, checkQuery, doc.UserID).Scan(&count)
	if err == nil && count > 0 {
		existingSubmission = true
	}

	// Create a submission record if it doesn't exist
	if !existingSubmission {
		submissionQuery := `
			INSERT INTO kyc_submissions (user_id, submission_type, status, submitted_at, created_at, updated_at)
			VALUES ($1, 'INITIAL', 'PENDING', NOW(), NOW(), NOW())
		`
		_, err = s.db.GetPool().Exec(ctx, submissionQuery, doc.UserID)
		if err != nil {
			s.log.Warn("Failed to create KYC submission record", zap.Error(err))
			// Don't fail the document creation if submission creation fails
		} else {
			s.log.Info("KYC submission record created automatically", zap.String("userID", doc.UserID.String()))
		}
	}

	s.log.Info("KYC document created successfully", zap.String("docID", createdDoc.ID.String()))
	return &createdDoc, nil
}

// GetDocumentByID retrieves a document by ID
func (s *kycImpl) GetDocumentByID(ctx context.Context, id uuid.UUID) (*dto.KYCDocument, error) {
	query := `
		SELECT id, user_id, document_type, file_url, file_name, upload_date, status, rejection_reason, reviewed_by, review_date, created_at, updated_at
		FROM kyc_documents
		WHERE id = $1
	`

	var doc dto.KYCDocument
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.UserID, &doc.DocumentType, &doc.FileUrl, &doc.FileName,
		&doc.UploadDate, &doc.Status, &doc.RejectionReason, &doc.ReviewedBy,
		&doc.ReviewDate, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.log.Error("Failed to get document by ID", zap.Error(err))
		return nil, err
	}

	return &doc, nil
}

// GetDocumentsByUserID retrieves all documents for a user
func (s *kycImpl) GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCDocument, error) {
	query := `
		SELECT id, user_id, document_type, file_url, file_name, upload_date, status, rejection_reason, reviewed_by, review_date, created_at, updated_at
		FROM kyc_documents
		WHERE user_id = $1
		ORDER BY upload_date DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.log.Error("Failed to get documents by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var documents []dto.KYCDocument
	for rows.Next() {
		var doc dto.KYCDocument
		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.DocumentType, &doc.FileUrl, &doc.FileName,
			&doc.UploadDate, &doc.Status, &doc.RejectionReason, &doc.ReviewedBy,
			&doc.ReviewDate, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			s.log.Error("Failed to scan document", zap.Error(err))
			return nil, err
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// UpdateDocumentStatus updates the status of a document
func (s *kycImpl) UpdateDocumentStatus(ctx context.Context, documentID uuid.UUID, status string, rejectionReason *string, reviewedBy uuid.UUID) error {
	s.log.Info("Updating document status", zap.String("docID", documentID.String()))

	query := `
		UPDATE kyc_documents
		SET status = $1, rejection_reason = $2, reviewed_by = $3, review_date = NOW(), updated_at = NOW()
		WHERE id = $4
	`

	result, err := s.db.GetPool().Exec(ctx, query, status, rejectionReason, reviewedBy, documentID)
	if err != nil {
		s.log.Error("Failed to update document status", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.log.Warn("No document found to update", zap.String("docID", documentID.String()))
		return sql.ErrNoRows
	}

	s.log.Info("Document status updated successfully")
	return nil
}

// CreateSubmission creates a new KYC submission
func (s *kycImpl) CreateSubmission(ctx context.Context, submission *dto.KYCSubmission) (*dto.KYCSubmission, error) {
	s.log.Info("Creating KYC submission", zap.String("userID", submission.UserID.String()))

	query := `
		INSERT INTO kyc_submissions (id, user_id, submission_date, overall_status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, submission_date, overall_status, notes, reviewed_by, review_date, created_at, updated_at
	`

	var createdSubmission dto.KYCSubmission
	err := s.db.GetPool().QueryRow(ctx, query,
		submission.ID, submission.UserID, submission.SubmissionDate,
		submission.OverallStatus, submission.Notes, submission.CreatedAt, submission.UpdatedAt,
	).Scan(
		&createdSubmission.ID, &createdSubmission.UserID, &createdSubmission.SubmissionDate,
		&createdSubmission.OverallStatus, &createdSubmission.Notes, &createdSubmission.ReviewedBy,
		&createdSubmission.ReviewDate, &createdSubmission.CreatedAt, &createdSubmission.UpdatedAt,
	)

	if err != nil {
		s.log.Error("Failed to create KYC submission", zap.Error(err))
		return nil, err
	}

	s.log.Info("KYC submission created successfully")
	return &createdSubmission, nil
}

// GetSubmissionByID retrieves a submission by ID
func (s *kycImpl) GetSubmissionByID(ctx context.Context, id uuid.UUID) (*dto.KYCSubmission, error) {
	query := `
		SELECT id, user_id, submission_date, overall_status, notes, reviewed_by, review_date, created_at, updated_at
		FROM kyc_submissions
		WHERE id = $1
	`

	var submission dto.KYCSubmission
	err := s.db.GetPool().QueryRow(ctx, query, id).Scan(
		&submission.ID, &submission.UserID, &submission.SubmissionDate,
		&submission.OverallStatus, &submission.Notes, &submission.ReviewedBy,
		&submission.ReviewDate, &submission.CreatedAt, &submission.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.log.Error("Failed to get submission by ID", zap.Error(err))
		return nil, err
	}

	return &submission, nil
}

// GetSubmissionsByUserID retrieves all submissions for a user
func (s *kycImpl) GetSubmissionsByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCSubmission, error) {
	query := `
		SELECT id, user_id, submitted_at, status, admin_notes, reviewed_by, reviewed_at, created_at, updated_at
		FROM kyc_submissions
		WHERE user_id = $1
		ORDER BY submitted_at DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.log.Error("Failed to get submissions by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var submissions []dto.KYCSubmission
	for rows.Next() {
		var submission dto.KYCSubmission
		err := rows.Scan(
			&submission.ID, &submission.UserID, &submission.SubmissionDate,
			&submission.OverallStatus, &submission.Notes, &submission.ReviewedBy,
			&submission.ReviewDate, &submission.CreatedAt, &submission.UpdatedAt,
		)
		if err != nil {
			s.log.Error("Failed to scan submission", zap.Error(err))
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// UpdateSubmissionStatus updates the status of a submission
func (s *kycImpl) UpdateSubmissionStatus(ctx context.Context, submissionID uuid.UUID, status string, notes *string, reviewedBy uuid.UUID) error {
	s.log.Info("Updating submission status", zap.String("submissionID", submissionID.String()))

	query := `
		UPDATE kyc_submissions
		SET overall_status = $1, notes = $2, reviewed_by = $3, review_date = NOW(), updated_at = NOW()
		WHERE id = $4
	`

	result, err := s.db.GetPool().Exec(ctx, query, status, notes, reviewedBy, submissionID)
	if err != nil {
		s.log.Error("Failed to update submission status", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.log.Warn("No submission found to update", zap.String("submissionID", submissionID.String()))
		return sql.ErrNoRows
	}

	s.log.Info("Submission status updated successfully")
	return nil
}

// CreateStatusChange creates a new status change record
func (s *kycImpl) CreateStatusChange(ctx context.Context, change *dto.KYCStatusChange) (*dto.KYCStatusChange, error) {
	s.log.Info("Creating KYC status change", zap.String("userID", change.UserID.String()))

	query := `
		INSERT INTO kyc_status_changes (id, user_id, old_status, new_status, changed_by, changed_at, change_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, old_status, new_status, changed_by, changed_at, change_reason
	`

	var createdChange dto.KYCStatusChange
	var changeReason sql.NullString
	err := s.db.GetPool().QueryRow(ctx, query,
		change.ID, change.UserID, change.OldStatus, change.NewStatus,
		change.ChangedBy, change.ChangeDate, change.Reason,
	).Scan(
		&createdChange.ID, &createdChange.UserID, &createdChange.OldStatus, &createdChange.NewStatus,
		&createdChange.ChangedBy, &createdChange.ChangeDate, &changeReason,
	)
	if changeReason.Valid {
		createdChange.Reason = &changeReason.String
	}

	if err != nil {
		s.log.Error("Failed to create status change", zap.Error(err))
		return nil, err
	}

	s.log.Info("Status change created successfully")
	return &createdChange, nil
}

// GetStatusChangesByUserID retrieves all status changes for a user
func (s *kycImpl) GetStatusChangesByUserID(ctx context.Context, userID uuid.UUID) ([]dto.KYCStatusChange, error) {
	query := `
		SELECT id, user_id, old_status, new_status, changed_by, changed_at, change_reason, admin_notes
		FROM kyc_status_changes
		WHERE user_id = $1
		ORDER BY changed_at DESC
	`

	rows, err := s.db.GetPool().Query(ctx, query, userID)
	if err != nil {
		s.log.Error("Failed to get status changes by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var changes []dto.KYCStatusChange
	for rows.Next() {
		var change dto.KYCStatusChange
		var changeReason, adminNotes sql.NullString
		err := rows.Scan(
			&change.ID, &change.UserID, &change.OldStatus, &change.NewStatus,
			&change.ChangedBy, &change.ChangeDate, &changeReason, &adminNotes,
		)
		if changeReason.Valid {
			change.Reason = &changeReason.String
		}
		if err != nil {
			s.log.Error("Failed to scan status change", zap.Error(err))
			return nil, err
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// GetAllSubmissions lists KYC submissions with optional status filter and pagination
func (s *kycImpl) GetAllSubmissions(ctx context.Context, status *string, limit, offset int) ([]dto.KYCSubmission, int64, error) {
	s.log.Info("Listing KYC submissions", zap.String("status", func() string {
		if status != nil {
			return *status
		}
		return ""
	}()))

	base := `
        SELECT 
            ks.id, ks.user_id, ks.submitted_at, ks.status, ks.admin_notes, ks.reviewed_by, ks.reviewed_at, ks.created_at, ks.updated_at,
            u.username, u.first_name, u.last_name, u.email, u.phone_number
        FROM kyc_submissions ks
        LEFT JOIN users u ON ks.user_id = u.id
    `
	countQ := `SELECT COUNT(*) FROM kyc_submissions`
	args := []interface{}{}
	where := ""
	if status != nil && *status != "" {
		where = " WHERE status = $1"
		args = append(args, *status)
	}
	orderLimit := " ORDER BY submitted_at DESC LIMIT $%d OFFSET $%d"
	// add limit/offset args
	args = append(args, limit, offset)

	// build order/limit with correct placeholders
	order := fmt.Sprintf(orderLimit, len(args)-1, len(args))

	// count total
	var total int64
	if where == "" {
		if err := s.db.GetPool().QueryRow(ctx, countQ).Scan(&total); err != nil {
			s.log.Error("Failed to count submissions", zap.Error(err))
			return nil, 0, err
		}
	} else {
		if err := s.db.GetPool().QueryRow(ctx, countQ+where, args[:1]...).Scan(&total); err != nil {
			s.log.Error("Failed to count submissions with filter", zap.Error(err))
			return nil, 0, err
		}
	}

	// query rows
	q := base + where + order
	rows, err := s.db.GetPool().Query(ctx, q, args...)
	if err != nil {
		s.log.Error("Failed to query submissions", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var items []dto.KYCSubmission
	for rows.Next() {
		var sub dto.KYCSubmission
		var username, firstName, lastName, email, phoneNumber sql.NullString
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.SubmissionDate, // Maps to submitted_at in DB
			&sub.OverallStatus,  // Maps to status in DB
			&sub.Notes,          // Maps to admin_notes in DB
			&sub.ReviewedBy,
			&sub.ReviewDate, // Maps to reviewed_at in DB
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&username,
			&firstName,
			&lastName,
			&email,
			&phoneNumber,
		); err != nil {
			s.log.Error("Failed to scan submission", zap.Error(err))
			return nil, 0, err
		}
		// Store user info in Notes field as JSON-like string for display
		if username.Valid {
			userInfo := fmt.Sprintf("__USER_INFO__:%s:%s:%s:%s:%s", username.String, firstName.String, lastName.String, email.String, phoneNumber.String)
			if sub.Notes != nil {
				updatedNotes := fmt.Sprintf("%s | %s", *sub.Notes, userInfo)
				sub.Notes = &updatedNotes
			} else {
				sub.Notes = &userInfo
			}
		}
		items = append(items, sub)
	}

	return items, total, nil
}

// UpdateUserKYCStatus updates the KYC status of a user
func (s *kycImpl) UpdateUserKYCStatus(ctx context.Context, userID uuid.UUID, newStatus string, reason *string, changedBy uuid.UUID) error {
	s.log.Info("Updating user KYC status", zap.String("userID", userID.String()))

	// Get current status
	currentStatus := "NO_KYC"
	var currentStatusStr string
	err := s.db.GetPool().QueryRow(ctx, "SELECT kyc_status FROM users WHERE id = $1", userID).Scan(&currentStatusStr)
	if err != nil && err != sql.ErrNoRows {
		s.log.Error("Failed to get current KYC status", zap.Error(err))
		return err
	}
	if currentStatusStr != "" {
		currentStatus = currentStatusStr
	}

	// Update user KYC status (the trigger will create the status change record)
	query := `UPDATE users SET kyc_status = $1 WHERE id = $2`
	result, err := s.db.GetPool().Exec(ctx, query, newStatus, userID)
	if err != nil {
		s.log.Error("Failed to update user KYC status", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.log.Warn("No user found to update", zap.String("userID", userID.String()))
		return sql.ErrNoRows
	}

	// Map KYC status to submission status
	var submissionStatus string
	switch newStatus {
	case "NO_KYC":
		submissionStatus = "PENDING"
	case "ID_VERIFIED", "ID_SOF_VERIFIED":
		submissionStatus = "APPROVED"
	case "KYC_FAILED":
		submissionStatus = "REJECTED"
	default:
		submissionStatus = "PENDING"
	}

	// Update pending submissions to match the new KYC status
	// Only update PENDING submissions to avoid overriding manual reviews
	updateSubmissionQuery := `
		UPDATE kyc_submissions 
		SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW()
		WHERE user_id = $3 AND (status = 'PENDING' OR status = 'UNDER_REVIEW')
	`
	_, err = s.db.GetPool().Exec(ctx, updateSubmissionQuery, submissionStatus, changedBy, userID)
	if err != nil {
		s.log.Warn("Failed to update submission status", zap.Error(err))
		// Don't fail the whole operation if this fails
	} else {
		s.log.Info("Submission status updated to match KYC status", zap.String("submissionStatus", submissionStatus))
	}

	// Create status change record manually (in case trigger doesn't work)
	statusChange := &dto.KYCStatusChange{
		ID:         uuid.New(),
		UserID:     userID,
		OldStatus:  currentStatus,
		NewStatus:  newStatus,
		ChangedBy:  &changedBy,
		ChangeDate: time.Now(),
		Reason:     reason,
		CreatedAt:  time.Now(),
	}
	_, err = s.CreateStatusChange(ctx, statusChange)
	if err != nil {
		s.log.Warn("Failed to create status change record", zap.Error(err))
		// Don't fail the whole operation if this fails
	}

	s.log.Info("User KYC status updated successfully")
	return nil
}

// GetUserKYCStatus retrieves the KYC status of a user
func (s *kycImpl) GetUserKYCStatus(ctx context.Context, userID uuid.UUID) (string, error) {
	query := `SELECT COALESCE(kyc_status, 'NO_KYC') FROM users WHERE id = $1`

	var statusStr string
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(&statusStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return "NO_KYC", nil
		}
		s.log.Error("Failed to get user KYC status", zap.Error(err))
		return "NO_KYC", err
	}

	return statusStr, nil
}

// BlockWithdrawals blocks withdrawals for a user
func (s *kycImpl) BlockWithdrawals(ctx context.Context, userID uuid.UUID, reason string, blockedBy uuid.UUID) error {
	s.log.Info("Blocking withdrawals", zap.String("userID", userID.String()))

	query := `UPDATE users SET withdrawal_restricted = TRUE WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query, userID)
	if err != nil {
		s.log.Error("Failed to block withdrawals", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.log.Warn("No user found to block", zap.String("userID", userID.String()))
		return sql.ErrNoRows
	}

	s.log.Info("Withdrawals blocked successfully")
	return nil
}

// UnblockWithdrawals unblocks withdrawals for a user
func (s *kycImpl) UnblockWithdrawals(ctx context.Context, userID uuid.UUID) error {
	s.log.Info("Unblocking withdrawals", zap.String("userID", userID.String()))

	query := `UPDATE users SET withdrawal_restricted = FALSE WHERE id = $1`

	result, err := s.db.GetPool().Exec(ctx, query, userID)
	if err != nil {
		s.log.Error("Failed to unblock withdrawals", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		s.log.Warn("No user found to unblock", zap.String("userID", userID.String()))
		return sql.ErrNoRows
	}

	s.log.Info("Withdrawals unblocked successfully")
	return nil
}

// IsWithdrawalBlocked checks if withdrawals are blocked for a user
func (s *kycImpl) IsWithdrawalBlocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT COALESCE(withdrawal_restricted, FALSE) FROM users WHERE id = $1`

	var blocked bool
	err := s.db.GetPool().QueryRow(ctx, query, userID).Scan(&blocked)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		s.log.Error("Failed to check withdrawal block status", zap.Error(err))
		return false, err
	}

	return blocked, nil
}

// GetAllKYCSettings retrieves all KYC settings ordered by id ASC
func (s *kycImpl) GetAllKYCSettings(ctx context.Context) ([]dto.KYCSettings, error) {
	s.log.Info("Getting all KYC settings")

	query := `
		SELECT id, setting_key, setting_value, description, is_active, created_at, updated_at
		FROM kyc_settings
		ORDER BY id ASC
	`

	rows, err := s.db.GetPool().Query(ctx, query)
	if err != nil {
		s.log.Error("Failed to query KYC settings", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var settings []dto.KYCSettings
	for rows.Next() {
		var setting dto.KYCSettings
		var settingValueJSON []byte
		var description sql.NullString

		err := rows.Scan(
			&setting.ID,
			&setting.SettingKey,
			&settingValueJSON,
			&description,
			&setting.IsActive,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			s.log.Error("Failed to scan KYC setting", zap.Error(err))
			continue
		}

		// Parse JSONB to map
		var settingValue map[string]interface{}
		if err := json.Unmarshal(settingValueJSON, &settingValue); err != nil {
			s.log.Error("Failed to unmarshal setting_value JSON", zap.Error(err))
			settingValue = make(map[string]interface{})
		}
		setting.SettingValue = settingValue

		if description.Valid {
			setting.Description = &description.String
		}

		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		s.log.Error("Error iterating KYC settings", zap.Error(err))
		return nil, err
	}

	return settings, nil
}

// UpdateKYCSettings updates a KYC setting
func (s *kycImpl) UpdateKYCSettings(ctx context.Context, id uuid.UUID, settingValue map[string]interface{}, description *string, isActive bool) error {
	s.log.Info("Updating KYC setting", zap.String("id", id.String()))

	// Marshal setting_value to JSON
	settingValueJSON, err := json.Marshal(settingValue)
	if err != nil {
		s.log.Error("Failed to marshal setting_value", zap.Error(err))
		return err
	}

	query := `
		UPDATE kyc_settings
		SET setting_value = $1,
		    description = $2,
		    is_active = $3,
		    updated_at = NOW()
		WHERE id = $4
	`

	result, err := s.db.GetPool().Exec(ctx, query, settingValueJSON, description, isActive, id)
	if err != nil {
		s.log.Error("Failed to update KYC setting", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("KYC setting with id %s not found", id.String())
	}

	s.log.Info("KYC setting updated successfully", zap.String("id", id.String()))
	return nil
}
