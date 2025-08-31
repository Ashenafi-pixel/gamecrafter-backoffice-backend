package types

import "context"

type EventHandler func(ctx context.Context, message []byte) (bool, error)
