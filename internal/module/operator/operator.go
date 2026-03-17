package operator

import (
	"context"

	"github.com/tucanbit/internal/constant/dto"
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

func (m *operatorModule) CreateOperator(ctx context.Context, req dto.CreateOperatorReq) (dto.Operator, error) {
	if req.OperatorID < 100000 || req.OperatorID > 999999 {
		return dto.Operator{}, errors.ErrInvalidUserInput.New("operator_id must be a 6-digit number")
	}
	if req.Name == "" || req.Code == "" {
		return dto.Operator{}, errors.ErrInvalidUserInput.New("name and code are required")
	}
	return m.storage.CreateOperator(ctx, req)
}

func (m *operatorModule) GetOperatorByID(ctx context.Context, operatorID int32) (dto.Operator, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return dto.Operator{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	op, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return dto.Operator{}, err
	}
	if !exists {
		return dto.Operator{}, errors.ErrResourceNotFound.New("operator not found")
	}
	return op, nil
}

func (m *operatorModule) GetOperators(ctx context.Context, req dto.GetOperatorsReq) (dto.GetOperatorsRes, error) {
	return m.storage.GetOperators(ctx, req)
}

func (m *operatorModule) UpdateOperator(ctx context.Context, req dto.UpdateOperatorReq) (dto.Operator, error) {
	if req.OperatorID < 100000 || req.OperatorID > 999999 {
		return dto.Operator{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	return m.storage.UpdateOperator(ctx, req)
}

func (m *operatorModule) DeleteOperator(ctx context.Context, operatorID int32) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.DeleteOperator(ctx, operatorID)
}

func (m *operatorModule) ChangeOperatorStatus(ctx context.Context, operatorID int32, req dto.ChangeOperatorStatusReq) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.UpdateOperatorStatus(ctx, operatorID, req.IsActive)
}

func (m *operatorModule) CreateOperatorCredential(ctx context.Context, operatorID int32) (dto.OperatorCredentialRes, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return dto.OperatorCredentialRes{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return dto.OperatorCredentialRes{}, err
	}
	if !exists {
		return dto.OperatorCredentialRes{}, errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.CreateOperatorCredential(ctx, operatorID)
}

func (m *operatorModule) RotateOperatorCredential(ctx context.Context, operatorID int32, credentialID int32) (dto.RotateOperatorCredentialRes, error) {
	if operatorID < 100000 || operatorID > 999999 || credentialID <= 0 {
		return dto.RotateOperatorCredentialRes{}, errors.ErrInvalidUserInput.New("invalid operator or credential ID")
	}
	return m.storage.RotateOperatorCredential(ctx, operatorID, credentialID)
}

func (m *operatorModule) GetActiveSigningKeyByOperatorID(ctx context.Context, operatorID int32) (string, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return "", errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	return m.storage.GetActiveSigningKeyByOperatorID(ctx, operatorID)
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

