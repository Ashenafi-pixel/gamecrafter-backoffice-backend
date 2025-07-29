package platform

import (
	"context"
	"encoding/json"
)

type Kafka interface {
	WriteMessage(ctx context.Context, topic, key string, value interface{}) error
	RegisterKafkaEventHandler(eventType string, handler EventHandler)
	StartConsumer(ctx context.Context)
}

// EventHandler defines the signature for event handler functions
type EventHandler func(ctx context.Context, event json.RawMessage) (bool, error)

// Event represents a generic event structure used across the platform
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
