package operator

import (
	"context"

	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
)

type operatorModule struct {
	storage storage.Operator
}

func NewOperatorModule(operatorStorage storage.Operator) module.Operator {
	return &operatorModule{
		storage: operatorStorage,
	}
}

func (m *operatorModule) AssignGamesToOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if operatorID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	if len(gameIDs) == 0 {
		return errors.ErrInvalidUserInput.New("game_ids is required")
	}

	return m.storage.AssignGamesToOperator(ctx, operatorID, gameIDs)
}

func (m *operatorModule) RevokeGamesFromOperator(ctx context.Context, operatorID int32, gameIDs []string) error {
	if operatorID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	if len(gameIDs) == 0 {
		return errors.ErrInvalidUserInput.New("game_ids is required")
	}

	return m.storage.RevokeGamesFromOperator(ctx, operatorID, gameIDs)
}

var _ module.Operator = (*operatorModule)(nil)

