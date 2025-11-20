package dto

import (
	"github.com/google/uuid"
)

// Page represents a page/route in the system
type Page struct {
	ID       uuid.UUID  `json:"id"`
	Path     string     `json:"path"`
	Label    string     `json:"label"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Icon     string     `json:"icon,omitempty"`
}

// CreatePageReq represents a request to create a new page
type CreatePageReq struct {
	Path     string     `json:"path" validate:"required"`
	Label    string     `json:"label" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Icon     string     `json:"icon,omitempty"`
}

// UserAllowedPages represents the allowed pages for a user
type UserAllowedPages struct {
	UserID uuid.UUID `json:"user_id"`
	Pages  []Page    `json:"pages"`
}

// AssignPagesToUserReq represents a request to assign pages to a user
type AssignPagesToUserReq struct {
	UserID  uuid.UUID   `json:"user_id" validate:"required"`
	PageIDs []uuid.UUID `json:"page_ids" validate:"required,min=1"`
}

// GetUserAllowedPagesRes represents the response for getting user's allowed pages
type GetUserAllowedPagesRes struct {
	UserID uuid.UUID `json:"user_id"`
	Pages  []Page    `json:"pages"`
}

