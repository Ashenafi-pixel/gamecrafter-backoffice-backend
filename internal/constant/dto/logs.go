package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	IPAddress   string    `json:"ip_address" validate:"max=50"`
	Success     bool      `json:"success"`
	AttemptTime time.Time `json:"attempt_time"`
	UserAgent   string    `json:"user_agent"`
}

type UserSessions struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"user_id"`
	Token                 string    `json:"token"`
	ExpiresAt             time.Time `json:"expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	IpAddress             string    `json:"ip_address"`
	UserAgent             string    `json:"user_agent"`
	CreatedAt             time.Time `json:"created_at"`
}

type SystemLogs struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	Module    string      `json:"module"`
	Detail    interface{} `json:"detail"`
	IPAddress string      `json:"ip_address"`
	Timestamp time.Time   `json:"timestamp"`
	Roles     interface{} `json:"roles"`
}

type GetSystemLogsReq struct {
	From    *time.Time `json:"from"`
	To      *time.Time `json:"to"`
	Module  string     `json:"module"`
	UserID  uuid.UUID  `json:"user_id"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
}

type SystemLogsRes struct {
	SystemLogs []SystemLogs `json:"system_logs"`
	TotalPage  int          `json:"total_page"`
}
