package operator

import (
	"context"
	"strings"

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

func (m *operatorModule) AssignAllGamesToOperator(ctx context.Context, operatorID int32) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.AssignAllGamesToOperator(ctx, operatorID)
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

func (m *operatorModule) AssignProviderToOperator(ctx context.Context, operatorID int32, providerID string) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	if providerID == "" {
		return errors.ErrInvalidUserInput.New("provider_id is required")
	}

	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}

	return m.storage.AssignProviderToOperator(ctx, operatorID, providerID)
}

func (m *operatorModule) RevokeProviderFromOperator(ctx context.Context, operatorID int32, providerID string) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	if providerID == "" {
		return errors.ErrInvalidUserInput.New("provider_id is required")
	}

	return m.storage.RevokeProviderFromOperator(ctx, operatorID, providerID)
}

func (m *operatorModule) GetOperatorGameIDs(ctx context.Context, operatorID int32) ([]string, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return nil, errors.ErrInvalidUserInput.New("invalid operator ID")
	}

	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.ErrResourceNotFound.New("operator not found")
	}

	return m.storage.GetOperatorGameIDs(ctx, operatorID)
}

func (m *operatorModule) GetOperatorGames(ctx context.Context, operatorID int32) ([]dto.GameResponse, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return nil, errors.ErrInvalidUserInput.New("invalid operator ID")
	}

	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.ErrResourceNotFound.New("operator not found")
	}

	return m.storage.GetOperatorGames(ctx, operatorID)
}

func (m *operatorModule) AddOperatorAllowedOrigin(ctx context.Context, operatorID int32, req dto.AddOperatorAllowedOriginReq) (dto.OperatorAllowedOriginRes, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return dto.OperatorAllowedOriginRes{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	origin := strings.TrimSpace(req.Origin)
	if origin == "" {
		return dto.OperatorAllowedOriginRes{}, errors.ErrInvalidUserInput.New("origin is required")
	}
	if len(origin) > 255 {
		return dto.OperatorAllowedOriginRes{}, errors.ErrInvalidUserInput.New("origin is too long")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return dto.OperatorAllowedOriginRes{}, err
	}
	if !exists {
		return dto.OperatorAllowedOriginRes{}, errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.AddOperatorAllowedOrigin(ctx, operatorID, origin)
}

func (m *operatorModule) RemoveOperatorAllowedOrigin(ctx context.Context, operatorID int32, originID int32) error {
	if operatorID < 100000 || operatorID > 999999 || originID <= 0 {
		return errors.ErrInvalidUserInput.New("invalid operator or origin ID")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.RemoveOperatorAllowedOrigin(ctx, operatorID, originID)
}

func (m *operatorModule) ListOperatorAllowedOrigins(ctx context.Context, operatorID int32) (dto.ListOperatorAllowedOriginsRes, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return dto.ListOperatorAllowedOriginsRes{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return dto.ListOperatorAllowedOriginsRes{}, err
	}
	if !exists {
		return dto.ListOperatorAllowedOriginsRes{}, errors.ErrResourceNotFound.New("operator not found")
	}
	origins, err := m.storage.ListOperatorAllowedOrigins(ctx, operatorID)
	if err != nil {
		return dto.ListOperatorAllowedOriginsRes{}, err
	}
	return dto.ListOperatorAllowedOriginsRes{Origins: origins}, nil
}

func (m *operatorModule) GetOperatorFeatureFlags(ctx context.Context, operatorID int32) (dto.OperatorFeatureFlagsRes, error) {
	if operatorID < 100000 || operatorID > 999999 {
		return dto.OperatorFeatureFlagsRes{}, errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return dto.OperatorFeatureFlagsRes{}, err
	}
	if !exists {
		return dto.OperatorFeatureFlagsRes{}, errors.ErrResourceNotFound.New("operator not found")
	}

	flags, err := m.storage.GetOperatorFeatureFlags(ctx, operatorID)
	if err != nil {
		return dto.OperatorFeatureFlagsRes{}, err
	}
	if flags == nil {
		flags = map[string]bool{}
	}
	return dto.OperatorFeatureFlagsRes{Flags: flags}, nil
}

func (m *operatorModule) UpdateOperatorFeatureFlags(ctx context.Context, operatorID int32, req dto.UpdateOperatorFeatureFlagsReq) error {
	if operatorID < 100000 || operatorID > 999999 {
		return errors.ErrInvalidUserInput.New("invalid operator ID")
	}
	if req.Flags == nil {
		return errors.ErrInvalidUserInput.New("flags is required")
	}
	// Ensure operator exists
	_, exists, err := m.storage.GetOperatorByID(ctx, operatorID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrResourceNotFound.New("operator not found")
	}
	return m.storage.UpdateOperatorFeatureFlags(ctx, operatorID, req.Flags)
}

var _ module.Operator = (*operatorModule)(nil)

