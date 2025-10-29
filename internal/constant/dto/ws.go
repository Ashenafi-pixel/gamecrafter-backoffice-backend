package dto

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket message types for session management
const (
	WS_SESSION_EXPIRED   = "session_expired"
	WS_SESSION_REFRESHED = "session_refreshed"
)

type WSMessageRequest struct {
	Type        string `json:"type"  swaggerignore:"true"`
	AccessToken string `json:"access_token" `
	Data        any    `json:"data"  swaggerignore:"true"`
}

type BroadCastPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Conn   *websocket.Conn
}

type WSRes struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// SessionEventMessage represents WebSocket messages for session-related events
type SessionEventMessage struct {
	Type    string                 `json:"type"`           // "session_warning", "session_expired", "session_refreshed"
	Message string                 `json:"message"`        // Human-readable message
	Action  string                 `json:"action"`         // "login", "refresh", "dismiss"
	Timeout int                    `json:"timeout"`        // Seconds until auto-action (optional)
	Data    map[string]interface{} `json:"data,omitempty"` // Additional data
}
