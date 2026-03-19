package dto

import (
	"time"

	"github.com/google/uuid"
)

type OperatorKYCDocument struct {
	ID              uuid.UUID  `json:"id"`
	OperatorID      int32      `json:"operator_id"`
	DocumentType    string     `json:"document_type"`
	FileURL         string     `json:"file_url"`
	FileName        string     `json:"file_name"`
	UploadDate      time.Time  `json:"upload_date"`
	Status          string     `json:"status"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewDate      *time.Time `json:"review_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type OperatorKYCSubmission struct {
	ID            uuid.UUID  `json:"id"`
	OperatorID    int32      `json:"operator_id"`
	SubmissionType string    `json:"submission_type"`
	Status        string     `json:"status"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	ReviewedBy    *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	AdminNotes    *string    `json:"admin_notes,omitempty"`
	AutoTriggered bool       `json:"auto_triggered"`
	TriggerReason *string    `json:"trigger_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

