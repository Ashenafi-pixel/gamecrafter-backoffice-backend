package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AccountBlockReq struct {
	ID          uuid.UUID  `json:"id"  swaggerignore:"true"`
	UserID      uuid.UUID  `json:"user_id"   swaggerignore:"true"`
	BlockedBy   uuid.UUID  `json:"blocked_by"  swaggerignore:"true"`
	BlockedFrom *time.Time `json:"blocked_from"`
	BlockedTo   *time.Time `json:"blocked_to"`
	Duration    string     `json:"duration"`
	Type        string     `json:"type"`
	Reason      string     `json:"reason"`
	Note        string     `json:"note"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AccountBlockRes struct {
	Message  string    `json:"message"`
	UserID   uuid.UUID `json:"user_id"`
	Type     string    `json:"type"`
	Duration string    `json:"duration"`
}

type GetBlockedAccountReq struct {
	UserID      uuid.UUID  `json:"user_id"`
	Type        string     `json:"type"`
	BlockedFrom *time.Time `json:"blocked_from"`
	BlockedTo   *time.Time `json:"blocked_to"`
}

type GetBlockedAccountLogReq struct {
	AdminID  uuid.UUID `json:"admin_id"  swaggerignore:"true"`
	Type     string    `json:"type"`
	UserID   uuid.UUID `json:"user_id"`
	Duration string    `json:"duration"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}

type GetBlockedAccountLogRep struct {
	BlockedAccount AccountBlockReq `json:"blocked_account"`
	User           User            `json:"user"`
	BlockedBy      User            `json:"blocked_by"`
	Total_pages    int             `json:"total_pages"`
}

type IPFilter struct {
	ID          uuid.UUID `json:"id"`
	StartIP     string    `json:"start_ip"`
	EndIP       string    `json:"end_ip"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Hits        int       `json:"hits"`
	LastHit     time.Time `json:"last_hit"`
	Type        string    `json:"type"`
}

type IpFilterReq struct {
	StartIP     string    `json:"start_ip" validate:"required"`
	EndIP       string    `json:"end_ip"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Type        string    `json:"type"`
}

type IPFilterRes struct {
	Message string   `json:"message"`
	Data    IPFilter `json:"data"`
}

func ValidateIP(rp IPFilter) error {
	validate := validator.New()
	validate.RegisterValidation("validpassword", isValidPassword)
	return validate.Struct(rp)
}

type GetIPFilterReq struct {
	Type    string `json:"type"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

type GetIPFilterRes struct {
	IPFilters  []GetIPFilterResData `json:"ip_filters"`
	TotalPages int                  `json:"total_pages"`
}
type GetIPFilterResData struct {
	ID          uuid.UUID `json:"id"`
	StartIP     string    `json:"start_ip"`
	EndIP       string    `json:"end_ip"`
	Description string    `json:"description"`
	CreatedBy   User      `json:"created_by"`
	Hists       int       `json:"hits"`
	Type        string    `json:"type"`
	LastHit     time.Time `json:"last_hit"`
}

type RemoveIPBlockReq struct {
	ID uuid.UUID `json:"id"`
}

type RemoveIPBlockRes struct {
	Message string `json:"message"`
}
