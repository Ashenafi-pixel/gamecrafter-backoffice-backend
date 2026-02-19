package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Player represents a player entity
type Player struct {
	ID                    int32     `json:"id"`
	Email                 string    `json:"email"`
	Username              string    `json:"username"`
	Password              string    `json:"password,omitempty" swaggerignore:"true"`
	Phone                 *string   `json:"phone,omitempty"`
	FirstName             *string   `json:"first_name,omitempty"`
	LastName              *string   `json:"last_name,omitempty"`
	DefaultCurrency       string    `json:"default_currency"`
	Brand                 *string   `json:"brand,omitempty"`
	DateOfBirth           time.Time `json:"date_of_birth"`
	Country               string    `json:"country"`
	State                 *string   `json:"state,omitempty"`
	StreetAddress         *string   `json:"street_address,omitempty"`
	PostalCode            *string   `json:"postal_code,omitempty"`
	TestAccount           bool      `json:"test_account"`
	EnableWithdrawalLimit bool      `json:"enable_withdrawal_limit"`
	BrandID               *int32    `json:"brand_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DateOfBirth is a custom time type that handles empty strings gracefully
type DateOfBirth struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler interface
func (dob *DateOfBirth) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	timeStr := strings.Trim(string(data), `"`)

	// If empty string or null, set to zero time
	if timeStr == "" || timeStr == "null" {
		dob.Time = time.Time{}
		return nil
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,       // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05Z", // 2006-01-02T15:04:05Z
		"2006-01-02T15:04:05",  // 2006-01-02T15:04:05
		"2006-01-02T15:04",     // 2006-01-02T15:04
		"2006-01-02 15:04:05",  // 2006-01-02 15:04:05
		"2006-01-02 15:04",     // 2006-01-02 15:04
		"2006-01-02",           // 2006-01-02
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			dob.Time = t
			return nil
		}
	}

	// If no format matches, try adding default time if only date is provided
	if !strings.Contains(timeStr, ":") && !strings.Contains(timeStr, "T") {
		timeStr += "T00:00:00Z"
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			dob.Time = t
			return nil
		}
	}

	// If still can't parse, return error
	return &time.ParseError{
		Layout:     "multiple formats",
		Value:      timeStr,
		LayoutElem: "time",
		ValueElem:  timeStr,
		Message:    "unable to parse date_of_birth",
	}
}

// MarshalJSON implements json.Marshaler interface
func (dob DateOfBirth) MarshalJSON() ([]byte, error) {
	if dob.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(dob.Time.Format(time.RFC3339))
}

// CreatePlayerReq represents the request to create a new player
type CreatePlayerReq struct {
	Email                 string      `json:"email" validate:"required,email"`
	Username              string      `json:"username" validate:"required,min=3,max=255"`
	Password              string      `json:"password" validate:"required,min=8"`
	Phone                 *string     `json:"phone,omitempty" validate:"omitempty,max=20"`
	FirstName             *string     `json:"first_name,omitempty" validate:"omitempty,max=255"`
	LastName              *string     `json:"last_name,omitempty" validate:"omitempty,max=255"`
	DefaultCurrency       string      `json:"default_currency" validate:"required,max=10"`
	Brand                 *string     `json:"brand,omitempty" validate:"omitempty,max=255"`
	DateOfBirth           DateOfBirth `json:"date_of_birth" validate:"required"`
	Country               string      `json:"country" validate:"required,max=100"`
	State                 *string     `json:"state,omitempty" validate:"omitempty,max=100"`
	StreetAddress         *string     `json:"street_address,omitempty"`
	PostalCode            *string     `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	TestAccount           bool        `json:"test_account,omitempty"`
	EnableWithdrawalLimit bool        `json:"enable_withdrawal_limit,omitempty"`
	BrandID               *int32      `json:"brand_id,omitempty"`
}

// CreatePlayerRes represents the response after creating a player
type CreatePlayerRes struct {
	ID        int32     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdatePlayerReq represents the request to update a player
type UpdatePlayerReq struct {
	ID                    int32      `json:"id" validate:"required"`
	Email                 *string    `json:"email,omitempty" validate:"omitempty,email"`
	Username              *string    `json:"username,omitempty" validate:"omitempty,min=3,max=255"`
	Phone                 *string    `json:"phone,omitempty" validate:"omitempty,max=20"`
	FirstName             *string    `json:"first_name,omitempty" validate:"omitempty,max=255"`
	LastName              *string    `json:"last_name,omitempty" validate:"omitempty,max=255"`
	DefaultCurrency       *string    `json:"default_currency,omitempty" validate:"omitempty,max=10"`
	Brand                 *string    `json:"brand,omitempty" validate:"omitempty,max=255"`
	DateOfBirth           *time.Time `json:"date_of_birth,omitempty"`
	Country               *string    `json:"country,omitempty" validate:"omitempty,max=100"`
	State                 *string    `json:"state,omitempty" validate:"omitempty,max=100"`
	StreetAddress         *string    `json:"street_address,omitempty"`
	PostalCode            *string    `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	TestAccount           *bool      `json:"test_account,omitempty"`
	EnableWithdrawalLimit *bool      `json:"enable_withdrawal_limit,omitempty"`
	BrandID               *int32     `json:"brand_id,omitempty"`
}

// UpdatePlayerRes represents the response after updating a player
type UpdatePlayerRes struct {
	Player Player `json:"player"`
}

// GetPlayerReq represents the request to get a player by ID
type GetPlayerReq struct {
	ID int32 `uri:"id" binding:"required"`
}

// GetPlayerRes represents the response with player details
type GetPlayerRes struct {
	Player Player `json:"player"`
}

// GetPlayersReq represents the request to get a list of players
type GetPlayersReqs struct {
	Page        int     `form:"page" validate:"min=1"`
	PerPage     int     `form:"per_page" validate:"min=1,max=100"`
	Search      string  `form:"search,omitempty"`
	BrandID     *string `form:"brand_id,omitempty"`
	Country     *string `form:"country,omitempty"`
	TestAccount *bool   `form:"test_account,omitempty"`
	SortBy      string  `form:"sort_by,omitempty" validate:"omitempty,oneof=email username created_at updated_at date_of_birth country"`
	SortOrder   string  `form:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetPlayersRes represents the response with a list of players
type GetPlayersRess struct {
	Players     []Player `json:"players"`
	TotalCount  int      `json:"total_count"`
	TotalPages  int      `json:"total_pages"`
	CurrentPage int      `json:"current_page"`
	PerPage     int      `json:"per_page"`
}

// ValidateCreatePlayer validates the CreatePlayerReq
func ValidateCreatePlayer(req CreatePlayerReq) error {
	validate := validator.New()

	// Custom validation: DateOfBirth must not be zero
	if req.DateOfBirth.Time.IsZero() {
		return fmt.Errorf("date_of_birth is required and cannot be empty")
	}

	return validate.Struct(req)
}

// ValidateUpdatePlayer validates the UpdatePlayerReq
func ValidateUpdatePlayer(req UpdatePlayerReq) error {
	validate := validator.New()
	return validate.Struct(req)
}
