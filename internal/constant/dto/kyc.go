package dto

import (
	"time"

	"github.com/google/uuid"
)

// KYC Document Types
const (
	DocumentTypeIDFront        = "ID_FRONT"
	DocumentTypeIDBack         = "ID_BACK"
	DocumentTypeProofOfAddress = "PROOF_OF_ADDRESS"
	DocumentTypeSelfieWithID   = "SELFIE_WITH_ID"
	DocumentTypeBankStatement  = "BANK_STATEMENT"
	DocumentTypeSOFDocument    = "SOF_DOCUMENT"
)

// KYC Status Values
const (
	StatusNoKYC         = "NO_KYC"
	StatusIDVerified    = "ID_VERIFIED"
	StatusIDSOFVerified = "ID_SOF_VERIFIED"
	StatusKYCFailed     = "KYC_FAILED"
)

// Document Status
const (
	DocumentStatusPending  = "PENDING"
	DocumentStatusApproved = "APPROVED"
	DocumentStatusRejected = "REJECTED"
)

// Submission Status
const (
	SubmissionStatusSubmitted   = "SUBMITTED"
	SubmissionStatusUnderReview = "UNDER_REVIEW"
	SubmissionStatusApproved    = "APPROVED"
	SubmissionStatusRejected    = "REJECTED"
	SubmissionStatusExpired     = "EXPIRED"
)

// KYC Document represents an uploaded document
type KYCDocument struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	DocumentType    string     `json:"document_type"`
	FileUrl         string     `json:"file_url"`
	FileName        string     `json:"file_name"`
	UploadDate      time.Time  `json:"upload_date"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewDate      *time.Time `json:"review_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// KYC Submission represents a KYC application
type KYCSubmission struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	SubmissionDate time.Time  `json:"submission_date"`
	OverallStatus  string     `json:"overall_status"`
	Notes          *string    `json:"notes,omitempty"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewDate     *time.Time `json:"review_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// KYC Status Change represents an audit log entry
type KYCStatusChange struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	OldStatus  string     `json:"old_status"`
	NewStatus  string     `json:"new_status"`
	ChangedBy  *uuid.UUID `json:"changed_by,omitempty"`
	ChangeDate time.Time  `json:"change_date"`
	Reason     *string    `json:"reason,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Request DTOs for KYC operations

// UploadKYCDocumentRequest represents a request to upload a KYC document
type UploadKYCDocumentRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	DocumentType string    `json:"document_type" binding:"required"`
	FileName     string    `json:"file_name" binding:"required"`
	FileSize     int64     `json:"file_size" binding:"required"`
	MimeType     string    `json:"mime_type" binding:"required"`
}

// UpdateDocumentStatusRequest represents a request to update document status
type UpdateDocumentStatusRequest struct {
	DocumentID      uuid.UUID `json:"document_id" binding:"required"`
	Status          string    `json:"status" binding:"required"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
}

// UpdateUserKYCStatusRequest represents a request to update user KYC status
type UpdateUserKYCStatusRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	NewStatus string    `json:"new_status" binding:"required"`
	Reason    *string   `json:"reason,omitempty"`
}

// KYCSettings represents a KYC setting from the kyc_settings table
type KYCSettings struct {
	ID           uuid.UUID              `json:"id"`
	SettingKey   string                 `json:"setting_key"`
	SettingValue map[string]interface{} `json:"setting_value"` // JSONB field
	Description  *string                `json:"description,omitempty"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// BlockWithdrawalsRequest represents a request to block user withdrawals
type BlockWithdrawalsRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Reason string    `json:"reason" binding:"required"`
}
